<script lang="ts">
	import type { Component } from 'svelte';
	import { MoreHorizontal } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	export interface ActionMenuItem {
		label: string;
		icon?: any;
		onclick: () => void;
		variant?: 'default' | 'primary' | 'warning' | 'danger' | 'success';
		divider?: boolean;
		disabled?: boolean;
	}

	interface Props {
		items: ActionMenuItem[];
		triggerClass?: string;
		align?: 'left' | 'right';
		ariaLabel?: string;
	}

	let {
		items,
		triggerClass = '',
		align = 'right',
		ariaLabel = 'Действия'
	}: Props = $props();

	let isOpen = $state(false);
	let menuRef = $state<HTMLDivElement | null>(null);

	function toggle(e: MouseEvent) {
		e.stopPropagation();
		e.preventDefault();
		isOpen = !isOpen;
	}

	function close() {
		isOpen = false;
	}

	function handleAction(e: MouseEvent, item: ActionMenuItem) {
		e.stopPropagation();
		e.preventDefault();
		if (item.disabled) return;
		close();
		item.onclick();
	}

	function handleWindowClick(e: MouseEvent) {
		if (isOpen && menuRef && !menuRef.contains(e.target as Node)) {
			close();
		}
	}

	function handleWindowKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && isOpen) {
			close();
		}
	}
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleWindowKeydown} />

<div class="relative inline-block text-left" bind:this={menuRef}>
	<button
		type="button"
		onclick={toggle}
		aria-label={ariaLabel}
		aria-haspopup="true"
		aria-expanded={isOpen}
		class={cn(
			'inline-flex items-center justify-center h-8 w-8 rounded-xl border border-zinc-200/80 bg-white text-zinc-600 transition-all hover:bg-zinc-50 hover:text-zinc-900 hover:border-zinc-300 active:scale-95 shadow-2xs dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground dark:hover:bg-muted/50 cursor-pointer',
			isOpen && 'bg-zinc-100 text-zinc-900 border-zinc-300 dark:bg-muted dark:text-foreground',
			triggerClass
		)}
	>
		<MoreHorizontal class="h-4 w-4" />
	</button>

	{#if isOpen}
		<div
			class={cn(
				'absolute top-full mt-1.5 z-40 min-w-[180px] w-max max-w-xs rounded-2xl border border-zinc-200/80 bg-white p-1.5 shadow-xl transition-all animate-in fade-in zoom-in-95 dark:border-border dark:bg-card',
				align === 'right' ? 'right-0' : 'left-0'
			)}
			role="menu"
		>
			{#each items as item}
				{#if item.divider}
					<div class="my-1 border-t border-zinc-100 dark:border-border"></div>
				{/if}

				{@const Icon = item.icon}
				<button
					type="button"
					role="menuitem"
					disabled={item.disabled}
					onclick={(e) => handleAction(e, item)}
					class={cn(
						'flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-xs font-medium transition-colors text-left cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed',
						item.variant === 'danger'
							? 'text-rose-600 hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-950/50'
							: item.variant === 'warning'
								? 'text-amber-700 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-950/50'
								: item.variant === 'success'
									? 'text-emerald-700 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-950/50'
									: item.variant === 'primary'
										? 'text-zinc-900 font-semibold hover:bg-zinc-100 dark:text-foreground dark:hover:bg-muted'
										: 'text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 dark:text-muted-foreground dark:hover:text-foreground dark:hover:bg-muted'
					)}
				>
					{#if Icon}
						<Icon class="h-3.5 w-3.5 shrink-0" />
					{/if}
					<span class="truncate">{item.label}</span>
				</button>
			{/each}
		</div>
	{/if}
</div>