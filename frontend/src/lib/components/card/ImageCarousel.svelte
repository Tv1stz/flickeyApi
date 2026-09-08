<script lang="ts">
	import { onMount } from 'svelte';
	import type { EmblaCarouselType } from 'embla-carousel';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import type { ImageCarouselProps } from './types';

	let {
		images = [],
		alt = 'Фото',
		loop = false,
		showDots = true,
		showArrows = true,
		maxImages = 0
	}: ImageCarouselProps = $props();

	let viewport: HTMLDivElement | undefined = $state();
	let embla: EmblaCarouselType | null = null;

	let selectedIndex = $state(0);
	let loadedImages = $state<Set<number>>(new Set([0, 1]));
	let canScrollPrev = $state(false);
	let canScrollNext = $state(true);

	const FALLBACK_IMAGE = `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='433'%3E%3Crect fill='%23f4f4f5' width='400' height='433'/%3E%3C/svg%3E`;

	const displayImages = $derived(maxImages > 0 ? images.slice(0, maxImages) : images);

	const visibleDots = $derived.by(() => {
		const total = displayImages.length;
		if (total <= 5) return displayImages.map((_, i) => i);

		const current = selectedIndex;
		if (current <= 2) return [0, 1, 2, 3, 4];
		if (current >= total - 3) return [total - 5, total - 4, total - 3, total - 2, total - 1];
		return [current - 2, current - 1, current, current + 1, current + 2];
	});

	function markLoaded(index: number) {
		if (!loadedImages.has(index)) {
			loadedImages = new Set([...loadedImages, index]);
		}
	}

	function preloadAdjacent(index: number) {
		const toLoad = [index - 1, index + 1, index + 2].filter(
			(i) => i >= 0 && i < displayImages.length && !loadedImages.has(i)
		);
		toLoad.forEach((i) => markLoaded(i));
	}

	function handleImageError(e: Event) {
		const img = e.target as HTMLImageElement;
		if (img.src !== FALLBACK_IMAGE) {
			img.src = FALLBACK_IMAGE;
		}
	}

	function scrollPrev(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		embla?.scrollPrev();
	}

	function scrollNext(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		embla?.scrollNext();
	}

	function scrollTo(index: number, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		embla?.scrollTo(index);
	}

	onMount(() => {
		if (!viewport || displayImages.length === 0) return;

		let destroyed = false;

		let wheelTimeout: ReturnType<typeof setTimeout> | null = null;
		let wheelDelta = 0;

		function handleWheel(e: WheelEvent) {
			if (!embla) return;

			if (Math.abs(e.deltaX) > Math.abs(e.deltaY) && Math.abs(e.deltaX) > 10) {
				e.preventDefault();
				e.stopPropagation();

				wheelDelta += e.deltaX;

				if (wheelTimeout) clearTimeout(wheelTimeout);

				wheelTimeout = setTimeout(() => {
					if (!embla) return;
					if (wheelDelta > 50) embla.scrollNext();
					else if (wheelDelta < -50) embla.scrollPrev();
					wheelDelta = 0;
				}, 80);
			}
		}

		import('embla-carousel').then(({ default: EmblaCarousel }) => {
			if (destroyed || !viewport) return;

			embla = EmblaCarousel(viewport, {
				loop,
				align: 'start',
				containScroll: 'trimSnaps',
				duration: 20,
				dragFree: false,
				skipSnaps: false
			});

			viewport.addEventListener('wheel', handleWheel, { passive: false });

			const onSelect = () => {
				if (!embla) return;
				const index = embla.selectedScrollSnap();
				selectedIndex = index;
				canScrollPrev = embla.canScrollPrev();
				canScrollNext = embla.canScrollNext();
				preloadAdjacent(index);
			};

			embla.on('select', onSelect);
			onSelect();
		});

		return () => {
			destroyed = true;
			viewport?.removeEventListener('wheel', handleWheel);
			if (wheelTimeout) clearTimeout(wheelTimeout);
			embla?.destroy();
		};
	});
</script>

{#if displayImages.length === 0}
	<div class="flex aspect-[12/13] w-full items-center justify-center rounded-2xl bg-zinc-100">
		<svg class="h-10 w-10 text-zinc-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="1"
				d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
			/>
		</svg>
	</div>
{:else}
	<div class="group/carousel relative contain-layout">
		<div bind:this={viewport} class="touch-pan-y overflow-hidden rounded-2xl">
			<div class="flex will-change-transform">
				{#each displayImages as src, i (src + i)}
					<div class="min-w-full flex-shrink-0">
						<div class="relative aspect-[12/13] w-full bg-zinc-100">
							{#if !loadedImages.has(i)}
								<div class="absolute inset-0 animate-pulse bg-zinc-200"></div>
							{/if}
							{#if loadedImages.has(i) || i <= 1}
								<img
									{src}
									loading={i <= 1 ? 'eager' : 'lazy'}
									decoding="async"
									onload={() => markLoaded(i)}
									onerror={handleImageError}
									class="absolute inset-0 h-full w-full object-cover
                                           transition-opacity duration-300
                                           {loadedImages.has(i) ? 'opacity-100' : 'opacity-0'}"
									alt={`${alt} ${i + 1}`}
									draggable="false"
								/>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>

		{#if showArrows && displayImages.length > 1}
			<button
				type="button"
				onclick={scrollPrev}
				disabled={!loop && !canScrollPrev}
				class="absolute top-1/2 left-2.5 z-10 flex
                       h-8 w-8 -translate-y-1/2 touch-manipulation items-center
                       justify-center rounded-full bg-white/90 opacity-0
                       shadow-md backdrop-blur-sm
                       transition-all duration-200 group-hover/carousel:opacity-100
                       hover:scale-105 hover:bg-white hover:shadow-lg
                       active:scale-95 disabled:hidden"
				aria-label="Предыдущее фото"
			>
				<ChevronLeft size={18} strokeWidth={2} class="text-zinc-700" />
			</button>

			<button
				type="button"
				onclick={scrollNext}
				disabled={!loop && !canScrollNext}
				class="absolute top-1/2 right-2.5 z-10 flex
                       h-8 w-8 -translate-y-1/2 touch-manipulation items-center
                       justify-center rounded-full bg-white/90 opacity-0
                       shadow-md backdrop-blur-sm
                       transition-all duration-200 group-hover/carousel:opacity-100
                       hover:scale-105 hover:bg-white hover:shadow-lg
                       active:scale-95 disabled:hidden"
				aria-label="Следующее фото"
			>
				<ChevronRight size={18} strokeWidth={2} class="text-zinc-700" />
			</button>
		{/if}

		{#if showDots && displayImages.length > 1}
			<div
				class="absolute bottom-2.5 left-1/2 z-10 flex
                        -translate-x-1/2 items-center gap-1 rounded-full bg-black/30
                        px-2 py-1.5 backdrop-blur-sm"
			>
				{#each visibleDots as dotIndex (dotIndex)}
					{@const isActive = selectedIndex === dotIndex}
					<button
						type="button"
						onclick={(e) => scrollTo(dotIndex, e)}
						aria-label="Фото {dotIndex + 1}"
						aria-current={isActive ? 'true' : undefined}
						class="touch-manipulation rounded-full transition-all duration-200
                               {isActive
							? 'h-1.5 w-4 bg-white'
							: 'h-1.5 w-1.5 bg-white/50 hover:bg-white/80'}"
					></button>
				{/each}
			</div>
		{/if}
	</div>
{/if}
