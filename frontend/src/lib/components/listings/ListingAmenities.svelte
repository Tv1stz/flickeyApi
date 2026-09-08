<script lang="ts">
	import {
		Wifi,
		Flame,
		UtensilsCrossed,
		Droplets,
		BedDouble,
		Sparkles,
		Refrigerator,
		Bell,
		Tv,
		Car,
		Wind,
		Waves,
		CheckCircle2
	} from 'lucide-svelte';
	import { translateAmenity } from '$lib/utils';

	interface Props {
		amenities?: string[];
	}

	let { amenities = [] }: Props = $props();

	let isExpanded = $state(false);

	let activeAmenities = $derived(amenities || []);
	let displayedAmenities = $derived(isExpanded ? activeAmenities : activeAmenities.slice(0, 8));

	function getAmenityIcon(key: string) {
		const lower = key.toLowerCase();
		if (lower.includes('wifi') || lower.includes('вайфай') || lower.includes('интернет')) return Wifi;
		if (lower.includes('heat') || lower.includes('отоплен')) return Flame;
		if (lower.includes('kitch') || lower.includes('кухн') || lower.includes('посуд')) return UtensilsCrossed;
		if (lower.includes('water') || lower.includes('вод') || lower.includes('душ')) return Droplets;
		if (lower.includes('bed') || lower.includes('бель') || lower.includes('кроват')) return BedDouble;
		if (lower.includes('towel') || lower.includes('полотенц')) return Sparkles;
		if (lower.includes('fridge') || lower.includes('refriger') || lower.includes('холодильн')) return Refrigerator;
		if (lower.includes('smoke') || lower.includes('дым') || lower.includes('сигнал')) return Bell;
		if (lower.includes('tv') || lower.includes('телевизор')) return Tv;
		if (lower.includes('park') || lower.includes('парковк')) return Car;
		if (lower.includes('air') || lower.includes('кондиционер')) return Wind;
		if (lower.includes('pool') || lower.includes('бассейн')) return Waves;
		return CheckCircle2;
	}
</script>

{#if activeAmenities.length > 0}
	<div class="space-y-4 pt-6 border-t border-border/60">
		<h2 class="text-base sm:text-lg font-bold text-foreground">Что вас ждёт</h2>

		<!-- 3-Column Amenities Grid matching screenshot style with dynamic items -->
		<div class="grid grid-cols-2 sm:grid-cols-3 gap-y-3.5 gap-x-4">
			{#each displayedAmenities as item}
				{@const IconComponent = getAmenityIcon(item)}
				<div class="flex items-center gap-2.5 text-xs text-foreground/85 font-normal">
					<IconComponent class="h-4 w-4 text-foreground/70 shrink-0" />
					<span class="truncate">{translateAmenity(item)}</span>
				</div>
			{/each}
		</div>

		{#if activeAmenities.length > 8}
			<div class="pt-1">
				<button
					type="button"
					onclick={() => (isExpanded = !isExpanded)}
					class="rounded-xl border border-border/80 bg-background px-4 py-2 text-xs font-semibold text-foreground hover:bg-muted/40 transition-colors cursor-pointer"
				>
					{isExpanded ? 'Скрыть часть удобств' : `Показать все ${activeAmenities.length} удобств`}
				</button>
			</div>
		{/if}
	</div>
{/if}
