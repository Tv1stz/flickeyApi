<script lang="ts">
	import ImageCarousel from '$lib/components/card/ImageCarousel.svelte';
	import type { Listing } from '$lib/components/card/types';
	import { PROPERTY_TYPE_LABELS } from '$lib/components/card/types';
	import { formatBYN, pluralRu } from '$lib/utils/format';
	import { formatListingAddressCompact } from '$lib/utils/location';
	import { Heart, X } from 'lucide-svelte';

	interface Props {
		listing: Listing;
		onClose: () => void;
		onOpen: () => void;
		isFavorite?: boolean;
		onFavoriteToggle?: () => void;
	}

	let { listing, onClose, onOpen, isFavorite = false, onFavoriteToggle }: Props = $props();

	const propertyLabel = $derived(PROPERTY_TYPE_LABELS[listing.propertyType] ?? 'Жильё');
	const locationLine = $derived(
		[formatListingAddressCompact(listing.address, listing.location), propertyLabel]
			.filter(Boolean)
			.join(', ')
	);
	const detailsLine = $derived(
		`${listing.bedrooms} ${pluralRu(listing.bedrooms, ['комната', 'комнаты', 'комнат'])} · ${listing.beds} ${pluralRu(listing.beds, ['кровать', 'кровати', 'кроватей'])} · ${listing.bathrooms} ${pluralRu(listing.bathrooms, ['ванная', 'ванные', 'ванных'])}`
	);
</script>

<div
	class="w-[min(310px,calc(100vw-32px))] cursor-pointer overflow-hidden rounded-2xl bg-white shadow-[0_6px_24px_rgba(0,0,0,0.16)] ring-1 ring-black/10 transition-all hover:shadow-[0_8px_28px_rgba(0,0,0,0.20)]"
	role="button"
	tabindex="0"
	aria-label="Краткий просмотр объявления"
	onclick={onOpen}
	onkeydown={(e) => {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onOpen();
		}
	}}
>
	<div class="group/carousel relative">
		<div role="none" onclick={(e) => e.stopPropagation()}>
			<ImageCarousel
				images={listing.images}
				alt={listing.title}
				maxImages={8}
				showArrows={listing.images.length > 1}
				aspectRatio="video"
			/>
		</div>
		<div class="pointer-events-none absolute inset-x-0 top-0 z-20 flex justify-end gap-2 p-2.5">
			{#if onFavoriteToggle}
				<button
					type="button"
					class="pointer-events-auto flex h-8 w-8 items-center justify-center rounded-full bg-white/95 shadow-md backdrop-blur-sm transition hover:scale-105 active:scale-95"
					aria-label={isFavorite ? 'Убрать из избранного' : 'В избранное'}
					onclick={(e) => {
						e.stopPropagation();
						onFavoriteToggle();
					}}
				>
					<Heart
						size={18}
						strokeWidth={1.75}
						class={isFavorite ? 'fill-rose-500 text-rose-500' : 'text-zinc-700'}
					/>
				</button>
			{/if}
			<button
				type="button"
				class="pointer-events-auto flex h-8 w-8 items-center justify-center rounded-full bg-white/95 shadow-md backdrop-blur-sm transition hover:scale-105 active:scale-95"
				aria-label="Закрыть"
				onclick={(e) => {
					e.stopPropagation();
					onClose();
				}}
			>
				<X size={18} strokeWidth={2} class="text-zinc-700" />
			</button>
		</div>
	</div>

	<div class="px-3.5 pb-3.5 pt-3 text-left">
		<div class="flex items-start justify-between gap-2">
			<p class="line-clamp-1 text-xs font-medium text-zinc-600">{locationLine}</p>
		</div>
		<h3 class="mt-1 line-clamp-2 text-[15px] font-semibold leading-snug text-zinc-900">
			{listing.title}
		</h3>
		<p class="mt-0.5 text-xs text-zinc-500">{detailsLine}</p>
		<p class="mt-2 text-sm font-semibold text-zinc-900">
			<span class="underline decoration-zinc-300 decoration-1 underline-offset-2">
				{formatBYN(listing.pricePerNight)}
			</span>
			<span class="font-normal text-zinc-600"> за ночь</span>
		</p>
	</div>
</div>
