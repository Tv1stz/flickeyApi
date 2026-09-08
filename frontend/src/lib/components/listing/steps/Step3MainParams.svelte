<script lang="ts">
	import TextInputField from '$lib/components/form/TextInputField.svelte';
	import CounterField from '$lib/components/form/CounterField.svelte';
	import { AlertCircle } from 'lucide-svelte';
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

	const isApartment = $derived(form.propertyType === 'apartment');
</script>

<section class={cardStyles.sectionDivided}>
	<h2 class={cardStyles.title}>Название объявления</h2>
	<p class={cardStyles.subtitle}>Придумайте короткий и понятный заголовок</p>

	<div class={cardStyles.contentGap}>
		<TextInputField
			id="title"
			label="Название объявления"
			placeholder="Уютная квартира с видом на парк"
			required
			value={form.title}
			error={errors.title}
			touched={touched.title}
			valid={touched.title && !errors.title && form.title.trim() !== ''}
			onInput={(v) => onInput('title', v)}
			onBlur={() => onBlur('title')}
		/>
	</div>
</section>

<section class={cardStyles.sectionDivided}>
	<h2 class={cardStyles.title}>Основные характеристики</h2>
	<p class={cardStyles.subtitle}>Укажите площадь, этажность и вместимость</p>

	<div class={cardStyles.contentGap}>
		<div class={formStyles.fieldGroup}>
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-5">
				<TextInputField
					id="size"
					label="Площадь"
					required
					numeric
					placeholder="45"
					suffix="м²"
					value={form.size}
					error={errors.size}
					touched={touched.size}
					valid={touched.size && !errors.size && form.size !== ''}
					onInput={(v) => onInput('size', v)}
					onBlur={() => onBlur('size')}
				/>

				{#if isApartment}
					<div class="space-y-2">
						<div class="flex items-end gap-3">
							<div class="flex-1">
								<TextInputField
									id="floor"
									label="Этаж"
									required
									numeric
									placeholder="5"
									value={form.floor}
									error={errors.floor}
									touched={touched.floor}
									valid={touched.floor && !errors.floor && form.floor !== ''}
									hideErrorMessage
									onInput={(v) => onInput('floor', v)}
									onBlur={() => onBlur('floor')}
								/>
							</div>
							<span class="pb-3.5 font-medium text-zinc-400">из</span>
							<div class="flex-1">
								<TextInputField
									id="totalFloors"
									label="Всего"
									required
									numeric
									placeholder="9"
									value={form.totalFloors}
									error={errors.totalFloors}
									touched={touched.totalFloors}
									valid={touched.totalFloors && !errors.totalFloors && form.totalFloors !== ''}
									hideErrorMessage
									onInput={(v) => onInput('totalFloors', v)}
									onBlur={() => onBlur('totalFloors')}
								/>
							</div>
						</div>

						{#if (touched.floor && errors.floor) || (touched.totalFloors && errors.totalFloors)}
							<div class={formStyles.errorBlock}>
								<AlertCircle class="h-5 w-5 shrink-0 text-red-500" />
								<p class="text-sm font-medium text-red-700">{errors.floor || errors.totalFloors}</p>
							</div>
						{/if}
					</div>
				{:else}
					<TextInputField
						id="totalFloors"
						label="Количество этажей"
						required
						numeric
						placeholder="2"
						value={form.totalFloors}
						error={errors.totalFloors}
						touched={touched.totalFloors}
						valid={touched.totalFloors && !errors.totalFloors && form.totalFloors !== ''}
						onInput={(v) => onInput('totalFloors', v)}
						onBlur={() => onBlur('totalFloors')}
					/>
				{/if}
			</div>

			<TextInputField
				id="maxGuests"
				label="Максимум гостей"
				required
				numeric
				placeholder="4"
				value={form.maxGuests}
				error={errors.maxGuests}
				touched={touched.maxGuests}
				valid={touched.maxGuests && !errors.maxGuests && form.maxGuests !== ''}
				onInput={(v) => onInput('maxGuests', v)}
				onBlur={() => onBlur('maxGuests')}
			/>
		</div>
	</div>
</section>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Сколько комнат и спальных мест?</h2>
	<p class={cardStyles.subtitle}>Гости хотят знать, где они будут спать</p>

	<div class={cardStyles.contentGap}>
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-3 sm:gap-6">
			<CounterField
				id="bedrooms"
				label="Комнат"
				required
				min={1}
				step={1}
				value={form.bedrooms}
				error={errors.bedrooms}
				touched={touched.bedrooms}
				onInput={(v) => onInput('bedrooms', v)}
				onBlur={() => onBlur('bedrooms')}
			/>
			<CounterField
				id="beds"
				label="Спальных мест"
				required
				min={1}
				step={1}
				value={form.beds}
				error={errors.beds}
				touched={touched.beds}
				onInput={(v) => onInput('beds', v)}
				onBlur={() => onBlur('beds')}
			/>
			<CounterField
				id="bathrooms"
				label="Ванных комнат"
				required
				min={1}
				step={1}
				value={form.bathrooms}
				error={errors.bathrooms}
				touched={touched.bathrooms}
				onInput={(v) => onInput('bathrooms', v)}
				onBlur={() => onBlur('bathrooms')}
			/>
		</div>
	</div>
</section>
