// src/lib/config/amenities.ts
import type { PropertyType } from '$lib/components/card/types';

export type AmenityPresetId = 'must_have' | 'family_friendly' | 'remote_work';

export type AmenityDefinition = {
	id: string;
	label: string;
	icon: string;
	appliesTo?: PropertyType[];
	priority: number;
	highlightScore: number;
	aliases?: string[];
};

export type AmenityCategory = {
	id: string;
	label: string;
	description: string;
	appliesTo?: PropertyType[];
	amenities: AmenityDefinition[];
};

export type AmenityPresetOption = {
	id: AmenityPresetId;
	label: string;
	description: string;
	amenityIds: string[];
};

export type GroupedAmenityCategory = {
	id: string;
	label: string;
	amenities: AmenityDefinition[];
};

const ALL_PROPERTY_TYPES: PropertyType[] = ['apartment', 'house', 'estate'];

// ─────────────────────────────────────────────────────────────────────────────
// CREATION AMENITIES
// Показываются при создании объявления (~65 позиций).
// Принцип: только то, что хост может однозначно подтвердить за 2 минуты.
// ─────────────────────────────────────────────────────────────────────────────

export const creationAmenitiesCategories: AmenityCategory[] = [
	{
		id: 'essentials',
		label: 'Основное',
		description: 'Базовые удобства, которые гости ожидают в первую очередь',
		amenities: [
			{ id: 'wifi', label: 'Wi‑Fi', icon: 'wifi', priority: 100, highlightScore: 100 },
			{ id: 'heating', label: 'Отопление', icon: 'heating', priority: 95, highlightScore: 88 },
			{ id: 'ac', label: 'Кондиционер', icon: 'ac', priority: 92, highlightScore: 84 },
			{
				id: 'hot_water',
				label: 'Горячая вода',
				icon: 'hot-water',
				priority: 90,
				highlightScore: 82,
				aliases: ['hot-water']
			},
			{
				id: 'washer',
				label: 'Стиральная машина',
				icon: 'washer',
				priority: 78,
				highlightScore: 72
			},
			{
				id: 'dryer',
				label: 'Сушильная машина',
				icon: 'dryer',
				priority: 62,
				highlightScore: 48
			},
			{ id: 'tv', label: 'Телевизор', icon: 'tv', priority: 68, highlightScore: 56 },
			{
				id: 'iron',
				label: 'Утюг',
				icon: 'iron',
				priority: 65,
				highlightScore: 52,
				aliases: ['iron']
			},
			{
				id: 'hair_dryer',
				label: 'Фен',
				icon: 'hair-dryer',
				priority: 64,
				highlightScore: 50,
				aliases: ['hair-dryer']
			},
			{
				id: 'towels',
				label: 'Полотенца',
				icon: 'towels',
				priority: 88,
				highlightScore: 80,
				aliases: ['towels']
			},
			{
				id: 'bed_linen',
				label: 'Постельное бельё',
				icon: 'bed-linen',
				priority: 89,
				highlightScore: 81,
				aliases: ['bed-linen']
			}
		]
	},
	{
		id: 'kitchen',
		label: 'Кухня',
		description: 'Оснащение для самостоятельной готовки',
		amenities: [
			{
				id: 'kitchen',
				label: 'Полноценная кухня',
				icon: 'kitchen',
				priority: 93,
				highlightScore: 86
			},
			{ id: 'fridge', label: 'Холодильник', icon: 'fridge', priority: 88, highlightScore: 80 },
			{ id: 'stove', label: 'Плита', icon: 'stove', priority: 76, highlightScore: 68 },
			{ id: 'oven', label: 'Духовка', icon: 'oven', priority: 62, highlightScore: 52 },
			{
				id: 'dishwasher',
				label: 'Посудомоечная машина',
				icon: 'dishwasher',
				priority: 58,
				highlightScore: 44
			},
			{
				id: 'microwave',
				label: 'Микроволновка',
				icon: 'microwave',
				priority: 60,
				highlightScore: 47
			},
			{
				id: 'coffee_machine',
				label: 'Кофемашина',
				icon: 'coffee',
				priority: 54,
				highlightScore: 40,
				aliases: ['coffee']
			},
			{
				id: 'kettle',
				label: 'Чайник',
				icon: 'kettle',
				priority: 56,
				highlightScore: 42,
				aliases: ['kettle']
			},
			{
				id: 'toaster',
				label: 'Тостер',
				icon: 'toaster',
				priority: 44,
				highlightScore: 30,
				aliases: ['toaster']
			},
			{
				id: 'dining_area',
				label: 'Обеденная зона',
				icon: 'dining-area',
				priority: 70,
				highlightScore: 60,
				aliases: ['dining-area']
			}
		]
	},
	{
		id: 'bedroom_bathroom',
		label: 'Спальня и ванная',
		description: 'Комфорт для сна и личной гигиены',
		amenities: [
			{
				id: 'extra_pillows',
				label: 'Дополнительные подушки и одеяла',
				icon: 'extra-pillows',
				priority: 60,
				highlightScore: 46,
				aliases: ['extra-pillows']
			},
			{
				id: 'blackout_curtains',
				label: 'Блэкаут-шторы',
				icon: 'blackout-curtains',
				priority: 58,
				highlightScore: 44,
				aliases: ['blackout-curtains']
			},
			{
				id: 'bathtub',
				label: 'Ванна',
				icon: 'bathtub',
				priority: 66,
				highlightScore: 55,
				aliases: ['bathtub']
			},
			{
				id: 'shower',
				label: 'Душ',
				icon: 'shower',
				priority: 85,
				highlightScore: 76,
				aliases: ['shower']
			},
			{
				id: 'bidet',
				label: 'Биде',
				icon: 'bidet',
				priority: 40,
				highlightScore: 28,
				aliases: ['bidet']
			}
		]
	},
	{
		id: 'work',
		label: 'Работа',
		description: 'Для тех, кто совмещает поездку с удалённой работой',
		amenities: [
			{
				id: 'workspace',
				label: 'Рабочее место',
				icon: 'desk',
				priority: 82,
				highlightScore: 74,
				aliases: ['desk']
			},
			{
				id: 'monitor',
				label: 'Внешний монитор',
				icon: 'monitor',
				priority: 50,
				highlightScore: 36,
				aliases: ['monitor']
			}
		]
	},
	{
		id: 'comfort',
		label: 'Комфорт и досуг',
		description: 'Удобства, которые улучшают впечатление от проживания',
		amenities: [
			{
				id: 'balcony',
				label: 'Балкон / терраса',
				icon: 'balcony',
				priority: 60,
				highlightScore: 50,
				appliesTo: ['apartment', 'house']
			},
			{
				id: 'garden',
				label: 'Сад / двор',
				icon: 'garden',
				priority: 74,
				highlightScore: 66,
				appliesTo: ['house', 'estate']
			},
			{
				id: 'pool',
				label: 'Бассейн',
				icon: 'pool',
				priority: 76,
				highlightScore: 70,
				appliesTo: ['house', 'estate']
			},
			{
				id: 'hot_tub',
				label: 'Джакузи / горячая ванна',
				icon: 'hot-tub',
				priority: 64,
				highlightScore: 58,
				appliesTo: ['house', 'estate'],
				aliases: ['hot-tub']
			},
			{ id: 'gym', label: 'Спортзал', icon: 'gym', priority: 49, highlightScore: 38 },
			{
				id: 'sauna',
				label: 'Сауна',
				icon: 'sauna',
				priority: 70,
				highlightScore: 63,
				appliesTo: ['house', 'estate']
			},
			{
				id: 'fireplace',
				label: 'Камин',
				icon: 'fireplace',
				priority: 62,
				highlightScore: 54,
				appliesTo: ['house', 'estate'],
				aliases: ['fireplace']
			},
			{ id: 'view', label: 'Красивый вид', icon: 'view', priority: 59, highlightScore: 54 },
			{
				id: 'sofa',
				label: 'Диван / зона отдыха',
				icon: 'sofa',
				priority: 55,
				highlightScore: 42,
				aliases: ['sofa']
			},
			{
				id: 'games',
				label: 'Настольные игры',
				icon: 'games',
				priority: 42,
				highlightScore: 30,
				aliases: ['games']
			},
			{
				id: 'books',
				label: 'Книги',
				icon: 'books',
				priority: 38,
				highlightScore: 26,
				aliases: ['books']
			}
		]
	},
	{
		id: 'family',
		label: 'Для семьи',
		description: 'Удобства для путешествий с детьми и питомцами',
		amenities: [
			{ id: 'crib', label: 'Детская кроватка', icon: 'crib', priority: 66, highlightScore: 57 },
			{
				id: 'high_chair',
				label: 'Детский стульчик',
				icon: 'high-chair',
				priority: 61,
				highlightScore: 51,
				aliases: ['high-chair']
			},
			{
				id: 'pets_allowed',
				label: 'Можно с животными',
				icon: 'pet',
				priority: 71,
				highlightScore: 62,
				aliases: ['pet']
			},
			{
				id: 'children_toys',
				label: 'Игрушки',
				icon: 'children-toys',
				priority: 46,
				highlightScore: 33,
				aliases: ['children-toys']
			},
			{
				id: 'baby_bath',
				label: 'Детская ванночка',
				icon: 'baby-bath',
				priority: 44,
				highlightScore: 31,
				aliases: ['baby-bath']
			}
		]
	},
	{
		id: 'safety',
		label: 'Безопасность',
		description: 'Базовые элементы безопасности жилья',
		amenities: [
			{
				id: 'smoke_alarm',
				label: 'Датчик дыма',
				icon: 'smoke',
				priority: 89,
				highlightScore: 78,
				aliases: ['smoke']
			},
			{
				id: 'co_alarm',
				label: 'Датчик угарного газа',
				icon: 'co',
				priority: 87,
				highlightScore: 76,
				aliases: ['co']
			},
			{
				id: 'fire_extinguisher',
				label: 'Огнетушитель',
				icon: 'fire',
				priority: 73,
				highlightScore: 60,
				aliases: ['fire']
			},
			{
				id: 'first_aid',
				label: 'Аптечка',
				icon: 'first-aid',
				priority: 75,
				highlightScore: 64,
				aliases: ['first-aid']
			},
			{
				id: 'lockbox',
				label: 'Кодовый замок / сейф',
				icon: 'lock',
				priority: 69,
				highlightScore: 55,
				aliases: ['lock']
			}
		]
	},
	{
		id: 'access',
		label: 'Доступ и парковка',
		description: 'Как гость попадает в жильё и где оставляет машину',
		amenities: [
			{
				id: 'self_checkin',
				label: 'Самостоятельный заезд',
				icon: 'self-checkin',
				priority: 81,
				highlightScore: 73,
				aliases: ['self-checkin']
			},
			{
				id: 'elevator',
				label: 'Лифт',
				icon: 'elevator',
				priority: 63,
				highlightScore: 49,
				appliesTo: ['apartment']
			},
			{ id: 'parking', label: 'Парковка', icon: 'parking', priority: 84, highlightScore: 75 },
			{
				id: 'ev_charger',
				label: 'Зарядка для электромобиля',
				icon: 'ev-charger',
				priority: 48,
				highlightScore: 35,
				aliases: ['ev-charger']
			}
		]
	},
	{
		id: 'outdoor',
		label: 'На улице',
		description: 'Уличные зоны для отдыха и встреч',
		appliesTo: ['house', 'estate'],
		amenities: [
			{ id: 'bbq', label: 'Мангал / барбекю', icon: 'bbq', priority: 72, highlightScore: 65 },
			{
				id: 'outdoor_seating',
				label: 'Зона отдыха на улице',
				icon: 'outdoor',
				priority: 70,
				highlightScore: 61,
				aliases: ['outdoor']
			},
			{
				id: 'outdoor_shower',
				label: 'Уличный душ',
				icon: 'outdoor-shower',
				priority: 45,
				highlightScore: 32,
				appliesTo: ['house', 'estate'],
				aliases: ['outdoor-shower']
			},
			{
				id: 'sun_loungers',
				label: 'Шезлонги',
				icon: 'sun-loungers',
				priority: 47,
				highlightScore: 34,
				appliesTo: ['house', 'estate'],
				aliases: ['sun-loungers']
			}
		]
	},
	{
		id: 'accessibility',
		label: 'Доступность',
		description: 'Для гостей с ограниченными возможностями',
		amenities: [
			{
				id: 'wide_entrance',
				label: 'Широкий вход (для коляски / кресла)',
				icon: 'wide-entrance',
				priority: 55,
				highlightScore: 45,
				aliases: ['wide-entrance']
			},
			{
				id: 'step_free',
				label: 'Вход без ступенек',
				icon: 'step-free',
				priority: 54,
				highlightScore: 44,
				aliases: ['step-free']
			},
			{
				id: 'accessible_bathroom',
				label: 'Ванная для людей с ОВЗ',
				icon: 'accessible-bathroom',
				priority: 52,
				highlightScore: 42,
				aliases: ['accessible-bathroom']
			}
		]
	}
];

// ─────────────────────────────────────────────────────────────────────────────
// FULL AMENITIES LIST
// Используется для детальной страницы и расширенного редактирования.
// Включает всё из creationAmenitiesCategories + редкие/специфичные позиции.
// ─────────────────────────────────────────────────────────────────────────────

const FULL_ONLY_CATEGORIES: AmenityCategory[] = [
	{
		id: 'services',
		label: 'Услуги',
		description: 'Дополнительные сервисы от хоста',
		amenities: [
			{
				id: 'breakfast',
				label: 'Завтрак включён',
				icon: 'breakfast',
				priority: 60,
				highlightScore: 50,
				aliases: ['breakfast']
			},
			{
				id: 'long_term_stays',
				label: 'Длительное проживание (от месяца)',
				icon: 'long-term',
				priority: 50,
				highlightScore: 40,
				aliases: ['long-term']
			}
		]
	},
	{
		id: 'accessibility_extended',
		label: 'Доступность (расширенная)',
		description: 'Дополнительные опции для гостей с ОВЗ',
		amenities: [
			{
				id: 'pool_hoist',
				label: 'Подъёмник для бассейна',
				icon: 'pool-hoist',
				priority: 30,
				highlightScore: 25,
				appliesTo: ['house', 'estate'],
				aliases: ['pool-hoist']
			}
		]
	}
];

export const fullAmenitiesCategories: AmenityCategory[] = [
	...creationAmenitiesCategories,
	...FULL_ONLY_CATEGORIES
];

// ─────────────────────────────────────────────────────────────────────────────
// По умолчанию экспортируем список для создания объявления.
// Весь остальной код (getAmenityHighlights и т.д.) работает без изменений.
// ─────────────────────────────────────────────────────────────────────────────

export const amenitiesCategories = creationAmenitiesCategories;

// ─────────────────────────────────────────────────────────────────────────────
// Presets
// ─────────────────────────────────────────────────────────────────────────────

const PRESET_TEMPLATES: Array<{
	id: AmenityPresetId;
	label: string;
	description: string;
	amenityIds: string[];
}> = [
	{
		id: 'must_have',
		label: 'Обязательный минимум',
		description: 'Базовый набор, который ждут большинство гостей',
		amenityIds: [
			'wifi',
			'heating',
			'hot_water',
			'towels',
			'bed_linen',
			'kitchen',
			'fridge',
			'smoke_alarm',
			'first_aid'
		]
	},
	{
		id: 'family_friendly',
		label: 'Для семей',
		description: 'Подборка для гостей с детьми и питомцами',
		amenityIds: ['crib', 'high_chair', 'washer', 'microwave', 'pets_allowed', 'children_toys']
	},
	{
		id: 'remote_work',
		label: 'Для удалённой работы',
		description: 'Подходит для workation и длительных поездок',
		amenityIds: ['wifi', 'workspace', 'coffee_machine', 'self_checkin', 'monitor']
	}
];

// ─────────────────────────────────────────────────────────────────────────────
// Internal maps — строятся по fullAmenitiesCategories,
// чтобы resolveAmenity работал для любых id включая full-only.
// ─────────────────────────────────────────────────────────────────────────────

function normalizeLookupValue(value: string): string {
	return value.trim().toLowerCase();
}

function appliesToPropertyType(
	appliesTo: PropertyType[] | undefined,
	propertyType: PropertyType
): boolean {
	return !appliesTo || appliesTo.includes(propertyType);
}

const _amenityMap = new Map<string, AmenityDefinition>();
const _amenityAliasMap = new Map<string, string>();

for (const category of fullAmenitiesCategories) {
	for (const amenity of category.amenities) {
		_amenityMap.set(amenity.id, amenity);
		_amenityAliasMap.set(normalizeLookupValue(amenity.id), amenity.id);
		_amenityAliasMap.set(normalizeLookupValue(amenity.label), amenity.id);
		for (const alias of amenity.aliases ?? []) {
			_amenityAliasMap.set(normalizeLookupValue(alias), amenity.id);
		}
	}
}

export const amenityMap = _amenityMap;
export const allAmenityIds = Array.from(_amenityMap.keys());

export function normalizeAmenityId(idOrAlias: string): string {
	const normalizedKey = normalizeLookupValue(idOrAlias);
	return _amenityAliasMap.get(normalizedKey) ?? idOrAlias.trim();
}

export function normalizeAmenityIds(ids: string[]): string[] {
	const seen = new Set<string>();
	const normalized: string[] = [];
	for (const id of ids) {
		const canonical = normalizeAmenityId(id);
		if (!canonical || seen.has(canonical)) continue;
		seen.add(canonical);
		normalized.push(canonical);
	}
	return normalized;
}

export function resolveAmenity(idOrLabel: string): AmenityDefinition {
	const normalizedId = normalizeAmenityId(idOrLabel);
	const byId = _amenityMap.get(normalizedId);
	if (byId) return byId;
	for (const def of _amenityMap.values()) {
		if (def.label === idOrLabel) return def;
	}
	return {
		id: normalizedId,
		label: idOrLabel,
		icon: '',
		appliesTo: ALL_PROPERTY_TYPES,
		priority: 0,
		highlightScore: 0
	};
}

export function getAmenityCategoriesForPropertyType(
	propertyType: PropertyType,
	source: AmenityCategory[] = amenitiesCategories
): AmenityCategory[] {
	return source
		.filter((category) => appliesToPropertyType(category.appliesTo, propertyType))
		.map((category) => ({
			...category,
			amenities: category.amenities.filter((amenity) =>
				appliesToPropertyType(amenity.appliesTo, propertyType)
			)
		}))
		.filter((category) => category.amenities.length > 0);
}

export function getAmenityIdsForPropertyType(
	propertyType: PropertyType,
	source?: AmenityCategory[]
): string[] {
	return getAmenityCategoriesForPropertyType(propertyType, source).flatMap((category) =>
		category.amenities.map((amenity) => amenity.id)
	);
}

export function getAmenityPresetOptions(propertyType: PropertyType): AmenityPresetOption[] {
	const available = new Set(getAmenityIdsForPropertyType(propertyType));
	return PRESET_TEMPLATES.map((preset) => ({
		...preset,
		amenityIds: preset.amenityIds.filter((id) => available.has(id))
	})).filter((preset) => preset.amenityIds.length > 0);
}

export function getAmenityHighlights(
	selectedAmenities: string[],
	propertyType: PropertyType,
	limit = 8
): AmenityDefinition[] {
	const normalized = normalizeAmenityIds(selectedAmenities);
	return normalized
		.map((id) => resolveAmenity(id))
		.filter((amenity) => appliesToPropertyType(amenity.appliesTo, propertyType))
		.sort((a, b) => {
			if (b.highlightScore !== a.highlightScore) return b.highlightScore - a.highlightScore;
			if (b.priority !== a.priority) return b.priority - a.priority;
			return a.label.localeCompare(b.label, 'ru-RU');
		})
		.slice(0, limit);
}

export function groupSelectedAmenitiesByCategory(
	selectedAmenities: string[],
	propertyType: PropertyType,
	source: AmenityCategory[] = fullAmenitiesCategories
): GroupedAmenityCategory[] {
	const selectedSet = new Set(normalizeAmenityIds(selectedAmenities));
	return getAmenityCategoriesForPropertyType(propertyType, source)
		.map((category) => ({
			id: category.id,
			label: category.label,
			amenities: category.amenities.filter((amenity) => selectedSet.has(amenity.id))
		}))
		.filter((category) => category.amenities.length > 0);
}
