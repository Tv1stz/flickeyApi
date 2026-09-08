<!-- src/lib/components/listing-page/ListingHeader.svelte -->
<script lang="ts">
	import { onDestroy } from 'svelte';
	import { ArrowLeft, Heart } from 'lucide-svelte';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/ui/Button.svelte';
	import ShareControl from '$lib/components/ui/ShareControl.svelte';
	import type { Listing } from '$lib/components/card/types';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import { toast } from '$lib/stores/toastStore';

	interface Props {
		title: string;
		address?: string;
		transparent?: boolean;
		listing?: Listing;
		showActions?: boolean;
		showTitle?: boolean;
	}

	let {
		title,
		address = '',
		transparent = false,
		listing,
		showActions = true,
		showTitle = false
	}: Props = $props();

	let isFavorite = $derived(listing ? favoritesStore.has(listing.id) : false);
	let isAnimating = $state(false);
	let authenticated = $derived(authStore.isAuthenticated);
	let currentUserId = $derived(authStore.user?.id ?? null);
	let favoriteAnimationTimeout: ReturnType<typeof setTimeout> | null = null;

	function goBack() {
		if (typeof window !== 'undefined' && window.history.length > 1) {
			window.history.back();
		} else {
			goto(resolve('/'));
		}
	}

	function toggleFavorite(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		if (!listing) return;

		if (!authenticated) {
			authStore.setPendingAction('favorite', `/listings/${listing.id}`);
			toast.info('Войдите в аккаунт', 'Чтобы добавлять объявления в избранное');
			goto(resolve('/auth'));
			return;
		}

		favoritesStore.toggle(listing);
		isAnimating = true;
		if (typeof window !== 'undefined' && 'vibrate' in navigator) navigator.vibrate(10);
		if (favoriteAnimationTimeout) clearTimeout(favoriteAnimationTimeout);
		favoriteAnimationTimeout = setTimeout(() => {
			isAnimating = false;
			favoriteAnimationTimeout = null;
		}, 400);
	}

	const transparentButtonClass = $derived(
		transparent
			? '!bg-white/70 !backdrop-blur-md !border-transparent !text-zinc-900 hover:!bg-white/85 !shadow-md'
			: ''
	);

	const favoriteButtonClass = $derived.by(() => {
		if (transparent) return transparentButtonClass;
		return isFavorite ? '!bg-rose-50 !border-rose-200' : '';
	});

	const canToggleFavorite = $derived(
		Boolean(listing) && showActions && listing?.ownerId !== currentUserId
	);

	onDestroy(() => {
		if (favoriteAnimationTimeout) {
			clearTimeout(favoriteAnimationTimeout);
			favoriteAnimationTimeout = null;
		}
	});
</script>

<!-- MOBILE HEADER -->
<nav class="flex items-center justify-between px-4 py-3 lg:hidden">
	<Button
		variant="glass"
		tone="neutral"
		size="icon"
		radius="pill"
		onclick={goBack}
		class={transparentButtonClass}
		aria-label="Назад"
	>
		<ArrowLeft size={18} strokeWidth={2} />
	</Button>

	{#if showActions}
		<div class="flex items-center gap-2">
			<ShareControl
				{title}
				appearance="icon"
				variant="glass"
				tone="neutral"
				size="icon"
				radius="pill"
				class={transparentButtonClass}
			/>

			{#if canToggleFavorite}
				<Button
					variant="glass"
					tone={isFavorite ? 'rose' : 'neutral'}
					size="icon"
					radius="pill"
					onclick={toggleFavorite}
					class={favoriteButtonClass}
					aria-label={isFavorite ? 'Удалить из избранного' : 'Добавить в избранное'}
					aria-pressed={isFavorite}
				>
					<span class="relative">
						<Heart
							size={18}
							strokeWidth={2}
							class="transition-all duration-300
                                   {isFavorite ? 'fill-rose-500 text-rose-500' : ''}
                                   {isAnimating ? 'scale-125' : 'scale-100'}"
						/>
						{#if isAnimating && isFavorite}
							<span class="pointer-events-none absolute inset-0 flex items-center justify-center">
								<Heart size={18} strokeWidth={2} class="animate-ping fill-rose-500" />
							</span>
						{/if}
					</span>
				</Button>
			{/if}
		</div>
	{/if}
</nav>

<!-- DESKTOP HEADER -->
<nav class="hidden items-center justify-between gap-8 py-4 lg:flex">
	<div class="min-w-0 flex-1">
		{#if showTitle}
			<h1 class="truncate text-[22px] font-semibold tracking-tight text-zinc-900">
				{title}
			</h1>
			{#if address}
				<p class="mt-0.5 text-[14px] text-zinc-500">{address}</p>
			{/if}
		{/if}
	</div>

	{#if showActions}
		<div class="flex shrink-0 items-center gap-0.5">
			<ShareControl
				{title}
				appearance="text"
				variant="ghost"
				tone="neutral"
				size="sm"
				radius="xl"
			/>

			{#if canToggleFavorite}
				<Button
					variant="ghost"
					tone={isFavorite ? 'rose' : 'neutral'}
					size="sm"
					radius="xl"
					onclick={toggleFavorite}
					class={isFavorite ? '!bg-rose-50 !text-rose-600' : ''}
					aria-pressed={isFavorite}
				>
					<span class="relative flex items-center justify-center">
						<Heart
							size={15}
							strokeWidth={2}
							class="transition-all duration-300
                                   {isFavorite ? 'fill-rose-500 text-rose-500' : ''}
                                   {isAnimating ? 'scale-125' : 'scale-100'}"
						/>
						{#if isAnimating && isFavorite}
							<span class="pointer-events-none absolute inset-0 flex items-center justify-center">
								<Heart size={15} strokeWidth={2} class="animate-ping fill-rose-500" />
							</span>
						{/if}
					</span>
					{isFavorite ? 'В избранном' : 'Сохранить'}
				</Button>
			{/if}
		</div>
	{/if}
</nav>
