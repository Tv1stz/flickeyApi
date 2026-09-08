<!-- src/lib/components/host-listing-edit/sections/EditPricing.svelte -->
<script lang="ts">
	import TextInputField from '$lib/components/form/TextInputField.svelte';
	import { cardStyles, formStyles } from '$lib/config/styles';
	import { formatBYN } from '$lib/utils/format';
	import { Banknote } from 'lucide-svelte';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	function updateField(field: keyof ListingEditFormData, value: any) {
		onUpdate({ [field]: value });
	}

	const formattedPrice = $derived(form.pricePerNight ? formatBYN(form.pricePerNight) : '');
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Стоимость проживания</h2>
	<p class={cardStyles.subtitle}>Укажите базовую цену за одну ночь</p>

	<div class={cardStyles.contentGap}>
		<div class={formStyles.fieldRow} style="grid-template-columns: 2fr 1fr;">
			<TextInputField
				id="pricePerNight"
				label="Цена за ночь"
				placeholder="100"
				type="number"
				required
				value={String(form.pricePerNight || '')}
				onInput={(v) => updateField('pricePerNight', parseFloat(v) || 0)}
				onBlur={() => {}}
			/>

			<TextInputField
				id="currency"
				label="Валюта"
				value={form.currency || 'BYN'}
				onInput={(v) => updateField('currency', v)}
				onBlur={() => {}}
				disabled
			/>
		</div>

		{#if form.pricePerNight && form.pricePerNight > 0}
			<div class="flex items-center gap-3 rounded-2xl bg-zinc-50 p-4">
				<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-zinc-200/60 text-zinc-700">
					<Banknote class="h-5 w-5" />
				</div>
				<div>
					<p class="text-xs text-zinc-500">Предпросмотр для гостей:</p>
					<p class="text-sm font-semibold text-zinc-900">
						<span class="text-base font-bold">{formattedPrice}</span> за 1 ночь
					</p>
				</div>
			</div>
		{/if}
	</div>
</section>
