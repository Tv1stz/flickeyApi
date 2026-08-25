<script lang="ts">
	import { Clock, CalendarCheck, Check, X } from 'lucide-svelte';
	import type { ListingRules as ListingRulesType } from '$lib/types/listings';

	interface Props {
		checkinFrom?: string;
		checkoutUntil?: string;
		minNights?: number;
		rules?: ListingRulesType;
	}

	let {
		checkinFrom = '14:00',
		checkoutUntil = '12:00',
		minNights = 1,
		rules
	}: Props = $props();
</script>

<div class="space-y-4 pt-4 border-t border-border">
	<h3 class="text-lg font-bold text-foreground">Правила и порядок проживания</h3>

	<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
		<div class="flex items-center gap-3 p-3.5 rounded-xl border border-border bg-card">
			<Clock class="h-5 w-5 text-primary shrink-0" />
			<div class="text-xs">
				<span class="text-muted-foreground">Время заезда / выезда:</span>
				<div class="font-bold text-foreground">Заезд с {checkinFrom} • Выезд до {checkoutUntil}</div>
			</div>
		</div>

		<div class="flex items-center gap-3 p-3.5 rounded-xl border border-border bg-card">
			<CalendarCheck class="h-5 w-5 text-primary shrink-0" />
			<div class="text-xs">
				<span class="text-muted-foreground">Минимальный срок аренды:</span>
				<div class="font-bold text-foreground">{minNights} {minNights === 1 ? 'сутки' : 'суток'}</div>
			</div>
		</div>
	</div>

	{#if rules}
		<div class="grid grid-cols-2 sm:grid-cols-3 gap-2.5 pt-2">
			<div class="flex items-center gap-2 p-2.5 rounded-xl bg-muted/20 text-xs">
				{#if rules.allow_children}
					<Check class="h-4 w-4 text-emerald-600" />
				{:else}
					<X class="h-4 w-4 text-destructive" />
				{/if}
				<span>Можно с детьми</span>
			</div>

			<div class="flex items-center gap-2 p-2.5 rounded-xl bg-muted/20 text-xs">
				{#if rules.allow_pets}
					<Check class="h-4 w-4 text-emerald-600" />
				{:else}
					<X class="h-4 w-4 text-destructive" />
				{/if}
				<span>Можно с питомцами</span>
			</div>

			<div class="flex items-center gap-2 p-2.5 rounded-xl bg-muted/20 text-xs">
				{#if rules.allow_smoking}
					<Check class="h-4 w-4 text-emerald-600" />
				{:else}
					<X class="h-4 w-4 text-destructive" />
				{/if}
				<span>Разрешено курение</span>
			</div>

			<div class="flex items-center gap-2 p-2.5 rounded-xl bg-muted/20 text-xs">
				{#if rules.allow_parties}
					<Check class="h-4 w-4 text-emerald-600" />
				{:else}
					<X class="h-4 w-4 text-destructive" />
				{/if}
				<span>Вечеринки / праздники</span>
			</div>
		</div>
	{/if}
</div>
