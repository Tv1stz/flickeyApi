// Package db provides GORM models matching the exact PostgreSQL schema
// created by the Alembic migration 1527e02c9a71_initial_schema.py.
package db

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// UserRole mirrors Python UserRole enum.
type UserRole string

const (
	UserRoleGuest UserRole = "guest"
	UserRoleHost  UserRole = "host"
	UserRoleAdmin UserRole = "admin"
)

// UserStatus mirrors Python UserStatus enum.
type UserStatus string

const (
	UserStatusPendingProfile UserStatus = "pending_profile"
	UserStatusActive         UserStatus = "active"
	UserStatusSuspended      UserStatus = "suspended"
	UserStatusBanned         UserStatus = "banned"
)

// User is the primary user entity with phone-based authentication.
// Table: users
type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Phone     string     `gorm:"type:varchar(20);not null;uniqueIndex:ix_users_phone" json:"phone"`
	FirstName *string    `gorm:"type:varchar(100)" json:"first_name"`
	LastName  *string    `gorm:"type:varchar(100)" json:"last_name"`
	Email     *string    `gorm:"type:varchar(255)" json:"email"`
	Role      UserRole   `gorm:"type:varchar(20);not null;default:'guest'" json:"role"`
	Status    UserStatus `gorm:"type:varchar(30);not null;default:'pending_profile'" json:"status"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "users" }

// MediaStatus mirrors Python MediaStatus enum.
type MediaStatus string

const (
	MediaStatusPending  MediaStatus = "pending"
	MediaStatusUploaded MediaStatus = "uploaded"
	MediaStatusAttached MediaStatus = "attached"
)

// Media represents an uploaded file in S3.
// Table: media
type Media struct {
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HostID        uuid.UUID   `gorm:"type:uuid;not null;index:ix_media_host_id" json:"host_id"`
	ListingID     *uuid.UUID  `gorm:"type:uuid;index:ix_media_listing_id" json:"listing_id"`
	FileKey       string      `gorm:"type:varchar(255);not null;uniqueIndex:ix_media_file_key" json:"file_key"`
	Status        MediaStatus `gorm:"type:varchar(50);not null;index:ix_media_status" json:"status"`
	ContentType   string      `gorm:"type:varchar(100);not null" json:"content_type"`
	FileSizeBytes int64       `gorm:"not null" json:"file_size_bytes"`
	CreatedAt     time.Time   `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (Media) TableName() string { return "media" }

func (m *Media) PublicURL(uploadBaseURL string) string {
	if uploadBaseURL != "" {
		return strings.TrimRight(uploadBaseURL, "/") + "/" + strings.TrimLeft(m.FileKey, "/")
	}
	return "/api/v1/media/dev-upload/" + strings.TrimLeft(m.FileKey, "/")
}

// ListingDraft stores multi-step wizard state.
// media_ids and amenities are JSON arrays.
// Table: listing_drafts
type ListingDraft struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HostID          uuid.UUID      `gorm:"type:uuid;not null;index:ix_listing_drafts_host_id" json:"host_id"`
	SourceListingID *uuid.UUID     `gorm:"type:uuid;index:ix_listing_drafts_source_listing_id" json:"source_listing_id,omitempty"`
	Mode            string         `gorm:"type:varchar(20);not null;default:'create'" json:"mode"` // create | edit
	CurrentStep     int            `gorm:"not null;default:1;check:ck_listing_drafts_current_step_range,current_step >= 1 AND current_step <= 6" json:"current_step"`
	Status          string         `gorm:"type:varchar(50);not null;default:'draft';index:ix_listing_drafts_status" json:"status"`
	Type            *string        `gorm:"type:varchar(50)" json:"type"`
	Name            *string        `gorm:"type:varchar(100)" json:"name"`
	Address         *string        `gorm:"type:text" json:"address"`
	City            *string        `gorm:"type:varchar(100)" json:"city"`
	Street          *string        `gorm:"type:varchar(150)" json:"street"`
	HouseNumber     *string        `gorm:"type:varchar(50)" json:"house_number"`
	Latitude        *float64       `gorm:"type:numeric(10,6)" json:"latitude"`
	Longitude       *float64       `gorm:"type:numeric(10,6)" json:"longitude"`
	Square          *float64       `gorm:"type:numeric(10,2)" json:"square"`
	Floor           *int           `gorm:"type:integer" json:"floor"`
	TotalFloors     *int           `gorm:"type:integer" json:"total_floors"`
	MaxGuests       *int           `gorm:"type:integer" json:"max_guests"`
	RoomsCount      *int           `gorm:"type:integer" json:"rooms_count"`
	BedsCount       *int           `gorm:"type:integer" json:"beds_count"`
	BathroomsCount  *int           `gorm:"type:integer" json:"bathrooms_count"`
	MediaIDs            datatypes.JSON `gorm:"type:json" json:"media_ids"` // []string of UUID strings
	Amenities           datatypes.JSON `gorm:"type:json" json:"amenities"` // []string of amenity IDs
	VerificationVideoID *uuid.UUID     `gorm:"type:uuid" json:"verification_video_id,omitempty"`
	PricePerNight       *float64       `gorm:"type:numeric(10,2)" json:"price_per_night"`
	Currency        *string        `gorm:"type:varchar(10)" json:"currency"`
	MinNights       *int           `gorm:"type:integer" json:"min_nights"`
	CheckinFrom     *string        `gorm:"type:time" json:"checkin_from"`  // stored as "HH:MM:SS"
	CheckoutUntil   *string        `gorm:"type:time" json:"checkout_until"`
	AllowChildren   *bool          `json:"allow_children"`
	AllowPets       *bool          `json:"allow_pets"`
	AllowSmoking    *bool          `json:"allow_smoking"`
	AllowParties    *bool          `json:"allow_parties"`
	DepositRequired *bool          `json:"deposit_required"`
	WithInvoicing   *bool          `json:"with_invoicing"`
	Description     *string        `gorm:"type:text" json:"description"`
	CreatedAt       time.Time      `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ListingDraft) TableName() string { return "listing_drafts" }

// Listing is a finalized submitted property listing.
// Table: listings
type Listing struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HostID          uuid.UUID `gorm:"type:uuid;not null;index:ix_listings_host_id" json:"host_id"`
	Status          string    `gorm:"type:varchar(50);not null;default:'pending_review';index:ix_listings_status" json:"status"`
	Type            string    `gorm:"type:varchar(50);not null" json:"type"`
	Name            string    `gorm:"type:varchar(100);not null" json:"name"`
	Address         string    `gorm:"type:text;not null;default:''" json:"address"`
	City            string    `gorm:"type:varchar(100);not null;default:'';index:ix_listings_city" json:"city"`
	Street          string    `gorm:"type:varchar(150);not null;default:''" json:"street"`
	HouseNumber     string    `gorm:"type:varchar(50);not null;default:''" json:"house_number"`
	Latitude        float64   `gorm:"type:numeric(10,6);not null;default:0;index:ix_listings_lat_lng" json:"latitude"`
	Longitude       float64   `gorm:"type:numeric(10,6);not null;default:0" json:"longitude"`
	Square          float64   `gorm:"type:numeric(10,2);not null" json:"square"`
	Floor           int       `gorm:"not null" json:"floor"`
	TotalFloors     int       `gorm:"not null" json:"total_floors"`
	MaxGuests       int       `gorm:"not null" json:"max_guests"`
	RoomsCount      int       `gorm:"not null" json:"rooms_count"`
	BedsCount       int       `gorm:"not null" json:"beds_count"`
	BathroomsCount  int       `gorm:"not null" json:"bathrooms_count"`
	PricePerNight   float64   `gorm:"type:numeric(10,2);not null" json:"price_per_night"`
	Currency        string    `gorm:"type:varchar(10);not null;default:'BYN'" json:"currency"`
	MinNights       int       `gorm:"not null;default:1" json:"min_nights"`
	CheckinFrom     string    `gorm:"type:time;not null" json:"checkin_from"` // "HH:MM:SS"
	CheckoutUntil   string    `gorm:"type:time;not null" json:"checkout_until"`
	AllowChildren   bool      `gorm:"not null;default:true" json:"allow_children"`
	AllowPets       bool      `gorm:"not null;default:false" json:"allow_pets"`
	AllowSmoking    bool      `gorm:"not null;default:false" json:"allow_smoking"`
	AllowParties    bool      `gorm:"not null;default:false" json:"allow_parties"`
	DepositRequired bool      `gorm:"not null;default:false" json:"deposit_required"`
	WithInvoicing   bool      `gorm:"not null;default:false" json:"with_invoicing"`
	Description     string    `gorm:"type:text;not null" json:"description"`
	CreatedAt       time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`

	VerificationVideoID *uuid.UUID `gorm:"type:uuid;index:ix_listings_verification_video_id" json:"verification_video_id,omitempty"`

	RejectionReason   *string `gorm:"type:varchar(255)" json:"rejection_reason,omitempty"`
	ModerationComment *string `gorm:"type:text" json:"moderation_comment,omitempty"`

	// Relationships
	Host              *User            `gorm:"foreignKey:HostID" json:"host,omitempty"`
	ListingAmenities  []ListingAmenity `gorm:"foreignKey:ListingID;constraint:OnDelete:CASCADE" json:"listing_amenities,omitempty"`
	Media             []Media          `gorm:"foreignKey:ListingID;constraint:OnDelete:SET NULL" json:"media,omitempty"`
	VerificationVideo *Media           `gorm:"foreignKey:VerificationVideoID;constraint:OnDelete:SET NULL" json:"verification_video,omitempty"`
}

func (Listing) TableName() string { return "listings" }

// ListingAmenity is the junction table for listing<->amenity relationships.
// Table: listing_amenities
type ListingAmenity struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ListingID uuid.UUID `gorm:"type:uuid;not null;index:ix_listing_amenities_listing_id" json:"listing_id"`
	AmenityID string    `gorm:"type:varchar(100);not null;index:ix_listing_amenities_amenity_id" json:"amenity_id"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
}

func (ListingAmenity) TableName() string { return "listing_amenities" }

// AuditLog is an immutable administrative audit record.
type AuditLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AdminID    uuid.UUID `gorm:"type:uuid;not null;index:ix_audit_logs_admin_id" json:"admin_id"`
	AdminName  string    `gorm:"type:varchar(200);not null" json:"admin_name"`
	Action     string    `gorm:"type:varchar(100);not null;index:ix_audit_logs_action" json:"action"`
	TargetType string    `gorm:"type:varchar(50);not null;index:ix_audit_logs_target_type" json:"target_type"`
	TargetID   string    `gorm:"type:varchar(100);not null;index:ix_audit_logs_target_id" json:"target_id"`
	Reason     string    `gorm:"type:varchar(255)" json:"reason"`
	Note       string    `gorm:"type:text" json:"note"`
	OldStatus  string    `gorm:"type:varchar(50)" json:"old_status"`
	NewStatus  string    `gorm:"type:varchar(50)" json:"new_status"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// VerificationRequest is a legal entity or host verification application.
type VerificationRequest struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index:ix_verification_requests_user_id" json:"user_id"`
	ProviderType     string         `gorm:"type:varchar(50);not null" json:"provider_type"` // individual, individual_entrepreneur, self_employed, legal_entity
	LegalName        string         `gorm:"type:varchar(255);not null" json:"legal_name"`
	UNP              string         `gorm:"type:varchar(50)" json:"unp"`
	Status           string         `gorm:"type:varchar(50);not null;default:'pending';index:ix_verification_requests_status" json:"status"`
	RejectionReason  string         `gorm:"type:varchar(255)" json:"rejection_reason"`
	AdminNote        string         `gorm:"type:text" json:"admin_note"`
	RequestedChanges datatypes.JSON `gorm:"type:json" json:"requested_changes"`
	Requisites       datatypes.JSON `gorm:"type:json" json:"requisites"`
	Documents        datatypes.JSON `gorm:"type:json" json:"documents"`
	CreatedAt        time.Time      `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (VerificationRequest) TableName() string { return "verification_requests" }

// Report represents a user or listing complaint.
type Report struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReporterID       uuid.UUID  `gorm:"type:uuid;not null;index:ix_reports_reporter_id" json:"reporter_id"`
	TargetType       string     `gorm:"type:varchar(50);not null" json:"target_type"` // listing, user
	TargetID         string     `gorm:"type:varchar(100);not null" json:"target_id"`
	Reason           string     `gorm:"type:varchar(100);not null" json:"reason"`
	Description      string     `gorm:"type:text" json:"description"`
	Status           string     `gorm:"type:varchar(50);not null;default:'open';index:ix_reports_status" json:"status"`
	ResolutionReason string     `gorm:"type:varchar(255)" json:"resolution_reason"`
	ResolutionNote   string     `gorm:"type:text" json:"resolution_note"`
	ResolverAdminID  *uuid.UUID `gorm:"type:uuid" json:"resolver_admin_id"`
	CreatedAt        time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (Report) TableName() string { return "reports" }

// UserRestriction tracks administrative enforcement actions applied to a user.
type UserRestriction struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index:ix_user_restrictions_user_id" json:"user_id"`
	AdminID         uuid.UUID  `gorm:"type:uuid;not null" json:"admin_id"`
	EnforcementType string     `gorm:"type:varchar(50);not null" json:"enforcement_type"` // warning, temporary_restriction, temporary_block, permanent_block
	Reason          string     `gorm:"type:varchar(255);not null" json:"reason"`
	Note            string     `gorm:"type:text" json:"note"`
	ExpiresAt       *time.Time `gorm:"type:timestamptz" json:"expires_at"`
	IsActive        bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
}

func (UserRestriction) TableName() string { return "user_restrictions" }

// Notification represents an in-app user notification.
// Table: notifications
type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index:ix_notifications_user_id;index:ix_notifications_user_is_read,priority:1;index:ix_notifications_user_created_at,priority:1" json:"user_id"`
	Type      string         `gorm:"type:varchar(64);not null" json:"type"`
	Title     string         `gorm:"type:varchar(255);not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Payload   datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"payload"`
	IsRead    bool           `gorm:"not null;default:false;index:ix_notifications_user_is_read,priority:2" json:"is_read"`
	ReadAt    *time.Time     `gorm:"type:timestamptz" json:"read_at"`
	CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:now();index:ix_notifications_user_created_at,priority:2" json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }
