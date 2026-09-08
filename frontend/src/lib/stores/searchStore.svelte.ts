import {
	DEFAULT_SEARCH_PARAMS,
	DEFAULT_FILTERS,
	type CityOption,
	type SearchParams,
	type SearchFilters
} from '$lib/components/search/taxonomy';
import type { Listing } from '$lib/components/card/types';

function createSearchStore() {
	type ActiveField = 'location' | 'dates' | 'propertyType' | 'guests' | null;
	let draft = $state<SearchParams>({ ...DEFAULT_SEARCH_PARAMS });
	let draftFilters = $state<SearchFilters>({ ...DEFAULT_FILTERS });
	let committed = $state<SearchParams>({ ...DEFAULT_SEARCH_PARAMS });
	let committedFilters = $state<SearchFilters>({ ...DEFAULT_FILTERS });
	let activeField = $state<ActiveField>(null);
	let activeFieldOwner = $state<string | null>(null);
	let resultsCount = $state(0);
	let availableCities = $state<CityOption[]>([]);
	let popularCities = $state<CityOption[]>([]);
	let listings = $state<Listing[]>([]);

	return {
		get draft() {
			return draft;
		},
		get draftFilters() {
			return draftFilters;
		},
		get params() {
			return committed;
		},
		get filters() {
			return committedFilters;
		},
		get activeField() {
			return activeField;
		},
		get activeFieldOwner() {
			return activeFieldOwner;
		},
		get resultsCount() {
			return resultsCount;
		},
		get availableCities() {
			return availableCities;
		},
		get popularCities() {
			return popularCities;
		},
		get listings() {
			return listings;
		},

		setDraft(updates: Partial<SearchParams>) {
			draft = { ...draft, ...updates };
		},
		setDraftFilters(updates: Partial<SearchFilters>) {
			draftFilters = { ...draftFilters, ...updates };
		},
		commit() {
			committed = { ...draft };
			committedFilters = { ...draftFilters };
			activeField = null;
			activeFieldOwner = null;
		},
		commitFilters(filters: SearchFilters) {
			committedFilters = { ...filters };
			draftFilters = { ...filters };
		},
		resetAll() {
			draft = { ...DEFAULT_SEARCH_PARAMS };
			draftFilters = { ...DEFAULT_FILTERS };
			committed = { ...DEFAULT_SEARCH_PARAMS };
			committedFilters = { ...DEFAULT_FILTERS };
			activeField = null;
			activeFieldOwner = null;
			resultsCount = 0;
		},
		setActiveField(field: ActiveField, ownerId?: string) {
			activeField = field;
			activeFieldOwner = field ? (ownerId ?? null) : null;
		},
		closeDropdowns(ownerId?: string) {
			if (ownerId && activeFieldOwner && ownerId !== activeFieldOwner) return;
			activeField = null;
			activeFieldOwner = null;
		},
		setResultsCount(count: number) {
			resultsCount = count;
		},
		setAvailableCities(cities: CityOption[]) {
			availableCities = cities;
		},
		setPopularCities(cities: CityOption[]) {
			popularCities = cities;
		},
		setListings(nextListings: Listing[]) {
			listings = nextListings;
		}
	};
}

export const searchStore = createSearchStore();
