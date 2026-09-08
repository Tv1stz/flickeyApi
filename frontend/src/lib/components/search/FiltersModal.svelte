<script lang="ts">
	import Modal from '$lib/components/ui/Modal.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import {
		AMENITY_FILTER_OPTIONS,
		DEFAULT_FILTERS,
		filterListings,
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
	const previewCount = $derived(
		listings && params ? filterListings(listings, params, local).length : resultsCount
	);

	$effect(() => {
		if (open) local = { ...filters };
	});

	function onPriceMinInput(e: Event) {
		const v = parseInt((e.target as HTMLInputElement).value);
		local = { ...local, priceMin: isNaN(v) ? null : v };
	}

	function onPriceMaxInput(e: Event) {
		const v = parseInt((e.target as HTMLInputElement).value);
		local = { ...local, priceMax: isNaN(v) ? null : v };
	}

	function toggleAmenity(id: string) {
		const has = local.amenities.includes(id);
		local = {
			...local,
			amenities: has ? local.amenities.filter((a) => a !== id) : [...local.amenities, id]
		};
	}

	function setBedrooms(n: number | null) {
		local = { ...local, bedroomsMin: local.bedroomsMin === n ? null : n };
	}

	function reset() {
		local = { ...DEFAULT_FILTERS };
	}

	function apply() {
		onApply({ ...local });
		onClose();
	}
</script>

<Modal {open} {onClose} title="Фильтры" maxWidth="md">
	<div class="modal-content-stack">
		<!-- Диапазон цен -->
		<section class="modal-section">
			<h4 class="mb-1.5 text-[17px] font-semibold tracking-[-0.01em] text-zinc-900">
				Диапазон цен
			</h4>
			<p class="mb-6 text-[14.5px] text-zinc-500">Цена за ночь без учета сборов и налогов</p>

			<div class="flex items-center gap-4">
				<!-- От -->
				<div
					class="relative flex-1 rounded-2xl border border-zinc-400/70 p-3 transition-colors duration-200 focus-within:border-zinc-900 focus-within:bg-zinc-50/30 focus-within:ring-1 focus-within:ring-zinc-900"
				>
					<label
						for="price-min"
						class="block text-[11px] font-medium tracking-wide text-zinc-500 uppercase"
						>Минимум</label
					>
					<div class="mt-1 flex items-center gap-1.5">
						<span class="text-[15px] text-zinc-800">BYN</span>
						<input
							id="price-min"
							type="number"
							min="0"
							placeholder="0"
							value={local.priceMin ?? ''}
							oninput={onPriceMinInput}
							class="w-full [appearance:textfield] bg-transparent text-[15px] text-zinc-900 placeholder:text-zinc-300 focus:outline-none [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
						/>
					</div>
				</div>

				<div class="h-[1px] w-4 shrink-0 bg-zinc-300"></div>

				<!-- До -->
				<div
					class="relative flex-1 rounded-2xl border border-zinc-400/70 p-3 transition-colors duration-200 focus-within:border-zinc-900 focus-within:bg-zinc-50/30 focus-within:ring-1 focus-within:ring-zinc-900"
				>
					<label
						for="price-max"
						class="block text-[11px] font-medium tracking-wide text-zinc-500 uppercase"
						>Максимум</label
					>
					<div class="mt-1 flex items-center gap-1.5">
						<span class="text-[15px] text-zinc-800">BYN</span>
						<input
							id="price-max"
							type="number"
							min="0"
							placeholder="∞"
							value={local.priceMax ?? ''}
							oninput={onPriceMaxInput}
							class="w-full [appearance:textfield] bg-transparent text-[15px] text-zinc-900 placeholder:text-zinc-300 focus:outline-none [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
						/>
					</div>
				</div>
			</div>
		</section>

		<!-- Количество спален -->
		<section class="modal-section">
			<h4 class="mb-6 text-[17px] font-semibold tracking-[-0.01em] text-zinc-900">
				Количество спален
			</h4>
			<div class="flex flex-wrap gap-2.5">
				<button
					type="button"
					onclick={() => setBedrooms(null)}
					class="min-w-[72px] touch-manipulation rounded-full border px-5 py-2.5 text-[14.5px] font-medium transition-all duration-200 active:scale-[0.96]
                 {local.bedroomsMin === null
						? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
						: 'border-zinc-300 bg-white text-zinc-800 hover:border-zinc-900'}"
				>
					Любое
				</button>
				{#each [1, 2, 3, 4] as n (n)}
					<button
						type="button"
						onclick={() => setBedrooms(n)}
						class="min-w-[64px] touch-manipulation rounded-full border px-5 py-2.5 text-[14.5px] font-medium transition-all duration-200 active:scale-[0.96]
                   {local.bedroomsMin === n
							? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
							: 'border-zinc-300 bg-white text-zinc-800 hover:border-zinc-900'}"
					>
						{n === 4 ? '4+' : n}
					</button>
				{/each}
			</div>
		</section>

		<!-- Удобства -->
		<section class="modal-section">
			<h4 class="mb-6 text-[17px] font-semibold tracking-[-0.01em] text-zinc-900">Удобства</h4>
			<div class="grid grid-cols-2 gap-3.5">
				{#each AMENITY_FILTER_OPTIONS as amenity (amenity.id)}
					{@const selected = local.amenities.includes(amenity.id)}
					<button
						type="button"
						onclick={() => toggleAmenity(amenity.id)}
						class="group flex touch-manipulation items-start gap-3.5 rounded-2xl border p-4 text-left transition-all duration-200 active:scale-[0.98]
                   {selected
							? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
							: 'border-zinc-300/80 bg-white text-zinc-800 hover:border-zinc-900 hover:bg-zinc-50/50'}"
					>
						<span
							class="text-2xl leading-none opacity-90 transition-transform duration-200 group-hover:scale-110"
							>{amenity.emoji}</span
						>
						<span class="mt-0.5 text-[14.5px] leading-tight font-medium tracking-[-0.01em]"
							>{amenity.label}</span
						>
					</button>
				{/each}
			</div>
		</section>
	</div>

	{#snippet footer()}
		<div class="modal-footer flex items-center justify-between gap-4">
			<button
				type="button"
				onclick={reset}
				class="touch-manipulation text-[15px] font-semibold text-zinc-900 underline decoration-zinc-900/30 underline-offset-4 transition-all duration-200 hover:text-black hover:decoration-zinc-900 active:scale-95"
			>
				Очистить всё
			</button>
			<Button
				variant="solid"
				tone="primary"
				size="lg"
				radius="pill"
				disabled={!canApply}
				onclick={apply}
			>
				<span class="px-2 font-semibold">
					Показать {previewCount}
					{pluralRu(previewCount, ['объявление', 'объявления', 'объявлений'])}
				</span>
			</Button>
		</div>
	{/snippet}
</Modal>
