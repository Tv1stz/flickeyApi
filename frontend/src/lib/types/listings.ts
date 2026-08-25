export type HousingType = 'apartment' | 'house' | 'manor';

export type ListingStatus =
	| 'draft'
	| 'pending_review'
	| 'awaiting_company_verification'
	| 'published'
	| 'rejected'
	| 'archived';

export interface ListingRules {
	allow_children?: boolean;
	allow_pets?: boolean;
	allow_smoking?: boolean;
	allow_parties?: boolean;
	deposit_required?: boolean;
	with_invoicing?: boolean;
}

export interface HostInfo {
	id: string;
	name?: string;
	first_name?: string;
	last_name?: string;
	phone?: string;
	email?: string;
	role?: string;
}

export interface ListingPublic {
	id: string;
	host_id?: string;
	host?: HostInfo;
	status?: string;
	type: HousingType;
	name: string;
	square: number;
	floor: number;
	total_floors: number;
	max_guests: number;
	rooms_count: number;
	beds_count: number;
	bathrooms_count: number;
	price_per_night: number;
	currency: string;
	min_nights: number;
	checkin_from: string;
	checkout_until: string;
	description: string;
	amenities: string[];
	media: string[];
	created_at: string;
}

export interface ListingHost extends ListingPublic {
	status: ListingStatus;
	host_id: string;
	updated_at: string;
}

export interface CreateDraftRequest {
	type: HousingType;
}

export interface CreateDraftResponse {
	draft_id: string;
	current_step: number;
}

export interface StepResponse {
	draft_id: string;
	current_step: number;
	status: string;
}

export interface DraftStep2Request {
	name: string;
	square: number;
	floor: number;
	total_floors: number;
	max_guests: number;
	rooms_count: number;
	beds_count: number;
	bathrooms_count: number;
}

export interface DraftStep3Request {
	media_ids: string[];
}

export interface DraftStep4Request {
	amenities: string[];
}

export interface DraftStep5Request {
	price_per_night: number;
	currency: string;
	min_nights: number;
	checkin_from: string;
	checkout_until: string;
	rules?: ListingRules;
}

export interface DraftStep6Request {
	description: string;
}

export interface SubmitDraftResponse {
	listing_id: string;
	status: string;
}

export interface DraftDetail {
	id: string;
	host_id: string;
	current_step: number;
	status: string;
	type: HousingType | null;
	name: string | null;
	square: number | null;
	floor: number | null;
	total_floors: number | null;
	max_guests: number | null;
	rooms_count: number | null;
	beds_count: number | null;
	bathrooms_count: number | null;
	media_ids: string[];
	amenities: string[];
	price_per_night: number | null;
	currency: string | null;
	min_nights: number | null;
	checkin_from: string | null;
	checkout_until: string | null;
	allow_children: boolean | null;
	allow_pets: boolean | null;
	allow_smoking: boolean | null;
	allow_parties: boolean | null;
	deposit_required: boolean | null;
	with_invoicing: boolean | null;
	description: string | null;
	created_at: string;
	updated_at: string;
}
