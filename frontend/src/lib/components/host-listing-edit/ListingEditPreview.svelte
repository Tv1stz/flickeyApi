<!-- src/lib/components/host-listing-edit/ListingEditPreview.svelte -->
<script lang="ts">
	import type { Listing } from '$lib/components/card/types';
	import { formatBYN } from '$lib/utils/format';
	import { Eye } from 'lucide-svelte';
	import StatusIndicator from '$lib/components/ui/StatusIndicator.svelte';

	interface Props {
		listing: Listing | null;
		onPreview?: () => void;
	}

	let { listing, onPreview }: Props = $props();
</script>

{#if listing}
	<div class="space-y-3">
		<!-- Listing Thumbnail -->
		<div class="relative aspect-[4/3] overflow-hidden rounded-xl bg-zinc-100">
			{#if listing.images?.[0]}
				<img
					src={listing.images[0]}
					alt={listing.title}
					class="h-full w-full object-cover"
				/>
			{:else}
				<div class="flex h-full items-center justify-center bg-zinc-100">
					<span class="text-sm text-zinc-400">Нет фото</span>
				</div>
			{/if}
			
			<!-- Status Badge -->
			<div class="absolute top-2 left-2">
				<StatusIndicator status={listing.status} active={listing.isActive} size="sm" />
			</div>

			<!-- Preview Button -->
			{#if onPreview}
				<button
					type="button"
					class="absolute right-2 bottom-2 flex h-8 items-center gap-1.5 rounded-lg bg-white/95 backdrop-blur-sm px-3 text-[12px] font-medium text-zinc-900 shadow-md transition-all hover:bg-white active:scale-95"
					onclick={onPreview}
				>
					<Eye class="h-3.5 w-3.5" strokeWidth={1.8} />
					<span>Просмотр</span>
				</button>
			{/if}
		</div>

		<!-- Listing Info -->
		<div class="space-y-1">
			<h3 class="line-clamp-2 text-[15px] font-semibold text-zinc-900 leading-snug">
				{listing.title}
			</h3>
			<p class="text-[13px] text-zinc-500">
				{listing.location?.city || listing.address}
			</p>
			<div class="flex items-baseline gap-1 pt-1">
				<span class="text-lg font-bold text-zinc-900">
					{formatBYN(listing.pricePerNight)}
				</span>
				<span class="text-[13px] text-zinc-500">/ ночь</span>
			</div>
		</div>

		<!-- Quick Stats -->
		<div class="grid grid-cols-3 gap-2 pt-2">
			<div class="rounded-lg bg-zinc-50 px-2 py-2 text-center">
				<p class="text-[12px] font-semibold text-zinc-900">{listing.maxGuests}</p>
				<p class="text-[11px] text-zinc-500">гостей</p>
			</div>
			<div class="rounded-lg bg-zinc-50 px-2 py-2 text-center">
				<p class="text-[12px] font-semibold text-zinc-900">{listing.bedrooms}</p>
				<p class="text-[11px] text-zinc-500">спален</p>
			</div>
			<div class="rounded-lg bg-zinc-50 px-2 py-2 text-center">
				<p class="text-[12px] font-semibold text-zinc-900">{listing.bathrooms}</p>
				<p class="text-[11px] text-zinc-500">санузлов</p>
			</div>
		</div>
	</div>
{:else}
	<div class="space-y-3">
		<div class="relative aspect-[4/3] overflow-hidden rounded-xl bg-zinc-100">
			<div class="flex h-full flex-col items-center justify-center gap-2 text-zinc-400">
				<svg class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
				<span class="text-sm">Загрузка...</span>
			</div>
		</div>
	</div>
{/if}
