<!-- src/lib/components/card/ListingGrid.svelte -->
<script lang="ts">
	import ListingCard from './ListingCard.svelte';
	import ListingSkeleton from './ListingSkeleton.svelte';
	import { Home } from 'lucide-svelte';
	import type { ListingGridProps } from './types';

	let {
		listings = [],
		isLoading = false,
		skeletonCount = 8,
		emptyMessage = 'Пока здесь пусто',
		emptyDescription = 'Объявления скоро появятся',
		favorites = new Set<string>(),
		onFavoriteToggle,
		layout = 'default'
	}: ListingGridProps = $props();

	let visibleItems = $state<Set<number>>(new Set());
	let gridRef: HTMLElement | undefined = $state();

	// Intersection Observer для анимации появления
	$effect(() => {
		if (!gridRef || isLoading || listings.length === 0) return;

		const observer = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					const index = Number(entry.target.getAttribute('data-index'));
					if (entry.isIntersecting && !isNaN(index)) {
						visibleItems = new Set([...visibleItems, index]);
					}
				});
			},
			{
				rootMargin: '50px',
				threshold: 0.1
			}
		);

		const items = gridRef.querySelectorAll('[data-index]');
		items.forEach((item) => observer.observe(item));

		return () => observer.disconnect();
	});

	function isFavorite(id: string): boolean {
		return favorites.has(id);
	}

	const gridClass = $derived(
		layout === 'two-column'
			? 'grid grid-cols-1 gap-x-6 gap-y-12 sm:grid-cols-2 sm:gap-x-8 sm:gap-y-14 lg:grid-cols-2'
			: 'grid grid-cols-1 gap-x-6 gap-y-12 sm:grid-cols-2 sm:gap-x-8 sm:gap-y-14 lg:grid-cols-3 xl:grid-cols-4'
	);
</script>

<section aria-label="Список объявлений" aria-busy={isLoading}>
	{#if isLoading}
		<div class={gridClass}>
			{#each Array.from({ length: skeletonCount }, (_, i) => i) as i (i)}
				<div
					class="animate-in fade-in slide-in-from-bottom-4 fill-mode-both duration-500"
					style="animation-delay: {i * 75}ms"
				>
					<ListingSkeleton />
				</div>
			{/each}
		</div>
	{:else if listings.length === 0}
		<div class="flex flex-col items-center justify-center py-24 text-center sm:py-32">
			<div
				class="mb-6 flex h-20 w-20 items-center
                        justify-center rounded-3xl bg-zinc-100"
			>
				<Home size={32} strokeWidth={1.5} class="text-zinc-400" />
			</div>
			<h3 class="mb-2 text-xl font-semibold text-zinc-800">{emptyMessage}</h3>
			<p class="max-w-sm text-base text-zinc-500">{emptyDescription}</p>
		</div>
	{:else}
		<ul bind:this={gridRef} class="m-0 list-none p-0 {gridClass}">
			{#each listings as listing, i (listing.id)}
				<li
					data-index={i}
					class="transition-all duration-500 ease-out
                           {visibleItems.has(i)
						? 'translate-y-0 opacity-100'
						: 'translate-y-6 opacity-0'}"
					style="transition-delay: {Math.min(i * 50, 300)}ms"
				>
					<ListingCard {listing} isFavorite={isFavorite(listing.id)} {onFavoriteToggle} />
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style>
	@keyframes fade-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes slide-in-from-bottom-4 {
		from {
			transform: translateY(1rem);
		}
		to {
			transform: translateY(0);
		}
	}

	.animate-in {
		animation-name: fade-in, slide-in-from-bottom-4;
		animation-timing-function: cubic-bezier(0.16, 1, 0.3, 1);
	}

	.fill-mode-both {
		animation-fill-mode: both;
	}
</style>
