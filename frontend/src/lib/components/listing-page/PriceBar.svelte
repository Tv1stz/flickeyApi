<!-- src/lib/components/listing-page/PriceBar.svelte -->
<script lang="ts">
	import { formatBYN, pluralRu } from '$lib/utils/format';
	import type { OwnerContacts } from '$lib/components/card/types';

	interface Props {
		pricePerNight: number;
		minNights?: number;
		selectedNights?: number;
		ownerContacts?: OwnerContacts;
		mode?: 'default' | 'preview';
		onOpenContacts?: () => void;
	}

	let {
		pricePerNight,
		minNights = 1,
		selectedNights = 0,
		ownerContacts,
		mode = 'default',
		onOpenContacts
	}: Props = $props();

	const hasContacts = $derived(
		ownerContacts &&
			(ownerContacts.phone ||
				ownerContacts.email ||
				ownerContacts.telegram ||
				ownerContacts.viber ||
				ownerContacts.whatsapp ||
				ownerContacts.signal)
	);

	const nightsText = $derived(
		selectedNights > 0
			? `${selectedNights} ${pluralRu(selectedNights, ['ночь', 'ночи', 'ночей'])} (${formatBYN(pricePerNight)} / ночь)`
			: `мин. ${minNights} ${pluralRu(minNights ?? 1, ['ночь', 'ночи', 'ночей'])}`
	);
</script>

<div
	class="pointer-events-none fixed right-0 bottom-0 left-0 z-[999] px-4"
	data-listing-price-bar
	style="padding-bottom: max(0.75rem, env(safe-area-inset-bottom));"
>
	<div
		class="pointer-events-auto mx-auto flex max-w-lg items-center justify-between gap-4
	               rounded-3xl border border-zinc-200/60 bg-white/90 px-5 py-3
	               shadow-lg"
		style="backdrop-filter: blur(24px) saturate(180%);
               -webkit-backdrop-filter: blur(24px) saturate(180%);"
	>
		<!-- Price -->
		<div class="flex flex-col">
			{#if selectedNights > 0}
				<div class="flex items-baseline gap-1">
					<span class="text-[19px] font-bold tracking-tight text-zinc-900">
						{formatBYN(pricePerNight * selectedNights)}
					</span>
					<span class="text-[13px] text-zinc-500 font-medium"> за {selectedNights} {pluralRu(selectedNights, ['ночь', 'ночи', 'ночей'])}</span>
				</div>
				<p class="mt-0.5 text-[11px] text-zinc-400">{formatBYN(pricePerNight)} / ночь</p>
			{:else}
				<div class="flex items-baseline gap-1">
					<span class="text-[19px] font-bold tracking-tight text-zinc-900">
						от {formatBYN(pricePerNight)}
					</span>
					<span class="text-[13px] text-zinc-400"> / ночь </span>
				</div>
				{#if minNights > 1}
					<p class="mt-0.5 text-[11px] text-zinc-400">{nightsText}</p>
				{/if}
			{/if}
		</div>

		<!-- CTA -->
		{#if hasContacts && onOpenContacts && mode === 'default'}
			<button
				onclick={onOpenContacts}
				class="relative h-11 shrink-0 overflow-hidden rounded-xl bg-zinc-900
	                       px-6 text-[14px] font-semibold
	                       text-white shadow-md
	                       transition-all
	                       duration-150 hover:bg-zinc-800
	                       active:scale-[0.96]"
				style="cursor: pointer;"
			>
				Связаться
			</button>
		{/if}
	</div>
</div>
