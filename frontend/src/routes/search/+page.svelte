<script lang="ts">
	import { onMount } from 'svelte';
	import ListingGrid from '$lib/components/card/ListingGrid.svelte';
	import SearchBarMobile from '$lib/components/search/panels/SearchBarMobile.svelte';
	import SearchPanelMobile from '$lib/components/search/panels/SearchPanelMobile.svelte';
	import { searchStore } from '$lib/stores/searchStore.svelte';
	import {
		countActiveFilters,
		filterListings,
		getCityOptionsFromListings,
		getPopularCitiesFromListings,
		hasActiveSearch,
		isSearchComplete,
		PROPERTY_TYPE_OPTIONS
	} from '$lib/components/search/taxonomy';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import { fade } from 'svelte/transition';
	import { List, Maximize2, Minimize2, Map as MapIcon, RotateCcw } from 'lucide-svelte';
	import type { Listing, PropertyType } from '$lib/components/card/types';
	import { pluralRu } from '$lib/utils/format';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import FiltersModal from '$lib/components/search/FiltersModal.svelte';
	import SearchMap from '$lib/components/search/SearchMap.svelte';
	import { listingsApi } from '$lib/api/listings';
	import { apiListingToCardListing } from '$lib/utils/listingConverters';

	interface PageData {
		listings: Promise<Listing[]>;
	}

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	let favorites = $derived(favoritesStore.favoriteIds);
	let listings = $state<Listing[]>([]);
	let isLoaded = $state(false);
	let mobileSearchOpen = $state(false);
	let mapExpanded = $state(false);
	let mobileFiltersOpen = $state(false);

	const params = $derived(searchStore.params);
	const filters = $derived(searchStore.filters);
	const filteredListings = $derived(isLoaded ? filterListings(listings, params, filters) : []);
	const hasSearch = $derived(isSearchComplete(params));
	const activeFiltersCount = $derived(countActiveFilters(filters));
	const totalGuests = $derived(params.adults + params.children);
	const propertyLabel = $derived(
		PROPERTY_TYPE_OPTIONS.find((o) => o.id === params.propertyType)?.label ?? 'Тип жилья'
	);
	const titleLine = $derived(
		`${params.location || 'Новый поиск'}${params.propertyType !== 'any' ? ` · ${propertyLabel}` : ''}`
	);
	const subtitleLine = $derived(
		totalGuests > 0
			? `${totalGuests} ${pluralRu(totalGuests, ['гость', 'гостя', 'гостей'])}`
			: 'Укажите гостей'
	);

	function handleBack() {
		if (typeof history !== 'undefined' && history.length > 1) {
			history.back();
			return;
		}
		goto(resolve('/'));
	}

	$effect(() => {
		if (typeof document === 'undefined') return;
		if (window.matchMedia('(min-width: 1024px)').matches) return;
		if (!mapExpanded) {
			const saved = document.body.style.top;
			document.body.style.position = '';
			document.body.style.top = '';
			document.body.style.left = '';
			document.body.style.right = '';
			document.body.style.width = '';
			if (saved) {
				const offset = Math.abs(parseInt(saved, 10) || 0);
				window.scrollTo(0, offset);
			}
			return;
		}

		const scrollY = window.scrollY;
		document.body.style.position = 'fixed';
		document.body.style.top = `-${scrollY}px`;
		document.body.style.left = '0';
		document.body.style.right = '0';
		document.body.style.width = '100%';
	});



	onMount(() => {
		const url = $page.url;
		const loc = url.searchParams.get('location') || '';
		const pt = (url.searchParams.get('propertyType') as PropertyType | 'any' | null) || 'any';
		const ci = url.searchParams.get('checkin') || null;
		const co = url.searchParams.get('checkout') || null;
		const ad = parseInt(url.searchParams.get('adults') || '0', 10) || 0;
		const ch = parseInt(url.searchParams.get('children') || '0', 10) || 0;

		if (loc || pt !== 'any' || ci || co || ad > 0 || ch > 0) {
			searchStore.setDraft({ location: loc, propertyType: pt, checkin: ci, checkout: co, adults: ad, children: ch });
			searchStore.commit();
		}
	});

	$effect(() => {
		data.listings.then((list) => {
			listings = list;
			searchStore.setListings(list);
			searchStore.setAvailableCities(getCityOptionsFromListings(list));
			searchStore.setPopularCities(getPopularCitiesFromListings(list));
			isLoaded = true;
		});
	});

	let lastLoadedDates = $state<string>('');
	let hasInitializedDates = false;

	$effect(() => {
		const ci = params.checkin;
		const co = params.checkout;
		const key = `${ci ?? ''}_${co ?? ''}`;

		if (!hasInitializedDates) {
			hasInitializedDates = true;
			lastLoadedDates = key;
			return;
		}

		if (key !== lastLoadedDates) {
			lastLoadedDates = key;
			listingsApi.getPublicListings(undefined, ci || undefined, co || undefined).then((raw) => {
				const list = (raw || []).map(apiListingToCardListing);
				listings = list;
				searchStore.setListings(list);
			});
		}
	});

	const hasAnyActiveFilters = $derived(hasActiveSearch(params) || activeFiltersCount > 0);

	function handleResetAll() {
		searchStore.resetAll();
		listingsApi.getPublicListings().then((raw) => {
			const list = (raw || []).map(apiListingToCardListing);
			listings = list;
			searchStore.setListings(list);
		});
	}

	$effect(() => {
		searchStore.setResultsCount(filteredListings.length);
		return () => searchStore.setResultsCount(0);
	});
</script>

<svelte:head>
	<title>Поиск жилья — Flickey</title>
	<meta name="description" content="Подбор жилья по вашим критериям." />
</svelte:head>

<div class="min-h-screen bg-white">
	<SearchPanelMobile
		variant="search"
		title={titleLine}
		subtitle={subtitleLine}
		showFilters={hasSearch}
		filtersCount={activeFiltersCount}
		onBack={handleBack}
		onOpenSearch={() => (mobileSearchOpen = true)}
		onOpenFilters={() => (mobileFiltersOpen = true)}
		class="lg:hidden"
	/>

	<section class="pt-0 lg:pt-0">
		<div class="mx-auto max-w-[1600px] px-4 sm:px-6 lg:px-8"></div>
	</section>

	<section class="pb-32 lg:pb-20">
		<div class="mx-auto max-w-[1600px] px-4 sm:px-6 lg:px-8">
			<div class="mt-6 flex flex-wrap items-center justify-between gap-3">
				<h2 class="text-xl font-semibold text-zinc-900">
					Нашли для вас
					{filteredListings.length}
					{pluralRu(filteredListings.length, ['вариант', 'варианта', 'вариантов'])}
					жилья
				</h2>

				{#if hasAnyActiveFilters}
					<button
						type="button"
						onclick={handleResetAll}
						class="flex items-center gap-1.5 rounded-full border border-zinc-200 bg-white px-4 py-2 text-xs font-semibold text-zinc-700 shadow-sm transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95"
					>
						<RotateCcw size={13} strokeWidth={2} />
						<span>Сбросить все фильтры</span>
					</button>
				{/if}
			</div>

			<div class="mt-6 grid gap-8 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
				<div class={mapExpanded ? 'lg:hidden' : ''}>
					<div in:fade={{ duration: 300 }}>
						{#if isLoaded}
							{#if filteredListings.length === 0}
								<div class="flex flex-col items-center justify-center py-20 text-center">
									<div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400">
										<RotateCcw size={26} strokeWidth={1.75} />
									</div>
									<h3 class="mb-1 text-lg font-semibold text-zinc-900">Ничего не найдено</h3>
									<p class="mb-5 max-w-sm text-sm text-zinc-500">
										Попробуйте изменить параметры поиска или сбросить фильтры
									</p>
									{#if hasAnyActiveFilters}
										<button
											type="button"
											onclick={handleResetAll}
											class="rounded-full bg-zinc-900 px-5 py-2.5 text-xs font-semibold text-white shadow transition hover:bg-zinc-800 active:scale-95"
										>
											Сбросить все фильтры
										</button>
									{/if}
								</div>
							{:else}
								<ListingGrid
									listings={filteredListings}
									{favorites}
									layout="two-column"
								/>
							{/if}
						{:else}
							<ListingGrid listings={[]} isLoading layout="two-column" />
						{/if}
					</div>
				</div>

				<div class={mapExpanded ? 'lg:col-span-2' : ''}>
					<div
						class="relative hidden h-[420px] overflow-hidden rounded-3xl border border-zinc-200 bg-zinc-100 shadow-sm lg:sticky lg:top-[112px] lg:block lg:h-[calc(100vh-112px-24px)]"
					>
						<SearchMap listings={filteredListings} />

						<button
							type="button"
							onclick={() => (mapExpanded = !mapExpanded)}
							class="absolute top-4 right-4 z-10 flex items-center gap-2 rounded-full bg-white/90 px-3 py-2 text-xs font-semibold text-zinc-800 shadow-md backdrop-blur transition-all hover:bg-white active:scale-[0.98]"
							aria-pressed={mapExpanded}
						>
							{#if mapExpanded}
								<Minimize2 size={14} strokeWidth={2} />
								<span>Свернуть карту</span>
							{:else}
								<Maximize2 size={14} strokeWidth={2} />
								<span>На всю ширину</span>
							{/if}
						</button>
					</div>
				</div>
			</div>

			{#if mapExpanded}
				<div class="fixed inset-0 z-50 bg-white lg:hidden">
					<div class="relative h-full w-full overflow-hidden bg-zinc-100">
						<SearchMap listings={filteredListings} />

						<button
							type="button"
							onclick={() => (mapExpanded = false)}
							class="absolute top-4 right-4 z-10 flex items-center gap-2 rounded-full bg-white/90 px-3 py-2 text-xs font-semibold text-zinc-800 shadow-md backdrop-blur transition-all hover:bg-white active:scale-[0.98]"
						>
							<List size={14} strokeWidth={2} />
							<span>К списку</span>
						</button>
					</div>
				</div>
			{/if}

			<button
				type="button"
				onclick={() => (mapExpanded = !mapExpanded)}
				class="fixed bottom-[calc(24px+var(--mobile-nav-height,0px))] left-1/2 z-40 flex -translate-x-1/2 items-center gap-2 rounded-full bg-zinc-900 px-5 py-3 text-sm font-semibold text-white shadow-lg transition-all hover:bg-zinc-800 active:scale-[0.98] lg:hidden"
				aria-pressed={mapExpanded}
			>
				{#if mapExpanded}
					<List size={16} strokeWidth={2} />
					<span>Списком</span>
				{:else}
					<MapIcon size={16} strokeWidth={2} />
					<span>На карте</span>
				{/if}
			</button>
		</div>
	</section>
</div>

<SearchBarMobile open={mobileSearchOpen} {listings} onClose={() => (mobileSearchOpen = false)} />

<FiltersModal
	open={mobileFiltersOpen}
	filters={searchStore.filters}
	resultsCount={filteredListings.length}
	{listings}
	{params}
	canApply={true}
	onApply={(f) => searchStore.commitFilters(f)}
	onClose={() => (mobileFiltersOpen = false)}
/>
