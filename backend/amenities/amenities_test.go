package amenities_test

import (
	"testing"

	"flickey/go-backend/amenities"
)

func TestRegistry_AllAmenitiiesHaveCategory(t *testing.T) {
	for id, a := range amenities.Registry {
		if a.Category == "" {
			t.Errorf("amenity %q has empty category", id)
		}
		if a.DisplayName == "" {
			t.Errorf("amenity %q has empty display name", id)
		}
		if a.ID != id {
			t.Errorf("amenity ID mismatch: key=%q, ID=%q", id, a.ID)
		}
	}
}

func TestRegistry_AllowedTypes(t *testing.T) {
	for id, a := range amenities.Registry {
		if len(a.AllowedTypes) == 0 {
			t.Errorf("amenity %q has no allowed types", id)
		}
	}
}

func TestGetAmenity_Found(t *testing.T) {
	known := []string{"wifi", "heating", "parking", "pets_allowed", "smoke_detector"}
	for _, id := range known {
		a, ok := amenities.GetAmenity(id)
		if !ok {
			t.Errorf("GetAmenity(%q) returned not found", id)
		}
		if a.ID != id {
			t.Errorf("GetAmenity(%q) returned wrong ID: %q", id, a.ID)
		}
	}
}

func TestGetAmenity_NotFound(t *testing.T) {
	_, ok := amenities.GetAmenity("does_not_exist")
	if ok {
		t.Error("expected not found for unknown amenity")
	}
}

func TestIsValidHousingType(t *testing.T) {
	valid := []string{"apartment", "house", "manor"}
	for _, t_ := range valid {
		if !amenities.IsValidHousingType(t_) {
			t.Errorf("expected %q to be valid", t_)
		}
	}
	invalid := []string{"condo", "APARTMENT", "", "villa"}
	for _, inv := range invalid {
		if amenities.IsValidHousingType(inv) {
			t.Errorf("expected %q to be invalid", inv)
		}
	}
}

func TestGetGroupedAmenities_All(t *testing.T) {
	groups := amenities.GetGroupedAmenities("")
	if len(groups) == 0 {
		t.Error("expected at least one category")
	}
	total := 0
	for _, g := range groups {
		total += len(g.Amenities)
	}
	if total != len(amenities.Registry) {
		t.Errorf("expected %d amenities total, got %d", len(amenities.Registry), total)
	}
}

func TestGetGroupedAmenities_FilterByType(t *testing.T) {
	// All amenities are allowed for all types in the registry.
	groups := amenities.GetGroupedAmenities("apartment")
	if len(groups) == 0 {
		t.Error("expected categories for apartment type")
	}
}

func TestGetGroupedAmenities_InvalidTypeEmpty(t *testing.T) {
	// Invalid housing type: GetGroupedAmenities should handle it gracefully
	// (it returns all if none match — but since we don't filter unknown types, test that it doesn't panic).
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetGroupedAmenities panicked on invalid type: %v", r)
		}
	}()
	amenities.GetGroupedAmenities("invalid_type")
}

func TestAmenityIsAllowedFor(t *testing.T) {
	wifi, _ := amenities.GetAmenity("wifi")
	for ht := range amenities.AllHousingTypes {
		if !wifi.IsAllowedFor(ht) {
			t.Errorf("wifi should be allowed for %q", ht)
		}
	}
}

func TestGroupedAmenities_NoDuplicates(t *testing.T) {
	groups := amenities.GetGroupedAmenities("")
	seen := map[string]struct{}{}
	for _, g := range groups {
		for _, a := range g.Amenities {
			if _, dup := seen[a.ID]; dup {
				t.Errorf("duplicate amenity %q in grouped response", a.ID)
			}
			seen[a.ID] = struct{}{}
		}
	}
}
