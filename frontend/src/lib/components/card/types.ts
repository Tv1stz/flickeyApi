// src/lib/components/card/types.ts

// ═══════════════════════════════════════════════════════════════
// OWNER
// ═══════════════════════════════════════════════════════════════

export interface OwnerContacts {
	phone?: string;
	email?: string;
	telegram?: string;
	viber?: string;
	whatsapp?: string;
	signal?: string;
}

export type ContactMethodType = 'email' | 'telegram' | 'whatsapp' | 'viber' | 'signal';

export interface ContactMethod {
	type: ContactMethodType;
	value: string;
}

export interface Owner {
	id: string;
	name: string;
	avatar?: string;
	isVerified?: boolean;
	listingsCount: number;
	contacts?: OwnerContacts;
	createdAt: string;
}

// ═══════════════════════════════════════════════════════════════
// LOCATION
// ═══════════════════════════════════════════════════════════════

export interface ListingLocation {
	lat: number;
	lng: number;
	city: string;
	district?: string;
	address: string;
}

// ═══════════════════════════════════════════════════════════════
// RULES
// ═══════════════════════════════════════════════════════════════

export interface ListingRules {
	checkIn: string;
	checkOut: string;
	minNights: number;
	depositRequired: boolean;
	documentsProvided: boolean;
	childrenAllowed: boolean;
	petsAllowed: boolean;
	smokingAllowed: boolean;
	partiesAllowed: boolean;
}

// ═══════════════════════════════════════════════════════════════
// PROPERTY TYPE
// ═══════════════════════════════════════════════════════════════

export type PropertyType = 'apartment' | 'house' | 'estate';

export const PROPERTY_TYPE_LABELS: Record<PropertyType, string> = {
	apartment: 'Квартира',
	house: 'Дом / Коттедж',
	estate: 'Усадьба'
};

// ═══════════════════════════════════════════════════════════════
// LISTING
// ═══════════════════════════════════════════════════════════════

export interface Listing {
	id: string;
	title: string;
	description: string;
	propertyType: PropertyType;

	// Адрес и локация
	address: string;
	location?: ListingLocation;

	// Параметры жилья
	area: number;
	floor?: number;
	totalFloors: number;

	// Вместимость
	maxGuests: number;
	bedrooms: number;
	beds: number;
	bathrooms: number;

	// Цена
	pricePerNight: number;
	currency: 'BYN' | 'USD' | 'EUR';

	// Медиа и удобства
	images: string[];
	amenities: string[];

	// Правила
	rules: ListingRules;

	// Владелец
	owner?: Owner;
	ownerId: string;

	// Мета
	createdAt: string;
	updatedAt: string;
	isActive: boolean;
	status?: string;
	verificationVideoUrl?: string;
	verificationVideoId?: string;
	rejectionReason?: string;
	moderationComment?: string;
}

// ═══════════════════════════════════════════════════════════════
// API TYPES
// ═══════════════════════════════════════════════════════════════

export type ListingCreateBase = Pick<
	Listing,
	| 'title'
	| 'description'
	| 'propertyType'
	| 'address'
	| 'location'
	| 'area'
	| 'floor'
	| 'totalFloors'
	| 'maxGuests'
	| 'bedrooms'
	| 'beds'
	| 'bathrooms'
	| 'currency'
	| 'amenities'
>;

export type ListingCreateRules = ListingRules;

export interface CreateListingRequest extends ListingCreateBase {
	pricePerNight: Listing['pricePerNight'];
	phone: string;
	rules: ListingCreateRules;
	contactMethods: ContactMethod[];
}

export interface ListingResponse {
	success: boolean;
	data?: Listing;
	error?: string;
}

export interface ListingsResponse {
	success: boolean;
	data?: {
		listings: Listing[];
		total: number;
		page: number;
		perPage: number;
	};
	error?: string;
}

// ═══════════════════════════════════════════════════════════════
// COMPONENT PROPS
// ═══════════════════════════════════════════════════════════════

export interface ListingCardProps {
	listing: Listing;
	onFavoriteToggle?: (id: string, isFavorite: boolean) => void;
	isFavorite?: boolean;
	priority?: boolean;
}

export interface ImageCarouselProps {
	images: string[];
	alt?: string;
	loop?: boolean;
	showDots?: boolean;
	showArrows?: boolean;
	aspectRatio?: 'square' | 'video' | 'portrait';
	priority?: boolean;
	onSlideChange?: (index: number) => void;
	maxImages?: number;
}

export interface ListingGridProps {
	listings: Listing[];
	isLoading?: boolean;
	skeletonCount?: number;
	emptyMessage?: string;
	emptyDescription?: string;
	onLoadMore?: () => void;
	hasMore?: boolean;
	isLoadingMore?: boolean;
	favorites?: Set<string>;
	onFavoriteToggle?: (id: string, isFavorite: boolean) => void;
	layout?: 'default' | 'two-column';
}
