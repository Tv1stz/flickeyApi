<script lang="ts">
	import { Maximize2, Users, Bed, Bath } from 'lucide-svelte';

	interface Props {
		square: number;
		maxGuests: number;
		roomsCount: number;
		bedsCount: number;
		bathroomsCount: number;
		floor?: number;
		totalFloors?: number;
	}

	let { square, maxGuests, roomsCount, bedsCount, bathroomsCount, floor, totalFloors }: Props = $props();
</script>

<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 p-5 rounded-2xl border border-border bg-card shadow-sm">
	<div class="flex flex-col items-center justify-center p-3 rounded-xl bg-muted/30 text-center">
		<Maximize2 class="h-5 w-5 text-primary mb-1" />
		<span class="text-base font-extrabold text-foreground">{square} м²</span>
		<span class="text-[11px] text-muted-foreground">
			{#if floor && totalFloors}
				{floor}/{totalFloors} этаж
			{:else}
				Площадь
			{/if}
		</span>
	</div>

	<div class="flex flex-col items-center justify-center p-3 rounded-xl bg-muted/30 text-center">
		<Users class="h-5 w-5 text-primary mb-1" />
		<span class="text-base font-extrabold text-foreground">{maxGuests}</span>
		<span class="text-[11px] text-muted-foreground">Макс. гостей</span>
	</div>

	<div class="flex flex-col items-center justify-center p-3 rounded-xl bg-muted/30 text-center">
		<Bed class="h-5 w-5 text-primary mb-1" />
		<span class="text-base font-extrabold text-foreground">{roomsCount} / {bedsCount}</span>
		<span class="text-[11px] text-muted-foreground">Комнат / Спальных</span>
	</div>

	<div class="flex flex-col items-center justify-center p-3 rounded-xl bg-muted/30 text-center">
		<Bath class="h-5 w-5 text-primary mb-1" />
		<span class="text-base font-extrabold text-foreground">{bathroomsCount}</span>
		<span class="text-[11px] text-muted-foreground">Санузлов</span>
	</div>
</div>
