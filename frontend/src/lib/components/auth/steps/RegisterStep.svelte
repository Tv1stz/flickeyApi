<!-- src/lib/components/auth/steps/RegisterStep.svelte -->
<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import TextInputField from '$lib/components/form/TextInputField.svelte';
	import PasswordField from '$lib/components/form/PasswordField.svelte';

	interface Props {
		name: string;
		password: string;
		confirmPassword: string;
		nameTouched: boolean;
		passwordTouched: boolean;
		confirmPasswordTouched: boolean;
		loading: boolean;
		onNameInput: (value: string) => void;
		onNameBlur: () => void;
		onPasswordInput: (value: string) => void;
		onPasswordBlur: () => void;
		onConfirmPasswordInput: (value: string) => void;
		onConfirmPasswordBlur: () => void;
		onSubmit: () => void;
	}

	let {
		name,
		password,
		confirmPassword,
		nameTouched,
		passwordTouched,
		confirmPasswordTouched,
		loading,
		onNameInput,
		onNameBlur,
		onPasswordInput,
		onPasswordBlur,
		onConfirmPasswordInput,
		onConfirmPasswordBlur,
		onSubmit
	}: Props = $props();

	const isNameValid = $derived(name.trim().length >= 2);
	const isPasswordValid = $derived(password.length >= 6);
	const doPasswordsMatch = $derived(password === confirmPassword);

	const nameError = $derived.by(() => {
		if (!nameTouched || !name) return '';
		if (name.trim().length < 2) return 'Минимум 2 символа';
		return '';
	});

	const passwordError = $derived.by(() => {
		if (!passwordTouched || !password) return '';
		if (password.length < 6) return 'Минимум 6 символов';
		return '';
	});

	const confirmPasswordError = $derived.by(() => {
		if (!confirmPasswordTouched || !confirmPassword) return '';
		if (!doPasswordsMatch) return 'Пароли не совпадают';
		return '';
	});

	const canSubmit = $derived(isNameValid && isPasswordValid && doPasswordsMatch && !loading);
</script>

<div class="space-y-5">
	<TextInputField
		id="auth-name"
		label="Ваше имя"
		required
		value={name}
		error={nameError}
		touched={nameTouched}
		valid={isNameValid && nameTouched}
		onInput={onNameInput}
		onBlur={onNameBlur}
	/>

	<PasswordField
		id="auth-password-new"
		label="Пароль"
		required
		value={password}
		error={passwordError}
		touched={passwordTouched}
		valid={isPasswordValid && passwordTouched}
		showStrength
		autocomplete="new-password"
		onInput={onPasswordInput}
		onBlur={onPasswordBlur}
	/>

	<PasswordField
		id="auth-password-confirm"
		label="Повторите пароль"
		required
		value={confirmPassword}
		error={confirmPasswordError}
		touched={confirmPasswordTouched}
		valid={doPasswordsMatch && confirmPasswordTouched && confirmPassword.length > 0}
		autocomplete="new-password"
		onInput={onConfirmPasswordInput}
		onBlur={onConfirmPasswordBlur}
	/>

	<Button
		variant="solid"
		tone="primary"
		size="xl"
		radius="2xl"
		fullWidth
		disabled={!canSubmit}
		{loading}
		loadingText="Создание..."
		onclick={onSubmit}
	>
		Создать аккаунт
	</Button>
</div>
