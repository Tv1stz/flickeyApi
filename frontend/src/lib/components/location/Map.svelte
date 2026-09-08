<!-- src/lib/components/location/Map.svelte -->
<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { locationStore } from '$lib/stores/location.svelte';
	import { reverseGeocode } from '$lib/services/geocoding';
	import { loadYandexMaps } from '$lib/services/yandexMaps';

	interface Props {
		onAddressResolved?: () => void;
	}
	const { onAddressResolved }: Props = $props();

	let mapContainer: HTMLDivElement;
	let map: ymaps.Map | null = null;
	let isDestroyed = false;

	// ── Управление состоянием через счётчик движений ──────────────────────────
	let moveGeneration = 0;
	let geocodeTimer: ReturnType<typeof setTimeout>;
	let activeAbort: AbortController | null = null;

	const GEOCODE_DEBOUNCE = 400;

	onMount(() => {
		isDestroyed = false;

		loadYandexMaps()
			.then((ymapsInstance) => {
				if (isDestroyed || !mapContainer) return;

				map = new ymapsInstance.Map(
					mapContainer,
					{
						center: [locationStore.lat, locationStore.lng],
						zoom: locationStore.zoom,
						controls: []
					},
					{
						suppressMapOpenBlock: true,
						yandexMapDisablePoiInteractivity: true,
						autoFitToViewport: 'always'
					}
				);

				// Отслеживаем начало движения
				map.events.add('actionbegin', () => {
					activeAbort?.abort();
					activeAbort = null;
					clearTimeout(geocodeTimer);
					moveGeneration++;
					locationStore.setMapMoving(true);
					locationStore.setError(null);
				});

				// Промежуточное обновление координат
				map.events.add('boundschange', (e: any) => {
					if (!map) return;
					const newCenter = e.get('newCenter');
					if (newCenter && newCenter.length >= 2) {
						locationStore.setCoords(newCenter[0], newCenter[1]);
					}
				});

				// Окончание движения карты (drag или fly)
				map.events.add('actionend', () => {
					if (!map) return;
					const c = map.getCenter();
					const z = map.getZoom();
					locationStore.setCoords(c[0], c[1]);
					locationStore.setZoom(z);

					const generation = moveGeneration;
					clearTimeout(geocodeTimer);

					geocodeTimer = setTimeout(async () => {
						if (generation !== moveGeneration) return;

						activeAbort?.abort();
						const abort = new AbortController();
						activeAbort = abort;

						locationStore.setLoading(true);

						try {
							const result = await reverseGeocode(c[0], c[1], abort.signal);

							if (generation !== moveGeneration) return;
							if (abort.signal.aborted) return;

							if (result) {
								locationStore.setAddress(result);
								onAddressResolved?.();
							} else {
								locationStore.setError('Не удалось определить адрес');
							}
						} finally {
							if (generation === moveGeneration) {
								locationStore.setLoading(false);
								locationStore.setMapMoving(false);
							}
						}
					}, GEOCODE_DEBOUNCE);
				});

				// Первоначальное определение адреса при старте
				const generation = moveGeneration;
				locationStore.setLoading(true);
				reverseGeocode(locationStore.lat, locationStore.lng).then((result) => {
					if (generation === moveGeneration && result) {
						locationStore.setAddress(result);
						locationStore.setLoading(false);
					}
				});
			})
			.catch((err) => {
				console.error('Failed to load Yandex Maps in Step2Location:', err);
				locationStore.setError('Не удалось загрузить карты Яндекса');
			});
	});

	onDestroy(() => {
		isDestroyed = true;
		clearTimeout(geocodeTimer);
		activeAbort?.abort();
		if (map) {
			map.destroy();
			map = null;
		}
	});

	export function flyTo(lat: number, lng: number, zoom = 16) {
		if (!map) return;
		map.setCenter([lat, lng], zoom, {
			duration: 500,
			checkZoomRange: true
		});
	}

	export function zoomIn(step = 1) {
		if (!map) return;
		map.setZoom(map.getZoom() + step, { duration: 180 });
	}

	export function zoomOut(step = 1) {
		if (!map) return;
		map.setZoom(map.getZoom() - step, { duration: 180 });
	}
</script>

<div bind:this={mapContainer} class="absolute inset-0 h-full w-full"></div>
