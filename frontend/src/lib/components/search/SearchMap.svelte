<!-- src/lib/components/search/SearchMap.svelte -->
<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import type { Listing } from '$lib/components/card/types';
	import { loadYandexMaps } from '$lib/services/yandexMaps';
	import SearchMapQuickCard from './SearchMapQuickCard.svelte';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';

	interface Props {
		listings: Listing[];
		selectedId?: string | null;
		class?: string;
		onSelect?: (listing: Listing | null) => void;
	}

	let {
		listings = [],
		selectedId = null,
		class: className = '',
		onSelect
	}: Props = $props();

	let mapContainer = $state<HTMLDivElement | null>(null);
	let mapInstance: ymaps.Map | null = null;
	let clustererInstance: ymaps.Clusterer | null = null;
	let placemarksMap = new Map<string, ymaps.Placemark>();
	let selectedListing = $state<Listing | null>(null);
	let isDestroyed = false;

	function highlightPlacemark(id: string | null) {
		placemarksMap.forEach((pm, pmId) => {
			const isActive = pmId === id;
			pm.properties.set('activeClass', isActive ? 'active' : '');
			pm.options.set('zIndex', isActive ? 1000 : 1);
		});
	}

	function selectListing(item: Listing | null) {
		selectedListing = item;
		highlightPlacemark(item?.id ?? null);
		onSelect?.(item);
	}

	function syncListingsOnMap(ymaps: typeof window.ymaps) {
		if (!mapInstance || !ymaps) return;

		if (clustererInstance) {
			mapInstance.geoObjects.remove(clustererInstance as any);
			clustererInstance = null;
		}
		placemarksMap.clear();

		const validListings = listings.filter((l) => {
			const lat = l.location?.lat;
			const lng = l.location?.lng;
			return lat && lng && !isNaN(Number(lat)) && !isNaN(Number(lng));
		});

		if (validListings.length === 0) {
			mapInstance.setCenter([53.9006, 27.5590], 12);
			return;
		}

		// Создаем кастомный HTML-layout для бейджей цен в стиле Airbnb
		const PricePinLayout = ymaps.templateLayoutFactory.createClass(
			'<div class="flickey-price-pin $[properties.activeClass]">' +
				'<span class="flickey-price-val">$[properties.priceText]</span>' +
			'</div>'
		);

		clustererInstance = new (ymaps.Clusterer as any)({
			preset: 'islands#darkGreenClusterIcons',
			groupByCoordinates: false,
			clusterHideIconOnBalloonOpen: false,
			geoObjectHideIconOnBalloonOpen: false,
			hasBalloon: false
		});

		const placemarks = validListings.map((item) => {
			const lat = Number(item.location?.lat);
			const lng = Number(item.location?.lng);
			const priceText = `${item.pricePerNight} BYN`;
			const isCurrent = selectedListing?.id === item.id || selectedId === item.id;

			const placemark = new ymaps.Placemark(
				[lat, lng],
				{
					listingId: item.id,
					priceText,
					activeClass: isCurrent ? 'active' : ''
				},
				{
					iconLayout: PricePinLayout,
					iconOffset: [-36, -16],
					iconShape: {
						type: 'Rectangle',
						coordinates: [[-36, -16], [36, 16]]
					} as any,
					openBalloonOnClick: false,
					hasBalloon: false,
					zIndex: isCurrent ? 1000 : 1
				}
			);

			placemark.events.add('click', (e: any) => {
				e.preventDefault();
				selectListing(item);
			});

			placemarksMap.set(item.id, placemark);
			return placemark;
		});

		if (clustererInstance) {
			clustererInstance.add(placemarks);
			mapInstance.geoObjects.add(clustererInstance as any);

			const bounds = clustererInstance.getBounds();
			if (bounds) {
				mapInstance.setBounds(bounds, {
					checkZoomRange: true,
					zoomMargin: [40, 40]
				});
			}
		}
	}

	onMount(() => {
		isDestroyed = false;

		loadYandexMaps()
			.then((ymaps) => {
				if (isDestroyed || !mapContainer) return;

				mapInstance = new ymaps.Map(
					mapContainer,
					{
						center: [53.9006, 27.5590],
						zoom: 12,
						controls: ['zoomControl', 'geolocationControl']
					},
					{
						suppressMapOpenBlock: true,
						autoFitToViewport: 'always'
					}
				);

				mapInstance.events.add('click', (e: any) => {
					// Клик по карте закрывает превью если кликнули не на метку
					if (!e.get('target')?.properties?.get('listingId')) {
						selectListing(null);
					}
				});

				syncListingsOnMap(ymaps);
			})
			.catch((err) => {
				console.error('Failed to load Yandex Maps in SearchMap:', err);
			});
	});

	$effect(() => {
		const currentListings = listings;
		if (mapInstance && typeof window !== 'undefined' && (window as any).ymaps) {
			syncListingsOnMap((window as any).ymaps);
		}
	});

	$effect(() => {
		if (selectedId !== undefined) {
			const found = listings.find((l) => l.id === selectedId) ?? null;
			if (found && selectedListing?.id !== found.id) {
				selectedListing = found;
				highlightPlacemark(found.id);
			} else if (!selectedId && selectedListing) {
				selectedListing = null;
				highlightPlacemark(null);
			}
		}
	});

	onDestroy(() => {
		isDestroyed = true;
		if (mapInstance) {
			mapInstance.destroy();
			mapInstance = null;
			clustererInstance = null;
			placemarksMap.clear();
		}
	});

	export function fitToViewport() {
		if (mapInstance && mapInstance.container) {
			mapInstance.container.fitToViewport();
		}
	}
</script>

<div class="relative h-full w-full overflow-hidden {className}">
	<div bind:this={mapContainer} class="h-full w-full"></div>

	{#if selectedListing}
		<div
			class="pointer-events-auto absolute bottom-6 left-1/2 -translate-x-1/2 z-30 transition-all duration-200"
		>
			<SearchMapQuickCard
				listing={selectedListing}
				isFavorite={favoritesStore.has(selectedListing.id)}
				onFavoriteToggle={() => {
					if (selectedListing) favoritesStore.toggle(selectedListing);
				}}
				onClose={() => selectListing(null)}
				onOpen={() => {
					if (selectedListing) goto(`/listings/${selectedListing.id}`);
				}}
			/>
		</div>
	{/if}
</div>

<style>
	:global(.flickey-price-pin) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 6px 11px;
		background: #ffffff;
		color: #18181b;
		border: 1px solid #d4d4d8;
		border-radius: 9999px;
		font-size: 13px;
		font-weight: 700;
		line-height: 1;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.14);
		cursor: pointer;
		transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);
		white-space: nowrap;
		user-select: none;
	}

	:global(.flickey-price-pin:hover),
	:global(.flickey-price-pin.active) {
		background: #18181b;
		color: #ffffff;
		border-color: #18181b;
		transform: scale(1.08);
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.24);
		z-index: 1000;
	}
</style>
