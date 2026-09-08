<script lang="ts">
	import type { AccountNavItem, AccountSectionId } from './types';

	interface Props {
		items: AccountNavItem[];
		activeSection: AccountSectionId;
		onSelect: (id: AccountSectionId) => void;
	}

	let { items, activeSection, onSelect }: Props = $props();
</script>

<nav class="space-y-1">
	{#each items as item, index (item.id)}
		{@const Icon = item.icon}
		{#if index > 0 && items[index - 1].group !== item.group}
			<div class="my-2 h-px bg-zinc-200"></div>
		{/if}

		<button
			type="button"
			class={`group flex w-full items-center gap-3 rounded-xl px-3.5 py-2.5 text-left text-[15px] font-medium transition-colors ${
				activeSection === item.id
					? 'bg-zinc-100 font-semibold text-zinc-900 shadow-sm'
					: 'text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900'
			}`}
			onclick={() => onSelect(item.id)}
		>
			<Icon
				class={`h-4 w-4 shrink-0 transition-colors ${
					activeSection === item.id ? 'text-zinc-900' : 'text-zinc-400 group-hover:text-zinc-700'
				}`}
				strokeWidth={2}
			/>
			<span>{item.title}</span>
		</button>
	{/each}
</nav>
