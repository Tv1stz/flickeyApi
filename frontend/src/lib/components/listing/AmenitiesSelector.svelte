<!-- src/lib/components/listing/AmenitiesSelector.svelte -->
<script lang="ts">
	import type { AmenityCategory } from '$lib/config/amenities';
	import { getAmenityIcon, fallbackAmenityIcon } from '$lib/config/amenityIcons';

	interface Props {
		categories?: AmenityCategory[];
		selected?: string[];
		onToggle: (id: string) => void;
	}

	let { categories = [], selected = [], onToggle }: Props = $props();

	// O(1) вместо O(n) на каждый isSelected — важно при 65+ удобствах
	const selectedSet = $derived(new Set(selected));

	function isSelected(id: string): boolean {
		return selectedSet.has(id);
	}
</script>

<div class="space-y-10">
	{#each categories as category (category.id)}
		<section>
			<h3 class="mb-1 text-lg font-semibold text-zinc-900 sm:text-xl">
				{category.label}
			</h3>
			{#if category.description}
				<p class="mb-4 text-sm text-zinc-500">{category.description}</p>
			{/if}

			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
				{#each category.amenities as amenity (amenity.id)}
					{@const active = isSelected(amenity.id)}
					{@const Icon = getAmenityIcon(amenity.icon) ?? fallbackAmenityIcon}

					<button
						type="button"
						class="group relative flex h-32 touch-manipulation flex-col justify-between rounded-xl p-4 text-left transition-all duration-200 select-none active:scale-95
						{active
							? 'border-2 border-zinc-900 bg-zinc-50'
							: 'border border-zinc-200 bg-white hover:border-zinc-900'}"
						onclick={() => onToggle(amenity.id)}
						aria-pressed={active}
					>
						<Icon
							class="h-8 w-8 {active ? 'text-zinc-900' : 'text-zinc-700 group-hover:text-zinc-900'}"
							strokeWidth={1.5}
						/>

						<span class="leading-tight font-medium text-zinc-900">
							{amenity.label}
						</span>
					</button>
				{/each}
			</div>
		</section>
	{/each}
</div>
