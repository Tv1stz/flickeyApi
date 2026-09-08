<script lang="ts">
	import { PROPERTY_TYPE_OPTIONS } from '../taxonomy';
	import type { PropertyType } from '$lib/components/card/types';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface Props {
		value: PropertyType | 'any';
		isOpen: boolean;
		onSelect: (value: PropertyType | 'any') => void;
		onFocus: () => void;
		compact?: boolean;
	}

	let { value, isOpen, onSelect, onFocus, compact = false }: Props = $props();

	const selectedOption = $derived(
		PROPERTY_TYPE_OPTIONS.find((o) => o.id === value) ?? PROPERTY_TYPE_OPTIONS[0]
	);
</script>

<div class="relative min-w-0 flex-1">
	<button
		type="button"
		onclick={onFocus}
		class="flex w-full cursor-pointer flex-col items-start justify-center focus:outline-none
               {compact ? 'px-3 py-1' : 'px-5 py-2.5'}"
	>
		{#if !compact}
			<span class="mb-0.5 text-[14px] font-medium text-zinc-800 select-none"> Тип жилья </span>
		{/if}
		<span
			class="w-full truncate text-left font-medium
                     {value === 'any' ? 'text-zinc-400' : 'text-zinc-900'}
                     {compact ? 'text-[12px]' : 'text-[14.5px]'}"
		>
			{compact && value === 'any' ? 'Тип жилья' : selectedOption.label}
		</span>
	</button>

	{#if isOpen}
		<div
			class="dropdown-surface-lg absolute top-full left-0 z-[100] mt-3 w-[280px] py-3"
			transition:fly={{ y: -8, duration: 220, easing: cubicOut }}
		>
			<div class="px-2">
				{#each PROPERTY_TYPE_OPTIONS as option (option.id)}
					{@const isSelected = option.id === value}
					<button
						type="button"
						onclick={(e) => {
							e.stopPropagation();
							onSelect(option.id as PropertyType | 'any');
						}}
						class="flex w-full touch-manipulation items-center gap-4 rounded-2xl px-4 py-3 text-left transition-colors
                               {isSelected
							? 'bg-zinc-900 text-white'
							: 'hover:bg-zinc-50 active:bg-zinc-100'}"
					>
						<span class="text-[22px] leading-none">{option.emoji}</span>
						<div class="min-w-0 flex-1">
							<p class="text-[14.5px] font-semibold {isSelected ? 'text-white' : 'text-zinc-900'}">
								{option.label}
							</p>
						</div>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style></style>
