package calendar

import (
	"strings"
	"testing"
	"time"

	"flickey/go-backend/db"
	"github.com/google/uuid"
)

func TestGenerateICalFeed(t *testing.T) {
	listingID := uuid.New()
	listing := &db.Listing{
		ID:   listingID,
		Name: "Уютная студия в центре",
	}

	resID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	reservations := []ListingReservation{
		{
			ID:        resID,
			ListingID: listingID,
			Source:    SourceFlickey,
			StartDate: time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC), // 2 nights
			Status:    StatusConfirmed,
			UpdatedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:        uuid.New(),
			ListingID: listingID,
			Source:    SourceICalImport, // Should be ignored in export
			StartDate: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
			Status:    StatusConfirmed,
		},
	}

	feedBytes := GenerateICalFeed(listing, reservations)
	feed := string(feedBytes)

	if !strings.Contains(feed, "BEGIN:VCALENDAR") || !strings.Contains(feed, "END:VCALENDAR") {
		t.Fatal("feed missing calendar boundary tags")
	}

	if !strings.Contains(feed, "UID:flickey-res-11111111-2222-3333-4444-555555555555@flickey.by") {
		t.Errorf("feed missing expected deterministic UID: %s", feed)
	}

	if !strings.Contains(feed, "DTSTART;VALUE=DATE:20260710") {
		t.Errorf("feed missing DTSTART 20260710: %s", feed)
	}

	if !strings.Contains(feed, "DTEND;VALUE=DATE:20260712") {
		t.Errorf("feed missing DTEND 20260712: %s", feed)
	}

	// Verify imported event was excluded
	if strings.Contains(feed, "20260715") {
		t.Errorf("feed must not contain external imported events: %s", feed)
	}
}

func TestParseICalFeed(t *testing.T) {
	rawICal := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Avito//Calendar//RU
BEGIN:VEVENT
UID:avito-item-12345
DTSTART;VALUE=DATE:20260801
DTEND;VALUE=DATE:20260805
SUMMARY:Бронь на Авито
STATUS:CONFIRMED
END:VEVENT
BEGIN:VEVENT
UID:sutochno-long-summary-unfolded
DTSTART;TZID=Europe/Moscow:20260810T140000
DTEND;TZID=Europe/Moscow:20260812T120000
SUMMARY:Это очень длинная строка с перено
 сом строки по правилам RFC 5545
STATUS:CONFIRMED
END:VEVENT
BEGIN:VEVENT
UID:cancelled-booking
DTSTART;VALUE=DATE:20260820
DTEND;VALUE=DATE:20260825
STATUS:CANCELLED
END:VEVENT
END:VCALENDAR`

	events, err := ParseICalFeed(strings.NewReader(rawICal))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 active events (cancelled excluded), got %d", len(events))
	}

	// Check event 1
	ev1 := events[0]
	if ev1.UID != "avito-item-12345" {
		t.Errorf("unexpected UID: %s", ev1.UID)
	}
	if ev1.StartDate.Format("2006-01-02") != "2026-08-01" {
		t.Errorf("unexpected start date: %v", ev1.StartDate)
	}
	if ev1.EndDate.Format("2006-01-02") != "2026-08-05" {
		t.Errorf("unexpected end date: %v", ev1.EndDate)
	}

	// Check event 2 (unfolded line and datetime)
	ev2 := events[1]
	if ev2.UID != "sutochno-long-summary-unfolded" {
		t.Errorf("unexpected UID: %s", ev2.UID)
	}
	if ev2.StartDate.Format("2006-01-02") != "2026-08-10" {
		t.Errorf("unexpected start date: %v", ev2.StartDate)
	}
	if ev2.EndDate.Format("2006-01-02") != "2026-08-12" {
		t.Errorf("unexpected end date: %v", ev2.EndDate)
	}
	if !strings.Contains(ev2.Summary, "переносом") {
		t.Errorf("expected unfolded summary to be combined, got: %s", ev2.Summary)
	}
}
