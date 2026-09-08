<!-- src/lib/components/ui/Modal.svelte -->
<script lang="ts">
	import { X } from 'lucide-svelte';
	import { tick, type Snippet } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { backOut, cubicOut } from 'svelte/easing';

	interface Props {
		open: boolean;
		onClose: () => void;
		title?: string;
		id?: string;
		maxWidth?: 'sm' | 'md' | 'lg';
		showHandle?: boolean;
		swipeDismiss?: boolean;
		showClose?: boolean;
		closeOnBackdrop?: boolean;
		closeOnEscape?: boolean;
		children: Snippet;
		footer?: Snippet;
	}

	let {
		open,
		onClose,
		title = '',
		id: propId,
		maxWidth = 'md',
		showHandle = true,
		swipeDismiss = true,
		showClose = true,
		closeOnBackdrop = true,
		closeOnEscape = true,
		children,
		footer
	}: Props = $props();

	const modalId = $derived(propId ?? `modal-${Math.random().toString(36).slice(2, 8)}`);
	const titleId = $derived(`${modalId}-title`);

	const maxWidthClass: Record<string, string> = {
		sm: 'sm:max-w-sm',
		md: 'sm:max-w-md',
		lg: 'sm:max-w-lg'
	};

	const LOCK_COUNT_ATTR = 'data-modal-lock-count';
	const LOCK_OVERFLOW_ATTR = 'data-modal-lock-overflow';
	const LOCK_PADDING_ATTR = 'data-modal-lock-padding-right';

	// ── Portal: монтируем прямо в <body>, минуя любые stacking context ──
	function portal(node: HTMLElement) {
		document.body.appendChild(node);
		return {
			destroy() {
				node.remove();
			}
		};
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

	$effect(() => {
		if (!open) return;
		lockBodyScroll();
		return () => unlockBodyScroll();
	});

	let wrapperEl: HTMLDivElement | undefined = $state();
	let panelEl: HTMLElement | undefined = $state();
	let lastFocusedEl: HTMLElement | null = null;

	$effect(() => {
		if (typeof document === 'undefined' || !open) return;
		lastFocusedEl = document.activeElement instanceof HTMLElement ? document.activeElement : null;
		tick().then(() => wrapperEl?.focus());
		return () => {
			if (lastFocusedEl && document.contains(lastFocusedEl)) {
				lastFocusedEl.focus();
			}
		};
	});

	let swipeY = $state(0);
	let isSwiping = $state(false);
	let isSwipeCandidate = $state(false);
	let allowSwipe = $state(false);
	let swipeStartY = 0;
	let swipeStartX = 0;
	let swipeStartTime = 0;

	const backdropOpacity = $derived(Math.max(0, 1 - swipeY / 400));

	function handleTouchStart(e: TouchEvent) {
		if (!swipeDismiss || e.touches.length !== 1) return;
		const target = e.target as Element | null;
		const isHandle = !!target?.closest('[data-swipe-handle]');
		if (!isHandle) return;
		allowSwipe = true;
		isSwipeCandidate = true;
		isSwiping = false;
		swipeStartX = e.touches[0].clientX;
		swipeStartY = e.touches[0].clientY;
		swipeStartTime = Date.now();
		swipeY = 0;
	}

	function handleTouchMove(e: TouchEvent) {
		if (!isSwipeCandidate || !allowSwipe || e.touches.length !== 1) return;
		const currentX = e.touches[0].clientX;
		const currentY = e.touches[0].clientY;
		const diffX = currentX - swipeStartX;
		const diffY = currentY - swipeStartY;

		if (!isSwiping) {
			const absX = Math.abs(diffX);
			const absY = Math.abs(diffY);
			if (absX < 8 && absY < 8) return;
			if (absX > absY) {
				isSwipeCandidate = false;
				return;
			}
			if (diffY <= 0) {
				isSwipeCandidate = false;
				return;
			}
			isSwiping = true;
		}

		if (diffY > 0) {
			e.preventDefault();
			swipeY = diffY * 0.8;
		}
	}

	function handleTouchEnd() {
		if (!isSwipeCandidate) return;
		const wasSwiping = isSwiping;
		isSwiping = false;
		isSwipeCandidate = false;
		allowSwipe = false;
		if (!wasSwiping) return;
		const elapsed = Date.now() - swipeStartTime;
		const velocity = elapsed > 0 ? swipeY / elapsed : 0;
		if (swipeY > 180 || velocity > 0.8) {
			swipeY = 500;
			setTimeout(() => {
				swipeY = 0;
				onClose();
			}, 200);
		} else {
			swipeY = 0;
		}
	}

	$effect(() => {
		if (!open) {
			swipeY = 0;
			isSwiping = false;
			isSwipeCandidate = false;
			allowSwipe = false;
		}
	});

	function handleWrapperClick(e: MouseEvent) {
		if (!closeOnBackdrop || isSwiping) return;
		const target = e.target as HTMLElement;
		if (target === wrapperEl || target.hasAttribute('data-modal-backdrop')) {
			onClose();
		}
	}

	function getFocusableElements() {
		if (!panelEl) return [];
		const selectors =
			'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), ' +
			'select:not([disabled]), [tabindex]:not([tabindex="-1"])';
		return Array.from(panelEl.querySelectorAll<HTMLElement>(selectors)).filter(
			(el) => el.getAttribute('aria-hidden') !== 'true' && !el.hasAttribute('disabled')
		);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && closeOnEscape) {
			e.preventDefault();
			onClose();
			return;
		}

		if (e.key !== 'Tab') return;
		const focusable = getFocusableElements();
		if (focusable.length === 0) {
			e.preventDefault();
			wrapperEl?.focus();
			return;
		}

		const first = focusable[0];
		const last = focusable[focusable.length - 1];
		const active = document.activeElement as HTMLElement | null;

		if (e.shiftKey) {
			if (active === first || active === wrapperEl) {
				e.preventDefault();
				last.focus();
			}
			return;
		}

		if (active === last) {
			e.preventDefault();
			first.focus();
		}
	}
</script>

{#if open}
	<div use:portal>
		<div
			bind:this={wrapperEl}
			class="fixed inset-0 z-[10000] flex items-end pb-[env(safe-area-inset-bottom)] sm:items-center sm:justify-center sm:p-4"
			onclick={handleWrapperClick}
			onkeydown={handleKeydown}
			role="dialog"
			aria-modal="true"
			aria-labelledby={title ? titleId : undefined}
			tabindex="-1"
		>
			<div
				class="absolute inset-0 bg-black/40 backdrop-blur-sm"
				data-modal-backdrop
				style="opacity: {backdropOpacity};"
				transition:fade={{ duration: 150 }}
			></div>

			<section
				bind:this={panelEl}
				role="dialog"
				aria-modal="true"
				tabindex="-1"
				class="relative flex max-h-[calc(100dvh-env(safe-area-inset-top))] w-full flex-col overflow-hidden rounded-t-3xl border border-zinc-200/80 bg-white shadow-[0_28px_80px_rgba(0,0,0,0.24)] will-change-transform
	                   sm:max-h-[min(88dvh,52rem)] sm:rounded-3xl {maxWidthClass[maxWidth] ??
					maxWidthClass.md}"
				style="transform: translateY({Math.max(0, swipeY)}px);
	                   transition: {isSwiping
					? 'none'
					: 'transform 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94)'};"
				ontouchstart={handleTouchStart}
				ontouchmove={handleTouchMove}
				ontouchend={handleTouchEnd}
				in:fly={{ y: 100, duration: 300, easing: backOut }}
				out:fly={{ y: 100, duration: 200, easing: cubicOut }}
			>
				{#if showHandle}
					<div class="flex justify-center pt-3 pb-1 sm:hidden" data-swipe-handle>
						<div
							class="h-1.5 w-12 rounded-full bg-zinc-300 transition-all duration-200
	                               {isSwiping ? 'w-16 bg-zinc-400' : ''}"
						></div>
					</div>
				{/if}

				{#if title || showClose}
					<div class="flex items-center justify-between px-5 py-4 sm:px-6" data-swipe-handle>
						{#if showClose}
							<div class="w-11"></div>
						{/if}

						{#if title}
							<h3 id={titleId} class="min-w-0 flex-1 text-center text-base font-bold text-zinc-900">
								{title}
							</h3>
						{:else}
							<div class="flex-1"></div>
						{/if}

						{#if showClose}
							<button
								type="button"
								onclick={onClose}
								class="-mr-1 flex h-11 w-11 shrink-0 touch-manipulation items-center
	                                   justify-center rounded-full transition-colors hover:bg-zinc-100"
								aria-label="Закрыть"
							>
								<X size={20} strokeWidth={2.5} class="text-zinc-700" />
							</button>
						{/if}
					</div>
					<div class="h-px bg-zinc-100"></div>
				{/if}

				<div class="min-h-0 flex-1 touch-pan-y overflow-y-auto overscroll-contain">
					{@render children()}
				</div>

				{#if footer}
					{@render footer()}
				{/if}

				<div class="h-[max(env(safe-area-inset-bottom),0.5rem)] sm:hidden"></div>
			</section>
		</div>
	</div>
{/if}
