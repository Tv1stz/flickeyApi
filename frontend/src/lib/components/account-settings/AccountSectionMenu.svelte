<script lang="ts">
	import { ChevronRight } from 'lucide-svelte';
	import type { AccountNavItem, AccountSectionId } from './types';

	interface Props {
		items: AccountNavItem[];
		onSelect: (id: AccountSectionId) => void;
	}

	let { items, onSelect }: Props = $props();
</script>

<nav class="divide-y divide-zinc-100">
	{#each items as item, index (item.id)}
		{@const Icon = item.icon}
		{#if index > 0 && items[index - 1].group !== item.group}
			<div class="py-2"></div>
		{/if}

		<button
			type="button"
			class="group flex w-full items-center gap-3 py-3.5 text-left transition-colors active:opacity-70 sm:py-4"
			onclick={() => onSelect(item.id)}
		>
			<Icon class="h-[18px] w-[18px] shrink-0 text-zinc-500" strokeWidth={1.8} />
			<span class="min-w-0 flex-1 text-[15px] font-medium text-zinc-900">{item.title}</span>
			<ChevronRight class="h-[18px] w-[18px] shrink-0 text-zinc-400" strokeWidth={1.8} />
		</button>
	{/each}
</nav>
