import { apiRequest, API_BASE } from '$lib/api/client';
import type {
	AvailabilityRange,
	HostCalendarResponse,
	ListingReservation,
	ListingCalendarSync,
	BlockDatesRequest,
	AddSyncFeedRequest,
	UpdateSyncFeedRequest
} from '$lib/types/calendar';

export const calendarApi = {
	/**
	 * Get public booked/blocked date ranges for a listing (anonymous).
	 */
	async getPublicAvailability(
		listingId: string,
		from?: string,
		to?: string
	): Promise<AvailabilityRange[]> {
		const params = new URLSearchParams();
		if (from) params.set('from', from);
		if (to) params.set('to', to);
		const qs = params.toString() ? `?${params.toString()}` : '';
		return apiRequest<AvailabilityRange[]>(`/listings/${listingId}/availability${qs}`, {
			skipAuth: true
		});
	},

	/**
	 * Get full calendar for host management (requires auth).
	 */
	async getHostCalendar(
		listingId: string,
		from?: string,
		to?: string
	): Promise<HostCalendarResponse> {
		const params = new URLSearchParams();
		if (from) params.set('from', from);
		if (to) params.set('to', to);
		const qs = params.toString() ? `?${params.toString()}` : '';
		return apiRequest<HostCalendarResponse>(`/listings/${listingId}/calendar${qs}`);
	},

	/**
	 * Manually block a date range for a listing.
	 */
	async blockDates(
		listingId: string,
		data: BlockDatesRequest
	): Promise<ListingReservation> {
		return apiRequest<ListingReservation>(`/listings/${listingId}/calendar/block`, {
			method: 'POST',
			body: data
		});
	},

	/**
	 * Unblock dates / delete a manual reservation.
	 */
	async unblockDates(listingId: string, resId: string): Promise<{ success: boolean }> {
		return apiRequest<{ success: boolean }>(
			`/listings/${listingId}/calendar/reservations/${resId}`,
			{
				method: 'DELETE'
			}
		);
	},

	/**
	 * Add an external iCal subscription URL.
	 */
	async addSyncFeed(
		listingId: string,
		data: AddSyncFeedRequest
	): Promise<ListingCalendarSync> {
		return apiRequest<ListingCalendarSync>(`/listings/${listingId}/calendar/syncs`, {
			method: 'POST',
			body: data
		});
	},

	/**
	 * Remove an external iCal subscription feed.
	 */
	async deleteSyncFeed(listingId: string, syncId: string): Promise<{ success: boolean }> {
		return apiRequest<{ success: boolean }>(
			`/listings/${listingId}/calendar/syncs/${syncId}`,
			{
				method: 'DELETE'
			}
		);
	},

	/**
	 * Update an external iCal subscription feed (name, color).
	 */
	async updateSyncFeed(
		listingId: string,
		syncId: string,
		data: UpdateSyncFeedRequest
	): Promise<ListingCalendarSync> {
		return apiRequest<ListingCalendarSync>(
			`/listings/${listingId}/calendar/syncs/${syncId}`,
			{
				method: 'PATCH',
				body: data
			}
		);
	},

	/**
	 * Trigger immediate background sync for all feeds of a listing.
	 */
	async syncNow(listingId: string): Promise<{ success: boolean; message: string }> {
		return apiRequest<{ success: boolean; message: string }>(
			`/listings/${listingId}/calendar/sync-now`,
			{
				method: 'POST'
			}
		);
	},

	/**
	 * Returns the direct URL for downloading the .ics feed.
	 */
	getExportFeedUrl(listingId: string): string {
		return `${API_BASE}/listings/${listingId}/calendar.ics`;
	}
};
