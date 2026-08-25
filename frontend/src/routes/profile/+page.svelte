<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Alert from '$lib/components/ui/Alert.svelte';
	import { toast } from '$lib/components/ui/toast.svelte';
	import { formatDate, maskPhone } from '$lib/utils';
	import {
		User,
		Phone,
		Mail,
		ShieldCheck,
		Calendar,
		LogOut,
		Save,
		Building
	} from 'lucide-svelte';

	let firstName = $state(auth.user?.first_name || '');
	let lastName = $state(auth.user?.last_name || '');
	let email = $state(auth.user?.email || '');

	let saving = $state(false);
	let errorMessage = $state<string | null>(null);

	$effect(() => {
		if (auth.user) {
			firstName = auth.user.first_name || '';
			lastName = auth.user.last_name || '';
			email = auth.user.email || '';
		}
	});

	async function handleSaveProfile() {
		if (!firstName.trim() || !lastName.trim()) {
			errorMessage = 'Имя и фамилия обязательны для заполнения';
			return;
		}

		errorMessage = null;
		saving = true;
		try {
			await auth.completeProfile(firstName.trim(), lastName.trim(), email.trim() || undefined);
			toast.success('Данные профиля успешно обновлены');
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось сохранить профиль';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Мой профиль — Flickey</title>
</svelte:head>

<div class="container mx-auto px-4 sm:px-6 py-8 max-w-4xl space-y-8">
	<div>
		<h1 class="text-2xl sm:text-3xl font-extrabold text-foreground tracking-tight">
			Профиль пользователя
		</h1>
		<p class="text-xs sm:text-sm text-muted-foreground mt-1">
			Управление личными данными, статусом аккаунта и безопасностью сессий.
		</p>
	</div>

	{#if !auth.isAuthenticated}
		<div class="rounded-3xl border border-border bg-card p-8 sm:p-12 text-center max-w-md mx-auto space-y-4 shadow-sm">
			<h3 class="text-lg font-bold text-foreground">Требуется авторизация</h3>
			<p class="text-xs text-muted-foreground leading-relaxed">
				Войдите в систему для просмотра и редактирования профиля.
			</p>
			<Button onclick={() => auth.openAuthModal('phone')} class="font-semibold">
				Войти в аккаунт
			</Button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
			<!-- Left Col: User Identity Card -->
			<div class="md:col-span-1 space-y-6">
				<div class="rounded-3xl border border-border bg-card p-6 shadow-sm text-center space-y-4">
					<div class="mx-auto flex h-20 w-20 items-center justify-center rounded-3xl bg-primary/10 text-primary font-extrabold text-2xl shadow-inner">
						{auth.user?.first_name ? auth.user.first_name[0].toUpperCase() : 'U'}
					</div>

					<div>
						<h3 class="font-bold text-lg text-foreground">{auth.displayName}</h3>
						<p class="text-xs text-muted-foreground font-mono mt-0.5">{auth.user?.phone}</p>
					</div>

					<div class="pt-2 border-t border-border flex flex-col gap-2">
						<div class="flex items-center justify-between text-xs">
							<span class="text-muted-foreground">Роль в системе:</span>
							<Badge variant="default" class="uppercase font-bold text-[10px]">
								{auth.user?.role}
							</Badge>
						</div>

						<div class="flex items-center justify-between text-xs">
							<span class="text-muted-foreground">Статус профиля:</span>
							<Badge
								variant={auth.user?.status === 'active' ? 'success' : 'warning'}
								class="text-[10px]"
							>
								{auth.user?.status}
							</Badge>
						</div>

						<div class="flex items-center justify-between text-xs">
							<span class="text-muted-foreground">Регистрация:</span>
							<span class="font-medium text-foreground">{auth.user?.created_at ? formatDate(auth.user.created_at) : '—'}</span>
						</div>
					</div>
				</div>

				<!-- Session Actions -->
				<div class="rounded-2xl border border-border bg-card p-4 space-y-2">
					<button
						type="button"
						onclick={() => auth.logout()}
						class="flex w-full items-center justify-center gap-2 rounded-xl p-2.5 text-xs font-semibold text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
					>
						<LogOut class="h-4 w-4" /> Выйти из аккаунта
					</button>
				</div>
			</div>

			<!-- Right Col: Profile Form -->
			<div class="md:col-span-2 space-y-6">
				<div class="rounded-3xl border border-border bg-card p-6 sm:p-8 shadow-sm space-y-6">
					<div>
						<h3 class="text-lg font-bold text-foreground">Личные данные</h3>
						<p class="text-xs text-muted-foreground mt-1">
							Используются для оформления бронирований и взаимодействия с гостями/хостами.
						</p>
					</div>

					{#if errorMessage}
						<Alert variant="destructive">{errorMessage}</Alert>
					{/if}

					<form
						onsubmit={(e) => {
							e.preventDefault();
							handleSaveProfile();
						}}
						class="space-y-4"
					>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
							<div class="space-y-1.5">
								<label for="p-first-name" class="text-xs font-semibold text-foreground">
									Имя <span class="text-destructive">*</span>
								</label>
								<Input id="p-first-name" bind:value={firstName} required />
							</div>

							<div class="space-y-1.5">
								<label for="p-last-name" class="text-xs font-semibold text-foreground">
									Фамилия <span class="text-destructive">*</span>
								</label>
								<Input id="p-last-name" bind:value={lastName} required />
							</div>
						</div>

						<div class="space-y-1.5">
							<label for="p-phone-dis" class="text-xs font-semibold text-foreground">
								Номер телефона
							</label>
							<Input id="p-phone-dis" value={auth.user?.phone || ''} disabled class="bg-muted/50 font-mono" />
							<p class="text-[11px] text-muted-foreground">Номер телефона привязан к вашей учетной записи</p>
						</div>

						<div class="space-y-1.5">
							<label for="p-email" class="text-xs font-semibold text-foreground">
								Электронная почта
							</label>
							<Input id="p-email" type="email" placeholder="example@mail.ru" bind:value={email} />
						</div>

						<div class="pt-4 border-t border-border flex justify-end">
							<Button type="submit" loading={saving} class="gap-2 font-semibold shadow-md">
								<Save class="h-4 w-4" /> Сохранить изменения
							</Button>
						</div>
					</form>
				</div>
			</div>
		</div>
	{/if}
</div>
