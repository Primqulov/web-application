package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ishchibormi/backend/config"
	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/pkg/db"
	"github.com/ishchibormi/backend/pkg/envfile"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var firstNames = []string{
	"Alisher", "Sardor", "Bekzod", "Jasur", "Davron", "Sherzod", "Otabek",
	"Akmal", "Rustam", "Shokhrux", "Aziz", "Murod", "Nodir", "Ulug'bek",
	"Doniyor", "Farrukh",
}
var firstNamesF = []string{
	"Malika", "Madina", "Nilufar", "Zilola", "Mavluda", "Sevara", "Dilshoda",
	"Munisa", "Shahnoza", "Gulchehra",
}
var lastNames = []string{
	"Rustamov", "Yusupov", "Karimov", "Akhmedov", "Tursunov", "Olimov",
	"Mahmudov", "Saidov", "Bekmirzayev", "Khojayev", "Norqulov", "Rasulov",
	"Toshpo'latov", "Egamberdiyev", "Mirzayev", "Soliyev",
}

var regions = []struct{ Region, District string }{
	{"Toshkent", "Chilonzor"}, {"Toshkent", "Yunusobod"}, {"Toshkent", "Sergeli"},
	{"Toshkent", "Mirzo Ulug'bek"}, {"Toshkent", "Yashnobod"},
	{"Samarqand", "Bag'ishamol"}, {"Samarqand", "Urgut"},
	{"Buxoro", "Olot"}, {"Buxoro", "Kogon"},
	{"Farg'ona", "Marg'ilon"}, {"Farg'ona", "Quvasoy"},
	{"Namangan", "Pop"}, {"Namangan", "Chust"},
	{"Andijon", "Asaka"}, {"Qashqadaryo", "Qarshi"},
	{"Surxondaryo", "Termiz"},
}

type catSeed struct {
	Name, Slug, Icon string
	SystemDefault    bool
}

var categories = []catSeed{
	{"Mebel tashish", "mebel-tashish", "📦", true},
	{"Santexnika", "santexnika", "🔧", true},
	{"Tozalash", "tozalash", "🧹", true},
	{"Elektr toki", "elektr-toki", "⚡", true},
	{"Kuryerlik", "kuryerlik", "🛵", true},
	{"Qurilish", "qurilish", "🏗️", true},
	{"Yuk tashish", "yuk-tashish", "🚚", true},
	{"Ustachilik", "ustachilik", "🛠️", true},
	{"Bog'dorchilik", "bogdorchilik", "🌳", true},
	{"Devor bo'yash", "devor-boyash", "🎨", true},
	{"Deraza yuvish", "deraza-yuvish", "🪟", true},
	{"Ko'chirish", "kochirish", "🏠", true},
	{"Hovli tozalash", "hovli-tozalash", "🍂", true},
	{"Ovqat tayyorlash", "ovqat-tayyorlash", "🍲", true},
	{"Bolalar parvarishi", "bolalar-parvarishi", "👶", true},
	{"Boshqalar", "boshqalar", "📌", true},
}

type elonSeed struct {
	Title, Description string
	CategoryName       string
	WorkersNeeded      int
	PricingType        string // per_worker|total|negotiable
	PriceAmount        int64
	Status             string // draft|recruiting|filled|in_progress|completed|cancelled
}

var elonSeeds = []elonSeed{
	{"Mebel tashishga ishchilar kerak", "3 xonali kvartiradan yangi uyga mebel tashish.", "Mebel tashish", 3, "per_worker", 150000, "recruiting"},
	{"Hovli tozalash", "Katta hovlini yig'ishtirish va xashagini chiqarish.", "Hovli tozalash", 2, "per_worker", 100000, "recruiting"},
	{"Santexnika xizmati", "Hammomdagi kran va lavabo ta'mirlash.", "Santexnika", 1, "total", 250000, "recruiting"},
	{"Devorlarni bo'yash", "Yangi yotoq xonasi devorlarini bo'yash.", "Devor bo'yash", 2, "per_worker", 200000, "filled"},
	{"Kuryer kerak (1 kunlik)", "Shahar bo'ylab paket yetkazib berish.", "Kuryerlik", 1, "per_worker", 120000, "recruiting"},
	{"Deraza yuvish", "Ofisdagi 20 ta derazani tozalash.", "Deraza yuvish", 2, "per_worker", 80000, "completed"},
	{"Yuk mashina yordamida ko'chirish", "Buyumlarni boshqa shaharga olib borish.", "Ko'chirish", 4, "total", 1200000, "completed"},
	{"Elektr tarmog'ini tortish", "Yangi qurilgan uyda elektr o'tkazish.", "Elektr toki", 2, "per_worker", 350000, "in_progress"},
	{"Bog'dorchilik", "Bog'dagi olma va o'rik daraxtlarini parvarish qilish.", "Bog'dorchilik", 1, "total", 180000, "recruiting"},
	{"Ovqat tayyorlash (to'y uchun)", "100 kishi uchun palov tayyorlash.", "Ovqat tayyorlash", 2, "per_worker", 500000, "recruiting"},
	{"Bolalar parvarishi (kunduzi)", "2 yoshli bola bilan ishlash kerak.", "Bolalar parvarishi", 1, "total", 200000, "draft"},
	{"Qurilish ishchilari (3 kun)", "Pol va devor qoplash ishlari.", "Qurilish", 5, "per_worker", 300000, "recruiting"},
	{"Tozalash (do'kon)", "Yangi ochilgan do'konni umumiy tozalash.", "Tozalash", 3, "per_worker", 90000, "completed"},
	{"Uy ko'chirish", "1 xonali uyni Sergeli tumaniga ko'chirish.", "Ko'chirish", 2, "negotiable", 0, "cancelled"},
	{"Ustachilik xizmati", "Yog'och eshikni tuzatish.", "Ustachilik", 1, "total", 150000, "recruiting"},
	{"Sayqal tozalash (ofis)", "Ofisdagi har kungi tozalash.", "Tozalash", 1, "per_worker", 70000, "filled"},
	{"Mebel yig'ish", "IKEA tipidagi mebellarni yig'ish.", "Ustachilik", 2, "per_worker", 120000, "completed"},
}

func main() {
	envfile.Load()
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	mdb, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongo: %v", err)
	}
	if err := db.EnsureIndexes(ctx, mdb); err != nil {
		log.Printf("indexes: %v", err)
	}

	// idempotency: clear domain collections
	cols := []string{"users", "categories", "elons", "applications", "reviews",
		"conversations", "messages", "notifications", "reports", "finance_entries",
		"admins", "admin_audit", "otp_codes"}
	for _, c := range cols {
		if _, err := mdb.Collection(c).DeleteMany(ctx, bson.M{}); err != nil {
			log.Printf("clear %s: %v", c, err)
		}
	}
	rand.Seed(time.Now().UnixNano())
	now := time.Now()

	// ---- categories ----
	catIDs := map[string]primitive.ObjectID{}
	for _, c := range categories {
		oid := primitive.NewObjectID()
		_, err := mdb.Collection("categories").InsertOne(ctx, models.Category{
			ID: oid, Name: c.Name, Slug: c.Slug, Icon: c.Icon,
			IsSystemDefault: c.SystemDefault, IsActive: true,
			UsageCount: rand.Intn(20), CreatedAt: now,
		})
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			log.Printf("category %s: %v", c.Name, err)
			continue
		}
		catIDs[c.Name] = oid
	}
	fmt.Printf("categories: %d\n", len(catIDs))

	// ---- users ----
	users := []models.User{}
	for i := 0; i < 18; i++ {
		var fn string
		if i%4 == 0 {
			fn = pick(firstNamesF)
		} else {
			fn = pick(firstNames)
		}
		ln := pick(lastNames)
		reg := regions[i%len(regions)]
		u := models.User{
			ID:                  primitive.NewObjectID(),
			TelegramID:          int64(1000000 + i),
			Phone:               fmt.Sprintf("+998 9%d %03d %02d %02d", 0+(i%5), rand.Intn(1000), rand.Intn(100), rand.Intn(100)),
			FirstName:           fn,
			LastName:            ln,
			Region:              reg.Region,
			District:            reg.District,
			AvatarURL:           "",
			Bio:                 "Tajribali ishchi, Toshkent va atrofdagi viloyatlarda ish bajaraman.",
			Skills:              []string{pick([]string{"Mebel tashish", "Tozalash", "Qurilish", "Santexnika", "Elektr", "Bog'dorchilik"})},
			Rating:              roundTo1(3.5 + rand.Float64()*1.5),
			ReviewsCount:        2 + rand.Intn(15),
			CompletedJobsCount:  1 + rand.Intn(20),
			IsPhoneVerified:     true,
			IsPremium:           i == 0,
			OnboardingCompleted: true,
			LangPref:            "latin",
			ThemePref:           "light",
			CreatedAt:           now.Add(-time.Duration(rand.Intn(60)) * 24 * time.Hour),
			UpdatedAt:           now,
		}
		users = append(users, u)
	}
	{
		docs := make([]any, len(users))
		for i, u := range users {
			docs[i] = u
		}
		if _, err := mdb.Collection("users").InsertMany(ctx, docs); err != nil {
			log.Printf("users: %v", err)
		}
	}
	fmt.Printf("users: %d\n", len(users))

	// ---- elons ----
	elons := []models.Elon{}
	for i, e := range elonSeeds {
		owner := users[i%len(users)]
		cid, ok := catIDs[e.CategoryName]
		if !ok {
			cid = catIDs["Boshqalar"]
		}
		pType := e.PricingType
		per := e.PriceAmount
		total := per * int64(e.WorkersNeeded)
		if pType == "total" {
			total = e.PriceAmount
			if e.WorkersNeeded > 0 {
				per = total / int64(e.WorkersNeeded)
			}
		}
		if pType == "negotiable" {
			per, total = 0, 0
		}
		var publishedAt *time.Time
		if e.Status != "draft" {
			t := now.Add(-time.Duration(rand.Intn(40)) * time.Hour)
			publishedAt = &t
		}
		reg := regions[i%len(regions)]
		el := models.Elon{
			ID:              primitive.NewObjectID(),
			OwnerID:         owner.ID,
			Title:           e.Title,
			CategoryID:      cid,
			CategoryName:    e.CategoryName,
			Description:     e.Description,
			LocationText:    reg.Region + ", " + reg.District,
			LocationURL:     "https://maps.google.com/?q=" + reg.Region,
			Region:          reg.Region,
			District:        reg.District,
			WorkersNeeded:   e.WorkersNeeded,
			PricingType:     pType,
			PriceAmount:     total,
			PerWorkerAmount: per,
			StartDate:       now.Format("2006-01-02"),
			WorkTimeFrom:    "09:00",
			WorkTimeTo:      "18:00",
			ContactPhone:    owner.Phone,
			Status:          e.Status,
			AcceptedCount:   0,
			PublishedAt:     publishedAt,
			CreatedAt:       now.Add(-time.Duration(rand.Intn(60)) * time.Hour),
			UpdatedAt:       now,
			OwnerName:       owner.FirstName + " " + owner.LastName,
			OwnerRating:     owner.Rating,
		}
		elons = append(elons, el)
	}
	{
		docs := make([]any, len(elons))
		for i, e := range elons {
			docs[i] = e
		}
		if _, err := mdb.Collection("elons").InsertMany(ctx, docs); err != nil {
			log.Printf("elons: %v", err)
		}
		// usageCount already initialized above; nothing else needed.
	}
	fmt.Printf("elons: %d\n", len(elons))

	// ---- applications ----
	apps := []models.Application{}
	statuses := []string{"pending", "accepted", "rejected", "cancelled", "completed", "completed", "pending", "accepted"}
	for i := 0; i < 20; i++ {
		e := elons[i%len(elons)]
		worker := users[(i+3)%len(users)]
		if worker.ID == e.OwnerID {
			worker = users[(i+4)%len(users)]
		}
		st := statuses[i%len(statuses)]
		a := models.Application{
			ID:           primitive.NewObjectID(),
			ElonID:       e.ID,
			ElonTitle:    e.Title,
			WorkerID:     worker.ID,
			EmployerID:   e.OwnerID,
			WorkerPhone:  worker.Phone,
			Amount:       e.PerWorkerAmount,
			IsNegotiable: e.PricingType == "negotiable",
			Status:       st,
			AppliedAt:    now.Add(-time.Duration(rand.Intn(120)) * time.Hour),
		}
		switch st {
		case "completed":
			c := a.AppliedAt.Add(time.Duration(12+rand.Intn(48)) * time.Hour)
			a.DecidedAt = &c
			a.CompletedAt = &c
			a.EmployerConfirmedDone = true
			a.WorkerConfirmedDone = true
		case "accepted":
			d := a.AppliedAt.Add(time.Duration(2+rand.Intn(20)) * time.Hour)
			a.DecidedAt = &d
		case "rejected", "cancelled":
			d := a.AppliedAt.Add(time.Duration(1+rand.Intn(8)) * time.Hour)
			a.DecidedAt = &d
			if st == "cancelled" {
				a.CancelledBy = "worker"
			}
		}
		apps = append(apps, a)
	}
	{
		docs := make([]any, len(apps))
		for i, a := range apps {
			docs[i] = a
		}
		if _, err := mdb.Collection("applications").InsertMany(ctx, docs); err != nil {
			log.Printf("applications: %v", err)
		}
	}
	fmt.Printf("applications: %d\n", len(apps))

	// ---- reviews (for completed apps) ----
	reviews := []models.Review{}
	for _, a := range apps {
		if a.Status != "completed" {
			continue
		}
		reviews = append(reviews, models.Review{
			ID:            primitive.NewObjectID(),
			ApplicationID: a.ID,
			ElonID:        a.ElonID,
			FromUserID:    a.EmployerID,
			ToUserID:      a.WorkerID,
			Direction:     "employer_to_worker",
			Rating:        4 + rand.Intn(2),
			Comment:       "Yaxshi ish bajardi, rahmat!",
			CreatedAt:     time.Now(),
		})
		reviews = append(reviews, models.Review{
			ID:            primitive.NewObjectID(),
			ApplicationID: a.ID,
			ElonID:        a.ElonID,
			FromUserID:    a.WorkerID,
			ToUserID:      a.EmployerID,
			Direction:     "worker_to_employer",
			Rating:        4 + rand.Intn(2),
			Comment:       "Yaxshi buyurtmachi, vaqtida to'ladi.",
			CreatedAt:     time.Now(),
		})
	}
	// pad to ≥15 with extra reviews on random users
	for len(reviews) < 16 {
		from := users[rand.Intn(len(users))]
		to := users[rand.Intn(len(users))]
		if from.ID == to.ID {
			continue
		}
		reviews = append(reviews, models.Review{
			ID:            primitive.NewObjectID(),
			ApplicationID: primitive.NewObjectID(),
			ElonID:        elons[rand.Intn(len(elons))].ID,
			FromUserID:    from.ID,
			ToUserID:      to.ID,
			Direction:     "employer_to_worker",
			Rating:        3 + rand.Intn(3),
			Comment:       "Yaxshi.",
			CreatedAt:     time.Now(),
		})
	}
	{
		docs := make([]any, len(reviews))
		for i, r := range reviews {
			docs[i] = r
		}
		if _, err := mdb.Collection("reviews").InsertMany(ctx, docs); err != nil {
			log.Printf("reviews: %v", err)
		}
	}
	fmt.Printf("reviews: %d\n", len(reviews))

	// ---- conversations + messages ----
	conversations := []models.Conversation{}
	messages := []models.Message{}
	for i := 0; i < 16; i++ {
		a := users[i%len(users)]
		b := users[(i+5)%len(users)]
		if a.ID == b.ID {
			b = users[(i+6)%len(users)]
		}
		cid := primitive.NewObjectID()
		unread := map[string]int{a.ID.Hex(): rand.Intn(3), b.ID.Hex(): rand.Intn(3)}
		messages = append(messages, models.Message{
			ID: primitive.NewObjectID(), ConversationID: cid, SenderID: a.ID,
			Text: "Assalomu alaykum, ish haqida gaplashsak bo'ladimi?",
			CreatedAt: now.Add(-time.Duration(rand.Intn(120)) * time.Hour),
		})
		messages = append(messages, models.Message{
			ID: primitive.NewObjectID(), ConversationID: cid, SenderID: b.ID,
			Text: "Albatta, qachon kelishingiz mumkin?",
			CreatedAt: now.Add(-time.Duration(rand.Intn(60)) * time.Hour),
		})
		conversations = append(conversations, models.Conversation{
			ID: cid,
			ParticipantIDs:  []primitive.ObjectID{a.ID, b.ID},
			LastMessageText: "Albatta, qachon kelishingiz mumkin?",
			LastMessageAt:   now,
			LastSenderID:    b.ID,
			Unread:          unread,
			CreatedAt:       now.Add(-72 * time.Hour),
		})
	}
	{
		docs := make([]any, len(conversations))
		for i, c := range conversations {
			docs[i] = c
		}
		if _, err := mdb.Collection("conversations").InsertMany(ctx, docs); err != nil {
			log.Printf("conversations: %v", err)
		}
		md := make([]any, len(messages))
		for i, m := range messages {
			md[i] = m
		}
		if _, err := mdb.Collection("messages").InsertMany(ctx, md); err != nil {
			log.Printf("messages: %v", err)
		}
	}
	fmt.Printf("conversations: %d, messages: %d\n", len(conversations), len(messages))

	// ---- notifications ----
	notifs := []any{}
	types := []string{"new_application", "application_accepted", "application_rejected",
		"new_message", "job_completed_request", "job_completed", "new_review", "system"}
	for i := 0; i < 20; i++ {
		t := types[i%len(types)]
		u := users[i%len(users)]
		notifs = append(notifs, models.Notification{
			UserID: u.ID, Type: t,
			Title:     titleFor(t),
			Body:      "Tafsilotlar uchun bosing.",
			IsRead:    i%3 == 0,
			CreatedAt: now.Add(-time.Duration(rand.Intn(120)) * time.Hour),
		})
	}
	if _, err := mdb.Collection("notifications").InsertMany(ctx, notifs); err != nil {
		log.Printf("notifs: %v", err)
	}
	fmt.Printf("notifications: %d\n", len(notifs))

	// ---- reports ----
	reports := []any{}
	reasons := []string{"Spam", "Yolg'on ma'lumot", "Haqorat", "Mos kelmaydigan kontent", "Boshqa"}
	statusesR := []string{"open", "open", "open", "resolved", "dismissed"}
	for i := 0; i < 16; i++ {
		r := users[i%len(users)]
		t := users[(i+2)%len(users)]
		reports = append(reports, models.Report{
			ReporterID: r.ID, TargetType: "user", TargetID: t.ID,
			Reason: reasons[i%len(reasons)], Description: "Sinov shikoyati.",
			Status: statusesR[i%len(statusesR)], CreatedAt: now.Add(-time.Duration(rand.Intn(240)) * time.Hour),
		})
	}
	if _, err := mdb.Collection("reports").InsertMany(ctx, reports); err != nil {
		log.Printf("reports: %v", err)
	}
	fmt.Printf("reports: %d\n", len(reports))

	// ---- finance entries (from completed + cancelled apps) ----
	finance := []any{}
	for _, a := range apps {
		switch a.Status {
		case "completed":
			finance = append(finance, models.FinanceEntry{
				UserID: a.WorkerID, Role: "worker", ApplicationID: a.ID, ElonID: a.ElonID,
				ElonTitle: a.ElonTitle, CounterpartyID: a.EmployerID,
				Type: "earned", Status: "completed",
				Amount: a.Amount, IsNegotiable: a.IsNegotiable, CreatedAt: now,
			})
			finance = append(finance, models.FinanceEntry{
				UserID: a.EmployerID, Role: "employer", ApplicationID: a.ID, ElonID: a.ElonID,
				ElonTitle: a.ElonTitle, CounterpartyID: a.WorkerID,
				Type: "spent", Status: "completed",
				Amount: a.Amount, IsNegotiable: a.IsNegotiable, CreatedAt: now,
			})
		case "cancelled":
			finance = append(finance, models.FinanceEntry{
				UserID: a.WorkerID, Role: "worker", ApplicationID: a.ID, ElonID: a.ElonID,
				ElonTitle: a.ElonTitle, CounterpartyID: a.EmployerID,
				Type: "earned", Status: "cancelled", IsNegotiable: a.IsNegotiable, CreatedAt: now,
			})
			finance = append(finance, models.FinanceEntry{
				UserID: a.EmployerID, Role: "employer", ApplicationID: a.ID, ElonID: a.ElonID,
				ElonTitle: a.ElonTitle, CounterpartyID: a.WorkerID,
				Type: "spent", Status: "cancelled", IsNegotiable: a.IsNegotiable, CreatedAt: now,
			})
		}
	}
	for len(finance) < 16 {
		u := users[rand.Intn(len(users))]
		e := elons[rand.Intn(len(elons))]
		finance = append(finance, models.FinanceEntry{
			UserID: u.ID, Role: "worker", ApplicationID: primitive.NewObjectID(),
			ElonID: e.ID, ElonTitle: e.Title, CounterpartyID: e.OwnerID,
			Type: "earned", Status: "completed",
			Amount: int64(50000 + rand.Intn(500000)), IsNegotiable: false, CreatedAt: now,
		})
	}
	if _, err := mdb.Collection("finance_entries").InsertMany(ctx, finance); err != nil {
		log.Printf("finance: %v", err)
	}
	fmt.Printf("finance_entries: %d\n", len(finance))

	// ---- admins ----
	admins := []models.Admin{}
	pw, _ := bcrypt.GenerateFromPassword([]byte(cfg.AdminSeedPass), bcrypt.DefaultCost)
	admins = append(admins, models.Admin{
		Username: cfg.AdminSeedUser, PasswordHash: string(pw),
		Role: "superadmin", IsActive: true, CreatedAt: now,
	})
	pw2, _ := bcrypt.GenerateFromPassword([]byte("Moder123!"), bcrypt.DefaultCost)
	admins = append(admins, models.Admin{
		Username: "moderator", PasswordHash: string(pw2),
		Role: "moderator", IsActive: true, CreatedAt: now,
	})
	for _, a := range admins {
		_, err := mdb.Collection("admins").InsertOne(ctx, a)
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			log.Printf("admin %s: %v", a.Username, err)
		}
	}
	fmt.Printf("admins: %d\n", len(admins))

	// recompute ratings from reviews to make data coherent
	if err := recomputeRatings(ctx, mdb); err != nil {
		log.Printf("recompute: %v", err)
	}

	fmt.Println("seed done")
}

func recomputeRatings(ctx context.Context, mdb *mongo.Database) error {
	cur, err := mdb.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var u models.User
		if err := cur.Decode(&u); err != nil {
			continue
		}
		rcur, err := mdb.Collection("reviews").Find(ctx, bson.M{"toUserId": u.ID})
		if err != nil {
			continue
		}
		var sum, n int
		for rcur.Next(ctx) {
			var r models.Review
			if err := rcur.Decode(&r); err == nil {
				sum += r.Rating
				n++
			}
		}
		rcur.Close(ctx)
		avg := 0.0
		if n > 0 {
			avg = float64(sum) / float64(n)
			avg = float64(int(avg*10+0.5)) / 10
		}
		_, _ = mdb.Collection("users").UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": bson.M{"rating": avg, "reviewsCount": n}})
	}
	return nil
}

func pick(s []string) string { return s[rand.Intn(len(s))] }

func roundTo1(f float64) float64 { return float64(int(f*10+0.5)) / 10 }

func titleFor(t string) string {
	switch t {
	case "new_application":
		return "Yangi ariza"
	case "application_accepted":
		return "Arizangiz qabul qilindi"
	case "application_rejected":
		return "Arizangiz rad etildi"
	case "new_message":
		return "Yangi xabar"
	case "job_completed_request":
		return "Tasdiqlash so'rovi"
	case "job_completed":
		return "Ish yakunlandi"
	case "new_review":
		return "Yangi baho"
	case "system":
		return "Tizim xabari"
	}
	return "Bildirishnoma"
}
