<script lang="ts">
	import { onMount } from 'svelte';
	import SearchBar from './SearchBar.svelte';
	import { searchStore } from '$lib/stores/searchStore.svelte';
	import { headerSearchStore } from '$lib/stores/headerSearchStore.svelte';

	interface Props {
		enabled?: boolean;
		filteredCount?: number;
		onSearch?: () => void;
		compactByDefault?: boolean;
		alwaysShowFilters?: boolean;
	}

	let { enabled = false, filteredCount = 0, onSearch, compactByDefault = false, alwaysShowFilters = false }: Props = $props();

	const HEADER_ID = 'header-search';

	const COMPACT_AT = 88;
	const EXPAND_AT = 40;
	const COMPACT_SHIFT_PX = -71;

	const isHeaderActive = $derived(searchStore.activeFieldOwner === HEADER_ID);
	let filtersOpen = $state(false);
	const isCompact = $derived(enabled && headerSearchStore.isCompact);
	const renderCompact = $derived(
		enabled && (compactByDefault ? !(isHeaderActive || headerSearchStore.isLocked) : isCompact)
	);
	const renderExpanded = $derived(enabled && !renderCompact);

	function openExpanded(e?: MouseEvent) {
		e?.preventDefault();
		e?.stopPropagation();
		if (!enabled) return;
		headerSearchStore.setExpanded(true);
		searchStore.setActiveField('location', HEADER_ID);
	}

	function syncFromScroll(scrollY: number) {
		if (!enabled) return;
		if (filtersOpen) return;
		if (headerSearchStore.isLocked || isHeaderActive) return;

		if (compactByDefault) {
			if (!headerSearchStore.isCompact) {
				searchStore.closeDropdowns(HEADER_ID);
				headerSearchStore.setCompact();
			}
			return;
		}

		if (scrollY > COMPACT_AT && headerSearchStore.isExpanded) {
			searchStore.closeDropdowns(HEADER_ID);
			headerSearchStore.setCompact();
			return;
		}

		if (scrollY < EXPAND_AT && headerSearchStore.isCompact) {
			headerSearchStore.setExpanded(false);
		}
	}

	$effect(() => {
		if (!enabled) {
			headerSearchStore.reset();
			searchStore.closeDropdowns(HEADER_ID);
			return;
		}
		if (typeof window === 'undefined') return;
		syncFromScroll(window.scrollY);
	});

	$effect(() => {
		if (!enabled) return;
		if (isHeaderActive) {
			headerSearchStore.setExpanded(true);
			return;
		}
		if (headerSearchStore.isLocked && !filtersOpen) {
			headerSearchStore.setLocked(false);
			if (typeof window !== 'undefined') syncFromScroll(window.scrollY);
		}
	});

	function handleFiltersOpen() {
		filtersOpen = true;
	}

	function handleFiltersClose() {
		filtersOpen = false;
		if (headerSearchStore.isLocked) {
			headerSearchStore.setLocked(false);
		}
		if (typeof window !== 'undefined') {
			syncFromScroll(window.scrollY);
		}
	}

	onMount(() => {
		if (typeof window === 'undefined') return;

		const media = window.matchMedia('(min-width: 1024px)');
		const handleViewport = () => {
			if (!enabled) return;
			if (!media.matches) {
				headerSearchStore.setCompact();
				return;
			}
			syncFromScroll(window.scrollY);
		};

		handleViewport();

		let ticking = false;
		const handleScroll = () => {
			if (!media.matches) return;
			if (ticking) return;
			ticking = true;
			requestAnimationFrame(() => {
				ticking = false;
				syncFromScroll(window.scrollY);
			});
		};

		window.addEventListener('scroll', handleScroll, { passive: true });
		media.addEventListener('change', handleViewport);

		return () => {
			window.removeEventListener('scroll', handleScroll);
			media.removeEventListener('change', handleViewport);
		};
	});
</script>

<div class="contents">
	<!-- Morphing bar (single instance) -->
	<div
		class="col-span-3 row-start-2 flex w-full justify-center transition-[max-height,padding] duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]
			{renderExpanded ? 'max-h-[140px] pt-1 pb-4' : 'max-h-0 pt-0 pb-0'}"
	>
		<div
			class="w-full transition-[transform,filter,max-width] duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] will-change-[transform]
				{renderExpanded ? 'max-w-[860px]' : 'max-w-[min(670px,calc(100vw-420px))]'} origin-center"
			style={`transform: translate3d(0, ${renderCompact ? `${COMPACT_SHIFT_PX}px` : '0px'}, 0);`}
		>
			<SearchBar
				compact={renderCompact}
				instanceId={HEADER_ID}
				{filteredCount}
				listings={searchStore.listings}
				onExpand={openExpanded}
				expandOnCompactClick
				onFiltersOpen={handleFiltersOpen}
				onFiltersClose={handleFiltersClose}
				{onSearch}
				{alwaysShowFilters}
			/>
		</div>
	</div>
</div>
