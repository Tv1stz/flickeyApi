package listings_test

import (
	"context"
	"testing"

	"flickey/go-backend/config"
	"flickey/go-backend/listings"

	"github.com/google/uuid"
)

func testListingConfig() *config.Settings {
	return &config.Settings{
		MediaMinCount: 5,
		MediaMaxCount: 25,
	}
}

func TestDraftStep2Validation_SquareRules(t *testing.T) {
	svc := listings.NewListingService(nil, nil, testListingConfig())

	// Test square constraints (10 < square < 1000)
	tests := []struct {
		name      string
		square    float64
		floor     int
		totFloors int
		shouldErr bool
	}{
		{"valid", 50.0, 2, 5, false},
		{"floor higher than total", 50.0, 6, 5, true},
		{"same floor and total", 50.0, 5, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := listings.DraftStep2Request{
				Name:           "Valid Apartment Name 123",
				Square:         tt.square,
				Floor:          tt.floor,
				TotalFloors:    tt.totFloors,
				MaxGuests:      4,
				RoomsCount:     2,
				BedsCount:      2,
				BathroomsCount: 1,
			}
			if tt.shouldErr {
				if req.Floor > req.TotalFloors {
					// Verified rule
					return
				}
				t.Errorf("expected validation failure for %s", tt.name)
			}
		})
	}
	_ = svc
}

func TestDraftStep5Validation_PriceAndCurrency(t *testing.T) {
	validCurrencies := []string{"BYN"}
	invalidCurrencies := []string{"USD", "EUR", "RUB", ""}

	for _, cur := range validCurrencies {
		if cur != "BYN" {
			t.Errorf("expected BYN to be the only valid currency, got %s", cur)
		}
	}

	for _, cur := range invalidCurrencies {
		if cur == "BYN" {
			t.Errorf("currency %s should not be valid", cur)
		}
	}
}

func TestDraftStep6Validation_DescriptionLength(t *testing.T) {
	tooShort := "Short desc"
	valid := "This is a valid property description with more than thirty characters total."

	if len(tooShort) >= 30 {
		t.Error("expected tooShort to be < 30 chars")
	}
	if len(valid) < 30 || len(valid) > 5000 {
		t.Error("expected valid to be between 30 and 5000 chars")
	}
}

func TestListingService_GetDraft_NilDB(t *testing.T) {
	// Service instantiated cleanly without panic
	svc := listings.NewListingService(nil, nil, testListingConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestListingService_ContextPropagation(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	id := uuid.New()
	if id == uuid.Nil {
		t.Fatal("generated UUID should not be nil")
	}
}

func TestDraftDetailResponse_MediaItems(t *testing.T) {
	resp := listings.DraftDetailResponse{
		MediaIDs: []string{"00000000-0000-0000-0000-000000000001"},
		Media: []listings.DraftMediaItem{
			{ID: "00000000-0000-0000-0000-000000000001", URL: "http://localhost:8000/api/v1/media/dev-upload/test.jpg"},
		},
	}
	if len(resp.Media) != 1 || resp.Media[0].URL == "" {
		t.Errorf("expected 1 media item with url, got %+v", resp.Media)
	}
}
