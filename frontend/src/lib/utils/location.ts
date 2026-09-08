import type { ListingLocation } from '$lib/components/card/types';

type CoordinatesLike = {
	lat?: unknown;
	lng?: unknown;
};

function toTrimmed(value: unknown): string {
	return typeof value === 'string' ? value.trim() : '';
}

function splitAddress(address: string): string[] {
	return normalizeAddress(address)
		.split(',')
		.map((part) => part.trim())
		.filter(Boolean);
}

function stripHouseNumber(street: string): string {
	return street.replace(/\s+\d+[A-Za-zА-Яа-я/-]*\s*$/u, '').trim();
}

export function normalizeAddress(address: string): string {
	return address
		.trim()
		.replace(/\s+/g, ' ')
		.replace(/\s*,\s*/g, ', ')
		.replace(/,+$/g, '');
}

export function hasValidCoordinates(
	location?: CoordinatesLike | null
): location is CoordinatesLike & { lat: number; lng: number } {
	return Boolean(
		location &&
		typeof location.lat === 'number' &&
		Number.isFinite(location.lat) &&
		typeof location.lng === 'number' &&
		Number.isFinite(location.lng)
	);
}

export function formatDistrictCity(
	location?: Pick<ListingLocation, 'district' | 'city'> | null
): string {
	const district = toTrimmed(location?.district);
	const city = toTrimmed(location?.city);

	if (district && city) return `${district}, ${city}`;
	return district || city;
}

export function formatListingAddressFull(
	address: string,
	location?: Pick<ListingLocation, 'district' | 'city'> | null
): string {
	const normalizedAddress = normalizeAddress(address);
	if (normalizedAddress) return normalizedAddress;
	const districtCity = formatDistrictCity(location);
	return districtCity || 'Адрес не указан';
}

export function formatListingAddress(address: string, location?: ListingLocation | null): string {
	return formatListingAddressFull(address, location);
}

export function formatListingAddressCompact(
	address: string,
	location?: Pick<ListingLocation, 'district' | 'city'> | null
): string {
	const districtCity = formatDistrictCity(location);
	if (districtCity) return districtCity;

	const parts = splitAddress(address);
	if (parts.length >= 2) {
		const [first, second] = parts;
		return `${first}, ${second}`;
	}

	if (parts.length === 1) return parts[0];
	return 'Локация не указана';
}

export function formatCoordinates(
	location?: Pick<ListingLocation, 'lat' | 'lng'> | null,
	precision = 6
): string | null {
	if (!hasValidCoordinates(location)) return null;
	return `${location.lat.toFixed(precision)}, ${location.lng.toFixed(precision)}`;
}

export function formatCityStreet(
	address: string,
	location?: Pick<ListingLocation, 'city'> | null
): string {
	const parts = splitAddress(address);
	const city = toTrimmed(location?.city) || parts[0] || 'Город не указан';
	const streetSource = parts[1] || '';
	const street = stripHouseNumber(streetSource);

	if (street) return `${city}, ${street}`;
	return city;
}
