package elon

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ishchibormi/backend/internal/category"
	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/internal/upload"
	"github.com/ishchibormi/backend/pkg/httpx"
	"github.com/ishchibormi/backend/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	Col        *mongo.Collection
	Categories *mongo.Collection
	Users      *mongo.Collection
	Storage    *storage.Service
}

func NewHandler(db *mongo.Database, s *storage.Service) *Handler {
	return &Handler{
		Col:        db.Collection("elons"),
		Categories: db.Collection("categories"),
		Users:      db.Collection("users"),
		Storage:    s,
	}
}

type upsertReq struct {
	Title         string   `json:"title" validate:"required"`
	CategoryID    string   `json:"categoryId" validate:"required"`
	Description   string   `json:"description" validate:"required"`
	LocationURL   string   `json:"locationUrl"`
	LocationText  string   `json:"locationText"`
	Region        string   `json:"region"`
	District      string   `json:"district"`
	WorkersNeeded int      `json:"workersNeeded" validate:"required,gte=1"`
	PricingType   string   `json:"pricingType"` // per_worker|total|negotiable
	PriceAmount   int64    `json:"priceAmount"`
	StartDate     string   `json:"startDate"`
	WorkTimeFrom  string   `json:"workTimeFrom"`
	WorkTimeTo    string   `json:"workTimeTo"`
	ContactPhone  string   `json:"contactPhone"`
	Images        []string `json:"images"`
}

func (req *upsertReq) computePrice() (pType string, total int64, perWorker int64) {
	switch req.PricingType {
	case "per_worker":
		return "per_worker", req.PriceAmount * int64(req.WorkersNeeded), req.PriceAmount
	case "total":
		if req.WorkersNeeded <= 0 {
			return "negotiable", 0, 0
		}
		return "total", req.PriceAmount, req.PriceAmount / int64(req.WorkersNeeded)
	default:
		if req.PriceAmount <= 0 {
			return "negotiable", 0, 0
		}
		return "per_worker", req.PriceAmount * int64(req.WorkersNeeded), req.PriceAmount
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	var req upsertReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Description) == "" || req.WorkersNeeded < 1 {
		httpx.Err(w, httpx.NewError(400, "bad_request", "title, description and workersNeeded required"))
		return
	}
	catID, err := primitive.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_request", "bad categoryId"))
		return
	}
	var cat models.Category
	if err := h.Categories.FindOne(r.Context(), bson.M{"_id": catID}).Decode(&cat); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "category not found"))
		return
	}
	pType, total, per := req.computePrice()

	var owner models.User
	_ = h.Users.FindOne(r.Context(), bson.M{"_id": uid}).Decode(&owner)

	now := time.Now()
	e := models.Elon{
		OwnerID:         uid,
		Title:           strings.TrimSpace(req.Title),
		CategoryID:      catID,
		CategoryName:    cat.Name,
		Description:     req.Description,
		LocationURL:     req.LocationURL,
		LocationText:    req.LocationText,
		Region:          req.Region,
		District:        req.District,
		WorkersNeeded:   req.WorkersNeeded,
		PricingType:     pType,
		PriceAmount:     total,
		PerWorkerAmount: per,
		StartDate:       req.StartDate,
		WorkTimeFrom:    req.WorkTimeFrom,
		WorkTimeTo:      req.WorkTimeTo,
		ContactPhone:    req.ContactPhone,
		Status:          "draft",
		CreatedAt:       now,
		UpdatedAt:       now,
		OwnerName:       strings.TrimSpace(owner.FirstName + " " + owner.LastName),
		OwnerRating:     owner.Rating,
		Images:          req.Images,
	}
	res, err := h.Col.InsertOne(r.Context(), e)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	e.ID = res.InsertedID.(primitive.ObjectID)
	category.IncrementUsage(r.Context(), h.Categories, catID)
	httpx.JSON(w, 201, e)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var e models.Elon
	if err := h.Col.FindOne(r.Context(), bson.M{"_id": id, "isDeleted": bson.M{"$ne": true}}).Decode(&e); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "elon not found"))
		return
	}
	// bump view count async
	go func() {
		_, _ = h.Col.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$inc": bson.M{"viewsCount": 1}})
	}()
	httpx.JSON(w, 200, e)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var req upsertReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	pType, total, per := req.computePrice()
	// Image diff: delete any S3 images that are removed from the new list.
	if req.Images != nil {
		var prev models.Elon
		if err := h.Col.FindOne(r.Context(), bson.M{"_id": id, "ownerId": uid}).Decode(&prev); err == nil {
			keep := map[string]bool{}
			for _, u := range req.Images {
				keep[u] = true
			}
			for _, u := range prev.Images {
				if !keep[u] {
					go upload.DeleteByURL(h.Storage, u)
				}
			}
		}
	}
	set := bson.M{
		"title":           req.Title,
		"description":     req.Description,
		"locationUrl":     req.LocationURL,
		"locationText":    req.LocationText,
		"region":          req.Region,
		"district":        req.District,
		"workersNeeded":   req.WorkersNeeded,
		"pricingType":     pType,
		"priceAmount":     total,
		"perWorkerAmount": per,
		"startDate":       req.StartDate,
		"workTimeFrom":    req.WorkTimeFrom,
		"workTimeTo":      req.WorkTimeTo,
		"contactPhone":    req.ContactPhone,
		"updatedAt":       time.Now(),
	}
	if req.Images != nil {
		set["images"] = req.Images
	}
	if req.CategoryID != "" {
		if catID, err := primitive.ObjectIDFromHex(req.CategoryID); err == nil {
			var cat models.Category
			if err := h.Categories.FindOne(r.Context(), bson.M{"_id": catID}).Decode(&cat); err == nil {
				set["categoryId"] = catID
				set["categoryName"] = cat.Name
			}
		}
	}
	res := h.Col.FindOneAndUpdate(r.Context(),
		bson.M{"_id": id, "ownerId": uid, "status": bson.M{"$in": []string{"draft", "recruiting", "filled"}}},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After))
	var e models.Elon
	if err := res.Decode(&e); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found_or_forbidden", "elon not found or not yours"))
		return
	}
	httpx.JSON(w, 200, e)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	// Look up images before soft-deleting so we can remove them from S3.
	var prev models.Elon
	_ = h.Col.FindOne(r.Context(), bson.M{"_id": id, "ownerId": uid}).Decode(&prev)
	res, err := h.Col.UpdateOne(r.Context(),
		bson.M{"_id": id, "ownerId": uid},
		bson.M{"$set": bson.M{"isDeleted": true, "status": "cancelled", "updatedAt": time.Now()}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	if res.MatchedCount == 0 {
		httpx.Err(w, httpx.NewError(404, "not_found_or_forbidden", "elon not found or not yours"))
		return
	}
	for _, u := range prev.Images {
		go upload.DeleteByURL(h.Storage, u)
	}
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	now := time.Now()
	res := h.Col.FindOneAndUpdate(r.Context(),
		bson.M{"_id": id, "ownerId": uid, "status": "draft"},
		bson.M{"$set": bson.M{"status": "recruiting", "publishedAt": now, "updatedAt": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After))
	var e models.Elon
	if err := res.Decode(&e); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found_or_forbidden", "draft not found"))
		return
	}
	httpx.JSON(w, 200, e)
}

// Feed: public listing for recruiting + filled (paged).
func (h *Handler) Feed(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	cat := strings.TrimSpace(r.URL.Query().Get("categoryId"))
	sort := r.URL.Query().Get("sort") // price|time|rating
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 24
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	filter := bson.M{"isDeleted": bson.M{"$ne": true}, "status": bson.M{"$in": []string{"recruiting", "filled"}}}
	if q != "" {
		rx := primitive.Regex{Pattern: regexpEscape(q), Options: "i"}
		filter["$or"] = []bson.M{{"title": rx}, {"description": rx}, {"locationText": rx}, {"categoryName": rx}}
	}
	if cat != "" {
		if cid, err := primitive.ObjectIDFromHex(cat); err == nil {
			filter["categoryId"] = cid
		}
	}
	sortDoc := bson.D{{Key: "publishedAt", Value: -1}}
	switch sort {
	case "price":
		sortDoc = bson.D{{Key: "perWorkerAmount", Value: -1}}
	case "rating":
		sortDoc = bson.D{{Key: "ownerRating", Value: -1}}
	case "time":
		sortDoc = bson.D{{Key: "publishedAt", Value: -1}}
	}
	cur, err := h.Col.Find(r.Context(), filter,
		options.Find().SetSort(sortDoc).SetSkip(int64((page-1)*limit)).SetLimit(int64(limit)))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Elon{}
	for cur.Next(r.Context()) {
		var e models.Elon
		if err := cur.Decode(&e); err == nil {
			out = append(out, e)
		}
	}
	total, _ := h.Col.CountDocuments(r.Context(), filter)
	httpx.JSON(w, 200, map[string]any{"items": out, "page": page, "limit": limit, "total": total})
}

// MyElons: owner's elons grouped by status.
func (h *Handler) MyElons(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cur, err := h.Col.Find(r.Context(),
		bson.M{"ownerId": uid},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	drafts := []models.Elon{}
	active := []models.Elon{}
	archived := []models.Elon{}
	for cur.Next(r.Context()) {
		var e models.Elon
		if err := cur.Decode(&e); err == nil {
			switch e.Status {
			case "draft":
				drafts = append(drafts, e)
			case "recruiting", "filled", "in_progress":
				if e.IsDeleted {
					archived = append(archived, e)
				} else {
					active = append(active, e)
				}
			default: // completed/cancelled
				archived = append(archived, e)
			}
		}
	}
	httpx.JSON(w, 200, map[string]any{"drafts": drafts, "active": active, "archived": archived})
}

func regexpEscape(s string) string {
	r := strings.NewReplacer(
		".", `\.`, "*", `\*`, "+", `\+`, "?", `\?`, "(", `\(`,
		")", `\)`, "[", `\[`, "]", `\]`, "{", `\{`, "}", `\}`,
		"|", `\|`, "^", `\^`, "$", `\$`, `\`, `\\`,
	)
	return r.Replace(s)
}
