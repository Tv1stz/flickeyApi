import { apiRequest } from './client';
import type {
	ListingPublic,
	ListingHost,
	CreateDraftRequest,
	CreateDraftResponse,
	StepResponse,
	DraftStep2Request,
	DraftStep3Request,
	DraftStep4Request,
	DraftStep5Request,
	DraftStep6Request,
	SubmitDraftResponse,
	DraftDetail,
	HousingType
} from '$lib/types/listings';
import type { PublicUser } from '$lib/types/users';

export const listingsApi = {
	/**
	 * Public catalog of published listings.
	 */
	getPublicListings(
		housingType?: HousingType,
		checkin?: string | null,
		checkout?: string | null,
		hostId?: string | null
	): Promise<ListingPublic[]> {
		const params = new URLSearchParams();
		if (housingType) params.set('type', housingType);
		if (checkin) params.set('checkin', checkin);
		if (checkout) params.set('checkout', checkout);
		if (hostId) params.set('host_id', hostId);
		const query = params.toString() ? `?${params.toString()}` : '';
		return apiRequest<ListingPublic[]>(`/listings${query}`, {
			method: 'GET',
			skipAuth: true
		});
	},

	/**
	 * Public profile of a host / user.
	 */
	getPublicUser(userId: string): Promise<PublicUser> {
		return apiRequest<PublicUser>(`/users/${userId}`, {
			method: 'GET',
			skipAuth: true
		});
	},

	/**
	 * Public details of a published listing.
	 */
	getPublicListing(id: string): Promise<ListingPublic> {
		return apiRequest<ListingPublic>(`/listings/${id}`, {
			method: 'GET'
		});
	},

	/**
	 * Host dashboard: Fetch all listings owned by authenticated host.
	 */
	getMyListings(): Promise<ListingHost[]> {
		return apiRequest<ListingHost[]>('/listings/my', {
			method: 'GET'
		});
	},

	/**
	 * Delete a listing owned by authenticated host.
	 */
	deleteListing(listingId: string): Promise<void> {
		return apiRequest<void>(`/listings/${listingId}`, {
			method: 'DELETE'
		});
	},

	/**
	 * Archive a listing owned by authenticated host.
	 */
	archiveListing(listingId: string): Promise<ListingHost> {
		return apiRequest<ListingHost>(`/listings/${listingId}/archive`, {
			method: 'POST'
		});
	},

	/**
	 * Unarchive a listing owned by authenticated host.
	 */
	unarchiveListing(listingId: string): Promise<ListingHost> {
		return apiRequest<ListingHost>(`/listings/${listingId}/unarchive`, {
			method: 'POST'
		});
	},

	/**
	 * Attach private apartment walkthrough video to a listing for verification.
	 */
	attachVerificationVideo(listingId: string, mediaId: string): Promise<ListingHost> {
		return apiRequest<ListingHost>(`/listings/${listingId}/verification-video`, {
			method: 'POST',
			body: { media_id: mediaId }
		});
	},

	/**
	 * Step 1: Create a new listing draft with a specified housing type.
	 */
	createDraft(payload: CreateDraftRequest): Promise<CreateDraftResponse> {
		return apiRequest<CreateDraftResponse>('/listings/drafts', {
			method: 'POST',
			body: payload
		});
	},

	/**
	 * Clone or create an edit draft from an existing listing.
	 */
	createDraftFromListing(listingId: string, mode?: 'create' | 'edit'): Promise<CreateDraftResponse> {
		const query = mode ? `?mode=${mode}` : '';
		return apiRequest<CreateDraftResponse>(`/listings/drafts/from-listing/${listingId}${query}`, {
			method: 'POST'
		});
	},


	/**
	 * Fetch current state of a draft.
	 */
	getDraft(draftId: string): Promise<DraftDetail> {
		return apiRequest<DraftDetail>(`/listings/drafts/${draftId}`, {
			method: 'GET'
		});
	},

	/**
	 * Step 2: Update property physical parameters.
	 */
	updateStep2(draftId: string, payload: DraftStep2Request): Promise<StepResponse> {
		return apiRequest<StepResponse>(`/listings/drafts/${draftId}/step-2`, {
			method: 'PATCH',
			body: payload
		});
	},

	/**
	 * Step 3: Attach uploaded media IDs (5-15 images).
	 */
	updateStep3(draftId: string, payload: DraftStep3Request): Promise<StepResponse> {
		return apiRequest<StepResponse>(`/listings/drafts/${draftId}/step-3`, {
			method: 'PATCH',
			body: payload
		});
	},

	/**
	 * Step 4: Select amenities.
	 */
	updateStep4(draftId: string, payload: DraftStep4Request): Promise<StepResponse> {
		return apiRequest<StepResponse>(`/listings/drafts/${draftId}/step-4`, {
			method: 'PATCH',
			body: payload
		});
	},

	/**
	 * Step 5: Pricing, currency (BYN), minimum stay, check-in/out, and rules.
	 */
	updateStep5(draftId: string, payload: DraftStep5Request): Promise<StepResponse> {
		return apiRequest<StepResponse>(`/listings/drafts/${draftId}/step-5`, {
			method: 'PATCH',
			body: payload
		});
	},

	/**
	 * Step 6: Property description (30-5000 chars).
	 */
	updateStep6(draftId: string, payload: DraftStep6Request): Promise<StepResponse> {
		return apiRequest<StepResponse>(`/listings/drafts/${draftId}/step-6`, {
			method: 'PATCH',
			body: payload
		});
	},

	/**
	 * Step 7: Atomic submission of the draft with Idempotency-Key.
	 */
	submitDraft(draftId: string, idempotencyKey: string): Promise<SubmitDraftResponse> {
		return apiRequest<SubmitDraftResponse>(`/listings/drafts/${draftId}/submit`, {
			method: 'POST',
			headers: {
				'Idempotency-Key': idempotencyKey
			}
		});
	}
};
