<script lang="ts">
	import type { ListingPublic } from '$lib/types/listings';
	import ListingGallery from './ListingGallery.svelte';
	import ListingHeader from './ListingHeader.svelte';
	import ListingSpecs from './ListingSpecs.svelte';
	import ListingDescription from './ListingDescription.svelte';
	import ListingAmenities from './ListingAmenities.svelte';
	import ListingRules from './ListingRules.svelte';
	import ListingHostCard from './ListingHostCard.svelte';
	import ListingPriceCard from './ListingPriceCard.svelte';

	interface Props {
		listing: ListingPublic;
		mode?: 'public' | 'preview';
		onpublish?: () => void;
		onedit?: () => void;
	}

	let {
		listing,
		mode = 'public',
		onpublish,
		onedit
	}: Props = $props();
</script>

<div class="space-y-8">
	<!-- Title & Badges Header -->
	<ListingHeader
		name={listing.name}
		type={listing.type}
		createdAt={listing.created_at}
		id={listing.id}
	/>

	<!-- Photos Gallery Grid & Lightbox -->
	<ListingGallery photos={listing.media || []} name={listing.name} />

	<!-- Main Details & Sticky Pricing Grid -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-10">
		<!-- Left Column: Specs, Description, Amenities, Rules, Host -->
		<div class="lg:col-span-2 space-y-8">
			<!-- Key Metrics Grid -->
			<ListingSpecs
				square={listing.square}
				maxGuests={listing.max_guests}
				roomsCount={listing.rooms_count}
				bedsCount={listing.beds_count}
				bathroomsCount={listing.bathrooms_count}
				floor={listing.floor}
				totalFloors={listing.total_floors}
			/>

			<!-- Description -->
			<ListingDescription description={listing.description} />

			<!-- Amenities -->
			<ListingAmenities amenities={listing.amenities || []} />

			<!-- House Rules -->
			<ListingRules
				checkinFrom={listing.checkin_from}
				checkoutUntil={listing.checkout_until}
				minNights={listing.min_nights}
			/>

			<!-- Host Info Card -->
			<div class="pt-4 border-t border-border">
				<h3 class="text-lg font-bold text-foreground mb-3">Информация об арендодателе</h3>
				<ListingHostCard host={listing.host} isVerified={true} />
			</div>
		</div>

		<!-- Right Column: Sticky Price / Booking Action Card -->
		<div class="lg:col-span-1">
			<ListingPriceCard
				pricePerNight={listing.price_per_night}
				currency={listing.currency}
				minNights={listing.min_nights}
				{mode}
				{onpublish}
				{onedit}
			/>
		</div>
	</div>
</div>
