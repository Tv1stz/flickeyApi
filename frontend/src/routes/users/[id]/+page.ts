import type { PageLoad } from './$types';
import { listingsApi } from '$lib/api/listings';
import { apiListingToCardListing } from '$lib/utils/listingConverters';
import type { Listing } from '$lib/components/card/types';
import type { PublicUser } from '$lib/types/users';

export const load: PageLoad = async ({ params }) => {
	const userId = params.id;

	const [userResult, listingsResult] = await Promise.allSettled([
		listingsApi.getPublicUser(userId),
		listingsApi.getPublicListings(undefined, undefined, undefined, userId)
	]);

	const user: PublicUser | null =
		userResult.status === 'fulfilled' ? userResult.value : null;

	const listings: Listing[] =
		listingsResult.status === 'fulfilled'
			? (listingsResult.value || []).map(apiListingToCardListing)
			: [];

	return {
		user,
		listings,
		userId
	};
};
