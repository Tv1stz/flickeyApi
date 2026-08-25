<script lang="ts">
	import { onMount } from 'svelte';
	import type * as maplibregl from 'maplibre-gl';
	import { MapPin, Compass, Navigation, Info } from 'lucide-svelte';

	interface Props {
		latitude?: number;
		longitude?: number;
		address?: string;
		zoom?: number;
		readonly?: boolean;
	}

	let {
		latitude = $bindable(53.9006),
		longitude = $bindable(27.5590),
		address = '',
		zoom = 15,
		readonly = false
	}: Props = $props();

	// Explicitly non-reactive DOM and Map references to avoid Svelte 5 deep Proxy interception & memory leaks
	let mapContainer: HTMLDivElement | null = null;
	let map: maplibregl.Map | null = null;
	let marker: maplibregl.Marker | null = null;
	let isMapReady = $state(false);

	// Epsilon tracking to prevent cyclical reactivity feedback loops between drag events and $effect()
	let lastSyncedLat = Number(latitude) || 53.9006;
	let lastSyncedLng = Number(longitude) || 27.5590;

	function updateCoordsFromUser(lat: number, lng: number) {
		lastSyncedLat = lat;
		lastSyncedLng = lng;
		latitude = Number(lat.toFixed(6));
		longitude = Number(lng.toFixed(6));
	}

	// Apply Russian language override to vector symbol text layers
	function applyRussianLabels(mapInstance: maplibregl.Map) {
		const style = mapInstance.getStyle();
		if (!style || !style.layers) return;
		for (const layer of style.layers) {
			if (layer.type === 'symbol' && layer.layout && (layer.layout as any)['text-field']) {
				try {
					mapInstance.setLayoutProperty(layer.id, 'text-field', [
						'coalesce',
						['get', 'name:ru'],
						['get', 'name_ru'],
						['get', 'name:latin'],
						['get', 'name']
					]);
				} catch {
					// Skip layers with fixed non-coalesce strings
				}
			}
		}
	}

	onMount(() => {
		if (!mapContainer || typeof window === 'undefined') return;

		let isDestroyed = false;

		// Dynamically import MapLibre GL to prevent SSR issues
		import('maplibre-gl').then((mlgl) => {
			if (isDestroyed || !mapContainer) return;
			const maplibreglInstance = mlgl;

			const initialLat = Number(latitude) || 53.9006;
			const initialLng = Number(longitude) || 27.5590;
			lastSyncedLat = initialLat;
			lastSyncedLng = initialLng;

			// Clean CartoDB Voyager raster + vector tiles with Russian localization
			const mapStyle: maplibregl.StyleSpecification = {
				version: 8,
				sources: {
					'carto-voyager': {
						type: 'raster',
						tiles: [
							'https://a.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}@2x.png',
							'https://b.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}@2x.png',
							'https://c.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}@2x.png',
							'https://d.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}@2x.png'
						],
						tileSize: 256,
						attribution:
							'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
					}
				},
				layers: [
					{
						id: 'carto-voyager-tiles',
						type: 'raster',
						source: 'carto-voyager',
						minzoom: 0,
						maxzoom: 20
					}
				]
			};

			map = new maplibreglInstance.Map({
				container: mapContainer,
				style: mapStyle,
				center: [initialLng, initialLat],
				zoom: zoom,
				attributionControl: false
			});

			map.addControl(new maplibreglInstance.NavigationControl({ showCompass: false }), 'top-right');

			map.on('styledata', () => {
				if (map) applyRussianLabels(map);
			});

			// Create custom pin marker element
			const pinEl = document.createElement('div');
			pinEl.className = 'flickey-maplibre-pin';
			pinEl.innerHTML = `
				<div class="relative flex items-center justify-center -translate-x-1/2 -translate-y-full cursor-grab active:cursor-grabbing">
					<div class="w-8 h-8 rounded-full bg-primary text-white flex items-center justify-center shadow-lg border-2 border-white transform transition-transform hover:scale-110">
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
							<path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/>
							<circle cx="12" cy="10" r="3"/>
						</svg>
					</div>
					<div class="absolute -bottom-1 w-2 h-2 rounded-full bg-primary/40 animate-ping"></div>
				</div>
			`;

			marker = new maplibreglInstance.Marker({
				element: pinEl,
				draggable: !readonly,
				anchor: 'bottom'
			})
				.setLngLat([initialLng, initialLat])
				.addTo(map);

			// Marker drag event
			if (!readonly) {
				marker.on('dragend', () => {
					if (!marker) return;
					const lngLat = marker.getLngLat();
					updateCoordsFromUser(lngLat.lat, lngLat.lng);
				});

				// Map click event
				map.on('click', (e: maplibregl.MapMouseEvent) => {
					if (!marker) return;
					marker.setLngLat([e.lngLat.lng, e.lngLat.lat]);
					updateCoordsFromUser(e.lngLat.lat, e.lngLat.lng);
				});
			}

			map.on('load', () => {
				isMapReady = true;
				if (map) map.resize();
			});

			// Fallback ready in case load fired early
			setTimeout(() => {
				if (!isDestroyed && !isMapReady) {
					isMapReady = true;
					if (map) map.resize();
				}
			}, 200);
		});

		// Destroy and cleanup on unmount to prevent WebGL leaks
		return () => {
			isDestroyed = true;
			isMapReady = false;
			if (map) {
				map.remove();
				map = null;
				marker = null;
			}
		};
	});

	// Reactive position sync when coordinates change externally (e.g. from Autocomplete selection)
	$effect(() => {
		const ready = isMapReady;
		const lat = Number(latitude);
		const lng = Number(longitude);

		if (!ready || !map || !marker || isNaN(lat) || isNaN(lng)) return;

		const deltaLat = Math.abs(lat - lastSyncedLat);
		const deltaLng = Math.abs(lng - lastSyncedLng);

		// Fly to new location only if change is significant (> 0.00005)
		if (deltaLat > 0.00005 || deltaLng > 0.00005) {
			lastSyncedLat = lat;
			lastSyncedLng = lng;
			marker.setLngLat([lng, lat]);
			map.flyTo({
				center: [lng, lat],
				zoom: 15,
				essential: true
			});
		}
	});
</script>

<div class="space-y-2.5">
	<!-- Map Container with explicit fixed height to prevent layout shifts & canvas glitches -->
	<div class="relative w-full h-[320px] rounded-3xl overflow-hidden border border-border shadow-sm bg-muted/30">
		<div bind:this={mapContainer} class="w-full h-full z-0"></div>

		<!-- Coordinates badge overlay -->
		<div
			class="absolute bottom-3 left-3 z-10 flex items-center gap-2 bg-background/95 backdrop-blur-md px-3 py-1.5 rounded-xl border border-border/80 text-[11px] font-mono text-foreground shadow-xs pointer-events-none"
		>
			<Compass class="h-3.5 w-3.5 text-primary shrink-0" />
			<span>{Number(latitude).toFixed(4)}° N, {Number(longitude).toFixed(4)}° E</span>
		</div>

		<!-- Instruction tooltip on top -->
		{#if !readonly}
			<div
				class="absolute top-3 left-3 z-10 flex items-center gap-1.5 bg-background/90 backdrop-blur-md px-2.5 py-1 rounded-lg border border-border/80 text-[10px] font-medium text-muted-foreground shadow-xs pointer-events-none"
			>
				<Info class="h-3 w-3 text-primary shrink-0" />
				<span>Перетащите метку или нажмите на карту для уточнения</span>
			</div>
		{/if}
	</div>

	<!-- Selected Address Caption -->
	{#if address}
		<div class="flex items-center gap-2 px-1 text-xs text-muted-foreground">
			<MapPin class="h-3.5 w-3.5 text-primary shrink-0" />
			<span class="font-medium text-foreground truncate">{address}</span>
		</div>
	{/if}
</div>

<style>
	:global(.maplibregl-canvas) {
		outline: none;
	}
	:global(.flickey-maplibre-pin) {
		background: transparent;
		border: none;
	}
</style>
