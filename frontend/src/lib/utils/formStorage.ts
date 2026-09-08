// src/lib/utils/formStorage.ts
import type { ListingFormValues } from '$lib/validation/listingValidation';

const STORAGE_PREFIX = 'listing_draft';
const STORAGE_VERSION = 2;
const MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000; // 7 дней

export interface DraftMeta {
	step?: number;
	photoCount?: number;
	coverIndex?: number;
}

interface StoredDraft {
	version: number;
	timestamp: number;
	data: Omit<ListingFormValues, 'photos'>;
	meta?: DraftMeta;
}

function hasTextValue(value: string | undefined): boolean {
	return Boolean(value?.trim());
}

function hasDraftContent(data: Omit<ListingFormValues, 'photos'>, meta?: DraftMeta): boolean {
	return Boolean(
		hasTextValue(data.title) ||
		hasTextValue(data.address) ||
		data.propertyType !== 'apartment' ||
		hasTextValue(data.size) ||
		hasTextValue(data.floor) ||
		hasTextValue(data.totalFloors) ||
		hasTextValue(data.maxGuests) ||
		hasTextValue(data.bedrooms) ||
		hasTextValue(data.beds) ||
		hasTextValue(data.bathrooms) ||
		hasTextValue(data.basePrice) ||
		hasTextValue(data.minNights) ||
		data.checkInFrom !== '14:00' ||
		data.checkOutBefore !== '12:00' ||
		hasTextValue(data.description) ||
		data.amenities.length > 0 ||
		data.allowChildren ||
		data.allowPets ||
		data.smokingAllowed ||
		data.partiesAllowed ||
		data.securityDepositRequired ||
		data.reportingDocuments ||
		hasTextValue(data.phone) ||
		data.contactMethods.some((method) => hasTextValue(method.value)) ||
		(meta?.photoCount ?? 0) > 0
	);
}

function getStorageKey(userId: string): string {
	return `${STORAGE_PREFIX}:v${STORAGE_VERSION}:${userId}`;
}

export function saveDraft(
	userId: string | null | undefined,
	form: ListingFormValues,
	meta: DraftMeta = {}
): void {
	if (typeof localStorage === 'undefined' || !userId) return;

	try {
		const key = getStorageKey(userId);
		const draft: StoredDraft = {
			version: STORAGE_VERSION,
			timestamp: Date.now(),
			data: {
				title: form.title,
				address: form.address,
				propertyType: form.propertyType,
				size: form.size,
				floor: form.floor,
				totalFloors: form.totalFloors,
				maxGuests: form.maxGuests,
				bedrooms: form.bedrooms,
				beds: form.beds,
				bathrooms: form.bathrooms,
				basePrice: form.basePrice,
				minNights: form.minNights,
				checkInFrom: form.checkInFrom,
				checkOutBefore: form.checkOutBefore,
				description: form.description,
				amenities: form.amenities,
				allowChildren: form.allowChildren,
				allowPets: form.allowPets,
				smokingAllowed: form.smokingAllowed,
				partiesAllowed: form.partiesAllowed,
				securityDepositRequired: form.securityDepositRequired,
				reportingDocuments: form.reportingDocuments,
				phone: form.phone,
				contactMethods: form.contactMethods
			},
			meta
		};

		if (!hasDraftContent(draft.data, draft.meta)) {
			localStorage.removeItem(key);
			return;
		}

		localStorage.setItem(key, JSON.stringify(draft));
	} catch (e) {
		console.warn('Не удалось сохранить черновик:', e);
	}
}

export function loadDraft(
	userId: string | null | undefined
): { data: Omit<ListingFormValues, 'photos'>; meta?: DraftMeta } | null {
	if (typeof localStorage === 'undefined' || !userId) return null;

	try {
		const raw = localStorage.getItem(getStorageKey(userId));
		if (!raw) return null;

		const draft: StoredDraft = JSON.parse(raw);

		if (draft.version !== STORAGE_VERSION) {
			clearDraft(userId);
			return null;
		}

		if (Date.now() - draft.timestamp > MAX_AGE_MS) {
			clearDraft(userId);
			return null;
		}

		if (!hasDraftContent(draft.data, draft.meta)) {
			clearDraft(userId);
			return null;
		}

		return { data: draft.data, meta: draft.meta };
	} catch (e) {
		console.warn('Не удалось загрузить черновик:', e);
		return null;
	}
}

export function clearDraft(userId: string | null | undefined): void {
	if (typeof localStorage === 'undefined' || !userId) return;

	try {
		localStorage.removeItem(getStorageKey(userId));
	} catch (e) {
		console.warn('Не удалось очистить черновик:', e);
	}
}

export function hasDraft(userId: string | null | undefined): boolean {
	return loadDraft(userId) !== null;
}

export function getDraftAge(userId: string | null | undefined): string | null {
	if (typeof localStorage === 'undefined' || !userId) return null;

	try {
		const raw = localStorage.getItem(getStorageKey(userId));
		if (!raw) return null;

		const draft: StoredDraft = JSON.parse(raw);
		const diff = Date.now() - draft.timestamp;

		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(diff / 3600000);
		const days = Math.floor(diff / 86400000);

		if (days > 0) return `${days} ${pluralize(days, 'день', 'дня', 'дней')} назад`;
		if (hours > 0) return `${hours} ${pluralize(hours, 'час', 'часа', 'часов')} назад`;
		if (minutes > 0) return `${minutes} ${pluralize(minutes, 'минуту', 'минуты', 'минут')} назад`;
		return 'только что';
	} catch {
		return null;
	}
}

function pluralize(n: number, one: string, few: string, many: string): string {
	const mod10 = n % 10;
	const mod100 = n % 100;

	if (mod100 >= 11 && mod100 <= 14) return many;
	if (mod10 === 1) return one;
	if (mod10 >= 2 && mod10 <= 4) return few;
	return many;
}
