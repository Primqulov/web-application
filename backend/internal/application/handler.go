package application

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/internal/notification"
	"github.com/ishchibormi/backend/pkg/httpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	Apps   *mongo.Collection
	Elons  *mongo.Collection
	Users  *mongo.Collection
	Notify *notification.Service
}

func NewHandler(db *mongo.Database, n *notification.Service) *Handler {
	return &Handler{
		Apps:   db.Collection("applications"),
		Elons:  db.Collection("elons"),
		Users:  db.Collection("users"),
		Notify: n,
	}
}

type applyReq struct {
	Phone string `json:"phone"`
}

func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	elonID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad elon id"))
		return
	}
	var req applyReq
	_ = httpx.Decode(r, &req)

	var elon models.Elon
	if err := h.Elons.FindOne(r.Context(), bson.M{"_id": elonID, "isDeleted": bson.M{"$ne": true}}).Decode(&elon); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "elon not found"))
		return
	}
	if elon.OwnerID == uid {
		httpx.Err(w, httpx.NewError(400, "self_apply", "cannot apply to own elon"))
		return
	}
	if elon.Status != "recruiting" {
		httpx.Err(w, httpx.NewError(400, "not_recruiting", "elon not accepting applications"))
		return
	}
	worker, _ := loadUser(r.Context(), h.Users, uid)
	phone := req.Phone
	if phone == "" && worker != nil {
		phone = worker.Phone
	}
	app := models.Application{
		ElonID:       elonID,
		ElonTitle:    elon.Title,
		WorkerID:     uid,
		EmployerID:   elon.OwnerID,
		WorkerPhone:  phone,
		Amount:       elon.PerWorkerAmount,
		IsNegotiable: elon.PricingType == "negotiable",
		Status:       "pending",
		AppliedAt:    time.Now(),
		// Elon snapshot (ishchining arizalar ro'yxati uchun).
		ElonCategoryName: elon.CategoryName,
		ElonRegion:       elon.Region,
		ElonDistrict:     elon.District,
		OwnerName:        elon.OwnerName,
		OwnerRating:      elon.OwnerRating,
	}
	if worker != nil {
		// Worker snapshot (ish beruvchining nomzodlar ro'yxati uchun).
		app.WorkerName = strings.TrimSpace(worker.FirstName + " " + worker.LastName)
		app.WorkerRating = worker.WorkerRating
		app.WorkerReviewsCount = worker.WorkerReviewsCount
		app.WorkerAvatarURL = worker.AvatarURL
		app.WorkerVerified = worker.IsPhoneVerified
	}
	res, err := h.Apps.InsertOne(r.Context(), app)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			httpx.Err(w, httpx.NewError(409, "duplicate", "you already applied"))
			return
		}
		httpx.Err(w, err)
		return
	}
	app.ID = res.InsertedID.(primitive.ObjectID)
	h.Notify.Push(r.Context(), elon.OwnerID, "new_application", "Yangi ariza", "Sizning e'loningizga ariza tushdi: "+elon.Title, &models.RelatedEntity{Type: "application", ID: app.ID})
	httpx.JSON(w, 201, app)
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, "accepted")
}
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, "rejected")
}

func (h *Handler) decide(w http.ResponseWriter, r *http.Request, decision string) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	appID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var app models.Application
	if err := h.Apps.FindOne(r.Context(), bson.M{"_id": appID}).Decode(&app); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "application not found"))
		return
	}
	if app.EmployerID != uid {
		httpx.Err(w, httpx.NewError(403, "forbidden", "not your application to decide"))
		return
	}
	if app.Status != "pending" {
		httpx.Err(w, httpx.NewError(400, "bad_state", "application is not pending"))
		return
	}
	now := time.Now()
	set := bson.M{"status": decision, "decidedAt": now}
	if _, err := h.Apps.UpdateOne(r.Context(), bson.M{"_id": appID}, bson.M{"$set": set}); err != nil {
		httpx.Err(w, err)
		return
	}
	if decision == "accepted" {
		// increment elon acceptedCount, mark filled if reached
		var elon models.Elon
		_ = h.Elons.FindOneAndUpdate(r.Context(),
			bson.M{"_id": app.ElonID},
			bson.M{"$inc": bson.M{"acceptedCount": 1}, "$set": bson.M{"updatedAt": now}},
			options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&elon)
		if elon.AcceptedCount >= elon.WorkersNeeded && elon.Status == "recruiting" {
			_, _ = h.Elons.UpdateOne(r.Context(), bson.M{"_id": elon.ID}, bson.M{"$set": bson.M{"status": "filled"}})
		}
		h.Notify.Push(r.Context(), app.WorkerID, "application_accepted", "Arizangiz qabul qilindi", elon.Title, &models.RelatedEntity{Type: "application", ID: appID})
	} else {
		h.Notify.Push(r.Context(), app.WorkerID, "application_rejected", "Arizangiz rad etildi", app.ElonTitle, &models.RelatedEntity{Type: "application", ID: appID})
	}
	httpx.JSON(w, 200, map[string]string{"status": decision})
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	appID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var app models.Application
	if err := h.Apps.FindOne(r.Context(), bson.M{"_id": appID}).Decode(&app); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "application not found"))
		return
	}
	var who string
	switch uid {
	case app.WorkerID:
		who = "worker"
	case app.EmployerID:
		who = "employer"
	default:
		httpx.Err(w, httpx.NewError(403, "forbidden", "not your application"))
		return
	}
	if app.Status != "pending" && app.Status != "accepted" {
		httpx.Err(w, httpx.NewError(400, "bad_state", "cannot cancel from this state"))
		return
	}
	wasAccepted := app.Status == "accepted"
	_, err = h.Apps.UpdateOne(r.Context(), bson.M{"_id": appID}, bson.M{"$set": bson.M{"status": "cancelled", "cancelledBy": who, "decidedAt": time.Now()}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	if wasAccepted {
		// roll back acceptedCount + status if needed
		var elon models.Elon
		_ = h.Elons.FindOneAndUpdate(r.Context(),
			bson.M{"_id": app.ElonID},
			bson.M{"$inc": bson.M{"acceptedCount": -1}, "$set": bson.M{"updatedAt": time.Now()}},
			options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&elon)
		if elon.Status == "filled" && elon.AcceptedCount < elon.WorkersNeeded {
			_, _ = h.Elons.UpdateOne(r.Context(), bson.M{"_id": elon.ID}, bson.M{"$set": bson.M{"status": "recruiting"}})
		}
	}
	// notify the other party
	other := app.EmployerID
	if uid == app.EmployerID {
		other = app.WorkerID
	}
	h.Notify.Push(r.Context(), other, "application_cancelled", "Ariza bekor qilindi", app.ElonTitle, &models.RelatedEntity{Type: "application", ID: appID})
	httpx.JSON(w, 200, map[string]string{"status": "cancelled"})
}

// ConfirmDone: dual-confirm completion.
func (h *Handler) ConfirmDone(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	appID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var app models.Application
	if err := h.Apps.FindOne(r.Context(), bson.M{"_id": appID}).Decode(&app); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "application not found"))
		return
	}
	if app.Status != "accepted" {
		httpx.Err(w, httpx.NewError(400, "bad_state", "only accepted applications can be confirmed"))
		return
	}
	var setField string
	switch uid {
	case app.WorkerID:
		setField = "workerConfirmedDone"
	case app.EmployerID:
		setField = "employerConfirmedDone"
	default:
		httpx.Err(w, httpx.NewError(403, "forbidden", "not your application"))
		return
	}
	_, err = h.Apps.UpdateOne(r.Context(), bson.M{"_id": appID}, bson.M{"$set": bson.M{setField: true}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	// refetch and decide
	if err := h.Apps.FindOne(r.Context(), bson.M{"_id": appID}).Decode(&app); err != nil {
		httpx.Err(w, err)
		return
	}
	if app.EmployerConfirmedDone && app.WorkerConfirmedDone {
		now := time.Now()
		_, _ = h.Apps.UpdateOne(r.Context(), bson.M{"_id": appID}, bson.M{"$set": bson.M{"status": "completed", "completedAt": now}})
		// bump completedJobsCount on both users
		_, _ = h.Users.UpdateOne(r.Context(), bson.M{"_id": app.WorkerID}, bson.M{"$inc": bson.M{"completedJobsCount": 1}})
		_, _ = h.Users.UpdateOne(r.Context(), bson.M{"_id": app.EmployerID}, bson.M{"$inc": bson.M{"completedJobsCount": 1}})
		// notify both
		h.Notify.Push(r.Context(), app.WorkerID, "job_completed", "Ish yakunlandi", app.ElonTitle, &models.RelatedEntity{Type: "application", ID: appID})
		h.Notify.Push(r.Context(), app.EmployerID, "job_completed", "Ish yakunlandi", app.ElonTitle, &models.RelatedEntity{Type: "application", ID: appID})
		httpx.JSON(w, 200, map[string]string{"status": "completed"})
		return
	}
	// notify the other side to confirm
	other := app.EmployerID
	if uid == app.EmployerID {
		other = app.WorkerID
	}
	h.Notify.Push(r.Context(), other, "job_completed_request", "Tasdiqlash so'rovi", "Ish yakunlanganini tasdiqlang: "+app.ElonTitle, &models.RelatedEntity{Type: "application", ID: appID})
	httpx.JSON(w, 200, map[string]string{"status": "awaiting_other"})
}

// MyApplications: applications I made as worker.
func (h *Handler) MyApplications(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cur, err := h.Apps.Find(r.Context(), bson.M{"workerId": uid}, options.Find().SetSort(bson.D{{Key: "appliedAt", Value: -1}}))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Application{}
	for cur.Next(r.Context()) {
		var a models.Application
		if err := cur.Decode(&a); err == nil {
			out = append(out, a)
		}
	}
	httpx.JSON(w, 200, out)
}

// MyElonsApplications: applications received on my elons (grouped by elon).
func (h *Handler) MyElonsApplications(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cur, err := h.Apps.Find(r.Context(), bson.M{"employerId": uid}, options.Find().SetSort(bson.D{{Key: "appliedAt", Value: -1}}))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	grouped := map[string][]models.Application{}
	for cur.Next(r.Context()) {
		var a models.Application
		if err := cur.Decode(&a); err == nil {
			grouped[a.ElonID.Hex()] = append(grouped[a.ElonID.Hex()], a)
		}
	}
	httpx.JSON(w, 200, grouped)
}

// History: completed/cancelled/rejected for the user (worker or employer).
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	filter := bson.M{
		"$or":    []bson.M{{"workerId": uid}, {"employerId": uid}},
		"status": bson.M{"$in": []string{"completed", "cancelled", "rejected"}},
	}
	cur, err := h.Apps.Find(r.Context(), filter, options.Find().SetSort(bson.D{{Key: "appliedAt", Value: -1}}))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Application{}
	for cur.Next(r.Context()) {
		var a models.Application
		if err := cur.Decode(&a); err == nil {
			out = append(out, a)
		}
	}
	httpx.JSON(w, 200, out)
}

func loadUser(ctx context.Context, col *mongo.Collection, id primitive.ObjectID) (*models.User, error) {
	var u models.User
	if err := col.FindOne(ctx, bson.M{"_id": id}).Decode(&u); err != nil {
		return nil, errors.New("not_found")
	}
	return &u, nil
}
