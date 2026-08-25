<script lang="ts">
	import { toast } from './toast.svelte';
	import { X, CheckCircle2, AlertCircle, Info, AlertTriangle } from 'lucide-svelte';
	import { cn } from '$lib/utils';
</script>

<div
	class="fixed bottom-0 right-0 z-50 flex max-h-screen w-full flex-col-reverse p-4 sm:bottom-0 sm:right-0 sm:top-auto sm:flex-col md:max-w-[420px] gap-2 pointer-events-none"
	aria-live="polite"
>
	{#each toast.toasts as item (item.id)}
		<div
			class={cn(
				'pointer-events-auto flex w-full items-start gap-3 overflow-hidden rounded-xl border p-4 shadow-xl transition-all animate-in slide-in-from-bottom-5 duration-200',
				item.type === 'success' && 'bg-card border-emerald-500/30 text-card-foreground',
				item.type === 'error' && 'bg-card border-destructive/40 text-card-foreground',
				item.type === 'warning' && 'bg-card border-amber-500/30 text-card-foreground',
				(!item.type || item.type === 'default' || item.type === 'info') &&
					'bg-card border-border text-card-foreground'
			)}
		>
			<div class="mt-0.5 shrink-0">
				{#if item.type === 'success'}
					<CheckCircle2 class="h-5 w-5 text-emerald-600" />
				{:else if item.type === 'error'}
					<AlertCircle class="h-5 w-5 text-destructive" />
				{:else if item.type === 'warning'}
					<AlertTriangle class="h-5 w-5 text-amber-600" />
				{:else}
					<Info class="h-5 w-5 text-primary" />
				{/if}
			</div>

			<div class="grid flex-1 gap-1">
				{#if item.title}
					<div class="text-sm font-semibold text-foreground">{item.title}</div>
				{/if}
				{#if item.description}
					<div class="text-xs text-muted-foreground leading-relaxed">{item.description}</div>
				{/if}
			</div>

			<button
				type="button"
				onclick={() => toast.dismiss(item.id)}
				class="rounded-md p-1 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
				aria-label="Закрыть уведомление"
			>
				<X class="h-4 w-4" />
			</button>
		</div>
	{/each}
</div>
