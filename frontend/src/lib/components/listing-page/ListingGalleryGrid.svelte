<!-- src/lib/components/listing-page/ListingGalleryGrid.svelte -->
<script lang="ts">
	import { Grid2x2, ImageOff } from 'lucide-svelte';
	import type { EmblaCarouselType } from 'embla-carousel';

	interface Props {
		images?: string[];
		onOpen?: (index: number) => void;
		fullHeight?: boolean;
		mobileFullWidth?: boolean;
	}

	let {
		images = [],
		onOpen = () => {},
		fullHeight = false,
		mobileFullWidth = false
	}: Props = $props();

	const FALLBACK = `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='400'%3E%3Crect fill='%23f4f4f5' width='400' height='400'/%3E%3C/svg%3E`;

	const safeImages = $derived(images.filter(Boolean));
	const gridImages = $derived(safeImages.slice(0, 5));
	const extraCount = $derived(Math.max(0, safeImages.length - 5));

	let loadedImages = $state<Set<number>>(new Set([0, 1, 2, 3, 4]));

	let mobileViewport: HTMLDivElement | undefined = $state();
	let embla: EmblaCarouselType | null = null;
	let currentIndex = $state(0);

	function markLoaded(i: number) {
		if (!loadedImages.has(i)) loadedImages = new Set([...loadedImages, i]);
	}

	function handleError(e: Event) {
		const img = e.currentTarget as HTMLImageElement;
		if (img.src !== FALLBACK) img.src = FALLBACK;
	}

	$effect(() => {
		if (!mobileViewport || safeImages.length === 0) {
			currentIndex = 0;
			return;
		}

		const viewport = mobileViewport;
		let disposed = false;
		let wheelTimeout: ReturnType<typeof setTimeout> | null = null;
		let wheelDelta = 0;
		let emblaInstance: EmblaCarouselType | null = null;

		function handleWheel(e: WheelEvent) {
			if (!emblaInstance) return;
			if (Math.abs(e.deltaX) <= Math.abs(e.deltaY) || Math.abs(e.deltaX) <= 10) return;

			e.preventDefault();
			wheelDelta += e.deltaX;
			if (wheelTimeout) clearTimeout(wheelTimeout);
			wheelTimeout = setTimeout(() => {
				if (!emblaInstance) return;
				if (wheelDelta > 50) emblaInstance.scrollNext();
				else if (wheelDelta < -50) emblaInstance.scrollPrev();
				wheelDelta = 0;
			}, 80);
		}

		const handleSelect = () => {
			currentIndex = emblaInstance?.selectedScrollSnap() ?? 0;
		};

		(async () => {
			try {
				const EmblaCarousel = (await import('embla-carousel')).default;
				if (disposed) return;

				emblaInstance = EmblaCarousel(viewport, {
					loop: false,
					align: 'start',
					containScroll: 'trimSnaps',
					dragFree: false,
					skipSnaps: false,
					duration: 20
				});
				embla = emblaInstance;
				viewport.addEventListener('wheel', handleWheel, { passive: false });
				emblaInstance.on('select', handleSelect);
				handleSelect();
			} catch (error) {
				console.warn('Failed to initialize gallery carousel:', error);
			}
		})();

		return () => {
			disposed = true;
			viewport.removeEventListener('wheel', handleWheel);
			if (wheelTimeout) clearTimeout(wheelTimeout);
			if (emblaInstance) {
				emblaInstance.off('select', handleSelect);
				emblaInstance.destroy();
			}
			if (embla === emblaInstance) embla = null;
		};
	});

	const mobileContainerClass = $derived(mobileFullWidth ? '' : 'rounded-3xl');
	const mobileAspectClass = $derived(mobileFullWidth ? 'aspect-[4/5]' : 'aspect-square');
</script>

<!-- ── EMPTY STATE ──────────────────────────────────────────────────── -->
{#if safeImages.length === 0}
	<div
		class="flex flex-col items-center justify-center gap-3 bg-zinc-100 text-zinc-400
               {mobileFullWidth ? 'aspect-[4/5]' : 'aspect-[4/3] rounded-2xl'}
               {fullHeight ? 'h-full min-h-[50vh]' : ''}"
	>
		<ImageOff size={48} strokeWidth={1} />
		<span class="text-sm">Фотографии не добавлены</span>
	</div>
{:else}
	<!-- ── MOBILE CAROUSEL ─────────────────────────────────────────── -->
	<div class="relative lg:hidden">
		<div bind:this={mobileViewport} class="overflow-hidden {mobileContainerClass}">
			<div class="flex will-change-transform">
				{#each safeImages as image, i (image + i)}
					<button
						type="button"
						class="relative min-w-0 flex-[0_0_100%] touch-manipulation"
						onclick={() => onOpen(i)}
					>
						<div class="relative {mobileAspectClass} w-full bg-zinc-100">
							{#if !loadedImages.has(i)}
								<div class="absolute inset-0 animate-pulse bg-zinc-200"></div>
							{/if}
							<img
								class="absolute inset-0 h-full w-full object-cover
                                       transition-opacity duration-300
                                       {loadedImages.has(i) ? 'opacity-100' : 'opacity-0'}"
								src={image}
								alt="Фото {i + 1}"
								loading={i <= 2 ? 'eager' : 'lazy'}
								onload={() => markLoaded(i)}
								onerror={handleError}
								draggable="false"
							/>
						</div>
					</button>
				{/each}
			</div>
		</div>

		<div class="absolute right-4 bottom-10 z-10">
			<button
				type="button"
				onclick={() => onOpen(currentIndex)}
				class="flex h-8 touch-manipulation items-center gap-1.5
                       rounded-full bg-black/50 px-3 text-[13px] font-medium
                       text-white backdrop-blur-md transition-transform active:scale-95"
			>
				<Grid2x2 size={13} />
				<span class="tabular-nums">{currentIndex + 1} из {safeImages.length}</span>
			</button>
		</div>

		{#if safeImages.length <= 5}
			<div class="absolute bottom-4 left-1/2 z-10 flex -translate-x-1/2 items-center gap-1.5">
				{#each Array.from({ length: safeImages.length }, (_, i) => i) as i (i)}
					<div
						class="h-1.5 rounded-full bg-white transition-all duration-300
                               {i === currentIndex ? 'w-5 opacity-100' : 'w-1.5 opacity-50'}"
					></div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- ── DESKTOP GRID ──────────────────────────────── -->
	<!-- ── DESKTOP GRID ──────────────────────────────── -->
	<div class="hidden lg:block">
		<div class="grid h-[360px] grid-cols-[2fr_1fr_1fr] grid-rows-2 gap-2 xl:h-[480px]">
			<!-- 1 — Hero -->
			<button
				type="button"
				onclick={() => onOpen(0)}
				class="group relative row-span-2 overflow-hidden rounded-3xl bg-zinc-100"
			>
				{#if !loadedImages.has(0)}
					<div class="absolute inset-0 animate-pulse bg-zinc-200"></div>
				{/if}
				<img
					src={gridImages[0]}
					alt="Фото 1"
					class="h-full w-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.03]"
					loading="eager"
					onload={() => markLoaded(0)}
					onerror={handleError}
				/>
				<div
					class="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/20 via-transparent to-transparent"
				></div>
			</button>

			<!-- 2 -->
			{#if gridImages[1]}
				<button
					type="button"
					onclick={() => onOpen(1)}
					class="group relative overflow-hidden rounded-3xl bg-zinc-100"
				>
					{#if !loadedImages.has(1)}
						<div class="absolute inset-0 animate-pulse bg-zinc-200"></div>
					{/if}
					<img
						src={gridImages[1]}
						alt="Фото 2"
						class="h-full w-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.03]"
						loading="lazy"
						onload={() => markLoaded(1)}
						onerror={handleError}
					/>
				</button>
			{:else}
				<div class="rounded-3xl bg-zinc-100"></div>
			{/if}

			<!-- 3 -->
			{#if gridImages[2]}
				<button
					type="button"
					onclick={() => onOpen(2)}
					class="group relative overflow-hidden rounded-3xl bg-zinc-100"
				>
					{#if !loadedImages.has(2)}
						<div class="absolute inset-0 animate-pulse bg-zinc-200"></div>
					{/if}
					<img
						src={gridImages[2]}
						alt="Фото 3"
						class="h-full w-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.03]"
						loading="lazy"
						onload={() => markLoaded(2)}
						onerror={handleError}
					/>
				</button>
			{:else}
				<div class="rounded-3xl bg-zinc-100"></div>
			{/if}

			<!-- 4 -->
			{#if gridImages[3]}
				<button
					type="button"
					onclick={() => onOpen(3)}
					class="group relative overflow-hidden rounded-3xl bg-zinc-100"
				>
					{#if !loadedImages.has(3)}
						<div class="absolute inset-0 animate-pulse bg-zinc-200"></div>
					{/if}
					<img
						src={gridImages[3]}
						alt="Фото 4"
						class="h-full w-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.03]"
						loading="lazy"
						onload={() => markLoaded(3)}
						onerror={handleError}
					/>
				</button>
			{:else}
				<div class="rounded-3xl bg-zinc-100"></div>
			{/if}

			<!-- 5 + badge -->
			<button
				type="button"
				onclick={() => onOpen(4)}
				class="group relative overflow-hidden rounded-3xl bg-zinc-100"
			>
				{#if gridImages[4]}
					{#if !loadedImages.has(4)}
						<div class="absolute inset-0 animate-pulse bg-zinc-200"></div>
					{/if}
					<img
						src={gridImages[4]}
						alt="Фото 5"
						class="h-full w-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.03]"
						loading="lazy"
						onload={() => markLoaded(4)}
						onerror={handleError}
					/>
				{:else}
					<div class="h-full w-full bg-zinc-100"></div>
				{/if}

				{#if extraCount > 0}
					<div
						class="absolute inset-0 flex flex-col items-center justify-center gap-1.5
                                bg-black/45 backdrop-blur-[2px]
                                transition-colors duration-200 group-hover:bg-black/55"
					>
						<span class="text-sm font-semibold text-white">ещё {extraCount} фото</span>
					</div>
				{/if}
			</button>
		</div>
	</div>
{/if}
