<script lang="ts">
	import { onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { Check, Copy, Share2 } from 'lucide-svelte';
	import { scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import Button from '$lib/components/ui/Button.svelte';

	type Appearance = 'icon' | 'text';
	type Variant = 'solid' | 'outline' | 'ghost' | 'soft' | 'link' | 'glass';
	type Tone = 'primary' | 'neutral' | 'danger' | 'success' | 'rose';
	type Size = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'icon-xs' | 'icon-sm' | 'icon' | 'icon-lg';
	type Radius = 'none' | 'md' | 'lg' | 'xl' | '2xl' | 'pill';

	interface Props {
		title: string;
		text?: string;
		url?: string;
		appearance?: Appearance;
		variant?: Variant;
		tone?: Tone;
		size?: Size;
		radius?: Radius;
		class?: string;
	}

	let {
		title,
		text = 'Поделиться',
		url,
		appearance = 'text',
		variant = 'ghost',
		tone = 'neutral',
		size = 'sm',
		radius = 'xl',
		class: className = ''
	}: Props = $props();

	let menuOpen = $state(false);
	let copied = $state(false);
	let rootEl = $state<HTMLElement | null>(null);
	let copyResetTimeout: ReturnType<typeof setTimeout> | null = null;

	const iconSize = $derived(appearance === 'icon' ? 18 : 15);
	const hasNativeShare = $derived(browser && typeof navigator.share === 'function');
	const shareUrl = $derived(url || (browser ? window.location.href : ''));

	function closeMenu() {
		menuOpen = false;
	}

	async function copyLink() {
		if (!shareUrl) return;

		try {
			await navigator.clipboard.writeText(shareUrl);
		} catch {
			const input = document.createElement('input');
			input.value = shareUrl;
			document.body.appendChild(input);
			input.select();
			document.execCommand('copy');
			document.body.removeChild(input);
		}

		copied = true;
		if (copyResetTimeout) clearTimeout(copyResetTimeout);
		copyResetTimeout = setTimeout(() => {
			copied = false;
			closeMenu();
			copyResetTimeout = null;
		}, 1400);
	}

	async function handleTriggerClick(e: MouseEvent) {
		e.stopPropagation();
		if (hasNativeShare) {
			try {
				await navigator.share({
					title,
					url: shareUrl
				});
			} catch {
				/* cancelled */
			}
			return;
		}
		menuOpen = !menuOpen;
	}

	function handleWindowPointerDown(e: PointerEvent) {
		if (!menuOpen || !rootEl) return;
		const target = e.target as Node | null;
		if (target && rootEl.contains(target)) return;
		closeMenu();
	}

	function handleWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeMenu();
	}

	onDestroy(() => {
		if (copyResetTimeout) {
			clearTimeout(copyResetTimeout);
			copyResetTimeout = null;
		}
	});
</script>

<svelte:window onpointerdown={handleWindowPointerDown} onkeydown={handleWindowKeydown} />

<div class="relative" bind:this={rootEl}>
	<Button
		{variant}
		{tone}
		{size}
		{radius}
		onclick={handleTriggerClick}
		class={className}
		aria-label="Поделиться"
		aria-expanded={menuOpen}
		aria-haspopup={hasNativeShare ? undefined : 'menu'}
	>
		<Share2 size={iconSize} strokeWidth={2} />
		{#if appearance === 'text'}
			{text}
		{/if}
	</Button>

	{#if menuOpen}
		<div
			transition:scale={{ duration: 180, easing: backOut, start: 0.95 }}
			class="dropdown-surface absolute top-full right-0 z-[60] mt-2 w-64 max-w-[min(20rem,calc(100vw-1.5rem))] bg-white/95 backdrop-blur-sm"
			role="menu"
		>
			<button
				type="button"
				onclick={copyLink}
				class="flex w-full touch-manipulation items-center gap-3 px-4 py-3.5 text-left transition-colors hover:bg-zinc-50 active:bg-zinc-100"
				role="menuitem"
				disabled={copied}
			>
				{#if copied}
					<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-emerald-50">
						<Check size={17} class="text-emerald-600" />
					</div>
					<span class="text-[14px] font-medium text-emerald-700">Скопировано!</span>
				{:else}
					<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-zinc-100">
						<Copy size={16} class="text-zinc-600" />
					</div>
					<div>
						<div class="text-[14px] font-medium text-zinc-900">Скопировать ссылку</div>
						<div class="text-[12px] text-zinc-400">Поделитесь с друзьями</div>
					</div>
				{/if}
			</button>
		</div>
	{/if}
</div>
