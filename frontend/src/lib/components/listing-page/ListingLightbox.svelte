<!-- src/lib/components/listing-page/ListingLightbox.svelte -->
<script lang="ts">
	import { X, ChevronLeft, ChevronRight, ZoomIn, ZoomOut, RotateCcw } from 'lucide-svelte';
	import { onMount } from 'svelte';

	interface Props {
		images: string[];
		startIndex?: number;
		onClose: () => void;
	}

	let { images, startIndex = 0, onClose }: Props = $props();
	const LOCK_COUNT_ATTR = 'data-modal-lock-count';
	const LOCK_OVERFLOW_ATTR = 'data-modal-lock-overflow';
	const LOCK_PADDING_ATTR = 'data-modal-lock-padding-right';

	let currentIndex = $state(0);
	let isVisible = $state(false);
	let isClosing = $state(false);

	let containerEl: HTMLDivElement | undefined = $state();
	let thumbnailsEl: HTMLDivElement | undefined = $state();

	let zoom = $state(1);
	let panX = $state(0);
	let panY = $state(0);
	const MIN_ZOOM = 1;
	const MAX_ZOOM = 4;
	const ZOOM_STEP = 0.5;

	const isZoomed = $derived(zoom > 1);
	const canPrev = $derived(currentIndex > 0);
	const canNext = $derived(currentIndex < images.length - 1);

	let isDragging = $state(false);
	let isSwiping = $state(false);
	let isPinching = $state(false);
	let swipeX = $state(0);
	let imageLoaded = $state(false);
	let loadedImages = $state<Set<number>>(new Set());
	let closeTimeout: ReturnType<typeof setTimeout> | null = null;
	let lightboxEl: HTMLDivElement | undefined = $state();

	let startX = 0,
		startY = 0,
		lastX = 0,
		lastY = 0;
	let velocityX = 0,
		lastTime = 0,
		lastTapTime = 0;
	let pinchStartDistance = 0,
		pinchStartZoom = 1;

	function resetView() {
		zoom = 1;
		panX = 0;
		panY = 0;
	}

	function setZoom(newZoom: number, centerX?: number, centerY?: number) {
		const prevZoom = zoom;
		zoom = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, newZoom));

		if (zoom <= 1) {
			resetView();
			return;
		}

		if (centerX !== undefined && centerY !== undefined && containerEl) {
			const rect = containerEl.getBoundingClientRect();
			const cx = centerX - rect.left - rect.width / 2;
			const cy = centerY - rect.top - rect.height / 2;
			const scale = zoom / prevZoom;
			panX = cx - (cx - panX) * scale;
			panY = cy - (cy - panY) * scale;
		}

		constrainPan();
	}

	function constrainPan() {
		if (!containerEl || zoom <= 1) {
			panX = 0;
			panY = 0;
			return;
		}

		const rect = containerEl.getBoundingClientRect();
		const maxX = (rect.width * (zoom - 1)) / 2;
		const maxY = (rect.height * (zoom - 1)) / 2;
		panX = Math.max(-maxX, Math.min(maxX, panX));
		panY = Math.max(-maxY, Math.min(maxY, panY));
	}

	function goTo(index: number) {
		if (index === currentIndex || index < 0 || index >= images.length) return;
		resetView();
		currentIndex = index;
		imageLoaded = false;
		preloadAdjacent(index);
		scrollThumbnailIntoView(index);
	}

	function next() {
		if (canNext && !isZoomed) goTo(currentIndex + 1);
	}
	function prev() {
		if (canPrev && !isZoomed) goTo(currentIndex - 1);
	}

	function scrollThumbnailIntoView(index: number) {
		if (!thumbnailsEl) return;
		const thumb = thumbnailsEl.children[index] as HTMLElement;
		if (!thumb) return;
		const container = thumbnailsEl.getBoundingClientRect();
		const scrollLeft = thumb.offsetLeft - container.width / 2 + thumb.offsetWidth / 2;
		thumbnailsEl.scrollTo({ left: scrollLeft, behavior: 'smooth' });
	}

	function preloadAdjacent(index: number) {
		[-1, 1, 2].forEach((offset) => {
			const i = index + offset;
			if (i >= 0 && i < images.length && !loadedImages.has(i)) {
				const img = new Image();
				img.src = images[i];
				img.onload = () => {
					loadedImages = new Set([...loadedImages, i]);
				};
			}
		});
	}

	function close() {
		if (isClosing) return;
		isClosing = true;
		isVisible = false;
		if (closeTimeout) clearTimeout(closeTimeout);
		closeTimeout = setTimeout(() => {
			closeTimeout = null;
			onClose();
		}, 200);
	}

	function lockBodyScroll() {
		if (typeof document === 'undefined') return;
		const body = document.body;
		const html = document.documentElement;
		const lockCount = Number(body.getAttribute(LOCK_COUNT_ATTR) ?? '0');

		if (lockCount === 0) {
			body.setAttribute(LOCK_OVERFLOW_ATTR, body.style.overflow);
			body.setAttribute(LOCK_PADDING_ATTR, body.style.paddingRight);

			const scrollbarWidth = Math.max(0, window.innerWidth - html.clientWidth);
			body.style.overflow = 'hidden';
			if (scrollbarWidth > 0) body.style.paddingRight = `${scrollbarWidth}px`;
		}

		body.setAttribute(LOCK_COUNT_ATTR, String(lockCount + 1));
	}

	function unlockBodyScroll() {
		if (typeof document === 'undefined') return;
		const body = document.body;
		const lockCount = Number(body.getAttribute(LOCK_COUNT_ATTR) ?? '0');
		const nextCount = Math.max(0, lockCount - 1);

		if (nextCount === 0) {
			body.style.overflow = body.getAttribute(LOCK_OVERFLOW_ATTR) ?? '';
			body.style.paddingRight = body.getAttribute(LOCK_PADDING_ATTR) ?? '';
			body.removeAttribute(LOCK_COUNT_ATTR);
			body.removeAttribute(LOCK_OVERFLOW_ATTR);
			body.removeAttribute(LOCK_PADDING_ATTR);
			return;
		}

		body.setAttribute(LOCK_COUNT_ATTR, String(nextCount));
	}

	function handleKeydown(e: KeyboardEvent) {
		switch (e.key) {
			case 'Escape':
				if (isZoomed) {
					resetView();
				} else {
					close();
				}
				break;
			case 'ArrowRight':
				if (!isZoomed) next();
				break;
			case 'ArrowLeft':
				if (!isZoomed) prev();
				break;
			case '+':
			case '=':
				setZoom(zoom + ZOOM_STEP);
				break;
			case '-':
				setZoom(zoom - ZOOM_STEP);
				break;
			case '0':
				resetView();
				break;
		}
	}

	function handleWheel(e: WheelEvent) {
		e.preventDefault();
		setZoom(zoom + (e.deltaY > 0 ? -0.3 : 0.3), e.clientX, e.clientY);
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget && !isZoomed) close();
	}

	function handleDoubleTap(x: number, y: number) {
		if (isZoomed) {
			resetView();
		} else {
			setZoom(2.5, x, y);
		}
	}

	function handleMouseDown(e: MouseEvent) {
		if (e.button !== 0) return;
		const now = Date.now();
		if (now - lastTapTime < 300) {
			handleDoubleTap(e.clientX, e.clientY);
			lastTapTime = 0;
			return;
		}
		lastTapTime = now;

		if (!isZoomed) return;
		e.preventDefault();
		isDragging = true;
		startX = e.clientX;
		startY = e.clientY;
		lastX = panX;
		lastY = panY;
	}

	function handleMouseMove(e: MouseEvent) {
		if (!isDragging || !isZoomed) return;
		panX = lastX + (e.clientX - startX);
		panY = lastY + (e.clientY - startY);
		constrainPan();
	}

	function handleMouseUp() {
		isDragging = false;
	}

	function getDistance(t1: Touch, t2: Touch) {
		return Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY);
	}

	function getCenter(t1: Touch, t2: Touch) {
		return { x: (t1.clientX + t2.clientX) / 2, y: (t1.clientY + t2.clientY) / 2 };
	}

	function handleTouchStart(e: TouchEvent) {
		if (e.touches.length === 2) {
			isPinching = true;
			isSwiping = false;
			isDragging = false;
			pinchStartDistance = getDistance(e.touches[0], e.touches[1]);
			pinchStartZoom = zoom;
			return;
		}

		if (e.touches.length !== 1) return;
		const touch = e.touches[0];
		const now = Date.now();

		if (now - lastTapTime < 300) {
			e.preventDefault();
			handleDoubleTap(touch.clientX, touch.clientY);
			lastTapTime = 0;
			return;
		}
		lastTapTime = now;

		startX = touch.clientX;
		startY = touch.clientY;
		lastTime = now;
		velocityX = 0;

		if (isZoomed) {
			isDragging = true;
			lastX = panX;
			lastY = panY;
		} else {
			isSwiping = true;
			swipeX = 0;
		}
	}

	function handleTouchMove(e: TouchEvent) {
		if (e.touches.length === 2 && isPinching) {
			e.preventDefault();
			const distance = getDistance(e.touches[0], e.touches[1]);
			const center = getCenter(e.touches[0], e.touches[1]);
			setZoom(pinchStartZoom * (distance / pinchStartDistance), center.x, center.y);
			return;
		}

		if (e.touches.length !== 1) return;
		const touch = e.touches[0];
		const diffX = touch.clientX - startX;
		const diffY = touch.clientY - startY;

		if (isDragging && isZoomed) {
			e.preventDefault();
			panX = lastX + diffX;
			panY = lastY + diffY;
			constrainPan();
		} else if (isSwiping && !isZoomed && Math.abs(diffX) > Math.abs(diffY) * 1.5) {
			e.preventDefault();
			const atEdge =
				(currentIndex === 0 && diffX > 0) || (currentIndex === images.length - 1 && diffX < 0);
			swipeX = atEdge ? diffX * 0.15 : diffX;
			const now = Date.now();
			if (now - lastTime > 0) velocityX = diffX / (now - lastTime);
		}
	}

	function handleTouchEnd() {
		if (isPinching) {
			isPinching = false;
			if (zoom < 1.05) resetView();
			return;
		}

		if (isSwiping && !isZoomed) {
			const shouldSwipe = Math.abs(swipeX) > 60 || Math.abs(velocityX) > 0.3;
			if (shouldSwipe) {
				if (swipeX > 0 || velocityX > 0.3) {
					prev();
				} else {
					next();
				}
			}
			swipeX = 0;
		}

		isSwiping = false;
		isDragging = false;
	}

	onMount(() => {
		if (!images.length) {
			onClose();
			return;
		}

		lockBodyScroll();
		currentIndex = Math.max(0, Math.min(startIndex, images.length - 1));
		loadedImages = new Set([currentIndex]);
		requestAnimationFrame(() => {
			isVisible = true;
			lightboxEl?.focus();
		});
		preloadAdjacent(currentIndex);
		scrollThumbnailIntoView(currentIndex);

		return () => {
			if (closeTimeout) {
				clearTimeout(closeTimeout);
				closeTimeout = null;
			}
			unlockBodyScroll();
		};
	});

	const imageTransform = $derived.by(() => {
		const t: string[] = [];
		if (swipeX) t.push(`translateX(${swipeX}px)`);
		if (zoom !== 1) t.push(`scale(${zoom})`);
		if (panX || panY) t.push(`translate(${panX / zoom}px, ${panY / zoom}px)`);
		return t.join(' ') || 'none';
	});
</script>

<svelte:window onkeydown={handleKeydown} onmousemove={handleMouseMove} onmouseup={handleMouseUp} />

<div
	bind:this={lightboxEl}
	class="fixed inset-0 z-[10000] touch-none overscroll-contain bg-zinc-950
           transition-opacity duration-200 ease-out
           {isVisible ? 'opacity-100' : 'opacity-0'}"
	role="dialog"
	aria-modal="true"
	aria-label="Image gallery"
	tabindex="-1"
>
	<header
		class="pointer-events-none absolute top-0 right-0 left-0 z-50 h-18 px-4 pt-[max(env(safe-area-inset-top),0.75rem)] pb-2
               transition-all duration-300 ease-out
               {isZoomed ? '-translate-y-full opacity-0' : ''}"
	>
		<div class="pointer-events-auto flex h-full items-center justify-between">
			<div class="flex items-center gap-3">
				<button
					type="button"
					onclick={close}
					aria-label="Close"
					class="flex size-11 cursor-pointer touch-manipulation items-center justify-center
                           rounded-full bg-white/10 text-white
                           backdrop-blur-xl transition-all duration-200
                           hover:scale-105 hover:bg-white/15 active:scale-95"
				>
					<X size={20} strokeWidth={2.5} />
				</button>

				<div
					class="hidden h-10 items-center gap-1.5 rounded-full bg-white/10 px-4
                            tabular-nums backdrop-blur-xl sm:flex"
				>
					<span class="font-semibold text-white">{currentIndex + 1}</span>
					<span class="text-white/40">/</span>
					<span class="text-white/60">{images.length}</span>
				</div>
			</div>

			<div
				class="hidden items-center gap-1 rounded-full bg-white/10 p-1
                        backdrop-blur-xl lg:flex"
			>
				<button
					type="button"
					onclick={() => setZoom(zoom - ZOOM_STEP)}
					disabled={zoom <= MIN_ZOOM}
					aria-label="Zoom out"
					class="flex size-9 cursor-pointer touch-manipulation items-center justify-center
                           rounded-full text-white/60 transition-all
                           duration-200 hover:bg-white/10
                           hover:text-white disabled:cursor-not-allowed disabled:opacity-30"
				>
					<ZoomOut size={18} />
				</button>

				<span
					class="min-w-13 text-center text-sm font-medium text-white/60 tabular-nums select-none"
				>
					{Math.round(zoom * 100)}%
				</span>

				<button
					type="button"
					onclick={() => setZoom(zoom + ZOOM_STEP)}
					disabled={zoom >= MAX_ZOOM}
					aria-label="Zoom in"
					class="flex size-9 cursor-pointer touch-manipulation items-center justify-center
                           rounded-full text-white/60 transition-all
                           duration-200 hover:bg-white/10
                           hover:text-white disabled:cursor-not-allowed disabled:opacity-30"
				>
					<ZoomIn size={18} />
				</button>

				{#if isZoomed}
					<button
						type="button"
						onclick={resetView}
						aria-label="Reset zoom"
						class="ml-1 flex size-9 cursor-pointer touch-manipulation items-center justify-center
                               rounded-full border-l border-white/10
                               pl-2 text-white/60 transition-all
                               duration-200 hover:bg-white/10 hover:text-white"
					>
						<RotateCcw size={16} />
					</button>
				{/if}
			</div>
		</div>
	</header>

	<button
		type="button"
		onclick={prev}
		aria-label="Previous image"
		class="absolute top-1/2 left-6 z-40 hidden size-14 -translate-y-1/2 cursor-pointer touch-manipulation
               items-center justify-center rounded-full bg-white
               text-zinc-900 shadow-2xl transition-all
               duration-200 hover:scale-110
               active:scale-95 lg:flex xl:left-10
               {!canPrev || isZoomed ? 'pointer-events-none opacity-0' : ''}"
	>
		<ChevronLeft size={28} strokeWidth={2.5} />
	</button>

	<button
		type="button"
		onclick={next}
		aria-label="Next image"
		class="absolute top-1/2 right-6 z-40 hidden size-14 -translate-y-1/2 cursor-pointer touch-manipulation
               items-center justify-center rounded-full bg-white
               text-zinc-900 shadow-2xl transition-all
               duration-200 hover:scale-110
               active:scale-95 lg:flex xl:right-10
               {!canNext || isZoomed ? 'pointer-events-none opacity-0' : ''}"
	>
		<ChevronRight size={28} strokeWidth={2.5} />
	</button>

	<div
		bind:this={containerEl}
		class="absolute inset-0 flex items-center justify-center
               overflow-hidden px-4 pt-[calc(4.5rem+env(safe-area-inset-top))] pb-[calc(8rem+env(safe-area-inset-bottom))] sm:px-6 sm:pt-20 sm:pb-36 lg:px-30"
		onclick={handleBackdropClick}
		onwheel={handleWheel}
		onmousedown={handleMouseDown}
		ontouchstart={handleTouchStart}
		ontouchmove={handleTouchMove}
		ontouchend={handleTouchEnd}
		role="presentation"
	>
		<div class="relative flex h-full w-full items-center justify-center">
			{#key currentIndex}
				<img
					src={images[currentIndex]}
					alt="Image {currentIndex + 1} of {images.length}"
					class="max-h-full max-w-full rounded-xl object-contain will-change-transform
                           select-none sm:rounded-2xl
                           {imageLoaded ? 'opacity-100' : 'opacity-0'}
                           {isDragging
						? 'cursor-grabbing'
						: isZoomed
							? 'cursor-grab'
							: 'cursor-zoom-in'}"
					style="transform: {imageTransform};
                           transition: {isDragging || isSwiping || isPinching
						? 'none'
						: 'transform 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94), opacity 0.3s ease'};"
					onload={() => (imageLoaded = true)}
					draggable="false"
				/>
			{/key}

			{#if !imageLoaded}
				<div class="absolute inset-0 flex items-center justify-center">
					<div
						class="size-10 animate-spin rounded-full border-3 border-white/20 border-t-white"
					></div>
				</div>
			{/if}
		</div>
	</div>

	<footer
		class="absolute right-0 bottom-0 left-0 z-40
               transition-all duration-300 ease-out
               {isZoomed ? 'pointer-events-none translate-y-full opacity-0' : ''}"
	>
		<div
			class="pointer-events-none absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/60 to-transparent"
		></div>

		<div class="relative px-4 py-5 pb-[calc(1.25rem+env(safe-area-inset-bottom))] sm:px-6">
			<div
				bind:this={thumbnailsEl}
				class="-my-2 flex gap-3 overflow-x-auto overflow-y-visible
                       scroll-smooth py-2 [-ms-overflow-style:none]
                       [scrollbar-width:none] lg:justify-center [&::-webkit-scrollbar]:hidden"
			>
				{#each images as image, i (image)}
					<button
						type="button"
						onclick={() => goTo(i)}
						aria-label="Go to image {i + 1}"
						aria-current={i === currentIndex ? 'true' : undefined}
						class="h-12 w-16 shrink-0 cursor-pointer touch-manipulation overflow-hidden rounded-lg transition-all
                               duration-250 sm:h-14 sm:w-20 sm:rounded-xl
                               lg:h-16 lg:w-24 lg:rounded-xl
                               {i === currentIndex
							? 'scale-110 opacity-100 ring-3 ring-white'
							: 'opacity-40 hover:scale-105 hover:opacity-70'}"
					>
						<img
							src={image}
							alt=""
							class="h-full w-full object-cover"
							loading="lazy"
							draggable="false"
						/>
					</button>
				{/each}
			</div>
		</div>
	</footer>
</div>
