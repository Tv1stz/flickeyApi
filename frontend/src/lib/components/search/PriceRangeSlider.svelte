<script lang="ts">
	import { onDestroy } from 'svelte';
	import { getPriceDistribution, type PriceDistribution } from './taxonomy';
	import type { Listing } from '$lib/components/card/types';

	interface Props {
		min: number;
		max: number;
		valueMin: number | null;
		valueMax: number | null;
		listings?: Listing[];
		onChangeMin: (value: number | null) => void;
		onChangeMax: (value: number | null) => void;
	}

	let { min, max, valueMin, valueMax, listings, onChangeMin, onChangeMax }: Props = $props();

	type DragTarget = 'min' | 'max' | 'range' | null;

	let sliderRef = $state<HTMLDivElement | null>(null);
	let captureElement = $state<HTMLElement | null>(null);
	let activePointerId = $state<number | null>(null);
	let dragTarget = $state<DragTarget>(null);
	let dragStartValue = $state<number | null>(null);
	let dragStartMin = $state<number | null>(null);
	let dragStartMax = $state<number | null>(null);
	let dragTrackRect = $state<{ left: number; width: number } | null>(null);
	let dragMin = $state<number | null>(null);
	let dragMax = $state<number | null>(null);
	let sliderWidth = $state(0);
	let inputMinValue = $state('');
	let inputMaxValue = $state('');
	let inputMinFocused = $state(false);
	let inputMaxFocused = $state(false);

	const range = $derived(max - min);
	const isValidRange = $derived(Number.isFinite(min) && Number.isFinite(max) && range > 0);
	const STEP = $derived(getPriceStep(range));
	const MIN_GAP = $derived(isValidRange ? Math.min(STEP, range) : 0);

	function getPriceStep(priceRange: number): number {
		if (priceRange >= 5000) return 50;
		if (priceRange >= 1000) return 10;
		if (priceRange >= 500) return 5;
		return 1;
	}

	function clamp(value: number, lower: number, upper: number): number {
		return Math.min(Math.max(value, lower), upper);
	}

	function snapToStep(value: number): number {
		if (!isValidRange) return min;
		return clamp(Math.round((value - min) / STEP) * STEP + min, min, max);
	}

	function normalizeExternalValue(value: number | null, fallback: number): number {
		return value !== null && Number.isFinite(value) ? clamp(value, min, max) : fallback;
	}

	function getInitialValues(): { minValue: number; maxValue: number } {
		if (!isValidRange) return { minValue: min, maxValue: max };

		let minValue = normalizeExternalValue(valueMin, min);
		let maxValue = normalizeExternalValue(valueMax, max);

		if (minValue > maxValue) {
			[minValue, maxValue] = [maxValue, minValue];
		}

		if (maxValue - minValue < MIN_GAP) {
			if (minValue + MIN_GAP <= max) maxValue = minValue + MIN_GAP;
			else minValue = maxValue - MIN_GAP;
		}

		return { minValue, maxValue };
	}

	const initialValues = $derived(getInitialValues());
	const currentMin = $derived(
		dragTarget ? (dragMin ?? initialValues.minValue) : initialValues.minValue
	);
	const currentMax = $derived(
		dragTarget ? (dragMax ?? initialValues.maxValue) : initialValues.maxValue
	);
	const minPercent = $derived(isValidRange ? ((currentMin - min) / range) * 100 : 0);
	const maxPercent = $derived(isValidRange ? ((currentMax - min) / range) * 100 : 100);
	const bubblesOverlap = $derived(
		Math.abs(maxPercent - minPercent) * (Math.max(sliderWidth, 280) / 100) < 112
	);

	$effect(() => {
		if (!inputMinFocused) inputMinValue = currentMin.toString();
		if (!inputMaxFocused) inputMaxValue = currentMax.toString();
	});

	$effect(() => {
		if (!sliderRef) return;
		const updateSliderWidth = () => (sliderWidth = sliderRef?.getBoundingClientRect().width ?? 0);
		updateSliderWidth();
		const observer = new ResizeObserver(updateSliderWidth);
		observer.observe(sliderRef);
		return () => observer.disconnect();
	});

	const HISTOGRAM_BUCKETS = 40;
	const dataPrices = $derived(
		listings?.map((listing) => listing.pricePerNight).filter(Number.isFinite) ?? []
	);
	function getDataDomain(prices: number[]): { min: number; max: number } | null {
		if (prices.length === 0) return null;
		let domainMin = prices[0];
		let domainMax = prices[0];
		for (let index = 1; index < prices.length; index += 1) {
			domainMin = Math.min(domainMin, prices[index]);
			domainMax = Math.max(domainMax, prices[index]);
		}
		return { min: domainMin, max: domainMax };
	}
	const dataDomain = $derived(getDataDomain(dataPrices));

	function getPercentile(sortedValues: number[], percentile: number): number {
		const index = (sortedValues.length - 1) * percentile;
		const lower = Math.floor(index);
		const upper = Math.ceil(index);
		return sortedValues[lower] + (sortedValues[upper] - sortedValues[lower]) * (index - lower);
	}

	function getVisualHistogramDomain(prices: number[]): { min: number; max: number } | null {
		if (!isValidRange || prices.length === 0) return null;
		if (prices.length < 4) return { min, max };

		const sortedPrices = [...prices].sort((a, b) => a - b);
		const firstQuartile = getPercentile(sortedPrices, 0.25);
		const thirdQuartile = getPercentile(sortedPrices, 0.75);
		const interquartileRange = thirdQuartile - firstQuartile;
		if (interquartileRange <= 0) return { min, max };

		const lowerWhisker = clamp(firstQuartile - interquartileRange * 1.5, min, max);
		const upperWhisker = clamp(thirdQuartile + interquartileRange * 1.5, min, max);
		const whiskerRange = upperWhisker - lowerWhisker;
		if (whiskerRange <= 0) return { min, max };

		const padding = Math.max(STEP, whiskerRange * 0.08);
		return {
			min: clamp(lowerWhisker - padding, min, max),
			max: clamp(upperWhisker + padding, min, max)
		};
	}

	const visualHistogramDomain = $derived(dataDomain ? getVisualHistogramDomain(dataPrices) : null);
	const distribution = $derived<PriceDistribution[]>(
		listings && visualHistogramDomain
			? getPriceDistribution(
					listings,
					visualHistogramDomain.min,
					visualHistogramDomain.max,
					HISTOGRAM_BUCKETS
				)
			: []
	);

	const histogramGeometry = $derived.by(() => {
		if (!visualHistogramDomain || distribution.length === 0) return [];
		const maxCount = Math.max(...distribution.map((bucket) => bucket.count), 1);
		const bucketWidth = 100 / distribution.length;
		const visualRange = visualHistogramDomain.max - visualHistogramDomain.min;

		return distribution.map((bucket, index) => ({
			count: bucket.count,
			x: index * bucketWidth,
			width: bucketWidth,
			height: bucket.count === 0 ? 0 : Math.max((bucket.count / maxCount) * 100, 3),
			dataMin:
				index === 0 ? min : visualHistogramDomain.min + (index / distribution.length) * visualRange,
			dataMax:
				index === distribution.length - 1
					? max
					: visualHistogramDomain.min + ((index + 1) / distribution.length) * visualRange
		}));
	});

	const histogramBars = $derived(
		histogramGeometry.map((bar) => {
			const overlap = Math.max(
				0,
				Math.min(currentMax, bar.dataMax) - Math.max(currentMin, bar.dataMin)
			);
			const isActive =
				overlap > 0 ||
				(currentMin === currentMax && currentMin >= bar.dataMin && currentMin <= bar.dataMax);
			const isEdge = isActive && (currentMin > bar.dataMin || currentMax < bar.dataMax);
			return { ...bar, isActive, isEdge };
		})
	);

	function getSliderRect(): { left: number; width: number } | null {
		if (!sliderRef) return null;
		const { left, width } = sliderRef.getBoundingClientRect();
		return width > 0 ? { left, width } : null;
	}

	function valueFromPointer(
		event: PointerEvent,
		trackRect: { left: number; width: number } | null = getSliderRect()
	): number {
		if (!isValidRange || !trackRect) return min;
		const position = clamp((event.clientX - trackRect.left) / trackRect.width, 0, 1);
		return snapToStep(min + position * range);
	}

	function constrainMin(value: number, maxValue: number): number {
		return clamp(snapToStep(value), min, maxValue - MIN_GAP);
	}

	function constrainMax(value: number, minValue: number): number {
		return clamp(snapToStep(value), minValue + MIN_GAP, max);
	}

	function emitValues(nextMin: number, nextMax: number, keepDragValues = false) {
		const previousMin = currentMin;
		const previousMax = currentMax;
		if (nextMin === previousMin && nextMax === previousMax) return;

		if (keepDragValues) {
			dragMin = nextMin;
			dragMax = nextMax;
		}
		if (nextMin !== previousMin) onChangeMin(nextMin === min ? null : nextMin);
		if (nextMax !== previousMax) onChangeMax(nextMax === max ? null : nextMax);
	}

	function startDrag(target: Exclude<DragTarget, null>, event: PointerEvent) {
		if (
			!isValidRange ||
			activePointerId !== null ||
			!event.isPrimary ||
			(event.pointerType === 'mouse' && event.button !== 0)
		)
			return;
		event.preventDefault();
		event.stopPropagation();
		const element = event.currentTarget as HTMLElement;
		element.setPointerCapture(event.pointerId);
		if (target !== 'range' && element instanceof HTMLButtonElement) {
			element.focus({ preventScroll: true });
		}
		captureElement = element;
		activePointerId = event.pointerId;
		dragTarget = target;
		dragStartValue = valueFromPointer(event);
		dragStartMin = currentMin;
		dragStartMax = currentMax;
		dragMin = currentMin;
		dragMax = currentMax;
	}

	function handleThumbPointerDown(target: 'min' | 'max', event: PointerEvent) {
		startDrag(target, event);
	}

	function handleRangePointerDown(event: PointerEvent) {
		startDrag('range', event);
	}

	function handleTrackPointerDown(event: PointerEvent) {
		if (
			!isValidRange ||
			activePointerId !== null ||
			!event.isPrimary ||
			(event.pointerType === 'mouse' && event.button !== 0)
		)
			return;
		const value = valueFromPointer(event);
		if (Math.abs(value - currentMin) <= Math.abs(value - currentMax)) {
			emitValues(constrainMin(value, currentMax), currentMax);
		} else {
			emitValues(currentMin, constrainMax(value, currentMin));
		}
	}

	function handlePointerMove(event: PointerEvent) {
		if (
			!isValidRange ||
			event.pointerId !== activePointerId ||
			!dragTarget ||
			dragStartValue === null ||
			dragStartMin === null ||
			dragStartMax === null
		)
			return;

		const delta = valueFromPointer(event) - dragStartValue;
		if (dragTarget === 'min') {
			emitValues(constrainMin(dragStartMin + delta, currentMax), currentMax, true);
		} else if (dragTarget === 'max') {
			emitValues(currentMin, constrainMax(dragStartMax + delta, currentMin), true);
		} else {
			const shift = clamp(delta, min - dragStartMin, max - dragStartMax);
			emitValues(dragStartMin + shift, dragStartMax + shift, true);
		}
	}

	function clearDrag(releaseCapture = true) {
		if (
			releaseCapture &&
			captureElement &&
			activePointerId !== null &&
			captureElement.hasPointerCapture(activePointerId)
		) {
			captureElement.releasePointerCapture(activePointerId);
		}
		captureElement = null;
		activePointerId = null;
		dragTarget = null;
		dragStartValue = null;
		dragStartMin = null;
		dragStartMax = null;
		dragMin = null;
		dragMax = null;
	}

	function handlePointerEnd(event: PointerEvent) {
		if (event.pointerId === activePointerId) clearDrag();
	}

	function handleLostPointerCapture(event: PointerEvent) {
		if (event.pointerId === activePointerId) clearDrag(false);
	}

	function handleKeydown(target: 'min' | 'max', event: KeyboardEvent) {
		if (!isValidRange) return;
		const step = STEP * (event.shiftKey ? 5 : 1);
		let nextValue: number | null = null;

		switch (event.key) {
			case 'ArrowLeft':
			case 'ArrowDown':
				nextValue = (target === 'min' ? currentMin : currentMax) - step;
				break;
			case 'ArrowRight':
			case 'ArrowUp':
				nextValue = (target === 'min' ? currentMin : currentMax) + step;
				break;
			case 'Home':
				nextValue = min;
				break;
			case 'End':
				nextValue = max;
				break;
			default:
				return;
		}

		event.preventDefault();
		if (target === 'min') emitValues(constrainMin(nextValue, currentMax), currentMax);
		else emitValues(currentMin, constrainMax(nextValue, currentMin));
	}

	function handleMinInput(e: Event) {
		inputMinValue = (e.target as HTMLInputElement).value;
	}
	function handleMinFocus() {
		inputMinFocused = true;
	}
	function handleMinBlur() {
		inputMinFocused = false;
		const value = Number(inputMinValue);
		if (!inputMinValue.trim() || !Number.isFinite(value)) {
			emitValues(min, currentMax);
			inputMinValue = min.toString();
			return;
		}
		const clampedValue = constrainMin(value, currentMax);
		emitValues(clampedValue, currentMax);
		inputMinValue = clampedValue.toString();
	}
	function handleInputKeydown(target: 'min' | 'max', event: KeyboardEvent) {
		const input = event.target as HTMLInputElement;
		if (event.key === 'Escape') {
			event.preventDefault();
			if (target === 'min') inputMinValue = currentMin.toString();
			else inputMaxValue = currentMax.toString();
			input.blur();
		} else if (event.key === 'Enter') {
			input.blur();
		}
	}

	function handleMaxInput(e: Event) {
		inputMaxValue = (e.target as HTMLInputElement).value;
	}
	function handleMaxFocus() {
		inputMaxFocused = true;
	}
	function handleMaxBlur() {
		inputMaxFocused = false;
		const value = Number(inputMaxValue);
		if (!inputMaxValue.trim() || !Number.isFinite(value)) {
			emitValues(currentMin, max);
			inputMaxValue = max.toString();
			return;
		}
		const clampedValue = constrainMax(value, currentMin);
		emitValues(currentMin, clampedValue);
		inputMaxValue = clampedValue.toString();
	}
	onDestroy(() => clearDrag());
</script>

<svelte:window onblur={() => clearDrag()} />

<div class="price-range-slider space-y-0">
	{#if histogramBars.length > 0 && isValidRange}
		<div class="relative h-20 w-full px-1 sm:h-24 sm:px-2">
			<div class="relative h-full" role="img" aria-label="Распределение цен по объявлениям">
				{#each histogramBars as bar, i (i)}
					<div
						class="absolute bottom-0 rounded-[3px] {dragTarget
							? ''
							: 'transition-colors duration-150 ease-out'} {bar.isActive
							? bar.isEdge
								? 'bg-zinc-600/90'
								: 'bg-zinc-900'
							: 'bg-zinc-200/60'}"
						style="left: calc({bar.x}% + 1px); width: calc({bar.width}% - 2px); height: {bar.height}%"
						aria-hidden="true"
					></div>
				{/each}
			</div>
		</div>
	{:else}
		<div class="relative h-20 w-full px-1 sm:h-24 sm:px-2">
			<div
				class="flex h-full items-end gap-[3px] opacity-60"
				role="img"
				aria-label="Данные о распределении цен недоступны"
			>
				{#each Array(16) as _}
					<div class="h-3 w-full flex-1 rounded-sm bg-zinc-100"></div>
				{/each}
			</div>
		</div>
	{/if}

	<div class="relative pt-11 pb-7 sm:pt-12 sm:pb-8">
		<div
			bind:this={sliderRef}
			role="presentation"
			class="relative mx-2 h-1 touch-manipulation rounded-full bg-zinc-200/90"
			onpointerdown={handleTrackPointerDown}
		>
			<div
				role="presentation"
				class="absolute h-full rounded-full bg-zinc-900 {dragTarget
					? ''
					: 'transition-[left,right] duration-150'}"
				style="left: {minPercent}%; right: {100 - maxPercent}%"
				onpointerdown={handleRangePointerDown}
				onpointermove={handlePointerMove}
				onpointerup={handlePointerEnd}
				onpointercancel={handlePointerEnd}
				onlostpointercapture={handleLostPointerCapture}
			></div>

			<div
				class="pointer-events-none absolute inset-x-0 top-0 {!isValidRange
					? 'hidden'
					: dragTarget
						? 'block'
						: 'hidden sm:block'}"
				aria-hidden="true"
			>
				<div
					class="absolute bottom-5 max-w-[116px] -translate-x-1/2 truncate rounded-xl border border-zinc-200/90 bg-white px-2.5 py-1 text-xs font-semibold text-zinc-900 tabular-nums shadow-[0_3px_12px_rgba(24,24,27,0.10)] {dragTarget
						? ''
						: 'transition-[left,bottom,opacity] duration-150'}"
					style="left: clamp(58px, {minPercent}%, calc(100% - 58px))"
				>
					{currentMin} BYN
				</div>
				<div
					class="absolute max-w-[116px] -translate-x-1/2 truncate rounded-xl border border-zinc-200/90 bg-white px-2.5 py-1 text-xs font-semibold text-zinc-900 tabular-nums shadow-[0_3px_12px_rgba(24,24,27,0.10)] {dragTarget
						? ''
						: 'transition-[left,bottom,opacity] duration-150'}"
					class:bottom-5={!bubblesOverlap}
					class:bottom-14={bubblesOverlap}
					style="left: clamp(58px, {maxPercent}%, calc(100% - 58px))"
				>
					{currentMax} BYN
				</div>
			</div>

			<button
				type="button"
				disabled={!isValidRange}
				onpointerdown={(event) => handleThumbPointerDown('min', event)}
				onpointermove={handlePointerMove}
				onpointerup={handlePointerEnd}
				onpointercancel={handlePointerEnd}
				onlostpointercapture={handleLostPointerCapture}
				onkeydown={(event) => handleKeydown('min', event)}
				class="group absolute top-1/2 flex h-11 w-11 -translate-x-1/2 -translate-y-1/2 touch-manipulation items-center justify-center rounded-full focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-900 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
				style="left: {minPercent}%"
				role="slider"
				aria-valuemin={min}
				aria-valuemax={currentMax - MIN_GAP}
				aria-valuenow={currentMin}
				aria-valuetext={`${currentMin} BYN`}
				aria-label="Минимальная цена"
			>
				<span
					class="h-6 w-6 rounded-full border border-zinc-300 bg-white shadow-[0_2px_8px_rgba(24,24,27,0.14)] {dragTarget
						? ''
						: 'transition-transform duration-150'} group-hover:scale-110 group-active:scale-95 {dragTarget ===
					'min'
						? 'scale-110 shadow-[0_3px_12px_rgba(24,24,27,0.20)]'
						: ''}"
				></span>
				<span class="sr-only">Минимальная цена: {currentMin} BYN</span>
			</button>

			<button
				type="button"
				disabled={!isValidRange}
				onpointerdown={(event) => handleThumbPointerDown('max', event)}
				onpointermove={handlePointerMove}
				onpointerup={handlePointerEnd}
				onpointercancel={handlePointerEnd}
				onlostpointercapture={handleLostPointerCapture}
				onkeydown={(event) => handleKeydown('max', event)}
				class="group absolute top-1/2 flex h-11 w-11 -translate-x-1/2 -translate-y-1/2 touch-manipulation items-center justify-center rounded-full focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-900 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
				style="left: {maxPercent}%"
				role="slider"
				aria-valuemin={currentMin + MIN_GAP}
				aria-valuemax={max}
				aria-valuenow={currentMax}
				aria-valuetext={`${currentMax} BYN`}
				aria-label="Максимальная цена"
			>
				<span
					class="h-6 w-6 rounded-full border border-zinc-300 bg-white shadow-[0_2px_8px_rgba(24,24,27,0.14)] {dragTarget
						? ''
						: 'transition-transform duration-150'} group-hover:scale-110 group-active:scale-95 {dragTarget ===
					'max'
						? 'scale-110 shadow-[0_3px_12px_rgba(24,24,27,0.20)]'
						: ''}"
				></span>
				<span class="sr-only">Максимальная цена: {currentMax} BYN</span>
			</button>
		</div>

		<div class="mt-7 flex items-start gap-3 px-1 sm:mt-8 sm:gap-5 sm:px-2">
			<div class="flex-1">
				<label for="price-input-min" class="mb-2 block text-sm font-medium text-zinc-700">
					Минимум
				</label>
				<div
					class="relative flex items-center rounded-2xl bg-zinc-50 px-3.5 py-3 ring-1 ring-inset {inputMinFocused
						? 'bg-white shadow-[0_2px_10px_rgba(24,24,27,0.06)] ring-zinc-900'
						: 'ring-zinc-200 transition-[background-color,box-shadow] duration-150 hover:bg-zinc-100/70 hover:ring-zinc-300'}"
				>
					<input
						id="price-input-min"
						type="number"
						inputmode="numeric"
						disabled={!isValidRange}
						{min}
						max={currentMax - MIN_GAP}
						step={STEP}
						value={inputMinValue}
						oninput={handleMinInput}
						onfocus={handleMinFocus}
						onblur={handleMinBlur}
						onkeydown={(event) => handleInputKeydown('min', event)}
						class="w-full min-w-0 [appearance:textfield] bg-transparent text-lg font-semibold text-zinc-900 tabular-nums outline-none placeholder:text-zinc-400 disabled:cursor-not-allowed disabled:opacity-50 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
						placeholder={min.toString()}
					/>
					<span class="ml-2 shrink-0 text-sm font-medium text-zinc-500">BYN</span>
				</div>
			</div>

			<div class="flex-1">
				<label for="price-input-max" class="mb-2 block text-sm font-medium text-zinc-700">
					Максимум
				</label>
				<div
					class="relative flex items-center rounded-2xl bg-zinc-50 px-3.5 py-3 ring-1 ring-inset {inputMaxFocused
						? 'bg-white shadow-[0_2px_10px_rgba(24,24,27,0.06)] ring-zinc-900'
						: 'ring-zinc-200 transition-[background-color,box-shadow] duration-150 hover:bg-zinc-100/70 hover:ring-zinc-300'}"
				>
					<input
						id="price-input-max"
						type="number"
						inputmode="numeric"
						disabled={!isValidRange}
						min={currentMin + MIN_GAP}
						{max}
						step={STEP}
						value={inputMaxValue}
						oninput={handleMaxInput}
						onfocus={handleMaxFocus}
						onblur={handleMaxBlur}
						onkeydown={(event) => handleInputKeydown('max', event)}
						class="w-full min-w-0 [appearance:textfield] bg-transparent text-lg font-semibold text-zinc-900 tabular-nums outline-none placeholder:text-zinc-400 disabled:cursor-not-allowed disabled:opacity-50 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
						placeholder={max.toString()}
					/>
					<span class="ml-2 shrink-0 text-sm font-medium text-zinc-500">BYN</span>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.touch-manipulation {
		touch-action: none;
		user-select: none;
		-webkit-user-select: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.price-range-slider *,
		.price-range-slider *::before,
		.price-range-slider *::after {
			animation-duration: 0.01ms !important;
			animation-iteration-count: 1 !important;
			transition-duration: 0.01ms !important;
			scroll-behavior: auto !important;
		}
	}
</style>
