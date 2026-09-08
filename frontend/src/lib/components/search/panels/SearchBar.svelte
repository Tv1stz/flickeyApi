<script lang="ts">
	import { searchStore } from '$lib/stores/searchStore.svelte';
	import LocationField from '../fields/LocationField.svelte';
	import DatesField from '../fields/DatesField.svelte';
	import PropertyTypeField from '../fields/PropertyTypeField.svelte';
	import GuestsField from '../fields/GuestsField.svelte';
	import FiltersModal from '../FiltersModal.svelte';
	import {
		countActiveFilters,
		getCityOptionsFromListings,
		getPopularCitiesFromListings,
		isSearchComplete,
		type SearchParams
	} from '../taxonomy';
	import type { Listing, PropertyType } from '$lib/components/card/types';
	import { SlidersHorizontal } from 'lucide-svelte';
	import { saveRecentSearch } from '$lib/services/localStorage';

	// ---------------------------------------------------------------------------
	// Props
	// ---------------------------------------------------------------------------
	interface Props {
		compact?: boolean;
		filteredCount?: number;
		onSearch?: () => void;
		instanceId?: string;
		listings?: Listing[];
		onExpand?: () => void;
		expandOnCompactClick?: boolean;
		onFiltersOpen?: () => void;
		onFiltersClose?: () => void;
		alwaysShowFilters?: boolean;
	}

	let {
		compact = false,
		filteredCount,
		onSearch,
		instanceId,
		listings,
		onExpand,
		expandOnCompactClick = false,
		onFiltersOpen,
		onFiltersClose,
		alwaysShowFilters = false
	}: Props = $props();
	const resolvedInstanceId = $derived(instanceId ?? (compact ? 'searchbar-compact' : 'searchbar'));
	const allowCompactExpand = $derived(Boolean(onExpand && expandOnCompactClick && compact));

	// ---------------------------------------------------------------------------
	// UI state
	// ---------------------------------------------------------------------------
	let filtersOpen = $state(false);
	let hoveredField: 'location' | 'dates' | 'propertyType' | 'guests' | null = $state(null);

	// Ширина круглой подложки кнопки поиска (px) — компенсация при расчёте hover-пилюли.
	const SEARCH_BTN_EXPAND_PX = $derived(compact ? 40 : 48);

	// ---------------------------------------------------------------------------
	// DOM refs — обычные переменные, не $state (bind:this этого не требует)
	// ---------------------------------------------------------------------------
	let barEl: HTMLDivElement | undefined;
	let locationEl: HTMLDivElement | undefined;
	let datesEl: HTMLDivElement | undefined;
	let propertyTypeEl: HTMLDivElement | undefined;
	let guestsEl: HTMLDivElement | undefined;

	// ---------------------------------------------------------------------------
	// Store-derived values
	// ---------------------------------------------------------------------------
	const activeField = $derived(
		searchStore.activeFieldOwner === resolvedInstanceId ? searchStore.activeField : null
	);
	const draft = $derived(searchStore.draft);
	const committedParams = $derived(searchStore.params);
	const committedFilters = $derived(searchStore.filters);
	const activeFiltersCount = $derived(countActiveFilters(committedFilters));
	const hasSearch = $derived(isSearchComplete(committedParams));
	const canApplyFilters = $derived(isSearchComplete(draft));

	const availableCities = $derived.by(() => {
		if (listings && listings.length > 0) {
			return getCityOptionsFromListings(listings);
		}
		return searchStore.availableCities;
	});
	const popularCities = $derived.by(() => {
		if (listings && listings.length > 0) {
			return getPopularCitiesFromListings(listings);
		}
		return searchStore.popularCities.length > 0
			? searchStore.popularCities
			: availableCities.slice(0, 6);
	});

	const DEFAULT_LOCATION = $derived.by(() => availableCities[0]?.label ?? '');
	const DEFAULT_ADULTS = 2;

	const hideFirstDivider = $derived(
		activeField === 'location' ||
			activeField === 'dates' ||
			hoveredField === 'location' ||
			hoveredField === 'dates'
	);

	const hideSecondDivider = $derived(
		activeField === 'dates' ||
			activeField === 'propertyType' ||
			hoveredField === 'dates' ||
			hoveredField === 'propertyType'
	);

	const hideThirdDivider = $derived(
		activeField === 'propertyType' ||
			activeField === 'guests' ||
			hoveredField === 'propertyType' ||
			hoveredField === 'guests'
	);

	// Сколько px справа занимает кнопка поиска (absolute right-2 = 8px gap + ширина кнопки)
	// При active-состоянии кнопка расширяется, поэтому резервируем больше
	const guestsReserveRight = $derived(
		activeField ? (compact ? 108 : 136) : (compact ? 48 : 58)
	);

	$effect(() => {
		if (!hasSearch && !alwaysShowFilters && filtersOpen) closeFilters();
	});

	// ---------------------------------------------------------------------------
	// Pill animation state
	// ---------------------------------------------------------------------------
	interface PillState {
		visible: boolean;
		left: number;
		width: number;
		transition: string;
	}

	const TRANSITION_FULL = 'transition: all 300ms cubic-bezier(0.2,0,0,1);';
	const TRANSITION_OPACITY = 'transition: opacity 300ms cubic-bezier(0.2,0,0,1);';

	let activePill = $state<PillState>({
		visible: false,
		left: 0,
		width: 0,
		transition: TRANSITION_FULL
	});
	let hoverPill = $state<PillState>({
		visible: false,
		left: 0,
		width: 0,
		transition: TRANSITION_FULL
	});

	let locationDropdownOffset = $state(0);
	let datesDropdownOffset = $state(0);
	let guestsDropdownOffset = $state(0);
	let locationOffsetLocked = $state(false);
	let guestsOffsetLocked = $state(false);

	const activePillStyle = $derived(
		`opacity:${activePill.visible ? 1 : 0};transform:translate3d(${activePill.left}px,0,0);width:${activePill.width}px;${activePill.transition}`
	);

	const hoverPillStyle = $derived(
		`opacity:${hoverPill.visible ? 1 : 0};transform:translate3d(${hoverPill.left}px,0,0);width:${hoverPill.width}px;${hoverPill.transition}`
	);

	// ---------------------------------------------------------------------------
	// Layout-stability flag (не реактивна — используется только в updatePills)
	// ---------------------------------------------------------------------------
	let isLayoutStable = true;
	let prevActiveField: typeof activeField = null;

	// ---------------------------------------------------------------------------
	// Helpers
	// ---------------------------------------------------------------------------
	function getFieldEl(field: typeof activeField): HTMLDivElement | undefined {
		if (!field) return undefined;
		const map: Record<NonNullable<typeof activeField>, HTMLDivElement | undefined> = {
			location: locationEl,
			dates: datesEl,
			propertyType: propertyTypeEl,
			guests: guestsEl
		};
		return map[field];
	}

	function getMetrics(node: HTMLDivElement | undefined): { left: number; width: number } | null {
		if (!node || !barEl) return null;
		const barRect = barEl.getBoundingClientRect();
		const nodeRect = node.getBoundingClientRect();
		return { left: nodeRect.left - barRect.left, width: nodeRect.width };
	}

	function updatePills(): void {
		if (!barEl) return;

		const baseTransition = isLayoutStable ? TRANSITION_FULL : TRANSITION_OPACITY;
		const barWidth = barEl.getBoundingClientRect().width;

		// --- Активная пилюля ---
		const activeNode = getFieldEl(activeField);
		const rawActiveMetrics = getMetrics(activeNode);

		if (rawActiveMetrics) {
			let { left, width } = rawActiveMetrics;

			// МАГИЯ ДЛЯ ЛЕВОГО КРАЯ (Где)
			if (activeField === 'location') {
				// Прибавляем отступ (1px) к ширине, чтобы правый край остался на месте,
				// и жестко прижимаем пилюлю к левому краю.
				width = width + left;
				left = 0;
			}
			// МАГИЯ ДЛЯ ПРАВОГО КРАЯ (Кто)
			else if (activeField === 'guests') {
				width = barWidth - left - 2;
			}

			activePill = { visible: true, left, width, transition: baseTransition };
		} else {
			activePill = { ...activePill, visible: false, transition: baseTransition };
		}

		// --- Hover-пилюля ---
		const hoverNode = getFieldEl(hoveredField);
		const rawHoverMetrics = getMetrics(hoverNode);

		if (rawHoverMetrics && hoverNode !== activeNode) {
			let { left, width } = rawHoverMetrics;

			if (hoveredField === 'location') {
				width = width + left;
				left = 0;
			} else if (hoveredField === 'guests') {
				width = barWidth - left - 2;
			}

			hoverPill = { visible: true, left, width, transition: baseTransition };
		} else {
			hoverPill = { ...hoverPill, visible: false, transition: baseTransition };
		}

		const barRect = barEl.getBoundingClientRect();
		if (locationEl) {
			const locationRect = locationEl.getBoundingClientRect();
			if (!locationOffsetLocked || isLayoutStable || activeField !== 'location') {
				locationDropdownOffset = Math.max(0, Math.round(locationRect.left - barRect.left));
				if (activeField === 'location' && !isLayoutStable) locationOffsetLocked = true;
			}
		}
		if (datesEl) {
			const datesRect = datesEl.getBoundingClientRect();
			datesDropdownOffset = Math.max(-120, Math.round(datesRect.left - barRect.left - 100));
		}
		if (guestsEl) {
			const guestsRect = guestsEl.getBoundingClientRect();
			if (!guestsOffsetLocked || isLayoutStable || activeField !== 'guests') {
				guestsDropdownOffset = Math.max(0, Math.round(barRect.right - guestsRect.right));
				if (activeField === 'guests' && !isLayoutStable) guestsOffsetLocked = true;
			}
		}
	}

	// ---------------------------------------------------------------------------
	// Effect 1: управление isLayoutStable при открытии/закрытии панели
	// ---------------------------------------------------------------------------
	let stableTimeoutId: ReturnType<typeof setTimeout> | undefined;

	$effect(() => {
		const currentOpen = !!activeField;
		const prevOpen = !!prevActiveField;

		if (currentOpen !== prevOpen) {
			isLayoutStable = false;
			clearTimeout(stableTimeoutId);
			stableTimeoutId = setTimeout(() => {
				isLayoutStable = true;
				locationOffsetLocked = false;
				guestsOffsetLocked = false;
				updatePills();
			}, 320);
		}

		if (activeField !== prevActiveField) {
			if (activeField === 'location') locationOffsetLocked = false;
			if (activeField === 'guests') guestsOffsetLocked = false;
		}

		prevActiveField = activeField;

		return () => clearTimeout(stableTimeoutId);
	});

	// ---------------------------------------------------------------------------
	// Effect 2: ResizeObserver — единственное место синхронного updatePills()
	// ---------------------------------------------------------------------------
	$effect(() => {
		if (!barEl) return;

		let frameId = 0;

		const ro = new ResizeObserver(() => {
			cancelAnimationFrame(frameId);
			frameId = requestAnimationFrame(updatePills);
		});

		ro.observe(barEl);
		if (locationEl) ro.observe(locationEl);
		if (propertyTypeEl) ro.observe(propertyTypeEl);
		if (guestsEl) ro.observe(guestsEl);

		void activeField;
		void hoveredField;
		void compact;

		cancelAnimationFrame(frameId);
		frameId = requestAnimationFrame(updatePills);

		return () => {
			ro.disconnect();
			cancelAnimationFrame(frameId);
		};
	});

	// ---------------------------------------------------------------------------
	// Event handlers
	// ---------------------------------------------------------------------------
	function handleSearch(e: MouseEvent): void {
		e.preventDefault();
		e.stopPropagation();
		searchStore.commit();
		searchStore.closeDropdowns(resolvedInstanceId);
		saveRecentSearch(searchStore.params, searchStore.filters);
		onSearch?.();
	}

	function handleBarClick(e: MouseEvent): void {
		e.stopPropagation();
	}

	function handleBarKeydown(e: KeyboardEvent): void {
		e.stopPropagation();
	}

	function handleWindowClick(e: MouseEvent): void {
		const target = e.target as Element | null;
		if (target?.closest('[data-filters-modal]')) return;
		searchStore.closeDropdowns(resolvedInstanceId);
	}

	function openFilters() {
		filtersOpen = true;
		onFiltersOpen?.();
	}

	function closeFilters() {
		filtersOpen = false;
		onFiltersClose?.();
	}

	function handleExpandClick(e: MouseEvent): void {
		e.preventDefault();
		e.stopPropagation();
		onExpand?.();
	}

	// Hover через pointer-события — доступны и для мыши, и не мешают клавиатуре.
	// onmouseenter/leave оставляем только для визуального hover-эффекта пилюли,
	// который по природе своей мышиный. Предупреждение Svelte про ARIA-роль
	// снимается добавлением role="presentation" на обёртки-пилюли,
	// а сами поля уже имеют интерактивные элементы внутри.
</script>

<svelte:window onclick={handleWindowClick} />

<div class="relative w-full" role="search">
	<!--
        barEl — role="presentation", т.к. интерактивность обеспечивается дочерними элементами.
        onclick нужен только для stopPropagation, поэтому добавляем onkeydown-аналог.
    -->
	<div class="flex items-center {compact ? 'gap-2' : 'gap-3'}">
		<div class="relative flex-1">
			<div
				bind:this={barEl}
				role="presentation"
				onclick={handleBarClick}
				onkeydown={handleBarKeydown}
				class="relative flex w-full items-center rounded-full transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)]
	                   {compact ? 'h-[52px]' : 'h-[66px]'}
	                   {activeField
					? 'border border-zinc-300/80 bg-[#ebebeb]'
					: 'border border-zinc-200/80 bg-white shadow-[0_3px_12px_rgba(0,0,0,0.06)] hover:shadow-[0_4px_16px_rgba(0,0,0,0.08)]'}"
			>
				<!-- Hover-пилюля -->
				<div
					class="pointer-events-none absolute inset-y-0 left-0 z-0 rounded-full will-change-[transform,width]
                   {activeField ? 'bg-[#dddddd]' : 'bg-zinc-100/80'}"
					style={hoverPillStyle}
				></div>

				<!-- Активная пилюля -->
				<div
					class="pointer-events-none absolute inset-y-0 left-0 z-10 rounded-full bg-white shadow-[0_6px_20px_rgba(0,0,0,0.08)] will-change-[transform,width]"
					style={activePillStyle}
				></div>

				<!-- 1. Location -->
				<div
					bind:this={locationEl}
					role="group"
					onmouseenter={() => (hoveredField = 'location')}
					onmouseleave={() => (hoveredField = null)}
					class="relative flex h-full min-w-0 flex-[1.2] items-center rounded-full
                   transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)]
                   {activeField === 'location' ? 'z-20' : 'z-10 cursor-pointer'}"
				>
					<div class="w-full {compact ? 'pl-1.5' : 'pl-0'}">
						<LocationField
							value={draft.location}
							isOpen={activeField === 'location'}
							cities={availableCities}
							{popularCities}
							dropdownOffsetLeft={locationDropdownOffset}
							onSelect={(v) => searchStore.setDraft({ location: v })}
							onComplete={() => searchStore.setActiveField('dates', resolvedInstanceId)}
							onFocus={() => searchStore.setActiveField('location', resolvedInstanceId)}
							{compact}
						/>
					</div>
				</div>

				<!-- Divider 1 -->
				<div
					class="relative z-20 w-px shrink-0 bg-zinc-300 transition-opacity duration-300 ease-[cubic-bezier(0.2,0,0,1)]
                   {compact ? 'h-6' : 'h-8'}
                   {hideFirstDivider ? 'opacity-0' : 'opacity-100'}"
					aria-hidden="true"
				></div>

				<!-- 2. Dates -->
				<div
					bind:this={datesEl}
					role="group"
					onmouseenter={() => (hoveredField = 'dates')}
					onmouseleave={() => (hoveredField = null)}
					class="relative flex h-full flex-[1.2] items-center rounded-full transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)] {activeField ===
					'dates'
						? 'z-20'
						: 'z-10 cursor-pointer'} {compact ? 'min-w-[95px]' : 'min-w-[135px]'}"
				>
					<div class="w-full">
						<DatesField
							checkin={draft.checkin ?? null}
							checkout={draft.checkout ?? null}
							isOpen={activeField === 'dates'}
							dropdownOffsetLeft={datesDropdownOffset}
							onSelect={(ci, co) => searchStore.setDraft({ checkin: ci, checkout: co })}
							onComplete={() => searchStore.setActiveField('propertyType', resolvedInstanceId)}
							onFocus={() => searchStore.setActiveField('dates', resolvedInstanceId)}
							{compact}
						/>
					</div>
				</div>

				<!-- Divider 2 -->
				<div
					class="relative z-20 w-px shrink-0 bg-zinc-300 transition-opacity duration-300 ease-[cubic-bezier(0.2,0,0,1)]
                   {compact ? 'h-6' : 'h-8'}
                   {hideSecondDivider ? 'opacity-0' : 'opacity-100'}"
					aria-hidden="true"
				></div>

				<!-- 3. Property Type -->
				<div
					bind:this={propertyTypeEl}
					role="group"
					onmouseenter={() => (hoveredField = 'propertyType')}
					onmouseleave={() => (hoveredField = null)}
					class="relative flex h-full flex-1 items-center rounded-full transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)] {activeField ===
					'propertyType'
						? 'z-20'
						: 'z-10 cursor-pointer'} {compact ? 'min-w-[75px]' : 'min-w-[135px]'}"
				>
					<div class="w-full">
						<PropertyTypeField
							value={draft.propertyType}
							isOpen={activeField === 'propertyType'}
							onSelect={(v) => {
								searchStore.setDraft({ propertyType: v as PropertyType | 'any' });
								searchStore.setActiveField('guests', resolvedInstanceId);
							}}
							onFocus={() => searchStore.setActiveField('propertyType', resolvedInstanceId)}
							{compact}
						/>
					</div>
				</div>

				<!-- Divider 3 -->
				<div
					class="relative z-20 w-px shrink-0 bg-zinc-300 transition-opacity duration-300 ease-[cubic-bezier(0.2,0,0,1)]
                   {compact ? 'h-6' : 'h-8'}
                   {hideThirdDivider ? 'opacity-0' : 'opacity-100'}"
					aria-hidden="true"
				></div>

				<!-- 4. Guests -->
				<div
					bind:this={guestsEl}
					role="group"
					onmouseenter={() => (hoveredField = 'guests')}
					onmouseleave={() => (hoveredField = null)}
					class="relative flex h-full flex-1 items-center rounded-full
                   transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)]
                   {activeField === 'guests' ? 'z-20' : 'z-10 cursor-pointer'}
                   {compact ? (activeField ? 'min-w-[155px]' : 'min-w-[110px]') : (activeField ? 'min-w-[240px]' : 'min-w-[170px]')}"
				>
					<div class="w-full">
						<GuestsField
							adults={draft.adults}
							children={draft.children}
							isOpen={activeField === 'guests'}
							reserveRight={guestsReserveRight}
							dropdownOffsetRight={guestsDropdownOffset}
							onSelect={(adults, children) => searchStore.setDraft({ adults, children })}
							onFocus={() => searchStore.setActiveField('guests', resolvedInstanceId)}
							{compact}
						/>
					</div>
				</div>

				<!-- Search button (absolute, no layout shift) -->
				<button
					type="button"
					onclick={handleSearch}
					class="group absolute top-1/2 right-2 z-20 -translate-y-1/2 overflow-hidden rounded-full text-white
                       transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)]
                       active:scale-[0.96]
                       {compact
							? activeField ? 'h-[40px] w-[96px]' : 'h-[40px] w-[40px]'
							: activeField ? 'h-[48px] w-[120px]' : 'h-[48px] w-[48px]'}"
					aria-label="Найти"
				>
					<span
						class="absolute inset-y-0 right-0 rounded-full bg-zinc-900 shadow-sm transition-[width,background-color,box-shadow]
                           duration-300 ease-[cubic-bezier(0.2,0,0,1)] group-hover:bg-zinc-800 group-hover:shadow-md
                           {activeField ? 'w-full' : compact ? 'w-[40px]' : 'w-[48px]'}"
						aria-hidden="true"
					></span>
					<span
						class="absolute inset-y-0 right-0 z-10 flex items-center justify-center
                           {compact ? 'w-[40px]' : 'w-[48px]'}"
						aria-hidden="true"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 32 32"
							aria-hidden="true"
							class="block {compact ? 'h-3.5 w-3.5' : 'h-4 w-4'}"
							style="fill:none;stroke:currentcolor;stroke-width:4;overflow:visible"
						>
							<path fill="none" d="M13 24a11 11 0 1 0 0-22 11 11 0 0 0 0 22zm8-3 9 9"></path>
						</svg>
					</span>

					<div
						class="pointer-events-none absolute inset-y-0 right-0 z-10 flex items-center overflow-hidden
                           transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)]
                           {activeField
							? compact
								? 'pr-9 pl-3.5 opacity-100'
								: 'pr-11 pl-4 opacity-100'
							: 'pr-0 pl-0 opacity-0'}"
						aria-hidden="true"
					>
						<span
							class="pt-[1px] text-[15px] leading-none font-bold tracking-wide whitespace-nowrap {compact
								? 'text-[12px]'
								: 'text-[15px]'}">Искать</span
						>
					</div>
				</button>
			</div>

			{#if allowCompactExpand}
				<button
					type="button"
					onclick={handleExpandClick}
					class="absolute inset-0 z-30 rounded-full"
					aria-label="Открыть поиск"
				></button>
			{/if}
		</div>

		{#if hasSearch || alwaysShowFilters}
			<button
				type="button"
				onclick={openFilters}
				class="flex shrink-0 items-center gap-2 rounded-full border border-zinc-200/80 bg-white font-semibold text-zinc-900 shadow-sm transition-all duration-300 hover:border-zinc-300 hover:shadow-md active:scale-[0.98]
	                   {compact ? 'h-[48px] px-3.5 text-[13px]' : 'h-[66px] px-5 text-[14px]'}"
				aria-label="Фильтры"
			>
				<SlidersHorizontal size={18} strokeWidth={2} class="text-zinc-700" />
				{#if !compact}
					<span>Фильтры</span>
				{/if}
				{#if activeFiltersCount > 0}
					<span
						class="ml-1 flex h-[22px] min-w-[22px] items-center justify-center rounded-full bg-zinc-900 px-1.5 text-[11px] font-bold text-white shadow-inner"
						aria-label="{activeFiltersCount} активных фильтра"
					>
						{activeFiltersCount}
					</span>
				{/if}
			</button>
		{/if}
	</div>
</div>

<div data-filters-modal>
	<FiltersModal
		open={filtersOpen}
		filters={searchStore.draftFilters}
		resultsCount={filteredCount ?? 0}
		{listings}
		params={committedParams}
		canApply={true}
		onApply={(f) => searchStore.commitFilters(f)}
		onClose={closeFilters}
	/>
</div>
