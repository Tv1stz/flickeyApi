// Package db provides all database query functions.
// Functions are organized by entity: User, Draft, Listing, Media, Verification.
package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─────────────────────────────────────────────────────────────────────────────
// UserStore
// ─────────────────────────────────────────────────────────────────────────────

// FindUserByID fetches a user by primary key.
func FindUserByID(db *gorm.DB, id uuid.UUID) (*User, error) {
	var u User
	if err := db.WithContext(context.Background()).First(&u, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("FindUserByID: %w", err)
	}
	return &u, nil
}

// FindUserByIDCtx fetches a user by primary key with context.
func FindUserByIDCtx(ctx context.Context, db *gorm.DB, id uuid.UUID) (*User, error) {
	var u User
	if err := db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("FindUserByIDCtx: %w", err)
	}
	return &u, nil
}

// FindUserByPhone fetches a user by phone number.
func FindUserByPhone(ctx context.Context, db *gorm.DB, phone string) (*User, error) {
	var u User
	if err := db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("FindUserByPhone: %w", err)
	}
	return &u, nil
}

// FindUserByEmail fetches a user by email address.
func FindUserByEmail(ctx context.Context, db *gorm.DB, email string) (*User, error) {
	var u User
	if err := db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("FindUserByEmail: %w", err)
	}
	return &u, nil
}

// CreateUser inserts a new user record.
func CreateUser(ctx context.Context, db *gorm.DB, u *User) error {
	if err := db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	return nil
}

// UpdateUserStatus updates only the status field.
func UpdateUserStatus(ctx context.Context, tx *gorm.DB, userID uuid.UUID, status UserStatus) error {
	result := tx.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("UpdateUserStatus: %w", result.Error)
	}
	return nil
}

// UpdateUserProfile updates first_name, last_name, email, and status atomically.
func UpdateUserProfile(ctx context.Context, tx *gorm.DB, userID uuid.UUID, firstName, lastName, email string) error {
	result := tx.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(map[string]any{
		"first_name": firstName,
		"last_name":  lastName,
		"email":      email,
		"status":     UserStatusActive,
	})
	if result.Error != nil {
		return fmt.Errorf("UpdateUserProfile: %w", result.Error)
	}
	return nil
}

// UpdateUserRole updates only the role field within the given transaction.
func UpdateUserRole(ctx context.Context, tx *gorm.DB, userID uuid.UUID, role UserRole) error {
	result := tx.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("role", role)
	if result.Error != nil {
		return fmt.Errorf("UpdateUserRole: %w", result.Error)
	}
	return nil
}

// LockUserByID fetches a user row with SELECT FOR UPDATE inside an existing transaction.
func LockUserByID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) (*User, error) {
	var u User
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("LockUserByID: %w", err)
	}
	return &u, nil
}

// FindListingByHostExists checks if at least one listing exists for this host.
func FindListingByHostExists(ctx context.Context, db *gorm.DB, hostID uuid.UUID) (bool, error) {
	var count int64
	if err := db.WithContext(ctx).Model(&Listing{}).Where("host_id = ?", hostID).Limit(1).Count(&count).Error; err != nil {
		return false, fmt.Errorf("FindListingByHostExists: %w", err)
	}
	return count > 0, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DraftStore
// ─────────────────────────────────────────────────────────────────────────────

// FindDraftByIDAndHost fetches a draft scoped by host_id (Anti-IDOR).
func FindDraftByIDAndHost(ctx context.Context, db *gorm.DB, draftID, hostID uuid.UUID) (*ListingDraft, error) {
	var d ListingDraft
	if err := db.WithContext(ctx).Where("id = ? AND host_id = ?", draftID, hostID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("FindDraftByIDAndHost: %w", err)
	}
	return &d, nil
}

// LockDraftByIDAndHost fetches a draft with SELECT FOR UPDATE inside a transaction.
func LockDraftByIDAndHost(ctx context.Context, tx *gorm.DB, draftID, hostID uuid.UUID) (*ListingDraft, error) {
	var d ListingDraft
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND host_id = ?", draftID, hostID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("LockDraftByIDAndHost: %w", err)
	}
	return &d, nil
}

// CreateDraft inserts a new draft.
func CreateDraft(ctx context.Context, db *gorm.DB, d *ListingDraft) error {
	if err := db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("CreateDraft: %w", err)
	}
	return nil
}

// SaveDraft persists all changes to an existing draft.
func SaveDraft(ctx context.Context, db *gorm.DB, d *ListingDraft) error {
	if err := db.WithContext(ctx).Save(d).Error; err != nil {
		return fmt.Errorf("SaveDraft: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ListingStore
// ─────────────────────────────────────────────────────────────────────────────

// CreateListing inserts a new listing and flushes to get the generated ID.
func CreateListing(ctx context.Context, tx *gorm.DB, l *Listing) error {
	if err := tx.WithContext(ctx).Create(l).Error; err != nil {
		return fmt.Errorf("CreateListing: %w", err)
	}
	return nil
}

// AddListingAmenities inserts amenity junction records in batch.
func AddListingAmenities(ctx context.Context, tx *gorm.DB, listingID uuid.UUID, amenityIDs []string) error {
	if len(amenityIDs) == 0 {
		return nil
	}
	records := make([]ListingAmenity, len(amenityIDs))
	for i, id := range amenityIDs {
		records[i] = ListingAmenity{
			ID:        uuid.New(),
			ListingID: listingID,
			AmenityID: id,
		}
	}
	if err := tx.WithContext(ctx).Create(&records).Error; err != nil {
		return fmt.Errorf("AddListingAmenities: %w", err)
	}
	return nil
}

// FindListingsByHost returns all listings for a host ordered by created_at desc.
func FindListingsByHost(ctx context.Context, db *gorm.DB, hostID uuid.UUID) ([]Listing, error) {
	var listings []Listing
	if err := db.WithContext(ctx).
		Preload("ListingAmenities").
		Preload("Media").
		Where("host_id = ?", hostID).
		Order("created_at DESC").
		Find(&listings).Error; err != nil {
		return nil, fmt.Errorf("FindListingsByHost: %w", err)
	}
	return listings, nil
}

// FindPublishedListings returns all listings with status = 'published'.
func FindPublishedListings(ctx context.Context, db *gorm.DB) ([]Listing, error) {
	var listings []Listing
	if err := db.WithContext(ctx).
		Preload("Host").
		Preload("ListingAmenities").
		Preload("Media").
		Where("status = ?", "published").
		Order("created_at DESC").
		Find(&listings).Error; err != nil {
		return nil, fmt.Errorf("FindPublishedListings: %w", err)
	}
	return listings, nil
}

// FindPublishedListingByID returns a single listing by ID.
func FindPublishedListingByID(ctx context.Context, db *gorm.DB, listingID uuid.UUID) (*Listing, error) {
	var l Listing
	if err := db.WithContext(ctx).
		Preload("Host").
		Preload("ListingAmenities").
		Preload("Media").
		Where("id = ?", listingID).
		First(&l).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("FindPublishedListingByID: %w", err)
	}
	return &l, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// MediaStore
// ─────────────────────────────────────────────────────────────────────────────

// CreateMedia inserts a new media record.
func CreateMedia(ctx context.Context, db *gorm.DB, m *Media) error {
	if err := db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("CreateMedia: %w", err)
	}
	return nil
}

// FindMediaByIDAndHost fetches a media record scoped by host_id (Anti-IDOR).
func FindMediaByIDAndHost(ctx context.Context, db *gorm.DB, mediaID, hostID uuid.UUID) (*Media, error) {
	var m Media
	if err := db.WithContext(ctx).Where("id = ? AND host_id = ?", mediaID, hostID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("FindMediaByIDAndHost: %w", err)
	}
	return &m, nil
}

// FindUploadedUnattachedMediaByIDsAndHost fetches media that are uploaded, unattached, and owned by host.
func FindUploadedUnattachedMediaByIDsAndHost(ctx context.Context, db *gorm.DB, mediaIDs []uuid.UUID, hostID uuid.UUID) ([]Media, error) {
	var media []Media
	if err := db.WithContext(ctx).Where(
		"id IN ? AND host_id = ? AND status = ? AND listing_id IS NULL",
		mediaIDs, hostID, MediaStatusUploaded,
	).Find(&media).Error; err != nil {
		return nil, fmt.Errorf("FindUploadedUnattachedMediaByIDsAndHost: %w", err)
	}
	return media, nil
}

// AttachMediaToListing bulk-updates media records to attached status.
// Uses raw SQL for an atomic rowcount check — critical for security.
// Returns the number of rows actually updated.
func AttachMediaToListing(ctx context.Context, tx *gorm.DB, mediaIDs []uuid.UUID, hostID, listingID uuid.UUID) (int64, error) {
	result := tx.WithContext(ctx).Exec(
		`UPDATE media
		 SET listing_id = ?, status = ?, updated_at = NOW()
		 WHERE id IN ?
		   AND host_id = ?
		   AND status = ?
		   AND listing_id IS NULL`,
		listingID, MediaStatusAttached, mediaIDs, hostID, MediaStatusUploaded,
	)
	if result.Error != nil {
		return 0, fmt.Errorf("AttachMediaToListing: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// UpdateMediaStatus updates a single media record's status.
func UpdateMediaStatus(ctx context.Context, db *gorm.DB, mediaID uuid.UUID, status MediaStatus, fileSizeBytes int64) error {
	result := db.WithContext(ctx).Model(&Media{}).Where("id = ?", mediaID).Updates(map[string]any{
		"status":          status,
		"file_size_bytes": fileSizeBytes,
	})
	if result.Error != nil {
		return fmt.Errorf("UpdateMediaStatus: %w", result.Error)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// VerificationStore
// ─────────────────────────────────────────────────────────────────────────────

// HasApprovedBusinessVerification checks if a host has an approved business verification.
// Currently stubbed to return false — listings always start as awaiting_company_verification.
// This matches the existing FastAPI behavior where the verifications table is not yet active.
func HasApprovedBusinessVerification(_ context.Context, _ *gorm.DB, _ uuid.UUID) (bool, error) {
	return false, nil
}
