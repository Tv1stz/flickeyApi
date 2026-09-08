<!-- src/lib/components/listing-page/ListingAmenities.svelte -->
<script lang="ts">
	import type { PropertyType } from '$lib/components/card/types';
	import {
		getAmenityHighlights,
		groupSelectedAmenitiesByCategory,
		type GroupedAmenityCategory
	} from '$lib/config/amenities';
	import {
		fallbackAmenityIcon,
		getAmenityIcon,
		type AmenityIconComponent
	} from '$lib/config/amenityIcons';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';

	type AmenityItemUI = {
		id: string;
		label: string;
		icon: AmenityIconComponent;
	};

	type GroupedAmenityUI = Omit<GroupedAmenityCategory, 'amenities'> & {
		amenities: AmenityItemUI[];
	};

	interface Props {
		amenities?: string[];
		propertyType?: PropertyType;
	}

	let { amenities = [], propertyType = 'apartment' }: Props = $props();

	let modalOpen = $state(false);

	const highlights = $derived(
		getAmenityHighlights(amenities, propertyType, 8).map((item) => ({
			id: item.id,
			label: item.label,
			icon: getAmenityIcon(item.icon) ?? fallbackAmenityIcon
		}))
	);

	const groupedAmenities = $derived(
		groupSelectedAmenitiesByCategory(amenities, propertyType).map(
			(category): GroupedAmenityUI => ({
				...category,
				amenities: category.amenities.map((item) => ({
					id: item.id,
					label: item.label,
					icon: getAmenityIcon(item.icon) ?? fallbackAmenityIcon
				}))
			})
		)
	);

	const totalAmenitiesCount = $derived(
		groupedAmenities.reduce((sum, category) => sum + category.amenities.length, 0)
	);

	const hasMore = $derived(totalAmenitiesCount > highlights.length);
</script>

<div>
	<h2 class="mb-5 text-[20px] font-semibold tracking-tight text-zinc-900">Что вас ждёт</h2>

	{#if totalAmenitiesCount === 0}
		<p class="text-[14px] text-zinc-400">Информация не указана</p>
	{:else}
		<ul class="grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-3">
			{#each highlights as item (item.id)}
				{@const AmenityIcon = item.icon}
				<li class="flex items-center gap-2.5">
					<AmenityIcon size={18} strokeWidth={1.75} class="shrink-0 text-zinc-900" />
					<span class="text-[14px] text-zinc-900">{item.label}</span>
				</li>
			{/each}
		</ul>

		{#if hasMore}
			<div class="mt-6">
				<Button variant="outline" tone="neutral" size="lg" onclick={() => (modalOpen = true)}>
					Показать все {totalAmenitiesCount} удобств
				</Button>
			</div>
		{/if}
	{/if}
</div>

<!-- Modal -->
<Modal open={modalOpen} onClose={() => (modalOpen = false)} title="Все удобства" maxWidth="md">
	<div class="modal-body py-5 sm:py-6">
		<div class="divide-y divide-zinc-100">
			{#each groupedAmenities as category (category.id)}
				<section class="py-5 first:pt-0 last:pb-0">
					<h3 class="mb-3 text-sm font-semibold text-zinc-900">{category.label}</h3>
					<ul class="space-y-2.5">
						{#each category.amenities as item (item.id)}
							{@const AmenityIcon = item.icon}
							<li class="flex items-center gap-3 py-1.5">
								<AmenityIcon size={18} strokeWidth={1.75} class="shrink-0 text-zinc-900" />
								<span class="text-[14px] text-zinc-900">{item.label}</span>
							</li>
						{/each}
					</ul>
				</section>
			{/each}
		</div>
	</div>
</Modal>
