package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ishchibormi/bot/internal/envfile"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type otpDoc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	TGToken    string             `bson:"tgToken"`
	Phone      string             `bson:"phone,omitempty"`
	TelegramID int64              `bson:"telegramId,omitempty"`
	Code       string             `bson:"code,omitempty"`
	ExpiresAt  time.Time          `bson:"expiresAt"`
	Used       bool               `bson:"used"`
	CreatedAt  time.Time          `bson:"createdAt"`
}

type userDoc struct {
	Phone      string `bson:"phone"`
	TelegramID int64  `bson:"telegramId"`
}

func main() {
	envfile.Load()
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN required")
	}
	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getenv("MONGO_DB", "ishchibormi")
	otpTTL := time.Duration(envInt("OTP_TTL_SECONDS", 300)) * time.Second
	otpLen := envInt("OTP_LENGTH", 6)

	ctx := context.Background()
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("mongo: %v", err)
	}
	defer func() { _ = mc.Disconnect(ctx) }()
	otpCol := mc.Database(dbName).Collection("otp_codes")
	usersCol := mc.Database(dbName).Collection("users")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}
	log.Printf("bot started: @%s", bot.Self.UserName)

	// Pending tokens per chatID (so /start <token>, then contact comes later)
	pending := map[int64]string{}

	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 30
	updates := bot.GetUpdatesChan(upd)
	for u := range updates {
		if u.Message == nil {
			continue
		}
		m := u.Message
		switch {
		case m.IsCommand() && m.Command() == "start":
			args := strings.TrimSpace(m.CommandArguments())

			// ── Known user shortcut ───────────────────────────────────
			// If we've seen this Telegram user before AND we have their
			// phone in the users collection, skip the contact-share step
			// and just issue a fresh code immediately.
			if phone, ok := findKnownPhone(ctx, usersCol, m.From.ID); ok {
				code, err := generateAndStore(ctx, otpCol, args, phone, m.From.ID, otpTTL, otpLen)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(m.Chat.ID, "Xatolik yuz berdi. Iltimos, keyinroq qayta urinib ko'ring."))
					continue
				}
				msg := tgbotapi.NewMessage(m.Chat.ID,
					fmt.Sprintf("Qaytib xush kelibsiz!\n\nTasdiqlash kodingiz: *%s*\n\nKodni saytda kiriting. Kod %d daqiqa amal qiladi.",
						code, int(otpTTL/time.Minute)))
				msg.ParseMode = "Markdown"
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				_, _ = bot.Send(msg)
				delete(pending, m.Chat.ID)
				continue
			}

			// ── First-time flow: ask for contact ──────────────────────
			if args != "" {
				pending[m.Chat.ID] = args
			}
			req := tgbotapi.NewMessage(m.Chat.ID, "Salom! \"Ishchi Bormi\" ga xush kelibsiz.\n\nIltimos, telefon raqamingizni ulashing.")
			kb := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: "📞 Telefon raqamni ulashish", RequestContact: true}),
			)
			kb.OneTimeKeyboard = true
			kb.ResizeKeyboard = true
			req.ReplyMarkup = kb
			_, _ = bot.Send(req)

		case m.Contact != nil:
			phone := normalizePhone(m.Contact.PhoneNumber)
			code, err := generateAndStore(ctx, otpCol, pending[m.Chat.ID], phone, m.From.ID, otpTTL, otpLen)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(m.Chat.ID, "Xatolik yuz berdi. Iltimos, keyinroq qayta urinib ko'ring."))
				continue
			}
			ack := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Tasdiqlash kodingiz: *%s*\n\nKodni saytda kiriting. Kod %d daqiqa amal qiladi.", code, int(otpTTL/time.Minute)))
			ack.ParseMode = "Markdown"
			ack.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = bot.Send(ack)
			delete(pending, m.Chat.ID)

		default:
			_, _ = bot.Send(tgbotapi.NewMessage(m.Chat.ID, "Iltimos, /start buyrug'idan boshlang."))
		}
	}
}

// findKnownPhone looks up a user by telegramId in the users collection.
// Returns the verified phone if available.
func findKnownPhone(ctx context.Context, users *mongo.Collection, tgID int64) (string, bool) {
	if tgID == 0 {
		return "", false
	}
	var u userDoc
	err := users.FindOne(ctx,
		bson.M{"telegramId": tgID, "phone": bson.M{"$ne": ""}},
		options.FindOne().SetProjection(bson.M{"phone": 1, "telegramId": 1}),
	).Decode(&u)
	if err != nil || u.Phone == "" {
		return "", false
	}
	return u.Phone, true
}

func generateAndStore(ctx context.Context, col *mongo.Collection, token, phone string, tgID int64, ttl time.Duration, length int) (string, error) {
	code, err := randDigits(length)
	if err != nil {
		return "", err
	}
	now := time.Now()
	if token != "" {
		res := col.FindOneAndUpdate(ctx,
			bson.M{"tgToken": token, "used": false, "expiresAt": bson.M{"$gt": now}},
			bson.M{"$set": bson.M{
				"phone": phone, "telegramId": tgID, "code": code, "expiresAt": now.Add(ttl),
			}},
		)
		if res.Err() == nil {
			return code, nil
		}
	}
	// fallback: insert a phone-based code (web verifies by code-only fallback)
	_, err = col.InsertOne(ctx, otpDoc{
		Phone: phone, TelegramID: tgID, Code: code,
		ExpiresAt: now.Add(ttl), CreatedAt: now,
	})
	return code, err
}

func normalizePhone(p string) string {
	p = strings.TrimSpace(p)
	if !strings.HasPrefix(p, "+") {
		p = "+" + p
	}
	return p
}

func randDigits(n int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, x := range b {
		out[i] = digits[int(x)%len(digits)]
	}
	return string(out), nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
