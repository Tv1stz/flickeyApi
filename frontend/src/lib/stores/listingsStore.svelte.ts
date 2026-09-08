// src/lib/stores/listingsStore.svelte.ts
import type { Listing } from '$lib/components/card/types';
import { browser } from '$app/environment';
import { apiListingToCardListing, normalizeListing } from '$lib/utils/listingConverters';
import { listingsApi } from '$lib/api/listings';

function createHostListingsStore() {
	let items = $state<Listing[]>([]);
	let isLoading = $state(false);
	let isInitialized = $state(false);

	return {
		async refresh() {
			if (!browser) {
				items = [];
				return;
			}
			isLoading = true;
			try {
				const raw = await listingsApi.getMyListings();
				items = (raw || []).map((l) => apiListingToCardListing(l));
			} catch (e) {
				console.warn('Failed to fetch host listings from API:', e);
			} finally {
				isLoading = false;
				isInitialized = true;
			}
		},

		async initialize() {
			if (!browser) return;
			await this.refresh();
		},

		add(listing: Listing) {
			items = [normalizeListing(listing), ...items];
		},

		update(id: string, updates: Partial<Listing>) {
			items = items.map((item) =>
				item.id === id
					? normalizeListing({
							...item,
							...updates,
							rules: updates.rules ? { ...item.rules, ...updates.rules } : item.rules,
							updatedAt: new Date().toISOString()
						})
					: item
			);
		},

		async remove(id: string) {
			const previous = items;
			items = items.filter((item) => item.id !== id);
			try {
				await listingsApi.deleteListing(id);
			} catch (e) {
				items = previous;
				console.error('Failed to delete listing on server:', e);
				throw e;
			}
		},

		async archive(id: string) {
			const updated = await listingsApi.archiveListing(id);
			items = items.map((item) => (item.id === id ? apiListingToCardListing(updated) : item));
			return updated;
		},

		async unarchive(id: string) {
			const updated = await listingsApi.unarchiveListing(id);
			items = items.map((item) => (item.id === id ? apiListingToCardListing(updated) : item));
			return updated;
		},

		getById(id: string): Listing | undefined {
			const listing = items.find((item) => item.id === id);
			return listing ? normalizeListing(listing) : undefined;
		},

		get items() {
			return items;
		},

		get isLoading() {
			return isLoading;
		},

		get isInitialized() {
			return isInitialized;
		}
	};
}

export const hostListingsStore = createHostListingsStore();
