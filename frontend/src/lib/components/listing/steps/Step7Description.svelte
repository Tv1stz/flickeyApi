<script lang="ts">
	import TextareaField from '$lib/components/form/TextareaField.svelte';
	import { cardStyles } from '$lib/config/styles';

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
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Создайте описание</h2>
	<p class={cardStyles.subtitle}>Расскажите гостям, что делает ваше жильё особенным</p>

	<div class={cardStyles.contentGap}>
		<TextareaField
			id="description"
			label="Описание"
			required
			rows={6}
			placeholder="Расскажите о планировке, особенностях района, что находится рядом и какие правила проживания..."
			value={form.description}
			error={errors.description}
			touched={touched.description}
			onInput={(v) => onInput('description', v)}
			onBlur={() => onBlur('description')}
		/>
	</div>
</section>
