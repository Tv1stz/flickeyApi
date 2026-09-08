// Package listings provides listing draft management, validation, and multi-step progression.
package listings

import (
	"encoding/json"
	"math"
	"sort"
	"strings"

	"flickey/go-backend/db"
)

// PublicationAction represents the moderation decision upon draft submission.
type PublicationAction string

const (
	PublicationActionNoChange          PublicationAction = "NO_CHANGE"
	PublicationActionRequireModeration PublicationAction = "REQUIRE_MODERATION"
)

// DeterminePublicationAction compares a draft against its original listing to detect meaningful changes.
// Returns (PublicationAction, list of changed field names).
func DeterminePublicationAction(
	original *db.Listing,
	draft *db.ListingDraft,
	originalAmenities []string,
	originalMedia []db.Media,
) (PublicationAction, []string) {
	if original == nil || draft == nil {
		return PublicationActionRequireModeration, []string{"all"}
	}

	var changedFields []string

	// 1. Housing Type
	if draft.Type != nil && *draft.Type != original.Type {
		changedFields = append(changedFields, "type")
	}

	// 2. Name
	if draft.Name != nil && strings.TrimSpace(*draft.Name) != strings.TrimSpace(original.Name) {
		changedFields = append(changedFields, "name")
	}

	// 3. Description
	if draft.Description != nil && strings.TrimSpace(*draft.Description) != strings.TrimSpace(original.Description) {
		changedFields = append(changedFields, "description")
	}

	// 4. Address & Location
	if draft.Address != nil && strings.TrimSpace(*draft.Address) != strings.TrimSpace(original.Address) {
		changedFields = append(changedFields, "address")
	}
	if draft.City != nil && strings.TrimSpace(*draft.City) != strings.TrimSpace(original.City) {
		changedFields = append(changedFields, "city")
	}
	if draft.Street != nil && strings.TrimSpace(*draft.Street) != strings.TrimSpace(original.Street) {
		changedFields = append(changedFields, "street")
	}
	if draft.HouseNumber != nil && strings.TrimSpace(*draft.HouseNumber) != strings.TrimSpace(original.HouseNumber) {
		changedFields = append(changedFields, "house_number")
	}
	if draft.Latitude != nil && math.Abs(*draft.Latitude-original.Latitude) > 0.0001 {
		changedFields = append(changedFields, "latitude")
	}
	if draft.Longitude != nil && math.Abs(*draft.Longitude-original.Longitude) > 0.0001 {
		changedFields = append(changedFields, "longitude")
	}

	// 5. Physical Parameters
	if draft.Square != nil && math.Abs(*draft.Square-original.Square) > 0.01 {
		changedFields = append(changedFields, "square")
	}
	if draft.Floor != nil && *draft.Floor != original.Floor {
		changedFields = append(changedFields, "floor")
	}
	if draft.TotalFloors != nil && *draft.TotalFloors != original.TotalFloors {
		changedFields = append(changedFields, "total_floors")
	}
	if draft.MaxGuests != nil && *draft.MaxGuests != original.MaxGuests {
		changedFields = append(changedFields, "max_guests")
	}
	if draft.RoomsCount != nil && *draft.RoomsCount != original.RoomsCount {
		changedFields = append(changedFields, "rooms_count")
	}
	if draft.BedsCount != nil && *draft.BedsCount != original.BedsCount {
		changedFields = append(changedFields, "beds_count")
	}
	if draft.BathroomsCount != nil && *draft.BathroomsCount != original.BathroomsCount {
		changedFields = append(changedFields, "bathrooms_count")
	}

	// 6. Pricing & Currency
	if draft.PricePerNight != nil && math.Abs(*draft.PricePerNight-original.PricePerNight) > 0.01 {
		changedFields = append(changedFields, "price_per_night")
	}
	if draft.Currency != nil && *draft.Currency != original.Currency {
		changedFields = append(changedFields, "currency")
	}
	if draft.MinNights != nil && *draft.MinNights != original.MinNights {
		changedFields = append(changedFields, "min_nights")
	}

	// 7. Check-in / Check-out
	if draft.CheckinFrom != nil {
		dCheckin := normalizeTimeStr(*draft.CheckinFrom)
		oCheckin := normalizeTimeStr(original.CheckinFrom)
		if dCheckin != oCheckin {
			changedFields = append(changedFields, "checkin_from")
		}
	}
	if draft.CheckoutUntil != nil {
		dCheckout := normalizeTimeStr(*draft.CheckoutUntil)
		oCheckout := normalizeTimeStr(original.CheckoutUntil)
		if dCheckout != oCheckout {
			changedFields = append(changedFields, "checkout_until")
		}
	}

	// 8. House Rules
	if draft.AllowChildren != nil && *draft.AllowChildren != original.AllowChildren {
		changedFields = append(changedFields, "allow_children")
	}
	if draft.AllowPets != nil && *draft.AllowPets != original.AllowPets {
		changedFields = append(changedFields, "allow_pets")
	}
	if draft.AllowSmoking != nil && *draft.AllowSmoking != original.AllowSmoking {
		changedFields = append(changedFields, "allow_smoking")
	}
	if draft.AllowParties != nil && *draft.AllowParties != original.AllowParties {
		changedFields = append(changedFields, "allow_parties")
	}
	if draft.DepositRequired != nil && *draft.DepositRequired != original.DepositRequired {
		changedFields = append(changedFields, "deposit_required")
	}
	if draft.WithInvoicing != nil && *draft.WithInvoicing != original.WithInvoicing {
		changedFields = append(changedFields, "with_invoicing")
	}

	// 9. Amenities Comparison
	var draftAmenities []string
	if draft.Amenities != nil {
		_ = json.Unmarshal(draft.Amenities, &draftAmenities)
	}
	if !slicesEqualUnordered(originalAmenities, draftAmenities) {
		changedFields = append(changedFields, "amenities")
	}

	// 10. Media Comparison (order and identity)
	var draftMediaIDs []string
	if draft.MediaIDs != nil {
		_ = json.Unmarshal(draft.MediaIDs, &draftMediaIDs)
	}
	originalMediaIDs := make([]string, 0, len(originalMedia))
	for _, m := range originalMedia {
		originalMediaIDs = append(originalMediaIDs, m.ID.String())
	}
	if !slicesEqualOrdered(originalMediaIDs, draftMediaIDs) {
		changedFields = append(changedFields, "media")
	}

	if len(changedFields) == 0 {
		return PublicationActionNoChange, nil
	}

	return PublicationActionRequireModeration, changedFields
}

func slicesEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ca := make([]string, len(a))
	copy(ca, a)
	sort.Strings(ca)

	cb := make([]string, len(b))
	copy(cb, b)
	sort.Strings(cb)

	for i := range ca {
		if ca[i] != cb[i] {
			return false
		}
	}
	return true
}

func slicesEqualOrdered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
