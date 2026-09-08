<script lang="ts">
	import Modal from '$lib/components/ui/Modal.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import PriceRangeSlider from './PriceRangeSlider.svelte';
	import { fade, slide } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import {
		DEFAULT_FILTERS,
		filterListings,
		getCategoriesWithAmenities,
		getPopularAmenities,
		getPriceRange,
		filterAmenitiesByPropertyType,
		type SearchFilters,
		type SearchParams
	} from './taxonomy';
	import type { Listing } from '$lib/components/card/types';
	import { pluralRu } from '$lib/utils/format';

	interface Props {
		open: boolean;
		filters: SearchFilters;
		resultsCount: number;
		listings?: Listing[];
		params?: SearchParams;
		canApply?: boolean;
		onApply: (filters: SearchFilters) => void;
		onClose: () => void;
	}

	let {
		open,
		filters,
		resultsCount,
		listings,
		params,
		canApply = true,
		onApply,
		onClose
	}: Props = $props();

	let local = $state<SearchFilters>({ ...DEFAULT_FILTERS });
	let selectedCategory = $state<string>('popular');
	let showAllAmenities = $state(false);

	const previewCount = $derived(
		listings && params ? filterListings(listings, params, local).length : resultsCount
	);

	const chartListings = $derived.by(() => {
		if (!listings || !params) return listings ?? [];
		const tempFilters = { ...local, priceMin: null, priceMax: null };
		return filterListings(listings, params, tempFilters);
	});

	const categoriesWithAmenities = getCategoriesWithAmenities();
	const popularAmenities = getPopularAmenities(6);

	const filteredPopularAmenities = $derived.by(() => {
		if (!params || params.propertyType === 'any') return popularAmenities;
		return filterAmenitiesByPropertyType(popularAmenities, params.propertyType);
	});

	const filteredCategoriesWithAmenities = $derived.by(() => {
		if (!params || params.propertyType === 'any') return categoriesWithAmenities;
		return categoriesWithAmenities
			.map(({ category, amenities }) => ({
				category,
				amenities: filterAmenitiesByPropertyType(amenities, params.propertyType)
			}))
			.filter(({ amenities }) => amenities.length > 0);
	});

	const priceRange = $derived(
		chartListings.length > 0 ? getPriceRange(chartListings) : { min: 0, max: 1000 }
	);

	$effect(() => {
		if (open) {
			local = { ...filters };
			showAllAmenities = false;
			selectedCategory = 'popular';
		}
	});

	function onPriceMinChange(value: number | null) {
		local = { ...local, priceMin: value };
	}

	function onPriceMaxChange(value: number | null) {
		local = { ...local, priceMax: value };
	}

	function toggleAmenity(id: string) {
		const has = local.amenities.includes(id);
		local = {
			...local,
			amenities: has ? local.amenities.filter((a) => a !== id) : [...local.amenities, id]
		};
	}

	function setBeds(n: number | null) {
		local = { ...local, bedsMin: local.bedsMin === n ? null : n };
	}

	function setBathrooms(n: number | null) {
		local = { ...local, bathroomsMin: local.bathroomsMin === n ? null : n };
	}

	function reset() {
		local = { ...DEFAULT_FILTERS };
	}

	function apply() {
		onApply({ ...local });
		onClose();
	}
</script>

<Modal {open} {onClose} title="Фильтры" maxWidth="lg">
	<div class="modal-content-stack px-1">
		<!-- Диапазон цен с гистограммой -->
		<section class="modal-section border-b border-zinc-100 pb-8 mb-8">
			<h4 class="mb-1 text-lg font-semibold tracking-tight text-zinc-900">
				Диапазон цен
			</h4>
			<p class="mb-8 text-sm text-zinc-500">
				Цена за ночь без учета сборов и налогов
			</p>

			<PriceRangeSlider
				min={priceRange.min}
				max={priceRange.max}
				valueMin={local.priceMin}
				valueMax={local.priceMax}
				listings={chartListings}
				onChangeMin={onPriceMinChange}
				onChangeMax={onPriceMaxChange}
			/>
		</section>

		<!-- Количество кроватей -->
		<section class="modal-section border-b border-zinc-100 pb-8 mb-8">
			<h4 class="mb-5 text-lg font-semibold tracking-tight text-zinc-900">
				Количество кроватей
			</h4>
			<div class="flex flex-wrap gap-3">
				<button
					type="button"
					onclick={() => setBeds(null)}
					class="min-w-[72px] touch-manipulation rounded-full border px-5 py-2.5 text-sm font-medium transition-all duration-200 active:scale-95
                 {local.bedsMin === null
						? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
						: 'border-zinc-200 bg-white text-zinc-700 hover:border-zinc-900 hover:text-zinc-900'}"
				>
					Любое
				</button>
				{#each [1, 2, 3, 4, 5, 6] as n (n)}
					<button
						type="button"
						onclick={() => setBeds(n)}
						class="min-w-[64px] touch-manipulation rounded-full border px-5 py-2.5 text-sm font-medium transition-all duration-200 active:scale-95
                   {local.bedsMin === n
							? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
							: 'border-zinc-200 bg-white text-zinc-700 hover:border-zinc-900 hover:text-zinc-900'}"
					>
						{n === 6 ? '6+' : n}
					</button>
				{/each}
			</div>
		</section>

		<!-- Количество ванных комнат -->
		<section class="modal-section border-b border-zinc-100 pb-8 mb-8">
			<h4 class="mb-5 text-lg font-semibold tracking-tight text-zinc-900">
				Количество ванных комнат
			</h4>
			<div class="flex flex-wrap gap-3">
				<button
					type="button"
					onclick={() => setBathrooms(null)}
					class="min-w-[72px] touch-manipulation rounded-full border px-5 py-2.5 text-sm font-medium transition-all duration-200 active:scale-95
                 {local.bathroomsMin === null
						? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
						: 'border-zinc-200 bg-white text-zinc-700 hover:border-zinc-900 hover:text-zinc-900'}"
				>
					Любое
				</button>
				{#each [1, 2, 3, 4] as n (n)}
					<button
						type="button"
						onclick={() => setBathrooms(n)}
						class="min-w-[64px] touch-manipulation rounded-full border px-5 py-2.5 text-sm font-medium transition-all duration-200 active:scale-95
                   {local.bathroomsMin === n
							? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
							: 'border-zinc-200 bg-white text-zinc-700 hover:border-zinc-900 hover:text-zinc-900'}"
					>
						{n === 4 ? '4+' : n}
					</button>
				{/each}
			</div>
		</section>

		<!-- Удобства -->
		<section class="modal-section pb-4">
			<div class="mb-6 flex items-center justify-between">
				<h4 class="text-lg font-semibold tracking-tight text-zinc-900">Удобства</h4>
				{#if !showAllAmenities}
					<button
						type="button"
						onclick={() => (showAllAmenities = true)}
						class="text-sm font-medium text-zinc-900 underline decoration-zinc-300 underline-offset-4 transition-all duration-200 hover:decoration-zinc-900"
					>
						Показать все
					</button>
				{/if}
			</div>

			{#if !showAllAmenities}
				<div class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-3" in:fade={{ duration: 200, delay: 100 }}>
					{#each filteredPopularAmenities as amenity (amenity.id)}
						{@const selected = local.amenities.includes(amenity.id)}
						<button
							type="button"
							onclick={() => toggleAmenity(amenity.id)}
							class="group flex touch-manipulation flex-col items-start gap-3 rounded-2xl p-4 text-left transition-all duration-200 active:scale-[0.97]
        {selected
								? 'bg-zinc-50 ring-2 ring-inset ring-zinc-900'
								: 'bg-white ring-1 ring-inset ring-zinc-200 hover:ring-zinc-900'}"
						>
							<span class="text-2xl leading-none transition-transform duration-200">
								{amenity.emoji}
							</span>
							<span class="text-sm font-medium tracking-tight text-zinc-900">
								{amenity.label}
							</span>
						</button>
					{/each}
				</div>
			{:else}
				<div class="space-y-8" in:slide={{ duration: 400, easing: cubicOut }}>
					<div class="relative -mx-1">
						<div class="flex gap-2.5 overflow-x-auto px-1 pb-3 scrollbar-hide">
							{#each filteredCategoriesWithAmenities as { category } (category.id)}
								<button
									type="button"
									onclick={() => (selectedCategory = category.id)}
									class="flex shrink-0 touch-manipulation items-center gap-2 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-200 active:scale-95
                       {selectedCategory === category.id
										? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
										: 'border-zinc-200 bg-white text-zinc-700 hover:border-zinc-900 hover:text-zinc-900'}"
								>
									<span class="text-base">{category.emoji}</span>
									<span>{category.label}</span>
								</button>
							{/each}
						</div>
					</div>

					{#each filteredCategoriesWithAmenities as { category, amenities } (category.id)}
						{#if selectedCategory === category.id}
							<div class="space-y-4" in:fade={{ duration: 200 }}>
								<div>
									<h5 class="text-base font-semibold text-zinc-900">
										{category.label}
									</h5>
									{#if category.description}
										<p class="mt-1 text-sm text-zinc-500">{category.description}</p>
									{/if}
								</div>
								<div class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-3">
									{#each amenities as amenity (amenity.id)}
										{@const selected = local.amenities.includes(amenity.id)}
										<button
											type="button"
											onclick={() => toggleAmenity(amenity.id)}
											class="group flex touch-manipulation flex-col items-start gap-3 rounded-2xl p-4 text-left transition-all duration-200 active:scale-[0.97]
                               {selected
												? 'bg-zinc-50 ring-2 ring-inset ring-zinc-900'
												: 'border border-zinc-200 bg-white hover:border-zinc-900'}"
										>
											<span class="text-2xl leading-none transition-transform duration-200 group-hover:scale-110">
												{amenity.emoji}
											</span>
											<span class="text-sm font-medium tracking-tight text-zinc-900">
												{amenity.label}
											</span>
										</button>
									{/each}
								</div>
							</div>
						{/if}
					{/each}

					<div class="pt-4">
						<button
							type="button"
							onclick={() => (showAllAmenities = false)}
							class="w-full rounded-xl py-3 text-center text-sm font-semibold text-zinc-900 transition-colors hover:bg-zinc-100"
						>
							Свернуть список
						</button>
					</div>
				</div>
			{/if}
		</section>
	</div>

	{#snippet footer()}
		<div class="modal-footer flex flex-col gap-3 border-t border-zinc-100 bg-white px-6 py-4 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
			<button
				type="button"
				onclick={reset}
				class="touch-manipulation text-sm font-medium text-zinc-900 underline decoration-zinc-300 underline-offset-4 transition-all duration-200 hover:decoration-zinc-900 active:scale-95"
			>
				Очистить всё
			</button>
			<Button
				variant="solid"
				tone="primary"
				size="lg"
				radius="xl"
				class="min-w-[160px]"
				disabled={!canApply}
				onclick={apply}
			>
				<span class="font-semibold text-base">
					Показать {previewCount}
					{pluralRu(previewCount, ['объявление', 'объявления', 'объявлений'])}
				</span>
			</Button>
		</div>
	{/snippet}
</Modal>

<style>
	.scrollbar-hide {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
	.scrollbar-hide::-webkit-scrollbar {
		display: none;
	}
</style>
