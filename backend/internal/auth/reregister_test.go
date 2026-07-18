package auth

import (
	"context"
	"testing"
	"time"

	"github.com/ishchibormi/backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Runs against a real local Mongo; skips when none is reachable.
func testDB(t *testing.T) *mongo.Database {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cl, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skipf("no local mongo: %v", err)
	}
	if err := cl.Ping(ctx, nil); err != nil {
		t.Skipf("no local mongo: %v", err)
	}
	db := cl.Database("ishchibormi_authtest")
	t.Cleanup(func() {
		bg := context.Background()
		_ = db.Drop(bg)
		_ = cl.Disconnect(bg)
	})
	return db
}

const reusedPhone = "+998900000077"

// Signing in with a number whose previous account was deleted must produce a
// brand-new account — new _id, empty profile, zeroed counters — never a revival
// of the old (isDeleted) record, which would leave the user permanently locked
// out by RequireActiveUser and would surface their old data.
//
// The deletion side is simulated exactly as account.softDelete leaves things:
// isDeleted set, phone/telegramId unset.
func TestUpsertUserAfterDeletionCreatesFreshAccount(t *testing.T) {
	db := testDB(t)
	h := NewHandler(config.Config{}, db)
	ctx := context.Background()

	oldID := primitive.NewObjectID()
	if _, err := db.Collection("users").InsertOne(ctx, bson.M{
		"_id": oldID, "phone": reusedPhone, "telegramId": int64(777001),
		"firstName": "Eski", "lastName": "Foydalanuvchi", "bio": "eski bio",
		"avatarUrl": "https://cdn/old.png", "rating": 4.8, "reviewsCount": 12,
		"completedJobsCount": 30, "isDeleted": false, "createdAt": time.Now(),
	}); err != nil {
		t.Fatalf("seed old user: %v", err)
	}

	// What account.softDelete does to the identity.
	if _, err := db.Collection("users").UpdateOne(ctx,
		bson.M{"_id": oldID},
		bson.M{
			"$set":   bson.M{"isDeleted": true, "deletedPhone": reusedPhone, "deletedTelegramId": int64(777001)},
			"$unset": bson.M{"phone": "", "telegramId": ""},
		}); err != nil {
		t.Fatalf("simulate delete: %v", err)
	}

	fresh, err := h.upsertUser(ctx, reusedPhone, 777001)
	if err != nil {
		t.Fatalf("upsertUser: %v", err)
	}

	if fresh.ID == oldID {
		t.Fatal("re-registration revived the deleted account instead of creating a new one")
	}
	if fresh.IsDeleted {
		t.Fatal("the new account is flagged deleted — the user would be locked out")
	}

	// None of the previous account's data may carry over.
	if fresh.FirstName != "" || fresh.LastName != "" || fresh.Bio != "" || fresh.AvatarURL != "" {
		t.Fatalf("old profile leaked into the new account: %+v", fresh)
	}
	if fresh.Rating != 0 || fresh.ReviewsCount != 0 || fresh.CompletedJobsCount != 0 {
		t.Fatalf("old reputation leaked: rating=%v reviews=%d jobs=%d",
			fresh.Rating, fresh.ReviewsCount, fresh.CompletedJobsCount)
	}

	// The old record is still there, still deleted, still detached.
	var old bson.M
	if err := db.Collection("users").FindOne(ctx, bson.M{"_id": oldID}).Decode(&old); err != nil {
		t.Fatalf("old record vanished: %v", err)
	}
	if old["isDeleted"] != true {
		t.Fatal("old record is no longer marked deleted")
	}
	if _, ok := old["phone"]; ok {
		t.Fatal("old record reclaimed the phone number")
	}
}

// Signing in again on the fresh account must keep returning that same account
// rather than piling up a new one per login.
func TestUpsertUserIsStableAcrossLogins(t *testing.T) {
	db := testDB(t)
	h := NewHandler(config.Config{}, db)
	ctx := context.Background()

	first, err := h.upsertUser(ctx, reusedPhone, 777002)
	if err != nil {
		t.Fatalf("first login: %v", err)
	}
	second, err := h.upsertUser(ctx, reusedPhone, 777002)
	if err != nil {
		t.Fatalf("second login: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("logging in twice created two accounts (%s, %s)", first.ID.Hex(), second.ID.Hex())
	}
}
