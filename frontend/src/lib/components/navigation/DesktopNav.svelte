<!-- src/lib/components/navigation/DesktopNav.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { auth } from '$lib/auth/auth.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { isHostUser } from '$lib/auth/permissions';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import { toast } from '$lib/stores/toastStore';
	import {
		ChevronDown,
		ChevronRight,
		Heart,
		Home,
		User as UserIcon,
		BarChart3,
		LayoutGrid,
		Plus,
		Shield,
		ShieldCheck,
		LogOut,
		HelpCircle,
		ArrowLeftRight,
		Calendar
	} from 'lucide-svelte';
	import { scale, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import Button from '$lib/components/ui/Button.svelte';
	import HeaderSearch from '$lib/components/search/panels/HeaderSearch.svelte';
	import { searchStore } from '$lib/stores/searchStore.svelte';
	import { headerSearchStore } from '$lib/stores/headerSearchStore.svelte';
	import { buildSearchUrl } from '$lib/components/search/taxonomy';
	import NotificationBell from '$lib/components/notifications/NotificationBell.svelte';

	interface Props {
		scrolled?: boolean;
	}
	let { scrolled = false }: Props = $props();

	let currentPath = $derived($page.url.pathname);
	let authenticated = $derived(auth.isAuthenticated || authStore.isAuthenticated);
	let user = $derived(
		authStore.user ||
			(auth.user
				? {
						id: auth.user.id,
						phone: auth.user.phone,
						name: auth.displayName,
						createdAt: auth.user.created_at || new Date().toISOString(),
						roles: (auth.user.role === 'admin'
							? ['guest', 'host', 'admin']
							: auth.user.role === 'host'
								? ['guest', 'host']
								: ['guest']) as ('guest' | 'host' | 'admin')[],
						hostProfile: { verificationStatus: 'verified' as const, listingsCount: 0 }
					}
				: null)
	);
	let initials = $derived(
		user?.name
			? user.name
					.split(' ')
					.filter(Boolean)
					.map((n) => n[0])
					.join('')
					.toUpperCase()
					.slice(0, 2)
			: authStore.userInitials
	);
	let verifiedHost = $derived(auth.isHost || authStore.isVerifiedHost);
	let hostMode = $derived(auth.activeContext === 'host' || authStore.isHostMode);
	let canSwitch = $derived(auth.isHost || authStore.canSwitchModes);
	let showCreateInHeader = $derived(authStore.showCreateButtonInHeader);
	let isAdmin = $derived(
		auth.isAdmin || authStore.user?.roles?.includes('admin') || Boolean(user?.roles?.includes('admin'))
	);

	const isListingDetailPage = $derived(/^\/listings\/[^/]+$/.test(currentPath) && !currentPath.endsWith('/new'));
	const isListingCreationPage = $derived(currentPath === '/listings/new');
	const isHomePage = $derived(currentPath === '/');
	const isSearchPage = $derived(currentPath.startsWith('/search'));

	const showHeaderSearch = $derived(isHomePage || isSearchPage);
	const headerSearchDim = $derived(showHeaderSearch && headerSearchStore.isExpanded && (scrolled || isSearchPage));

	let showUserMenu = $state(false);

	type KnownRoute = '/' | '/auth' | '/profile' | '/favorites' | '/host/listings' | '/host/calendar' | '/host/stats' | '/host/verification' | '/listings/new';
	type MenuItem = { href: KnownRoute; icon: any; label: string; description?: string; };

	const guestMenuItems: MenuItem[] = [
		{ href: '/favorites', icon: Heart, label: 'Избранное', description: 'Сохранённые объявления' }
	];

	const hostMenuItems: MenuItem[] = [
		{ href: '/host/listings', icon: LayoutGrid, label: 'Мои объявления', description: 'Управление жильём' },
		{ href: '/host/calendar', icon: Calendar, label: 'Календарь (iCal)', description: 'Синхронизация и занятость' },
		{ href: '/host/verification', icon: ShieldCheck, label: 'Верификация партнера', description: 'Юридические реквизиты (РБ)' },
		{ href: '/host/stats', icon: BarChart3, label: 'Статистика', description: 'Аналитика и доходы' }
	];

	export function closeUserMenu() {
		showUserMenu = false;
	}

	function handleLogout() {
		auth.logout();
		showUserMenu = false;
		goto(resolve('/'));
	}

	function handleOutsideClick(e: PointerEvent) {
		if (showUserMenu) {
			const target = e.target as HTMLElement;
			if (!target.closest('[data-user-menu]')) {
				showUserMenu = false;
			}
		}
	}

	function handleCreateListing() {
		if (!authenticated) {
			authStore.setPendingAction('create-listing');
			goto(resolve('/auth'));
			return;
		}
		if (!isHostUser(user)) {
			authStore.becomeHost();
			toast.success('Режим хозяина активирован', 'Теперь можно создавать объявления');
		}
		goto(resolve('/listings/new'));
	}

	function switchMode() {
		if (!isHostUser(user) || !canSwitch) {
			showUserMenu = false;
			toast.info('Режим хозяина недоступен', 'Сначала начните размещать жилье');
			return;
		}
		const newMode = authStore.viewMode === 'guest' ? 'host' : 'guest';
		authStore.setViewMode(newMode);
		showUserMenu = false;
		if (newMode === 'host') {
			toast.success('Режим хозяина', 'Управляйте объявлениями и статистикой');
			goto(resolve('/host/listings'));
		} else {
			toast.success('Режим гостя', 'Ищите и бронируйте жильё');
			goto(resolve('/'));
		}
	}
</script>

<svelte:window onpointerdown={handleOutsideClick} />

{#if !isListingCreationPage}
	<header
		class="right-0 left-0 z-50 hidden lg:block {isListingDetailPage ? 'relative' : 'fixed top-0'}"
	>
		<div
			class="absolute inset-0 transition-all duration-500 ease-out
                   {isListingDetailPage
				? 'border-b border-black/5 bg-white'
				: headerSearchDim
					? 'border-b border-black/5 bg-white/90 shadow-md backdrop-blur-2xl'
					: scrolled
						? 'border-b border-black/5 bg-white/80 shadow-sm backdrop-blur-2xl'
						: 'border-b border-transparent bg-white/60 backdrop-blur-xl'}"
		></div>

		<div class="relative mx-auto max-w-[1600px] px-8">
			<div class="grid grid-cols-[1fr_auto_1fr] grid-rows-[96px_auto] items-center gap-x-4">
				<a
					href={resolve(hostMode ? '/host/listings' : '/')}
					class="group col-start-1 row-start-1 shrink-0 justify-self-start"
				>
					<img src="/logo.svg" alt="Flickey" class="h-8 w-auto" />
				</a>

				<nav class="col-start-3 row-start-1 flex shrink-0 items-center gap-3 justify-self-end">

					{#if canSwitch}
						<Button
							variant="ghost"
							tone="neutral"
							size="lg"
							radius="pill"
							iconLeft={ArrowLeftRight}
							onclick={switchMode}
						>
							{hostMode ? 'В путешествие' : 'К приему гостей'}
						</Button>
					{/if}

					{#if showCreateInHeader && !hostMode}
						<Button
							variant="ghost"
							tone="neutral"
							size="lg"
							radius="pill"
							onclick={handleCreateListing}
						>
							Сдать жильё
						</Button>
					{/if}

					{#if authenticated && user}
						<NotificationBell />

						<div class="relative" data-user-menu>
							<button
								type="button"
								class="group flex h-11 items-center gap-2.5 rounded-full border border-zinc-200 bg-white/80 pr-4 pl-1.5 backdrop-blur-2xl transition-all duration-500 hover:shadow-lg active:scale-[0.97]"
								onclick={() => (showUserMenu = !showUserMenu)}
							>
								<div class="relative">
									<div class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-full bg-zinc-100 ring-1 ring-black/5 transition-transform duration-500 group-hover:scale-105">
										{#if user.avatar}
											<img src={user.avatar} alt="" class="h-full w-full object-cover" />
										{:else}
											<span class="text-sm font-medium text-zinc-600">{initials}</span>
										{/if}
									</div>
									{#if verifiedHost}
										<div class="absolute -right-0.5 -bottom-0.5 flex h-[14px] w-[14px] items-center justify-center rounded-full bg-white ring-1 ring-black/5" in:scale={{ duration: 400, easing: cubicOut }}>
											<Shield strokeWidth={2} class="h-2.5 w-2.5 text-zinc-900" />
										</div>
									{/if}
								</div>
								<ChevronDown class="h-4 w-4 text-zinc-500 transition-transform duration-500 ease-out {showUserMenu ? 'rotate-180' : ''}" strokeWidth={1.5} />
							</button>

							{#if showUserMenu}
								<div class="menu-surface absolute top-full right-0 z-[70] mt-6 w-70 origin-top-right p-2" in:fly={{ y: -16, duration: 500, easing: cubicOut }} out:scale={{ duration: 300, start: 0.96, opacity: 0, easing: cubicOut }} role="menu">
									<a href={resolve('/profile')} class="group/profile flex items-center gap-4 rounded-3xl p-3 transition-all duration-400 hover:bg-white/60 hover:shadow-sm" onclick={() => (showUserMenu = false)}>
										<div class="relative shrink-0">
											<div class="flex h-[52px] w-[52px] items-center justify-center overflow-hidden rounded-full bg-white shadow-sm ring-1 ring-black/5 transition-transform duration-500 group-hover/profile:scale-105">
												{#if user.avatar}
													<img src={user.avatar} alt="" class="h-full w-full object-cover" />
												{:else}
													<span class="text-lg font-medium text-zinc-700">{initials}</span>
												{/if}
											</div>
										</div>
										<div class="min-w-0 flex-1">
											<p class="truncate text-[16px] font-semibold tracking-tight text-zinc-900">{user.name}</p>
											<p class="truncate text-[13px] text-zinc-500">{user.phone}</p>
										</div>
										<ChevronRight class="h-5 w-5 text-zinc-300 transition-all duration-400 group-hover/profile:translate-x-1 group-hover/profile:text-zinc-600" strokeWidth={1.5} />
									</a>

									{#if isAdmin}
										<div class="mt-2 px-1">
											<a href={resolve('/admin')} class="group/item flex items-center gap-4 rounded-2xl p-3 bg-emerald-50 text-emerald-800 transition-all duration-300 hover:bg-emerald-100" role="menuitem" onclick={() => (showUserMenu = false)}>
												<Shield strokeWidth={1.75} class="h-[20px] w-[20px] text-emerald-600" />
												<div class="min-w-0 flex-1">
													<p class="text-[15px] font-semibold text-emerald-900">Админ-панель</p>
													<p class="text-[12px] text-emerald-600">Модерация объявлений</p>
												</div>
											</a>
										</div>
									{/if}

									<div class="my-2 flex justify-center opacity-70">
										<div class="h-px w-[85%] bg-gradient-to-r from-transparent via-zinc-300/60 to-transparent"></div>
									</div>

									{#if hostMode}
										<div class="mb-3 px-1">
											<a href={resolve('/listings/new')} class="group/cta relative flex items-center justify-between overflow-hidden rounded-3xl bg-zinc-900/95 p-3 text-white shadow-lg backdrop-blur-xl transition-all duration-500 hover:scale-[1.02] hover:shadow-xl active:scale-[0.98]" onclick={() => (showUserMenu = false)}>
												<div class="absolute inset-0 -translate-x-full bg-gradient-to-r from-transparent via-white/10 to-transparent transition-transform duration-1000 ease-out group-hover/cta:translate-x-full"></div>
												<div class="relative min-w-0 flex-1">
													<p class="text-[15px] font-semibold tracking-wide">Создать объявление</p>
													<p class="text-[13px] font-medium text-zinc-400">Сдать жильё в аренду</p>
												</div>
												<div class="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white/10 backdrop-blur-md transition-transform duration-500 group-hover/cta:rotate-90 group-hover/cta:bg-white/20">
													<Plus class="h-5 w-5" strokeWidth={1.5} />
												</div>
											</a>
										</div>
									{/if}

									<div class="flex flex-col gap-1 px-1">
										{#each hostMode ? hostMenuItems : guestMenuItems as item (item.href)}
											{@const MenuIcon = item.icon}
											<a href={resolve(item.href)} class="group/item flex items-center gap-4 rounded-2xl p-3 transition-all duration-400 hover:bg-zinc-200/60" role="menuitem" onclick={() => (showUserMenu = false)}>
												<MenuIcon strokeWidth={1.75} class="h-[20px] w-[20px] text-zinc-600 transition-colors duration-400 group-hover/item:text-zinc-900" />
												<div class="min-w-0 flex-1">
													<p class="text-[15px] font-medium text-zinc-600 transition-colors duration-400 group-hover/item:text-zinc-900">{item.label}</p>
												</div>
											</a>
										{/each}
									</div>

									<div class="my-2 flex justify-center opacity-70">
										<div class="h-px w-[85%] bg-gradient-to-r from-transparent via-zinc-300/60 to-transparent"></div>
									</div>

									<div class="flex flex-col gap-1 px-1 pb-1">
										<a href={resolve('/help')} class="group/help flex items-center gap-4 rounded-2xl p-3 transition-all duration-400 hover:bg-zinc-200/60" onclick={() => (showUserMenu = false)}>
											<HelpCircle strokeWidth={1.75} class="h-[20px] w-[20px] text-zinc-600 transition-colors duration-400 group-hover/help:text-zinc-900" />
											<span class="text-[15px] font-medium text-zinc-600 transition-colors duration-400 group-hover/help:text-zinc-900">Помощь</span>
										</a>
										<button type="button" class="group/logout flex items-center gap-4 rounded-2xl p-3 transition-all duration-400 hover:bg-red-50" onclick={handleLogout}>
											<LogOut strokeWidth={1.75} class="h-[20px] w-[20px] text-zinc-600 transition-colors duration-400 group-hover/logout:text-red-500" />
											<span class="text-[15px] font-medium text-zinc-600 transition-colors duration-400 group-hover/logout:text-red-600">Выйти</span>
										</button>
									</div>
								</div>
							{/if}
						</div>
					{:else}
						<Button href={resolve('/auth')} variant="solid" tone="primary" size="lg" radius="pill">Войти</Button>
					{/if}
				</nav>

				{#if showHeaderSearch}
					<HeaderSearch
						enabled={showHeaderSearch}
						compactByDefault={isSearchPage}
						filteredCount={searchStore.resultsCount}
						alwaysShowFilters={isSearchPage}
						onSearch={() => goto(buildSearchUrl(searchStore.params))}
					/>
				{/if}
			</div>
		</div>
	</header>
{/if}
