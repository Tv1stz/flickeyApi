import { apiRequest } from './client';
import type { GeoSuggestItem, GeoSuggestResponse } from '$lib/types/geo';

export const geoApi = {
	/**
	 * Fetch address suggestions matching query string.
	 * Calls GET /api/v1/geo/suggest?q={query}&lang={lang}&limit={limit}
	 */
	async suggest(
		query: string,
		lang: string = 'ru',
		limit: number = 5,
		signal?: AbortSignal
	): Promise<GeoSuggestItem[]> {
		const q = query.trim();
		if (q.length < 2) {
			return [];
		}

		const params = new URLSearchParams();
		params.set('q', q);
		if (lang) params.set('lang', lang);
		if (limit) params.set('limit', limit.toString());

		try {
			const res = await apiRequest<GeoSuggestResponse>(`/geo/suggest?${params.toString()}`, {
				method: 'GET',
				signal
			});
			return res?.results || [];
		} catch (err) {
			if (err instanceof DOMException && err.name === 'AbortError') {
				return [];
			}
			console.warn('Geocoding request failed:', err);
			return [];
		}
	}
};
