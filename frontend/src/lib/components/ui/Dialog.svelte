<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';
	import { X } from 'lucide-svelte';

	interface Props {
		open?: boolean;
		title?: string;
		description?: string;
		class?: string;
		onclose?: () => void;
		children?: Snippet;
		footer?: Snippet;
	}

	let {
		open = $bindable(false),
		title = '',
		description = '',
		class: className = '',
		onclose,
		children,
		footer
	}: Props = $props();

	function close() {
		open = false;
		onclose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open) {
			close();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto"
		role="dialog"
		aria-modal="true"
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity animate-in fade-in"
			onclick={close}
		></div>

		<!-- Dialog Panel -->
		<div
			class={cn(
				'relative z-50 w-full max-w-lg rounded-2xl bg-card p-6 shadow-2xl border border-border transition-all animate-in zoom-in-95 my-8 max-h-[90vh] flex flex-col',
				className
			)}
		>
			<!-- Header -->
			<div class="flex items-start justify-between gap-4 pb-3">
				<div>
					{#if title}
						<h3 class="text-lg font-bold text-foreground leading-none tracking-tight">{title}</h3>
					{/if}
					{#if description}
						<p class="text-sm text-muted-foreground mt-1.5">{description}</p>
					{/if}
				</div>
				<button
					type="button"
					onclick={close}
					class="rounded-full p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground transition-colors cursor-pointer"
					aria-label="Закрыть"
				>
					<X class="h-4 w-4" />
				</button>
			</div>

			<!-- Content body with scroll -->
			<div class="overflow-y-auto py-2 flex-1 pr-1">
				{@render children?.()}
			</div>

			<!-- Footer -->
			{#if footer}
				<div class="mt-4 pt-3 border-t border-border flex items-center justify-end gap-3">
					{@render footer()}
				</div>
			{/if}
		</div>
	</div>
{/if}
