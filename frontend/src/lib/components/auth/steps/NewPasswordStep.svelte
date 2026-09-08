<!-- src/lib/components/auth/steps/NewPasswordStep.svelte -->
<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import PasswordField from '$lib/components/form/PasswordField.svelte';

	interface Props {
		newPassword: string;
		confirmNewPassword: string;
		newPasswordTouched: boolean;
		confirmNewPasswordTouched: boolean;
		loading: boolean;
		onNewPasswordInput: (value: string) => void;
		onNewPasswordBlur: () => void;
		onConfirmNewPasswordInput: (value: string) => void;
		onConfirmNewPasswordBlur: () => void;
		onSubmit: () => void;
	}

	let {
		newPassword,
		confirmNewPassword,
		newPasswordTouched,
		confirmNewPasswordTouched,
		loading,
		onNewPasswordInput,
		onNewPasswordBlur,
		onConfirmNewPasswordInput,
		onConfirmNewPasswordBlur,
		onSubmit
	}: Props = $props();

	const isValid = $derived(newPassword.length >= 6);
	const doPasswordsMatch = $derived(newPassword === confirmNewPassword);

	const newPasswordError = $derived.by(() => {
		if (!newPasswordTouched || !newPassword) return '';
		if (newPassword.length < 6) return 'Минимум 6 символов';
		return '';
	});

	const confirmNewPasswordError = $derived.by(() => {
		if (!confirmNewPasswordTouched || !confirmNewPassword) return '';
		if (!doPasswordsMatch) return 'Пароли не совпадают';
		return '';
	});

	const canSubmit = $derived(isValid && doPasswordsMatch && !loading);
</script>

<div class="space-y-5">
	<PasswordField
		id="new-password"
		label="Новый пароль"
		required
		value={newPassword}
		error={newPasswordError}
		touched={newPasswordTouched}
		valid={isValid && newPasswordTouched}
		showStrength
		autocomplete="new-password"
		onInput={onNewPasswordInput}
		onBlur={onNewPasswordBlur}
	/>

	<PasswordField
		id="confirm-new-password"
		label="Повторите пароль"
		required
		value={confirmNewPassword}
		error={confirmNewPasswordError}
		touched={confirmNewPasswordTouched}
		valid={doPasswordsMatch && confirmNewPasswordTouched && confirmNewPassword.length > 0}
		autocomplete="new-password"
		onInput={onConfirmNewPasswordInput}
		onBlur={onConfirmNewPasswordBlur}
	/>

	<Button
		variant="solid"
		tone="primary"
		size="xl"
		radius="2xl"
		fullWidth
		disabled={!canSubmit}
		{loading}
		loadingText="Сохранение..."
		onclick={onSubmit}
	>
		Сохранить пароль
	</Button>
</div>
