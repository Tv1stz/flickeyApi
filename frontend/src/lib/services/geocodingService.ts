import type { ListingLocation } from '$lib/components/card/types';
import { hasValidCoordinates, normalizeAddress } from '$lib/utils/location';
import { geoApi } from '$lib/api/geo';

type NominatimAddress = {
	city?: string;
	town?: string;
	village?: string;
	hamlet?: string;
	municipality?: string;
	county?: string;
	state_district?: string;
	city_district?: string;
	district?: string;
	suburb?: string;
	borough?: string;
	quarter?: string;
	neighbourhood?: string;
};

type NominatimResult = {
	lat: string;
	lon: string;
	address?: NominatimAddress;
};

const SEARCH_ENDPOINT = 'https://nominatim.openstreetmap.org/search';
const REQUEST_TIMEOUT_MS = 6000;
const cache = new Map<string, ListingLocation | null>();

function toFinite(value: string | number): number | null {
	const numeric = Number(value);
	return Number.isFinite(numeric) ? numeric : null;
}

function pickCity(address?: NominatimAddress): string {
	return (
		address?.city ||
		address?.town ||
		address?.village ||
		address?.hamlet ||
		address?.municipality ||
		address?.county ||
		'Не указан'
	);
}

function pickDistrict(address?: NominatimAddress): string | undefined {
	return (
		address?.city_district ||
		address?.district ||
		address?.suburb ||
		address?.borough ||
		address?.quarter ||
		address?.neighbourhood ||
		address?.state_district
	);
}

export async function geocodeAddress(address: string): Promise<ListingLocation | null> {
	const normalizedAddress = normalizeAddress(address);
	if (!normalizedAddress) return null;

	const cacheKey = normalizedAddress.toLowerCase();

	if (cache.has(cacheKey)) {
		return cache.get(cacheKey) ?? null;
	}

	// 1. Try backend geoApi proxy
	try {
		const suggestions = await geoApi.suggest(normalizedAddress, 'ru', 1);
		if (suggestions && suggestions.length > 0) {
			const first = suggestions[0];
			const location: ListingLocation = {
				lat: first.latitude,
				lng: first.longitude,
				city: first.city || 'Беларусь',
				address: first.raw_address || normalizedAddress
			};
			cache.set(cacheKey, location);
			return location;
		}
	} catch {
		// Fallback to nominatim
	}

	const controller = new AbortController();
	const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

	try {
		const query = new URLSearchParams({
			format: 'jsonv2',
			addressdetails: '1',
			limit: '1',
			'accept-language': 'ru',
			q: normalizedAddress
		});

		const response = await fetch(`${SEARCH_ENDPOINT}?${query.toString()}`, {
			method: 'GET',
			headers: { Accept: 'application/json' },
			signal: controller.signal
		});

		if (!response.ok) {
			cache.set(cacheKey, null);
			return null;
		}

		const payload = (await response.json()) as NominatimResult[];
		const first = Array.isArray(payload) ? payload[0] : null;
		if (!first) {
			cache.set(cacheKey, null);
			return null;
		}

		const lat = toFinite(first.lat);
		const lng = toFinite(first.lon);
		if (lat === null || lng === null) {
			cache.set(cacheKey, null);
			return null;
		}

		const location: ListingLocation = {
			lat,
			lng,
			city: pickCity(first.address),
			district: pickDistrict(first.address),
			address: normalizedAddress
		};

		cache.set(cacheKey, location);
		return location;
	} catch {
		cache.set(cacheKey, null);
		return null;
	} finally {
		clearTimeout(timeout);
	}
}

export async function resolveListingLocation(
	address: string,
	currentLocation?: ListingLocation
): Promise<ListingLocation | undefined> {
	const normalizedAddress = normalizeAddress(address);
	if (!normalizedAddress) return undefined;

	const geocoded = await geocodeAddress(normalizedAddress);
	if (geocoded) return geocoded;

	if (hasValidCoordinates(currentLocation)) {
		const existingCity = currentLocation.city?.trim() || 'Не указан';
		const existingDistrict = currentLocation.district?.trim() || undefined;

		return {
			...currentLocation,
			address: normalizedAddress,
			city: existingCity,
			district: existingDistrict
		};
	}

	return undefined;
}
