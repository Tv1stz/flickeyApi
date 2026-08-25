<script lang="ts">
	import { geoApi } from '$lib/api/geo';
	import type { GeoSuggestItem } from '$lib/types/geo';
	import { MapPin, Search, Loader2, X, Building, Navigation } from 'lucide-svelte';

	interface Props {
		value?: string;
		latitude?: number;
		longitude?: number;
		selected?: GeoSuggestItem | null;
		placeholder?: string;
		disabled?: boolean;
		onselect?: (item: GeoSuggestItem) => void;
	}

	let {
		value = $bindable(''),
		latitude = $bindable(53.9006),
		longitude = $bindable(27.5590),
		selected = $bindable(null),
		placeholder = 'Начните вводить адрес (город, улица, дом)...',
		disabled = false,
		onselect
	}: Props = $props();

	let query = $state(value || '');
	let results = $state<GeoSuggestItem[]>([]);
	let isLoading = $state(false);
	let isOpen = $state(false);
	let highlightedIndex = $state(-1);

	let abortController: AbortController | null = null;
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	// Keep internal query in sync if parent value changes externally
	$effect(() => {
		if (value !== query && !isOpen) {
			query = value || '';
		}
	});

	function handleInput(e: Event) {
		const target = e.target as HTMLInputElement;
		query = target.value;
		value = query;

		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}

		if (query.trim().length < 2) {
			results = [];
			isOpen = false;
			isLoading = false;
			if (abortController) {
				abortController.abort();
			}
			return;
		}

		isLoading = true;
		isOpen = true;
		highlightedIndex = -1;

		debounceTimer = setTimeout(async () => {
			if (abortController) {
				abortController.abort();
			}
			abortController = new AbortController();

			try {
				const items = await geoApi.suggest(query, 'ru', 6, abortController.signal);
				results = items;
				isOpen = items.length > 0;
			} catch {
				results = [];
			} finally {
				isLoading = false;
			}
		}, 280);
	}

	function handleSelect(item: GeoSuggestItem) {
		const parsedLat = typeof item.latitude === 'number' ? item.latitude : parseFloat(String(item.latitude));
		const parsedLng = typeof item.longitude === 'number' ? item.longitude : parseFloat(String(item.longitude));

		const cleanItem: GeoSuggestItem = {
			...item,
			latitude: !isNaN(parsedLat) ? parsedLat : latitude,
			longitude: !isNaN(parsedLng) ? parsedLng : longitude
		};

		query = cleanItem.raw_address;
		value = cleanItem.raw_address;
		selected = cleanItem;

		if (!isNaN(parsedLat) && !isNaN(parsedLng)) {
			latitude = parsedLat;
			longitude = parsedLng;
		}

		isOpen = false;
		results = [];
		if (onselect) {
			onselect(cleanItem);
		}
	}

	function handleClear() {
		query = '';
		value = '';
		selected = null;
		results = [];
		isOpen = false;
		if (abortController) {
			abortController.abort();
		}
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (!isOpen || results.length === 0) return;

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			highlightedIndex = (highlightedIndex + 1) % results.length;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			highlightedIndex = (highlightedIndex - 1 + results.length) % results.length;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (highlightedIndex >= 0 && highlightedIndex < results.length) {
				handleSelect(results[highlightedIndex]);
			}
		} else if (e.key === 'Escape') {
			isOpen = false;
		}
	}
</script>

<div class="relative w-full">
	<div class="relative flex items-center">
		<!-- Left icon -->
		<div class="absolute left-3.5 flex items-center pointer-events-none text-muted-foreground">
			<MapPin class="h-4 w-4 text-primary" />
		</div>

		<!-- Input field -->
		<input
			type="text"
			value={query}
			oninput={handleInput}
			onkeydown={handleKeyDown}
			onfocus={() => {
				if (results.length > 0) isOpen = true;
			}}
			{placeholder}
			{disabled}
			class="w-full rounded-2xl border border-input bg-background pl-10 pr-10 py-3 text-xs font-medium text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all shadow-xs disabled:opacity-50"
			autocomplete="off"
		/>

		<!-- Right status / clear icons -->
		<div class="absolute right-3 flex items-center gap-1.5">
			{#if isLoading}
				<Loader2 class="h-4 w-4 animate-spin text-primary" />
			{:else if query}
				<button
					type="button"
					onclick={handleClear}
					class="p-0.5 rounded-full hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
					title="Очистить"
				>
					<X class="h-3.5 w-3.5" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Suggestions dropdown -->
	{#if isOpen && results.length > 0}
		<!-- Backdrop for clicking outside -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="fixed inset-0 z-30" onclick={() => (isOpen = false)}></div>

		<div
			class="absolute left-0 right-0 top-full mt-1.5 z-40 rounded-2xl border border-border bg-card shadow-2xl overflow-hidden divide-y divide-border/50 max-h-72 overflow-y-auto animate-in zoom-in-95 duration-100"
		>
			<div class="px-3 py-2 bg-muted/40 text-[10px] font-bold text-muted-foreground uppercase tracking-wider flex items-center justify-between">
				<span>Подходящие адреса</span>
				<span class="font-normal lowercase text-[10px]">Photon OSM</span>
			</div>

			{#each results as item, index (item.raw_address + index)}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					onclick={() => handleSelect(item)}
					onmouseenter={() => (highlightedIndex = index)}
					class="p-3 text-left transition-colors cursor-pointer flex items-start gap-3 {highlightedIndex === index ? 'bg-primary/10 text-primary' : 'hover:bg-accent'}"
				>
					<div class="w-7 h-7 rounded-xl bg-primary/10 flex items-center justify-center text-primary shrink-0 mt-0.5">
						{#if item.house_number}
							<Building class="h-3.5 w-3.5" />
						{:else}
							<Navigation class="h-3.5 w-3.5" />
						{/if}
					</div>

					<div class="min-w-0 flex-1">
						<div class="text-xs font-bold text-foreground leading-snug">
							{#if item.street}
								{item.street}{item.house_number ? ', ' + item.house_number : ''}
							{:else}
								{item.raw_address}
							{/if}
						</div>
						<div class="text-[11px] text-muted-foreground flex items-center gap-1.5 mt-0.5">
							{#if item.city}
								<span class="font-medium">{item.city}</span>
							{/if}
							{#if item.city && item.country}
								<span>•</span>
							{/if}
							{#if item.country}
								<span>{item.country}</span>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
