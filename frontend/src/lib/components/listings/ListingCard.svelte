<script lang="ts">
	import type { ListingPublic } from '$lib/types/listings';
	import { formatCurrency, translateHousingType } from '$lib/utils';
	import { Users, Bed, Bath, Maximize2, ShieldCheck, Image as ImageIcon } from 'lucide-svelte';

	interface Props {
		listing: ListingPublic;
	}

	let { listing }: Props = $props();

	// Fallback placeholder images for nice visuals
	const fallbackImages: Record<string, string> = {
		apartment: 'https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?w=800&auto=format&fit=crop&q=80',
		house: 'https://images.unsplash.com/photo-1580587771525-78b9dba3b914?w=800&auto=format&fit=crop&q=80',
		manor: 'https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=800&auto=format&fit=crop&q=80'
	};

	let primaryPhoto = $derived(
		listing.media && listing.media.length > 0
			? listing.media[0]
			: fallbackImages[listing.type] || fallbackImages.apartment
	);
</script>

<a
	href={`/listings/${listing.id}`}
	class="group flex flex-col overflow-hidden rounded-2xl border border-border bg-card shadow-sm hover:shadow-xl hover:border-primary/40 transition-all duration-300 transform hover:-translate-y-1"
>
	<!-- Card Image Banner -->
	<div class="relative aspect-[4/3] w-full overflow-hidden bg-muted">
		<img
			src={primaryPhoto}
			alt={listing.name}
			loading="lazy"
			class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
		/>

		<!-- Top Badges -->
		<div class="absolute top-3 left-3 flex items-center gap-1.5">
			<span class="rounded-full bg-background/90 backdrop-blur-md px-2.5 py-1 text-xs font-bold text-foreground shadow-sm">
				{translateHousingType(listing.type)}
			</span>
		</div>

		<div class="absolute top-3 right-3">
			<span class="rounded-full bg-emerald-500/90 text-white backdrop-blur-md px-2 py-0.5 text-[11px] font-semibold flex items-center gap-1 shadow-sm">
				<ShieldCheck class="h-3 w-3" /> Проверено
			</span>
		</div>
	</div>

	<!-- Content Body -->
	<div class="flex flex-1 flex-col p-4 sm:p-5">
		<!-- Title -->
		<h3 class="font-bold text-base text-foreground line-clamp-1 group-hover:text-primary transition-colors">
			{listing.name}
		</h3>

		<!-- Physical Parameters Line -->
		<div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
			<span class="flex items-center gap-1">
				<Maximize2 class="h-3.5 w-3.5" />
				<span>{listing.square} м²</span>
			</span>
			<span>•</span>
			<span class="flex items-center gap-1">
				<Users class="h-3.5 w-3.5" />
				<span>до {listing.max_guests} гостей</span>
			</span>
			<span>•</span>
			<span class="flex items-center gap-1">
				<Bed class="h-3.5 w-3.5" />
				<span>{listing.rooms_count} комн. ({listing.beds_count} спальн.)</span>
			</span>
		</div>

		<!-- Description snippet -->
		{#if listing.description}
			<p class="mt-2.5 text-xs text-muted-foreground line-clamp-2 leading-relaxed">
				{listing.description}
			</p>
		{/if}

		<!-- Footer with Pricing -->
		<div class="mt-auto pt-4 border-t border-border/80 flex items-end justify-between">
			<div>
				<div class="text-[11px] text-muted-foreground">Цена за сутки</div>
				<div class="text-lg font-extrabold text-foreground tracking-tight">
					{formatCurrency(listing.price_per_night, listing.currency)}
				</div>
			</div>

			<div class="text-right text-[11px] text-muted-foreground">
				мин. {listing.min_nights} {listing.min_nights === 1 ? 'сутки' : 'суток'}
			</div>
		</div>
	</div>
</a>
