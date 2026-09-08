<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import {
		LayoutDashboard,
		Building2,
		ShieldCheck,
		Users,
		AlertTriangle,
		FileText,
		ArrowUpRight,
		Compass,
		Home,
		User
	} from 'lucide-svelte';
	import { adminOverviewApi } from '$lib/api/admin/overview';
	import { authStore } from '$lib/stores/authStore.svelte';
	import type { AdminOverviewMetrics } from '$lib/types/admin';

	let metrics = $state<AdminOverviewMetrics | null>(null);

	onMount(() => {
		adminOverviewApi
			.getOverview()
			.then((res) => {
				metrics = res;
			})
			.catch(() => {
				// Metrics fail silently in nav
			});
	});

	const navItems = $derived([
		{ href: '/admin', label: 'Дашборд', icon: LayoutDashboard, exact: true, count: 0 },
		{
			href: '/admin/listings',
			label: 'Жилье',
			icon: Building2,
			exact: false,
			count: metrics?.pending_listings || 0,
			badgeColor: 'bg-amber-500 text-white'
		},
		{
			href: '/admin/verification',
			label: 'Верификация',
			icon: ShieldCheck,
			exact: false,
			count: metrics?.pending_verifications || 0,
			badgeColor: 'bg-blue-500 text-white'
		},
		{ href: '/admin/users', label: 'Пользователи', icon: Users, exact: false, count: 0 },
		{
			href: '/admin/reports',
			label: 'Жалобы',
			icon: AlertTriangle,
			exact: false,
			count: metrics?.open_reports || 0,
			badgeColor: 'bg-rose-500 text-white'
		},
		{ href: '/admin/audit', label: 'Аудит', icon: FileText, exact: false, count: 0 }
	]);

	function isActive(href: string, exact: boolean): boolean {
		const path = page.url.pathname;
		if (exact) {
			return path === href;
		}
		return path === href || path.startsWith(href + '/');
	}
</script>

<header class="sticky top-0 z-40 w-full border-b border-zinc-200/80 bg-white/90 backdrop-blur-md dark:border-border dark:bg-card/90 shadow-2xs">
	<div class="mx-auto max-w-[1440px] px-4 sm:px-6 lg:px-8">
		<div class="flex h-16 items-center justify-between gap-4">
			<!-- Brand & Badge -->
			<div class="flex items-center gap-6">
				<a href="/admin" class="group flex items-center gap-2.5 transition active:scale-98">
					<div class="flex h-9 w-9 items-center justify-center rounded-2xl bg-zinc-900 text-sm font-black text-white shadow-sm dark:bg-white dark:text-zinc-900">
						F
					</div>
					<div class="flex flex-col">
						<div class="flex items-center gap-1.5">
							<span class="text-base font-extrabold tracking-tight text-zinc-900 dark:text-foreground">Flickey</span>
							<span class="rounded-full bg-zinc-100 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-zinc-700 dark:bg-muted dark:text-muted-foreground">
								Admin
							</span>
						</div>
					</div>
				</a>

				<!-- Desktop Main Navigation -->
				<nav class="hidden lg:flex items-center gap-1">
					{#each navItems as item}
						{@const active = isActive(item.href, item.exact)}
						{@const Icon = item.icon}
						<a
							href={item.href}
							class="group relative flex items-center gap-2 rounded-full px-3.5 py-1.5 text-xs font-semibold transition-all {active
								? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
								: 'text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 dark:text-muted-foreground dark:hover:bg-muted dark:hover:text-foreground'}"
						>
							<Icon class="h-4 w-4 shrink-0 transition group-hover:scale-110" />
							<span>{item.label}</span>

							{#if item.count > 0}
								<span
									class="flex h-4 min-w-[16px] items-center justify-center rounded-full px-1 text-[10px] font-bold leading-none shadow-xs {active
										? 'bg-white text-zinc-900 dark:bg-zinc-900 dark:text-white'
										: item.badgeColor}"
								>
									{item.count}
								</span>
							{/if}
						</a>
					{/each}
				</nav>
			</div>

			<!-- Right Controls: Quick Navigation to Userland / Catalog -->
			<div class="flex items-center gap-2">
				<a
					href="/host/listings"
					class="hidden sm:inline-flex items-center gap-1.5 rounded-full border border-zinc-200/80 bg-white px-3 py-1.5 text-xs font-semibold text-zinc-700 shadow-2xs transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground"
					title="Перейти в кабинет хозяина"
				>
					<Home class="h-3.5 w-3.5 text-zinc-500" />
					<span>Мои объекты</span>
				</a>

				<a
					href="/search"
					class="inline-flex items-center gap-1.5 rounded-full border border-zinc-200/80 bg-white px-3 py-1.5 text-xs font-semibold text-zinc-700 shadow-2xs transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground"
					title="Перейти в каталог сайта"
				>
					<Compass class="h-3.5 w-3.5 text-zinc-500" />
					<span>Каталог</span>
				</a>

				<!-- Profile Info Pill -->
				{#if authStore.user}
					<a
						href="/profile"
						class="flex items-center gap-2 rounded-full border border-zinc-200/80 bg-zinc-50/80 px-2.5 py-1 text-xs font-semibold text-zinc-900 shadow-2xs transition hover:border-zinc-300 hover:bg-zinc-100 active:scale-95 dark:border-border dark:bg-muted dark:text-foreground"
					>
						<div class="flex h-6 w-6 items-center justify-center rounded-full bg-zinc-900 text-[10px] font-bold text-white dark:bg-white dark:text-zinc-900">
							A
						</div>
						<span class="hidden md:inline text-[11px] font-mono">
							{authStore.user.phone ? authStore.user.phone.slice(-4) : 'Admin'}
						</span>
					</a>
				{/if}
			</div>
		</div>

		<!-- Mobile/Tablet Sub-Navigation Bar -->
		<div class="lg:hidden flex items-center gap-1.5 overflow-x-auto pb-2.5 pt-0.5 scrollbar-none">
			{#each navItems as item}
				{@const active = isActive(item.href, item.exact)}
				{@const Icon = item.icon}
				<a
					href={item.href}
					class="flex items-center gap-1.5 whitespace-nowrap rounded-full px-3 py-1.5 text-xs font-semibold transition-all {active
						? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
						: 'text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 dark:text-muted-foreground dark:hover:bg-muted dark:hover:text-foreground'}"
				>
					<Icon class="h-3.5 w-3.5 shrink-0" />
					<span>{item.label}</span>
					{#if item.count > 0}
						<span
							class="flex h-4 min-w-[16px] items-center justify-center rounded-full px-1 text-[10px] font-bold leading-none {active
								? 'bg-white text-zinc-900 dark:bg-zinc-900 dark:text-white'
								: item.badgeColor}"
						>
							{item.count}
						</span>
					{/if}
				</a>
			{/each}
		</div>
	</div>
</header>