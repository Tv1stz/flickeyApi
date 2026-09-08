// src/lib/stores/location.svelte.ts

export interface ParsedAddress {
	country: string;
	city: string;
	district: string;
	street: string;
	house: string;
	displayName: string;
	lat: number;
	lng: number;
	confidence: 'full' | 'no-house' | 'street-only';
	raw?: unknown;
}

function createLocationStore() {
	let lat = $state(55.7558);
	let lng = $state(37.6173);
	let zoom = $state(13);
	let address = $state<ParsedAddress | null>(null);
	let isLoading = $state(false);
	let isMapMoving = $state(false);
	let error = $state<string | null>(null);

	return {
		get lat() {
			return lat;
		},
		get lng() {
			return lng;
		},
		get zoom() {
			return zoom;
		},
		get address() {
			return address;
		},
		get isLoading() {
			return isLoading;
		},
		get isMapMoving() {
			return isMapMoving;
		},
		get error() {
			return error;
		},

		setCoords(la: number, ln: number) {
			// Не обновляем если изменение меньше 1см — убирает дрожание
			if (Math.abs(la - lat) < 0.000001 && Math.abs(ln - lng) < 0.000001) return;
			lat = la;
			lng = ln;
		},

		setZoom(z: number) {
			if (Math.abs(z - zoom) < 0.01) return;
			zoom = z;
		},

		setAddress(a: ParsedAddress | null) {
			address = a;
		},

		setLoading(v: boolean) {
			if (isLoading === v) return; // не триггерим реактивность без изменений
			isLoading = v;
		},

		setMapMoving(v: boolean) {
			if (isMapMoving === v) return;
			isMapMoving = v;
		},

		setError(e: string | null) {
			error = e;
		}
	};
}

export const locationStore = createLocationStore();
