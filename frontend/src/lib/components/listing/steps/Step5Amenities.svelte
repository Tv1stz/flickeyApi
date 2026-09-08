<!-- src/lib/components/listing/steps/Step5Amenities.svelte -->
<script lang="ts">
	import type { ListingFormValues } from '$lib/validation/listingValidation';
	import type {
		AmenityCategory,
		AmenityPresetId,
		AmenityPresetOption
	} from '$lib/config/amenities';
	import type { PropertyType } from '$lib/components/card/types';
	import AmenitiesSelector from '$lib/components/listing/AmenitiesSelector.svelte';
	import { cardStyles } from '$lib/config/styles';
	import { pluralRu } from '$lib/utils/format';
	import { Building2, House, Castle } from 'lucide-svelte';

	interface Props {
		form: ListingFormValues;
		propertyType: PropertyType;
		categories: AmenityCategory[];
		presetOptions?: AmenityPresetOption[];
		error?: string;
		onToggleAmenity: (id: string) => void;
		onApplyPreset?: (presetId: AmenityPresetId) => void;
	}

	let {
		form,
		propertyType,
		categories,
		presetOptions = [],
		error = '',
		onToggleAmenity,
		onApplyPreset
	}: Props = $props();

	const selectedCount = $derived(form.amenities.length);

	const propertyTypeLabel: Record<PropertyType, { label: string; Icon: typeof Building2 }> = {
		apartment: { label: 'Квартира', Icon: Building2 },
		house: { label: 'Дом', Icon: House },
		estate: { label: 'Усадьба', Icon: Castle }
	};

	const currentTypeInfo = $derived(propertyTypeLabel[propertyType]);
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Какие удобства вы предлагаете?</h2>
	<p class={cardStyles.subtitle}>
		Отметьте всё, что есть в вашем жилье. Вы сможете добавить другие удобства позже.
	</p>

	<div class={cardStyles.contentGap}>
		<div class="rounded-2xl border border-zinc-200 bg-zinc-50 p-4">
			<!-- Тип жилья + счётчик -->
			<div class="flex items-center justify-between gap-3">
				<div class="flex items-center gap-2 text-sm text-zinc-500">
					{#if currentTypeInfo}
						{@const TypeIcon = currentTypeInfo.Icon}
						<TypeIcon class="h-4 w-4 shrink-0" />
						<span
							>Удобства для: <span class="font-medium text-zinc-700">{currentTypeInfo.label}</span
							></span
						>
					{/if}
				</div>

				<p class="shrink-0 text-sm font-medium text-zinc-700">
					{selectedCount}
					{pluralRu(selectedCount, ['удобство', 'удобства', 'удобств'])}
				</p>
			</div>

			{#if presetOptions.length > 0}
				<div class="mt-3 flex flex-wrap gap-2">
					{#each presetOptions as preset (preset.id)}
						<button
							type="button"
							class="rounded-full border border-zinc-300 bg-white px-3 py-1.5 text-xs font-medium text-zinc-700 transition-colors hover:border-zinc-400 hover:text-zinc-900"
							onclick={() => onApplyPreset?.(preset.id)}
						>
							{preset.label}
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<AmenitiesSelector {categories} selected={form.amenities} onToggle={onToggleAmenity} />

		{#if error}
			<p class="text-sm font-medium text-red-600">{error}</p>
		{/if}
	</div>
</section>
