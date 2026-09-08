// src/lib/types/verification.ts

export type ProviderType =
	| 'individual'
	| 'self_employed'
	| 'individual_entrepreneur'
	| 'legal_entity';

export interface BelarusAddress {
	postal_code: string;
	city: string;
	street_building: string;
}

export interface BelarusContactPerson {
	full_name: string;
	email: string;
	position: string;
	phone: string;
}

export interface BelarusIndividualRequisites {
	first_name: string;
	last_name: string;
	email: string;
	phone: string;
	personal_id: string; // 14 characters passport/ID identification number
	birth_date: string; // YYYY-MM-DD or DD.MM.YYYY
	bank_name: string;
	unp: string; // 9 alphanumeric chars
	bic: string; // 8 or 11 chars
	iban: string; // 28 chars, starts with BY
	currency: 'BYN';
}

export interface BelarusCompanyRequisites {
	legal_name: string;
	brand_name?: string;
	website?: string;
	email: string;
	phone: string;
	legal_address: BelarusAddress;
	actual_address_same: boolean;
	actual_address?: BelarusAddress;
	bank_name: string;
	unp: string; // 9 digits
	bic: string; // 8 or 11 chars
	egr_number: string;
	iban: string; // 28 chars, starts with BY
	currency: 'BYN';
	contact_person: BelarusContactPerson;
}

export interface SubmitVerificationPayload {
	provider_type: ProviderType;
	legal_name: string;
	unp: string;
	requisites: Record<string, any>;
	documents: string[]; // array of media UUIDs
}

export interface VerificationResponse {
	id?: string;
	status: 'none' | 'pending' | 'approved' | 'rejected' | 'changes_requested';
	provider_type?: ProviderType;
	legal_name?: string;
	unp?: string;
	requisites?: Record<string, any>;
	documents?: string[];
	rejection_reason?: string;
	admin_note?: string;
	created_at?: string;
	updated_at?: string;
}

// ── Validation regexes for Republic of Belarus legal forms ───────────────────
export const UNP_COMPANY_REGEX = /^[0-9]{9}$/;
export const UNP_INDIVIDUAL_REGEX = /^[A-Za-z0-9]{9}$/;
export const IBAN_BELARUS_REGEX = /^BY[0-9]{2}[A-Za-z0-9]{4}[0-9]{20}$/;
export const BIC_REGEX = /^[A-Za-z0-9]{8}([A-Za-z0-9]{3})?$/;
export const PERSONAL_ID_REGEX = /^[0-9]{7}[A-Za-z][0-9]{3}[A-Za-z]{2}[0-9]$/;
