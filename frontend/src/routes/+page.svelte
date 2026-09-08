<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';
	import ListingGrid from '$lib/components/card/ListingGrid.svelte';
	import SearchBarMobile from '$lib/components/search/panels/SearchBarMobile.svelte';
	import SearchPanelMobile from '$lib/components/search/panels/SearchPanelMobile.svelte';
	import { searchStore } from '$lib/stores/searchStore.svelte';
	import {
		getCityOptionsFromListings,
		getPopularCitiesFromListings,
		buildSearchUrl
	} from '$lib/components/search/taxonomy';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import { fade } from 'svelte/transition';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import type { Listing } from '$lib/components/card/types';
	import { apiListingToCardListing } from '$lib/utils/listingConverters';

	let { data }: { data: PageData } = $props();

	let favorites = $derived(favoritesStore.favoriteIds);
	let allListings = $derived((data.listings || []).map(apiListingToCardListing));
	let mobileSearchOpen = $state(false);

	onMount(() => {
		searchStore.resetAll();
		if (allListings.length > 0) {
			searchStore.setListings(allListings);
			searchStore.setAvailableCities(getCityOptionsFromListings(allListings));
			searchStore.setPopularCities(getPopularCitiesFromListings(allListings));
		}
	});
</script>

<svelte:head>
	<title>Flickey — Аренда жилья посуточно</title>
	<meta name="description" content="Квартиры, дома и апартаменты для посуточной аренды в Беларуси." />
</svelte:head>

<div class="min-h-screen bg-white">
	<SearchPanelMobile
		variant="home"
		title="Найти жильё"
		subtitle="Город · Тип · Гости"
		onOpenSearch={() => (mobileSearchOpen = true)}
		class="lg:hidden"
	/>

	<section class="pt-8 pb-8 lg:pt-32 lg:pb-10">
		<div class="mx-auto max-w-[1600px] px-4 sm:px-6 lg:px-8">
			<div class="mx-auto max-w-3xl text-center">
				<h1
					class="text-4xl leading-[1.05] font-bold tracking-tight text-slate-900 sm:text-5xl md:text-6xl lg:text-7xl"
				>
					Найдите место,<br />
					<span
						class="bg-gradient-to-r from-zinc-900 via-zinc-700 to-zinc-900 bg-clip-text text-transparent"
					>
						где хочется остаться
					</span>
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-lg leading-relaxed text-zinc-500 sm:text-xl">
					Квартиры, дома и апартаменты для незабываемого отдыха
				</p>
			</div>
		</div>
	</section>

	<section class="pb-32 lg:pb-20">
		<div class="mx-auto max-w-[1600px] px-4 sm:px-6 lg:px-8">
			<div in:fade={{ duration: 300 }}>
				{#if allListings.length > 0}
					<ListingGrid
						listings={allListings}
						{favorites}
						emptyMessage="Ничего не найдено"
						emptyDescription="Попробуйте изменить параметры поиска"
					/>
				{:else}
					<div class="rounded-3xl border border-dashed border-zinc-200 p-12 text-center max-w-lg mx-auto space-y-4">
						<h3 class="text-lg font-bold text-zinc-900">Объявлений пока нет</h3>
						<p class="text-sm text-zinc-500 leading-relaxed">
							В этой категории пока нет опубликованных объявлений. Вы можете стать первым арендодателем!
						</p>
						<div class="pt-2">
							<a
								href="/host/new"
								class="inline-flex items-center justify-center rounded-2xl bg-zinc-900 px-5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-zinc-800 transition-colors"
							>
								Сдать жильё
							</a>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</section>
</div>

<SearchBarMobile
	open={mobileSearchOpen}
	listings={allListings}
	onClose={() => (mobileSearchOpen = false)}
	onApply={() => goto(buildSearchUrl(searchStore.params))}
/>

