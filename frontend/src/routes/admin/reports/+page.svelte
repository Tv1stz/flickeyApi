<script lang="ts">
	import { adminReportsApi } from '$lib/api/admin/reports';
	import type { ReportItem } from '$lib/types/admin';
	import { formatDate } from '$lib/utils';
	import { AlertTriangle, CheckCircle2, XCircle, Loader2 } from 'lucide-svelte';

	let reports = $state<ReportItem[]>([]);
	let selectedStatus = $state('open');
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	function loadReports(statusFilter: string) {
		isLoading = true;
		error = null;
		adminReportsApi
			.getReports(statusFilter)
			.then((res) => {
				reports = res || [];
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки реестра жалоб';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	$effect(() => {
		loadReports(selectedStatus);
	});

	function handleResolve(reportId: string, action: 'resolve' | 'dismiss') {
		const reason = prompt(action === 'resolve' ? 'Укажите причину удовлетворения жалобы:' : 'Укажите причину отклонения жалобы:') || '';
		if (!reason) return;
		const note = prompt('Дополнительное примечание модератора:') || '';

		adminReportsApi
			.resolveReport(reportId, { action, reason, note })
			.then(() => {
				loadReports(selectedStatus);
			})
			.catch((err) => {
				alert(err.message || 'Ошибка вынесения решения по жалобе');
			});
	}
</script>

<div class="space-y-6">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-foreground tracking-tight">Реестр жалоб и конфликтных ситуаций</h1>
			<p class="text-sm text-muted-foreground">Обработка обращений пользователей на нарушения правил, недобросовестность и контент.</p>
		</div>

		<div class="flex items-center space-x-1 bg-muted/40 p-1 rounded-2xl border border-border overflow-x-auto scrollbar-none">
			{#each [
				{ id: 'open', label: 'Открытые' },
				{ id: 'resolved', label: 'Удовлетворены' },
				{ id: 'dismissed', label: 'Отклонены' }
			] as tab}
				<button
					type="button"
					onclick={() => (selectedStatus = tab.id)}
					class="px-3 py-1.5 rounded-xl text-xs font-semibold whitespace-nowrap transition-colors {selectedStatus === tab.id
						? 'bg-card text-foreground shadow-xs'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					{tab.label}
				</button>
			{/each}
		</div>
	</div>

	{#if isLoading}
		<div class="p-12 text-center">
			<Loader2 class="h-8 w-8 text-primary animate-spin mx-auto mb-2" />
			<p class="text-xs text-muted-foreground">Загрузка реестра жалоб...</p>
		</div>
	{:else if error}
		<div class="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-sm">
			{error}
		</div>
	{:else if reports.length === 0}
		<div class="p-12 rounded-2xl border border-dashed border-border bg-card text-center space-y-2">
			<CheckCircle2 class="h-10 w-10 text-emerald-500 mx-auto" />
			<h3 class="text-base font-bold text-foreground">Жалоб не обнаружено</h3>
			<p class="text-xs text-muted-foreground max-w-sm mx-auto">В данной категории нет обращений, требующих вмешательства модератора.</p>
		</div>
	{:else}
		<div class="rounded-2xl border border-border bg-card overflow-hidden shadow-xs">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-border bg-muted/40 text-[11px] font-bold text-muted-foreground uppercase">
							<th class="py-3.5 px-4">Тип / Объект жалобы</th>
							<th class="py-3.5 px-4">Заявитель</th>
							<th class="py-3.5 px-4">Причина нарушения</th>
							<th class="py-3.5 px-4">Статус</th>
							<th class="py-3.5 px-4">Дата подачи</th>
							<th class="py-3.5 px-4 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-border text-xs">
						{#each reports as r}
							<tr class="hover:bg-muted/20 transition-colors">
								<td class="py-3.5 px-4 font-bold text-foreground">
									<span class="uppercase text-[10px] text-muted-foreground">{r.target_type}:</span> {r.target_id.slice(0, 8)}...
								</td>
								<td class="py-3.5 px-4 font-mono text-muted-foreground">
									{r.reporter_id.slice(0, 8)}...
								</td>
								<td class="py-3.5 px-4 text-foreground max-w-xs">
									<p class="font-semibold">{r.reason}</p>
									{#if r.description}
										<p class="text-[11px] text-muted-foreground truncate">{r.description}</p>
									{/if}
								</td>
								<td class="py-3.5 px-4">
									<span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-rose-100 text-rose-800 border border-rose-200 dark:bg-rose-950 dark:text-rose-300">
										{r.status}
									</span>
								</td>
								<td class="py-3.5 px-4 text-muted-foreground whitespace-nowrap">
									{formatDate(r.created_at)}
								</td>
								<td class="py-3.5 px-4 text-right">
									{#if r.status === 'open'}
										<div class="flex items-center justify-end gap-1.5">
											<button
												type="button"
												onclick={() => handleResolve(r.id, 'resolve')}
												class="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-[11px] transition-colors"
											>
												Удовлетворить
											</button>
											<button
												type="button"
												onclick={() => handleResolve(r.id, 'dismiss')}
												class="px-3 py-1.5 rounded-xl bg-muted hover:bg-muted/80 text-foreground font-semibold text-[11px] transition-colors"
											>
												Отклонить
											</button>
										</div>
									{:else}
										<span class="text-[11px] text-muted-foreground italic">Решение вынесено</span>
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
