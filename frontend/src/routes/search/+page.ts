import type { PageLoad } from './$types';
import { listingsApi } from '$lib/api/listings';
import { apiListingToCardListing } from '$lib/utils/listingConverters';
import type { Listing } from '$lib/components/card/types';

export const load: PageLoad = async ({ url }) => {
	try {
		const checkin = url.searchParams.get('checkin');
		const checkout = url.searchParams.get('checkout');
		const rawListings = await listingsApi.getPublicListings(undefined, checkin, checkout);
		const listings: Listing[] = (rawListings || []).map(apiListingToCardListing);
		return {
			listings: Promise.resolve(listings)
		};
	} catch {
		return {
			listings: Promise.resolve([] as Listing[])
		};
	}
};

