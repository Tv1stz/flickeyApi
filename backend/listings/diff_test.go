package listings

import (
	"encoding/json"
	"testing"
	"time"

	"flickey/go-backend/db"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func TestDeterminePublicationAction_NoChange(t *testing.T) {
	listingID := uuid.New()
	hostID := uuid.New()
	media1 := uuid.New()
	media2 := uuid.New()

	original := &db.Listing{
		ID:              listingID,
		HostID:          hostID,
		Status:          "published",
		Type:            "apartment",
		Name:            "Clean Test Apartment",
		Address:         "г. Минск, пр. Победителей, 1",
		City:            "Минск",
		Street:          "пр. Победителей",
		HouseNumber:     "1",
		Latitude:        53.9006,
		Longitude:       27.5590,
		Square:          55.0,
		Floor:           3,
		TotalFloors:     9,
		MaxGuests:       4,
		RoomsCount:      2,
		BedsCount:       2,
		BathroomsCount:  1,
		PricePerNight:   150.0,
		Currency:        "BYN",
		MinNights:       2,
		CheckinFrom:     "14:00:00",
		CheckoutUntil:   "12:00:00",
		AllowChildren:   true,
		AllowPets:       false,
		AllowSmoking:    false,
		AllowParties:    false,
		DepositRequired: false,
		WithInvoicing:   false,
		Description:     "A very nice and quiet place to stay in Minsk.",
		CreatedAt:       time.Now().Add(-24 * time.Hour),
		UpdatedAt:       time.Now().Add(-12 * time.Hour),
	}

	originalAmenities := []string{"wifi", "heating"}
	originalMedia := []db.Media{
		{ID: media1},
		{ID: media2},
	}

	mediaJSON, _ := json.Marshal([]string{media1.String(), media2.String()})
	amenitiesJSON, _ := json.Marshal([]string{"heating", "wifi"}) // Different order is OK

	typeStr := "apartment"
	nameStr := "Clean Test Apartment"
	addrStr := "г. Минск, пр. Победителей, 1"
	cityStr := "Минск"
	streetStr := "пр. Победителей"
	houseStr := "1"
	lat := 53.9006
	lng := 27.5590
	sq := 55.0
	floor := 3
	totalFloors := 9
	maxGuests := 4
	rooms := 2
	beds := 2
	baths := 1
	price := 150.0
	curr := "BYN"
	minNights := 2
	checkin := "14:00"
	checkout := "12:00"
	allowChildren := true
	allowPets := false
	allowSmoking := false
	allowParties := false
	depositReq := false
	withInvoicing := false
	desc := "A very nice and quiet place to stay in Minsk."

	draft := &db.ListingDraft{
		ID:              uuid.New(),
		HostID:          hostID,
		SourceListingID: &listingID,
		Mode:            "edit",
		Type:            &typeStr,
		Name:            &nameStr,
		Address:         &addrStr,
		City:            &cityStr,
		Street:          &streetStr,
		HouseNumber:     &houseStr,
		Latitude:        &lat,
		Longitude:       &lng,
		Square:          &sq,
		Floor:           &floor,
		TotalFloors:     &totalFloors,
		MaxGuests:       &maxGuests,
		RoomsCount:      &rooms,
		BedsCount:       &beds,
		BathroomsCount:  &baths,
		PricePerNight:   &price,
		Currency:        &curr,
		MinNights:       &minNights,
		CheckinFrom:     &checkin,
		CheckoutUntil:   &checkout,
		AllowChildren:   &allowChildren,
		AllowPets:       &allowPets,
		AllowSmoking:    &allowSmoking,
		AllowParties:    &allowParties,
		DepositRequired: &depositReq,
		WithInvoicing:   &withInvoicing,
		Description:     &desc,
		MediaIDs:        datatypes.JSON(mediaJSON),
		Amenities:       datatypes.JSON(amenitiesJSON),
	}

	action, changed := DeterminePublicationAction(original, draft, originalAmenities, originalMedia)
	if action != PublicationActionNoChange {
		t.Fatalf("expected NO_CHANGE, got %s with changes: %v", action, changed)
	}
}

func TestDeterminePublicationAction_PriceChanged(t *testing.T) {
	listingID := uuid.New()
	original := &db.Listing{
		ID:            listingID,
		PricePerNight: 100.0,
		Name:          "Apartment",
		Description:   "Description",
	}

	newPrice := 150.0
	draft := &db.ListingDraft{
		SourceListingID: &listingID,
		PricePerNight:   &newPrice,
		Name:            &original.Name,
		Description:     &original.Description,
	}

	action, changed := DeterminePublicationAction(original, draft, nil, nil)
	if action != PublicationActionRequireModeration {
		t.Fatalf("expected REQUIRE_MODERATION, got %s", action)
	}
	if len(changed) != 1 || changed[0] != "price_per_night" {
		t.Fatalf("expected changed fields [price_per_night], got %v", changed)
	}
}

func TestDeterminePublicationAction_AmenitiesChanged(t *testing.T) {
	listingID := uuid.New()
	original := &db.Listing{
		ID:   listingID,
		Name: "Apartment",
	}

	amenitiesJSON, _ := json.Marshal([]string{"wifi", "pool"})
	draft := &db.ListingDraft{
		SourceListingID: &listingID,
		Name:            &original.Name,
		Amenities:       datatypes.JSON(amenitiesJSON),
	}

	action, changed := DeterminePublicationAction(original, draft, []string{"wifi"}, nil)
	if action != PublicationActionRequireModeration {
		t.Fatalf("expected REQUIRE_MODERATION, got %s", action)
	}
	if len(changed) != 1 || changed[0] != "amenities" {
		t.Fatalf("expected changed fields [amenities], got %v", changed)
	}
}
