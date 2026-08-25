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
	MediaIDs []uuid.UUID `json:"media_ids" binding:"required,min=5,max=15"`
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
	DraftID     uuid.UUID `json:"draft_id"`
	CurrentStep int       `json:"current_step"`
}

// DraftStepResponse is returned after each draft step update.
type DraftStepResponse struct {
	DraftID     uuid.UUID `json:"draft_id"`
	CurrentStep int       `json:"current_step"`
	Status      string    `json:"status"`
}

// DraftDetailResponse is returned for GET /listings/drafts/{id}.
type DraftDetailResponse struct {
	ID              uuid.UUID `json:"id"`
	HostID          uuid.UUID `json:"host_id"`
	CurrentStep     int       `json:"current_step"`
	Status          string    `json:"status"`
	Type            *string   `json:"type"`
	Name            *string   `json:"name"`
	Square          *float64  `json:"square"`
	Floor           *int      `json:"floor"`
	TotalFloors     *int      `json:"total_floors"`
	MaxGuests       *int      `json:"max_guests"`
	RoomsCount      *int      `json:"rooms_count"`
	BedsCount       *int      `json:"beds_count"`
	BathroomsCount  *int      `json:"bathrooms_count"`
	MediaIDs        []string  `json:"media_ids"`
	Amenities       []string  `json:"amenities"`
	PricePerNight   *float64  `json:"price_per_night"`
	Currency        *string   `json:"currency"`
	MinNights       *int      `json:"min_nights"`
	CheckinFrom     *string   `json:"checkin_from"`
	CheckoutUntil   *string   `json:"checkout_until"`
	AllowChildren   *bool     `json:"allow_children"`
	AllowPets       *bool     `json:"allow_pets"`
	AllowSmoking    *bool     `json:"allow_smoking"`
	AllowParties    *bool     `json:"allow_parties"`
	DepositRequired *bool     `json:"deposit_required"`
	WithInvoicing   *bool     `json:"with_invoicing"`
	Description     *string   `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ListingSubmitResponse is returned after POST /listings/drafts/{id}/submit.
type ListingSubmitResponse struct {
	ListingID uuid.UUID `json:"listing_id"`
	Status    string    `json:"status"`
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
	WithInvoicing   bool      `json:"with_invoicing"`
	Description     string    `json:"description"`
	Amenities       []string  `json:"amenities"`
	Media           []string  `json:"media"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
	Host            *HostSchema `json:"host,omitempty"`
	Type            string      `json:"type"`
	Name            string      `json:"name"`
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
