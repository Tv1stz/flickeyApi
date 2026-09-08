<!-- src/lib/components/host-listing-edit/sections/EditLocation.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import Map from '$lib/components/location/Map.svelte';
	import TextInputField from '$lib/components/form/TextInputField.svelte';
	import { cardStyles, mapOverlayStyles } from '$lib/config/styles';
	import { locationStore } from '$lib/stores/location.svelte';
	import { reverseGeocode, searchAddress } from '$lib/services/geocoding';
	import type { ParsedAddress } from '$lib/stores/location.svelte';
	import { LocateFixed, Search, MapPin, Loader2, Plus, Minus } from 'lucide-svelte';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	let mapRef: {
		flyTo: (lat: number, lng: number, zoom?: number) => void;
		zoomIn: (step?: number) => void;
		zoomOut: (step?: number) => void;
	} | null = $state(null);

	let searchQuery = $state('');
	let isSearching = $state(false);
	let isLocating = $state(false);
	let searchResults = $state<ParsedAddress[]>([]);
	let showSearchResults = $state(false);
	let searchDebounce: ReturnType<typeof setTimeout> | null = null;

	onMount(() => {
		if (form.latitude && form.longitude) {
			locationStore.setCoords(form.latitude, form.longitude);
		}
		if (form.address) {
			searchQuery = form.address;
		}
	});

	$effect(() => {
		if (form.address && !searchQuery) {
			searchQuery = form.address;
		}
	});

	function handleAddressInput(value: string) {
		searchQuery = value;
		onUpdate({ address: value });

		if (searchDebounce) clearTimeout(searchDebounce);
		if (value.trim().length < 3) {
			searchResults = [];
			showSearchResults = false;
			return;
		}

		searchDebounce = setTimeout(async () => {
			isSearching = true;
			try {
				const res = await searchAddress(value);
				searchResults = res;
				showSearchResults = res.length > 0;
			} catch (e) {
				console.error('Search error:', e);
			} finally {
				isSearching = false;
			}
		}, 350);
	}

	function selectResult(result: ParsedAddress) {
		searchQuery = result.displayName;
		showSearchResults = false;

		locationStore.setCoords(result.lat, result.lng);
		locationStore.setAddress(result);

		onUpdate({
			address: result.displayName,
			city: result.city,
			street: result.street,
			houseNumber: result.house,
			latitude: result.lat,
			longitude: result.lng
		});

		mapRef?.flyTo(result.lat, result.lng, 16);
	}

	function handleMapAddressResolved() {
		const addr = locationStore.address;
		if (addr) {
			searchQuery = addr.displayName;
			onUpdate({
				address: addr.displayName,
				city: addr.city,
				street: addr.street,
				houseNumber: addr.house,
				latitude: locationStore.lat,
				longitude: locationStore.lng
			});
		} else {
			onUpdate({
				latitude: locationStore.lat,
				longitude: locationStore.lng
			});
		}
	}

	function getCurrentLocation() {
		if (!navigator.geolocation) return;
		isLocating = true;
		navigator.geolocation.getCurrentPosition(
			async (position) => {
				const { latitude, longitude } = position.coords;
				locationStore.setCoords(latitude, longitude);
				mapRef?.flyTo(latitude, longitude, 16);

				try {
					const addr = await reverseGeocode(latitude, longitude);
					if (addr) {
						locationStore.setAddress(addr);
						searchQuery = addr.displayName;
						onUpdate({
							address: addr.displayName,
							city: addr.city,
							street: addr.street,
							houseNumber: addr.house,
							latitude,
							longitude
						});
					}
				} catch (e) {
					console.error('Reverse geocoding error:', e);
				} finally {
					isLocating = false;
				}
			},
			(err) => {
				console.warn('Geolocation error:', err);
				isLocating = false;
			},
			{ enableHighAccuracy: true, timeout: 10000 }
		);
	}
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Местоположение</h2>
	<p class={cardStyles.subtitle}>Укажите точный адрес вашего жилья на карте</p>

	<div class={cardStyles.contentGap}>
		<!-- Address Search -->
		<div class="relative">
			<TextInputField
				id="address"
				label="Адрес жилья"
				placeholder="Введите город, улицу и номер дома"
				value={searchQuery}
				onInput={handleAddressInput}
				onBlur={() => {
					setTimeout(() => {
						showSearchResults = false;
					}, 250);
				}}
			/>

			<div class="absolute right-3 top-8 flex items-center gap-1.5">
				{#if isSearching || isLocating}
					<Loader2 class="h-5 w-5 animate-spin text-zinc-400" />
				{:else}
					<button
						type="button"
						class={mapOverlayStyles.iconButton}
						onclick={getCurrentLocation}
						title="Моё местоположение"
					>
						<LocateFixed class="h-4 w-4 text-zinc-700" />
					</button>
				{/if}
			</div>

			<!-- Suggestions Dropdown -->
			{#if showSearchResults && searchResults.length > 0}
				<div
					class="absolute z-20 mt-2 max-h-60 w-full overflow-y-auto rounded-2xl border border-zinc-200 bg-white p-1.5 shadow-xl"
				>
					{#each searchResults as result}
						<button
							type="button"
							class="flex w-full items-start gap-3 rounded-xl px-3.5 py-2.5 text-left transition-colors hover:bg-zinc-100/70"
							onmousedown={() => selectResult(result)}
						>
							<MapPin class="mt-0.5 h-4 w-4 shrink-0 text-zinc-400" />
							<div class="min-w-0 flex-1">
								<p class="truncate text-xs font-semibold text-zinc-900">{result.displayName}</p>
								<p class="mt-0.5 truncate text-[11px] text-zinc-500">
									{[result.city, result.district].filter(Boolean).join(', ')}
								</p>
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Map Container -->
		<div class="relative h-[420px] overflow-hidden rounded-3xl border border-zinc-200 shadow-sm">
			<Map bind:this={mapRef} onAddressResolved={handleMapAddressResolved} />

			<!-- Center Pin Marker -->
			<div class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center">
				<div class="-translate-y-4 filter drop-shadow-md">
					<div
						class="flex h-10 w-10 items-center justify-center rounded-full bg-zinc-900 text-white ring-4 ring-white"
					>
						<MapPin class="h-5 w-5" />
					</div>
					<div class="mx-auto h-2 w-2 rotate-45 bg-zinc-900"></div>
				</div>
			</div>

			<!-- Map Controls -->
			<div class="absolute right-4 bottom-4 z-10 flex flex-col gap-2">
				<button
					type="button"
					class="flex h-9 w-9 items-center justify-center rounded-xl border border-zinc-200 bg-white/95 text-zinc-800 shadow-md backdrop-blur-sm transition-transform hover:scale-105 active:scale-95"
					onclick={() => mapRef?.zoomIn()}
					title="Приблизить"
				>
					<Plus class="h-4 w-4" />
				</button>
				<button
					type="button"
					class="flex h-9 w-9 items-center justify-center rounded-xl border border-zinc-200 bg-white/95 text-zinc-800 shadow-md backdrop-blur-sm transition-transform hover:scale-105 active:scale-95"
					onclick={() => mapRef?.zoomOut()}
					title="Отдалить"
				>
					<Minus class="h-4 w-4" />
				</button>
			</div>

			<!-- Coordinates Pill -->
			{#if form.latitude && form.longitude}
				<div
					class="absolute bottom-4 left-4 z-10 rounded-xl bg-white/90 px-3 py-1.5 text-[11px] font-medium text-zinc-600 shadow-sm backdrop-blur-sm"
				>
					{form.latitude.toFixed(5)}, {form.longitude.toFixed(5)}
				</div>
			{/if}
		</div>
	</div>
</section>
