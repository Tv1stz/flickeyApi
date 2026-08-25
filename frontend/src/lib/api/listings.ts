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

export const listingsApi = {
	/**
	 * Public catalog of published listings.
	 */
	getPublicListings(housingType?: HousingType): Promise<ListingPublic[]> {
		const query = housingType ? `?type=${encodeURIComponent(housingType)}` : '';
		return apiRequest<ListingPublic[]>(`/listings${query}`, {
			method: 'GET',
			skipAuth: true
		});
	},

	/**
	 * Public details of a published listing.
	 */
	getPublicListing(id: string): Promise<ListingPublic> {
		return apiRequest<ListingPublic>(`/listings/${id}`, {
			method: 'GET',
			skipAuth: true
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
	 * Step 1: Create a new listing draft with a specified housing type.
	 */
	createDraft(payload: CreateDraftRequest): Promise<CreateDraftResponse> {
		return apiRequest<CreateDraftResponse>('/listings/drafts', {
			method: 'POST',
			body: payload
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
