import { apiRequest } from '../client';
import type { ListingPublic } from '$lib/types/listings';
import type { ModerationPayload } from '$lib/types/admin';

export const adminListingsApi = {
	/**
	 * Fetch listing moderation queue.
	 */
	getListings(status?: string): Promise<ListingPublic[]> {
		const query = status ? `?status=${encodeURIComponent(status)}` : '';
		return apiRequest<ListingPublic[]>(`/admin/listings${query}`, {
			method: 'GET'
		});
	},

	/**
	 * Fetch a single listing by ID for moderation inspection.
	 */
	getListing(id: string): Promise<ListingPublic> {
		return apiRequest<ListingPublic>(`/admin/listings/${id}`, {
			method: 'GET'
		});
	},

	/**
	 * Moderate a listing (approve, reject, request_changes, suspend).
	 */
	moderateListing(id: string, payload: ModerationPayload): Promise<ListingPublic> {
		return apiRequest<ListingPublic>(`/admin/listings/${id}/moderate`, {
			method: 'POST',
			body: payload
		});
	}
};
