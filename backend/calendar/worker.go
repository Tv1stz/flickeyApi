package calendar

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SyncWorker manages background periodic iCal synchronization.
type SyncWorker struct {
	Service     *CalendarService
	DB          *gorm.DB
	Interval    time.Duration
	Concurrency int
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewSyncWorker creates a new SyncWorker.
func NewSyncWorker(svc *CalendarService, database *gorm.DB, interval time.Duration, concurrency int) *SyncWorker {
	if interval < 1*time.Minute {
		interval = 15 * time.Minute
	}
	if concurrency <= 0 {
		concurrency = 5
	}
	return &SyncWorker{
		Service:     svc,
		DB:          database,
		Interval:    interval,
		Concurrency: concurrency,
		stopCh:      make(chan struct{}),
	}
}

// Start launches the background worker loop in a separate goroutine.
func (w *SyncWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		log.Printf("[CalendarSyncWorker] Started with interval %v, concurrency %d", w.Interval, w.Concurrency)

		// Run an initial sync cycle shortly after startup
		time.Sleep(10 * time.Second)
		w.runCycle(ctx)

		ticker := time.NewTicker(w.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[CalendarSyncWorker] Context cancelled, stopping...")
				return
			case <-w.stopCh:
				log.Println("[CalendarSyncWorker] Stop requested, stopping...")
				return
			case <-ticker.C:
				w.runCycle(ctx)
			}
		}
	}()
}

// Stop gracefully shuts down the worker and waits for in-flight syncs to finish.
func (w *SyncWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	log.Println("[CalendarSyncWorker] Stopped cleanly.")
}

func (w *SyncWorker) runCycle(ctx context.Context) {
	var feedIDs []uuid.UUID
	err := w.DB.WithContext(ctx).
		Model(&ListingCalendarSync{}).
		Where("is_active = true").
		Pluck("id", &feedIDs).Error
	if err != nil {
		log.Printf("[CalendarSyncWorker] Error querying active feeds: %v", err)
		return
	}

	if len(feedIDs) == 0 {
		return
	}

	log.Printf("[CalendarSyncWorker] Starting sync cycle for %d feeds", len(feedIDs))

	sem := make(chan struct{}, w.Concurrency)
	var wg sync.WaitGroup

	for _, id := range feedIDs {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(feedID uuid.UUID) {
			defer func() {
				<-sem
				wg.Done()
			}()

			syncCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := w.Service.SyncFeed(syncCtx, feedID); err != nil {
				log.Printf("[CalendarSyncWorker] Feed %s error: %v", feedID, err)
			}
		}(id)
	}

	wg.Wait()
	log.Println("[CalendarSyncWorker] Completed sync cycle.")
}
