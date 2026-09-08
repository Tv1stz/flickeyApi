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

export interface FilterCategory {
	id: string;
	label: string;
	emoji: string;
	description: string;
	priority: number;
}

export const FILTER_CATEGORIES: FilterCategory[] = [
	{
		id: 'popular',
		label: 'Популярное',
		emoji: '⭐',
		description: 'Самые востребованные удобства',
		priority: 100
	},
	{
		id: 'essentials',
		label: 'Основное',
		emoji: '🏠',
		description: 'Базовые удобства для комфортного проживания',
		priority: 90
	},
	{
		id: 'work',
		label: 'Для работы',
		emoji: '💼',
		description: 'Удобства для удаленной работы',
		priority: 85
	},
	{
		id: 'family',
		label: 'Для семьи',
		emoji: '👨‍👩‍👧‍👦',
		description: 'Подходит для отдыха с детьми',
		priority: 80
	},
	{
		id: 'leisure',
		label: 'Отдых и развлечения',
		emoji: '🎉',
		description: 'Для активного отдыха',
		priority: 75
	},
	{
		id: 'outdoor',
		label: 'На улице',
		emoji: '🌳',
		description: 'Уличные зоны и активности',
		priority: 70
	},
	{
		id: 'accessibility',
		label: 'Доступность',
		emoji: '♿',
		description: 'Для людей с ограниченными возможностями',
		priority: 65
	},
	{
		id: 'safety',
		label: 'Безопасность',
		emoji: '🔒',
		description: 'Меры безопасности',
		priority: 60
	}
];

export interface AmenityFilterOption {
	id: string;
	label: string;
	emoji: string;
	category: string;
}

export const AMENITY_FILTER_OPTIONS: AmenityFilterOption[] = [
	// Популярное
	{ id: 'wifi', label: 'Wi-Fi', emoji: '📶', category: 'popular' },
	{ id: 'parking', label: 'Парковка', emoji: '🅿️', category: 'popular' },
	{ id: 'kitchen', label: 'Кухня', emoji: '🍳', category: 'popular' },
	{ id: 'ac', label: 'Кондиционер', emoji: '❄️', category: 'popular' },
	{ id: 'washer', label: 'Стиральная машина', emoji: '🌀', category: 'popular' },
	{ id: 'self_checkin', label: 'Самостоятельный заезд', emoji: '🔑', category: 'popular' },

	// Основное
	{ id: 'heating', label: 'Отопление', emoji: '🔥', category: 'essentials' },
	{ id: 'hot_water', label: 'Горячая вода', emoji: '💧', category: 'essentials' },
	{ id: 'tv', label: 'Телевизор', emoji: '📺', category: 'essentials' },
	{ id: 'iron', label: 'Утюг', emoji: '👔', category: 'essentials' },
	{ id: 'hair_dryer', label: 'Фен', emoji: '💨', category: 'essentials' },
	{ id: 'towels', label: 'Полотенца', emoji: '🧺', category: 'essentials' },
	{ id: 'bed_linen', label: 'Постельное белье', emoji: '🛏️', category: 'essentials' },
	{ id: 'dryer', label: 'Сушилка для белья', emoji: '🌬️', category: 'essentials' },

	// Для работы
	{ id: 'workspace', label: 'Рабочее место', emoji: '💻', category: 'work' },
	{ id: 'monitor', label: 'Монитор', emoji: '🖥️', category: 'work' },

	// Для семьи
	{ id: 'crib', label: 'Детская кроватка', emoji: '👶', category: 'family' },
	{ id: 'high_chair', label: 'Детский стульчик', emoji: '🪑', category: 'family' },
	{ id: 'children_toys', label: 'Детские игрушки', emoji: '🧸', category: 'family' },
	{ id: 'baby_bath', label: 'Детская ванночка', emoji: '🛁', category: 'family' },

	// Отдых и развлечения
	{ id: 'pool', label: 'Бассейн', emoji: '🏊', category: 'leisure' },
	{ id: 'sauna', label: 'Сауна', emoji: '🧖', category: 'leisure' },
	{ id: 'hot_tub', label: 'Джакузи', emoji: '♨️', category: 'leisure' },
	{ id: 'gym', label: 'Тренажерный зал', emoji: '🏋️', category: 'leisure' },
	{ id: 'fireplace', label: 'Камин', emoji: '🕯️', category: 'leisure' },
	{ id: 'games', label: 'Настольные игры', emoji: '🎲', category: 'leisure' },
	{ id: 'books', label: 'Книги', emoji: '📚', category: 'leisure' },

	// На улице
	{ id: 'bbq', label: 'Мангал / барбекю', emoji: '🍖', category: 'outdoor' },
	{ id: 'balcony', label: 'Балкон', emoji: '🪴', category: 'outdoor' },
	{ id: 'garden', label: 'Сад', emoji: '🌻', category: 'outdoor' },
	{ id: 'outdoor_seating', label: 'Зона отдыха на улице', emoji: '🪑', category: 'outdoor' },
	{ id: 'sun_loungers', label: 'Шезлонги', emoji: '🏖️', category: 'outdoor' },
	{ id: 'view', label: 'Красивый вид', emoji: '🌅', category: 'outdoor' },

	// Доступность
	{ id: 'elevator', label: 'Лифт', emoji: '🛗', category: 'accessibility' },
	{ id: 'wide_entrance', label: 'Широкий вход', emoji: '🚪', category: 'accessibility' },
	{ id: 'step_free', label: 'Без ступенек', emoji: '♿', category: 'accessibility' },
	{ id: 'accessible_bathroom', label: 'Доступная ванная', emoji: '🚿', category: 'accessibility' },

	// Безопасность
	{ id: 'smoke_alarm', label: 'Датчик дыма', emoji: '🚨', category: 'safety' },
	{ id: 'co_alarm', label: 'Датчик CO', emoji: '⚠️', category: 'safety' },
	{ id: 'fire_extinguisher', label: 'Огнетушитель', emoji: '🧯', category: 'safety' },
	{ id: 'first_aid', label: 'Аптечка', emoji: '💊', category: 'safety' },
	{ id: 'lockbox', label: 'Кодовый замок', emoji: '🔐', category: 'safety' },

	// Дополнительно
	{ id: 'pets_allowed', label: 'Можно с животными', emoji: '🐾', category: 'popular' },
	{ id: 'fridge', label: 'Холодильник', emoji: '🧊', category: 'essentials' },
	{ id: 'stove', label: 'Плита', emoji: '🔥', category: 'essentials' },
	{ id: 'oven', label: 'Духовка', emoji: '🍕', category: 'essentials' },
	{ id: 'dishwasher', label: 'Посудомоечная машина', emoji: '🍽️', category: 'essentials' },
	{ id: 'microwave', label: 'Микроволновка', emoji: '📻', category: 'essentials' },
	{ id: 'coffee_machine', label: 'Кофемашина', emoji: '☕', category: 'essentials' },
	{ id: 'kettle', label: 'Чайник', emoji: '🫖', category: 'essentials' },
	{ id: 'ev_charger', label: 'Зарядка для электромобиля', emoji: '🔌', category: 'popular' }
];

export interface RecentSearch {
	params: SearchParams;
	filters: SearchFilters;
	timestamp: number;
	id: string;
}

export interface SearchParams {
	location: string;
	propertyType: PropertyType | 'any';
	adults: number;
	children: number;
	checkin?: string | null;  // YYYY-MM-DD
	checkout?: string | null; // YYYY-MM-DD
}

export interface SearchFilters {
	priceMin: number | null;
	priceMax: number | null;
	bedsMin?: number | null;
	bathroomsMin?: number | null;
	bedroomsMin?: number | null;
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
	bedsMin: null,
	bathroomsMin: null,
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
			const title = listing.title ?? '';
			const cityLine = [dist, city].filter(Boolean).join(', ');
			const haystacks = [city, dist, addr, title, cityLine].filter(Boolean);
			if (!haystacks.some((h) => matchesSearch(q, h))) return false;
		}

		if (params.propertyType !== 'any' && listing.propertyType !== params.propertyType) return false;
		if (totalGuests > 0 && listing.maxGuests < totalGuests) return false;
		if (params.children > 0 && !listing.rules.childrenAllowed) return false;
		if (filters.priceMin !== null && listing.pricePerNight < filters.priceMin) return false;
		if (filters.priceMax !== null && listing.pricePerNight > filters.priceMax) return false;
		if (filters.bedsMin && listing.beds < filters.bedsMin) return false;
		if (filters.bathroomsMin && listing.bathrooms < filters.bathroomsMin) return false;
		if (filters.bedroomsMin && listing.bedrooms < filters.bedroomsMin) return false;

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
	if (filters.bedsMin != null) c++;
	if (filters.bathroomsMin != null) c++;
	if (filters.bedroomsMin != null) c++;
	c += filters.amenities.length;
	return c;
}

export function getAmenitiesByCategory(categoryId: string): AmenityFilterOption[] {
	return AMENITY_FILTER_OPTIONS.filter((a) => a.category === categoryId);
}

export interface CategoryWithAmenities {
	category: FilterCategory;
	amenities: AmenityFilterOption[];
}

export function getCategoriesWithAmenities(): CategoryWithAmenities[] {
	return FILTER_CATEGORIES.map((category) => ({
		category,
		amenities: getAmenitiesByCategory(category.id)
	})).filter((item) => item.amenities.length > 0);
}

export function getPopularAmenities(limit: number = 6): AmenityFilterOption[] {
	return getAmenitiesByCategory('popular').slice(0, limit);
}

export function filterAmenitiesByPropertyType(
	amenities: AmenityFilterOption[],
	propertyType: PropertyType | 'any'
): AmenityFilterOption[] {
	if (propertyType === 'any') return amenities;

	return amenities.filter((amenity) => {
		const definition = AMENITY_DEFINITIONS.get(amenity.id);
		if (!definition) return true;
		if (!definition.appliesTo) return true;
		return definition.appliesTo.includes(propertyType);
	});
}

const AMENITY_DEFINITIONS = new Map<string, { appliesTo?: PropertyType[] }>([
	['wifi', {}],
	['heating', {}],
	['ac', {}],
	['hot_water', {}],
	['washer', {}],
	['dryer', {}],
	['tv', {}],
	['iron', {}],
	['hair_dryer', {}],
	['towels', {}],
	['bed_linen', {}],
	['kitchen', {}],
	['fridge', {}],
	['stove', {}],
	['oven', {}],
	['dishwasher', {}],
	['microwave', {}],
	['coffee_machine', {}],
	['kettle', {}],
	['workspace', {}],
	['monitor', {}],
	['crib', {}],
	['high_chair', {}],
	['children_toys', {}],
	['baby_bath', {}],
	['pets_allowed', {}],
	['balcony', { appliesTo: ['apartment', 'house'] }],
	['garden', { appliesTo: ['house', 'estate'] }],
	['pool', { appliesTo: ['house', 'estate'] }],
	['hot_tub', { appliesTo: ['house', 'estate'] }],
	['gym', {}],
	['sauna', { appliesTo: ['house', 'estate'] }],
	['fireplace', { appliesTo: ['house', 'estate'] }],
	['view', {}],
	['games', {}],
	['books', {}],
	['bbq', { appliesTo: ['house', 'estate'] }],
	['outdoor_seating', { appliesTo: ['house', 'estate'] }],
	['sun_loungers', { appliesTo: ['house', 'estate'] }],
	['elevator', { appliesTo: ['apartment'] }],
	['wide_entrance', {}],
	['step_free', {}],
	['accessible_bathroom', {}],
	['smoke_alarm', {}],
	['co_alarm', {}],
	['fire_extinguisher', {}],
	['first_aid', {}],
	['lockbox', {}],
	['self_checkin', {}],
	['parking', {}],
	['ev_charger', {}]
]);

export interface PriceDistribution {
	price: number;
	count: number;
}

function getSyntheticRangeForSinglePrice(price: number): { min: number; max: number } {
	const syntheticRange = Math.max(price * 0.3, 100);
	return {
		min: Math.max(0, price - syntheticRange / 2),
		max: price + syntheticRange / 2
	};
}

export function getPriceDistribution(
	listings: Listing[],
	sliderMin: number,
	sliderMax: number,
	buckets: number = 40
): PriceDistribution[] {
	if (
		listings.length === 0 ||
		buckets <= 0 ||
		!Number.isFinite(sliderMin) ||
		!Number.isFinite(sliderMax)
	)
		return [];

	const range = sliderMax - sliderMin;
	if (range <= 0) return [];

	const bucketSize = range / buckets;
	const counts = Array<number>(buckets).fill(0);

	for (const listing of listings) {
		const price = listing.pricePerNight;
		if (!Number.isFinite(price)) continue;

		const bucketIndex = Math.max(
			0,
			Math.min(buckets - 1, Math.floor((price - sliderMin) / bucketSize))
		);
		counts[bucketIndex] += 1;
	}

	return counts.map((count, index) => ({
		price: Math.round(sliderMin + (index + 0.5) * bucketSize),
		count
	}));
}

export function getPriceRange(listings: Listing[]): { min: number; max: number } {
	if (listings.length === 0) return { min: 0, max: 1000 };

	const prices = listings.map((l) => l.pricePerNight);
	const actualMin = Math.min(...prices);
	const actualMax = Math.max(...prices);
	const range = actualMax - actualMin;

	if (range === 0) {
		const { min: syntheticMin, max: syntheticMax } = getSyntheticRangeForSinglePrice(actualMin);
		return {
			min: Math.floor(syntheticMin / 10) * 10,
			max: Math.ceil(syntheticMax / 10) * 10
		};
	}

	const padding = Math.max(10, Math.round(range * 0.1));

	return {
		min: Math.max(0, Math.floor((actualMin - padding) / 10) * 10),
		max: Math.ceil((actualMax + padding) / 10) * 10
	};
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
