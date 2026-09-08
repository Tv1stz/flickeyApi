<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import { formatCurrency } from '$lib/utils';
	import { ArrowLeft, Send } from 'lucide-svelte';

	interface Props {
		pricePerNight: number;
		currency?: string;
		minNights?: number;
		mode?: 'public' | 'preview';
		onpublish?: () => void;
		onedit?: () => void;
		oncontact?: () => void;
	}

	let {
		pricePerNight,
		currency = 'BYN',
		minNights = 1,
		mode = 'public',
		onpublish,
		onedit,
		oncontact
	}: Props = $props();

	function getNightWord(n: number): string {
		if (n === 1) return 'ночь';
		if (n >= 2 && n <= 4) return 'ночи';
		return 'ночей';
	}
</script>

<div class="rounded-3xl border border-border/80 bg-card p-6 shadow-xs space-y-4">
	{#if mode === 'preview'}
		<div class="rounded-xl bg-amber-500/10 border border-amber-500/20 p-2.5 text-center text-xs font-bold text-amber-700 dark:text-amber-400">
			✨ Режим предпросмотра
		</div>
	{/if}

	<!-- Dynamic Price Headline matching screenshot style -->
	<div>
		<div class="flex items-baseline gap-1.5">
			<span class="text-xs font-medium text-muted-foreground">от</span>
			<span class="text-2xl sm:text-3xl font-extrabold text-foreground tracking-tight">
				{pricePerNight} {currency}
			</span>
			<span class="text-xs text-muted-foreground font-normal">/ ночь</span>
		</div>
		{#if minNights && minNights > 1}
			<p class="text-[11px] text-muted-foreground mt-0.5 font-normal">
				мин. {minNights} {getNightWord(minNights)}
			</p>
		{/if}
	</div>

	<!-- Primary Action Button matching screenshot style -->
	{#if mode === 'public'}
		<button
			type="button"
			onclick={oncontact}
			class="w-full py-3.5 px-4 rounded-xl bg-neutral-900 hover:bg-neutral-800 dark:bg-white dark:hover:bg-neutral-100 text-white dark:text-neutral-900 font-bold text-xs sm:text-sm transition-all cursor-pointer shadow-xs active:scale-[0.99] text-center block"
		>
			Связаться с хозяином
		</button>
	{:else}
		<div class="space-y-2 pt-1">
			<Button onclick={onpublish} class="w-full h-11 text-xs sm:text-sm font-bold bg-emerald-600 hover:bg-emerald-700 text-white shadow-xs gap-2">
				<Send class="h-4 w-4" /> Опубликовать
			</Button>
			<Button variant="outline" onclick={onedit} class="w-full h-10 text-xs gap-2">
				<ArrowLeft class="h-3.5 w-3.5" /> Вернуться к редактированию
			</Button>
		</div>
	{/if}
</div>
