<script lang="ts">
	import { Users, DoorOpen, Bed, Bath, Maximize2, Building } from 'lucide-svelte';

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

	function getGuestWord(n: number): string {
		if (n === 1) return 'гость';
		if (n >= 2 && n <= 4) return 'гостя';
		return 'гостей';
	}

	function getRoomWord(n: number): string {
		if (n === 1) return 'комната';
		if (n >= 2 && n <= 4) return 'комнаты';
		return 'комнат';
	}

	function getBedWord(n: number): string {
		if (n === 1) return 'кровать';
		if (n >= 2 && n <= 4) return 'кровати';
		return 'кроватей';
	}

	function getBathWord(n: number): string {
		if (n === 1) return 'ванная';
		if (n >= 2 && n <= 4) return 'ванные';
		return 'ванных';
	}
</script>

<!-- Clean specs bar matching screenshot -->
<div class="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs font-medium text-foreground/80 py-1">
	<div class="flex items-center gap-1.5">
		<Users class="h-4 w-4 text-foreground/70 shrink-0" />
		<span>{maxGuests} {getGuestWord(maxGuests)}</span>
	</div>

	<div class="flex items-center gap-1.5">
		<DoorOpen class="h-4 w-4 text-foreground/70 shrink-0" />
		<span>{roomsCount} {getRoomWord(roomsCount)}</span>
	</div>

	<div class="flex items-center gap-1.5">
		<Bed class="h-4 w-4 text-foreground/70 shrink-0" />
		<span>{bedsCount} {getBedWord(bedsCount)}</span>
	</div>

	<div class="flex items-center gap-1.5">
		<Bath class="h-4 w-4 text-foreground/70 shrink-0" />
		<span>{bathroomsCount} {getBathWord(bathroomsCount)}</span>
	</div>

	<div class="flex items-center gap-1.5">
		<Maximize2 class="h-4 w-4 text-foreground/70 shrink-0" />
		<span>{square} м² площадь</span>
	</div>

	{#if floor}
		<div class="flex items-center gap-1.5">
			<Building class="h-4 w-4 text-foreground/70 shrink-0" />
			<span>{floor}{totalFloors ? '/' + totalFloors : ''} этаж</span>
		</div>
	{/if}
</div>
