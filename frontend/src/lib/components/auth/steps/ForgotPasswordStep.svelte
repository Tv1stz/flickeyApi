<!-- src/lib/components/auth/steps/ForgotPasswordStep.svelte -->
<script lang="ts">
	import { ArrowRight } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import PhoneInputField from '$lib/components/form/PhoneInputField.svelte';
	import { validatePhoneLength, type Country } from '$lib/config/countries';

	interface Props {
		phone: string;
		country: Country;
		touched: boolean;
		loading: boolean;
		onInput: (phone: string, country: Country) => void;
		onBlur: () => void;
		onSubmit: () => void;
	}

	let { phone, country, touched, loading, onInput, onBlur, onSubmit }: Props = $props();

	const isValid = $derived(validatePhoneLength(phone, country));

	const error = $derived.by(() => {
		if (!touched || !phone) return '';
		if (!isValid) return 'Введите полный номер телефона';
		return '';
	});

	const canSubmit = $derived(isValid && !loading);
</script>

<div class="space-y-6">
	<PhoneInputField
		id="reset-phone"
		label="Номер телефона"
		required
		value={phone}
		{country}
		{error}
		{touched}
		valid={isValid && touched}
		helpText="Введите номер, указанный при регистрации"
		{onInput}
		{onBlur}
	/>

	<Button
		variant="solid"
		tone="primary"
		size="xl"
		radius="2xl"
		fullWidth
		disabled={!canSubmit}
		{loading}
		loadingText="Отправка..."
		iconRight={ArrowRight}
		onclick={onSubmit}
	>
		Отправить код
	</Button>
</div>
