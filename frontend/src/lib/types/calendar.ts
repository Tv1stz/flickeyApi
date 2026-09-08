/**
 * Calendar and booking types matching Go backend schema (snake_case).
 */

export type ReservationSource = 'flickey' | 'manual_block' | 'ical_import';
export type SyncStatus = 'idle' | 'syncing' | 'success' | 'failed';

export interface AvailabilityRange {
	start_date: string; // YYYY-MM-DD
	end_date: string;   // YYYY-MM-DD (exclusive DTEND)
}

export interface ListingReservation {
	id: string;
	listing_id: string;
	source: ReservationSource;
	sync_feed_id?: string | null;
	start_date: string;
	end_date: string;
	status: 'confirmed' | 'blocked' | 'cancelled';
	guest_name?: string;
	guest_phone?: string;
	external_uid?: string;
	note?: string;
	created_at: string;
	updated_at: string;
}

export interface ListingCalendarSync {
	id: string;
	listing_id: string;
	name: string;
	feed_url: string;
	color?: string;
	last_synced_at?: string | null;
	sync_status: SyncStatus;
	error_message?: string | null;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface HostCalendarResponse {
	listing_id: string;
	export_url: string;
	reservations: ListingReservation[];
	sync_feeds: ListingCalendarSync[];
}

export interface BlockDatesRequest {
	start_date: string; // YYYY-MM-DD
	end_date: string;   // YYYY-MM-DD
	note?: string;
	guest_name?: string;
	guest_phone?: string;
}

export interface AddSyncFeedRequest {
	name: string;
	feed_url: string;
	color?: string;
}

export interface UpdateSyncFeedRequest {
	name?: string;
	color?: string;
}
