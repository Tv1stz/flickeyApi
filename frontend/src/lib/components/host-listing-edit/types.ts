// src/lib/components/host-listing-edit/types.ts
import type { ComponentType } from 'svelte';
import type { Listing } from '$lib/components/card/types';

export type ListingEditSectionId =
	| 'basics'
	| 'photos'
	| 'description'
	| 'location'
	| 'amenities'
	| 'pricing'
	| 'rules'
	| 'contacts';

export interface ListingEditNavItem {
	id: ListingEditSectionId;
	title: string;
	subtitle: string;
	icon: ComponentType;
	group: 'main' | 'details' | 'settings';
}

export interface ListingEditFormData {
	title: string;
	description: string;
	propertyType: 'apartment' | 'house' | 'estate' | 'manor' | 'room' | string;
	address: string;
	city?: string;
	street?: string;
	houseNumber?: string;
	latitude?: number;
	longitude?: number;
	area: number;
	floor?: number;
	totalFloors: number;
	maxGuests: number;
	bedrooms: number;
	beds: number;
	bathrooms: number;
	pricePerNight: number;
	currency: 'BYN' | 'USD' | 'EUR' | string;
	images: string[];
	amenities: string[];
	rules: {
		checkIn: string;
		checkOut: string;
		minNights: number;
		depositRequired: boolean;
		documentsProvided: boolean;
		childrenAllowed: boolean;
		petsAllowed: boolean;
		smokingAllowed: boolean;
		partiesAllowed: boolean;
	};
	contacts: {
		phone: string;
		email?: string;
		telegram?: string;
		whatsapp?: string;
	};
	isActive: boolean;
}

export interface ListingEditStore {
	original: Listing | null;
	draft: Partial<ListingEditFormData>;
	activeSection: ListingEditSectionId;
	isSaving: boolean;
	isDirty: boolean;
	lastSaved: Date | null;
}

export type ListingEditAction =
	| { type: 'save' }
	| { type: 'publish' }
	| { type: 'unpublish' }
	| { type: 'preview' }
	| { type: 'discard' };
