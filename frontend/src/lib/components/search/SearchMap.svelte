<!-- src/lib/components/search/SearchMap.svelte -->
<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import type { Listing } from '$lib/components/card/types';
	import { loadYandexMaps } from '$lib/services/yandexMaps';
	import { formatBYN } from '$lib/utils/format';

	interface Props {
		listings: Listing[];
		class?: string;
	}

	let { listings = [], class: className = '' }: Props = $props();

	let mapContainer = $state<HTMLDivElement | null>(null);
	let mapInstance: ymaps.Map | null = null;
	let clustererInstance: ymaps.Clusterer | null = null;
	let isDestroyed = false;

	function syncListingsOnMap(ymaps: typeof window.ymaps) {
		if (!mapInstance || !ymaps) return;

		if (clustererInstance) {
			mapInstance.geoObjects.remove(clustererInstance as any);
			clustererInstance = null;
		}

		const validListings = listings.filter((l) => {
			const lat = l.location?.lat;
			const lng = l.location?.lng;
			return lat && lng && !isNaN(Number(lat)) && !isNaN(Number(lng));
		});

		if (validListings.length === 0) {
			// Если нет объектов с координатами, центрируем на Минске
			mapInstance.setCenter([53.9006, 27.5590], 12);
			return;
		}

		clustererInstance = new (ymaps.Clusterer as any)({
			preset: 'islands#darkGreenClusterIcons',
			groupByCoordinates: false,
			clusterHideIconOnBalloonOpen: false,
			geoObjectHideIconOnBalloonOpen: false
		});

		const placemarks = validListings.map((item) => {
			const lat = Number(item.location?.lat);
			const lng = Number(item.location?.lng);
			const coverImg = item.images?.[0] || '';
			const priceText = formatBYN(item.pricePerNight);

			const balloonHtml = `
				<div style="padding: 6px; max-width: 220px; font-family: inherit;">
					<a href="/listings/${item.id}" style="text-decoration: none; color: inherit; display: block;">
						${coverImg ? `<img src="${coverImg}" alt="${item.title}" style="width: 100%; height: 110px; object-fit: cover; border-radius: 12px; margin-bottom: 8px;" />` : ''}
						<div style="font-weight: 800; font-size: 15px; color: #18181b;">от ${priceText} <span style="font-size: 12px; font-weight: 500; color: #71717a;">/ ночь</span></div>
						<div style="font-size: 12px; font-weight: 600; color: #3f3f46; margin-top: 4px; line-height: 1.3;">${item.title}</div>
						<div style="font-size: 11px; color: #a1a1aa; margin-top: 2px;">${item.address || ''}</div>
					</a>
				</div>
			`;

			const placemark = new ymaps.Placemark(
				[lat, lng],
				{
					hintContent: `${item.title} — ${priceText}/ночь`,
					balloonContent: balloonHtml,
					iconCaption: priceText
				},
				{
					preset: 'islands#darkGreenDotIconWithCaption',
					iconColor: '#059669'
				}
			);

			return placemark;
		});

		if (clustererInstance) {
			clustererInstance.add(placemarks);
			mapInstance.geoObjects.add(clustererInstance as any);

			// Автоматически подгоняем масштаб и границы
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

	onDestroy(() => {
		isDestroyed = true;
		if (mapInstance) {
			mapInstance.destroy();
			mapInstance = null;
			clustererInstance = null;
		}
	});

	export function fitToViewport() {
		if (mapInstance && mapInstance.container) {
			mapInstance.container.fitToViewport();
		}
	}
</script>

<div bind:this={mapContainer} class="h-full w-full {className}"></div>
