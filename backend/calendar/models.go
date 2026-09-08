// Package calendar provides models, services, iCal generation/parsing,
// and background workers for 2-way booking calendar synchronization.
package calendar

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReservationSource defines the origin of a calendar block/reservation.
type ReservationSource string

const (
	SourceFlickey     ReservationSource = "flickey"
	SourceManualBlock ReservationSource = "manual_block"
	SourceICalImport  ReservationSource = "ical_import"
)

// ReservationStatus defines the lifecycle status of a reservation.
type ReservationStatus string

const (
	StatusConfirmed ReservationStatus = "confirmed"
	StatusBlocked   ReservationStatus = "blocked"
	StatusCancelled ReservationStatus = "cancelled"
)

// SyncStatus defines the status of an external iCal feed sync.
type SyncStatus string

const (
	SyncStatusIdle    SyncStatus = "idle"
	SyncStatusSyncing SyncStatus = "syncing"
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
)

// ListingReservation represents a booked or blocked date interval for a listing.
// Dates are stored as a half-open interval [start_date, end_date),
// where start_date is the check-in date and end_date is the check-out date (exclusive).
// Table: listing_reservations
type ListingReservation struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ListingID   uuid.UUID         `gorm:"type:uuid;not null;index:ix_reservations_listing_id" json:"listing_id"`
	Source      ReservationSource `gorm:"type:varchar(30);not null;default:'flickey';index:ix_reservations_source" json:"source"`
	SyncFeedID  *uuid.UUID        `gorm:"type:uuid;index:ix_reservations_sync_feed_id" json:"sync_feed_id,omitempty"`
	StartDate   time.Time         `gorm:"type:date;not null;index:ix_reservations_start_date" json:"start_date"`
	EndDate     time.Time         `gorm:"type:date;not null;index:ix_reservations_end_date" json:"end_date"`
	Status      ReservationStatus `gorm:"type:varchar(30);not null;default:'confirmed';index:ix_reservations_status" json:"status"`
	GuestName   *string           `gorm:"type:varchar(150)" json:"guest_name,omitempty"`
	GuestPhone  *string           `gorm:"type:varchar(50)" json:"guest_phone,omitempty"`
	ExternalUID *string           `gorm:"type:varchar(255);index:ix_reservations_external_uid" json:"external_uid,omitempty"`
	Note        *string           `gorm:"type:text" json:"note,omitempty"`
	CreatedAt   time.Time         `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`

	// Relationship
	SyncFeed *ListingCalendarSync `gorm:"foreignKey:SyncFeedID" json:"sync_feed,omitempty"`
}

func (ListingReservation) TableName() string { return "listing_reservations" }

// ListingCalendarSync represents an external iCal subscription URL registered by a host.
// Table: listing_calendar_syncs
type ListingCalendarSync struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ListingID    uuid.UUID  `gorm:"type:uuid;not null;index:ix_calendar_syncs_listing_id" json:"listing_id"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	FeedURL      string     `gorm:"type:text;not null" json:"feed_url"`
	Color        string     `gorm:"type:varchar(30);not null;default:'indigo'" json:"color"`
	LastSyncedAt *time.Time `gorm:"type:timestamptz" json:"last_synced_at,omitempty"`
	SyncStatus   SyncStatus `gorm:"type:varchar(30);not null;default:'idle'" json:"sync_status"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ListingCalendarSync) TableName() string { return "listing_calendar_syncs" }

// AutoMigrateCalendar creates or updates calendar tables and indexes in PostgreSQL.
func AutoMigrateCalendar(db *gorm.DB) error {
	if err := db.AutoMigrate(&ListingCalendarSync{}, &ListingReservation{}); err != nil {
		return err
	}

	// 1. Unique index on (sync_feed_id, external_uid) for idempotent reconciliation
	_ = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_reservations_feed_uid 
		ON listing_reservations (sync_feed_id, external_uid) 
		WHERE sync_feed_id IS NOT NULL AND external_uid IS NOT NULL;
	`).Error

	// 2. Performance index for date-range overlap queries
	_ = db.Exec(`
		CREATE INDEX IF NOT EXISTS ix_reservations_listing_dates 
		ON listing_reservations (listing_id, start_date, end_date) 
		WHERE status != 'cancelled';
	`).Error

	// 3. Exclusion constraint for internal bookings and manual blocks (if btree_gist extension is available)
	_ = db.Exec(`CREATE EXTENSION IF NOT EXISTS btree_gist;`).Error
	_ = db.Exec(`
		DO $$ 
		BEGIN 
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'no_overlapping_internal_reservations'
			) THEN
				ALTER TABLE listing_reservations
				ADD CONSTRAINT no_overlapping_internal_reservations
				EXCLUDE USING gist (
					listing_id WITH =,
					daterange(start_date, end_date, '[)') WITH &&
				) WHERE (source IN ('flickey', 'manual_block') AND status != 'cancelled');
			END IF;
		EXCEPTION WHEN OTHERS THEN 
			-- If btree_gist is not permitted by DB permissions, row-level locks in service provide fallback.
			NULL;
		END $$;
	`).Error

	return nil
}
