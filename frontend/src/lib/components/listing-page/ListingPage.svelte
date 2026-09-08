<!-- src/lib/components/listing-page/ListingPage.svelte -->
<script lang="ts">
	import ListingHeader from './ListingHeader.svelte';
	import ListingParams from './ListingParams.svelte';
	import ListingDescription from './ListingDescription.svelte';
	import ListingAmenities from './ListingAmenities.svelte';
	import ListingOwnerCard from './ListingOwnerCard.svelte';
	import ListingMap from './ListingMap.svelte';
	import ListingGalleryGrid from './ListingGalleryGrid.svelte';
	import ListingLightbox from './ListingLightbox.svelte';
	import ListingRules from './ListingRules.svelte';
	import ListingAvailabilityCalendar from './ListingAvailabilityCalendar.svelte';
	import PriceBar from './PriceBar.svelte';
	import ContactsModal from './ContactsModal.svelte';
	import type { Listing } from '$lib/components/card/types';
	import { PROPERTY_TYPE_LABELS } from '$lib/components/card/types';
	import { pluralize, PLURAL_FORMS } from '$lib/utils/pluralize';
	import { formatBYN, pluralRu } from '$lib/utils/format';
	import { formatListingAddressFull, normalizeAddress } from '$lib/utils/location';
	import { AlertCircle, Pencil, Eye } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { saveDraft } from '$lib/utils/formStorage';
	import { toast } from '$lib/stores/toastStore';
	import type { ListingFormValues } from '$lib/validation/listingValidation';

	interface Props {
		listing: Listing;
		mode?: 'default' | 'preview';
	}

	let { listing, mode = 'default' }: Props = $props();

	const showActions = $derived(mode === 'default');

	let galleryOpen = $state(false);
	let startIndex = $state(0);
	let scrollY = $state(0);
	let contactsOpen = $state(false);
	let selectedCheckin = $state<string | null>(null);
	let selectedCheckout = $state<string | null>(null);
	let selectedNights = $state(0);

	const headerOpaque = $derived(scrollY > 280);

	function openGallery(index = 0) {
		startIndex = index;
		galleryOpen = true;
	}

	function closeGallery() {
		galleryOpen = false;
	}

	function openContacts() {
		contactsOpen = true;
	}

	function closeContacts() {
		contactsOpen = false;
	}

	const propertyLabel = $derived(PROPERTY_TYPE_LABELS[listing.propertyType] ?? '');
	const normalizedAddress = $derived(normalizeAddress(listing.address));
	const fullLocation = $derived(formatListingAddressFull(normalizedAddress, listing.location));
	const headerAddressText = $derived(fullLocation);

	const subtitleText = $derived.by(() => {
		if (propertyLabel && fullLocation) {
			return `${propertyLabel} · ${fullLocation}`;
		}
		if (propertyLabel) {
			return `${propertyLabel} · ${normalizedAddress}`;
		}
		return fullLocation || normalizedAddress;
	});
	const isOwner = $derived(
		Boolean(
			authStore.user?.id &&
				(authStore.user.id === listing.ownerId || authStore.user.id === listing.owner?.id)
		)
	);
	const isAdmin = $derived(Boolean(authStore.user?.roles?.includes('admin')));
	const isHostOrAdmin = $derived(isOwner || isAdmin);
	const isUnpublished = $derived(
		Boolean(listing.status && listing.status !== 'published' && listing.status !== 'active')
	);

	function handleEditAndResubmit() {
		goto(`/host/listings/${listing.id}`);
	}
</script>

<svelte:window bind:scrollY />

<article class="min-h-screen bg-white">
	<!-- ════════════════════════════════════════════════════ -->
	<!-- MOBILE LAYOUT                                       -->
	<!-- ════════════════════════════════════════════════════ -->
	<div class="relative lg:hidden">
		<!-- Floating header -->
		<div
			class="safe-top fixed top-0 right-0 left-0 z-[100] transition-all duration-500 ease-out"
			data-listing-mobile-header
			style="background-color: {headerOpaque
				? 'rgba(255,255,255,0.97)'
				: 'transparent'}; backdrop-filter: {headerOpaque ? 'blur(20px)' : 'none'};
                {headerOpaque ? 'box-shadow: 0 1px 3px rgba(0,0,0,0.05);' : ''}"
		>
			{#if !headerOpaque}
				<div
					class="pointer-events-none absolute inset-0 bg-gradient-to-b from-black/40 via-black/15 to-transparent"
				></div>
			{/if}
			<ListingHeader title={listing.title} transparent={!headerOpaque} {listing} {showActions} />
		</div>

		<!-- Gallery -->
		<div class="relative">
			<ListingGalleryGrid
				images={listing.images ?? []}
				onOpen={openGallery}
				fullHeight={false}
				mobileFullWidth={true}
			/>
		</div>

		<!-- Content sheet -->
		<div class="relative z-20 -mt-6 rounded-t-3xl bg-white shadow-xl">
			<div class="px-5 pt-7 pb-40">
				{#if isHostOrAdmin && isUnpublished}
					{#if listing.status === 'rejected'}
						<div class="mb-6 rounded-2xl border border-red-200 bg-red-50 p-4">
							<div class="flex items-start gap-3">
								<AlertCircle class="h-5 w-5 text-red-600 shrink-0 mt-0.5" />
								<div class="min-w-0 flex-1">
									<p class="text-sm font-semibold text-red-900">Объявление отклонено</p>
									<p class="text-xs text-red-700 mt-1">Отредактируйте данные и отправьте на повторную модерацию.</p>
									<button
										type="button"
										onclick={handleEditAndResubmit}
										class="mt-3 inline-flex items-center gap-2 rounded-xl bg-red-600 px-3.5 py-2 text-xs font-semibold text-white shadow-sm hover:bg-red-700 transition-colors cursor-pointer"
									>
										<Pencil class="h-3.5 w-3.5" />
										Редактировать и отправить заново
									</button>
								</div>
							</div>
						</div>
					{:else}
						<div class="mb-6 rounded-2xl border border-amber-200 bg-amber-50/80 p-4">
							<div class="flex items-start gap-3">
								<Eye class="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
								<div class="min-w-0 flex-1 text-xs">
									<p class="font-bold text-amber-950">
										Режим предпросмотра
										<span class="ml-1.5 inline-block rounded-md bg-amber-200/70 px-2 py-0.5 text-[11px] font-semibold text-amber-900">
											{listing.status === 'draft_video_required' ? 'Черновик: требуется видео' : listing.status === 'draft' ? 'Черновик' : listing.status === 'pending_review' ? 'На модерации' : listing.status === 'awaiting_company_verification' ? 'Ожидает верификации партнера' : 'Не опубликовано'}
										</span>
									</p>
									<p class="text-amber-800 mt-1 leading-relaxed">
										Объявление пока не видно гостям в поиске. Вы просматриваете его так, как оно будет отображаться после публикации.
									</p>
								</div>
							</div>
						</div>
					{/if}
				{/if}

				<!-- Title block -->
				<header>
					<h1 class="text-3xl leading-[1.15] font-bold tracking-tight text-zinc-900">
						{listing.title}
					</h1>
					<p class="mt-2 text-lg leading-snug text-zinc-900">{subtitleText}</p>
				</header>

				<!-- Divider -->
				<div class="my-6 h-px bg-zinc-100"></div>

				<!-- Params -->
				<section>
					<ListingParams
						maxGuests={listing.maxGuests}
						bedrooms={listing.bedrooms}
						beds={listing.beds}
						bathrooms={listing.bathrooms}
						area={listing.area}
						floor={listing.floor}
						totalFloors={listing.totalFloors}
						propertyType={listing.propertyType}
					/>
				</section>

				<!-- Sections -->
				<div class="mt-8 space-y-0">
					{#if listing.description}
						<section class=" pt-8 pb-8">
							<ListingDescription description={listing.description} />
						</section>
					{/if}
					{#if listing.amenities && listing.amenities.length > 0}
						<section class=" pt-8 pb-8">
							<ListingAmenities amenities={listing.amenities} propertyType={listing.propertyType} />
						</section>
					{/if}
					{#if listing.rules}
						<section class=" pt-8 pb-8">
							<ListingRules rules={listing.rules} />
						</section>
					{/if}
					<section class=" pt-8 pb-8">
						<ListingAvailabilityCalendar
							listingId={listing.id}
							pricePerNight={listing.pricePerNight}
							minNights={listing.rules?.minNights ?? 1}
							onDatesSelected={(ci, co, nights) => {
								selectedCheckin = ci;
								selectedCheckout = co;
								selectedNights = nights;
							}}
						/>
					</section>
					<section class=" pt-8 pb-8">
						<ListingMap location={listing.location} address={normalizedAddress} />
					</section>
					{#if listing.owner}
						<section class=" pt-8 pb-4">
							<ListingOwnerCard owner={listing.owner} />
						</section>
					{/if}
				</div>
			</div>
		</div>
	</div>

	<!-- ════════════════════════════════════════════════════ -->
	<!-- DESKTOP LAYOUT                                      -->
	<!-- ════════════════════════════════════════════════════ -->
	<div class="hidden lg:block">
		<!-- Sticky header -->
		<div class="sticky top-0 z-30 bg-white/95 backdrop-blur-md">
			<div class="mx-auto max-w-[1120px] px-8 xl:px-6">
				<ListingHeader
					title={listing.title}
					address={headerAddressText}
					{listing}
					{showActions}
					showTitle={true}
				/>
			</div>
		</div>

		<div class="mx-auto max-w-[1120px] px-8 xl:px-6">
			{#if isHostOrAdmin && isUnpublished}
				{#if listing.status === 'rejected'}
					<div class="mt-4 rounded-2xl border border-red-200 bg-red-50 p-5 flex items-center justify-between gap-4">
						<div class="flex items-start gap-3">
							<div class="rounded-xl bg-red-100 p-2.5 text-red-600">
								<AlertCircle class="h-6 w-6" />
							</div>
							<div>
								<p class="text-base font-semibold text-red-900">Объявление отклонено модератором</p>
								<p class="text-sm text-red-700 mt-0.5">Вы можете внести исправления в параметры, фото или описание и отправить объявление на повторную проверку.</p>
							</div>
						</div>
						<button
							type="button"
							onclick={handleEditAndResubmit}
							class="inline-flex items-center gap-2 rounded-xl bg-red-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-red-700 transition-colors shrink-0 cursor-pointer"
						>
							<Pencil class="h-4 w-4" />
							Редактировать и отправить заново
						</button>
					</div>
				{:else}
					<div class="mt-4 rounded-2xl border border-amber-200 bg-amber-50/80 p-5 flex items-center justify-between gap-4">
						<div class="flex items-start gap-3.5">
							<div class="rounded-xl bg-amber-100 p-2.5 text-amber-700 shrink-0">
								<Eye class="h-6 w-6" />
							</div>
							<div>
								<p class="text-base font-bold text-amber-950 flex items-center gap-2">
									<span>Режим предпросмотра</span>
									<span class="inline-block rounded-md bg-amber-200/80 px-2.5 py-0.5 text-xs font-semibold text-amber-900">
										{listing.status === 'draft_video_required' ? 'Черновик: требуется видео' : listing.status === 'draft' ? 'Черновик' : listing.status === 'pending_review' ? 'На модерации' : listing.status === 'awaiting_company_verification' ? 'Ожидает верификации партнера' : 'Не опубликовано'}
									</span>
								</p>
								<p class="text-xs sm:text-sm text-amber-800 mt-1 leading-relaxed">
									Это объявление пока не видно гостям в поиске. Вы просматриваете его в режиме хозяина.
								</p>
							</div>
						</div>
						<a
							href={resolve('/host/listings')}
							class="inline-flex items-center rounded-xl border border-amber-300 bg-white px-4 py-2 text-xs font-semibold text-amber-900 hover:bg-amber-100 transition shrink-0 cursor-pointer shadow-2xs"
						>
							В мои объявления
						</a>
					</div>
				{/if}
			{/if}

			<!-- Gallery -->
			<div class="pt-4 pb-2">
				<ListingGalleryGrid images={listing.images ?? []} onOpen={openGallery} />
			</div>

			<!-- Two-column layout -->
			<div class="grid grid-cols-[1fr_380px] gap-x-16 pt-8 pb-40 xl:gap-x-20">
				<!-- Left: main content -->
				<div class="min-w-0">
					<!-- Params -->
					<section class="pb-8">
						<ListingParams
							maxGuests={listing.maxGuests}
							bedrooms={listing.bedrooms}
							beds={listing.beds}
							bathrooms={listing.bathrooms}
							area={listing.area}
							floor={listing.floor}
							totalFloors={listing.totalFloors}
							propertyType={listing.propertyType}
						/>
					</section>

					{#if listing.description}
						<section class=" pt-8 pb-8">
							<ListingDescription description={listing.description} />
						</section>
					{/if}

					{#if listing.amenities && listing.amenities.length > 0}
						<section class=" pt-8 pb-8">
							<ListingAmenities amenities={listing.amenities} propertyType={listing.propertyType} />
						</section>
					{/if}

					{#if listing.rules}
						<section class=" pt-8 pb-8">
							<ListingRules rules={listing.rules} />
						</section>
					{/if}

					<section class="pt-8 pb-8">
						<ListingAvailabilityCalendar
							listingId={listing.id}
							pricePerNight={listing.pricePerNight}
							minNights={listing.rules?.minNights ?? 1}
							onDatesSelected={(ci, co, nights) => {
								selectedCheckin = ci;
								selectedCheckout = co;
								selectedNights = nights;
							}}
						/>
					</section>

					<section class="pt-8">
						<ListingMap location={listing.location} address={normalizedAddress} />
					</section>
				</div>

				<!-- Right: sticky sidebar -->
				<aside class="relative">
					<div class="sticky top-24 space-y-6">
						<!-- Price card -->
						<div
							class="overflow-hidden rounded-3xl border border-zinc-200/60 bg-white p-6 shadow-md"
						>
							{#if selectedNights > 0}
								<div class="flex items-baseline gap-1.5">
									<span class="text-[28px] font-bold tracking-tight text-zinc-900">
										{formatBYN(listing.pricePerNight * selectedNights)}
									</span>
									<span class="text-[14px] font-medium text-zinc-500">
										за {selectedNights} {pluralRu(selectedNights, ['ночь', 'ночи', 'ночей'])}
									</span>
								</div>
								<p class="mt-1 text-[13px] text-zinc-400">
									{formatBYN(listing.pricePerNight)} / ночь
								</p>
							{:else}
								<div class="flex items-baseline gap-1.5">
									<span class="text-[28px] font-bold tracking-tight text-zinc-900">
										от {formatBYN(listing.pricePerNight)}
									</span>
									<span class="text-[15px] text-zinc-400">/ ночь</span>
								</div>

								{#if listing.rules?.minNights && listing.rules.minNights > 1}
									<p class="mt-1 text-[13px] text-zinc-400">
										мин. {listing.rules.minNights}
										{pluralize(listing.rules.minNights, PLURAL_FORMS.nights)}
									</p>
								{/if}
							{/if}

							{#if listing.owner?.contacts && mode === 'default'}
								<button
									onclick={openContacts}
									class="mt-5 flex h-12 w-full items-center justify-center rounded-2xl
                                           bg-zinc-900 text-[15px] font-semibold text-white
                                           shadow-md transition-all
                                           duration-150 hover:bg-zinc-800
                                           active:scale-[0.98]"
									style="cursor: pointer;"
								>
									Связаться с хозяином
								</button>
							{/if}
						</div>

						<!-- Owner card (compact for sidebar) -->
						{#if listing.owner}
							<ListingOwnerCard owner={listing.owner} compact={true} />
						{/if}
					</div>
				</aside>
			</div>
		</div>
	</div>

	<!-- Mobile price bar -->
	<div class="lg:hidden">
		<PriceBar
			pricePerNight={listing.pricePerNight}
			minNights={listing.rules?.minNights}
			{selectedNights}
			ownerContacts={listing.owner?.contacts}
			{mode}
			onOpenContacts={openContacts}
		/>
	</div>
</article>

<!-- Contacts modal -->
<ContactsModal
	open={contactsOpen}
	onClose={closeContacts}
	ownerContacts={listing.owner?.contacts}
	ownerName={listing.owner?.name}
/>

{#if galleryOpen}
	<ListingLightbox images={listing.images ?? []} {startIndex} onClose={closeGallery} />
{/if}

<style>
	.safe-top {
		padding-top: env(safe-area-inset-top, 0px);
	}
</style>
