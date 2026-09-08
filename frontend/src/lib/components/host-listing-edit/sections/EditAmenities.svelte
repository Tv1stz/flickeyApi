<!-- src/lib/components/host-listing-edit/sections/EditAmenities.svelte -->
<script lang="ts">
	import AmenitiesSelector from '$lib/components/listing/AmenitiesSelector.svelte';
	import { cardStyles } from '$lib/config/styles';
	import { getAmenityCategoriesForPropertyType } from '$lib/config/amenities';
	import type { PropertyType } from '$lib/components/card/types';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	const categories = $derived(
		getAmenityCategoriesForPropertyType((form.propertyType as PropertyType) || 'apartment')
	);

	function handleToggleAmenity(id: string) {
		const current = form.amenities || [];
		const updated = current.includes(id)
			? current.filter((a) => a !== id)
			: [...current, id];
		onUpdate({ amenities: updated });
	}
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Удобства</h2>
	<p class={cardStyles.subtitle}>Отметьте всё, что доступно гостям в вашем жилье</p>

	<div class={cardStyles.contentGap}>
		<AmenitiesSelector
			selected={form.amenities || []}
			{categories}
			onToggle={handleToggleAmenity}
		/>
	</div>
</section>
