package calendar

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"flickey/go-backend/db"
)

// ICalEvent represents a parsed event from an external iCal feed.
type ICalEvent struct {
	UID       string
	StartDate time.Time // Check-in date (inclusive)
	EndDate   time.Time // Check-out date (exclusive)
	Summary   string
}

// GenerateICalFeed converts a listing and its active reservations into RFC 5545 compliant iCalendar text.
// Sync-loop protection: only internal reservations (flickey & manual_block) are exported.
func GenerateICalFeed(listing *db.Listing, reservations []ListingReservation) []byte {
	var buf bytes.Buffer

	listingName := "Жильё"
	if listing != nil && listing.Name != "" {
		listingName = sanitizeICalText(listing.Name)
	}

	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//Flickey//Booking Calendar 1.0//RU\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")
	buf.WriteString(fmt.Sprintf("X-WR-CALNAME:Flickey - %s\r\n", listingName))

	for _, res := range reservations {
		// Never export imported external events (prevents sync loops)
		if res.Source == SourceICalImport || res.Status == StatusCancelled {
			continue
		}

		// Ensure proper half-open interval
		start := res.StartDate.UTC()
		end := res.EndDate.UTC()
		if !end.After(start) {
			end = start.AddDate(0, 0, 1)
		}

		dtStamp := res.UpdatedAt.UTC().Format("20060102T150405Z")
		dtStart := start.Format("20060102")
		dtEnd := end.Format("20060102") // RFC 5545: VALUE=DATE is exclusive check-out date

		summary := "Занято (Flickey)"
		if res.Source == SourceManualBlock {
			summary = "Недоступно для бронирования"
		}

		buf.WriteString("BEGIN:VEVENT\r\n")
		// Deterministic, immutable UID based on reservation UUID
		buf.WriteString(fmt.Sprintf("UID:flickey-res-%s@flickey.by\r\n", res.ID.String()))
		buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", dtStamp))
		buf.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", dtStart))
		buf.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", dtEnd))
		buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", sanitizeICalText(summary)))
		buf.WriteString("STATUS:CONFIRMED\r\n")
		buf.WriteString("TRANSP:OPAQUE\r\n")
		buf.WriteString("END:VEVENT\r\n")
	}

	buf.WriteString("END:VCALENDAR\r\n")
	return buf.Bytes()
}

// ParseICalFeed reads an iCalendar data stream and extracts booking events.
// It handles RFC 5545 line unfolding, date/datetime formatting, and half-open intervals.
func ParseICalFeed(r io.Reader) ([]ICalEvent, error) {
	scanner := bufio.NewScanner(r)
	// Support up to 64KB lines
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 64*1024)

	// Step 1: Unfold lines (lines beginning with space or tab are continuation of previous line)
	var unfoldedLines []string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r\n")
		if len(line) == 0 {
			continue
		}

		if (line[0] == ' ' || line[0] == '\t') && len(unfoldedLines) > 0 {
			unfoldedLines[len(unfoldedLines)-1] += line[1:]
		} else {
			unfoldedLines = append(unfoldedLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading iCal stream: %w", err)
	}

	// Step 2: Parse VEVENT blocks
	var events []ICalEvent
	var inEvent bool
	var currentUID string
	var currentSummary string
	var currentStart, currentEnd *time.Time
	var currentStatus string

	for _, line := range unfoldedLines {
		upper := strings.ToUpper(strings.TrimSpace(line))

		if upper == "BEGIN:VEVENT" {
			inEvent = true
			currentUID = ""
			currentSummary = ""
			currentStart = nil
			currentEnd = nil
			currentStatus = "CONFIRMED"
			continue
		}

		if upper == "END:VEVENT" {
			if inEvent && currentStart != nil {
				// Ignore explicitly cancelled events
				if currentStatus != "CANCELLED" {
					// Fallback: if DTEND is missing or not after DTSTART, assume 1 night
					endDate := currentStart.AddDate(0, 0, 1)
					if currentEnd != nil && currentEnd.After(*currentStart) {
						endDate = *currentEnd
					}

					uid := currentUID
					if uid == "" {
						uid = fmt.Sprintf("ext-%d", currentStart.Unix())
					}

					events = append(events, ICalEvent{
						UID:       uid,
						StartDate: *currentStart,
						EndDate:   endDate,
						Summary:   currentSummary,
					})
				}
			}
			inEvent = false
			continue
		}

		if !inEvent {
			continue
		}

		// Parse properties
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		propKey := strings.ToUpper(strings.TrimSpace(line[:colonIdx]))
		propVal := strings.TrimSpace(line[colonIdx+1:])

		// Check property names (may include parameters like ;VALUE=DATE)
		switch {
		case propKey == "UID" || strings.HasPrefix(propKey, "UID;"):
			currentUID = propVal

		case propKey == "SUMMARY" || strings.HasPrefix(propKey, "SUMMARY;"):
			currentSummary = propVal

		case propKey == "STATUS" || strings.HasPrefix(propKey, "STATUS;"):
			currentStatus = strings.ToUpper(propVal)

		case propKey == "DTSTART" || strings.HasPrefix(propKey, "DTSTART;"):
			if t, err := parseICalDate(propVal); err == nil {
				currentStart = &t
			}

		case propKey == "DTEND" || strings.HasPrefix(propKey, "DTEND;"):
			if t, err := parseICalDate(propVal); err == nil {
				currentEnd = &t
			}
		}
	}

	return events, nil
}

// dateRegexp matches basic date/datetime patterns in iCal
var dateRegexp = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})`)

// parseICalDate parses RFC 5545 dates (YYYYMMDD or YYYYMMDDTHHMMSS[Z]) into a normalized pure date (UTC).
func parseICalDate(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	// Handle TZID prefixes like "TZID=Europe/Moscow:20260710T120000" if colon wasn't split
	if idx := strings.LastIndex(val, ":"); idx != -1 {
		val = val[idx+1:]
	}

	matches := dateRegexp.FindStringSubmatch(val)
	if len(matches) < 4 {
		return time.Time{}, fmt.Errorf("unsupported date format: %s", val)
	}

	year := matches[1]
	month := matches[2]
	day := matches[3]

	formatted := fmt.Sprintf("%s-%s-%s", year, month, day)
	t, err := time.Parse("2006-01-02", formatted)
	if err != nil {
		return time.Time{}, err
	}

	return t.UTC(), nil
}

func sanitizeICalText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}
