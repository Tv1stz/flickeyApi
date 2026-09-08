<script lang="ts">
	import { goto } from '$app/navigation';
	import { adminReportsApi } from '$lib/api/admin/reports';
	import type { ReportItem } from '$lib/types/admin';
	import ResolveReportDialog from '$lib/components/admin/ResolveReportDialog.svelte';
	import AdminActionMenu from '$lib/components/admin/AdminActionMenu.svelte';
	import { toast } from '$lib/stores/toastStore';
	import { formatDate } from '$lib/utils';
	import {
		AlertTriangle,
		CheckCircle2,
		XCircle,
		Loader2,
		Search,
		X,
		ExternalLink,
		Building2,
		User
	} from 'lucide-svelte';

	let reports = $state<ReportItem[]>([]);
	let selectedStatus = $state('open');
	let searchQuery = $state('');
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Dialog state
	let targetReport = $state<ReportItem | null>(null);
	let dialogAction = $state<'resolve' | 'dismiss'>('resolve');
	let isDialogOpen = $state(false);
	let isSubmitting = $state(false);

	function loadReports(statusFilter: string) {
		isLoading = true;
		error = null;
		adminReportsApi
			.getReports(statusFilter)
			.then((res) => {
				reports = res || [];
			})
			.catch((err) => {
				error = err.message || 'Не удалось загрузить реестр жалоб';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	$effect(() => {
		loadReports(selectedStatus);
	});

	function openDialog(report: ReportItem, action: 'resolve' | 'dismiss') {
		targetReport = report;
		dialogAction = action;
		isDialogOpen = true;
	}

	function handleConfirmResolve(data: { reason: string; note: string }) {
		if (!targetReport) return;
		isSubmitting = true;
		const action = dialogAction;
		adminReportsApi
			.resolveReport(targetReport.id, { action, reason: data.reason, note: data.note })
			.then(() => {
				isDialogOpen = false;
				targetReport = null;
				toast.success(
					action === 'resolve'
						? 'Жалоба успешно удовлетворена'
						: 'Жалоба отклонена'
				);
				loadReports(selectedStatus);
			})
			.catch((err) => {
				toast.error('Ошибка обработки', err.message || 'Не удалось обработать жалобу');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	const filteredReports = $derived.by(() => {
		if (!searchQuery.trim()) return reports;
		const q = searchQuery.toLowerCase().trim();
		return reports.filter((r) => {
			const reasonMatch = r.reason?.toLowerCase().includes(q);
			const descMatch = r.description?.toLowerCase().includes(q);
			const targetMatch = r.target_id?.toLowerCase().includes(q);
			const reporterMatch = r.reporter_id?.toLowerCase().includes(q);
			return reasonMatch || descMatch || targetMatch || reporterMatch;
		});
	});

	function openTarget(r: ReportItem) {
		if (r.target_type === 'listing') {
			goto(`/admin/listings/${r.target_id}`);
		} else {
			goto(`/admin/users/${r.target_id}`);
		}
	}
</script>

<svelte:head>
	<title>Жалобы и обращения — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header & Search -->
	<div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
		<div>
			<div class="flex items-center gap-2.5">
				<h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-foreground">
					Жалобы и инциденты
				</h1>
				<span class="flex h-6 min-w-[24px] items-center justify-center rounded-full bg-zinc-900 px-2 text-xs font-bold text-white dark:bg-white dark:text-zinc-900">
					{filteredReports.length}
				</span>
			</div>
			<p class="mt-1 text-sm text-zinc-500 dark:text-muted-foreground">
				Рассмотрение обращений гостей и хозяев по нарушениям правил проживания, спаму и мошенничеству.
			</p>
		</div>

		<!-- Search Input -->
		<div class="relative w-full md:w-80">
			<Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Поиск по причине, описанию, ID..."
				class="w-full rounded-full border border-zinc-200/80 bg-white py-2 pl-9 pr-9 text-xs text-zinc-900 placeholder:text-zinc-400 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-border dark:bg-card dark:text-foreground shadow-2xs"
			/>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200"
				>
					<X class="h-3.5 w-3.5" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Status Tabs -->
	<div class="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-none">
		{#each [
			{ id: 'open', label: 'Открытые жалобы' },
			{ id: 'resolved', label: 'Удовлетворенные' },
			{ id: 'dismissed', label: 'Отклоненные' },
			{ id: '', label: 'Все обращения' }
		] as tab}
			{@const active = selectedStatus === tab.id}
			<button
				type="button"
				onclick={() => (selectedStatus = tab.id)}
				class="whitespace-nowrap rounded-full px-4 py-2 text-xs font-semibold transition-all cursor-pointer {active
					? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
					: 'border border-zinc-200/80 bg-white text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900 hover:border-zinc-300 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>

	<!-- Content Area -->
	{#if isLoading}
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card">
			<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка очереди обращений...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={() => loadReports(selectedStatus)}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if filteredReports.length === 0}
		<div class="rounded-3xl border border-dashed border-zinc-200 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card space-y-3">
			<div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400 mx-auto dark:bg-muted">
				<CheckCircle2 class="h-7 w-7 text-emerald-500" />
			</div>
			<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">
				{searchQuery ? 'Жалобы не найдены' : 'Жалоб нет'}
			</h3>
			<p class="text-xs text-zinc-500 max-w-sm mx-auto dark:text-muted-foreground">
				{searchQuery ? 'Попробуйте изменить поисковый запрос.' : 'В данной категории нет нерассмотренных инцидентов.'}
			</p>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="mt-2 inline-flex items-center gap-1.5 rounded-full bg-zinc-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-zinc-800 dark:bg-white dark:text-zinc-900"
				>
					Сбросить поиск
				</button>
			{/if}
		</div>
	{:else}
		<!-- Reports Table -->
		<div class="rounded-3xl border border-zinc-200/80 bg-white overflow-hidden shadow-2xs dark:border-border dark:bg-card">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-zinc-100 bg-zinc-50/70 text-[11px] font-bold text-zinc-500 uppercase tracking-wider dark:border-border dark:bg-muted/40 dark:text-muted-foreground">
							<th class="py-4 px-5">Целевой объект</th>
							<th class="py-4 px-4">Заявитель</th>
							<th class="py-4 px-4">Причина и подробности</th>
							<th class="py-4 px-4">Статус</th>
							<th class="py-4 px-4">Подано</th>
							<th class="py-4 px-5 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-zinc-100 text-xs dark:divide-border">
						{#each filteredReports as r (r.id)}
							<tr class="transition-colors hover:bg-zinc-50/80 dark:hover:bg-muted/30">
								<!-- Target Object -->
								<td class="py-4 px-5 font-bold text-zinc-900 dark:text-foreground">
									<div class="flex items-center gap-2.5">
										{#if r.target_type === 'listing'}
											<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300">
												<Building2 class="h-4 w-4" />
											</div>
										{:else}
											<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-zinc-100 text-zinc-700 dark:bg-muted dark:text-foreground">
												<User class="h-4 w-4" />
											</div>
										{/if}
										<div>
											<button
												type="button"
												onclick={() => openTarget(r)}
												class="font-bold hover:underline inline-flex items-center gap-1 cursor-pointer text-left"
											>
												<span class="capitalize">{r.target_type === 'listing' ? 'Жилье' : 'Пользователь'}:</span>
												<span class="font-mono">{r.target_id.slice(0, 8)}...</span>
												<ExternalLink class="h-3 w-3 text-zinc-400" />
											</button>
										</div>
									</div>
								</td>

								<!-- Reporter -->
								<td class="py-4 px-4 font-mono text-zinc-500 dark:text-muted-foreground">
									<a href="/admin/users/{r.reporter_id}" class="hover:underline">
										{r.reporter_id.slice(0, 8)}...
									</a>
								</td>

								<!-- Reason & Description -->
								<td class="py-4 px-4 max-w-sm">
									<p class="font-bold text-zinc-900 dark:text-foreground">{r.reason}</p>
									{#if r.description}
										<p class="text-[11px] text-zinc-500 mt-0.5 line-clamp-2 dark:text-muted-foreground">{r.description}</p>
									{/if}
								</td>

								<!-- Status -->
								<td class="py-4 px-4">
									{#if r.status === 'open'}
										<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-950 dark:text-rose-300">
											Открыта
										</span>
									{:else if r.status === 'resolved'}
										<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300">
											Удовлетворена
										</span>
									{:else}
										<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-zinc-100 text-zinc-700 border border-zinc-200 dark:bg-muted dark:text-muted-foreground">
											{r.status === 'dismissed' ? 'Отклонена' : r.status}
										</span>
									{/if}
								</td>

								<!-- Created Date -->
								<td class="py-4 px-4 text-zinc-500 whitespace-nowrap dark:text-muted-foreground">
									{formatDate(r.created_at)}
								</td>

								<!-- Actions Menu -->
								<td class="py-4 px-5 text-right">
									{#if r.status === 'open'}
										<div class="flex items-center justify-end gap-1.5">
											<button
												type="button"
												onclick={() => openDialog(r, 'resolve')}
												class="rounded-full bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white shadow-2xs transition hover:bg-emerald-700 active:scale-95 cursor-pointer"
											>
												Решить
											</button>
											<button
												type="button"
												onclick={() => openDialog(r, 'dismiss')}
												class="rounded-full border border-zinc-200/80 bg-white px-3 py-1.5 text-xs font-semibold text-zinc-700 hover:bg-zinc-50 transition active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground cursor-pointer"
											>
												Отклонить
											</button>
										</div>
									{:else}
										<span class="text-[11px] text-zinc-400 italic">Рассмотрено</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

{#if targetReport}
	<ResolveReportDialog
		open={isDialogOpen}
		action={dialogAction}
		reportId={targetReport.id}
		targetType={targetReport.target_type === 'listing' ? 'Жилье' : 'Пользователь'}
		targetId={targetReport.target_id}
		reportedReason={targetReport.reason}
		{isSubmitting}
		onConfirm={handleConfirmResolve}
		onCancel={() => (isDialogOpen = false)}
	/>
{/if}