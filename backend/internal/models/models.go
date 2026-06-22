package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User -- platform user (both employer and worker)
type User struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TelegramID          int64              `bson:"telegramId,omitempty" json:"telegramId"`
	Phone               string             `bson:"phone" json:"phone"`
	FirstName           string             `bson:"firstName" json:"firstName"`
	LastName            string             `bson:"lastName" json:"lastName"`
	AvatarURL           string             `bson:"avatarUrl,omitempty" json:"avatarUrl"`
	Region              string             `bson:"region,omitempty" json:"region"`
	District            string             `bson:"district,omitempty" json:"district"`
	Bio                 string             `bson:"bio,omitempty" json:"bio"`
	Skills              []string           `bson:"skills,omitempty" json:"skills"`
	Rating              float64            `bson:"rating" json:"rating"`
	ReviewsCount        int                `bson:"reviewsCount" json:"reviewsCount"`
	CompletedJobsCount  int                `bson:"completedJobsCount" json:"completedJobsCount"`
	IsPhoneVerified     bool               `bson:"isPhoneVerified" json:"isPhoneVerified"`
	IsPremium           bool               `bson:"isPremium" json:"isPremium"`
	IsBlocked           bool               `bson:"isBlocked" json:"isBlocked"`
	IsDeleted           bool               `bson:"isDeleted" json:"isDeleted"`
	LangPref            string             `bson:"langPref" json:"langPref"`
	ThemePref           string             `bson:"themePref" json:"themePref"`
	BlockedUserIDs      []primitive.ObjectID `bson:"blockedUserIds,omitempty" json:"blockedUserIds"`
	OnboardingCompleted bool               `bson:"onboardingCompleted" json:"onboardingCompleted"`
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// PublicUser is a safe projection.
type PublicUser struct {
	ID                 primitive.ObjectID `json:"id"`
	FirstName          string             `json:"firstName"`
	LastName           string             `json:"lastName"`
	AvatarURL          string             `json:"avatarUrl"`
	Region             string             `json:"region"`
	District           string             `json:"district"`
	Bio                string             `json:"bio"`
	Skills             []string           `json:"skills"`
	Rating             float64            `json:"rating"`
	ReviewsCount       int                `json:"reviewsCount"`
	CompletedJobsCount int                `json:"completedJobsCount"`
	IsPhoneVerified    bool               `json:"isPhoneVerified"`
}

func (u *User) Public() PublicUser {
	return PublicUser{
		ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, AvatarURL: u.AvatarURL,
		Region: u.Region, District: u.District, Bio: u.Bio, Skills: u.Skills,
		Rating: u.Rating, ReviewsCount: u.ReviewsCount,
		CompletedJobsCount: u.CompletedJobsCount, IsPhoneVerified: u.IsPhoneVerified,
	}
}

// Category
type Category struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Slug            string             `bson:"slug" json:"slug"`
	Icon            string             `bson:"icon,omitempty" json:"icon"`
	CreatedBy       primitive.ObjectID `bson:"createdBy,omitempty" json:"createdBy"`
	IsSystemDefault bool               `bson:"isSystemDefault" json:"isSystemDefault"`
	IsActive        bool               `bson:"isActive" json:"isActive"`
	UsageCount      int                `bson:"usageCount" json:"usageCount"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}

// Elon (job listing)
type Elon struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID        primitive.ObjectID `bson:"ownerId" json:"ownerId"`
	Title          string             `bson:"title" json:"title"`
	CategoryID     primitive.ObjectID `bson:"categoryId" json:"categoryId"`
	CategoryName   string             `bson:"categoryName" json:"categoryName"`
	Description    string             `bson:"description" json:"description"`
	LocationURL    string             `bson:"locationUrl,omitempty" json:"locationUrl"`
	LocationText   string             `bson:"locationText,omitempty" json:"locationText"`
	Region         string             `bson:"region,omitempty" json:"region"`
	District       string             `bson:"district,omitempty" json:"district"`
	WorkersNeeded  int                `bson:"workersNeeded" json:"workersNeeded"`
	PricingType    string             `bson:"pricingType" json:"pricingType"` // per_worker|total|negotiable
	PriceAmount    int64              `bson:"priceAmount" json:"priceAmount"`
	PerWorkerAmount int64             `bson:"perWorkerAmount" json:"perWorkerAmount"`
	StartDate      string             `bson:"startDate,omitempty" json:"startDate"`
	WorkTimeFrom   string             `bson:"workTimeFrom,omitempty" json:"workTimeFrom"`
	WorkTimeTo     string             `bson:"workTimeTo,omitempty" json:"workTimeTo"`
	ContactPhone   string             `bson:"contactPhone,omitempty" json:"contactPhone"`
	Status         string             `bson:"status" json:"status"` // draft|recruiting|filled|in_progress|completed|cancelled
	AcceptedCount  int                `bson:"acceptedCount" json:"acceptedCount"`
	ViewsCount     int                `bson:"viewsCount" json:"viewsCount"`
	IsDeleted      bool               `bson:"isDeleted" json:"isDeleted"`
	PublishedAt    *time.Time         `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
	// Denormalized owner info for fast feed
	OwnerName   string  `bson:"ownerName,omitempty" json:"ownerName"`
	OwnerRating float64 `bson:"ownerRating,omitempty" json:"ownerRating"`
	// Image URLs (stored on S3).
	Images []string `bson:"images,omitempty" json:"images"`
}

// Application
type Application struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ElonID                 primitive.ObjectID `bson:"elonId" json:"elonId"`
	ElonTitle              string             `bson:"elonTitle" json:"elonTitle"`
	WorkerID               primitive.ObjectID `bson:"workerId" json:"workerId"`
	EmployerID             primitive.ObjectID `bson:"employerId" json:"employerId"`
	WorkerPhone            string             `bson:"workerPhone" json:"workerPhone"`
	Amount                 int64              `bson:"amount" json:"amount"`
	IsNegotiable           bool               `bson:"isNegotiable" json:"isNegotiable"`
	Status                 string             `bson:"status" json:"status"` // pending|accepted|rejected|cancelled|completed
	EmployerConfirmedDone  bool               `bson:"employerConfirmedDone" json:"employerConfirmedDone"`
	WorkerConfirmedDone    bool               `bson:"workerConfirmedDone" json:"workerConfirmedDone"`
	CancelledBy            string             `bson:"cancelledBy,omitempty" json:"cancelledBy"`
	AppliedAt              time.Time          `bson:"appliedAt" json:"appliedAt"`
	DecidedAt              *time.Time         `bson:"decidedAt,omitempty" json:"decidedAt,omitempty"`
	CompletedAt            *time.Time         `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
}

// Review
type Review struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ApplicationID primitive.ObjectID `bson:"applicationId" json:"applicationId"`
	ElonID        primitive.ObjectID `bson:"elonId" json:"elonId"`
	FromUserID    primitive.ObjectID `bson:"fromUserId" json:"fromUserId"`
	ToUserID      primitive.ObjectID `bson:"toUserId" json:"toUserId"`
	Direction     string             `bson:"direction" json:"direction"`
	Rating        int                `bson:"rating" json:"rating"`
	Comment       string             `bson:"comment,omitempty" json:"comment"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}

// Conversation
type Conversation struct {
	ID              primitive.ObjectID            `bson:"_id,omitempty" json:"id"`
	ParticipantIDs  []primitive.ObjectID          `bson:"participantIds" json:"participantIds"`
	LastMessageText string                        `bson:"lastMessageText" json:"lastMessageText"`
	LastMessageAt   time.Time                     `bson:"lastMessageAt" json:"lastMessageAt"`
	LastSenderID    primitive.ObjectID            `bson:"lastSenderId,omitempty" json:"lastSenderId"`
	Unread          map[string]int                `bson:"unread" json:"unread"`
	CreatedAt       time.Time                     `bson:"createdAt" json:"createdAt"`
}

// MessageAttachment is a single file attached to a message.
type MessageAttachment struct {
	URL  string `bson:"url" json:"url"`
	Name string `bson:"name,omitempty" json:"name"`
	Size int64  `bson:"size,omitempty" json:"size"`
	Mime string `bson:"mime,omitempty" json:"mime"`
}

// Message
type Message struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID  `bson:"conversationId" json:"conversationId"`
	SenderID       primitive.ObjectID  `bson:"senderId" json:"senderId"`
	Text           string              `bson:"text" json:"text"`
	Attachments    []MessageAttachment `bson:"attachments,omitempty" json:"attachments"`
	IsRead         bool                `bson:"isRead" json:"isRead"`
	CreatedAt      time.Time           `bson:"createdAt" json:"createdAt"`
}

// Notification
type Notification struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"userId" json:"userId"`
	Type           string             `bson:"type" json:"type"`
	Title          string             `bson:"title" json:"title"`
	Body           string             `bson:"body" json:"body"`
	RelatedEntity  *RelatedEntity     `bson:"relatedEntity,omitempty" json:"relatedEntity,omitempty"`
	IsRead         bool               `bson:"isRead" json:"isRead"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
}
type RelatedEntity struct {
	Type string             `bson:"type" json:"type"`
	ID   primitive.ObjectID `bson:"id" json:"id"`
}

// Report
type Report struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReporterID  primitive.ObjectID `bson:"reporterId" json:"reporterId"`
	TargetType  string             `bson:"targetType" json:"targetType"` // user|elon|message
	TargetID    primitive.ObjectID `bson:"targetId" json:"targetId"`
	Reason      string             `bson:"reason" json:"reason"`
	Description string             `bson:"description,omitempty" json:"description"`
	Status      string             `bson:"status" json:"status"` // open|resolved|dismissed
	ReviewedBy  primitive.ObjectID `bson:"reviewedBy,omitempty" json:"reviewedBy,omitempty"`
	ReviewedAt  *time.Time         `bson:"reviewedAt,omitempty" json:"reviewedAt,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

// FinanceEntry
type FinanceEntry struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"userId" json:"userId"`
	Role           string             `bson:"role" json:"role"` // worker|employer
	ApplicationID  primitive.ObjectID `bson:"applicationId" json:"applicationId"`
	ElonID         primitive.ObjectID `bson:"elonId" json:"elonId"`
	ElonTitle      string             `bson:"elonTitle" json:"elonTitle"`
	CounterpartyID primitive.ObjectID `bson:"counterpartyId" json:"counterpartyId"`
	Type           string             `bson:"type" json:"type"` // earned|spent
	Status         string             `bson:"status" json:"status"` // completed|cancelled
	Amount         int64              `bson:"amount" json:"amount"`
	IsNegotiable   bool               `bson:"isNegotiable" json:"isNegotiable"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
}

// Admin
type Admin struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Role         string             `bson:"role" json:"role"`
	IsActive     bool               `bson:"isActive" json:"isActive"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}

// AdminAuditLog
type AdminAudit struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminID   primitive.ObjectID `bson:"adminId" json:"adminId"`
	Action    string             `bson:"action" json:"action"`
	Target    string             `bson:"target,omitempty" json:"target"`
	Detail    string             `bson:"detail,omitempty" json:"detail"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// OTPCode
type OTPCode struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	TGToken    string             `bson:"tgToken"`
	Phone      string             `bson:"phone,omitempty"`
	TelegramID int64              `bson:"telegramId,omitempty"`
	Code       string             `bson:"code,omitempty"`
	ExpiresAt  time.Time          `bson:"expiresAt"`
	Used       bool               `bson:"used"`
	CreatedAt  time.Time          `bson:"createdAt"`
}
