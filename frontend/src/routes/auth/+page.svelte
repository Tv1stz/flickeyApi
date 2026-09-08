<!-- src/routes/auth/+page.svelte -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount, onDestroy, tick } from 'svelte';

	import AuthHeader from '$lib/components/auth/AuthHeader.svelte';
	import AuthFooter from '$lib/components/auth/AuthFooter.svelte';
	import PhoneStep from '$lib/components/auth/steps/PhoneStep.svelte';
	import CodeStep from '$lib/components/auth/steps/CodeStep.svelte';
	import PasswordStep from '$lib/components/auth/steps/PasswordStep.svelte';
	import RegisterStep from '$lib/components/auth/steps/RegisterStep.svelte';
	import ForgotPasswordStep from '$lib/components/auth/steps/ForgotPasswordStep.svelte';
	import NewPasswordStep from '$lib/components/auth/steps/NewPasswordStep.svelte';

	import { auth } from '$lib/auth/auth.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { authFlow, type AuthStep } from '$lib/stores/authFlowStore.svelte';
	import { isHostUser } from '$lib/auth/permissions';
	import { toast } from '$lib/stores/toastStore';
	import {
		sendVerificationCode,
		verifyCode,
		login,
		register,
		resendCode,
		resetPassword,
		getTestAccountsInfo
	} from '$lib/services/authService';
	import { getFullPhoneNumber, validatePhoneLength, type Country } from '$lib/config/countries';

	// ═══════════════════════════════════════════════════════════════
	// ANIMATION STATE
	// ═══════════════════════════════════════════════════════════════

	let mounted = $state(false);
	let isAnimating = $state(false);
	let direction = $state<'forward' | 'back'>('forward');
	let displayedStep = $state<AuthStep>('phone');
	let nextStep = $state<AuthStep | null>(null);

	// ═══════════════════════════════════════════════════════════════
	// COMPONENT REFS
	// ═══════════════════════════════════════════════════════════════

	let codeStepRef = $state<{ focus: () => void; clear: () => void } | undefined>();
	let resetCodeStepRef = $state<{ focus: () => void; clear: () => void } | undefined>();

	// ═══════════════════════════════════════════════════════════════
	// COMPUTED VALUES
	// ═══════════════════════════════════════════════════════════════

	const fullPhone = $derived(getFullPhoneNumber(authFlow.phone, authFlow.country));

	const stepContent = $derived.by(() => {
		const step = nextStep ?? displayedStep;
		const contents: Record<AuthStep, { title: string; subtitle: string }> = {
			phone: {
				title: 'Вход или регистрация',
				subtitle: 'Введите номер телефона для продолжения'
			},
			code: {
				title: 'Подтвердите номер',
				subtitle: `Введите код из SMS, отправленного на ${fullPhone}`
			},
			password: {
				title: 'Введите пароль',
				subtitle: 'Используйте пароль от вашего аккаунта'
			},
			register: {
				title: 'Создайте аккаунт',
				subtitle: 'Заполните данные для завершения регистрации'
			},
			'forgot-password': {
				title: 'Восстановление пароля',
				subtitle: 'Введите номер телефона, привязанный к аккаунту'
			},
			'reset-code': {
				title: 'Введите код',
				subtitle: `Мы отправили код подтверждения на ${fullPhone}`
			},
			'new-password': {
				title: 'Новый пароль',
				subtitle: 'Придумайте новый пароль для аккаунта'
			}
		};
		return contents[step];
	});

	const showBackButton = $derived(displayedStep !== 'phone');

	// ═══════════════════════════════════════════════════════════════
	// STEP TRANSITION
	// ═══════════════════════════════════════════════════════════════

	async function transitionTo(newStep: AuthStep, dir: 'forward' | 'back') {
		if (isAnimating || newStep === displayedStep) return;

		isAnimating = true;
		direction = dir;
		nextStep = newStep;

		await new Promise((resolve) => setTimeout(resolve, 180));

		displayedStep = newStep;
		authFlow.setStep(newStep);
		nextStep = null;

		await new Promise((resolve) => setTimeout(resolve, 200));

		isAnimating = false;

		if (newStep === 'code') {
			await tick();
			codeStepRef?.focus();
		} else if (newStep === 'reset-code') {
			await tick();
			resetCodeStepRef?.focus();
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS: PHONE STEP
	// ═══════════════════════════════════════════════════════════════

	function handlePhoneInput(phone: string, country: Country) {
		authFlow.setPhone(phone, country);
	}

	async function handleSendCode() {
		if (!validatePhoneLength(authFlow.phone, authFlow.country)) {
			authFlow.touchPhone();
			return;
		}

		authFlow.setLoading(true);

		try {
			const result = await sendVerificationCode(fullPhone);

			if (result.success) {
				authFlow.setIsNewUser(result.isNewUser);
				authFlow.startResendTimer();
				authFlow.resetCodeState();
				toast.success('Код отправлен', `SMS отправлено на ${fullPhone}`);
				await transitionTo('code', 'forward');
			} else {
				toast.error('Ошибка', result.error ?? 'Не удалось отправить код');
			}
		} catch {
			toast.error('Ошибка', 'Произошла ошибка. Попробуйте позже.');
		} finally {
			authFlow.setLoading(false);
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS: CODE STEP
	// ═══════════════════════════════════════════════════════════════

	function handleCodeChange(value: string) {
		authFlow.setCode(value);
	}

	async function handleVerifyCode(submittedCode?: string) {
		const codeToVerify = submittedCode ?? authFlow.code;
		if (codeToVerify.length !== 6) return;

		authFlow.setLoading(true);

		try {
			const result = await verifyCode(fullPhone, codeToVerify);

			if (result.success) {
				if (result.isNewUser) {
					toast.info('Почти готово', 'Заполните имя для завершения регистрации');
					await transitionTo('register', 'forward');
				} else {
					toast.success('Добро пожаловать!', 'Вход выполнен успешно');
					const pending = authStore.consumePendingAction();
					authFlow.resetAll();

					setTimeout(() => {
						if (pending.action === 'create-listing') {
							if (!isHostUser(authStore.user)) authStore.becomeHost();
							goto(resolve('/listings/new'));
						} else if (pending.redirect) {
							goto(resolve(pending.redirect as '/'));
						} else {
							goto(resolve('/'));
						}
					}, 300);
				}
			} else {
				toast.error('Неверный код', result.error ?? 'Проверьте код и попробуйте снова');
				codeStepRef?.clear();
			}
		} catch {
			toast.error('Ошибка', 'Произошла ошибка. Попробуйте позже.');
		} finally {
			authFlow.setLoading(false);
		}
	}

	async function handleQuickDevLogin(role: 'admin' | 'host' | 'guest') {
		authFlow.setLoading(true);
		try {
			const phone = role === 'admin' ? '+375290000001' : (role === 'host' ? '+375290000002' : '+375290000003');
			await auth.quickDevLogin(phone, role);
			toast.success(
				role === 'admin' ? 'Вход администратора' : (role === 'host' ? 'Вход хозяина' : 'Вход гостя'),
				'Добро пожаловать в Flickey'
			);
			const pending = authStore.consumePendingAction();
			authFlow.resetAll();
			setTimeout(() => {
				if (role === 'admin') {
					goto(resolve('/admin'));
				} else if (pending.action === 'create-listing') {
					if (!isHostUser(authStore.user)) authStore.becomeHost();
					goto(resolve('/listings/new'));
				} else if (pending.redirect) {
					goto(resolve(pending.redirect as '/'));
				} else {
					goto(resolve('/'));
				}
			}, 300);
		} catch (e: any) {
			toast.error('Ошибка входа', e?.message || 'Не удалось выполнить быстрый вход');
		} finally {
			authFlow.setLoading(false);
		}
	}

	async function handleResendCode() {
		if (authFlow.resendTimer > 0) return;

		authFlow.setLoading(true);

		try {
			const result = await resendCode(fullPhone);

			if (result.success) {
				toast.success('Код отправлен', 'Новый код отправлен на ваш номер');
				authFlow.startResendTimer();
				authFlow.resetCodeState();
				codeStepRef?.clear();
				resetCodeStepRef?.clear();
			} else {
				toast.error('Ошибка', result.error ?? 'Не удалось отправить код');
			}
		} catch {
			toast.error('Ошибка', 'Произошла ошибка');
		} finally {
			authFlow.setLoading(false);
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS: PASSWORD STEP
	// ═══════════════════════════════════════════════════════════════

	async function handleLogin() {
		if (authFlow.password.length < 6) {
			authFlow.touchPassword();
			return;
		}

		authFlow.setLoading(true);

		try {
			const result = await login(fullPhone, authFlow.password);

			if (result.success && result.user) {
				authStore.setUser(result.user);
				toast.success('Добро пожаловать!', `Рады видеть вас, ${result.user.name}`);

				const pending = authStore.consumePendingAction();
				authFlow.resetAll();

				setTimeout(() => {
					if (pending.action === 'create-listing') {
						if (!isHostUser(result.user)) {
							authStore.becomeHost();
						}
						goto(resolve('/listings/new'));
					} else if (pending.redirect) {
						goto(resolve(pending.redirect as '/'));
					} else {
						goto(resolve('/'));
					}
				}, 300);
			} else {
				toast.error('Ошибка входа', result.error ?? 'Неверный пароль');
			}
		} catch {
			toast.error('Ошибка', 'Произошла ошибка. Попробуйте позже.');
		} finally {
			authFlow.setLoading(false);
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS: REGISTER STEP
	// ═══════════════════════════════════════════════════════════════

	async function handleRegister() {
		const isNameValid = authFlow.name.trim().length >= 2;

		if (!isNameValid) {
			authFlow.touchName();
			return;
		}

		authFlow.setLoading(true);

		try {
			const result = await register(fullPhone, authFlow.name.trim(), authFlow.password);

			if (result.success && result.user) {
				authStore.setUser(result.user);
				toast.success('Аккаунт создан!', 'Добро пожаловать в Flickey');

				const pending = authStore.consumePendingAction();
				authFlow.resetAll();

				setTimeout(() => {
					if (pending.action === 'create-listing') {
						if (!isHostUser(result.user)) {
							authStore.becomeHost();
						}
						goto(resolve('/listings/new'));
					} else if (pending.redirect) {
						goto(resolve(pending.redirect as '/'));
					} else {
						goto(resolve('/'));
					}
				}, 300);
			} else {
				toast.error('Ошибка регистрации', result.error ?? 'Попробуйте ещё раз');
			}
		} catch {
			toast.error('Ошибка', 'Произошла ошибка. Попробуйте позже.');
		} finally {
			authFlow.setLoading(false);
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS: FORGOT PASSWORD FLOW
	// ═══════════════════════════════════════════════════════════════

	async function handleForgotPasswordSendCode() {
		if (!validatePhoneLength(authFlow.phone, authFlow.country)) {
			authFlow.touchPhone();
			return;
		}

		authFlow.setLoading(true);

		try {
			const result = await sendVerificationCode(fullPhone);

			if (result.success) {
				if (result.isNewUser) {
					toast.error('Аккаунт не найден', 'Пользователь с таким номером не зарегистрирован');
					authFlow.setLoading(false);
					return;
				}

				authFlow.startResendTimer();
				authFlow.resetCodeState();
				toast.success('Код отправлен', `SMS отправлено на ${fullPhone}`);
				await transitionTo('reset-code', 'forward');
			} else {
				toast.error('Ошибка', result.error ?? 'Не удалось отправить код');
			}
		} catch {
			toast.error('Ошибка', 'Произошла ошибка. Попробуйте позже.');
		} finally {
			authFlow.setLoading(false);
		}
	}

	async function handleResetCodeVerify(submittedCode?: string) {
		const codeToVerify = submittedCode ?? authFlow.code;
		if (codeToVerify.length !== 6) return;

		authFlow.setLoading(true);

		try {
			const result = await verifyCode(fullPhone, codeToVerify);

			if (result.success) {
				await transitionTo('new-password', 'forward');
			} else {
				toast.error('Неверный код', result.error ?? 'Проверьте код и попробуйте снова');
				resetCodeStepRef?.clear();
			}
		} catch {
			toast.error('Ошибка', 'Произошла ошибка. Попробуйте позже.');
		} finally {
			authFlow.setLoading(false);
		}
	}

	async function handleSetNewPassword() {
		const isPasswordValid = authFlow.newPassword.length >= 6;
		const doPasswordsMatch = authFlow.newPassword === authFlow.confirmNewPassword;

		if (!isPasswordValid || !doPasswordsMatch) {
			authFlow.touchNewPassword();
			authFlow.touchConfirmNewPassword();
			return;
		}

		authFlow.setLoading(true);

		try {
			const result = await resetPassword(fullPhone, authFlow.newPassword);

			if (!result.success) {
				toast.error('Ошибка', result.error ?? 'Не удалось изменить пароль');
				return;
			}

			toast.success('Пароль изменён', 'Теперь вы можете войти с новым паролем');

			authFlow.setPassword(authFlow.newPassword);
			authFlow.resetNewPasswordState();
			await transitionTo('password', 'forward');
		} catch {
			toast.error('Ошибка', 'Не удалось изменить пароль');
		} finally {
			authFlow.setLoading(false);
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// NAVIGATION
	// ═══════════════════════════════════════════════════════════════

	async function handleBack() {
		if (isAnimating) return;

		switch (displayedStep) {
			case 'code':
				authFlow.resetCodeState();
				await transitionTo('phone', 'back');
				break;
			case 'password':
			case 'register':
				authFlow.resetPasswordState();
				await transitionTo('code', 'back');
				break;
			case 'forgot-password':
				await transitionTo('password', 'back');
				break;
			case 'reset-code':
				authFlow.resetCodeState();
				await transitionTo('forgot-password', 'back');
				break;
			case 'new-password':
				authFlow.resetNewPasswordState();
				await transitionTo('reset-code', 'back');
				break;
		}
	}

	function handleClose() {
		authStore.clearPendingAction();
		authFlow.resetAll();
		goto(resolve('/'));
	}

	async function handleForgotPassword() {
		authFlow.touchPhone();
		await transitionTo('forgot-password', 'forward');
	}

	// ═══════════════════════════════════════════════════════════════
	// KEYBOARD HANDLER
	// ═══════════════════════════════════════════════════════════════

	function handleKeydown(e: KeyboardEvent) {
		if (e.key !== 'Enter' || isAnimating) return;

		e.preventDefault();

		switch (displayedStep) {
			case 'phone':
				handleSendCode();
				break;
			case 'code':
				handleVerifyCode();
				break;
			case 'password':
				handleLogin();
				break;
			case 'register':
				handleRegister();
				break;
			case 'forgot-password':
				handleForgotPasswordSendCode();
				break;
			case 'reset-code':
				handleResetCodeVerify();
				break;
			case 'new-password':
				handleSetNewPassword();
				break;
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// LIFECYCLE
	// ═══════════════════════════════════════════════════════════════

	onMount(() => {
		displayedStep = authFlow.step;

		requestAnimationFrame(() => {
			mounted = true;
		});

		authStore.onLogout(() => {
			authFlow.resetAll();
			displayedStep = 'phone';
		});

		if (import.meta.env.DEV) {
			console.log(getTestAccountsInfo());
		}
	});

	onDestroy(() => {
		authFlow.cleanup();
	});
</script>

<svelte:head>
	<title>{stepContent.title} | Flickey</title>
</svelte:head>

<svelte:window onkeydown={handleKeydown} />
<div class="auth-page" class:auth-page--mounted={mounted}>
	<!-- Header -->
	<AuthHeader
		progress={authFlow.progressPercent}
		{showBackButton}
		onBack={handleBack}
		onClose={handleClose}
	/>

	<!-- Main content -->
	<main class="auth-main">
		<div class="auth-container">
			<!-- Dev Quick Login Panel -->
			<div class="mb-5 rounded-2xl border border-emerald-500/20 bg-emerald-500/5 p-3 sm:p-4">
				<p class="mb-2.5 text-xs font-semibold text-emerald-800">Быстрый вход для разработки:</p>
				<div class="flex flex-wrap items-center gap-2">
					<button
						type="button"
						class="flex items-center gap-1.5 rounded-xl bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-emerald-700 transition-colors"
						onclick={() => handleQuickDevLogin('admin')}
						disabled={authFlow.loading}
					>
						🛡️ Администратор
					</button>
					<button
						type="button"
						class="flex items-center gap-1.5 rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-xs font-semibold text-zinc-700 hover:bg-zinc-50 transition-colors"
						onclick={() => handleQuickDevLogin('host')}
						disabled={authFlow.loading}
					>
						🏠 Хозяин
					</button>
					<button
						type="button"
						class="flex items-center gap-1.5 rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-xs font-semibold text-zinc-700 hover:bg-zinc-50 transition-colors"
						onclick={() => handleQuickDevLogin('guest')}
						disabled={authFlow.loading}
					>
						👤 Гость
					</button>
				</div>
			</div>

			<!-- Step header -->
			<div class="auth-header">
				<div
					class="auth-header__content"
					class:auth-header__content--exiting={isAnimating && nextStep}
				>
					<h1 class="auth-title">
						{stepContent.title}
					</h1>
					<p class="auth-subtitle">
						{stepContent.subtitle}
					</p>
				</div>
			</div>

			<!-- Step content -->
			<div class="auth-content">
				<div
					class="auth-step"
					class:auth-step--forward={direction === 'forward'}
					class:auth-step--back={direction === 'back'}
					class:auth-step--exiting={isAnimating && nextStep}
				>
					{#if displayedStep === 'phone'}
						<PhoneStep
							phone={authFlow.phone}
							country={authFlow.country}
							touched={authFlow.phoneTouched}
							loading={authFlow.loading}
							onInput={handlePhoneInput}
							onBlur={() => authFlow.touchPhone()}
							onSubmit={handleSendCode}
						/>
					{:else if displayedStep === 'code'}
						<CodeStep
							bind:this={codeStepRef}
							code={authFlow.code}
							loading={authFlow.loading}
							resendTimer={authFlow.resendTimer}
							onCodeChange={handleCodeChange}
							onCodeComplete={handleVerifyCode}
							onResend={handleResendCode}
							onSubmit={() => handleVerifyCode()}
						/>
					{:else if displayedStep === 'password'}
						<PasswordStep
							password={authFlow.password}
							touched={authFlow.passwordTouched}
							loading={authFlow.loading}
							onInput={(v) => authFlow.setPassword(v)}
							onBlur={() => authFlow.touchPassword()}
							onSubmit={handleLogin}
							onForgotPassword={handleForgotPassword}
						/>
					{:else if displayedStep === 'register'}
						<RegisterStep
							name={authFlow.name}
							password={authFlow.password}
							confirmPassword={authFlow.confirmPassword}
							nameTouched={authFlow.nameTouched}
							passwordTouched={authFlow.passwordTouched}
							confirmPasswordTouched={authFlow.confirmPasswordTouched}
							loading={authFlow.loading}
							onNameInput={(v) => authFlow.setName(v)}
							onNameBlur={() => authFlow.touchName()}
							onPasswordInput={(v) => authFlow.setPassword(v)}
							onPasswordBlur={() => authFlow.touchPassword()}
							onConfirmPasswordInput={(v) => authFlow.setConfirmPassword(v)}
							onConfirmPasswordBlur={() => authFlow.touchConfirmPassword()}
							onSubmit={handleRegister}
						/>
					{:else if displayedStep === 'forgot-password'}
						<ForgotPasswordStep
							phone={authFlow.phone}
							country={authFlow.country}
							touched={authFlow.phoneTouched}
							loading={authFlow.loading}
							onInput={handlePhoneInput}
							onBlur={() => authFlow.touchPhone()}
							onSubmit={handleForgotPasswordSendCode}
						/>
					{:else if displayedStep === 'reset-code'}
						<CodeStep
							bind:this={resetCodeStepRef}
							code={authFlow.code}
							loading={authFlow.loading}
							resendTimer={authFlow.resendTimer}
							onCodeChange={handleCodeChange}
							onCodeComplete={handleResetCodeVerify}
							onResend={handleResendCode}
							onSubmit={() => handleResetCodeVerify()}
						/>
					{:else if displayedStep === 'new-password'}
						<NewPasswordStep
							newPassword={authFlow.newPassword}
							confirmNewPassword={authFlow.confirmNewPassword}
							newPasswordTouched={authFlow.newPasswordTouched}
							confirmNewPasswordTouched={authFlow.confirmNewPasswordTouched}
							loading={authFlow.loading}
							onNewPasswordInput={(v) => authFlow.setNewPassword(v)}
							onNewPasswordBlur={() => authFlow.touchNewPassword()}
							onConfirmNewPasswordInput={(v) => authFlow.setConfirmNewPassword(v)}
							onConfirmNewPasswordBlur={() => authFlow.touchConfirmNewPassword()}
							onSubmit={handleSetNewPassword}
						/>
					{/if}
				</div>
			</div>
		</div>
	</main>

	<!-- Footer -->
	<AuthFooter />
</div>

<style>
	/* ═══════════════════════════════════════════════════════════════
       PAGE LAYOUT
       ═══════════════════════════════════════════════════════════════ */

	.auth-page {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		background-color: white;
		opacity: 0;
		transition: opacity 0.25s ease-out;
	}

	.auth-page--mounted {
		opacity: 1;
	}

	.auth-main {
		flex: 1;
		display: flex;
		flex-direction: column;
		padding: 1.5rem 1rem;
	}

	.auth-container {
		width: 100%;
		max-width: 28rem; /* max-w-lg */
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		flex: 1;
	}

	/* ═══════════════════════════════════════════════════════════════
       HEADER SECTION
       ═══════════════════════════════════════════════════════════════ */

	.auth-header {
		margin-bottom: 2rem;
		min-height: 4.5rem;
	}

	.auth-header__content {
		transition:
			opacity 0.15s ease-out,
			transform 0.18s ease-out;
	}

	.auth-header__content--exiting {
		opacity: 0;
		transform: translateY(-6px);
	}

	.auth-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: #18181b; /* zinc-900 */
		letter-spacing: -0.025em;
		margin-bottom: 0.5rem;
		line-height: 1.2;
	}

	.auth-subtitle {
		color: #71717a; /* zinc-500 */
		font-size: 0.875rem;
		line-height: 1.5;
	}

	/* ═══════════════════════════════════════════════════════════════
       CONTENT SECTION - NO OVERFLOW HIDDEN!
       ═══════════════════════════════════════════════════════════════ */

	.auth-content {
		flex: 1;
		/* Убрали overflow-hidden чтобы ring эффекты не обрезались */
		/* Добавили небольшой padding для ring-offset */
		padding: 2px;
		margin: -2px;
	}

	.auth-step {
		transition:
			opacity 0.18s ease-out,
			transform 0.22s cubic-bezier(0.16, 1, 0.3, 1);
	}

	/* Forward animation */
	.auth-step--forward.auth-step--exiting {
		opacity: 0;
		transform: translateX(-20px);
	}

	/* Back animation */
	.auth-step--back.auth-step--exiting {
		opacity: 0;
		transform: translateX(20px);
	}
</style>
