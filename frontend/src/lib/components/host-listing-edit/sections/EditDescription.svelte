<!-- src/lib/components/host-listing-edit/sections/EditDescription.svelte -->
<script lang="ts">
	import TextareaField from '$lib/components/form/TextareaField.svelte';
	import { cardStyles } from '$lib/config/styles';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	function updateField(field: keyof ListingEditFormData, value: any) {
		onUpdate({ [field]: value });
	}

	const charCount = $derived((form.description || '').length);
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Описание жилья</h2>
	<p class={cardStyles.subtitle}>Расскажите гостям об особенностях вашего жилья</p>

	<div class={cardStyles.contentGap}>
		<TextareaField
			id="description"
			label="Описание"
			placeholder="Опишите ваше жильё, атмосферу, виды из окон, удобства рядом..."
			required
			rows={8}
			value={form.description || ''}
			onInput={(v) => updateField('description', v)}
			onBlur={() => {}}
		/>

		<div class="flex items-center justify-between text-xs text-zinc-500">
			<span>Рекомендуемая длина: от 50 до 1000 символов</span>
			<span class={charCount < 30 ? 'text-amber-600 font-medium' : 'text-zinc-600'}>
				{charCount} символов
			</span>
		</div>
	</div>
</section>
