<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Alert from '$lib/components/ui/Alert.svelte';
	import { toast } from '$lib/components/ui/toast.svelte';
	import { Phone, KeyRound, UserCheck, Zap, ArrowLeft } from 'lucide-svelte';

	let phone = $state('+375291234567');
	let otpCode = $state('');
	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');

	let loading = $state(false);
	let errorMessage = $state<string | null>(null);
	let cooldownSeconds = $state(0);
	let cooldownInterval: ReturnType<typeof setInterval> | null = null;

	function startCooldown(seconds: number = 60) {
		cooldownSeconds = seconds;
		if (cooldownInterval) clearInterval(cooldownInterval);
		cooldownInterval = setInterval(() => {
			if (cooldownSeconds > 0) {
				cooldownSeconds--;
			} else if (cooldownInterval) {
				clearInterval(cooldownInterval);
				cooldownInterval = null;
			}
		}, 1000);
	}

	async function handleSendOTP() {
		if (!phone || phone.length < 9) {
			errorMessage = 'Введите корректный номер телефона в формате E.164 (напр. +375291234567)';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await auth.requestOTP(phone);
			startCooldown(60);
			toast.info('Код подтверждения отправлен (в режиме dev проверьте логи сервера)');
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось отправить SMS-код';
		} finally {
			loading = false;
		}
	}

	async function handleVerifyOTP() {
		if (!otpCode || otpCode.length < 4) {
			errorMessage = 'Введите код подтверждения';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await auth.verifyOTP(otpCode);
			toast.success('Вы успешно вошли в систему');
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Неверный или просроченный код';
		} finally {
			loading = false;
		}
	}

	async function handleCompleteProfile() {
		if (!firstName.trim() || !lastName.trim()) {
			errorMessage = 'Имя и фамилия обязательны для заполнения';
			return;
		}

		errorMessage = null;
		loading = true;
		try {
			await auth.completeProfile(firstName.trim(), lastName.trim(), email.trim() || undefined);
			toast.success('Профиль успешно заполнен');
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось сохранить профиль';
		} finally {
			loading = false;
		}
	}

	async function handleQuickLogin(role: 'guest' | 'host' | 'admin') {
		errorMessage = null;
		loading = true;
		try {
			const devPhone = role === 'admin' ? '+375290000000' : role === 'host' ? '+375292222222' : '+375291111111';
			await auth.quickDevLogin(devPhone, role);
			toast.success(`Вход выполнен в режиме ${role === 'admin' ? 'Администратора' : role === 'host' ? 'Хоста' : 'Гостя'}`);
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Ошибка быстрого входа';
		} finally {
			loading = false;
		}
	}
</script>

<Dialog
	bind:open={auth.isAuthModalOpen}
	title={auth.authModalView === 'phone'
		? 'Вход или регистрация'
		: auth.authModalView === 'otp'
			? 'Подтверждение номера'
			: 'Заполнение профиля'}
	description={auth.authModalView === 'phone'
		? 'Введите номер телефона для получения одноразового SMS-кода'
		: auth.authModalView === 'otp'
			? `Код отправлен на номер ${auth.pendingPhone}`
			: 'Укажите ваши имя и фамилию для завершения регистрации'}
	onclose={() => (errorMessage = null)}
>
	<!-- Dev Admin Login Banner -->
	<div class="mb-5 rounded-xl border border-dashed border-emerald-500/30 bg-emerald-500/5 p-3 flex items-center justify-between gap-3">
		<div class="flex items-center gap-2 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
			<Zap class="h-4 w-4 shrink-0" />
			<span>Вход в панель администратора:</span>
		</div>
		<Button
			variant="default"
			size="sm"
			class="text-[11px] h-7 px-3 bg-emerald-600 hover:bg-emerald-700 text-white font-semibold shrink-0"
			onclick={() => handleQuickLogin('admin')}
			disabled={loading}
		>
			🛡️ Войти как Администратор
		</Button>
	</div>

	{#if errorMessage}
		<div class="mb-4">
			<Alert variant="destructive">{errorMessage}</Alert>
		</div>
	{/if}

	<!-- View 1: Phone Request -->
	{#if auth.authModalView === 'phone'}
		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleSendOTP();
			}}
			class="space-y-4"
		>
			<div class="space-y-2">
				<label for="auth-phone" class="text-xs font-medium text-foreground">
					Номер телефона (в международном формате)
				</label>
				<div class="relative">
					<Phone class="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
					<Input
						id="auth-phone"
						type="tel"
						placeholder="+375291234567"
						bind:value={phone}
						class="pl-9 font-mono"
						required
					/>
				</div>
			</div>

			<Button type="submit" class="w-full" {loading}>
				Получить код по SMS
			</Button>
		</form>
	{/if}

	<!-- View 2: OTP Verification -->
	{#if auth.authModalView === 'otp'}
		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleVerifyOTP();
			}}
			class="space-y-4"
		>
			<div class="space-y-2">
				<label for="auth-otp" class="text-xs font-medium text-foreground">
					6-значный SMS-код
				</label>
				<div class="relative">
					<KeyRound class="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
					<Input
						id="auth-otp"
						type="text"
						inputmode="numeric"
						placeholder="123456"
						maxlength={6}
						bind:value={otpCode}
						class="pl-9 tracking-[0.3em] font-mono text-base font-bold"
						required
					/>
				</div>
			</div>

			<div class="flex items-center justify-between text-xs text-muted-foreground">
				<button
					type="button"
					onclick={() => (auth.authModalView = 'phone')}
					class="flex items-center gap-1 hover:text-foreground cursor-pointer"
				>
					<ArrowLeft class="h-3.5 w-3.5" /> Изменить номер
				</button>

				{#if cooldownSeconds > 0}
					<span>Повтор через {cooldownSeconds}с</span>
				{:else}
					<button
						type="button"
						onclick={handleSendOTP}
						class="text-primary hover:underline cursor-pointer font-medium"
					>
						Отправить повторно
					</button>
				{/if}
			</div>

			<Button type="submit" class="w-full" {loading}>
				Войти в аккаунт
			</Button>
		</form>
	{/if}

	<!-- View 3: Complete Profile -->
	{#if auth.authModalView === 'complete_profile'}
		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleCompleteProfile();
			}}
			class="space-y-4"
		>
			<div class="space-y-2">
				<label for="prof-first-name" class="text-xs font-medium text-foreground">
					Имя <span class="text-destructive">*</span>
				</label>
				<Input
					id="prof-first-name"
					type="text"
					placeholder="Иван"
					bind:value={firstName}
					required
				/>
			</div>

			<div class="space-y-2">
				<label for="prof-last-name" class="text-xs font-medium text-foreground">
					Фамилия <span class="text-destructive">*</span>
				</label>
				<Input
					id="prof-last-name"
					type="text"
					placeholder="Иванов"
					bind:value={lastName}
					required
				/>
			</div>

			<div class="space-y-2">
				<label for="prof-email" class="text-xs font-medium text-foreground">
					Email (необязательно)
				</label>
				<Input
					id="prof-email"
					type="email"
					placeholder="ivan@example.by"
					bind:value={email}
				/>
			</div>

			<Button type="submit" class="w-full" {loading}>
				Сохранить и продолжить
			</Button>
		</form>
	{/if}
</Dialog>
