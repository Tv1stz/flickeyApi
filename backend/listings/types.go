// Package listings provides request/response types for listing and draft endpoints.
package listings

import (
	"time"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// Draft Requests
// ─────────────────────────────────────────────────────────────────────────────

// DraftCreateRequest is the body for POST /listings/drafts (Step 1).
type DraftCreateRequest struct {
	Type string `json:"type" binding:"required"`
}

// DraftStep2Request is the body for PATCH /listings/drafts/{id}/step-2.
type DraftStep2Request struct {
	Name           string  `json:"name" binding:"required,min=10,max=100"`
	Address        string  `json:"address" binding:"required"`
	City           string  `json:"city"`
	Street         string  `json:"street"`
	HouseNumber    string  `json:"house_number"`
	Latitude       float64 `json:"latitude" binding:"required"`
	Longitude      float64 `json:"longitude" binding:"required"`
	Square         float64 `json:"square" binding:"required,gt=10,lt=1000"`
	Floor          int     `json:"floor" binding:"required,min=1,max=150"`
	TotalFloors    int     `json:"total_floors" binding:"required,min=1,max=150"`
	MaxGuests      int     `json:"max_guests" binding:"required,min=1,max=50"`
	RoomsCount     int     `json:"rooms_count" binding:"required,min=1,max=30"`
	BedsCount      int     `json:"beds_count" binding:"required,min=1,max=30"`
	BathroomsCount int     `json:"bathrooms_count" binding:"required,min=1,max=20"`
}

// DraftStep3Request is the body for PATCH /listings/drafts/{id}/step-3.
type DraftStep3Request struct {
	MediaIDs []uuid.UUID `json:"media_ids" binding:"required,min=5,max=25"`
}

// DraftStep4Request is the body for PATCH /listings/drafts/{id}/step-4.
type DraftStep4Request struct {
	Amenities []string `json:"amenities" binding:"required,min=1"`
}

// ListingRules represents house rules in step 5.
type ListingRules struct {
	AllowChildren   bool `json:"allow_children"`
	AllowPets       bool `json:"allow_pets"`
	AllowSmoking    bool `json:"allow_smoking"`
	AllowParties    bool `json:"allow_parties"`
	DepositRequired bool `json:"deposit_required"`
	WithInvoicing   bool `json:"with_invoicing"`
}

// DraftStep5Request is the body for PATCH /listings/drafts/{id}/step-5.
type DraftStep5Request struct {
	PricePerNight float64      `json:"price_per_night" binding:"required,gt=0"`
	Currency      string       `json:"currency" binding:"required"`
	MinNights     int          `json:"min_nights" binding:"required,min=1,max=30"`
	CheckinFrom   string       `json:"checkin_from" binding:"required"`   // "HH:MM" or "HH:MM:SS"
	CheckoutUntil string       `json:"checkout_until" binding:"required"` // "HH:MM" or "HH:MM:SS"
	Rules         ListingRules `json:"rules"`
}

// DraftStep6Request is the body for PATCH /listings/drafts/{id}/step-6.
type DraftStep6Request struct {
	Description string `json:"description" binding:"required,min=30,max=5000"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Draft Responses
// ─────────────────────────────────────────────────────────────────────────────

// DraftCreateResponse is returned after POST /listings/drafts.
type DraftCreateResponse struct {
	DraftID         uuid.UUID  `json:"draft_id"`
	SourceListingID *uuid.UUID `json:"source_listing_id,omitempty"`
	Mode            string     `json:"mode,omitempty"`
	CurrentStep     int        `json:"current_step"`
	CompletedSteps  []int      `json:"completed_steps"`
	TotalSteps      int        `json:"total_steps"`
	Status          string     `json:"status"`
}

// DraftStepResponse is returned after each draft step update.
type DraftStepResponse struct {
	DraftID        uuid.UUID `json:"draft_id"`
	CurrentStep    int       `json:"current_step"`
	CompletedSteps []int     `json:"completed_steps"`
	TotalSteps     int       `json:"total_steps"`
	Status         string    `json:"status"`
}

// DraftMediaItem contains an uploaded media item ID and its public URL.
type DraftMediaItem struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// DraftDetailResponse is returned for GET /listings/drafts/{id}.
type DraftDetailResponse struct {
	ID              uuid.UUID        `json:"id"`
	HostID          uuid.UUID        `json:"host_id"`
	SourceListingID *uuid.UUID       `json:"source_listing_id,omitempty"`
	Mode            string           `json:"mode"`
	CurrentStep     int              `json:"current_step"`
	CompletedSteps  []int            `json:"completed_steps"`
	TotalSteps      int              `json:"total_steps"`
	Status          string           `json:"status"`
	Type            *string          `json:"type"`
	Name            *string          `json:"name"`
	Address         *string          `json:"address"`
	City            *string          `json:"city"`
	Street          *string          `json:"street"`
	HouseNumber     *string          `json:"house_number"`
	Latitude        *float64         `json:"latitude"`
	Longitude       *float64         `json:"longitude"`
	Square          *float64         `json:"square"`
	Floor           *int             `json:"floor"`
	TotalFloors     *int             `json:"total_floors"`
	MaxGuests       *int             `json:"max_guests"`
	RoomsCount      *int             `json:"rooms_count"`
	BedsCount       *int             `json:"beds_count"`
	BathroomsCount  *int             `json:"bathrooms_count"`
	MediaIDs        []string         `json:"media_ids"`
	Media           []DraftMediaItem `json:"media"`
	Amenities       []string         `json:"amenities"`
	PricePerNight   *float64   `json:"price_per_night"`
	Currency        *string    `json:"currency"`
	MinNights       *int       `json:"min_nights"`
	CheckinFrom     *string    `json:"checkin_from"`
	CheckoutUntil   *string    `json:"checkout_until"`
	AllowChildren   *bool      `json:"allow_children"`
	AllowPets       *bool      `json:"allow_pets"`
	AllowSmoking    *bool      `json:"allow_smoking"`
	AllowParties    *bool      `json:"allow_parties"`
	DepositRequired *bool      `json:"deposit_required"`
	WithInvoicing   *bool      `json:"with_invoicing"`
	Description     *string    `json:"description"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ListingSubmitResponse is returned after POST /listings/drafts/{id}/submit.
type ListingSubmitResponse struct {
	ListingID         uuid.UUID  `json:"listing_id"`
	DraftID           *uuid.UUID `json:"draft_id,omitempty"`
	Status            string     `json:"status"`
	PublicationAction string     `json:"publication_action,omitempty"`
}


// ─────────────────────────────────────────────────────────────────────────────
// Listing Read Schemas
// ─────────────────────────────────────────────────────────────────────────────

// ListingHostReadSchema is the host-facing listing view (includes all fields).
type ListingHostReadSchema struct {
	ID              uuid.UUID `json:"id"`
	Status          string    `json:"status"`
	Type            string    `json:"type"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	City            string    `json:"city"`
	Street          string    `json:"street"`
	HouseNumber     string    `json:"house_number"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	Square          float64   `json:"square"`
	Floor           int       `json:"floor"`
	TotalFloors     int       `json:"total_floors"`
	MaxGuests       int       `json:"max_guests"`
	RoomsCount      int       `json:"rooms_count"`
	BedsCount       int       `json:"beds_count"`
	BathroomsCount  int       `json:"bathrooms_count"`
	PricePerNight   float64   `json:"price_per_night"`
	Currency        string    `json:"currency"`
	MinNights       int       `json:"min_nights"`
	CheckinFrom     string    `json:"checkin_from"`
	CheckoutUntil   string    `json:"checkout_until"`
	AllowChildren   bool      `json:"allow_children"`
	AllowPets       bool      `json:"allow_pets"`
	AllowSmoking    bool      `json:"allow_smoking"`
	AllowParties    bool      `json:"allow_parties"`
	DepositRequired bool      `json:"deposit_required"`
	WithInvoicing        bool       `json:"with_invoicing"`
	Description          string     `json:"description"`
	Amenities            []string   `json:"amenities"`
	Media                []string   `json:"media"`
	VerificationVideoURL *string    `json:"verification_video_url,omitempty"`
	VerificationVideoID  *uuid.UUID `json:"verification_video_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// AttachVerificationVideoRequest is the payload for POST /listings/:id/verification-video.
type AttachVerificationVideoRequest struct {
	MediaID uuid.UUID `json:"media_id" binding:"required"`
}

type HostSchema struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Role      string    `json:"role,omitempty"`
}

// ListingPublicSchema is the guest-facing published listing view.
type ListingPublicSchema struct {
	ID              uuid.UUID   `json:"id"`
	HostID          uuid.UUID   `json:"host_id,omitempty"`
	Host            *HostSchema `json:"host,omitempty"`
	Status          string      `json:"status,omitempty"`
	Type            string      `json:"type"`
	Name            string      `json:"name"`
	Address         string      `json:"address"`
	City            string      `json:"city"`
	Street          string      `json:"street"`
	HouseNumber     string      `json:"house_number"`
	Latitude        float64     `json:"latitude"`
	Longitude       float64     `json:"longitude"`
	Square          float64     `json:"square"`
	Floor           int         `json:"floor"`
	TotalFloors     int         `json:"total_floors"`
	MaxGuests       int         `json:"max_guests"`
	RoomsCount      int         `json:"rooms_count"`
	BedsCount       int         `json:"beds_count"`
	BathroomsCount  int         `json:"bathrooms_count"`
	PricePerNight   float64     `json:"price_per_night"`
	Currency        string      `json:"currency"`
	MinNights       int         `json:"min_nights"`
	CheckinFrom     string      `json:"checkin_from"`
	CheckoutUntil   string      `json:"checkout_until"`
	AllowChildren   bool        `json:"allow_children"`
	AllowPets       bool        `json:"allow_pets"`
	AllowSmoking    bool        `json:"allow_smoking"`
	AllowParties    bool        `json:"allow_parties"`
	DepositRequired bool        `json:"deposit_required"`
	WithInvoicing   bool        `json:"with_invoicing"`
	Description     string      `json:"description"`
	Amenities       []string    `json:"amenities"`
	Media           []string    `json:"media"`
	CreatedAt       time.Time   `json:"created_at"`
}

// PublicUserProfileSchema represents a host/user public profile.
type PublicUserProfileSchema struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	FirstName     *string   `json:"first_name,omitempty"`
	LastName      *string   `json:"last_name,omitempty"`
	Role          string    `json:"role"`
	IsVerified    bool      `json:"is_verified"`
	ListingsCount int       `json:"listings_count"`
	CreatedAt     time.Time `json:"created_at"`
}
