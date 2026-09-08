<!-- src/lib/components/auth/steps/PasswordStep.svelte -->
<script lang="ts">
	import { ArrowRight } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import PasswordField from '$lib/components/form/PasswordField.svelte';

	interface Props {
		password: string;
		touched: boolean;
		loading: boolean;
		onInput: (value: string) => void;
		onBlur: () => void;
		onSubmit: () => void;
		onForgotPassword: () => void;
	}

	let { password, touched, loading, onInput, onBlur, onSubmit, onForgotPassword }: Props = $props();

	const isValid = $derived(password.length >= 6);

	const error = $derived.by(() => {
		if (!touched || !password) return '';
		if (password.length < 6) return 'Минимум 6 символов';
		return '';
	});

	const canSubmit = $derived(isValid && !loading);
</script>

<div class="space-y-6">
	<PasswordField
		id="auth-password"
		label="Пароль"
		required
		value={password}
		{error}
		{touched}
		valid={isValid && touched}
		autocomplete="current-password"
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
		loadingText="Вход..."
		iconRight={ArrowRight}
		onclick={onSubmit}
	>
		Войти
	</Button>

	<div class="text-center">
		<button
			type="button"
			onclick={onForgotPassword}
			class="text-sm font-medium text-zinc-500 transition-colors
                   duration-200 hover:text-zinc-900"
		>
			Забыли пароль?
		</button>
	</div>
</div>
