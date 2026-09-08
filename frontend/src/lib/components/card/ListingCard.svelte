<script lang="ts">
	import { onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import ImageCarousel from './ImageCarousel.svelte';
	import type { ListingCardProps } from './types';
	import { Heart } from 'lucide-svelte';
	import { formatBYN, pluralRu } from '$lib/utils/format';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import { toast } from '$lib/stores/toastStore';
	import { formatListingAddressFull } from '$lib/utils/location';

	let {
		listing,
		isFavorite: initialFavorite = false,
		onFavoriteToggle
	}: ListingCardProps = $props();

	// ═══════════════════════════════════════════════════════════════
	// STATE
	// ═══════════════════════════════════════════════════════════════

	let isAnimating = $state(false);
	let authenticated = $derived(authStore.isAuthenticated);
	let favorites = $derived(favoritesStore.favoriteIds);
	let favoritesLoaded = $derived(favoritesStore.ready);
	let currentUserId = $derived(authStore.user?.id ?? null);
	let favoriteAnimationTimeout: ReturnType<typeof setTimeout> | null = null;

	// ═══════════════════════════════════════════════════════════════
	// DERIVED
	// ═══════════════════════════════════════════════════════════════

	const isFavoriteComputed = $derived(
		favoritesLoaded ? favorites.has(listing.id) : Boolean(initialFavorite)
	);

	const guestsText = $derived(
		`${listing.maxGuests} ${pluralRu(listing.maxGuests, ['гость', 'гостя', 'гостей'])}`
	);

	const roomsText = $derived(
		`${listing.bedrooms} ${pluralRu(listing.bedrooms, ['комната', 'комнаты', 'комнат'])}`
	);

	const bedsText = $derived(
		`${listing.beds} ${pluralRu(listing.beds, ['кровать', 'кровати', 'кроватей'])}`
	);

	const addressText = $derived(formatListingAddressFull(listing.address, listing.location));
	const canToggleFavorite = $derived(listing.ownerId !== currentUserId);

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS
	// ═══════════════════════════════════════════════════════════════

	function toggleFavorite(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();

		if (!authenticated) {
			authStore.setPendingAction('favorite', `/listings/${listing.id}`);
			toast.info('Войдите в аккаунт', 'Чтобы добавлять объявления в избранное');
			goto(resolve('/auth'));
			return;
		}

		const nextFavorite = favoritesStore.toggle(listing);
		isAnimating = true;

		if (typeof window !== 'undefined' && 'vibrate' in navigator) {
			navigator.vibrate(10);
		}

		onFavoriteToggle?.(listing.id, nextFavorite);

		if (favoriteAnimationTimeout) clearTimeout(favoriteAnimationTimeout);
		favoriteAnimationTimeout = setTimeout(() => {
			isAnimating = false;
			favoriteAnimationTimeout = null;
		}, 400);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			toggleFavorite(e as unknown as MouseEvent);
		}
	}

	onDestroy(() => {
		if (favoriteAnimationTimeout) {
			clearTimeout(favoriteAnimationTimeout);
			favoriteAnimationTimeout = null;
		}
	});
</script>

<a
	href={resolve('/listings/[id]', { id: listing.id })}
	class="group block rounded-4xl outline-none focus-visible:ring-2 focus-visible:ring-zinc-900/20 focus-visible:ring-offset-2"
	draggable="false"
>
	<article class="relative contain-layout">
		<!-- Image container -->
		<div
			class="relative overflow-hidden rounded-4xl shadow-sm transition-shadow duration-300 ease-out group-hover:shadow-md"
		>
			<ImageCarousel images={listing.images} alt={listing.title} maxImages={5} />

			<!-- Favorite button -->
			{#if canToggleFavorite}
				<button
					type="button"
					onclick={toggleFavorite}
					onkeydown={handleKeydown}
					class="absolute top-3 right-3 z-20 flex h-10 w-10 touch-manipulation items-center justify-center rounded-full border backdrop-blur-md transition-all duration-200 ease-out hover:scale-110 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:outline-none active:scale-95
		                {isFavoriteComputed
						? 'border-white/60 bg-white/85 hover:bg-white'
						: 'border-white/50 bg-zinc-200/50 hover:bg-zinc-200/90'}"
					aria-label={isFavoriteComputed ? 'Удалить из избранного' : 'Добавить в избранное'}
					aria-pressed={isFavoriteComputed}
				>
					<Heart
						size={22}
						strokeWidth={1.75}
						class="transition-all duration-200
	                       {isFavoriteComputed ? 'fill-rose-500 text-rose-500' : 'text-white'}
	                       {isAnimating ? 'scale-125' : 'scale-100'}"
					/>

					{#if isAnimating && isFavoriteComputed}
						<span class="pointer-events-none absolute inset-0 flex items-center justify-center">
							<Heart size={22} strokeWidth={1.75} class="animate-ping fill-rose-500 opacity-75" />
						</span>
					{/if}
				</button>
			{/if}
		</div>

		<!-- Content -->
		<div class="pt-3 sm:pt-3.5">
			<h3
				class="line-clamp-1 text-base leading-snug font-semibold text-zinc-900 transition-colors duration-150 group-hover:text-zinc-700 sm:text-lg"
			>
				{listing.title}
			</h3>

			<p class="mt-0.5 line-clamp-1 text-sm text-zinc-500 sm:mt-1 sm:text-base">
				{addressText}
			</p>

			<p class="mt-0.5 text-sm text-zinc-400 sm:mt-1">
				{guestsText} · {roomsText} · {bedsText}
			</p>

			<div class="mt-2 flex items-baseline gap-1 sm:mt-2.5">
				<span class="text-lg font-bold text-zinc-900 sm:text-xl">
					от {formatBYN(listing.pricePerNight)}
				</span>
				<span class="text-sm text-zinc-500 sm:text-base">/ ночь</span>
			</div>
		</div>
	</article>
</a>
