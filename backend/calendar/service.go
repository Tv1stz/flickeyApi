package calendar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"flickey/go-backend/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrListingNotFound = errors.New("listing not found")
	ErrUnauthorized    = errors.New("unauthorized to manage this listing")
	ErrInvalidDates    = errors.New("check-out date must be after check-in date")
	ErrDateConflict    = errors.New("selected dates conflict with an existing booking or block")
	ErrReservationNotFound = errors.New("reservation not found")
)

// CalendarService manages reservations, availability, and iCal synchronizations.
type CalendarService struct {
	DB         *gorm.DB
	HTTPClient *http.Client
	BaseURL    string
}

// NewCalendarService creates a new CalendarService.
func NewCalendarService(database *gorm.DB, baseURL string) *CalendarService {
	client := NewSafeHTTPClient(10 * time.Second)
	return &CalendarService{
		DB:         database,
		HTTPClient: client,
		BaseURL:    strings.TrimRight(baseURL, "/"),
	}
}

// AvailabilityRange represents a booked/blocked range for public consumption.
type AvailabilityRange struct {
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`   // YYYY-MM-DD (exclusive)
}

// HostCalendarResponse provides full calendar details for the host dashboard.
type HostCalendarResponse struct {
	ListingID    uuid.UUID             `json:"listing_id"`
	ExportURL    string                `json:"export_url"`
	Reservations []ListingReservation  `json:"reservations"`
	SyncFeeds    []ListingCalendarSync `json:"sync_feeds"`
}

// GetICalFeed generates the RFC 5545 .ics export for a published listing.
func (s *CalendarService) GetICalFeed(ctx context.Context, listingID uuid.UUID) ([]byte, error) {
	var listing db.Listing
	if err := s.DB.WithContext(ctx).First(&listing, "id = ?", listingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}

	var reservations []ListingReservation
	err := s.DB.WithContext(ctx).
		Where("listing_id = ? AND status != ?", listingID, StatusCancelled).
		Order("start_date ASC").
		Find(&reservations).Error
	if err != nil {
		return nil, fmt.Errorf("fetching reservations: %w", err)
	}

	return GenerateICalFeed(&listing, reservations), nil
}

// GetPublicAvailability returns anonymous occupied ranges for public listing pages.
func (s *CalendarService) GetPublicAvailability(ctx context.Context, listingID uuid.UUID, from, to time.Time) ([]AvailabilityRange, error) {
	var reservations []ListingReservation
	query := s.DB.WithContext(ctx).
		Where("listing_id = ? AND status != ?", listingID, StatusCancelled)

	if !from.IsZero() {
		query = query.Where("end_date >= ?", from)
	}
	if !to.IsZero() {
		query = query.Where("start_date <= ?", to)
	}

	if err := query.Order("start_date ASC").Find(&reservations).Error; err != nil {
		return nil, fmt.Errorf("fetching availability: %w", err)
	}

	ranges := make([]AvailabilityRange, len(reservations))
	for i, r := range reservations {
		ranges[i] = AvailabilityRange{
			StartDate: r.StartDate.Format("2006-01-02"),
			EndDate:   r.EndDate.Format("2006-01-02"),
		}
	}

	return ranges, nil
}

// GetHostCalendar returns full reservations and sync status for a listing host.
func (s *CalendarService) GetHostCalendar(ctx context.Context, hostID, listingID uuid.UUID, from, to time.Time) (*HostCalendarResponse, error) {
	if err := s.verifyHostListing(ctx, hostID, listingID); err != nil {
		return nil, err
	}

	var reservations []ListingReservation
	query := s.DB.WithContext(ctx).
		Where("listing_id = ? AND status != ?", listingID, StatusCancelled)

	if !from.IsZero() {
		query = query.Where("end_date >= ?", from)
	}
	if !to.IsZero() {
		query = query.Where("start_date <= ?", to)
	}

	if err := query.Order("start_date ASC").Find(&reservations).Error; err != nil {
		return nil, fmt.Errorf("fetching host reservations: %w", err)
	}

	var syncFeeds []ListingCalendarSync
	if err := s.DB.WithContext(ctx).Where("listing_id = ?", listingID).Order("created_at DESC").Find(&syncFeeds).Error; err != nil {
		return nil, fmt.Errorf("fetching sync feeds: %w", err)
	}

	exportURL := fmt.Sprintf("%s/api/v1/listings/%s/calendar.ics", s.BaseURL, listingID.String())

	return &HostCalendarResponse{
		ListingID:    listingID,
		ExportURL:    exportURL,
		Reservations: reservations,
		SyncFeeds:    syncFeeds,
	}, nil
}

// BlockDates manually blocks a date interval for an owner.
func (s *CalendarService) BlockDates(ctx context.Context, hostID, listingID uuid.UUID, start, end time.Time, note string) (*ListingReservation, error) {
	if err := s.verifyHostListing(ctx, hostID, listingID); err != nil {
		return nil, err
	}

	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	if !end.After(start) {
		return nil, ErrInvalidDates
	}

	// Check conflict with internal bookings / manual blocks
	var conflictCount int64
	err := s.DB.WithContext(ctx).Model(&ListingReservation{}).
		Where("listing_id = ? AND status IN (?, ?) AND source IN (?, ?) AND start_date < ? AND end_date > ?",
			listingID, StatusConfirmed, StatusBlocked, SourceFlickey, SourceManualBlock, end, start).
		Count(&conflictCount).Error
	if err != nil {
		return nil, err
	}
	if conflictCount > 0 {
		return nil, ErrDateConflict
	}

	res := ListingReservation{
		ListingID: listingID,
		Source:    SourceManualBlock,
		StartDate: start,
		EndDate:   end,
		Status:    StatusBlocked,
		Note:      &note,
	}

	if err := s.DB.WithContext(ctx).Create(&res).Error; err != nil {
		return nil, fmt.Errorf("creating manual block: %w", err)
	}

	return &res, nil
}

// UnblockDates removes a manual block created by the host.
func (s *CalendarService) UnblockDates(ctx context.Context, hostID, listingID, reservationID uuid.UUID) error {
	if err := s.verifyHostListing(ctx, hostID, listingID); err != nil {
		return err
	}

	res := s.DB.WithContext(ctx).
		Where("id = ? AND listing_id = ? AND source = ?", reservationID, listingID, SourceManualBlock).
		Delete(&ListingReservation{})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrReservationNotFound
	}

	return nil
}

// AddSyncFeed registers a new external iCal subscription URL.
func (s *CalendarService) AddSyncFeed(ctx context.Context, hostID, listingID uuid.UUID, name, feedURL, color string) (*ListingCalendarSync, error) {
	if err := s.verifyHostListing(ctx, hostID, listingID); err != nil {
		return nil, err
	}

	parsedURL, err := ValidateURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid calendar URL: %w", err)
	}

	chosenColor := strings.TrimSpace(color)
	if chosenColor == "" {
		chosenColor = "indigo"
	}

	feed := ListingCalendarSync{
		ListingID:  listingID,
		Name:       strings.TrimSpace(name),
		FeedURL:    parsedURL.String(),
		Color:      chosenColor,
		SyncStatus: SyncStatusIdle,
		IsActive:   true,
	}

	if feed.Name == "" {
		feed.Name = "Внешний календарь"
	}

	if err := s.DB.WithContext(ctx).Create(&feed).Error; err != nil {
		return nil, fmt.Errorf("saving sync feed: %w", err)
	}

	// Trigger immediate initial sync in background
	go func() {
		_ = s.SyncFeed(context.Background(), feed.ID)
	}()

	return &feed, nil
}

// UpdateSyncFeed updates the name or display color of an external iCal subscription.
func (s *CalendarService) UpdateSyncFeed(ctx context.Context, hostID, listingID, syncID uuid.UUID, name, color *string) (*ListingCalendarSync, error) {
	if err := s.verifyHostListing(ctx, hostID, listingID); err != nil {
		return nil, err
	}

	var feed ListingCalendarSync
	if err := s.DB.WithContext(ctx).Where("id = ? AND listing_id = ?", syncID, listingID).First(&feed).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("sync feed not found")
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed != "" {
			updates["name"] = trimmed
			feed.Name = trimmed
		}
	}
	if color != nil {
		trimmedColor := strings.TrimSpace(*color)
		if trimmedColor != "" {
			updates["color"] = trimmedColor
			feed.Color = trimmedColor
		}
	}

	if len(updates) > 0 {
		if err := s.DB.WithContext(ctx).Model(&feed).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("updating sync feed: %w", err)
		}
	}

	return &feed, nil
}

// DeleteSyncFeed removes an external iCal subscription and purges its imported reservations.
func (s *CalendarService) DeleteSyncFeed(ctx context.Context, hostID, listingID, syncID uuid.UUID) error {
	if err := s.verifyHostListing(ctx, hostID, listingID); err != nil {
		return err
	}

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete imported reservations associated with this feed
		if err := tx.Where("sync_feed_id = ?", syncID).Delete(&ListingReservation{}).Error; err != nil {
			return err
		}

		// Delete the sync feed record
		res := tx.Where("id = ? AND listing_id = ?", syncID, listingID).Delete(&ListingCalendarSync{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("sync feed not found")
		}

		return nil
	})
}

// SyncFeed performs an idempotent reconciliation of an external iCal feed.
func (s *CalendarService) SyncFeed(ctx context.Context, feedID uuid.UUID) error {
	var feed ListingCalendarSync
	if err := s.DB.WithContext(ctx).First(&feed, "id = ?", feedID).Error; err != nil {
		return err
	}

	if !feed.IsActive {
		return nil
	}

	// Update status to syncing
	_ = s.DB.WithContext(ctx).Model(&feed).Update("sync_status", SyncStatusSyncing)

	// Fetch with Anti-SSRF protection
	data, fetchErr := FetchSafeURL(ctx, s.HTTPClient, feed.FeedURL)
	if fetchErr != nil {
		errMsg := fetchErr.Error()
		now := time.Now().UTC()
		_ = s.DB.WithContext(ctx).Model(&feed).Updates(map[string]interface{}{
			"sync_status":   SyncStatusFailed,
			"error_message": errMsg,
			"last_synced_at": now,
		})
		return fmt.Errorf("sync feed %s (%s): %w", feed.Name, feed.ID, fetchErr)
	}

	// Parse events
	events, parseErr := ParseICalFeed(bytes.NewReader(data))
	if parseErr != nil {
		errMsg := parseErr.Error()
		now := time.Now().UTC()
		_ = s.DB.WithContext(ctx).Model(&feed).Updates(map[string]interface{}{
			"sync_status":   SyncStatusFailed,
			"error_message": errMsg,
			"last_synced_at": now,
		})
		return fmt.Errorf("parse feed %s (%s): %w", feed.Name, feed.ID, parseErr)
	}

	// Idempotent Transactional Reconciliation
	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seenUIDs := make([]string, 0, len(events))

		for _, ev := range events {
			uid := ev.UID
			seenUIDs = append(seenUIDs, uid)

			res := ListingReservation{
				ListingID:   feed.ListingID,
				Source:      SourceICalImport,
				SyncFeedID:  &feed.ID,
				StartDate:   ev.StartDate,
				EndDate:     ev.EndDate,
				Status:      StatusConfirmed,
				ExternalUID: &uid,
				Note:        &ev.Summary,
			}

			// Upsert on conflict (sync_feed_id, external_uid)
			err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "sync_feed_id"},
					{Name: "external_uid"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"start_date",
					"end_date",
					"status",
					"note",
					"updated_at",
				}),
			}).Create(&res).Error

			if err != nil {
				return fmt.Errorf("upserting reservation %s: %w", uid, err)
			}
		}

		// Delete events that were removed from the external feed
		deleteQuery := tx.Where("sync_feed_id = ? AND source = ?", feed.ID, SourceICalImport)
		if len(seenUIDs) > 0 {
			deleteQuery = deleteQuery.Where("external_uid NOT IN ?", seenUIDs)
		}
		if err := deleteQuery.Delete(&ListingReservation{}).Error; err != nil {
			return fmt.Errorf("deleting stale reservations: %w", err)
		}

		// Update feed sync status to success
		now := time.Now().UTC()
		return tx.Model(&feed).Updates(map[string]interface{}{
			"sync_status":    SyncStatusSuccess,
			"error_message":  nil,
			"last_synced_at": now,
		}).Error
	})

	return txErr
}

// SyncListingFeeds synchronizes all active iCal feeds for a specific listing.
func (s *CalendarService) SyncListingFeeds(ctx context.Context, listingID uuid.UUID) error {
	var feeds []ListingCalendarSync
	err := s.DB.WithContext(ctx).
		Where("listing_id = ? AND is_active = true", listingID).
		Find(&feeds).Error
	if err != nil {
		return err
	}

	for _, f := range feeds {
		if err := s.SyncFeed(ctx, f.ID); err != nil {
			// Log and continue to sync other feeds
			continue
		}
	}

	return nil
}

func (s *CalendarService) verifyHostListing(ctx context.Context, hostID, listingID uuid.UUID) error {
	var count int64
	err := s.DB.WithContext(ctx).Model(&db.Listing{}).
		Where("id = ? AND host_id = ?", listingID, hostID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrUnauthorized
	}
	return nil
}
