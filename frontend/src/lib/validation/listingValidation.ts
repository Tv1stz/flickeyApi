// src/lib/validation/listingValidation.ts

import type { PropertyType, ContactMethod, ContactMethodType, ListingLocation } from '$lib/components/card/types';

export type { ContactMethod, ContactMethodType, ListingLocation };

export interface ListingFormValidatedSlice {
	// О жилье
	title: string;
	propertyType: PropertyType | '';
	address: string;
	description: string;

	// Параметры (строки для инпутов)
	size: string; // area в Listing
	floor: string;
	totalFloors: string;
	maxGuests: string; // maxGuests в Listing
	bedrooms: string; // bedrooms в Listing
	beds: string;
	bathrooms: string;

	// Цена и правила
	basePrice: string; // pricePerNight в Listing
	minNights: string;
	checkInFrom: string; // rules.checkIn в Listing
	checkOutBefore: string; // rules.checkOut в Listing

	// Удобства и медиа
	photos: (File | { id: string; url: string; mediaId?: string })[];

	// Контакты
	contactMethods: ContactMethod[];
	phone: string;
}

export interface ListingFormValues extends ListingFormValidatedSlice {
	location?: ListingLocation;
	amenities: string[];
	allowChildren: boolean;
	allowPets: boolean;
	smokingAllowed: boolean;
	partiesAllowed: boolean;
	securityDepositRequired: boolean;
	reportingDocuments: boolean;
}

export type ValidatedListingField = keyof ListingFormValidatedSlice;
export type ListingErrors = Record<ValidatedListingField, string>;
export type ListingTouched = Record<ValidatedListingField, boolean>;

// --- Утилиты парсинга ---

function parseNumber(value: string): number | null {
	const trimmed = value.trim();
	if (!trimmed) return null;
	const num = Number(trimmed);
	return Number.isNaN(num) ? null : num;
}

function parseInteger(value: string): number | null {
	const num = parseNumber(value);
	if (num === null) return null;
	return Number.isInteger(num) ? num : null;
}

function isValidTime(value: string): boolean {
	const trimmed = value.trim();
	const match = /^(\d{2}):(\d{2})$/.exec(trimmed);
	if (!match) return false;

	const hours = Number(match[1]);
	const minutes = Number(match[2]);
	if (!Number.isFinite(hours) || !Number.isFinite(minutes)) return false;

	const totalMinutes = hours * 60 + minutes;
	return totalMinutes >= 0 && totalMinutes <= 23 * 60 && totalMinutes % 30 === 0;
}

function isValidPhone(value: string): boolean {
	const digits = value.replace(/\D/g, '');
	return digits.length >= 9 && digits.length <= 15;
}

function isValidEmail(value: string): boolean {
	return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim());
}

function isValidTelegramUsername(value: string): boolean {
	const cleaned = value.trim().replace(/^@/, '');
	return /^[a-zA-Z][a-zA-Z0-9_]{4,31}$/.test(cleaned);
}

// --- Константы ---

const ALLOWED_PROPERTY_TYPES = new Set(['apartment', 'house', 'estate']);
const MIN_PHOTOS = 5;
const MAX_PHOTOS = 25;
const MAX_PHOTO_SIZE_BYTES = 15 * 1024 * 1024;

// --- Валидаторы ---

type ValidatorFn<K extends ValidatedListingField> = (
	value: ListingFormValidatedSlice[K],
	form: ListingFormValues
) => string;

type ValidatorMap = {
	[K in ValidatedListingField]: ValidatorFn<K>;
};

export const listingValidators: ValidatorMap = {
	title(value) {
		const trimmed = value.trim();
		if (!trimmed) return 'Укажите заголовок объявления';
		if (trimmed.length < 10) return 'Заголовок должен быть не короче 10 символов';
		if (trimmed.length > 100) return 'Заголовок не должен превышать 100 символов';
		return '';
	},

	address(value) {
		const trimmed = value.trim();
		if (!trimmed) return 'Укажите адрес';
		if (trimmed.length < 5) return 'Адрес слишком короткий';
		return '';
	},

	propertyType(value) {
		if (!value) return 'Выберите тип жилья';
		if (!ALLOWED_PROPERTY_TYPES.has(value)) return 'Некорректный тип жилья';
		return '';
	},

	size(value) {
		const num = parseNumber(value);
		if (num === null) return 'Укажите площадь';
		if (num <= 10) return 'Площадь должна быть больше 10 м²';
		if (num >= 1000) return 'Площадь должна быть меньше 1000 м²';
		return '';
	},

	floor(value, form) {
		if (form.propertyType !== 'apartment') return '';

		const num = parseInteger(value);
		if (num === null) return 'Укажите этаж';
		if (num < 1) return 'Этаж должен быть не меньше 1';
		if (num > 150) return 'Максимальный этаж: 150';

		const totalFloors = parseInteger(form.totalFloors);
		if (totalFloors !== null && num > totalFloors) {
			return 'Этаж не может быть больше этажности дома';
		}
		return '';
	},

	totalFloors(value, form) {
		const num = parseInteger(value);
		if (num === null) {
			return form.propertyType === 'apartment'
				? 'Укажите этажность дома'
				: 'Укажите этажность объекта';
		}
		if (num < 1) return 'Этажность должна быть не меньше 1';
		if (num > 150) return 'Максимальная этажность: 150';

		if (form.propertyType === 'apartment') {
			const floor = parseInteger(form.floor);
			if (floor !== null && num < floor) {
				return 'Этажность дома не может быть меньше этажа квартиры';
			}
		}
		return '';
	},

	maxGuests(value) {
		const num = parseInteger(value);
		if (num === null) return 'Укажите максимальное число гостей';
		if (num < 1) return 'Минимум 1 гость';
		if (num > 50) return 'Максимум 50 гостей';
		return '';
	},

	bedrooms(value) {
		const num = parseInteger(value);
		if (num === null) return 'Укажите количество комнат';
		if (num < 1) return 'Минимум 1 комната';
		if (num > 30) return 'Максимум 30 комнат';
		return '';
	},

	beds(value) {
		const num = parseInteger(value);
		if (num === null) return 'Укажите количество спальных мест';
		if (num < 1) return 'Минимум 1 спальное место';
		if (num > 30) return 'Максимум 30 спальных мест';
		return '';
	},

	bathrooms(value) {
		const num = parseInteger(value);
		if (num === null) return 'Укажите количество ванных комнат';
		if (num < 1) return 'Минимум 1 ванная комната';
		if (num > 20) return 'Максимум 20 ванных комнат';
		return '';
	},

	basePrice(value) {
		const num = parseNumber(value);
		if (num === null) return 'Укажите базовую цену';
		if (num <= 0) return 'Цена должна быть больше 0';
		return '';
	},

	minNights(value) {
		const num = parseInteger(value);
		if (num === null) return 'Укажите минимальное количество ночей';
		if (num < 1) return 'Минимум 1 ночь';
		if (num > 30) return 'Максимум 30 ночей';
		return '';
	},

	checkInFrom(value) {
		const trimmed = value.trim();
		if (!trimmed) return 'Укажите время заезда';
		if (!isValidTime(trimmed)) return 'Выберите время с шагом 30 минут';
		return '';
	},

	checkOutBefore(value) {
		const trimmed = value.trim();
		if (!trimmed) return 'Укажите время выезда';
		if (!isValidTime(trimmed)) return 'Выберите время с шагом 30 минут';
		return '';
	},

	description(value) {
		const trimmed = value.trim();
		if (!trimmed) return 'Опишите жильё';
		if (trimmed.length < 30) return 'Описание должно быть не короче 30 символов';
		if (trimmed.length > 5000) return 'Описание не должно превышать 5000 символов';
		return '';
	},

	photos(files) {
		if (!Array.isArray(files) || files.length < MIN_PHOTOS) {
			return `Добавьте минимум ${MIN_PHOTOS} фото`;
		}
		if (files.length > MAX_PHOTOS) {
			return `Максимум ${MAX_PHOTOS} фото`;
		}

		for (const f of files) {
			if (typeof File !== 'undefined' && f instanceof File) {
				if (!f.type.startsWith('image/')) {
					return 'Только изображения (JPG/PNG/WEBP)';
				}
				if (f.size > MAX_PHOTO_SIZE_BYTES) {
					return 'Одно из фото слишком большое (до 15MB)';
				}
			}
		}

		return '';
	},

	phone(value) {
		const trimmed = value.trim();
		if (!trimmed) return 'Укажите номер телефона';
		if (!isValidPhone(trimmed)) return 'Некорректный номер телефона';
		return '';
	},

	contactMethods(methods) {
		if (!Array.isArray(methods)) return '';

		for (const method of methods) {
			const trimmedValue = method.value.trim();
			if (!trimmedValue) continue;

			switch (method.type) {
				case 'email':
					if (!isValidEmail(trimmedValue)) {
						return 'Некорректный email адрес';
					}
					break;
				case 'telegram':
					if (!isValidTelegramUsername(trimmedValue) && !isValidPhone(trimmedValue)) {
						return 'Укажите корректный username (@username) или номер телефона';
					}
					break;
				case 'whatsapp':
				case 'viber':
				case 'signal': {
					const labels: Record<string, string> = {
						whatsapp: 'WhatsApp',
						viber: 'Viber',
						signal: 'Signal'
					};
					if (!isValidPhone(trimmedValue)) {
						return `Некорректный номер для ${labels[method.type]}`;
					}
					break;
				}
			}
		}

		return '';
	}
};

export function validateListingField<K extends ValidatedListingField>(
	name: K,
	form: ListingFormValues
): string {
	const value = form[name] as ListingFormValidatedSlice[K];
	return listingValidators[name](value, form);
}

export function validateListingForm(form: ListingFormValues): ListingErrors {
	const errors = {} as ListingErrors;
	for (const name of validatedListingFields) {
		errors[name] = validateListingField(name, form);
	}
	return errors;
}

export function hasListingErrors(errors: ListingErrors): boolean {
	return Object.values(errors).some(Boolean);
}

export const validatedListingFields = Object.keys(listingValidators) as ValidatedListingField[];
