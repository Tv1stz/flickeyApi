<!-- src/lib/components/host-listing-edit/ListingEditNav.svelte -->
<script lang="ts">
	import type { ListingEditNavItem, ListingEditSectionId } from './types';

	interface Props {
		items: ListingEditNavItem[];
		activeSection: ListingEditSectionId;
		completedSections?: ListingEditSectionId[];
		onSelect: (id: ListingEditSectionId) => void;
	}

	let { items, activeSection, completedSections = [], onSelect }: Props = $props();

	function isCompleted(id: ListingEditSectionId): boolean {
		return completedSections.includes(id);
	}
</script>

<nav class="space-y-0.5">
	{#each items as item, index (item.id)}
		{#if index > 0 && items[index - 1].group !== item.group}
			<div class="my-3 h-px bg-zinc-200"></div>
		{/if}

		<button
			type="button"
			class={`group flex w-full items-start gap-3 rounded-lg px-3 py-3 text-left transition-colors xl:gap-3.5 xl:px-3.5 xl:py-3 ${
				activeSection === item.id
					? 'text-zinc-900 bg-zinc-100/70'
					: 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-50'
				}`}
			onclick={() => onSelect(item.id)}
		>
			<div class="relative mt-0.5 shrink-0">
				<item.icon 
					class={`h-[18px] w-[18px] transition-colors xl:h-5 xl:w-5 ${
						activeSection === item.id ? 'text-zinc-900' : 'text-zinc-500 group-hover:text-zinc-700'
					}`} 
					strokeWidth={1.7} 
				/>
				{#if isCompleted(item.id) && activeSection !== item.id}
					<div class="absolute -right-0.5 -bottom-0.5 h-2 w-2 rounded-full bg-emerald-500 ring-2 ring-white"></div>
				{/if}
			</div>
			<div class="min-w-0 flex-1">
				<div class="flex items-center gap-2">
					<span class={`text-[15px] leading-snug ${activeSection === item.id ? 'font-semibold' : 'font-medium'}`}>
						{item.title}
					</span>
					{#if isCompleted(item.id)}
						<svg class="h-4 w-4 text-emerald-500" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
						</svg>
					{/if}
				</div>
				<p class="mt-0.5 text-[13px] leading-snug text-zinc-500">
					{item.subtitle}
				</p>
			</div>
		</button>
	{/each}
</nav>
