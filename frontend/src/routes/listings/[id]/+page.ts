import type { PageLoad } from './$types';
import { listingsApi } from '$lib/api/listings';
import { apiListingToCardListing } from '$lib/utils/listingConverters';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
	try {
		const raw = await listingsApi.getPublicListing(params.id);
		if (!raw) {
			throw error(404, 'Объявление не найдено');
		}
		const listing = apiListingToCardListing(raw);
		return {
			listing
		};
	} catch (err) {
		throw error(404, err instanceof Error ? err.message : 'Объявление не найдено');
	}
};

