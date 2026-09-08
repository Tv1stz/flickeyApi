<!-- src/lib/components/auth/steps/PhoneStep.svelte -->
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
		id="auth-phone"
		label="Номер телефона"
		required
		value={phone}
		{country}
		{error}
		{touched}
		valid={isValid && touched}
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
		Продолжить
	</Button>

	<!-- DEV: Test accounts -->
	{#if import.meta.env.DEV}
		<details class="group">
			<summary
				class="flex cursor-pointer items-center gap-1
                           text-xs text-zinc-400 transition-colors hover:text-zinc-600"
			>
				<span>Тестовые аккаунты</span>
				<svg
					class="h-3 w-3 transition-transform group-open:rotate-180"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M19 9l-7 7-7-7"
					/>
				</svg>
			</summary>
			<div class="mt-3 space-y-2 rounded-xl bg-zinc-50 p-3 text-xs">
				<div class="grid grid-cols-2 gap-2 text-zinc-600">
					<div>
						<code class="rounded bg-zinc-200 px-1.5 py-0.5 text-[10px]">291111111</code>
						<span class="mt-0.5 block text-[10px] text-zinc-400">guest123</span>
					</div>
					<div>
						<code class="rounded bg-zinc-200 px-1.5 py-0.5 text-[10px]">293333333</code>
						<span class="mt-0.5 block text-[10px] text-zinc-400">host123</span>
					</div>
				</div>
				<p class="border-t border-zinc-200 pt-1 text-zinc-500">
					Код: <code class="rounded bg-zinc-200 px-1.5 py-0.5">000000</code>
				</p>
			</div>
		</details>
	{/if}
</div>
