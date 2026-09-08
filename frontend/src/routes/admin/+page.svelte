<script lang="ts">
	import { onMount } from 'svelte';
	import { adminOverviewApi } from '$lib/api/admin/overview';
	import type { AdminOverviewMetrics } from '$lib/types/admin';
	import {
		Clock,
		ShieldCheck,
		AlertTriangle,
		Ban,
		Building2,
		Users,
		ArrowUpRight,
		CheckCircle2,
		RotateCw,
		Loader2,
		Sparkles
	} from 'lucide-svelte';

	let metrics = $state<AdminOverviewMetrics | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let isRefreshing = $state(false);

	function fetchOverview() {
		isRefreshing = true;
		error = null;
		adminOverviewApi
			.getOverview()
			.then((res) => {
				metrics = res;
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки сводки показателей';
			})
			.finally(() => {
				isLoading = false;
				isRefreshing = false;
			});
	}

	onMount(() => {
		fetchOverview();
	});
</script>

<svelte:head>
	<title>Панель управления — Flickey Admin</title>
</svelte:head>

<div class="space-y-8">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<div class="flex items-center gap-2.5">
				<h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-foreground">
					Панель управления Flickey
				</h1>
				<span class="rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-bold text-emerald-700 border border-emerald-200 dark:bg-emerald-950 dark:border-emerald-900 dark:text-emerald-400">
					Live
				</span>
			</div>
			<p class="mt-1 text-sm text-zinc-500 dark:text-muted-foreground">
				Сводка очереди модерации, верификации партнеров, обращений пользователей и метрик платформы.
			</p>
		</div>

		<button
			type="button"
			disabled={isRefreshing}
			onclick={fetchOverview}
			class="inline-flex items-center gap-2 rounded-full border border-zinc-200/80 bg-white px-4 py-2 text-xs font-semibold text-zinc-700 shadow-2xs transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground cursor-pointer"
		>
			<RotateCw class="h-3.5 w-3.5 {isRefreshing ? 'animate-spin text-zinc-900 dark:text-foreground' : ''}" />
			<span>Обновить данные</span>
		</button>
	</div>

	{#if isLoading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 animate-pulse">
			{#each Array(4) as _}
				<div class="h-36 rounded-3xl bg-white border border-zinc-200/80 p-6 shadow-2xs dark:bg-card dark:border-border"></div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={fetchOverview}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Попробовать снова
			</button>
		</div>
	{:else if metrics}
		<!-- 4 Key Metric Action Cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<!-- Pending Listings -->
			<div class="group relative flex flex-col justify-between rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-2xs transition-all hover:border-amber-300 hover:shadow-sm dark:border-border dark:bg-card">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-muted-foreground">
						Модерация жилья
					</span>
					<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300">
						<Clock class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-black tracking-tight text-zinc-900 dark:text-foreground">
						{metrics.pending_listings}
					</div>
					<div class="mt-2 flex items-center justify-between">
						<span class="text-xs text-zinc-500 dark:text-muted-foreground">ожидают решения</span>
						<a
							href="/admin/listings?status=pending_review"
							class="inline-flex items-center gap-1 text-xs font-bold text-amber-600 hover:text-amber-700 dark:text-amber-400"
						>
							<span>Открыть</span>
							<ArrowUpRight class="h-3.5 w-3.5 transition group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
						</a>
					</div>
				</div>
			</div>

			<!-- Pending Verifications -->
			<div class="group relative flex flex-col justify-between rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-2xs transition-all hover:border-blue-300 hover:shadow-sm dark:border-border dark:bg-card">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-muted-foreground">
						Верификация
					</span>
					<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
						<ShieldCheck class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-black tracking-tight text-zinc-900 dark:text-foreground">
						{metrics.pending_verifications}
					</div>
					<div class="mt-2 flex items-center justify-between">
						<span class="text-xs text-zinc-500 dark:text-muted-foreground">заявок на проверку</span>
						<a
							href="/admin/verification"
							class="inline-flex items-center gap-1 text-xs font-bold text-blue-600 hover:text-blue-700 dark:text-blue-400"
						>
							<span>Открыть</span>
							<ArrowUpRight class="h-3.5 w-3.5 transition group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
						</a>
					</div>
				</div>
			</div>

			<!-- Open Reports -->
			<div class="group relative flex flex-col justify-between rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-2xs transition-all hover:border-rose-300 hover:shadow-sm dark:border-border dark:bg-card">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-muted-foreground">
						Открытые жалобы
					</span>
					<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-rose-50 text-rose-700 dark:bg-rose-950 dark:text-rose-300">
						<AlertTriangle class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-black tracking-tight text-zinc-900 dark:text-foreground">
						{metrics.open_reports}
					</div>
					<div class="mt-2 flex items-center justify-between">
						<span class="text-xs text-zinc-500 dark:text-muted-foreground">требуют разбора</span>
						<a
							href="/admin/reports"
							class="inline-flex items-center gap-1 text-xs font-bold text-rose-600 hover:text-rose-700 dark:text-rose-400"
						>
							<span>Открыть</span>
							<ArrowUpRight class="h-3.5 w-3.5 transition group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
						</a>
					</div>
				</div>
			</div>

			<!-- Blocked Users -->
			<div class="group relative flex flex-col justify-between rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-2xs transition-all hover:border-zinc-400 hover:shadow-sm dark:border-border dark:bg-card">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-muted-foreground">
						Ограниченные юзеры
					</span>
					<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300">
						<Ban class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-black tracking-tight text-zinc-900 dark:text-foreground">
						{metrics.blocked_users}
					</div>
					<div class="mt-2 flex items-center justify-between">
						<span class="text-xs text-zinc-500 dark:text-muted-foreground">активные санкции</span>
						<a
							href="/admin/users?status=suspended"
							class="inline-flex items-center gap-1 text-xs font-bold text-zinc-700 hover:text-zinc-900 dark:text-zinc-300 dark:hover:text-foreground"
						>
							<span>Открыть</span>
							<ArrowUpRight class="h-3.5 w-3.5 transition group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
						</a>
					</div>
				</div>
			</div>
		</div>

		<!-- Detailed Platform Breakdown Cards -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Listings Statistics Card -->
			<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-7 shadow-2xs dark:border-border dark:bg-card space-y-6">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-900 dark:bg-muted dark:text-foreground">
							<Building2 class="h-5 w-5" />
						</div>
						<div>
							<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">Каталог объявлений</h3>
							<p class="text-xs text-zinc-500 dark:text-muted-foreground">Текущий статус базы объектов посуточной аренды</p>
						</div>
					</div>

					<a
						href="/admin/listings"
						class="text-xs font-semibold text-zinc-600 hover:text-zinc-900 hover:underline dark:text-muted-foreground dark:hover:text-foreground"
					>
						Все объявления →
					</a>
				</div>

				<div class="grid grid-cols-3 gap-3 text-center">
					<div class="rounded-2xl border border-zinc-100 bg-zinc-50/80 p-4 dark:border-border dark:bg-muted/40">
						<div class="text-2xl font-black text-zinc-900 dark:text-foreground">
							{metrics.total_listings}
						</div>
						<div class="mt-1 text-[11px] font-medium text-zinc-500 dark:text-muted-foreground">
							Всего в базе
						</div>
					</div>

					<div class="rounded-2xl border border-emerald-100 bg-emerald-50/70 p-4 dark:border-emerald-950 dark:bg-emerald-950/40">
						<div class="text-2xl font-black text-emerald-700 dark:text-emerald-400">
							{metrics.approved_listings}
						</div>
						<div class="mt-1 text-[11px] font-semibold text-emerald-700 dark:text-emerald-400">
							Опубликовано
						</div>
					</div>

					<div class="rounded-2xl border border-rose-100 bg-rose-50/70 p-4 dark:border-rose-950 dark:bg-rose-950/40">
						<div class="text-2xl font-black text-rose-700 dark:text-rose-400">
							{metrics.rejected_listings}
						</div>
						<div class="mt-1 text-[11px] font-semibold text-rose-700 dark:text-rose-400">
							Отклонено
						</div>
					</div>
				</div>
			</div>

			<!-- Users Statistics Card -->
			<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-7 shadow-2xs dark:border-border dark:bg-card space-y-6">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-900 dark:bg-muted dark:text-foreground">
							<Users class="h-5 w-5" />
						</div>
						<div>
							<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">Пользователи сервиса</h3>
							<p class="text-xs text-zinc-500 dark:text-muted-foreground">Зарегистрированные гости, арендодатели и персонал</p>
						</div>
					</div>

					<a
						href="/admin/users"
						class="text-xs font-semibold text-zinc-600 hover:text-zinc-900 hover:underline dark:text-muted-foreground dark:hover:text-foreground"
					>
						Все пользователи →
					</a>
				</div>

				<div class="grid grid-cols-3 gap-3 text-center">
					<div class="rounded-2xl border border-zinc-100 bg-zinc-50/80 p-4 dark:border-border dark:bg-muted/40">
						<div class="text-2xl font-black text-zinc-900 dark:text-foreground">
							{metrics.total_users}
						</div>
						<div class="mt-1 text-[11px] font-medium text-zinc-500 dark:text-muted-foreground">
							Регистраций
						</div>
					</div>

					<div class="rounded-2xl border border-emerald-100 bg-emerald-50/70 p-4 dark:border-emerald-950 dark:bg-emerald-950/40">
						<div class="text-2xl font-black text-emerald-700 dark:text-emerald-400">
							{metrics.active_users}
						</div>
						<div class="mt-1 text-[11px] font-semibold text-emerald-700 dark:text-emerald-400">
							Активных
						</div>
					</div>

					<div class="rounded-2xl border border-amber-100 bg-amber-50/70 p-4 dark:border-amber-950 dark:bg-amber-950/40">
						<div class="text-2xl font-black text-amber-700 dark:text-amber-400">
							{metrics.blocked_users}
						</div>
						<div class="mt-1 text-[11px] font-semibold text-amber-700 dark:text-amber-400">
							С ограничением
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>