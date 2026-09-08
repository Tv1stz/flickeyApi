package listings_test

import (
	"encoding/json"
	"testing"

	"flickey/go-backend/db"
	"flickey/go-backend/listings"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func TestRegistry_TotalSteps(t *testing.T) {
	reg := listings.DefaultRegistry
	if reg.TotalSteps() != 6 {
		t.Fatalf("expected 6 total steps, got %d", reg.TotalSteps())
	}
}

func TestRegistry_GetCompletedSteps_EmptyDraft(t *testing.T) {
	reg := listings.DefaultRegistry
	draft := &db.ListingDraft{
		ID:     uuid.New(),
		HostID: uuid.New(),
		Status: "draft",
	}

	completed := reg.GetCompletedSteps(draft)
	if len(completed) != 0 {
		t.Fatalf("expected 0 completed steps for empty draft, got %v", completed)
	}

	next := reg.CalculateNextStep(draft)
	if next != 1 {
		t.Fatalf("expected next step 1 for empty draft, got %d", next)
	}
}

func TestRegistry_GetCompletedSteps_Step1Only(t *testing.T) {
	reg := listings.DefaultRegistry
	apt := "apartment"
	draft := &db.ListingDraft{
		ID:     uuid.New(),
		HostID: uuid.New(),
		Status: "draft",
		Type:   &apt,
	}

	completed := reg.GetCompletedSteps(draft)
	if len(completed) != 1 || completed[0] != 1 {
		t.Fatalf("expected completed steps [1], got %v", completed)
	}

	next := reg.CalculateNextStep(draft)
	if next != 2 {
		t.Fatalf("expected next step 2, got %d", next)
	}

	// Step 2 can be accessed
	if !reg.CanAccessStep(draft, 2) {
		t.Fatalf("expected step 2 to be accessible")
	}

	// Step 3 cannot be accessed yet (anti-skipping)
	if reg.CanAccessStep(draft, 3) {
		t.Fatalf("expected step 3 to NOT be accessible when step 2 is incomplete")
	}
}

func TestRegistry_CanAccessStep_AntiSkipping(t *testing.T) {
	reg := listings.DefaultRegistry
	apt := "apartment"
	draft := &db.ListingDraft{
		ID:     uuid.New(),
		HostID: uuid.New(),
		Status: "draft",
		Type:   &apt,
	}

	// Attempting to access step 4, 5, 6 when only step 1 is done
	for step := 3; step <= 6; step++ {
		if reg.CanAccessStep(draft, step) {
			t.Errorf("expected CanAccessStep(draft, %d) to be false, got true", step)
		}
	}
}

func TestRegistry_CanAccessStep_AllowEditPreviousStep(t *testing.T) {
	reg := listings.DefaultRegistry
	apt := "apartment"
	name := "Modern Loft in Central Minsk"
	addr := "Independence Ave 10"
	lat := 53.9006
	lng := 27.5590
	sq := 60.0
	fl := 3
	tf := 9
	mg := 4
	rc := 2
	bc := 2
	bac := 1

	mediaIDs := []string{
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	}
	mediaJSON, _ := json.Marshal(mediaIDs)

	draft := &db.ListingDraft{
		ID:             uuid.New(),
		HostID:         uuid.New(),
		Status:         "draft",
		Type:           &apt,
		Name:           &name,
		Address:        &addr,
		Latitude:       &lat,
		Longitude:      &lng,
		Square:         &sq,
		Floor:          &fl,
		TotalFloors:    &tf,
		MaxGuests:      &mg,
		RoomsCount:     &rc,
		BedsCount:      &bc,
		BathroomsCount: &bac,
		MediaIDs:       datatypes.JSON(mediaJSON),
	}

	completed := reg.GetCompletedSteps(draft)
	if len(completed) != 3 {
		t.Fatalf("expected 3 completed steps (1, 2, 3), got %v", completed)
	}

	// Step 1, 2, 3, 4 should all be accessible
	if !reg.CanAccessStep(draft, 1) {
		t.Errorf("expected step 1 to be accessible for editing")
	}
	if !reg.CanAccessStep(draft, 2) {
		t.Errorf("expected step 2 to be accessible for editing")
	}
	if !reg.CanAccessStep(draft, 3) {
		t.Errorf("expected step 3 to be accessible for editing")
	}
	if !reg.CanAccessStep(draft, 4) {
		t.Errorf("expected step 4 to be accessible as next step")
	}

	// Step 5 should NOT be accessible
	if reg.CanAccessStep(draft, 5) {
		t.Errorf("expected step 5 to NOT be accessible before step 4 is completed")
	}
}

func TestRegistry_InvalidateDependents(t *testing.T) {
	reg := listings.DefaultRegistry
	apt := "apartment"
	name := "Modern Loft in Central Minsk"
	mediaJSON, _ := json.Marshal([]string{uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()})
	amenitiesJSON, _ := json.Marshal([]string{"wifi", "kitchen"})
	price := 150.0
	desc := "This is a wonderful 30-character or longer apartment description."

	draft := &db.ListingDraft{
		ID:          uuid.New(),
		HostID:      uuid.New(),
		Status:      "draft",
		Type:        &apt,
		Name:        &name,
		MediaIDs:    datatypes.JSON(mediaJSON),
		Amenities:   datatypes.JSON(amenitiesJSON),
		PricePerNight: &price,
		Description: &desc,
	}

	// Editing Step 2 invalidates steps 3, 4, 5, 6
	reg.InvalidateDependents(draft, 2)

	if draft.MediaIDs != nil {
		t.Errorf("expected MediaIDs (step 3) to be nil after step 2 invalidation")
	}
	if draft.Amenities != nil {
		t.Errorf("expected Amenities (step 4) to be nil after step 2 invalidation")
	}
	if draft.PricePerNight != nil {
		t.Errorf("expected PricePerNight (step 5) to be nil after step 2 invalidation")
	}
	if draft.Description != nil {
		t.Errorf("expected Description (step 6) to be nil after step 2 invalidation")
	}
}

func TestRegistry_SubmittedDraft_Locked(t *testing.T) {
	reg := listings.DefaultRegistry
	apt := "apartment"
	draft := &db.ListingDraft{
		ID:     uuid.New(),
		HostID: uuid.New(),
		Status: "submitted",
		Type:   &apt,
	}

	for step := 1; step <= 6; step++ {
		if reg.CanAccessStep(draft, step) {
			t.Errorf("expected step %d to NOT be accessible on submitted draft", step)
		}
	}
}
