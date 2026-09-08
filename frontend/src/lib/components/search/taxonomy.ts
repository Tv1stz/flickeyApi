import type { Listing, PropertyType } from '$lib/components/card/types';
import { matchesSearch, normalizeForSearch } from '$lib/utils/searchNormalization';

export interface PropertyTypeOption {
	id: PropertyType | 'any';
	label: string;
	emoji: string;
	description: string;
}

export const PROPERTY_TYPE_OPTIONS: PropertyTypeOption[] = [
	{ id: 'any', label: 'Любой тип', emoji: '🏠', description: 'Все варианты жилья' },
	{ id: 'apartment', label: 'Квартира', emoji: '🏢', description: 'Апартаменты, студии' },
	{ id: 'house', label: 'Дом / Коттедж', emoji: '🏡', description: 'Частные дома' },
	{ id: 'estate', label: 'Усадьба', emoji: '🌿', description: 'Загородная недвижимость' }
];

export interface CityOption {
	id: string;
	label: string;
	region?: string;
	count?: number;
}

export const CITY_OPTIONS: CityOption[] = [
	{ id: 'minsk', label: 'Минск', region: 'Минская область' },
	{ id: 'brest', label: 'Брест', region: 'Брестская область' },
	{ id: 'grodno', label: 'Гродно', region: 'Гродненская область' },
	{ id: 'gomel', label: 'Гомель', region: 'Гомельская область' },
	{ id: 'vitebsk', label: 'Витебск', region: 'Витебская область' },
	{ id: 'mogilev', label: 'Могилёв', region: 'Могилёвская область' },
	{ id: 'minsk_region', label: 'Минский район', region: 'Пригород Минска' }
];

function normalizeCityKey(value: string): string {
	return normalizeForSearch(value);
}

function slugifyCityId(value: string): string {
	const cleaned = normalizeCityKey(value).replace(/[^a-z0-9а-яё]+/gi, '_');
	return cleaned.replace(/^_+|_+$/g, '') || 'city';
}

export function getCityOptionsFromListings(listings: Listing[]): CityOption[] {
	const knownByLabel = new Map(
		CITY_OPTIONS.map((option) => [normalizeCityKey(option.label), option])
	);
	const counts = new Map<string, { option: CityOption; count: number }>();

	listings.forEach((listing) => {
		const city = listing.location?.city?.trim();
		if (!city) return;
		const key = normalizeCityKey(city);
		const entry = counts.get(key);
		if (entry) {
			entry.count += 1;
			return;
		}

		const known = knownByLabel.get(key);
		const option = known ? { ...known } : { id: slugifyCityId(city), label: city };
		counts.set(key, { option, count: 1 });
	});

	return [...counts.values()]
		.map(({ option, count }) => ({ ...option, count }))
		.sort((a, b) => a.label.localeCompare(b.label, 'ru'));
}

export function getPopularCitiesFromListings(listings: Listing[], limit: number = 6): CityOption[] {
	return getCityOptionsFromListings(listings)
		.slice()
		.sort((a, b) => (b.count ?? 0) - (a.count ?? 0) || a.label.localeCompare(b.label, 'ru'))
		.slice(0, limit);
}

export interface AmenityFilterOption {
	id: string;
	label: string;
	emoji: string;
}

export const AMENITY_FILTER_OPTIONS: AmenityFilterOption[] = [
	{ id: 'wifi', label: 'Wi-Fi', emoji: '📶' },
	{ id: 'parking', label: 'Парковка', emoji: '🅿️' },
	{ id: 'pool', label: 'Бассейн', emoji: '🏊' },
	{ id: 'sauna', label: 'Сауна', emoji: '🧖' },
	{ id: 'ac', label: 'Кондиционер', emoji: '❄️' },
	{ id: 'washer', label: 'Стиральная машина', emoji: '🌀' },
	{ id: 'kitchen', label: 'Кухня', emoji: '🍳' },
	{ id: 'pets_allowed', label: 'Можно с животными', emoji: '🐾' },
	{ id: 'bbq', label: 'Мангал / барбекю', emoji: '🔥' },
	{ id: 'fireplace', label: 'Камин', emoji: '🕯️' },
	{ id: 'workspace', label: 'Рабочее место', emoji: '💻' },
	{ id: 'self_checkin', label: 'Самостоятельный заезд', emoji: '🔑' }
];

export interface SearchParams {
	location: string;
	propertyType: PropertyType | 'any';
	adults: number;
	children: number;
	checkin: string | null;  // YYYY-MM-DD
	checkout: string | null; // YYYY-MM-DD
}

export interface SearchFilters {
	priceMin: number | null;
	priceMax: number | null;
	bedroomsMin: number | null;
	amenities: string[];
}

export const DEFAULT_SEARCH_PARAMS: SearchParams = {
	location: '',
	propertyType: 'any',
	adults: 0,
	children: 0,
	checkin: null,
	checkout: null
};

export const DEFAULT_FILTERS: SearchFilters = {
	priceMin: null,
	priceMax: null,
	bedroomsMin: null,
	amenities: []
};

export function filterListings(
	listings: Listing[],
	params: SearchParams,
	filters: SearchFilters
): Listing[] {
	const totalGuests = params.adults + params.children;
	return listings.filter((listing) => {
		if (params.location) {
			const q = params.location;
			const city = listing.location?.city ?? '';
			const addr = listing.address ?? '';
			const dist = listing.location?.district ?? '';
			if (!matchesSearch(q, city) && !matchesSearch(q, addr) && !matchesSearch(q, dist))
				return false;
		}
		if (params.propertyType !== 'any' && listing.propertyType !== params.propertyType) return false;
		if (totalGuests > 0 && listing.maxGuests < totalGuests) return false;
		if (params.children > 0 && !listing.rules.childrenAllowed) return false;
		if (filters.priceMin !== null && listing.pricePerNight < filters.priceMin) return false;
		if (filters.priceMax !== null && listing.pricePerNight > filters.priceMax) return false;
		if (filters.bedroomsMin !== null && listing.bedrooms < filters.bedroomsMin) return false;
		if (filters.amenities.length > 0) {
			const set = new Set(listing.amenities);
			if (!filters.amenities.every((id) => set.has(id))) return false;
		}
		return true;
	});
}

export function hasActiveSearch(params: SearchParams): boolean {
	return (
		params.location.trim().length > 0 ||
		params.propertyType !== 'any' ||
		params.adults + params.children > 0 ||
		Boolean(params.checkin || params.checkout)
	);
}

export function isSearchComplete(params: SearchParams): boolean {
	return (
		(params.location.trim().length > 0 && params.adults + params.children > 0) ||
		Boolean(params.checkin && params.checkout)
	);
}

export function countActiveFilters(filters: SearchFilters): number {
	let c = 0;
	if (filters.priceMin !== null || filters.priceMax !== null) c++;
	if (filters.bedroomsMin !== null) c++;
	c += filters.amenities.length;
	return c;
}

export function buildSearchUrl(params: SearchParams): string {
	const sp = new URLSearchParams();
	if (params.location.trim()) sp.set('location', params.location.trim());
	if (params.propertyType && params.propertyType !== 'any') sp.set('propertyType', params.propertyType);
	if (params.adults > 0) sp.set('adults', String(params.adults));
	if (params.children > 0) sp.set('children', String(params.children));
	if (params.checkin) sp.set('checkin', params.checkin);
	if (params.checkout) sp.set('checkout', params.checkout);
	const qs = sp.toString();
	return qs ? `/search?${qs}` : '/search';
}
