<script lang="ts">
	import { searchStore } from '$lib/stores/searchStore.svelte';
	import {
		PROPERTY_TYPE_OPTIONS,
		DEFAULT_FILTERS,
		getCityOptionsFromListings,
		isSearchComplete,
		type SearchFilters
	} from '../taxonomy';
	import { isExactNormalizedMatch, matchesSearch } from '$lib/utils/searchNormalization';
	import type { Listing, PropertyType } from '$lib/components/card/types';
	import { Search, X, MapPin, Minus, Plus } from 'lucide-svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	interface Props {
		open: boolean;
		onClose: () => void;
		listings?: Listing[];
		onApply?: () => void;
	}

	let { open, onClose, onApply, listings }: Props = $props();

	let filtersOpen = $state(false);
	let loc = $state('');
	let pt = $state<PropertyType | 'any'>('any');
	let adults = $state(0);
	let children = $state(0);
	let checkin = $state<string | null>(null);
	let checkout = $state<string | null>(null);
	let localFilters = $state<SearchFilters>({ ...DEFAULT_FILTERS });
	const hasSearch = $derived(isSearchComplete(searchStore.params));

	$effect(() => {
		if (open) {
			loc = searchStore.params.location;
			pt = searchStore.params.propertyType;
			adults = searchStore.params.adults;
			children = searchStore.params.children;
			checkin = searchStore.params.checkin;
			checkout = searchStore.params.checkout;
			localFilters = { ...searchStore.filters };
		}
	});

	$effect(() => {
		if (!hasSearch && filtersOpen) filtersOpen = false;
	});

	const totalGuests = $derived(adults + children);

	const availableCities = $derived.by(() => {
		if (listings && listings.length > 0) {
			return getCityOptionsFromListings(listings);
		}
		return searchStore.availableCities;
	});

	const popularCities = $derived.by(() => searchStore.popularCities);
	const query = $derived(loc.trim());
	const cities = $derived(
		query.length > 0
			? availableCities.filter((c) => {
					if (matchesSearch(query, c.label)) return true;
					return c.region ? matchesSearch(query, c.region) : false;
				})
			: availableCities
	);
	const hasExactMatch = $derived(
		query.length > 0 && availableCities.some((c) => isExactNormalizedMatch(c.label, query))
	);
	const showSuggestions = $derived(query.length > 0 && !hasExactMatch);

	function apply() {
		searchStore.setDraft({ location: loc, propertyType: pt, adults, children, checkin, checkout });
		searchStore.setDraftFilters(localFilters);
		searchStore.commit();
		onClose();
		onApply?.();
	}

	function reset() {
		loc = '';
		pt = 'any';
		adults = 0;
		children = 0;
		checkin = null;
		checkout = null;
		localFilters = { ...DEFAULT_FILTERS };
	}

	function setAdults(next: number) {
		if (next < 0) return;
		if (next + children > 20) return;
		if (next === 0 && children > 0) return;
		adults = next;
	}

	function setChildren(next: number) {
		if (next < 0) return;
		if (adults + next > 20) return;
		if (next > 0 && adults === 0) return;
		children = next;
	}
</script>

<Modal {open} {onClose} title="Поиск жилья">
	<div class="modal-content-stack">
		<!-- Location -->
		<section class="modal-section">
			<h4 class="mb-5 text-[17px] font-semibold tracking-[-0.01em] text-zinc-900">Куда едем?</h4>
			<div
				class="group relative flex items-center gap-3 rounded-2xl border border-zinc-300/80 bg-white p-4
                        shadow-sm transition-colors duration-200 focus-within:border-zinc-900 focus-within:ring-1 focus-within:ring-zinc-900"
			>
				<Search size={20} strokeWidth={2} class="shrink-0 text-zinc-900" />
				<input
					type="text"
					placeholder="Город, район или адрес"
					bind:value={loc}
					class="min-w-0 flex-1 bg-transparent text-[15px] font-medium text-zinc-900
                       placeholder:font-normal placeholder:text-zinc-400 focus:outline-none"
				/>
				{#if loc}
					<button
						type="button"
						onclick={() => (loc = '')}
						class="flex h-8 w-8 touch-manipulation items-center justify-center rounded-full bg-zinc-200 text-zinc-600 transition-colors hover:bg-zinc-300"
					>
						<X size={14} strokeWidth={2.5} />
					</button>
				{/if}
			</div>

			<!-- Выпадающий список городов -->
			{#if showSuggestions || query.length === 0}
				<div class="mt-3 max-h-[260px] space-y-1 overflow-y-auto overscroll-contain">
					{#if query.length === 0}
						<p class="px-2 pb-1 text-[12px] font-semibold tracking-wide text-zinc-400 uppercase">
							Популярные города
						</p>
						{#each popularCities.length > 0 ? popularCities : availableCities.slice(0, 6) as city (city.id)}
							<button
								type="button"
								onclick={() => (loc = city.label)}
								class="flex w-full touch-manipulation items-center gap-4 rounded-2xl p-3 text-left
                           transition-all active:scale-[0.98] {loc === city.label
									? 'bg-zinc-100'
									: 'hover:bg-zinc-50'}"
							>
								<div
									class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-zinc-100"
								>
									<MapPin size={18} strokeWidth={2} class="text-zinc-600" />
								</div>
								<div class="min-w-0">
									<p class="text-[15px] font-medium text-zinc-900">{city.label}</p>
									{#if city.region}<p class="text-[13px] text-zinc-500">{city.region}</p>{/if}
								</div>
							</button>
						{/each}
					{:else if cities.length === 0}
						<div
							class="rounded-2xl border border-dashed border-zinc-200 px-4 py-5 text-center text-sm text-zinc-400"
						>
							Нет совпадений
						</div>
					{:else}
						{#each cities.slice(0, 5) as city (city.id)}
							<button
								type="button"
								onclick={() => (loc = city.label)}
								class="flex w-full touch-manipulation items-center gap-4 rounded-2xl p-3 text-left
                           transition-all active:scale-[0.98] {loc === city.label
									? 'bg-zinc-100'
									: 'hover:bg-zinc-50'}"
							>
								<div
									class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-zinc-100"
								>
									<MapPin size={18} strokeWidth={2} class="text-zinc-600" />
								</div>
								<div class="min-w-0">
									<p class="text-[15px] font-medium text-zinc-900">{city.label}</p>
									{#if city.region}<p class="text-[13px] text-zinc-500">{city.region}</p>{/if}
								</div>
							</button>
						{/each}
					{/if}
				</div>
			{/if}
		</section>

		<!-- Dates (Когда) -->
		<section class="modal-section">
			<div class="mb-3 flex items-center justify-between">
				<h4 class="text-[17px] font-semibold tracking-[-0.01em] text-zinc-900">Даты поездки</h4>
				{#if checkin || checkout}
					<button
						type="button"
						onclick={() => {
							checkin = null;
							checkout = null;
						}}
						class="text-xs font-semibold text-zinc-500 underline hover:text-zinc-900"
					>
						Сбросить
					</button>
				{/if}
			</div>
			<div class="grid grid-cols-2 gap-3">
				<div class="rounded-2xl border border-zinc-300/80 bg-white p-3 shadow-sm">
					<label class="block text-xs font-medium text-zinc-500 mb-1" for="mobile-checkin">Заезд</label>
					<input
						id="mobile-checkin"
						type="date"
						value={checkin ?? ''}
						min={new Date().toISOString().split('T')[0]}
						onchange={(e) => {
							const val = (e.target as HTMLInputElement).value;
							checkin = val || null;
							if (checkout && val && checkout <= val) checkout = null;
						}}
						class="w-full bg-transparent text-sm font-semibold text-zinc-900 focus:outline-none"
					/>
				</div>
				<div class="rounded-2xl border border-zinc-300/80 bg-white p-3 shadow-sm">
					<label class="block text-xs font-medium text-zinc-500 mb-1" for="mobile-checkout">Отъезд</label>
					<input
						id="mobile-checkout"
						type="date"
						value={checkout ?? ''}
						min={checkin ?? new Date().toISOString().split('T')[0]}
						onchange={(e) => {
							const val = (e.target as HTMLInputElement).value;
							checkout = val || null;
						}}
						class="w-full bg-transparent text-sm font-semibold text-zinc-900 focus:outline-none"
					/>
				</div>
			</div>
		</section>

		<!-- Property type (В стиле модалки фильтров) -->
		<section class="modal-section">
			<h4 class="mb-5 text-[17px] font-semibold tracking-[-0.01em] text-zinc-900">Тип жилья</h4>
			<div class="grid grid-cols-2 gap-3.5">
				{#each PROPERTY_TYPE_OPTIONS as option (option.id)}
					{@const sel = pt === option.id}
					<button
						type="button"
						onclick={() => (pt = option.id as PropertyType | 'any')}
						class="group flex touch-manipulation items-start gap-3.5 rounded-2xl border p-4 text-left transition-all duration-200 active:scale-[0.98]
                         {sel
							? 'border-zinc-900 bg-zinc-900 text-white shadow-md'
							: 'border-zinc-300/80 bg-white text-zinc-800 hover:border-zinc-900 hover:bg-zinc-50/50'}"
					>
						<span
							class="text-2xl leading-none opacity-90 transition-transform duration-200 group-hover:scale-110"
							>{option.emoji}</span
						>
						<span class="mt-0.5 text-[14.5px] leading-tight font-medium tracking-[-0.01em]"
							>{option.label}</span
						>
					</button>
				{/each}
			</div>
		</section>

		<!-- Guests -->
		<section class="modal-section">
			<h4 class="mb-5 text-[17px] font-semibold tracking-[-0.01em] text-zinc-900">Кто едет?</h4>
			<div class="space-y-4 rounded-2xl border border-zinc-300/80 bg-white p-5 shadow-sm">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-[15px] font-medium text-zinc-900">Взрослые</p>
						<p class="text-[13.5px] text-zinc-500">От 18 лет</p>
					</div>
					<div class="flex items-center gap-4">
						<button
							type="button"
							onclick={() => setAdults(adults - 1)}
							disabled={adults === 0 || (adults === 1 && children > 0)}
							class="flex h-11 w-11 touch-manipulation items-center justify-center rounded-full border border-zinc-300 text-zinc-600
                           transition-all active:scale-90 disabled:scale-100 disabled:border-zinc-100 disabled:bg-zinc-50 disabled:text-zinc-300"
						>
							<Minus size={18} strokeWidth={2} />
						</button>
						<span class="w-6 text-center text-[17px] font-semibold text-zinc-900 tabular-nums"
							>{adults}</span
						>
						<button
							type="button"
							onclick={() => setAdults(adults + 1)}
							disabled={totalGuests >= 20}
							class="flex h-11 w-11 touch-manipulation items-center justify-center rounded-full border border-zinc-300 text-zinc-600
                           transition-all active:scale-90 disabled:scale-100 disabled:border-zinc-100 disabled:bg-zinc-50 disabled:text-zinc-300"
						>
							<Plus size={18} strokeWidth={2} />
						</button>
					</div>
				</div>

				<div class="h-px bg-zinc-200/60"></div>

				<div class="flex items-center justify-between">
					<div>
						<p class="text-[15px] font-medium text-zinc-900">Дети</p>
						<p class="text-[13.5px] text-zinc-500">До 18 лет</p>
					</div>
					<div class="flex items-center gap-4">
						<button
							type="button"
							onclick={() => setChildren(children - 1)}
							disabled={children === 0}
							class="flex h-11 w-11 touch-manipulation items-center justify-center rounded-full border border-zinc-300 text-zinc-600
                           transition-all active:scale-90 disabled:scale-100 disabled:border-zinc-100 disabled:bg-zinc-50 disabled:text-zinc-300"
						>
							<Minus size={18} strokeWidth={2} />
						</button>
						<span class="w-6 text-center text-[17px] font-semibold text-zinc-900 tabular-nums"
							>{children}</span
						>
						<button
							type="button"
							onclick={() => setChildren(children + 1)}
							disabled={totalGuests >= 20 || adults === 0}
							class="flex h-11 w-11 touch-manipulation items-center justify-center rounded-full border border-zinc-300 text-zinc-600
                           transition-all active:scale-90 disabled:scale-100 disabled:border-zinc-100 disabled:bg-zinc-50 disabled:text-zinc-300"
						>
							<Plus size={18} strokeWidth={2} />
						</button>
					</div>
				</div>
			</div>
		</section>
	</div>

	{#snippet footer()}
		<div class="modal-footer shrink-0">
			<div class="flex items-center justify-between gap-10">
				<button
					type="button"
					onclick={reset}
					class="touch-manipulation text-[15px] font-semibold text-zinc-900 underline decoration-zinc-900/30 underline-offset-4 transition-all duration-200 hover:text-black hover:decoration-zinc-900 active:scale-95"
				>
					Очистить
				</button>

				<Button variant="solid" tone="primary" size="lg" radius="pill" onclick={apply}>
					<span class="flex items-center justify-center gap-2 font-semibold">
						<Search size={18} strokeWidth={2.5} />
						Найти жилье
					</span>
				</Button>
			</div>
		</div>
	{/snippet}
</Modal>
