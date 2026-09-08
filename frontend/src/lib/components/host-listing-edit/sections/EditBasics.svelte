<!-- src/lib/components/host-listing-edit/sections/EditBasics.svelte -->
<script lang="ts">
	import TextInputField from '$lib/components/form/TextInputField.svelte';
	import CounterField from '$lib/components/form/CounterField.svelte';
	import { cardStyles, formStyles } from '$lib/config/styles';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	const isApartment = $derived(form.propertyType === 'apartment');

	function updateField(field: keyof ListingEditFormData, value: any) {
		onUpdate({ [field]: value });
	}
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
			value={form.title || ''}
			onInput={(v) => updateField('title', v)}
			onBlur={() => {}}
		/>
	</div>
</section>

<section class={cardStyles.sectionDivided}>
	<h2 class={cardStyles.title}>Основные параметры</h2>
	<p class={cardStyles.subtitle}>Укажите вместительность и количество комнат</p>

	<div class={cardStyles.contentGap}>
		<div class={formStyles.fieldGroup}>
			<CounterField
				id="maxGuests"
				label="Максимальное количество гостей"
				min={1}
				max={20}
				value={String(form.maxGuests || 1)}
				onInput={(v) => updateField('maxGuests', parseInt(v) || 1)}
				onBlur={() => {}}
			/>

			<div class={formStyles.fieldRow} style="grid-template-columns: repeat(3, 1fr);">
				<CounterField
					id="bedrooms"
					label="Спальни"
					min={0}
					max={10}
					value={String(form.bedrooms || 0)}
					onInput={(v) => updateField('bedrooms', parseInt(v) || 0)}
					onBlur={() => {}}
				/>
				<CounterField
					id="beds"
					label="Кровати"
					min={1}
					max={20}
					value={String(form.beds || 1)}
					onInput={(v) => updateField('beds', parseInt(v) || 1)}
					onBlur={() => {}}
				/>
				<CounterField
					id="bathrooms"
					label="Санузлы"
					min={0}
					max={10}
					value={String(form.bathrooms || 0)}
					onInput={(v) => updateField('bathrooms', parseInt(v) || 0)}
					onBlur={() => {}}
				/>
			</div>

			{#if isApartment}
				<div class={formStyles.fieldRow} style="grid-template-columns: 1fr 1fr;">
					<CounterField
						id="floor"
						label="Этаж"
						min={1}
						max={50}
						value={String(form.floor || 1)}
						onInput={(v) => updateField('floor', parseInt(v) || 1)}
						onBlur={() => {}}
					/>
					<CounterField
						id="totalFloors"
						label="Этажей в доме"
						min={1}
						max={50}
						value={String(form.totalFloors || 1)}
						onInput={(v) => updateField('totalFloors', parseInt(v) || 1)}
						onBlur={() => {}}
					/>
				</div>
			{:else}
				<CounterField
					id="totalFloors"
					label="Этажей в доме"
					min={1}
					max={50}
					value={String(form.totalFloors || 1)}
					onInput={(v) => updateField('totalFloors', parseInt(v) || 1)}
					onBlur={() => {}}
				/>
			{/if}

			<CounterField
				id="area"
				label="Площадь (м²)"
				min={10}
				max={1000}
				value={String(form.area || 10)}
				onInput={(v) => updateField('area', parseInt(v) || 10)}
				onBlur={() => {}}
			/>
		</div>
	</div>
</section>
