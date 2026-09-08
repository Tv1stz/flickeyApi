<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authStore, type User } from '$lib/stores/authStore.svelte';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import { hostListingsStore } from '$lib/stores/listingsStore.svelte';
	import { toast } from '$lib/stores/toastStore';

	import {
		Check,
		ChevronRight,
		BarChart3,
		Heart,
		LayoutGrid,
		LogOut,
		User as UserIcon,
		CreditCard,
		HelpCircle,
		ShieldCheck
	} from 'lucide-svelte';
	import type { Listing } from '$lib/components/card/types';

	let authInitialized = $derived(authStore.initialized);
	let authenticated = $derived(authStore.isAuthenticated);
	let user = $derived(authStore.user);
	let initials = $derived(authStore.userInitials);
	let verified = $derived(authStore.isVerifiedHost);
	let hostMode = $derived(authStore.isHostMode);
	let canSwitch = $derived(authStore.canSwitchModes);
	
	let favoriteCount = $derived(favoritesStore.favoriteListings.length);
	let hostListingItems = $derived(hostListingsStore.items);
	let redirectingToAuth = $state(false);

	$effect(() => {
		if (!authStore.initialized || authStore.user || redirectingToAuth) return;
		redirectingToAuth = true;
		authStore.setPendingAction(null, '/profile');
		goto(resolve('/auth'));
	});

	function handleLogout() {
		authStore.logout();
		goto(resolve('/'));
	}

	function handleSoon(feature: string) {
		toast.info('Скоро', `${feature} будут доступны в следующем обновлении`);
	}

	const activeHostListings = $derived(hostListingItems.filter((item) => item.isActive).length);
	const dashboardActionClass =
		'group relative w-full p-4 text-left transition-all duration-300 ease-out hover:bg-zinc-50 active:bg-zinc-100 sm:block sm:overflow-hidden sm:rounded-2xl sm:border sm:border-zinc-200 sm:bg-white sm:p-6 sm:shadow-sm sm:hover:-translate-y-1 sm:hover:border-black/10 sm:hover:bg-white sm:hover:shadow-xl sm:active:scale-[0.98] sm:active:bg-transparent';

	let isAdmin = $derived(authStore.user?.roles?.includes('admin') || Boolean(user?.roles?.includes('admin')));

	type DashboardRoute = '/host/listings' | '/host/stats' | '/favorites' | '/help' | '/admin';

	type DashboardCardType = {
		label: string;
		description: string;
		icon: any;
		href?: DashboardRoute;
		action?: string;
	};

	type DashboardSection = {
		title: string;
		items: DashboardCardType[];
	};

	const dashboardSections = $derived.by(() => {
		const sections: DashboardSection[] = [];

		if (isAdmin) {
			sections.push({
				title: 'Администрирование платформы',
				items: [
					{
						label: 'Панель администратора',
						description: 'Модерация объявлений, пользователи, верификация',
						icon: ShieldCheck,
						href: '/admin'
					}
				]
			});
		}

		if (hostMode) {
			sections.push({
				title: 'Управление жильем',
				items: [
					{
						label: 'Мои объявления',
						description: `${activeHostListings} активных`,
						icon: LayoutGrid,
						href: '/host/listings'
					},
					{
						label: 'Статистика',
						description: 'Просмотры и доходы',
						icon: BarChart3,
						href: '/host/stats'
					}
				]
			});
		} else {
			sections.push({
				title: 'Путешествия',
				items: [
					{
						label: 'Избранное',
						description: `${favoriteCount} сохранено`,
						icon: Heart,
						href: '/favorites'
					}
				]
			});
		}

		sections.push({
			title: 'Настройки аккаунта',
			items: [
				{
					label: 'Личные данные + Добавить Юр.информация',
					description: 'Редактировать профиль',
					icon: UserIcon,
					action: 'Редактирование профиля'
				},
				{
					label: 'Безопасность',
					description: 'Двухфакторная аутентификация',
					icon: ShieldCheck,
					action: 'Настройки безопасности'
				},
				{
					label: 'Платежи',
					description: 'Карты',
					icon: CreditCard,
					action: 'Платежные настройки'
				}
			]
		});

		sections.push({
			title: 'Поддержка',
			items: [
				{
					label: 'Помощь',
					description: 'Служба поддержки',
					icon: HelpCircle,
					href: '/help'
				}
			]
		});

		return sections;
	});
</script>

<svelte:head>
	<title>Аккаунт — Flickey</title>
</svelte:head>

<!-- Сниппет: адаптирован под оба вида (строка на мобильных, колонка на ПК) -->
{#snippet dashboardCardContent(item: DashboardCardType)}
	<div
		class="absolute inset-0 bg-gradient-to-br from-zinc-50 to-transparent opacity-0 transition-opacity duration-300 sm:group-hover:opacity-100"
	></div>
	<div class="relative flex w-full items-center sm:h-full sm:flex-col sm:items-start">
		<item.icon class="h-6 w-6 shrink-0 text-zinc-900 sm:mb-4 sm:h-7 sm:w-7" strokeWidth={1.5} />
		<div class="ml-4 flex flex-1 flex-col sm:ml-0">
			<div class="font-medium text-zinc-900">{item.label}</div>
			<div class="mt-0.5 text-[13px] text-zinc-500 sm:mt-1 sm:text-sm">{item.description}</div>
		</div>
		<!-- Стрелочка видна только на мобильных, дополняет ощущение "iOS меню" -->
		<ChevronRight class="ml-2 h-5 w-5 shrink-0 text-zinc-300 sm:hidden" />
	</div>
{/snippet}

{#if !authInitialized}
	<div class="flex min-h-[50vh] items-center justify-center">
		<div class="flex flex-col items-center gap-3">
			<div
				class="h-8 w-8 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"
			></div>
			<p class="text-sm text-zinc-500">Загрузка аккаунта...</p>
		</div>
	</div>
{:else if !authenticated}
	<div class="flex min-h-[50vh] items-center justify-center">
		<p class="text-sm text-zinc-500">Перенаправляем на страницу входа...</p>
	</div>
{:else}
	<div class="min-h-dvh bg-white pt-6 pb-20 sm:pt-12">
		<div class="mx-auto max-w-6xl px-4 sm:px-6">
			<h1 class="mb-8 text-3xl font-bold tracking-tight text-zinc-900 sm:mb-12 sm:text-[32px]">
				Аккаунт
			</h1>

			<div class="flex flex-col gap-8 lg:flex-row lg:items-stretch lg:gap-16">
				<!-- ─── Left Column: Only Profile Card now ─── -->
				<aside class="w-full lg:w-[30%] lg:min-w-80">
					<div
						class="relative overflow-hidden rounded-3xl border border-black/5 bg-gradient-to-br from-zinc-50 to-white p-5 shadow-lg sm:p-8"
					>
						<div
							class="absolute inset-0 bg-[radial-gradient(circle_at_top_right,_black_0%,_transparent_70%)] opacity-[0.03]"
						></div>

						<div
							class="relative flex flex-col items-center gap-4 text-center sm:flex-row sm:items-center sm:gap-6 sm:text-left"
						>
							<div class="relative shrink-0">
								<div
									class="flex h-20 w-20 items-center justify-center overflow-hidden rounded-full bg-zinc-900 text-xl font-semibold text-white shadow-xl ring-4 ring-white sm:h-24 sm:w-24 sm:text-2xl"
								>
									{#if user?.avatar}
										<img src={user.avatar} alt={user.name} class="h-full w-full object-cover" />
									{:else}
										{initials}
									{/if}
								</div>
								{#if verified}
									<div
										class="absolute -right-1 -bottom-1 flex h-7 w-7 items-center justify-center rounded-full bg-[#FF385C] shadow-md ring-4 ring-white sm:h-8 sm:w-8"
									>
										<ShieldCheck class="h-3.5 w-3.5 text-white sm:h-4 sm:w-4" />
									</div>
								{/if}
							</div>

							<div class="min-w-0">
								<h2 class="truncate text-xl font-semibold tracking-tight text-zinc-900 sm:text-2xl">
									{user?.name || 'Пользователь'}
								</h2>
								<p class="mt-0.5 text-[13px] font-medium text-zinc-500 sm:mt-1 sm:text-sm">
									{hostMode ? 'Хозяин жилья' : 'Гость'}
								</p>
							</div>
						</div>

						<div class="my-6 border-t border-zinc-200 sm:my-8"></div>

						<div class="relative">
							<h3 class="mb-3 text-base font-semibold text-zinc-900 sm:mb-4 sm:text-lg">
								Подтвержденная информация
							</h3>
							<ul class="space-y-2.5 sm:space-y-3">
								{#if verified}
									<li class="flex items-center gap-3">
										<Check class="h-4 w-4 text-zinc-900 sm:h-5 sm:w-5" />
										<span class="text-sm text-zinc-700 sm:text-[15px]">Личность подтверждена</span>
									</li>
								{/if}
								<li class="flex items-center gap-3">
									<Check class="h-4 w-4 text-zinc-900 sm:h-5 sm:w-5" />
									<span class="text-sm text-zinc-700 sm:text-[15px]">Эл. почта подтверждена</span>
								</li>
								{#if user?.phone}
									<li class="flex items-center gap-3">
										<Check class="h-4 w-4 text-zinc-900 sm:h-5 sm:w-5" />
										<span class="text-sm text-zinc-700 sm:text-[15px]">Номер телефона</span>
									</li>
								{/if}
							</ul>
						</div>
					</div>
				</aside>

				<!-- ─── Right Column ─── -->
				<main class="flex flex-1 flex-col">
					<div class="mb-2 sm:mb-4">
						<h2 class="text-xl font-semibold tracking-tight text-zinc-900 sm:text-2xl">
							Привет, {user?.name?.split(' ')[0] || 'друг'}!
						</h2>
						<p class="mt-1 text-sm text-zinc-500 sm:mt-2 sm:text-base">
							{canSwitch
								? `Вы находитесь в режиме ${hostMode ? 'управления жильем' : 'гостя'}.`
								: 'Управляйте своим профилем и настройками здесь.'}
						</p>
					</div>

					<div class="flex-1">
						{#each dashboardSections as section (section.title)}
							<section class="mt-8 first:mt-6 sm:mt-12 sm:first:mt-8">
								<h3 class="mb-3 text-base font-semibold text-zinc-900 sm:mb-4 sm:text-lg">
									{section.title}
								</h3>

								<!-- На мобильных это единый блок со списком (flex-col, divide-y). На десктопе - сетка (grid, gap-4). -->
								<div
									class="flex flex-col divide-y divide-zinc-100 overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-sm sm:grid sm:grid-cols-2 sm:gap-4 sm:divide-y-0 sm:overflow-visible sm:border-none sm:bg-transparent sm:shadow-none lg:grid-cols-3"
								>
									{#each section.items as item (item.label)}
										<!-- Классы настроены так, чтобы элемент выглядел как строка списка на мобильных, и как отдельная красивая карточка на sm+ -->
										{#if item.href}
											<a href={resolve(item.href)} class={dashboardActionClass}>
												{@render dashboardCardContent(item)}
											</a>
										{:else}
											<button
												type="button"
												onclick={() => handleSoon(item.action || '')}
												class={dashboardActionClass}
											>
												{@render dashboardCardContent(item)}
											</button>
										{/if}
									{/each}
								</div>
							</section>
						{/each}
					</div>

					{#if authenticated}
						<div class="mt-12 flex justify-start border-t border-zinc-200 pt-6 sm:mt-16 sm:pt-8">
							<button
								type="button"
								onclick={handleLogout}
								class="group flex items-center gap-2 text-sm font-medium text-zinc-500 transition-all duration-300 ease-out hover:text-zinc-900 active:scale-95 sm:text-base"
							>
								<LogOut
									class="h-4 w-4 transition-transform duration-300 ease-out group-hover:-translate-x-1 sm:h-5 sm:w-5"
								/>
								Выйти из аккаунта
							</button>
						</div>
					{/if}
				</main>
			</div>
		</div>
	</div>
{/if}
