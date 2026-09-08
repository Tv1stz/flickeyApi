<script lang="ts">
	import type { ListingPublic } from '$lib/types/listings';
	import ListingGallery from './ListingGallery.svelte';
	import ListingHeader from './ListingHeader.svelte';
	import ListingSpecs from './ListingSpecs.svelte';
	import ListingDescription from './ListingDescription.svelte';
	import ListingAmenities from './ListingAmenities.svelte';
	import ListingRules from './ListingRules.svelte';
	import ListingMap from '$lib/components/listing-page/ListingMap.svelte';
	import ListingHostCard from './ListingHostCard.svelte';
	import ListingPriceCard from './ListingPriceCard.svelte';

	interface Props {
		listing: ListingPublic;
		mode?: 'public' | 'preview';
		onpublish?: () => void;
		onedit?: () => void;
		oncontact?: () => void;
	}

	let {
		listing,
		mode = 'public',
		onpublish,
		onedit,
		oncontact
	}: Props = $props();

	let joinedYear = $derived(
		listing.created_at ? new Date(listing.created_at).getFullYear() : undefined
	);
</script>

<div class="space-y-6 max-w-6xl mx-auto">
	<!-- 1. Header: Title, Address & Actions (Purely Dynamic) -->
	<ListingHeader
		name={listing.name}
		address={listing.address}
		city={listing.city}
		street={listing.street}
		type={listing.type}
		createdAt={listing.created_at}
		id={listing.id}
	/>

	<!-- 2. Bento Photo Gallery (Purely Dynamic) -->
	<ListingGallery photos={listing.media || []} name={listing.name} />

	<!-- 3. Main Two-Column Layout matching screenshot style -->
	<div class="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-10 pt-2">
		<!-- Left Column: Specs, Description, Amenities, Rules, Map (8 cols) -->
		<div class="lg:col-span-8 space-y-6">
			<!-- Key Specs Bar -->
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

			<!-- Rules -->
			<ListingRules
				checkinFrom={listing.checkin_from}
				checkoutUntil={listing.checkout_until}
				minNights={listing.min_nights}
				rules={listing.rules}
				allowChildren={listing.allow_children}
				allowPets={listing.allow_pets}
				allowSmoking={listing.allow_smoking}
				allowParties={listing.allow_parties}
				depositRequired={listing.deposit_required}
				withInvoicing={listing.with_invoicing}
			/>

			<!-- Location Map -->
			<ListingMap
				latitude={listing.latitude}
				longitude={listing.longitude}
				address={listing.address}
				city={listing.city}
				street={listing.street}
			/>
		</div>

		<!-- Right Column: Sticky Price & Host Cards (4 cols) -->
		<div class="lg:col-span-4 space-y-4">
			<div class="sticky top-20 space-y-4">
				<ListingPriceCard
					pricePerNight={listing.price_per_night}
					currency={listing.currency}
					minNights={listing.min_nights}
					{mode}
					{onpublish}
					{onedit}
					{oncontact}
				/>

				{#if listing.host}
					<ListingHostCard
						host={listing.host}
						isVerified={true}
						{joinedYear}
					/>
				{/if}
			</div>
		</div>
	</div>
</div>
