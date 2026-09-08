// src/lib/stores/favoritesStore.svelte.ts
import { browser } from '$app/environment';
import type { Listing } from '$lib/components/card/types';
import { authStore } from '$lib/stores/authStore.svelte';
import { normalizeListing } from '$lib/utils/listingConverters';
import { untrack } from 'svelte';

export interface FavoriteItem {
	id: string;
	listing?: Listing;
	addedAt: number;
}

const STORAGE_VERSION = 1;
const STORAGE_PREFIX = 'favorites';

function storageKey(userId: string): string {
	return `${STORAGE_PREFIX}:v${STORAGE_VERSION}:${userId}`;
}

function readFromStorage(userId: string): Record<string, FavoriteItem> {
	if (!browser) return {};
	try {
		const raw = localStorage.getItem(storageKey(userId));
		if (!raw) return {};
		const parsed = JSON.parse(raw) as Record<string, FavoriteItem>;
		return parsed ?? {};
	} catch (e) {
		console.warn('Не удалось загрузить избранное:', e);
		return {};
	}
}

function writeToStorage(userId: string, items: Record<string, FavoriteItem>): void {
	if (!browser) return;
	try {
		localStorage.setItem(storageKey(userId), JSON.stringify(items));
	} catch (e) {
		console.warn('Не удалось сохранить избранное:', e);
	}
}

function createFavoritesStore() {
	let userId = $state<string | null>(null);
	let items = $state<Record<string, FavoriteItem>>({});
	let ready = $state(false);
	let initialized = false;

	return {
		initialize() {
			if (!browser || initialized) return;
			initialized = true;

			$effect.root(() => {
				$effect(() => {
					const user = authStore.user;
					if (!user?.id) {
						userId = null;
						items = {};
						ready = true;
						return;
					}

					if (user.id !== userId) {
						userId = user.id;
						items = readFromStorage(user.id);
						ready = true;
					}
				});
			});
		},

		toggle(listing: Listing): boolean {
			if (!userId) return false;

			const id = listing.id;
			const existing = items[id];
			const nextItems = { ...items };

			if (existing) {
				delete nextItems[id];
			} else {
				nextItems[id] = {
					id,
					listing: normalizeListing(listing),
					addedAt: Date.now()
				};
			}

			items = nextItems;
			writeToStorage(userId, nextItems);
			return !existing;
		},

		add(listing: Listing): void {
			if (!userId) return;
			const id = listing.id;
			if (items[id]) return;

			const nextItems = {
				...items,
				[id]: { id, listing: normalizeListing(listing), addedAt: Date.now() }
			};
			items = nextItems;
			writeToStorage(userId, nextItems);
		},

		remove(id: string): void {
			if (!userId || !items[id]) return;
			const nextItems = { ...items };
			delete nextItems[id];
			items = nextItems;
			writeToStorage(userId, nextItems);
		},

		clearAll(): void {
			if (!userId) return;
			items = {};
			writeToStorage(userId, {});
		},

		has(id: string): boolean {
			return Boolean(items[id]);
		},

		// ═══════════════════════════════════════════════════════════════════════════
		// REACTIVE GETTERS
		// ═══════════════════════════════════════════════════════════════════════════

		get ready() {
			return ready;
		},
		get favoriteIds() {
			return new Set(Object.keys(items));
		},
		get favoriteItems() {
			return Object.values(items).sort((a, b) => b.addedAt - a.addedAt);
		},
		get favoriteListings() {
			return Object.values(items)
				.map((item) => item.listing)
				.filter((listing): listing is Listing => Boolean(listing));
		}
	};
}

export const favoritesStore = createFavoritesStore();
