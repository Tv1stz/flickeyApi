// src/lib/utils/listingConverters.ts

import type {
	Listing,
	CreateListingRequest,
	PropertyType,
	OwnerContacts,
	ListingRules,
	ContactMethod
} from '$lib/components/card/types';
import type { ListingFormValues } from '$lib/validation/listingValidation';
import { normalizeAddress } from '$lib/utils/location';

type LegacyListingRuleFields = {
	allowChildren?: boolean;
	allowPets?: boolean;
	securityDepositRequired?: boolean;
	reportingDocuments?: boolean;
};

function toFiniteNumber(value: unknown): number | null {
	if (typeof value === 'number' && Number.isFinite(value)) return value;
	if (typeof value === 'string') {
		const numeric = Number(value);
		return Number.isFinite(numeric) ? numeric : null;
	}
	return null;
}

function toBoolean(value: unknown, fallback = false): boolean {
	return typeof value === 'boolean' ? value : fallback;
}

function toNonEmptyString(value: unknown, fallback: string): string {
	return typeof value === 'string' && value.trim().length > 0 ? value : fallback;
}

function normalizeMultilineText(value: string): string {
	return value.replace(/\r\n?/g, '\n').trim();
}

function normalizeCurrency(value: unknown): Listing['currency'] {
	if (value === 'BYN' || value === 'USD' || value === 'EUR') return value;
	return 'BYN';
}

function toInteger(value: unknown): number | null {
	const numeric = toFiniteNumber(value);
	if (numeric === null) return null;
	return Math.trunc(numeric);
}

function normalizeInteger(value: unknown, fallback: number, min: number): number {
	const numeric = toInteger(value);
	if (numeric === null) return fallback;
	return Math.max(min, numeric);
}

function normalizeOwnerContacts(value: unknown): OwnerContacts {
	if (!value || typeof value !== 'object') return {};

	const source = value as OwnerContacts;
	const normalized: OwnerContacts = {};
	const contactKeys: Array<keyof OwnerContacts> = [
		'phone',
		'email',
		'telegram',
		'viber',
		'whatsapp',
		'signal'
	];

	for (const key of contactKeys) {
		const raw = source[key];
		if (typeof raw === 'string') {
			const trimmed = raw.trim();
			if (trimmed) normalized[key] = trimmed;
		}
	}

	return normalized;
}

function normalizeRules(listing: Partial<Listing> & LegacyListingRuleFields): ListingRules {
	const sourceRules = listing.rules ?? ({} as Partial<ListingRules>);
	const minNights = Math.max(1, toFiniteNumber(sourceRules.minNights) ?? 1);

	return {
		checkIn: toNonEmptyString(sourceRules.checkIn, '14:00'),
		checkOut: toNonEmptyString(sourceRules.checkOut, '12:00'),
		minNights,
		depositRequired: toBoolean(
			sourceRules.depositRequired,
			toBoolean(listing.securityDepositRequired)
		),
		documentsProvided: toBoolean(
			sourceRules.documentsProvided,
			toBoolean(listing.reportingDocuments)
		),
		childrenAllowed: toBoolean(sourceRules.childrenAllowed, toBoolean(listing.allowChildren)),
		petsAllowed: toBoolean(sourceRules.petsAllowed, toBoolean(listing.allowPets)),
		smokingAllowed: toBoolean(sourceRules.smokingAllowed),
		partiesAllowed: toBoolean(sourceRules.partiesAllowed)
	};
}

export function normalizeListing(
	listing: Listing | (Partial<Listing> & LegacyListingRuleFields)
): Listing {
	const propertyType = (listing.propertyType as PropertyType) || 'apartment';
	const normalizedAddress = normalizeAddress(
		toNonEmptyString(listing.address, listing.location?.address ?? 'Адрес не указан')
	);
	const normalizedRules = normalizeRules(listing);
	const totalFloors = normalizeInteger(listing.totalFloors, 1, 1);
	const floorValue = toInteger(listing.floor);
	const floor =
		propertyType === 'apartment' && floorValue !== null
			? Math.max(1, Math.min(floorValue, totalFloors))
			: undefined;
	const normalizedOwnerContacts = normalizeOwnerContacts(listing.owner?.contacts);
	const normalizedOwnerId = toNonEmptyString(listing.ownerId, listing.owner?.id ?? 'unknown-owner');

	const owner = listing.owner
		? {
				...listing.owner,
				id: toNonEmptyString(listing.owner.id, normalizedOwnerId),
				name: toNonEmptyString(listing.owner.name, 'Хозяин'),
				listingsCount: normalizeInteger(listing.owner.listingsCount, 1, 1),
				isVerified: toBoolean(listing.owner.isVerified),
				contacts: normalizedOwnerContacts,
				createdAt: toNonEmptyString(listing.owner.createdAt, new Date().toISOString())
			}
		: undefined;

	const locationLat = toFiniteNumber(listing.location?.lat);
	const locationLng = toFiniteNumber(listing.location?.lng);

	const normalizedLocation =
		listing.location && locationLat !== null && locationLng !== null
			? {
					...listing.location,
					lat: locationLat,
					lng: locationLng,
					address: normalizedAddress,
					city: toNonEmptyString(listing.location.city, 'Не указан'),
					district:
						typeof listing.location.district === 'string' &&
						listing.location.district.trim().length > 0
							? listing.location.district.trim()
							: undefined
				}
			: undefined;

	return {
		...(listing as Listing),
		id: toNonEmptyString(listing.id, crypto?.randomUUID?.() ?? `listing-${Date.now()}`),
		title: toNonEmptyString(listing.title, 'Объявление без названия'),
		description: normalizeMultilineText(toNonEmptyString(listing.description, '')),
		propertyType,
		address: normalizedAddress,
		area: Math.max(0, toFiniteNumber(listing.area) ?? 0),
		floor,
		totalFloors,
		maxGuests: normalizeInteger(listing.maxGuests, 1, 1),
		bedrooms: normalizeInteger(listing.bedrooms, 0, 0),
		beds: normalizeInteger(listing.beds, 1, 1),
		bathrooms: normalizeInteger(listing.bathrooms, 1, 1),
		pricePerNight: Math.max(0, toFiniteNumber(listing.pricePerNight) ?? 0),
		currency: normalizeCurrency(listing.currency),
		images: Array.isArray(listing.images)
			? listing.images.filter(
					(image): image is string => typeof image === 'string' && image.trim().length > 0
				)
			: [],
		amenities: Array.isArray(listing.amenities)
			? Array.from(
					new Set(
						listing.amenities.filter(
							(amenity): amenity is string =>
								typeof amenity === 'string' && amenity.trim().length > 0
						)
					)
				)
			: [],
		rules: normalizedRules,
		location: normalizedLocation,
		owner,
		ownerId: normalizedOwnerId,
		createdAt: toNonEmptyString(listing.createdAt, new Date().toISOString()),
		updatedAt: toNonEmptyString(listing.updatedAt, new Date().toISOString()),
		isActive: toBoolean(listing.isActive, true)
	};
}

// ═══════════════════════════════════════════════════════════════
// FORM → API REQUEST
// ═══════════════════════════════════════════════════════════════

function parseNumberField(value: string): number | null {
	const normalized = value.trim().replace(',', '.');
	if (!normalized) return null;
	const numeric = Number(normalized);
	return Number.isFinite(numeric) ? numeric : null;
}

export function parseIntegerField(value: string): number | null {
	const numeric = parseNumberField(value);
	if (numeric === null || !Number.isInteger(numeric)) return null;
	return numeric;
}

export function formToCreateRequest(form: ListingFormValues): CreateListingRequest {
	return {
		title: form.title.trim(),
		description: normalizeMultilineText(form.description),
		propertyType: form.propertyType as PropertyType,
		address: normalizeAddress(form.address),
		location: form.location,
		area: parseNumberField(form.size) ?? 0,
		floor:
			form.propertyType === 'apartment' ? (parseIntegerField(form.floor) ?? undefined) : undefined,
		totalFloors: parseIntegerField(form.totalFloors) ?? 1,
		maxGuests: parseIntegerField(form.maxGuests) ?? 1,
		bedrooms: parseIntegerField(form.bedrooms) ?? 0,
		beds: parseIntegerField(form.beds) ?? 1,
		bathrooms: parseIntegerField(form.bathrooms) ?? 1,
		pricePerNight: parseNumberField(form.basePrice) ?? 0,
		currency: 'BYN',
		amenities: Array.from(new Set(form.amenities)),
		phone: form.phone.trim(),
		rules: {
			checkIn: form.checkInFrom,
			checkOut: form.checkOutBefore,
			minNights: parseIntegerField(form.minNights) ?? 1,
			depositRequired: form.securityDepositRequired,
			documentsProvided: form.reportingDocuments,
			childrenAllowed: form.allowChildren,
			petsAllowed: form.allowPets,
			smokingAllowed: form.smokingAllowed,
			partiesAllowed: form.partiesAllowed
		},
		contactMethods: form.contactMethods
			.map((method) => ({
				type: method.type,
				value: method.value.trim()
			}))
			.filter((method): method is ContactMethod => Boolean(method.value))
	};
}

// ═══════════════════════════════════════════════════════════════
// LISTING → FORM (для редактирования)
// ═══════════════════════════════════════════════════════════════

export function listingToForm(listing: Listing): Partial<ListingFormValues> {
	const normalized = normalizeListing(listing);
	const contacts = normalized.owner?.contacts ?? {};
	const contactMethods: ListingFormValues['contactMethods'] = [];

	if (contacts.email) contactMethods.push({ type: 'email', value: contacts.email });
	if (contacts.telegram) contactMethods.push({ type: 'telegram', value: contacts.telegram });
	if (contacts.whatsapp) contactMethods.push({ type: 'whatsapp', value: contacts.whatsapp });
	if (contacts.viber) contactMethods.push({ type: 'viber', value: contacts.viber });
	if (contacts.signal) contactMethods.push({ type: 'signal', value: contacts.signal });

	return {
		title: normalized.title,
		description: normalized.description,
		propertyType: normalized.propertyType,
		address: normalized.address,
		size: String(normalized.area),
		floor: normalized.floor ? String(normalized.floor) : '',
		totalFloors: String(normalized.totalFloors),
		maxGuests: String(normalized.maxGuests),
		bedrooms: String(normalized.bedrooms),
		beds: String(normalized.beds),
		bathrooms: String(normalized.bathrooms),
		basePrice: String(normalized.pricePerNight),
		minNights: String(normalized.rules.minNights),
		checkInFrom: normalized.rules.checkIn,
		checkOutBefore: normalized.rules.checkOut,
		allowChildren: normalized.rules.childrenAllowed,
		allowPets: normalized.rules.petsAllowed,
		smokingAllowed: normalized.rules.smokingAllowed,
		partiesAllowed: normalized.rules.partiesAllowed,
		securityDepositRequired: normalized.rules.depositRequired,
		reportingDocuments: normalized.rules.documentsProvided,
		amenities: normalized.amenities,
		phone: contacts.phone ?? '',
		contactMethods
	};
}

// ═══════════════════════════════════════════════════════════════
// VALIDATION HELPERS
// ═══════════════════════════════════════════════════════════════

export function isValidPropertyType(value: string): value is PropertyType {
	return ['apartment', 'house', 'estate'].includes(value);
}

export function parseNumericField(value: string): number | null {
	return parseIntegerField(value);
}

export function buildListingImageFallbacks(_propertyType: PropertyType, count = 1): string[] {
	const pool = Array.from({ length: 10 }, (_, i) => `/${i + 1}.jpeg`);
	const safeCount = Math.max(1, count);
	return Array.from({ length: safeCount }, () => pool[Math.floor(Math.random() * pool.length)]);
}

// ═══════════════════════════════════════════════════════════════
// FORM → LISTING PREVIEW
// ═══════════════════════════════════════════════════════════════

function contactMethodsToOwnerContacts(
	methods: ListingFormValues['contactMethods']
): OwnerContacts {
	const contacts: OwnerContacts = {};

	for (const method of methods) {
		const value = method.value.trim();
		if (!value) continue;

		if (method.type === 'email') {
			contacts.email = value;
		} else if (method.type === 'telegram') {
			contacts.telegram = value;
		} else if (method.type === 'whatsapp') {
			contacts.whatsapp = value;
		} else if (method.type === 'viber') {
			contacts.viber = value;
		} else if (method.type === 'signal') {
			contacts.signal = value;
		}
	}

	return contacts;
}

export function formToListingPreview(
	form: ListingFormValues,
	imageUrls: string[],
	ownerName = 'Вы'
): Listing {
	const now = new Date().toISOString();
	const propertyType = (form.propertyType as PropertyType) || 'apartment';
	const totalFloors = parseIntegerField(form.totalFloors) ?? 1;
	const floor = propertyType === 'apartment' ? (parseIntegerField(form.floor) ?? 1) : undefined;
	const images =
		imageUrls.length > 0
			? imageUrls
			: buildListingImageFallbacks(propertyType, Math.max(1, imageUrls.length));
	const ownerContacts = contactMethodsToOwnerContacts(form.contactMethods);
	if (form.phone?.trim()) {
		ownerContacts.phone = form.phone.trim();
	}

	return normalizeListing({
		id: 'preview',
		title: form.title.trim() || 'Название объявления',
		description: normalizeMultilineText(form.description) || 'Описание объявления будет здесь.',
		propertyType,
		address: form.address.trim() || 'Адрес не указан',
		area: parseNumberField(form.size) ?? 0,
		floor,
		totalFloors,
		maxGuests: parseIntegerField(form.maxGuests) ?? 1,
		bedrooms: parseIntegerField(form.bedrooms) ?? 1,
		beds: parseIntegerField(form.beds) ?? 1,
		bathrooms: parseIntegerField(form.bathrooms) ?? 1,
		pricePerNight: parseNumberField(form.basePrice) ?? 0,
		currency: 'BYN',
		images,
		amenities: form.amenities,
		rules: {
			checkIn: form.checkInFrom || '14:00',
			checkOut: form.checkOutBefore || '12:00',
			minNights: parseIntegerField(form.minNights) ?? 1,
			depositRequired: form.securityDepositRequired,
			documentsProvided: form.reportingDocuments,
			childrenAllowed: form.allowChildren,
			petsAllowed: form.allowPets,
			smokingAllowed: form.smokingAllowed,
			partiesAllowed: form.partiesAllowed
		},
		owner: {
			id: 'preview-owner',
			name: ownerName,
			listingsCount: 1,
			contacts: ownerContacts,
			createdAt: now
		},
		ownerId: 'preview-owner',
		createdAt: now,
		updatedAt: now,
		isActive: true
	});
}

import type { ListingPublic } from '$lib/types/listings';

export function apiListingToCardListing(apiListing: ListingPublic): Listing {
	const rawType = apiListing.type as string;
	const propertyType: PropertyType =
		rawType === 'manor' ? 'estate' : rawType === 'house' ? 'house' : 'apartment';

	const ownerName = apiListing.host
		? apiListing.host.first_name
			? `${apiListing.host.first_name}${apiListing.host.last_name ? ' ' + apiListing.host.last_name : ''}`
			: apiListing.host.name || 'Хозяин'
		: 'Хозяин';

	const fullAddress =
		apiListing.address ||
		(apiListing.city
			? `${apiListing.city}${apiListing.street ? ', ' + apiListing.street : ''}${apiListing.house_number ? ', д. ' + apiListing.house_number : ''}`
			: 'Беларусь');

	return normalizeListing({
		id: apiListing.id,
		title: apiListing.name || 'Без названия',
		description: apiListing.description || '',
		propertyType,
		address: fullAddress,
		location:
			apiListing.latitude && apiListing.longitude
				? {
						lat: apiListing.latitude,
						lng: apiListing.longitude,
						city: apiListing.city || '',
						address: fullAddress
					}
				: undefined,
		area: apiListing.square || 0,
		floor: apiListing.floor,
		totalFloors: apiListing.total_floors || 1,
		maxGuests: apiListing.max_guests || 1,
		bedrooms: apiListing.rooms_count || 1,
		beds: apiListing.beds_count || 1,
		bathrooms: apiListing.bathrooms_count || 1,
		pricePerNight: apiListing.price_per_night || 0,
		currency: (apiListing.currency as 'BYN' | 'USD' | 'EUR') || 'BYN',
		images: apiListing.media && apiListing.media.length > 0 ? apiListing.media : [],
		amenities: apiListing.amenities || [],
		rules: {
			checkIn: apiListing.checkin_from || '14:00',
			checkOut: apiListing.checkout_until || '12:00',
			minNights: apiListing.min_nights || 1,
			depositRequired: Boolean(apiListing.deposit_required || apiListing.rules?.deposit_required),
			documentsProvided: Boolean(apiListing.with_invoicing || apiListing.rules?.with_invoicing),
			childrenAllowed: Boolean(apiListing.allow_children || apiListing.rules?.allow_children),
			petsAllowed: Boolean(apiListing.allow_pets || apiListing.rules?.allow_pets),
			smokingAllowed: Boolean(apiListing.allow_smoking || apiListing.rules?.allow_smoking),
			partiesAllowed: Boolean(apiListing.allow_parties || apiListing.rules?.allow_parties)
		},
		owner: {
			id: apiListing.host_id || (apiListing.host?.id ?? ''),
			name: ownerName,
			listingsCount: 1,
			contacts: {
				phone: apiListing.host?.phone,
				email: apiListing.host?.email
			},
			createdAt: apiListing.created_at || ''
		},
		ownerId: apiListing.host_id || apiListing.host?.id || '',
		createdAt: apiListing.created_at || '',
		updatedAt: apiListing.created_at || '',
		status: apiListing.status,
		isActive: apiListing.status === 'published' || apiListing.status === 'active' || !apiListing.status,
		verificationVideoUrl: (apiListing as any).verification_video_url,
		verificationVideoId: (apiListing as any).verification_video_id
	});
}

