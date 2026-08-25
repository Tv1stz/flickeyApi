<script lang="ts">
	import { adminOverviewApi } from '$lib/api/admin/overview';
	import type { AdminOverviewMetrics } from '$lib/types/admin';
	import { Building2, ShieldCheck, Users, AlertTriangle, ArrowUpRight, CheckCircle2, Clock, Ban } from 'lucide-svelte';

	let metrics = $state<AdminOverviewMetrics | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	$effect(() => {
		adminOverviewApi
			.getOverview()
			.then((res) => {
				metrics = res;
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки метрик платформ';
			})
			.finally(() => {
				isLoading = false;
			});
	});
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-foreground tracking-tight">Панель управления Flickey Compliance</h1>
		<p class="text-sm text-muted-foreground">Централизованный мониторинг жилья, верификации партнеров, жалоб и дисциплинарных мер.</p>
	</div>

	{#if isLoading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 animate-pulse">
			{#each Array(4) as _}
				<div class="h-32 bg-muted/40 rounded-2xl border border-border"></div>
			{/each}
		</div>
	{:else if error}
		<div class="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-sm">
			{error}
		</div>
	{:else if metrics}
		<!-- Metrics Cards Grid -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<!-- Pending Listings Card -->
			<div class="p-5 rounded-2xl bg-card border border-border shadow-xs hover:border-amber-500/50 transition-all flex flex-col justify-between">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Ожидают модерации</span>
					<div class="p-2.5 rounded-xl bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300">
						<Clock class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-extrabold text-foreground">{metrics.pending_listings}</div>
					<a href="/admin/listings?status=pending_review" class="text-xs font-semibold text-amber-600 dark:text-amber-400 hover:underline inline-flex items-center gap-1 mt-1">
						<span>Открыть очередь</span>
						<ArrowUpRight class="h-3.5 w-3.5" />
					</a>
				</div>
			</div>

			<!-- Pending Verifications Card -->
			<div class="p-5 rounded-2xl bg-card border border-border shadow-xs hover:border-blue-500/50 transition-all flex flex-col justify-between">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Заявки верификации</span>
					<div class="p-2.5 rounded-xl bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300">
						<ShieldCheck class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-extrabold text-foreground">{metrics.pending_verifications}</div>
					<a href="/admin/verification" class="text-xs font-semibold text-blue-600 dark:text-blue-400 hover:underline inline-flex items-center gap-1 mt-1">
						<span>Проверить документы</span>
						<ArrowUpRight class="h-3.5 w-3.5" />
					</a>
				</div>
			</div>

			<!-- Open Reports Card -->
			<div class="p-5 rounded-2xl bg-card border border-border shadow-xs hover:border-rose-500/50 transition-all flex flex-col justify-between">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Открытые жалобы</span>
					<div class="p-2.5 rounded-xl bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300">
						<AlertTriangle class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-extrabold text-foreground">{metrics.open_reports}</div>
					<a href="/admin/reports" class="text-xs font-semibold text-rose-600 dark:text-rose-400 hover:underline inline-flex items-center gap-1 mt-1">
						<span>Разрешить конфликты</span>
						<ArrowUpRight class="h-3.5 w-3.5" />
					</a>
				</div>
			</div>

			<!-- Blocked Users Card -->
			<div class="p-5 rounded-2xl bg-card border border-border shadow-xs hover:border-slate-500/50 transition-all flex flex-col justify-between">
				<div class="flex items-center justify-between">
					<span class="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Заблокированные аккаунты</span>
					<div class="p-2.5 rounded-xl bg-slate-100 text-slate-800 dark:bg-slate-900 dark:text-slate-300">
						<Ban class="h-5 w-5" />
					</div>
				</div>
				<div class="mt-4">
					<div class="text-3xl font-extrabold text-foreground">{metrics.blocked_users}</div>
					<a href="/admin/users?status=suspended" class="text-xs font-semibold text-slate-600 dark:text-slate-400 hover:underline inline-flex items-center gap-1 mt-1">
						<span>Управление пользователями</span>
						<ArrowUpRight class="h-3.5 w-3.5" />
					</a>
				</div>
			</div>
		</div>

		<!-- Quick Action Queues Summary -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 pt-4">
			<div class="p-6 rounded-2xl bg-card border border-border space-y-4">
				<h3 class="text-base font-bold text-foreground flex items-center gap-2">
					<Building2 class="h-5 w-5 text-primary" />
					<span>Статистика объектов жилья</span>
				</h3>
				<div class="grid grid-cols-3 gap-3 text-center">
					<div class="p-4 rounded-xl bg-muted/30 border border-border">
						<div class="text-2xl font-extrabold text-foreground">{metrics.total_listings}</div>
						<div class="text-[11px] text-muted-foreground font-medium mt-1">Всего создано</div>
					</div>
					<div class="p-4 rounded-xl bg-emerald-50 text-emerald-900 border border-emerald-200 dark:bg-emerald-950 dark:text-emerald-200 dark:border-emerald-900">
						<div class="text-2xl font-extrabold text-emerald-700 dark:text-emerald-400">{metrics.approved_listings}</div>
						<div class="text-[11px] font-semibold mt-1">Опубликовано</div>
					</div>
					<div class="p-4 rounded-xl bg-rose-50 text-rose-900 border border-rose-200 dark:bg-rose-950 dark:text-rose-200 dark:border-rose-900">
						<div class="text-2xl font-extrabold text-rose-700 dark:text-rose-400">{metrics.rejected_listings}</div>
						<div class="text-[11px] font-semibold mt-1">Отклонено</div>
					</div>
				</div>
			</div>

			<div class="p-6 rounded-2xl bg-card border border-border space-y-4">
				<h3 class="text-base font-bold text-foreground flex items-center gap-2">
					<Users class="h-5 w-5 text-primary" />
					<span>Состояние базы пользователей</span>
				</h3>
				<div class="grid grid-cols-3 gap-3 text-center">
					<div class="p-4 rounded-xl bg-muted/30 border border-border">
						<div class="text-2xl font-extrabold text-foreground">{metrics.total_users}</div>
						<div class="text-[11px] text-muted-foreground font-medium mt-1">Зарегистрировано</div>
					</div>
					<div class="p-4 rounded-xl bg-emerald-50 text-emerald-900 border border-emerald-200 dark:bg-emerald-950 dark:text-emerald-200 dark:border-emerald-900">
						<div class="text-2xl font-extrabold text-emerald-700 dark:text-emerald-400">{metrics.active_users}</div>
						<div class="text-[11px] font-semibold mt-1">Активные</div>
					</div>
					<div class="p-4 rounded-xl bg-amber-50 text-amber-900 border border-amber-200 dark:bg-amber-950 dark:text-amber-200 dark:border-amber-900">
						<div class="text-2xl font-extrabold text-amber-700 dark:text-amber-400">{metrics.blocked_users}</div>
						<div class="text-[11px] font-semibold mt-1">Под ограничениями</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
