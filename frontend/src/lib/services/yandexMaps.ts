// src/lib/services/yandexMaps.ts
import { browser } from '$app/environment';
import type ymaps from 'yandex-maps';

let ymapsPromise: Promise<typeof ymaps> | null = null;

/**
 * Loads the Yandex Maps JavaScript API (v2.1) dynamically on the client.
 * Guarantees singleton script injection and resolves when `ymaps.ready()` completes.
 * Uses `coordorder=latlong` so coordinates are always [latitude, longitude].
 */
export function loadYandexMaps(apiKey?: string): Promise<typeof ymaps> {
	if (!browser) {
		return Promise.reject(new Error('Yandex Maps can only be loaded in browser environment'));
	}

	if (typeof window !== 'undefined' && (window as any).ymaps && (window as any).ymaps.Map) {
		return new Promise((resolve) => {
			(window as any).ymaps.ready(() => resolve((window as any).ymaps));
		});
	}

	if (ymapsPromise) {
		return ymapsPromise;
	}

	ymapsPromise = new Promise((resolve, reject) => {
		const existingScript = document.querySelector('script[src*="api-maps.yandex.ru"]');
		if (existingScript) {
			const checkInterval = setInterval(() => {
				const ym = (window as any).ymaps;
				if (ym && ym.ready) {
					clearInterval(checkInterval);
					ym.ready(() => resolve(ym));
				}
			}, 50);
			return;
		}

		const script = document.createElement('script');
		script.type = 'text/javascript';
		const keyParam = apiKey ? `&apikey=${encodeURIComponent(apiKey)}` : '';
		script.src = `https://api-maps.yandex.ru/2.1/?lang=ru_RU&coordorder=latlong${keyParam}`;
		script.async = true;

		script.onload = () => {
			const ym = (window as any).ymaps;
			if (!ym) {
				reject(new Error('Failed to initialize Yandex Maps SDK'));
				return;
			}
			ym.ready(() => {
				resolve(ym);
			});
		};

		script.onerror = (err) => {
			ymapsPromise = null;
			reject(new Error('Failed to load Yandex Maps script: ' + String(err)));
		};

		document.head.appendChild(script);
	});

	return ymapsPromise;
}
