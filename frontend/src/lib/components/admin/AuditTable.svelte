<script lang="ts">
	import { FileText, ShieldCheck, Filter, Search, X, ArrowRight } from 'lucide-svelte';
	import type { AuditLogEntry } from '$lib/types/admin';
	import { formatDate } from '$lib/utils';

	interface Props {
		logs: AuditLogEntry[];
	}

	let { logs = [] }: Props = $props();

	let filterAction = $state('');
	let filterTarget = $state('');

	let filteredLogs = $derived(
		logs.filter((log) => {
			if (filterAction) {
				const q = filterAction.toLowerCase();
				const matchAct = log.action?.toLowerCase().includes(q);
				const matchReason = log.reason?.toLowerCase().includes(q);
				const matchNote = log.note?.toLowerCase().includes(q);
				const matchTarget = log.target_id?.toLowerCase().includes(q);
				const matchAdmin = log.admin_name?.toLowerCase().includes(q);
				if (!matchAct && !matchReason && !matchNote && !matchTarget && !matchAdmin) return false;
			}
			if (filterTarget && log.target_type !== filterTarget) return false;
			return true;
		})
	);

	function formatActionLabel(act: string): string {
		switch (act) {
			case 'listing_approve':
				return 'Одобрение жилья';
			case 'listing_reject':
				return 'Отклонение жилья';
			case 'listing_request_changes':
				return 'Запрос правок по жилью';
			case 'listing_suspend':
				return 'Приостановка публикации';
			case 'verification_approve':
				return 'Верификация подтверждена';
			case 'verification_reject':
				return 'Отказ в верификации';
			case 'verification_request_changes':
				return 'Запрос доработки документов';
			case 'user_warning':
				return 'Предупреждение пользователю';
			case 'user_temporary_restriction':
				return 'Ограничение аккаунта';
			case 'user_temporary_block':
				return 'Временная блокировка';
			case 'user_permanent_block':
				return 'Перманентный бан';
			case 'user_unblock':
				return 'Снятие ограничений';
			case 'report_resolve':
				return 'Удовлетворение жалобы';
			case 'report_dismiss':
				return 'Отклонение жалобы';
			default:
				return act;
		}
	}

	function getActionBadgeStyle(act: string): string {
		if (act.includes('approve') || act.includes('unblock') || act.includes('resolve')) {
			return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300';
		}
		if (act.includes('reject') || act.includes('block') || act.includes('ban')) {
			return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950 dark:text-rose-300';
		}
		if (act.includes('request') || act.includes('warning') || act.includes('restriction') || act.includes('suspend')) {
			return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950 dark:text-amber-300';
		}
		return 'bg-zinc-100 text-zinc-700 border-zinc-200 dark:bg-muted dark:text-muted-foreground';
	}
</script>

<div class="space-y-4">
	<!-- Filter Toolbar -->
	<div class="rounded-3xl border border-zinc-200/80 bg-white p-4 sm:p-5 shadow-2xs dark:border-border dark:bg-card flex flex-col sm:flex-row sm:items-center justify-between gap-3">
		<div class="flex items-center gap-2 text-xs font-bold text-zinc-700 dark:text-foreground">
			<Filter class="h-4 w-4 text-zinc-400 shrink-0" />
			<span>Фильтры журнала:</span>
		</div>

		<div class="flex flex-wrap items-center gap-2.5">
			<div class="relative min-w-[200px]">
				<Search class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-400" />
				<input
					type="text"
					bind:value={filterAction}
					placeholder="Поиск по действию, причине, ID..."
					class="w-full rounded-full border border-zinc-200 bg-white py-1.5 pl-8 pr-8 text-xs text-zinc-900 placeholder:text-zinc-400 focus:border-zinc-900 focus:outline-none dark:border-input dark:bg-background dark:text-foreground"
				/>
				{#if filterAction}
					<button
						type="button"
						onclick={() => (filterAction = '')}
						class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-600"
					>
						<X class="h-3 w-3" />
					</button>
				{/if}
			</div>

			<select
				bind:value={filterTarget}
				class="rounded-full border border-zinc-200 bg-white px-3 py-1.5 text-xs text-zinc-800 focus:border-zinc-900 focus:outline-none dark:border-input dark:bg-background dark:text-foreground cursor-pointer"
			>
				<option value="">Все объекты</option>
				<option value="listing">Жилье (Listing)</option>
				<option value="user">Пользователи (User)</option>
				<option value="verification">Верификация (Verification)</option>
				<option value="report">Жалобы (Report)</option>
			</select>
		</div>
	</div>

	<!-- Audit Table -->
	<div class="rounded-3xl border border-zinc-200/80 bg-white overflow-hidden shadow-2xs dark:border-border dark:bg-card">
		<div class="overflow-x-auto">
			<table class="w-full text-left border-collapse">
				<thead>
					<tr class="border-b border-zinc-100 bg-zinc-50/70 text-[11px] font-bold text-zinc-500 uppercase tracking-wider dark:border-border dark:bg-muted/40 dark:text-muted-foreground">
						<th class="py-4 px-5">Время</th>
						<th class="py-4 px-4">Администратор</th>
						<th class="py-4 px-4">Действие</th>
						<th class="py-4 px-4">Цель</th>
						<th class="py-4 px-4">Основание / Примечание</th>
						<th class="py-4 px-5">Изменение статуса</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-zinc-100 text-xs dark:divide-border">
					{#if filteredLogs.length === 0}
						<tr>
							<td colspan={6} class="py-12 text-center text-zinc-400 dark:text-muted-foreground">
								Записи аудита не найдены.
							</td>
						</tr>
					{:else}
						{#each filteredLogs as item}
							<tr class="transition-colors hover:bg-zinc-50/80 dark:hover:bg-muted/30">
								<!-- Timestamp -->
								<td class="py-4 px-5 font-mono text-[11px] text-zinc-500 whitespace-nowrap dark:text-muted-foreground">
									{formatDate(item.created_at)}
								</td>

								<!-- Admin Info -->
								<td class="py-4 px-4 font-semibold text-zinc-900 whitespace-nowrap dark:text-foreground">
									{item.admin_name || item.admin_id.slice(0, 8)}
								</td>

								<!-- Action Label -->
								<td class="py-4 px-4">
									<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold border {getActionBadgeStyle(item.action)}">
										{formatActionLabel(item.action)}
									</span>
								</td>

								<!-- Target Reference -->
								<td class="py-4 px-4 font-mono text-[11px] text-zinc-600 whitespace-nowrap dark:text-muted-foreground">
									<span class="font-bold uppercase text-[10px] text-zinc-400">{item.target_type}:</span> {item.target_id.slice(0, 8)}...
								</td>

								<!-- Reason -->
								<td class="py-4 px-4 text-zinc-700 max-w-xs truncate dark:text-zinc-300">
									{item.reason || item.note || '—'}
								</td>

								<!-- Status Transition -->
								<td class="py-4 px-5 font-mono text-[11px] whitespace-nowrap">
									{#if item.old_status || item.new_status}
										<div class="inline-flex items-center gap-1.5 rounded-full bg-zinc-100 px-2.5 py-0.5 text-zinc-700 dark:bg-muted dark:text-foreground">
											<span class="text-zinc-500">{item.old_status || 'none'}</span>
											<ArrowRight class="h-3 w-3 text-zinc-400" />
											<span class="font-bold text-zinc-900 dark:text-foreground">{item.new_status || 'none'}</span>
										</div>
									{:else}
										<span class="text-zinc-400">—</span>
									{/if}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>