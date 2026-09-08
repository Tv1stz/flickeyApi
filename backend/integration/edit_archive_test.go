package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"flickey/go-backend/db"
	"flickey/go-backend/listings"

	"github.com/google/uuid"
)

func TestListing_Edit_NoChange_And_WithChange(t *testing.T) {
	app := setupApp(t)
	_, hostToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)

	// 1. Create a full listing and mark it published in DB
	draftID := createDraftSteps1to6(t, app, hostToken)

	submitW := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Content-Type":    "application/json",
		"Idempotency-Key": uuid.New().String(),
	})
	if submitW.Code != http.StatusCreated {
		t.Fatalf("submit failed: %d %s", submitW.Code, submitW.Body.String())
	}
	var submitResp listings.ListingSubmitResponse
	_ = json.Unmarshal(submitW.Body.Bytes(), &submitResp)
	listingID := submitResp.ListingID

	// Mark as published in DB
	if err := app.Suite.DB.Model(&db.Listing{}).Where("id = ?", listingID).Update("status", "published").Error; err != nil {
		t.Fatalf("failed to mark listing published: %v", err)
	}

	// 2. Create edit draft from published listing
	w := app.DoAuth(http.MethodPost, "/api/v1/listings/drafts/from-listing/"+listingID.String()+"?mode=edit", nil, hostToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for edit draft creation, got %d: %s", w.Code, w.Body.String())
	}
	var editDraftResp listings.DraftCreateResponse
	_ = json.Unmarshal(w.Body.Bytes(), &editDraftResp)
	if editDraftResp.Mode != "edit" {
		t.Fatalf("expected mode edit, got %s", editDraftResp.Mode)
	}

	// 3. Submit edit draft without any changes -> should return NO_CHANGE
	submitEditW := app.Do("POST", "/api/v1/listings/drafts/"+editDraftResp.DraftID.String()+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Content-Type":    "application/json",
		"Idempotency-Key": uuid.New().String(),
	})
	if submitEditW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for submit edit without changes, got %d: %s", submitEditW.Code, submitEditW.Body.String())
	}
	var noChangeResp listings.ListingSubmitResponse
	_ = json.Unmarshal(submitEditW.Body.Bytes(), &noChangeResp)
	if noChangeResp.PublicationAction != "NO_CHANGE" {
		t.Fatalf("expected publication_action NO_CHANGE, got %s", noChangeResp.PublicationAction)
	}
	if noChangeResp.Status != "published" {
		t.Fatalf("expected status published, got %s", noChangeResp.Status)
	}

	// 4. Create another edit draft and modify the price (Step 5)
	w2 := app.DoAuth(http.MethodPost, "/api/v1/listings/drafts/from-listing/"+listingID.String()+"?mode=edit", nil, hostToken)
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w2.Code)
	}
	var editDraftResp2 listings.DraftCreateResponse
	_ = json.Unmarshal(w2.Body.Bytes(), &editDraftResp2)

	// Update Step 5 with new price
	step5Body := listings.DraftStep5Request{
		PricePerNight: 299.99,
		Currency:      "BYN",
		MinNights:     1,
		CheckinFrom:   "14:00",
		CheckoutUntil: "12:00",
		Rules: listings.ListingRules{
			AllowChildren: true,
		},
	}
	step5W := app.DoAuth(http.MethodPatch, "/api/v1/listings/drafts/"+editDraftResp2.DraftID.String()+"/step-5", step5Body, hostToken)
	if step5W.Code != http.StatusOK {
		t.Fatalf("expected 200 for step 5 update, got %d: %s", step5W.Code, step5W.Body.String())
	}

	// Submit edit draft with price change -> should return REQUIRE_MODERATION
	submitEditW2 := app.Do("POST", "/api/v1/listings/drafts/"+editDraftResp2.DraftID.String()+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Content-Type":    "application/json",
		"Idempotency-Key": uuid.New().String(),
	})
	if submitEditW2.Code != http.StatusCreated {
		t.Fatalf("expected 201 for submit edit with changes, got %d: %s", submitEditW2.Code, submitEditW2.Body.String())
	}
	var withChangeResp listings.ListingSubmitResponse
	_ = json.Unmarshal(submitEditW2.Body.Bytes(), &withChangeResp)
	if withChangeResp.PublicationAction != "REQUIRE_MODERATION" {
		t.Fatalf("expected publication_action REQUIRE_MODERATION, got %s", withChangeResp.PublicationAction)
	}
	if withChangeResp.Status != "pending_review" {
		t.Fatalf("expected status pending_review, got %s", withChangeResp.Status)
	}

	// Original public listing in DB is still published
	var l db.Listing
	_ = app.Suite.DB.Where("id = ?", listingID).First(&l)
	if l.Status != "published" {
		t.Fatalf("original listing status should still be published, got %s", l.Status)
	}
}

func TestListing_Archive_And_Unarchive(t *testing.T) {
	app := setupApp(t)
	_, hostToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)
	_, otherToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)

	// 1. Create a full listing and mark it published in DB
	draftID := createDraftSteps1to6(t, app, hostToken)
	submitW := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Content-Type":    "application/json",
		"Idempotency-Key": uuid.New().String(),
	})
	var submitResp listings.ListingSubmitResponse
	_ = json.Unmarshal(submitW.Body.Bytes(), &submitResp)
	listingID := submitResp.ListingID

	_ = app.Suite.DB.Model(&db.Listing{}).Where("id = ?", listingID).Update("status", "published")

	// 2. Anti-IDOR: other user cannot archive host's listing
	idorW := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"//archive", nil, otherToken)
	if idorW.Code == http.StatusOK {
		t.Fatalf("expected non-200 for IDOR archive, got %d", idorW.Code)
	}

	// 3. Host archives listing
	archW := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/archive", nil, hostToken)
	if archW.Code != http.StatusOK {
		t.Fatalf("expected 200 for archive, got %d: %s", archW.Code, archW.Body.String())
	}
	var archListing listings.ListingHostReadSchema
	_ = json.Unmarshal(archW.Body.Bytes(), &archListing)
	if archListing.Status != "archived" {
		t.Fatalf("expected archived status, got %s", archListing.Status)
	}

	// Verify archived listing is not in public catalog
	pubW := app.Do(http.MethodGet, "/api/v1/listings", nil, nil)
	var pubListings []listings.ListingPublicSchema
	_ = json.Unmarshal(pubW.Body.Bytes(), &pubListings)
	for _, pl := range pubListings {
		if pl.ID == listingID {
			t.Fatalf("archived listing should NOT appear in public catalog")
		}
	}

	// 4. Host unarchives listing without changes -> restored directly to published
	unarchW := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/unarchive", nil, hostToken)
	if unarchW.Code != http.StatusOK {
		t.Fatalf("expected 200 for unarchive, got %d: %s", unarchW.Code, unarchW.Body.String())
	}
	var unarchListing listings.ListingHostReadSchema
	_ = json.Unmarshal(unarchW.Body.Bytes(), &unarchListing)
	if unarchListing.Status != "published" {
		t.Fatalf("expected published status, got %s", unarchListing.Status)
	}

	// 5. If archived listing has submitted changes pending moderation -> unarchive routes through awaiting_company_verification
	// Archive again
	_ = app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/archive", nil, hostToken)

	// Create edit draft with changes and submit
	wDraft := app.DoAuth(http.MethodPost, "/api/v1/listings/drafts/from-listing/"+listingID.String()+"?mode=edit", nil, hostToken)
	var eDraft listings.DraftCreateResponse
	_ = json.Unmarshal(wDraft.Body.Bytes(), &eDraft)

	_ = app.DoAuth(http.MethodPatch, "/api/v1/listings/drafts/"+eDraft.DraftID.String()+"/step-5", listings.DraftStep5Request{
		PricePerNight: 999.0,
		Currency:      "BYN",
		MinNights:     1,
		CheckinFrom:   "14:00",
		CheckoutUntil: "12:00",
	}, hostToken)

	submitEditW3 := app.Do("POST", "/api/v1/listings/drafts/"+eDraft.DraftID.String()+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Content-Type":    "application/json",
		"Idempotency-Key": uuid.New().String(),
	})
	if submitEditW3.Code != http.StatusCreated {
		t.Fatalf("expected 201 for submit edit 3, got %d: %s", submitEditW3.Code, submitEditW3.Body.String())
	}

	// Unarchive -> should route through awaiting_company_verification (not directly published!)
	unarchW2 := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/unarchive", nil, hostToken)
	var unarchListing2 listings.ListingHostReadSchema
	_ = json.Unmarshal(unarchW2.Body.Bytes(), &unarchListing2)
	if unarchListing2.Status == "published" {
		t.Fatalf("unarchiving listing with pending changes MUST NOT be directly published")
	}
	if unarchListing2.Status != "awaiting_company_verification" {
		t.Fatalf("expected awaiting_company_verification, got %s", unarchListing2.Status)
	}
}

func TestListing_ArchiveBypassSecurity(t *testing.T) {
	app := setupApp(t)
	_, hostToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)

	// 1. Create a listing that starts in pending_review
	draftID := createDraftSteps1to6(t, app, hostToken)
	submitW := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Content-Type":    "application/json",
		"Idempotency-Key": uuid.New().String(),
	})
	if submitW.Code != http.StatusCreated {
		t.Fatalf("submit failed: %d %s", submitW.Code, submitW.Body.String())
	}
	var submitResp listings.ListingSubmitResponse
	_ = json.Unmarshal(submitW.Body.Bytes(), &submitResp)
	listingID := submitResp.ListingID

	// Listing is in pending_review. Attempting to archive it MUST fail with 400 CANNOT_ARCHIVE_UNPUBLISHED
	wArchivePending := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/archive", nil, hostToken)
	if wArchivePending.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when archiving pending_review listing, got %d: %s", wArchivePending.Code, wArchivePending.Body.String())
	}

	// 2. Set listing status to draft. Attempting to archive it MUST fail
	_ = app.Suite.DB.Model(&db.Listing{}).Where("id = ?", listingID).Update("status", "draft")
	wArchiveDraft := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/archive", nil, hostToken)
	if wArchiveDraft.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when archiving draft listing, got %d: %s", wArchiveDraft.Code, wArchiveDraft.Body.String())
	}

	// 3. Unarchiving a draft MUST set its status to pending_review, NOT published!
	wUnarchiveDraft := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/unarchive", nil, hostToken)
	if wUnarchiveDraft.Code != http.StatusOK {
		t.Fatalf("expected 200 on unarchiving draft, got %d: %s", wUnarchiveDraft.Code, wUnarchiveDraft.Body.String())
	}
	var unarchDraft listings.ListingHostReadSchema
	_ = json.Unmarshal(wUnarchiveDraft.Body.Bytes(), &unarchDraft)
	if unarchDraft.Status != "pending_review" {
		t.Fatalf("unarchived draft MUST have status pending_review, got %s", unarchDraft.Status)
	}

	// 4. Set listing status to rejected. Archiving it MUST fail
	_ = app.Suite.DB.Model(&db.Listing{}).Where("id = ?", listingID).Updates(map[string]any{
		"status":           "rejected",
		"rejection_reason": "Bad quality photos",
	})
	wArchiveRejected := app.DoAuth(http.MethodPost, "/api/v1/listings/"+listingID.String()+"/archive", nil, hostToken)
	if wArchiveRejected.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when archiving rejected listing, got %d: %s", wArchiveRejected.Code, wArchiveRejected.Body.String())
	}

	// 5. Updating a published listing:
	// Set to published first
	_ = app.Suite.DB.Model(&db.Listing{}).Where("id = ?", listingID).Updates(map[string]any{
		"status":           "published",
		"rejection_reason": nil,
	})

	// Host saves changes as draft
	saveDraftBody := listings.UpdateListingRequest{
		Name:        ptrString("Updated Title in Draft"),
		SaveAsDraft: ptrBool(true),
	}
	wEditDraft := app.DoAuth(http.MethodPatch, "/api/v1/listings/"+listingID.String(), saveDraftBody, hostToken)
	if wEditDraft.Code != http.StatusOK {
		t.Fatalf("expected 200 for update listing draft, got %d: %s", wEditDraft.Code, wEditDraft.Body.String())
	}
	var editDraftListing listings.ListingHostReadSchema
	_ = json.Unmarshal(wEditDraft.Body.Bytes(), &editDraftListing)
	if editDraftListing.Status != "draft" {
		t.Fatalf("expected draft status after saving as draft, got %s", editDraftListing.Status)
	}

	// Host submits draft for moderation
	submitModBody := listings.UpdateListingRequest{
		Name:                 ptrString("Updated Title for Review"),
		SubmitForModeration: ptrBool(true),
	}
	wEditMod := app.DoAuth(http.MethodPatch, "/api/v1/listings/"+listingID.String(), submitModBody, hostToken)
	if wEditMod.Code != http.StatusOK {
		t.Fatalf("expected 200 for submit for moderation, got %d: %s", wEditMod.Code, wEditMod.Body.String())
	}
	var editModListing listings.ListingHostReadSchema
	_ = json.Unmarshal(wEditMod.Body.Bytes(), &editModListing)
	if editModListing.Status != "pending_review" {
		t.Fatalf("expected pending_review status after submit for moderation, got %s", editModListing.Status)
	}
}

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }

