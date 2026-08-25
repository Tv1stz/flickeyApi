<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import { formatCurrency } from '$lib/utils';
	import { Sparkles, Edit3, ArrowLeft, Send } from 'lucide-svelte';

	interface Props {
		pricePerNight: number;
		currency?: string;
		minNights?: number;
		mode?: 'public' | 'preview';
		onpublish?: () => void;
		onedit?: () => void;
	}

	let {
		pricePerNight = 100,
		currency = 'BYN',
		minNights = 1,
		mode = 'public',
		onpublish,
		onedit
	}: Props = $props();

	let nightsCount = $state<number>(1);
	$effect(() => {
		nightsCount = minNights || 1;
	});

	let totalPrice = $derived(pricePerNight * nightsCount);
</script>

<div class="sticky top-24 rounded-3xl border border-border bg-card p-6 shadow-xl space-y-6">
	{#if mode === 'preview'}
		<div class="rounded-xl bg-amber-500/10 border border-amber-500/20 p-3 text-center text-xs font-bold text-amber-700 dark:text-amber-400">
			✨ Режим предпросмотра черновика
		</div>
	{/if}

	<div>
		<span class="text-xs text-muted-foreground font-medium">Стоимость</span>
		<div class="flex items-baseline gap-1 mt-1">
			<span class="text-3xl font-extrabold text-foreground tracking-tight">
				{formatCurrency(pricePerNight, currency)}
			</span>
			<span class="text-xs text-muted-foreground">/ сутки</span>
		</div>
	</div>

	<div class="rounded-2xl border border-border bg-muted/20 p-4 space-y-3">
		<div class="flex items-center justify-between text-xs">
			<span class="text-muted-foreground">Количество ночей:</span>
			<div class="flex items-center gap-2">
				<button
					type="button"
					onclick={() => (nightsCount = Math.max(minNights || 1, nightsCount - 1))}
					class="h-7 w-7 rounded-lg border border-border bg-card font-bold hover:bg-accent cursor-pointer"
				>
					-
				</button>
				<span class="font-bold text-sm min-w-[20px] text-center">{nightsCount}</span>
				<button
					type="button"
					onclick={() => (nightsCount = nightsCount + 1)}
					class="h-7 w-7 rounded-lg border border-border bg-card font-bold hover:bg-accent cursor-pointer"
				>
					+
				</button>
			</div>
		</div>

		<div class="pt-2 border-t border-border/60 flex items-center justify-between text-xs">
			<span class="text-muted-foreground">{pricePerNight} {currency} × {nightsCount} ноч.</span>
			<span class="font-bold text-foreground">{totalPrice} {currency}</span>
		</div>
	</div>

	{#if mode === 'public'}
		<Button class="w-full h-12 text-base font-bold shadow-lg gap-2">
			<Sparkles class="h-4 w-4" /> Забронировать онлайн
		</Button>
		<p class="text-center text-[11px] text-muted-foreground">
			Оплата происходит после подтверждения бронирования арендодателем
		</p>
	{:else}
		<div class="space-y-2">
			<Button onclick={onpublish} class="w-full h-12 text-base font-bold bg-emerald-600 hover:bg-emerald-700 text-white shadow-lg gap-2">
				<Send class="h-4 w-4" /> Опубликовать
			</Button>
			<Button variant="outline" onclick={onedit} class="w-full gap-2">
				<ArrowLeft class="h-4 w-4" /> Вернуться к редактированию
			</Button>
		</div>
	{/if}
</div>
