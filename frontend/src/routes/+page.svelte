<script lang="ts">
	import type { PageData } from './$types';
	import type { HousingType, ListingPublic } from '$lib/types/listings';
	import { listingsApi } from '$lib/api/listings';
	import ListingCard from '$lib/components/listings/ListingCard.svelte';
	import ListingFilterBar from '$lib/components/listings/ListingFilterBar.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Building2, Sparkles, PlusCircle } from 'lucide-svelte';

	let { data }: { data: PageData } = $props();

	let selectedType = $state<HousingType | 'all'>('all');
	let searchQuery = $state<string>('');
	let clientListings = $state<ListingPublic[]>([]);
	let isFiltering = $state<boolean>(false);

	$effect(() => {
		clientListings = data.listings || [];
	});

	// Client-side instant filter by query
	let filteredListings = $derived(
		clientListings.filter((item) => {
			if (!searchQuery.trim()) return true;
			const q = searchQuery.toLowerCase();
			return (
				item.name.toLowerCase().includes(q) ||
				(item.description && item.description.toLowerCase().includes(q))
			);
		})
	);

	async function handleFilterChange(type: HousingType | 'all', query: string) {
		selectedType = type;
		searchQuery = query;

		isFiltering = true;
		try {
			const res = await listingsApi.getPublicListings(type === 'all' ? undefined : type);
			clientListings = res || [];
		} catch {
			clientListings = [];
		} finally {
			isFiltering = false;
		}
	}
</script>

<svelte:head>
	<title>Flickey — Посуточная аренда квартир, домов и усадеб в Беларуси</title>
	<meta
		name="description"
		content="Сервис краткосрочной аренды проверенного жилья в Беларуси. Квартиры, коттеджи и агроусадьбы без посредников."
	/>
</svelte:head>

<div class="container mx-auto px-4 sm:px-6 py-8 space-y-10">
	<!-- Hero Section -->
	<section class="text-center max-w-3xl mx-auto space-y-4 pt-4 pb-2">
		<div class="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-3.5 py-1 text-xs font-bold text-primary">
			<Sparkles class="h-3.5 w-3.5" />
			<span>Проверенное жилье в Беларуси</span>
		</div>
		<h1 class="text-3xl sm:text-5xl font-extrabold tracking-tight text-foreground leading-[1.15]">
			Найдите идеальное место для отдыха или работы
		</h1>
		<p class="text-sm sm:text-base text-muted-foreground max-w-xl mx-auto leading-relaxed">
			Уютные квартиры, загородные дома и усадьбы с проверенными фотографиями и мгновенным бронированием.
		</p>
	</section>

	<!-- Filters & Search Bar -->
	<section class="max-w-4xl mx-auto">
		<ListingFilterBar
			bind:selectedType
			bind:searchQuery
			onfilter={handleFilterChange}
		/>
	</section>

	<!-- Listings Grid -->
	<section>
		{#if isFiltering}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
				{#each Array(8) as _}
					<div class="flex flex-col rounded-2xl border border-border bg-card p-4 space-y-3">
						<Skeleton class="aspect-[4/3] w-full rounded-xl" />
						<Skeleton class="h-5 w-3/4" />
						<Skeleton class="h-4 w-1/2" />
						<div class="pt-4 flex justify-between">
							<Skeleton class="h-6 w-1/3" />
							<Skeleton class="h-4 w-1/4" />
						</div>
					</div>
				{/each}
			</div>
		{:else if filteredListings.length > 0}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
				{#each filteredListings as listing (listing.id)}
					<ListingCard {listing} />
				{/each}
			</div>
		{:else}
			<!-- Empty State -->
			<div class="rounded-3xl border border-dashed border-border bg-muted/20 p-12 text-center max-w-lg mx-auto space-y-4">
				<div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
					<Building2 class="h-7 w-7" />
				</div>
				<h3 class="text-lg font-bold text-foreground">Объявлений пока нет</h3>
				<p class="text-xs text-muted-foreground leading-relaxed">
					{#if searchQuery}
						По запросу «{searchQuery}» ничего не найдено. Попробуйте изменить параметры поиска.
					{:else}
						В этой категории пока нет опубликованных объявлений. Вы можете стать первым арендодателем!
					{/if}
				</p>
				<div class="pt-2">
					<a href="/host/new">
						<Button size="sm" class="gap-2 font-semibold">
							<PlusCircle class="h-4 w-4" /> Сдать свое жилье
						</Button>
					</a>
				</div>
			</div>
		{/if}
	</section>
</div>
