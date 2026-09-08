// src/lib/services/geocoding.ts

import { nominatimQueue } from './requestQueue';
import { persistedGeoCache, persistGeoEntry } from './localStorage';

export interface ParsedAddress {
	country: string;
	city: string;
	district: string;
	street: string;
	house: string;
	displayName: string;
	lat: number;
	lng: number;
	confidence: 'full' | 'no-house' | 'street-only';
	raw?: unknown;
}

// ─── Кэш ─────────────────────────────────────────────────────────────────────

const reverseCache = new Map<string, ParsedAddress>();
const searchCache = new Map<string, ParsedAddress[]>();
const MAX_CACHE = 300;

function safeSet<V>(cache: Map<string, V>, key: string, val: V) {
	if (cache.size >= MAX_CACHE) cache.delete(cache.keys().next().value!);
	cache.set(key, val);
}

function cacheKey(lat: number, lng: number, precision = 4) {
	return `${lat.toFixed(precision)},${lng.toFixed(precision)}`;
}

// ─── Языковая коррекция ───────────────────────────────────────────────────────
//
// Nominatim иногда возвращает белорусские варианты названий даже при
// accept-language: ru. Исправляем специфичные символы и латинские омоглифы.

const LANG_FIXES: Array<[RegExp, string | ((s: string) => string)]> = [
	// Белорусские буквы → русские эквиваленты
	[/ў/g, 'у'],
	[/Ў/g, 'У'],
	[/і/g, 'и'],
	[/І/g, 'И'],

	// Латинские омоглифы → кириллица
	// (символы которые визуально идентичны но имеют другой Unicode code point)
	[/a/g, 'а'],
	[/A/g, 'А'],
	[/e/g, 'е'],
	[/E/g, 'Е'],
	[/o/g, 'о'],
	[/O/g, 'О'],
	[/p/g, 'р'],
	[/P/g, 'Р'],
	[/c/g, 'с'],
	[/C/g, 'С'],
	[/x/g, 'х'],
	[/X/g, 'Х'],
	[/B/g, 'В'],
	[/M/g, 'М'],
	[/T/g, 'Т'],
	[/y/g, 'у'],
	[/Y/g, 'У'],
	[/H/g, 'Н'],
	[/K/g, 'К']
];

function fixLanguage(s: string): string {
	if (!s) return s;

	// Применяем только если строка содержит кириллицу —
	// иначе сломаем английские названия
	const hasCyrillic = /[а-яёА-ЯЁ]/.test(s);
	if (!hasCyrillic) return s;

	let result = s;
	for (const [pattern, replacement] of LANG_FIXES) {
		result = result.replace(
			pattern,
			replacement as string // ts не видит перегрузку replace с функцией
		);
	}
	return result;
}

// ─── Нормализация улицы ───────────────────────────────────────────────────────

const STREET_TYPES = [
	{ pattern: /^проспект\s+/i, full: 'проспект ' },
	{ pattern: /^пр-кт\.?\s+/i, full: 'проспект ' },
	{ pattern: /^пр\.?\s*-?\s*кт\.?\s+/i, full: 'проспект ' },
	{ pattern: /^переулок\s+/i, full: 'переулок ' },
	{ pattern: /^пер\.?\s+/i, full: 'переулок ' },
	{ pattern: /^проезд\s+/i, full: 'проезд ' },
	{ pattern: /^пр-д\.?\s+/i, full: 'проезд ' },
	{ pattern: /^шоссе\s+/i, full: 'шоссе ' },
	{ pattern: /^ш\.?\s+/i, full: 'шоссе ' },
	{ pattern: /^бульвар\s+/i, full: 'бульвар ' },
	{ pattern: /^б-р\.?\s+/i, full: 'бульвар ' },
	{ pattern: /^набережная\s+/i, full: 'набережная ' },
	{ pattern: /^наб\.?\s+/i, full: 'набережная ' },
	{ pattern: /^площадь\s+/i, full: 'площадь ' },
	{ pattern: /^пл\.?\s+/i, full: 'площадь ' },
	{ pattern: /^тупик\s+/i, full: 'тупик ' },
	{ pattern: /^аллея\s+/i, full: 'аллея ' },
	{ pattern: /^микрорайон\s+/i, full: 'микрорайон ' },
	{ pattern: /^мкр\.?\s+/i, full: 'микрорайон ' },
	{ pattern: /^улица\s+/i, full: 'улица ' },
	{ pattern: /^ул\.?\s+/i, full: 'улица ' }
] as const;

export function normalizeStreet(raw: string): string {
	if (!raw) return '';
	// Сначала правим язык, потом нормализуем тип улицы
	const fixed = fixLanguage(raw.trim());
	for (const { pattern, full } of STREET_TYPES) {
		if (pattern.test(fixed)) {
			return (full + fixed.replace(pattern, '')).replace(/\s{2,}/g, ' ').trim();
		}
	}
	return fixed;
}

function stripPostal(s: string): string {
	return s
		.replace(/\b\d{6}\b/g, '')
		.replace(/\b\d{5}\b/g, '')
		.replace(/,\s*,/g, ',')
		.replace(/^[\s,]+|[\s,]+$/g, '')
		.replace(/\s{2,}/g, ' ')
		.trim();
}

// ─── Препроцессинг поискового запроса ────────────────────────────────────────

const CITY_HINTS = new Set([
	'москва',
	'санкт-петербург',
	'спб',
	'питер',
	'новосибирск',
	'екатеринбург',
	'казань',
	'нижний',
	'самара',
	'омск',
	'челябинск',
	'ростов',
	'уфа',
	'красноярск',
	'пермь',
	'воронеж',
	'волгоград',
	'краснодар',
	'саратов',
	'тюмень',
	'тольятти',
	'ижевск',
	'барнаул',
	'ульяновск',
	'иркутск',
	'хабаровск',
	'ярославль',
	'владивосток',
	'томск',
	'оренбург',
	'кемерово',
	'новокузнецк',
	'рязань',
	'астрахань',
	'пенза',
	'липецк',
	'тула',
	'киров',
	'чебоксары',
	'калининград',
	'минск',
	'алматы',
	'ташкент',
	'баку',
	'ереван',
	'тбилиси'
]);

export function preprocessQuery(raw: string): string {
	const q = raw.trim();
	if (!q || q.includes(',')) return q;

	const words = q.split(/\s+/);
	const houseIdx = words.findIndex((w) => /^\d+[а-яёa-z]?$/i.test(w));
	if (houseIdx === -1) return words.join(' ');

	const houseNum = words[houseIdx];
	const beforeHouse = words.slice(0, houseIdx);
	const afterHouse = words.slice(houseIdx + 1);
	const isFirstCity = CITY_HINTS.has(beforeHouse[0]?.toLowerCase() ?? '');

	const cityPart = isFirstCity ? beforeHouse[0] : '';
	const streetPart = isFirstCity ? beforeHouse.slice(1).join(' ') : beforeHouse.join(' ');

	return [streetPart, houseNum, cityPart, ...afterHouse].filter(Boolean).join(', ');
}

// ─── Парсинг ответа Nominatim ─────────────────────────────────────────────────

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function extractAddress(raw: any): ParsedAddress | null {
	if (!raw) return null;

	const a = raw.address ?? raw ?? {};

	const rawStreet =
		a.road ?? a.street ?? a.pedestrian ?? a.footway ?? a.path ?? a.cycleway ?? a.residential ?? '';

	const street = normalizeStreet(rawStreet);
	const house = fixLanguage(a.house_number ?? '');
	const country = fixLanguage(a.country ?? '');
	const city = fixLanguage(
		a.city ?? a.town ?? a.village ?? a.hamlet ?? a.municipality ?? a.county ?? ''
	);
	const district = fixLanguage(a.city_district ?? a.suburb ?? a.neighbourhood ?? a.quarter ?? '');

	if (!street && !city) return null;

	const parts = [city, street, house].filter(Boolean);
	const unique = parts.filter(
		(p, i, arr) => p.trim() && arr.findIndex((x) => x.toLowerCase() === p.toLowerCase()) === i
	);

	const confidence: ParsedAddress['confidence'] = house
		? 'full'
		: street
			? 'no-house'
			: 'street-only';

	return {
		country,
		city,
		district,
		street,
		house,
		displayName: stripPostal(unique.join(', ')),
		lat: parseFloat(raw.lat ?? raw.geometry?.coordinates?.[1] ?? 0),
		lng: parseFloat(raw.lon ?? raw.geometry?.coordinates?.[0] ?? 0),
		confidence,
		raw
	};
}

// ─── Fetch хелпер ────────────────────────────────────────────────────────────
//
// Accept-Language передаём цепочкой: ru → ru-RU → en
// Nominatim выбирает name:ru если есть, иначе name:en, иначе name

const HEADERS = {
	'Accept-Language': 'ru,ru-RU;q=0.9,en;q=0.5',
	'User-Agent': 'RentalListingApp/1.0',
	Accept: 'application/json',
	Referer: typeof window !== 'undefined' ? window.location.origin : ''
};

async function nominatimFetch(url: string, signal?: AbortSignal): Promise<unknown> {
	if (typeof navigator !== 'undefined' && !navigator.onLine) {
		throw new Error('OFFLINE');
	}
	const res = await fetch(url, { headers: HEADERS, signal });
	if (!res.ok) throw new Error(`HTTP ${res.status}`);
	return res.json();
}

function hasErrorField(value: unknown): value is { error: unknown } {
	return typeof value === 'object' && value !== null && 'error' in value;
}

// ─── Стратегия 1: стандартный reverse ────────────────────────────────────────

async function reverseStandard(
	lat: number,
	lng: number,
	zoom: number,
	signal?: AbortSignal
): Promise<ParsedAddress | null> {
	const params = new URLSearchParams({
		lat: String(lat),
		lon: String(lng),
		format: 'jsonv2',
		'accept-language': 'ru,ru-RU;q=0.9,en;q=0.5',
		addressdetails: '1',
		zoom: String(zoom)
	});

	const data = await nominatimQueue.enqueue(() =>
		nominatimFetch(`https://nominatim.openstreetmap.org/reverse?${params}`, signal)
	);

	if (hasErrorField(data) && data.error) return null;
	return extractAddress(data);
}

// ─── Стратегия 2: Overpass API — ближайшее здание с номером ──────────────────
//
// Бесплатный, без ключа. Ищем addr:housenumber в радиусе radiusM метров.
// Используем зеркало overpass-api.de (основное) с таймаутом 5с.

async function findNearestBuilding(
	lat: number,
	lng: number,
	radiusM = 80,
	signal?: AbortSignal
): Promise<{ house: string; street: string } | null> {
	const query = `
[out:json][timeout:5];
(
  node["addr:housenumber"](around:${radiusM},${lat},${lng});
  way["addr:housenumber"](around:${radiusM},${lat},${lng});
  relation["addr:housenumber"](around:${radiusM},${lat},${lng});
);
out body center 5;
    `.trim();

	try {
		const res = await fetch('https://overpass-api.de/api/interpreter', {
			method: 'POST',
			body: query,
			headers: { 'Content-Type': 'text/plain', Accept: 'application/json' },
			signal
		});

		if (!res.ok) throw new Error(`Overpass HTTP ${res.status}`);
		const data = await res.json();

		if (!data.elements?.length) return null;

		// Находим ближайший элемент к исходной точке
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const nearest = (data.elements as any[])
			.map((el) => {
				const elLat = el.lat ?? el.center?.lat ?? 0;
				const elLng = el.lon ?? el.center?.lon ?? 0;
				return { el, dist: Math.hypot(elLat - lat, elLng - lng) };
			})
			.sort((a, b) => a.dist - b.dist)[0]?.el;

		if (!nearest?.tags) return null;

		const house = fixLanguage(nearest.tags['addr:housenumber'] ?? '');
		const street = normalizeStreet(nearest.tags['addr:street'] ?? nearest.tags['addr:place'] ?? '');

		return house ? { house, street } : null;
	} catch (e) {
		console.warn('[overpass]', e);
		return null;
	}
}

// ─── Стратегия 3: поиск по сетке смещений ────────────────────────────────────
//
// Координата может попасть во двор или на границу участков.
// Пробуем 8 точек вокруг исходной на расстоянии ~15м.

const OFFSET_DEG = 0.00015; // ~15 метров

const GRID_OFFSETS = [
	[1, 0],
	[-1, 0],
	[0, 1],
	[0, -1],
	[1, 1],
	[1, -1],
	[-1, 1],
	[-1, -1]
] as const;

async function reverseWithGridSearch(
	lat: number,
	lng: number,
	signal?: AbortSignal
): Promise<ParsedAddress | null> {
	for (const [dLat, dLng] of GRID_OFFSETS) {
		const shiftedLat = lat + dLat * OFFSET_DEG;
		const shiftedLng = lng + dLng * OFFSET_DEG;

		const result = await reverseStandard(shiftedLat, shiftedLng, 18, signal);

		if (result?.house) {
			// Нашли — возвращаем с оригинальными координатами пользователя
			return { ...result, lat, lng, confidence: 'full' };
		}

		// Пауза между запросами — очередь уже throttle, это для надёжности
		await new Promise((r) => setTimeout(r, 50));
	}

	return null;
}

// ─── Главный reverse geocoding с fallback-цепочкой ───────────────────────────
//
// Шаг 1 → Nominatim zoom=18          (быстро, точно)
// Шаг 2 → Overpass радиус 80м        (ищем addr:housenumber рядом)
// Шаг 3 → Nominatim zoom=17          (чуть шире охват)
// Шаг 4 → Сетка 8 точек по ~15м      (обход дворов и границ)
// Шаг 5 → Возврат улицы без дома     (с confidence='no-house')

export async function reverseGeocode(
	lat: number,
	lng: number,
	signal?: AbortSignal
): Promise<ParsedAddress | null> {
	const key = cacheKey(lat, lng);

	// 1. In-memory кэш
	if (reverseCache.has(key)) return reverseCache.get(key)!;

	// 2. Персистентный кэш (localStorage, TTL 7 дней)
	const persisted = persistedGeoCache.get(key);
	if (persisted) {
		safeSet(reverseCache, key, persisted.data);
		return persisted.data;
	}

	try {
		// ── Шаг 1 ────────────────────────────────────────────────────────────
		const step1 = await reverseStandard(lat, lng, 18, signal);

		if (step1?.house) {
			const result = { ...step1, confidence: 'full' as const };
			safeSet(reverseCache, key, result);
			persistGeoEntry(key, result);
			return result;
		}

		const baseResult = step1; // улица без дома — запасной вариант

		// ── Шаг 2: Overpass ──────────────────────────────────────────────────
		const building = await findNearestBuilding(lat, lng, 80, signal);

		if (building?.house && baseResult) {
			const merged: ParsedAddress = {
				...baseResult,
				house: building.house,
				street: building.street || baseResult.street,
				confidence: 'full'
			};

			const parts = [merged.city, merged.street, merged.house].filter(Boolean);
			const unique = parts.filter(
				(p, i, arr) => p.trim() && arr.findIndex((x) => x.toLowerCase() === p.toLowerCase()) === i
			);
			merged.displayName = stripPostal(unique.join(', '));

			safeSet(reverseCache, key, merged);
			persistGeoEntry(key, merged);
			return merged;
		}

		// ── Шаг 3: zoom=17 ───────────────────────────────────────────────────
		const step3 = await reverseStandard(lat, lng, 17, signal);

		if (step3?.house) {
			const result = { ...step3, lat, lng, confidence: 'full' as const };
			safeSet(reverseCache, key, result);
			persistGeoEntry(key, result);
			return result;
		}

		// ── Шаг 4: сетка смещений ────────────────────────────────────────────
		const step4 = await reverseWithGridSearch(lat, lng, signal);

		if (step4?.house) {
			safeSet(reverseCache, key, step4);
			persistGeoEntry(key, step4);
			return step4;
		}

		// ── Шаг 5: возвращаем что есть (улица без дома) ──────────────────────
		const fallback = baseResult ?? step3;
		if (fallback) {
			const result = { ...fallback, lat, lng };
			safeSet(reverseCache, key, result);
			// Неполные результаты не сохраняем в localStorage
			return result;
		}

		return null;
	} catch (e: unknown) {
		if (e instanceof DOMException && e.name === 'AbortError') return null;
		const msg = e instanceof Error ? e.message : String(e);
		if (msg === 'OFFLINE') throw new Error('OFFLINE');
		console.error('[reverseGeocode]', e);
		return null;
	}
}

// ─── Search ───────────────────────────────────────────────────────────────────

interface SearchOptions {
	nearLat?: number;
	nearLng?: number;
	signal?: AbortSignal;
}

export async function searchAddress(
	query: string,
	opts: SearchOptions = {}
): Promise<ParsedAddress[]> {
	const processed = preprocessQuery(query);
	const cKey = `${processed}|${(opts.nearLat ?? 0).toFixed(2)}|${(opts.nearLng ?? 0).toFixed(2)}`;

	if (searchCache.has(cKey)) return searchCache.get(cKey)!;

	try {
		const params = new URLSearchParams({
			q: processed,
			format: 'jsonv2',
			'accept-language': 'ru,ru-RU;q=0.9,en;q=0.5',
			addressdetails: '1',
			limit: '8',
			dedupe: '1'
		});

		if (opts.nearLat !== undefined && opts.nearLng !== undefined) {
			const d = 0.5;
			params.set(
				'viewbox',
				[opts.nearLng - d, opts.nearLat - d, opts.nearLng + d, opts.nearLat + d].join(',')
			);
			params.set('bounded', '0');
		}

		const data = (await nominatimQueue.enqueue(() =>
			nominatimFetch(`https://nominatim.openstreetmap.org/search?${params}`, opts.signal)
		)) as unknown[];

		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const results = (data as any[])
			.map(extractAddress)
			.filter((a): a is ParsedAddress => a !== null)
			.filter((a, i, arr) => arr.findIndex((b) => b.displayName === a.displayName) === i);

		// Сортируем по расстоянию от пользователя
		if (opts.nearLat !== undefined && opts.nearLng !== undefined) {
			const uLat = opts.nearLat;
			const uLng = opts.nearLng;
			results.sort(
				(a, b) => Math.hypot(a.lat - uLat, a.lng - uLng) - Math.hypot(b.lat - uLat, b.lng - uLng)
			);
		}

		safeSet(searchCache, cKey, results);
		return results;
	} catch (e: unknown) {
		if (e instanceof DOMException && e.name === 'AbortError') return [];
		console.error('[searchAddress]', e);
		return [];
	}
}

// ─── Утилиты расстояния (Хаверсин) ───────────────────────────────────────────

export function haversineKm(lat1: number, lng1: number, lat2: number, lng2: number): number {
	const R = 6371;
	const dLat = ((lat2 - lat1) * Math.PI) / 180;
	const dLng = ((lng2 - lng1) * Math.PI) / 180;
	const a =
		Math.sin(dLat / 2) ** 2 +
		Math.cos((lat1 * Math.PI) / 180) * Math.cos((lat2 * Math.PI) / 180) * Math.sin(dLng / 2) ** 2;
	return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
}

export function formatDistance(km: number): string {
	if (km < 0.05) return 'рядом';
	if (km < 1) return `${Math.round(km * 1000)} м`;
	if (km < 10) return `${km.toFixed(1)} км`;
	return `${Math.round(km)} км`;
}
