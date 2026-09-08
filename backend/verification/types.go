// Package verification provides host verification DTOs and endpoints for Belarus partners.
package verification

import (
	"time"

	"github.com/google/uuid"
)

// SubmitVerificationRequest is the payload sent by hosts to submit Belarus requisites.
type SubmitVerificationRequest struct {
	ProviderType string         `json:"provider_type" binding:"required,oneof=individual self_employed individual_entrepreneur legal_entity"`
	LegalName    string         `json:"legal_name" binding:"required,min=2,max=255"`
	UNP          string         `json:"unp" binding:"required"`
	Requisites   map[string]any `json:"requisites" binding:"required"`
	Documents    []uuid.UUID    `json:"documents"`
}

// VerificationStatusResponse reports the host's current verification status and submitted requisites.
type VerificationStatusResponse struct {
	ID               *uuid.UUID     `json:"id,omitempty"`
	Status           string         `json:"status"` // none, pending, approved, rejected, changes_requested
	ProviderType     string         `json:"provider_type,omitempty"`
	LegalName        string         `json:"legal_name,omitempty"`
	UNP              string         `json:"unp,omitempty"`
	Requisites       map[string]any `json:"requisites,omitempty"`
	Documents        []string       `json:"documents,omitempty"`
	RejectionReason  string         `json:"rejection_reason,omitempty"`
	AdminNote        string         `json:"admin_note,omitempty"`
	CreatedAt        *time.Time     `json:"created_at,omitempty"`
	UpdatedAt        *time.Time     `json:"updated_at,omitempty"`
}
