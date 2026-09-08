// Package listings provides a configurable step registry for the listing creation wizard.
// The registry is the single source of truth for step ordering, completeness rules,
// and dependency invalidation. It contains no DB or Redis calls.
package listings

import (
	"encoding/json"
	"strings"

	"flickey/go-backend/db"
)

// StepConfig describes one wizard step.
type StepConfig struct {
	// ID is the step number (1-based, contiguous).
	ID int
	// Required means the step must be complete before later steps are accessible.
	Required bool
	// IsComplete reports whether the step's data on the draft is valid and persisted.
	// Derived from actual field values — never from current_step alone.
	IsComplete func(d *db.ListingDraft) bool
	// Invalidates lists step IDs whose data must be nulled when this step is updated.
	// Implements the "cascade downstream" rule: editing step N clears N+1..max.
	Invalidates []int
}

// StepRegistry is the ordered list of wizard steps.
type StepRegistry struct {
	Steps []StepConfig
}

// DefaultRegistry is the project-wide 6-step configuration.
// The number and order of steps is determined entirely here;
// no other file hardcodes the step count or sequence.
var DefaultRegistry = &StepRegistry{
	Steps: []StepConfig{
		{
			ID:          1,
			Required:    true,
			IsComplete:  step1Complete,
			Invalidates: nil, // Step 1 (type) is immutable after CreateDraft.
		},
		{
			ID:          2,
			Required:    true,
			IsComplete:  step2Complete,
			Invalidates: []int{3, 4, 5, 6},
		},
		{
			ID:          3,
			Required:    true,
			IsComplete:  step3Complete,
			Invalidates: []int{4, 5, 6},
		},
		{
			ID:          4,
			Required:    true,
			IsComplete:  step4Complete,
			Invalidates: []int{5, 6},
		},
		{
			ID:          5,
			Required:    true,
			IsComplete:  step5Complete,
			Invalidates: []int{6},
		},
		{
			ID:          6,
			Required:    true,
			IsComplete:  step6Complete,
			Invalidates: nil,
		},
	},
}

// nullFuncsByStep maps step ID → function that clears that step's draft fields.
// Invoked by InvalidateDependents when an earlier step is updated.
var nullFuncsByStep = map[int]func(*db.ListingDraft){
	3: func(d *db.ListingDraft) { d.MediaIDs = nil },
	4: func(d *db.ListingDraft) { d.Amenities = nil },
	5: func(d *db.ListingDraft) {
		d.PricePerNight = nil
		d.Currency = nil
		d.MinNights = nil
		d.CheckinFrom = nil
		d.CheckoutUntil = nil
		d.AllowChildren = nil
		d.AllowPets = nil
		d.AllowSmoking = nil
		d.AllowParties = nil
		d.DepositRequired = nil
		d.WithInvoicing = nil
	},
	6: func(d *db.ListingDraft) { d.Description = nil },
}

// ─────────────────────────────────────────────────────────────────────────────
// Registry methods
// ─────────────────────────────────────────────────────────────────────────────

// TotalSteps returns the total number of steps in the registry.
func (r *StepRegistry) TotalSteps() int { return len(r.Steps) }

// GetCompletedSteps returns the IDs of all steps whose data is complete in the draft.
// The result is derived from actual field values, not from draft.CurrentStep.
func (r *StepRegistry) GetCompletedSteps(draft *db.ListingDraft) []int {
	out := make([]int, 0, len(r.Steps))
	for _, s := range r.Steps {
		if s.IsComplete(draft) {
			out = append(out, s.ID)
		}
	}
	return out
}

// CanAccessStep reports whether the client is allowed to submit data for targetStepID.
//
// Rules:
//   - A submitted draft is always locked.
//   - Step 1 is always accessible (it is consumed by CreateDraft).
//   - For step N ≥ 2: every required step with ID < N must be complete.
func (r *StepRegistry) CanAccessStep(draft *db.ListingDraft, targetStepID int) bool {
	if draft.Status == "submitted" {
		return false
	}
	if targetStepID == 1 {
		return true
	}
	for _, s := range r.Steps {
		if s.ID >= targetStepID {
			break
		}
		if s.Required && !s.IsComplete(draft) {
			return false
		}
	}
	return true
}

// CalculateNextStep returns the ID of the first required step that is not yet complete.
// If all steps are complete, it returns the ID of the last step (ready to submit).
// The returned value is always in the range [1, TotalSteps].
func (r *StepRegistry) CalculateNextStep(draft *db.ListingDraft) int {
	for _, s := range r.Steps {
		if s.Required && !s.IsComplete(draft) {
			return s.ID
		}
	}
	// All steps complete: stay on the last step (submission screen).
	return r.Steps[len(r.Steps)-1].ID
}

// InvalidateDependents nulls the draft fields belonging to all steps listed
// in the Invalidates slice of the changedStepID. Called immediately after the
// new data for changedStepID has been applied to the draft (before Save).
func (r *StepRegistry) InvalidateDependents(draft *db.ListingDraft, changedStepID int) {
	// In edit mode (editing an existing listing), previous complete steps are already valid and
	// modifying a specific step (e.g. price or address) does not destroy existing photos/description.
	if draft != nil && draft.Mode == "edit" {
		return
	}
	for _, s := range r.Steps {
		if s.ID == changedStepID {
			for _, inv := range s.Invalidates {
				if fn, ok := nullFuncsByStep[inv]; ok {
					fn(draft)
				}
			}
			return
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Per-step IsComplete predicates
// ─────────────────────────────────────────────────────────────────────────────

func step1Complete(d *db.ListingDraft) bool {
	if d.Type == nil {
		return false
	}
	t := *d.Type
	return t == "apartment" || t == "house" || t == "manor"
}

func step2Complete(d *db.ListingDraft) bool {
	if d.Name == nil || len(*d.Name) < 10 || len(*d.Name) > 100 {
		return false
	}
	if d.Address == nil || len(strings.TrimSpace(*d.Address)) == 0 {
		return false
	}
	if d.Latitude == nil || *d.Latitude == 0 {
		return false
	}
	if d.Longitude == nil || *d.Longitude == 0 {
		return false
	}
	if d.Square == nil || *d.Square <= 10 || *d.Square >= 1000 {
		return false
	}
	if d.Floor == nil || d.TotalFloors == nil || *d.Floor > *d.TotalFloors {
		return false
	}
	if d.MaxGuests == nil || d.RoomsCount == nil || d.BedsCount == nil || d.BathroomsCount == nil {
		return false
	}
	return true
}

const step3MinMedia = 5

func step3Complete(d *db.ListingDraft) bool {
	if d.MediaIDs == nil {
		return false
	}
	var ids []string
	if err := json.Unmarshal(d.MediaIDs, &ids); err != nil {
		return false
	}
	return len(ids) >= step3MinMedia
}

func step4Complete(d *db.ListingDraft) bool {
	if d.Amenities == nil {
		return false
	}
	var ams []string
	if err := json.Unmarshal(d.Amenities, &ams); err != nil {
		return false
	}
	return len(ams) >= 1
}

func step5Complete(d *db.ListingDraft) bool {
	if d.PricePerNight == nil || *d.PricePerNight <= 0 {
		return false
	}
	if d.Currency == nil || *d.Currency != "BYN" {
		return false
	}
	if d.MinNights == nil {
		return false
	}
	if d.CheckinFrom == nil || d.CheckoutUntil == nil {
		return false
	}
	if d.AllowChildren == nil || d.AllowPets == nil || d.AllowSmoking == nil ||
		d.AllowParties == nil || d.DepositRequired == nil || d.WithInvoicing == nil {
		return false
	}
	return true
}

func step6Complete(d *db.ListingDraft) bool {
	return d.Description != nil && len(*d.Description) >= 30
}
