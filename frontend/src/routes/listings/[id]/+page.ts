import type { PageLoad } from './$types';
import { listingsApi } from '$lib/api/listings';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
	try {
		const listing = await listingsApi.getPublicListing(params.id);
		if (!listing) {
			throw error(404, 'Объявление не найдено');
		}
		return {
			listing
		};
	} catch (err) {
		throw error(404, err instanceof Error ? err.message : 'Объявление не найдено');
	}
};
