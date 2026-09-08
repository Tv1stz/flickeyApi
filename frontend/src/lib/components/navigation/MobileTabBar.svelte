<!-- src/lib/components/navigation/MobileTabBar.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	import { auth } from '$lib/auth/auth.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import { Heart, Home, User as UserIcon, BarChart3, LayoutGrid, Plus, Calendar } from 'lucide-svelte';

	let currentPath = $derived($page.url.pathname);
	let authenticated = $derived(auth.isAuthenticated || authStore.isAuthenticated);
	let hostMode = $derived(authStore.isHostMode);
	let user = $derived(authStore.user);
	let initials = $derived(authStore.userInitials);

	const isListingDetailPage = $derived(
		/^\/listings\/[^/]+$/.test(currentPath) && !currentPath.endsWith('/new')
	);
	const isListingCreationPage = $derived(currentPath === '/listings/new');

	type KnownRoute =
		| '/'
		| '/auth'
		| '/profile'
		| '/favorites'
		| '/host/listings'
		| '/host/calendar'
		| '/host/stats'
		| '/listings/new';

	type TabItem = {
		href: KnownRoute;
		icon: typeof Home;
		label: string;
		requiresAuth?: boolean;
		action?: 'create-listing';
		isCreate?: boolean;
	};

	const mobileTabs = $derived.by((): TabItem[] => {
		if (!authenticated) {
			return [
				{ href: '/', icon: Home, label: 'Главная' },
				{
					href: '/listings/new',
					icon: Plus,
					label: 'Создать',
					requiresAuth: true,
					action: 'create-listing',
					isCreate: true
				},
				{ href: '/auth', icon: UserIcon, label: 'Войти' }
			];
		}
		if (hostMode) {
			return [
				{ href: '/host/listings', icon: LayoutGrid, label: 'Объявления' },
				{ href: '/host/calendar', icon: Calendar, label: 'Календарь' },
				{ href: '/listings/new', icon: Plus, label: 'Создать', isCreate: true },
				{ href: '/host/stats', icon: BarChart3, label: 'Статистика' },
				{ href: '/profile', icon: UserIcon, label: 'Профиль' }
			];
		}
		return [
			{ href: '/', icon: Home, label: 'Главная' },
			{ href: '/favorites', icon: Heart, label: 'Избранное' },
			{ href: '/listings/new', icon: Plus, label: 'Создать', isCreate: true },
			{ href: '/profile', icon: UserIcon, label: 'Профиль' }
		];
	});

	function handleTabClick(tab: TabItem): boolean {
		if (tab.requiresAuth && !authenticated) {
			if (tab.action === 'create-listing') {
				authStore.setPendingAction('create-listing');
			}
			goto(resolve('/auth'));
			return true;
		}
		return false;
	}

	function isActive(href: string): boolean {
		if (href === '/') return currentPath === '/' || currentPath.startsWith('/search');
		return currentPath.startsWith(href);
	}
</script>

{#if !isListingDetailPage && !isListingCreationPage}
	<nav class="fixed right-0 bottom-0 left-0 z-40 lg:hidden" data-tab-bar>
		<div
			class="safe-bottom border-t border-zinc-200/60 bg-white/80 pb-1 backdrop-blur-2xl backdrop-saturate-150"
		>
			<div
				class="mx-auto grid max-w-md"
				style="grid-template-columns: repeat({mobileTabs.length}, 1fr);"
			>
				{#each mobileTabs as tab (tab.href + tab.label)}
					{@const active =
						(!tab.requiresAuth && isActive(tab.href)) || (tab.isCreate && isActive(tab.href))}
					{@const isProfile = tab.label === 'Профиль' && authenticated && !!user}
					{@const TabIcon = tab.icon}

					{#if tab.requiresAuth && !authenticated}
						<button
							type="button"
							class="group flex flex-col items-center justify-center gap-1 py-2
							       text-zinc-400 transition-colors duration-200 active:text-zinc-600"
							onclick={() => handleTabClick(tab)}
						>
							<TabIcon strokeWidth={1.75} size={24} />
							<span class="text-[10px] leading-none font-medium">{tab.label}</span>
						</button>
					{:else}
						<a
							href={resolve(tab.href)}
							class="group flex flex-col items-center justify-center gap-1 py-2
							       transition-colors duration-200
							       {active ? 'text-zinc-900' : 'text-zinc-400'}"
						>
							{#if isProfile}
								<div
									class="flex h-7 w-7 items-center justify-center overflow-hidden
							               rounded-full transition-all duration-200
							               {active ? 'ring-[2.5px] ring-zinc-900 ring-offset-1' : ''}"
								>
									{#if user?.avatar}
										<img src={user.avatar} alt="" class="h-full w-full object-cover" />
									{:else}
										<div
											class="flex h-full w-full items-center justify-center
							                       {active ? 'bg-zinc-900' : 'bg-zinc-200'}"
										>
											<span
												class="text-[10px] font-bold
							                           {active ? 'text-white' : 'text-zinc-500'}"
											>
												{initials}
											</span>
										</div>
									{/if}
								</div>
								<span class="text-[10px] leading-none font-medium">{tab.label}</span>
							{:else}
								<TabIcon strokeWidth={1.75} size={24} />
								<span class="text-[10px] leading-none font-medium">{tab.label}</span>
							{/if}
						</a>
					{/if}
				{/each}
			</div>
		</div>
	</nav>
{/if}
