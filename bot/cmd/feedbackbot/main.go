// Command feedbackbot ikkita Telegram botni bitta jarayonda ishga tushiradi:
//
//  1. Foydalanuvchi boti (FEEDBACK_BOT_TOKEN): foydalanuvchi /start bosadi,
//     telefon raqamini yuboradi, so'ng "Taklif" yoki "Shikoyat" ni tanlaydi va
//     matn / ovozli xabar / rasm yuboradi. Agar avval tur tanlamasdan xabar
//     yuborsa, xabar ostidan "Taklif/Shikoyat" so'rovi chiqadi. Qabul qilingach
//     foydalanuvchiga avtomatik tasdiq xabari boradi.
//
//  2. Admin boti (SUPPORT_BOT_TOKEN): yangi murojaatlar shu botga kimdan
//     kelganini ko'rsatib yuboriladi; admin "Javob berish" tugmasi orqali
//     javob yozadi va u foydalanuvchiga (1-bot orqali) yetkaziladi. Bu botdan
//     foydalanish faqat SUPPORT_ADMIN_PHONE raqami tasdiqlangach ochiladi.
//
// Har ikki bot bir xil Mongo bazadan foydalanadi ("bot_feedback" va
// "support_admins" kolleksiyalari).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ishchibormi/bot/internal/envfile"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type feedbackDoc struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	TelegramID  int64              `bson:"telegramId"`
	ChatID      int64              `bson:"chatId"`
	Phone       string             `bson:"phone"`
	Name        string             `bson:"name"`
	Username    string             `bson:"username,omitempty"`
	Type        string             `bson:"type"`        // suggestion|complaint
	ContentType string             `bson:"contentType"` // text|voice|photo
	Text        string             `bson:"text,omitempty"`
	FileID      string             `bson:"fileId,omitempty"`
	Status      string             `bson:"status"` // open|answered
	AdminReply  string             `bson:"adminReply,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt"`
	AnsweredAt  *time.Time         `bson:"answeredAt,omitempty"`
}

type adminDoc struct {
	Phone      string `bson:"phone"`
	ChatID     int64  `bson:"chatId"`
	TelegramID int64  `bson:"telegramId"`
}

// pendingContent — tur tanlanishidan oldin yuborilgan xabar mazmuni.
type pendingContent struct {
	contentType string
	text        string
	fileID      string
}

// session — foydalanuvchi boti uchun har bir chatning holati (xotirada).
type session struct {
	phone       string
	pendingType string          // tanlangan tur, xabar kutilmoqda
	pending     *pendingContent // xabar kelgan, tur kutilmoqda
}

type app struct {
	userBot          *tgbotapi.BotAPI // 1-bot (qabul)
	adminBot         *tgbotapi.BotAPI // 2-bot (javob)
	fb               *mongo.Collection
	admins           *mongo.Collection
	users            *mongo.Collection
	adminPhoneDigits string

	mu       sync.Mutex
	sessions map[int64]*session           // 1-bot chat holatlari
	replyTo  map[int64]primitive.ObjectID // 2-bot: admin chat -> javob berilayotgan murojaat
}

func main() {
	envfile.Load()
	userToken := os.Getenv("FEEDBACK_BOT_TOKEN")
	adminToken := os.Getenv("SUPPORT_BOT_TOKEN")
	if userToken == "" || adminToken == "" {
		log.Fatal("FEEDBACK_BOT_TOKEN and SUPPORT_BOT_TOKEN required")
	}
	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getenv("MONGO_DB", "ishchibormi")
	adminPhone := getenv("SUPPORT_ADMIN_PHONE", "+998900202535")

	ctx := context.Background()
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("mongo: %v", err)
	}
	defer func() { _ = mc.Disconnect(ctx) }()
	db := mc.Database(dbName)

	ub, err := tgbotapi.NewBotAPI(userToken)
	if err != nil {
		log.Fatalf("feedback bot: %v", err)
	}
	ab, err := tgbotapi.NewBotAPI(adminToken)
	if err != nil {
		log.Fatalf("support bot: %v", err)
	}
	log.Printf("feedback bot started: @%s | support bot: @%s", ub.Self.UserName, ab.Self.UserName)

	a := &app{
		userBot:          ub,
		adminBot:         ab,
		fb:               db.Collection("bot_feedback"),
		admins:           db.Collection("support_admins"),
		users:            db.Collection("users"),
		adminPhoneDigits: digitsOnly(adminPhone),
		sessions:         map[int64]*session{},
		replyTo:          map[int64]primitive.ObjectID{},
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); a.runUserBot(ctx) }()
	go func() { defer wg.Done(); a.runAdminBot(ctx) }()
	wg.Wait()
}

// ─────────────────────────── 1-bot: foydalanuvchi ──────────────────────────

func (a *app) runUserBot(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	for upd := range a.userBot.GetUpdatesChan(u) {
		switch {
		case upd.CallbackQuery != nil:
			a.userCallback(ctx, upd.CallbackQuery)
		case upd.Message != nil:
			a.userMessage(ctx, upd.Message)
		}
	}
}

func (a *app) userMessage(ctx context.Context, m *tgbotapi.Message) {
	chatID := m.Chat.ID
	s := a.sess(chatID)

	// /start — holatni tozalab, telefon raqamini so'raymiz (yoki tanlov menyusi).
	if m.IsCommand() && m.Command() == "start" {
		a.mu.Lock()
		s.pendingType = ""
		s.pending = nil
		phone := s.phone
		a.mu.Unlock()
		if phone == "" {
			if ph, ok := a.knownPhone(ctx, m.From.ID); ok {
				a.mu.Lock()
				s.phone = ph
				a.mu.Unlock()
				phone = ph
			}
		}
		if phone != "" {
			a.showMenu(chatID, "Assalomu alaykum! \"Ishchi Bormi\" taklif va shikoyatlar bo'limi.\n\nMurojaat turini tanlang yoki to'g'ridan-to'g'ri xabar (matn, ovozli xabar yoki rasm) yuboring.")
			return
		}
		msg := tgbotapi.NewMessage(chatID, "Assalomu alaykum! \"Ishchi Bormi\" taklif va shikoyatlar bo'limiga xush kelibsiz.\n\nDavom etish uchun telefon raqamingizni yuboring.")
		msg.ReplyMarkup = contactKb()
		_, _ = a.userBot.Send(msg)
		return
	}

	// Kontakt (telefon raqami).
	if m.Contact != nil {
		a.mu.Lock()
		s.phone = normalizePhone(m.Contact.PhoneNumber)
		a.mu.Unlock()
		ack := tgbotapi.NewMessage(chatID, "Rahmat! Telefon raqamingiz qabul qilindi.")
		ack.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, _ = a.userBot.Send(ack)
		a.showMenu(chatID, "Endi murojaat turini tanlang yoki to'g'ridan-to'g'ri xabar (matn, ovozli xabar yoki rasm) yuboring.")
		return
	}

	// Raqam berilmagan bo'lsa — avval uni so'raymiz.
	a.mu.Lock()
	phone := s.phone
	a.mu.Unlock()
	if phone == "" {
		msg := tgbotapi.NewMessage(chatID, "Iltimos, avval telefon raqamingizni yuboring.")
		msg.ReplyMarkup = contactKb()
		_, _ = a.userBot.Send(msg)
		return
	}

	// Xabar mazmunini ajratamiz.
	pc := extractContent(m)
	if pc == nil {
		a.showMenu(chatID, "Iltimos, taklif yoki shikoyatingizni matn, ovozli xabar yoki rasm ko'rinishida yuboring.")
		return
	}

	a.mu.Lock()
	pt := s.pendingType
	if pt != "" {
		s.pendingType = ""
		a.mu.Unlock()
		a.saveFeedback(ctx, chatID, m.From, phone, pt, pc)
		return
	}
	// Tur hali tanlanmagan — xabarni saqlab, turini so'raymiz.
	s.pending = pc
	a.mu.Unlock()
	msg := tgbotapi.NewMessage(chatID, "Bu taklifmi yoki shikoyatmi? Iltimos, tanlang:")
	msg.ReplyMarkup = typeMenu()
	_, _ = a.userBot.Send(msg)
}

func (a *app) userCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	_, _ = a.userBot.Request(tgbotapi.NewCallback(cq.ID, ""))
	if cq.Message == nil || !strings.HasPrefix(cq.Data, "type:") {
		return
	}
	typ := strings.TrimPrefix(cq.Data, "type:")
	if typ != "suggestion" && typ != "complaint" {
		return
	}
	chatID := cq.Message.Chat.ID
	s := a.sess(chatID)

	a.mu.Lock()
	phone := s.phone
	a.mu.Unlock()
	if phone == "" {
		msg := tgbotapi.NewMessage(chatID, "Iltimos, avval telefon raqamingizni yuboring.")
		msg.ReplyMarkup = contactKb()
		_, _ = a.userBot.Send(msg)
		return
	}

	a.mu.Lock()
	pc := s.pending
	if pc != nil {
		s.pending = nil
		a.mu.Unlock()
		a.saveFeedback(ctx, chatID, cq.From, phone, typ, pc)
		return
	}
	// Xabar hali yuborilmagan — turini eslab, xabarni so'raymiz.
	s.pendingType = typ
	a.mu.Unlock()
	a.send(a.userBot, chatID, fmt.Sprintf("Iltimos, %singizni yozing (matn, ovozli xabar yoki rasm).", typeLabel(typ)))
}

func (a *app) saveFeedback(ctx context.Context, chatID int64, from *tgbotapi.User, phone, typ string, pc *pendingContent) {
	name := ""
	username := ""
	var tgID int64
	if from != nil {
		name = strings.TrimSpace(from.FirstName + " " + from.LastName)
		username = from.UserName
		tgID = from.ID
	}
	doc := feedbackDoc{
		TelegramID:  tgID,
		ChatID:      chatID,
		Phone:       phone,
		Name:        name,
		Username:    username,
		Type:        typ,
		ContentType: pc.contentType,
		Text:        pc.text,
		FileID:      pc.fileID,
		Status:      "open",
		CreatedAt:   time.Now(),
	}
	res, err := a.fb.InsertOne(ctx, doc)
	if err != nil {
		a.send(a.userBot, chatID, "Xatolik yuz berdi. Iltimos, keyinroq qayta urinib ko'ring.")
		return
	}
	if id, ok := res.InsertedID.(primitive.ObjectID); ok {
		doc.ID = id
	}
	a.send(a.userBot, chatID, "✅ "+typeLabelCap(typ)+"ingiz qabul qilindi. Ko'rib chiqib, tez orada javob beramiz. Rahmat!")
	a.notifyAdmins(ctx, doc)
}

// notifyAdmins yangi murojaatni barcha tasdiqlangan adminlarga yuboradi.
func (a *app) notifyAdmins(ctx context.Context, doc feedbackDoc) {
	admins := a.adminChats(ctx)
	if len(admins) == 0 {
		return
	}
	mediaURL := a.mediaURL(doc.FileID)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✍️ Javob berish", "reply:"+doc.ID.Hex())),
	)
	for _, chatID := range admins {
		a.sendMedia(a.adminBot, chatID, doc.ContentType, mediaURL)
		msg := tgbotapi.NewMessage(chatID, a.adminText(doc))
		msg.ReplyMarkup = kb
		_, _ = a.adminBot.Send(msg)
	}
}

// ─────────────────────────── 2-bot: admin/javob ────────────────────────────

func (a *app) runAdminBot(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	for upd := range a.adminBot.GetUpdatesChan(u) {
		switch {
		case upd.CallbackQuery != nil:
			a.adminCallback(ctx, upd.CallbackQuery)
		case upd.Message != nil:
			a.adminMessage(ctx, upd.Message)
		}
	}
}

func (a *app) adminMessage(ctx context.Context, m *tgbotapi.Message) {
	chatID := m.Chat.ID

	// Kontakt — ruxsatni tekshirish.
	if m.Contact != nil {
		if digitsOnly(m.Contact.PhoneNumber) == a.adminPhoneDigits {
			_, _ = a.admins.UpdateOne(ctx,
				bson.M{"chatId": chatID},
				bson.M{"$set": bson.M{
					"chatId": chatID, "phone": normalizePhone(m.Contact.PhoneNumber),
					"telegramId": m.From.ID, "createdAt": time.Now(),
				}},
				options.Update().SetUpsert(true))
			ack := tgbotapi.NewMessage(chatID, "✅ Raqamingiz tasdiqlandi! Endi foydalanuvchilarning taklif va shikoyatlariga javob bera olasiz.\n\nOchiq murojaatlar ro'yxati: /list")
			ack.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = a.adminBot.Send(ack)
			a.sendOpenList(ctx, chatID)
		} else {
			ack := tgbotapi.NewMessage(chatID, "❌ Bu raqamga ruxsat berilmagan.")
			ack.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = a.adminBot.Send(ack)
		}
		return
	}

	// Ruxsat yo'q bo'lsa — raqam so'raymiz.
	if !a.isAdmin(ctx, chatID) {
		msg := tgbotapi.NewMessage(chatID, "Bu bot orqali javob berish uchun ruxsat kerak.\n\nIltimos, telefon raqamingizni yuboring.")
		msg.ReplyMarkup = contactKb()
		_, _ = a.adminBot.Send(msg)
		return
	}

	if m.IsCommand() {
		switch m.Command() {
		case "start":
			a.send(a.adminBot, chatID, "Assalomu alaykum! Siz admin sifatida tasdiqlangansiz.\n\nOchiq murojaatlar ro'yxati: /list")
		case "list":
			a.sendOpenList(ctx, chatID)
		default:
			a.send(a.adminBot, chatID, "Buyruqlar: /list — ochiq murojaatlar.")
		}
		return
	}

	// Javob kutilyaptimi?
	a.mu.Lock()
	fid, waiting := a.replyTo[chatID]
	if waiting {
		delete(a.replyTo, chatID)
	}
	a.mu.Unlock()
	if waiting && strings.TrimSpace(m.Text) != "" {
		a.deliverReply(ctx, chatID, fid, m.Text)
		return
	}
	a.send(a.adminBot, chatID, "Javob berish uchun murojaat ostidagi \"✍️ Javob berish\" tugmasini bosing yoki /list ni yuboring.")
}

func (a *app) adminCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	_, _ = a.adminBot.Request(tgbotapi.NewCallback(cq.ID, ""))
	if cq.Message == nil || !strings.HasPrefix(cq.Data, "reply:") {
		return
	}
	chatID := cq.Message.Chat.ID
	if !a.isAdmin(ctx, chatID) {
		a.send(a.adminBot, chatID, "Ruxsat yo'q. Telefon raqamingizni yuboring.")
		return
	}
	id, err := primitive.ObjectIDFromHex(strings.TrimPrefix(cq.Data, "reply:"))
	if err != nil {
		return
	}
	a.mu.Lock()
	a.replyTo[chatID] = id
	a.mu.Unlock()
	a.send(a.adminBot, chatID, "✍️ Javobingizni yozing:")
}

// deliverReply admin javobini foydalanuvchiga (1-bot orqali) yetkazadi.
func (a *app) deliverReply(ctx context.Context, adminChat int64, fid primitive.ObjectID, reply string) {
	var doc feedbackDoc
	if err := a.fb.FindOne(ctx, bson.M{"_id": fid}).Decode(&doc); err != nil {
		a.send(a.adminBot, adminChat, "Murojaat topilmadi.")
		return
	}
	a.send(a.userBot, doc.ChatID, fmt.Sprintf("📩 %singizga javob keldi:\n\n%s", typeLabel(doc.Type), reply))
	now := time.Now()
	_, _ = a.fb.UpdateOne(ctx, bson.M{"_id": fid},
		bson.M{"$set": bson.M{"status": "answered", "adminReply": reply, "answeredAt": now}})
	a.send(a.adminBot, adminChat, "✅ Javob foydalanuvchiga yuborildi.")
}

func (a *app) sendOpenList(ctx context.Context, chatID int64) {
	cur, err := a.fb.Find(ctx, bson.M{"status": "open"},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetLimit(20))
	if err != nil {
		a.send(a.adminBot, chatID, "Ro'yxatni olishda xatolik.")
		return
	}
	defer cur.Close(ctx)
	var docs []feedbackDoc
	if err := cur.All(ctx, &docs); err != nil || len(docs) == 0 {
		a.send(a.adminBot, chatID, "Hozircha ochiq murojaatlar yo'q.")
		return
	}
	for _, d := range docs {
		a.sendMedia(a.adminBot, chatID, d.ContentType, a.mediaURL(d.FileID))
		msg := tgbotapi.NewMessage(chatID, a.adminText(d))
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✍️ Javob berish", "reply:"+d.ID.Hex())),
		)
		_, _ = a.adminBot.Send(msg)
	}
}

// ─────────────────────────────── yordamchilar ──────────────────────────────

func (a *app) sess(chatID int64) *session {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.sessions[chatID]
	if s == nil {
		s = &session{}
		a.sessions[chatID] = s
	}
	return s
}

func (a *app) isAdmin(ctx context.Context, chatID int64) bool {
	n, _ := a.admins.CountDocuments(ctx, bson.M{"chatId": chatID})
	return n > 0
}

func (a *app) adminChats(ctx context.Context) []int64 {
	cur, err := a.admins.Find(ctx, bson.M{})
	if err != nil {
		return nil
	}
	defer cur.Close(ctx)
	var docs []adminDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil
	}
	out := make([]int64, 0, len(docs))
	for _, d := range docs {
		if d.ChatID != 0 {
			out = append(out, d.ChatID)
		}
	}
	return out
}

// knownPhone — foydalanuvchining telefon raqamini asosiy "users" kolleksiyasidan
// (agar u ilovada ro'yxatdan o'tgan bo'lsa) topadi.
func (a *app) knownPhone(ctx context.Context, tgID int64) (string, bool) {
	if tgID == 0 {
		return "", false
	}
	var u struct {
		Phone string `bson:"phone"`
	}
	err := a.users.FindOne(ctx,
		bson.M{"telegramId": tgID, "phone": bson.M{"$ne": ""}},
		options.FindOne().SetProjection(bson.M{"phone": 1}),
	).Decode(&u)
	if err != nil || u.Phone == "" {
		return "", false
	}
	return u.Phone, true
}

// mediaURL fayl_id bo'yicha 1-bot orqali to'g'ridan-to'g'ri yuklab olish
// havolasini qaytaradi (2-bot uni URL orqali qayta yuboradi — file_id'lar
// botlararo mos kelmaydi).
func (a *app) mediaURL(fileID string) string {
	if fileID == "" {
		return ""
	}
	if url, err := a.userBot.GetFileDirectURL(fileID); err == nil {
		return url
	}
	return ""
}

func (a *app) sendMedia(bot *tgbotapi.BotAPI, chatID int64, contentType, url string) {
	if url == "" {
		return
	}
	switch contentType {
	case "photo":
		_, _ = bot.Send(tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(url)))
	case "voice":
		_, _ = bot.Send(tgbotapi.NewVoice(chatID, tgbotapi.FileURL(url)))
	}
}

func (a *app) adminText(d feedbackDoc) string {
	body := strings.TrimSpace(d.Text)
	switch d.ContentType {
	case "voice":
		if body == "" {
			body = "🎤 (ovozli xabar — yuqorida)"
		} else {
			body = "🎤 " + body
		}
	case "photo":
		if body == "" {
			body = "🖼 (rasm — yuqorida)"
		} else {
			body = "🖼 " + body
		}
	}
	uname := ""
	if d.Username != "" {
		uname = " (@" + d.Username + ")"
	}
	return fmt.Sprintf("🆕 Yangi %s\n👤 %s%s\n📞 %s\n🕒 %s\n\n%s",
		typeLabelCap(d.Type), d.Name, uname, d.Phone, d.CreatedAt.Format("2006-01-02 15:04"), body)
}

func (a *app) showMenu(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = typeMenu()
	_, _ = a.userBot.Send(msg)
}

func (a *app) send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	_, _ = bot.Send(tgbotapi.NewMessage(chatID, text))
}

func extractContent(m *tgbotapi.Message) *pendingContent {
	switch {
	case len(m.Photo) > 0:
		return &pendingContent{contentType: "photo", text: m.Caption, fileID: m.Photo[len(m.Photo)-1].FileID}
	case m.Voice != nil:
		return &pendingContent{contentType: "voice", text: m.Caption, fileID: m.Voice.FileID}
	case m.Audio != nil:
		return &pendingContent{contentType: "voice", text: m.Caption, fileID: m.Audio.FileID}
	case strings.TrimSpace(m.Text) != "":
		return &pendingContent{contentType: "text", text: m.Text}
	}
	return nil
}

func contactKb() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.KeyboardButton{Text: "📞 Telefon raqamni yuborish", RequestContact: true}),
	)
	kb.OneTimeKeyboard = true
	kb.ResizeKeyboard = true
	return kb
}

func typeMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💡 Taklif", "type:suggestion"),
			tgbotapi.NewInlineKeyboardButtonData("⚠️ Shikoyat", "type:complaint"),
		),
	)
}

func typeLabel(t string) string {
	if t == "complaint" {
		return "shikoyat"
	}
	return "taklif"
}

func typeLabelCap(t string) string {
	if t == "complaint" {
		return "Shikoyat"
	}
	return "Taklif"
}

func normalizePhone(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	if !strings.HasPrefix(p, "+") {
		p = "+" + p
	}
	return p
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
