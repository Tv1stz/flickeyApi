import type { PageLoad } from './$types';
import { listingsApi } from '$lib/api/listings';
import type { ListingPublic } from '$lib/types/listings';

export const load: PageLoad = async () => {
	try {
		const listings = await listingsApi.getPublicListings();
		return {
			listings: listings || []
		};
	} catch {
		return {
			listings: [] as ListingPublic[]
		};
	}
};
