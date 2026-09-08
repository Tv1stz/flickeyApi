<!-- src/routes/favorites/+page.svelte -->
<script lang="ts">
	import { onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Heart } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';

	import Button from '$lib/components/ui/Button.svelte';
	import ListingCard from '$lib/components/card/ListingCard.svelte';

	import { authStore } from '$lib/stores/authStore.svelte';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import { toast } from '$lib/stores/toastStore';
	import { pluralRu } from '$lib/utils/format';
	import type { Listing } from '$lib/components/card/types';

	// ═══════════════════════════════════════════════════════════════
	// STATE
	// ═══════════════════════════════════════════════════════════════

	let authenticated = $derived(authStore.isAuthenticated);
	let authInitialized = $derived(authStore.initialized);
	let loading = $derived(!authStore.initialized);
	let redirectingToAuth = $state(false);
	let favoritesLoaded = $derived(favoritesStore.ready);
	let favoriteItems = $derived(favoritesStore.favoriteListings);
	
	let removingIds = $state<Set<string>>(new Set());
	let removeTimeouts = $state<Record<string, ReturnType<typeof setTimeout>>>({});

	$effect(() => {
		if (!authStore.initialized || authStore.user || redirectingToAuth) return;
		redirectingToAuth = true;
		authStore.setPendingAction('favorite', '/favorites');
		goto(resolve('/auth'));
	});

	onDestroy(() => {
		for (const timeoutId of Object.values(removeTimeouts)) {
			clearTimeout(timeoutId);
		}
		removeTimeouts = {};
	});

	// ═══════════════════════════════════════════════════════════════
	// DERIVED
	// ═══════════════════════════════════════════════════════════════

	const count = $derived(favoriteItems.length);

	const countLabel = $derived(
		count > 0 ? `${count} ${pluralRu(count, ['объявление', 'объявления', 'объявлений'])}` : ''
	);

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS
	// ═══════════════════════════════════════════════════════════════

	function clearAllFavorites() {
		toast.confirm({
			title: 'Очистить избранное?',
			message: 'Все объявления будут удалены из избранного.',
			confirmText: 'Очистить',
			cancelText: 'Отмена',
			type: 'warning',
			onConfirm: () => {
				favoritesStore.clearAll();
				toast.success('Избранное очищено');
			}
		});
	}

	function handleFavoriteToggle(id: string, isFavorite: boolean): void {
		if (isFavorite) return;

		const listing = favoriteItems.find((item) => item.id === id);
		if (!listing || removingIds.has(id)) return;

		removingIds = new Set([...removingIds, id]);
		const timeoutId = setTimeout(() => {
			favoritesStore.toggle(listing);
			removingIds = new Set([...removingIds].filter((itemId) => itemId !== id));
			delete removeTimeouts[id];
			toast.success('Удалено из избранного');
		}, 250);
		removeTimeouts[id] = timeoutId;
	}
</script>

<svelte:head>
	<title>Избранное — Flickey</title>
</svelte:head>

<div class="min-h-screen bg-white">
	<div class="mx-auto max-w-[1600px] px-4 pt-6 pb-32 sm:px-6 lg:px-8 lg:pb-20">
		<!-- ═══ HEADER ═══ -->
		<div class="mb-6 flex items-center justify-between sm:mb-8">
			<div>
				<h1 class="text-2xl font-bold text-zinc-900 sm:text-3xl">Избранное</h1>
				{#if countLabel}
					<p class="mt-0.5 text-sm text-zinc-400" in:fade={{ duration: 200 }}>
						{countLabel}
					</p>
				{/if}
			</div>

			{#if authenticated && count > 0}
				<div in:fade={{ duration: 200 }}>
					<Button variant="link" tone="neutral" size="sm" onclick={clearAllFavorites}>
						Очистить всё
					</Button>
				</div>
			{/if}
		</div>

		<!-- ═══ LOADING ═══ -->
		{#if loading || !authInitialized || !favoritesLoaded}
			<div
				class="grid grid-cols-2 gap-x-3 gap-y-8
					   sm:gap-x-5 sm:gap-y-10
					   lg:grid-cols-3 xl:grid-cols-4"
			>
				{#each Array.from({ length: 8 }, (_, i) => i) as i (i)}
					<!-- Inline skeleton — компактнее чем на главной -->
					<div class="animate-pulse" aria-hidden="true">
						<div class="overflow-hidden rounded-3xl">
							<div class="relative aspect-[12/13] w-full overflow-hidden bg-zinc-200">
								<div
									class="animate-shimmer absolute inset-0 -translate-x-full
										   bg-gradient-to-r from-transparent via-white/30 to-transparent"
								></div>
							</div>
						</div>
						<div class="space-y-1.5 pt-2.5">
							<div class="h-4 w-3/4 rounded-md bg-zinc-200"></div>
							<div class="h-3.5 w-1/2 rounded-md bg-zinc-100"></div>
							<div class="h-3.5 w-2/5 rounded-md bg-zinc-100"></div>
							<div class="!mt-2 h-4 w-1/3 rounded-md bg-zinc-200"></div>
						</div>
					</div>
				{/each}
			</div>

			<!-- ═══ NOT AUTHENTICATED ═══ -->
		{:else if !authenticated}
			<div
				class="flex flex-col items-center justify-center py-24 text-center"
				in:fade={{ duration: 300 }}
			>
				<div class="mb-5 flex h-20 w-20 items-center justify-center rounded-3xl bg-rose-50">
					<Heart size={34} class="text-rose-300" strokeWidth={1.5} />
				</div>
				<h2 class="mb-2 text-xl font-semibold text-zinc-900">Войдите в аккаунт</h2>
				<p class="mb-8 max-w-xs text-sm leading-relaxed text-zinc-500">
					Чтобы сохранять понравившиеся объявления и возвращаться к ним позже
				</p>
				<Button href={resolve('/auth')} variant="solid" tone="primary" size="lg" radius="pill">
					Войти
				</Button>
			</div>

			<!-- ═══ EMPTY ═══ -->
		{:else if count === 0}
			<div
				class="flex flex-col items-center justify-center py-24 text-center"
				in:fade={{ duration: 300 }}
			>
				<div class="mb-5 flex h-20 w-20 items-center justify-center rounded-3xl bg-zinc-100">
					<Heart size={34} class="text-zinc-300" strokeWidth={1.5} />
				</div>
				<h2 class="mb-2 text-xl font-semibold text-zinc-900">Пока пусто</h2>
				<p class="mb-8 max-w-xs text-sm leading-relaxed text-zinc-500">
					Нажмите на сердечко на карточке объявления, чтобы сохранить его здесь
				</p>
				<Button href={resolve('/')} variant="solid" tone="primary" size="lg" radius="xl">
					Смотреть объявления
				</Button>
			</div>

			<!-- ═══ GRID ═══ -->
		{:else}
			<div
				class="grid grid-cols-2 gap-x-3 gap-y-8
					   sm:gap-x-5 sm:gap-y-10
					   lg:grid-cols-3 xl:grid-cols-4"
			>
				{#each favoriteItems as listing, i (listing.id)}
					{@const isRemoving = removingIds.has(listing.id)}
					<div
						in:fly={{ y: 16, duration: 300, delay: i * 40 }}
						style="opacity: {isRemoving ? 0 : 1};
							   transform: {isRemoving ? 'scale(0.96)' : 'scale(1)'};
							   transition: opacity 250ms ease, transform 250ms ease;"
					>
						<ListingCard
							{listing}
							isFavorite={true}
							onFavoriteToggle={handleFavoriteToggle}
							priority={i < 4}
						/>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style>
	@keyframes shimmer {
		100% {
			transform: translateX(100%);
		}
	}
	.animate-shimmer {
		animation: shimmer 1.5s infinite;
	}
</style>
