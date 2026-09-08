<script lang="ts">
	import { AlertCircle, House, Building2, Castle, Check } from 'lucide-svelte';
	import { cardStyles, formStyles } from '$lib/config/styles';

	import type {
		ListingFormValues,
		ListingErrors,
		ListingTouched,
		ValidatedListingField
	} from '$lib/validation/listingValidation';

	interface Props {
		form: ListingFormValues;
		errors: ListingErrors;
		touched: ListingTouched;
		onInput: (
			field: Exclude<ValidatedListingField, 'photos' | 'contactMethods'>,
			value: string
		) => void;
		onBlur: (field: ValidatedListingField) => void;
	}

	let { form, errors, touched, onInput, onBlur }: Props = $props();

	const propertyTypeOptions = [
		{
			id: 'apartment',
			label: 'Квартира',
			description: 'Отдельная квартира в многоквартирном доме',
			Icon: Building2
		},
		{ id: 'house', label: 'Дом', description: 'Отдельный дом или коттедж', Icon: House },
		{ id: 'estate', label: 'Усадьба', description: 'Загородная усадьба или поместье', Icon: Castle }
	] as const;

	type PropertyType = (typeof propertyTypeOptions)[number]['id'];

	function selectPropertyType(type: PropertyType): void {
		onInput('propertyType', type);
		onBlur('propertyType');
	}
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Какой тип жилья вы сдаёте?</h2>
	<p class={cardStyles.subtitle}>Выберите формат, который лучше всего описывает ваше жильё</p>

	<div class={cardStyles.contentGap}>
		<div id="propertyType" tabindex="-1" class="space-y-3">
			{#each propertyTypeOptions as option (option.id)}
				{@const active = form.propertyType === option.id}
				{@const TypeIcon = option.Icon}
				<button
					type="button"
					onclick={() => selectPropertyType(option.id)}
					class="flex h-[88px] w-full touch-manipulation items-center gap-4 rounded-3xl border p-4 text-left transition-all duration-200
						focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-900 focus-visible:ring-offset-2 sm:h-[98px] sm:p-5
						{active ? 'border-zinc-900 bg-zinc-50' : 'border-zinc-200 bg-white hover:border-zinc-400'}"
				>
					<div
						class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl sm:h-14 sm:w-14
							{active ? 'bg-zinc-900 text-white' : 'text-zinc-900'}"
					>
						<TypeIcon class="h-6 w-6" />
					</div>

					<div class="min-w-0 flex-1">
						<p class="text-base font-semibold text-zinc-900 sm:text-lg">{option.label}</p>
						<p class="mt-0.5 text-sm text-zinc-500">{option.description}</p>
					</div>

					{#if active}
						<div class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-zinc-900">
							<Check class="h-4 w-4 text-white" strokeWidth={3} />
						</div>
					{/if}
				</button>
			{/each}

			{#if touched.propertyType && errors.propertyType}
				<div class={formStyles.errorBlock}>
					<AlertCircle class="h-5 w-5 shrink-0 text-red-500" />
					<p class="text-sm font-medium text-red-700">{errors.propertyType}</p>
				</div>
			{/if}
		</div>
	</div>
</section>
