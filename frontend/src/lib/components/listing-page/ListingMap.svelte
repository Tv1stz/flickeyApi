<!-- src/lib/components/listing-page/ListingMap.svelte -->
<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { MapPin, Maximize2, Minimize2, Minus, Plus } from 'lucide-svelte';
	import type { ListingLocation } from '$lib/components/card/types';
	import { mapOverlayStyles } from '$lib/config/styles';
	import { formatListingAddressFull } from '$lib/utils/location';
	import { loadYandexMaps } from '$lib/services/yandexMaps';

	interface Props {
		location?: ListingLocation;
		latitude?: number;
		longitude?: number;
		address?: string;
		city?: string;
		street?: string;
	}

	let { location, latitude, longitude, address = '', city = '', street = '' }: Props = $props();

	let mapContainer = $state<HTMLDivElement | null>(null);
	let mapInstance: ymaps.Map | null = null;
	let placemarkInstance: ymaps.Placemark | null = null;
	let isFullscreen = $state(false);
	let isDestroyed = false;

	const effectiveLat = $derived(location?.lat ?? latitude);
	const effectiveLng = $derived(location?.lng ?? longitude);

	const fullAddress = $derived(
		address ||
		(location ? formatListingAddressFull(location.address || '', location) : '') ||
		(city && street ? `${city}, ${street}` : city || 'Беларусь')
	);

	const hasCoordinates = $derived(
		effectiveLat !== undefined &&
		effectiveLng !== undefined &&
		!isNaN(Number(effectiveLat)) &&
		!isNaN(Number(effectiveLng)) &&
		Number(effectiveLat) !== 0 &&
		Number(effectiveLng) !== 0
	);

	const mapFrameClass = $derived(
		isFullscreen
			? 'fixed inset-0 z-[1200] h-[100dvh] w-screen overflow-hidden bg-zinc-100'
			: 'relative mb-4 aspect-square overflow-hidden rounded-3xl bg-zinc-100 sm:aspect-auto sm:h-[320px] lg:h-[380px]'
	);

	function initMap() {
		if (!mapContainer || typeof window === 'undefined' || !hasCoordinates || !effectiveLat || !effectiveLng) return;

		isDestroyed = false;

		loadYandexMaps()
			.then((ymaps) => {
				if (isDestroyed || !mapContainer) return;

				if (mapInstance) {
					mapInstance.destroy();
					mapInstance = null;
					placemarkInstance = null;
				}

				const lat = Number(effectiveLat);
				const lng = Number(effectiveLng);

				mapInstance = new ymaps.Map(
					mapContainer,
					{
						center: [lat, lng],
						zoom: 16,
						controls: []
					},
					{
						suppressMapOpenBlock: true,
						autoFitToViewport: 'always'
					}
				);

				// Создаем аккуратную метку с адресом
				placemarkInstance = new ymaps.Placemark(
					[lat, lng],
					{
						hintContent: fullAddress,
						balloonContentHeader: '<span style="font-weight: 700;">Объект размещения</span>',
						balloonContentBody: `<p style="margin: 4px 0 0; font-size: 13px;">${fullAddress}</p>`
					},
					{
						preset: 'islands#darkGreenDotIconWithCaption',
						iconColor: '#059669'
					}
				);

				mapInstance.geoObjects.add(placemarkInstance);
			})
			.catch((err) => {
				console.error('Failed to initialize Yandex Map in ListingMap:', err);
			});
	}

	onMount(() => {
		initMap();
	});

	$effect(() => {
		if (hasCoordinates && effectiveLat && effectiveLng) {
			const lat = Number(effectiveLat);
			const lng = Number(effectiveLng);

			if (mapInstance && placemarkInstance) {
				placemarkInstance.geometry?.setCoordinates([lat, lng]);
				mapInstance.setCenter([lat, lng], mapInstance.getZoom());
			} else if (mapContainer && !mapInstance) {
				initMap();
			}
		}
	});

	onDestroy(() => {
		isDestroyed = true;
		if (mapInstance) {
			mapInstance.destroy();
			mapInstance = null;
			placemarkInstance = null;
		}
	});

	function zoomInMap() {
		if (mapInstance) {
			mapInstance.setZoom(mapInstance.getZoom() + 1, { duration: 180 });
		}
	}

	function zoomOutMap() {
		if (mapInstance) {
			mapInstance.setZoom(mapInstance.getZoom() - 1, { duration: 180 });
		}
	}

	function toggleFullscreen() {
		isFullscreen = !isFullscreen;
		if (typeof document !== 'undefined') {
			document.body.classList.toggle('listing-map-fullscreen-active', isFullscreen);
		}
		setTimeout(() => {
			if (mapInstance && mapInstance.container) {
				mapInstance.container.fitToViewport();
			}
		}, 150);
	}
</script>

<div data-listing-map class={isFullscreen ? 'relative z-[1200]' : ''}>
	{#if !isFullscreen}
		<h2 class="mb-5 text-[20px] font-semibold tracking-tight text-zinc-900">Где вы будете</h2>
	{/if}

	<div class={mapFrameClass}>
		{#if hasCoordinates}
			<div bind:this={mapContainer} class="absolute inset-0 h-full w-full"></div>

			{#if isFullscreen}
				<div
					class="pointer-events-none absolute inset-x-0 top-0 z-20 h-20 bg-gradient-to-b from-black/30 to-transparent"
				></div>
			{/if}

			<div class="absolute top-3 right-3 z-30">
				<button
					type="button"
					onclick={() => toggleFullscreen()}
					class={mapOverlayStyles.iconButton}
					aria-label={isFullscreen ? 'Свернуть карту' : 'Развернуть карту на весь экран'}
					aria-pressed={isFullscreen}
				>
					{#if isFullscreen}
						<Minimize2 class="h-5 w-5" />
					{:else}
						<Maximize2 class="h-5 w-5" />
					{/if}
				</button>
			</div>

			<div class="absolute right-3 bottom-3 z-30 flex flex-col gap-2">
				<button
					type="button"
					onclick={zoomInMap}
					class={mapOverlayStyles.iconButton}
					aria-label="Приблизить карту"
				>
					<Plus class="h-5 w-5" />
				</button>
				<button
					type="button"
					onclick={zoomOutMap}
					class={mapOverlayStyles.iconButton}
					aria-label="Отдалить карту"
				>
					<Minus class="h-5 w-5" />
				</button>
			</div>
		{:else}
			<div class="absolute inset-0 flex items-center justify-center bg-zinc-50">
				<div class="relative flex flex-col items-center gap-3 px-5 text-center">
					<div
						class="flex h-12 w-12 items-center justify-center rounded-full bg-emerald-600 shadow-xl"
					>
						<MapPin size={20} class="text-white" />
					</div>
					<p class="text-[14px] font-medium text-zinc-500">Координаты для карты не указаны</p>
				</div>
			</div>
		{/if}
	</div>

	{#if !isFullscreen}
		<div class="flex items-start gap-2.5">
			<MapPin size={15} strokeWidth={1.5} class="mt-0.5 shrink-0 text-emerald-600" />
			<p class="min-w-0 text-[14px] leading-snug font-medium text-zinc-800">{fullAddress}</p>
		</div>
	{/if}
</div>

<style>
	:global(body.listing-map-fullscreen-active [data-listing-mobile-header]),
	:global(body.listing-map-fullscreen-active [data-listing-price-bar]) {
		opacity: 0;
		pointer-events: none;
	}
</style>
