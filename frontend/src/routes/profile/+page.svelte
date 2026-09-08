<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { notifications } from '$lib/stores/notifications.svelte';
	import { toast } from '$lib/stores/toastStore';
	import AccountSidebarNav from '$lib/components/account-settings/AccountSidebarNav.svelte';
	import AccountSectionMenu from '$lib/components/account-settings/AccountSectionMenu.svelte';
	import AccountSettingsRows from '$lib/components/account-settings/AccountSettingsRows.svelte';
	import AccountProfileCard from '$lib/components/account-settings/AccountProfileCard.svelte';
	import NotificationItem from '$lib/components/notifications/NotificationItem.svelte';
	import {
		getNavItems,
		buildRowsBySection,
		formatMemberSince,
		getUserInitials,
		isAccountSectionId
	} from '$lib/components/account-settings/accountSettingsData';
	import { LogOut, ArrowLeft, LayoutGrid, CheckCheck, Inbox, Sparkles } from 'lucide-svelte';
	import type {
		AccountSectionId,
		AccountSettingAction,
		AccountSettingsRoute
	} from '$lib/components/account-settings/types';

	let authInitialized = $derived(authStore.initialized);
	let authenticated = $derived(authStore.isAuthenticated);
	let user = $derived(authStore.user);
	let verified = $derived(authStore.isVerifiedHost);
	let hostMode = $derived(authStore.isHostMode);
	let isAdmin = $derived(authStore.user?.roles?.includes('admin') || Boolean(user?.roles?.includes('admin')));

	let redirectingToAuth = $state(false);
	let activeSection = $state<AccountSectionId>('personal');
	let mobileSection = $state<AccountSectionId | null>(null);
	let avatarInputRef = $state<HTMLInputElement | null>(null);
	let desktopContentRef = $state<HTMLElement | null>(null);
	let avatarUploading = $state(false);

	const navItems = $derived(getNavItems(isAdmin));
	let activeItem = $derived(navItems.find((item) => item.id === activeSection) ?? navItems[0]);
	let mobileActiveItem = $derived(
		mobileSection ? (navItems.find((item) => item.id === mobileSection) ?? navItems[0]) : null
	);
	let memberSince = $derived(formatMemberSince(user?.createdAt));
	let userInitials = $derived(getUserInitials(user?.name));
	let currentModeLabel = $derived(hostMode ? 'Хозяин' : 'Гость');
	let rowsBySection = $derived(buildRowsBySection(user, verified, memberSince));
	let activeRows = $derived(rowsBySection[activeSection]);
	let mobileRows = $derived(mobileSection ? rowsBySection[mobileSection] : []);

	let notifTab = $state<'all' | 'unread'>('all');
	let isSendingTest = $state(false);

	let notifItems = $derived(notifications.items);
	let notifUnreadCount = $derived(notifications.unreadCount);
	let notifHasUnread = $derived(notifications.hasUnread);
	let notifIsLoading = $derived(notifications.isLoading);

	let notifDisplayedItems = $derived.by(() => {
		if (notifTab === 'unread') {
			return notifItems.filter((i) => !i.is_read);
		}
		return notifItems;
	});

	$effect(() => {
		const sec = $page.url.searchParams.get('section');
		if (sec && isAccountSectionId(sec)) {
			activeSection = sec;
			mobileSection = sec;
		}
	});

	$effect(() => {
		if (!authStore.initialized || authStore.user || redirectingToAuth) return;
		redirectingToAuth = true;
		authStore.setPendingAction(null, '/profile');
		goto(resolve('/auth'));
	});

	async function handleSendTestNotif() {
		isSendingTest = true;
		try {
			await notifications.sendDevTestNotification();
		} finally {
			isSendingTest = false;
		}
	}

	function goTo(path: AccountSettingsRoute) {
		goto(resolve(path));
	}

	function handleSoon(feature: string) {
		toast.info('Скоро', `${feature} будут доступны в следующем обновлении`);
	}

	function handleSectionAction(action: Exclude<AccountSettingAction, { type: 'none' }>) {
		if (action.type === 'route') {
			goTo(action.path);
			return;
		}
		handleSoon(action.feature);
	}

	function selectSection(id: AccountSectionId) {
		if (id === 'help') {
			goTo('/help');
			return;
		}
		if (id === 'admin') {
			goTo('/admin');
			return;
		}

		activeSection = id;
		mobileSection = id;

		requestAnimationFrame(() => {
			desktopContentRef?.scrollTo({ top: 0, behavior: 'auto' });
		});
	}

	async function requestLogout() {
		const confirmed = await toast.confirm({
			title: 'Выход из аккаунта',
			message: 'Вы точно хотите завершить текущую сессию?',
			cancelText: 'Остаться',
			confirmText: 'Выйти',
			type: 'warning'
		});

		if (!confirmed) return;
		authStore.logout();
		goto(resolve('/'));
	}

	function triggerAvatarUpload() {
		avatarInputRef?.click();
	}

	function fileToDataUrl(file: File): Promise<string> {
		return new Promise((resolveRead, rejectRead) => {
			const reader = new FileReader();
			reader.onload = () => {
				if (typeof reader.result === 'string') {
					resolveRead(reader.result);
					return;
				}
				rejectRead(new Error('Не удалось прочитать файл'));
			};
			reader.onerror = () => rejectRead(new Error('Ошибка чтения файла'));
			reader.readAsDataURL(file);
		});
	}

	async function handleAvatarChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement | null;
		const file = input?.files?.[0];
		if (!file) return;

		try {
			if (!file.type.startsWith('image/')) {
				toast.error('Неверный формат', 'Выберите изображение');
				return;
			}
			if (file.size > 8 * 1024 * 1024) {
				toast.error('Слишком большой файл', 'Максимальный размер изображения 8 МБ');
				return;
			}

			avatarUploading = true;
			const avatar = await fileToDataUrl(file);
			authStore.updateProfile({ avatar });
			toast.success('Фото обновлено');
		} catch (error) {
			console.error('Avatar update error:', error);
			toast.error('Не удалось обновить фото');
		} finally {
			avatarUploading = false;
			if (input) input.value = '';
		}
	}
</script>

<svelte:head>
	<title>Настройки аккаунта — Flickey</title>
</svelte:head>

{#if !authInitialized}
	<div class="flex min-h-[50vh] items-center justify-center bg-white">
		<div class="flex flex-col items-center gap-3">
			<div class="h-8 w-8 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"></div>
			<p class="text-sm text-zinc-500">Загрузка аккаунта...</p>
		</div>
	</div>
{:else if !authenticated}
	<div class="flex min-h-[50vh] items-center justify-center bg-white">
		<p class="text-sm text-zinc-500">Перенаправляем на страницу входа...</p>
	</div>
{:else}
	<!-- Desktop View -->
	<div class="hidden bg-zinc-50/50 lg:flex lg:h-[calc(100dvh-80px)] lg:overflow-hidden">
		<aside
			class="h-full w-72 shrink-0 overflow-y-auto border-r border-zinc-200 bg-white px-5 py-8 xl:w-80 xl:px-6 xl:py-10"
		>
			<div class="px-1">
				<AccountProfileCard
					name={user?.name || 'Профиль'}
					initials={userInitials}
					modeLabel={currentModeLabel}
					avatar={user?.avatar}
					onAvatarClick={triggerAvatarUpload}
				/>
			</div>

			{#if hostMode}
				<div class="mt-6 px-1">
					<a
						href="/host/listings"
						class="flex items-center justify-between rounded-2xl border border-zinc-200 bg-zinc-50 p-3.5 text-sm font-semibold text-zinc-900 transition-colors hover:bg-zinc-100"
					>
						<span class="flex items-center gap-2.5">
							<LayoutGrid class="h-4 w-4 text-zinc-700" strokeWidth={2} />
							<span>Мои объявления</span>
						</span>
						<span class="rounded-full bg-zinc-200 px-2 py-0.5 text-xs text-zinc-700">Перейти</span>
					</a>
				</div>
			{/if}

			<div class="mt-8 px-1">
				<AccountSidebarNav
					items={navItems}
					{activeSection}
					onSelect={selectSection}
				/>
			</div>

			<div class="mt-auto border-t border-zinc-200 pt-6 px-1">
				<button
					type="button"
					class="group flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-left text-sm font-medium text-zinc-500 transition-colors hover:text-zinc-900"
					onclick={requestLogout}
				>
					<LogOut class="h-4 w-4" strokeWidth={1.8} />
					<span>Выйти</span>
				</button>
			</div>
		</aside>

		<section bind:this={desktopContentRef} class="min-w-0 flex-1 overflow-y-auto overscroll-contain bg-white">
			<div class="mx-auto max-w-3xl px-8 pt-10 pb-16 xl:px-12 xl:pt-12">
				<h2 class="text-2xl font-bold tracking-tight text-zinc-900 sm:text-3xl">
					{activeItem.title}
				</h2>
				<p class="mt-1 text-sm text-zinc-500">
					Управление настройками {activeItem.title.toLowerCase()}
				</p>
				{#if activeSection === 'notifications'}
					{@render notificationCenter(false)}
				{:else}
					<div class="mt-8">
						<AccountSettingsRows rows={activeRows} onAction={handleSectionAction} variant="desktop" />
					</div>
				{/if}
			</div>
		</section>
	</div>

	<!-- Mobile View -->
	<div class="min-h-dvh bg-white lg:hidden">
		<div class="mx-auto max-w-lg px-4 pt-6 pb-28 sm:px-6">
			{#if !mobileSection}
				<div class="flex items-center justify-between py-2">
					<h1 class="text-2xl font-bold text-zinc-900">Профиль</h1>
					<button
						type="button"
						class="flex items-center gap-1.5 text-sm font-medium text-zinc-500 transition-colors hover:text-zinc-900"
						onclick={requestLogout}
					>
						<span>Выйти</span>
						<LogOut class="h-4 w-4" strokeWidth={1.8} />
					</button>
				</div>

				<div class="mt-4 py-4">
					<AccountProfileCard
						name={user?.name || 'Профиль'}
						initials={userInitials}
						modeLabel={currentModeLabel}
						avatar={user?.avatar}
						onAvatarClick={triggerAvatarUpload}
						variant="mobile"
					/>
				</div>

				{#if hostMode}
					<div class="my-4">
						<a
							href="/host/listings"
							class="flex items-center justify-between rounded-2xl border border-zinc-200 bg-zinc-50 p-4 text-sm font-semibold text-zinc-900 transition-colors hover:bg-zinc-100 active:scale-[0.98]"
						>
							<span class="flex items-center gap-2.5">
								<LayoutGrid class="h-4 w-4 text-zinc-700" strokeWidth={2} />
								<span>Мои объявления</span>
							</span>
							<span class="rounded-full bg-zinc-200 px-2.5 py-0.5 text-xs text-zinc-700">Перейти</span>
						</a>
					</div>
				{/if}

				<div class="mt-4">
					<AccountSectionMenu items={navItems} onSelect={selectSection} />
				</div>
			{:else}
				<div class="flex items-center gap-3 py-2">
					<button
						type="button"
						class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-zinc-100 text-zinc-900 transition-colors hover:bg-zinc-200"
						onclick={() => (mobileSection = null)}
						aria-label="Назад к списку разделов"
					>
						<ArrowLeft class="h-5 w-5" strokeWidth={2} />
					</button>
					<h1 class="text-xl font-bold truncate text-zinc-900">
						{mobileActiveItem?.title}
					</h1>
				</div>

				{#if mobileSection === 'notifications'}
					{@render notificationCenter(true)}
				{:else}
					<div class="mt-4">
						<AccountSettingsRows rows={mobileRows} onAction={handleSectionAction} variant="mobile" />
					</div>
				{/if}
			{/if}
		</div>
	</div>

	{#snippet notificationCenter(isMobile: boolean)}
		<div class="{isMobile ? 'mt-4' : 'mt-8'} space-y-6">
			<div class="flex flex-wrap items-center justify-between gap-3 rounded-2xl bg-zinc-50 p-3.5 sm:p-4 border border-zinc-200/80">
				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={() => (notifTab = 'all')}
						class="rounded-xl px-3 py-1.5 text-xs font-semibold transition-all
						       {notifTab === 'all' ? 'bg-white text-zinc-900 shadow-xs' : 'text-zinc-500 hover:text-zinc-800'}"
					>
						Все ({notifItems.length})
					</button>
					<button
						type="button"
						onclick={() => (notifTab = 'unread')}
						class="rounded-xl px-3 py-1.5 text-xs font-semibold transition-all
						       {notifTab === 'unread' ? 'bg-white text-zinc-900 shadow-xs' : 'text-zinc-500 hover:text-zinc-800'}"
					>
						Непрочитанные ({notifUnreadCount})
					</button>
				</div>

				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={handleSendTestNotif}
						disabled={isSendingTest}
						class="flex items-center gap-1.5 rounded-xl border border-amber-200 bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-800 transition-all hover:bg-amber-100 disabled:opacity-50"
						title="Отправить тестовое уведомление через SSE"
					>
						<Sparkles class="h-3.5 w-3.5 text-amber-600" strokeWidth={2} />
						<span>{isSendingTest ? 'Отправка...' : 'Тест SSE'}</span>
					</button>

					{#if notifHasUnread}
						<button
							type="button"
							onclick={() => notifications.markAllAsRead()}
							class="flex items-center gap-1.5 rounded-xl bg-white px-3 py-1 text-xs font-medium text-zinc-700 shadow-xs transition-colors hover:bg-zinc-100"
						>
							<CheckCheck class="h-3.5 w-3.5 text-zinc-600" strokeWidth={2} />
							<span>Прочитать все</span>
						</button>
					{/if}
				</div>
			</div>

			<div class="space-y-2.5">
				{#if notifIsLoading && notifItems.length === 0}
					<div class="flex flex-col items-center justify-center py-12 text-center text-zinc-400">
						<div class="h-6 w-6 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"></div>
						<p class="mt-2 text-xs">Загрузка уведомлений...</p>
					</div>
				{:else if notifDisplayedItems.length === 0}
					<div class="flex flex-col items-center justify-center py-12 text-center">
						<div class="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400">
							<Inbox class="h-6 w-6" strokeWidth={1.5} />
						</div>
						<p class="text-sm font-semibold text-zinc-700">
							{notifTab === 'unread' ? 'Нет непрочитанных уведомлений' : 'У вас пока нет уведомлений'}
						</p>
						<p class="mt-0.5 text-xs text-zinc-400">
							{notifTab === 'unread' ? 'Все уведомления прочитаны' : 'Здесь будут появляться оповещения о событиях в сервисе'}
						</p>
					</div>
				{:else}
					{#each notifDisplayedItems as item (item.id)}
						<NotificationItem notification={item} compact={isMobile} />
					{/each}
				{/if}
			</div>

			<div class="pt-6 border-t border-zinc-100">
				<h3 class="text-base font-semibold text-zinc-900 mb-3">Каналы уведомлений</h3>
				<AccountSettingsRows rows={isMobile ? mobileRows : activeRows} onAction={handleSectionAction} variant={isMobile ? 'mobile' : 'desktop'} />
			</div>
		</div>
	{/snippet}

	<input
		bind:this={avatarInputRef}
		type="file"
		accept="image/*"
		class="hidden"
		onchange={handleAvatarChange}
		disabled={avatarUploading}
	/>
{/if}
