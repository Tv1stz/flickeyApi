<script lang="ts">
	import { FileText, ShieldCheck, Filter } from 'lucide-svelte';
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
			if (filterAction && !log.action.toLowerCase().includes(filterAction.toLowerCase())) return false;
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
				return 'Запрос правок жилья';
			case 'listing_suspend':
				return 'Приостановка жилья';
			case 'verification_approve':
				return 'Подтверждение верификации';
			case 'verification_reject':
				return 'Отказ верификации';
			case 'verification_request_changes':
				return 'Запрос правок верификации';
			case 'user_warning':
				return 'Предупреждение пользователю';
			case 'user_temporary_restriction':
				return 'Временное ограничение';
			case 'user_temporary_block':
				return 'Временный бан';
			case 'user_permanent_block':
				return 'Персистентный бан';
			case 'user_unblock':
				return 'Разблокировка';
			case 'report_resolve':
				return 'Решение жалобы';
			case 'report_dismiss':
				return 'Отклонение жалобы';
			default:
				return act;
		}
	}
</script>

<div class="space-y-4">
	<!-- Filters Bar -->
	<div class="p-4 rounded-2xl bg-card border border-border flex flex-wrap items-center justify-between gap-4">
		<div class="flex items-center gap-2">
			<Filter class="h-4 w-4 text-muted-foreground shrink-0" />
			<span class="text-xs font-semibold text-foreground">Фильтрация логов:</span>
		</div>
		<div class="flex flex-wrap items-center gap-3">
			<input
				type="text"
				bind:value={filterAction}
				placeholder="Поиск по действию..."
				class="px-3 py-1.5 rounded-xl border border-input bg-background text-xs text-foreground focus:ring-2 focus:ring-primary focus:outline-none"
			/>
			<select
				bind:value={filterTarget}
				class="px-3 py-1.5 rounded-xl border border-input bg-background text-xs text-foreground focus:ring-2 focus:ring-primary focus:outline-none"
			>
				<option value="">Все объекты</option>
				<option value="listing">Жилье (Listing)</option>
				<option value="user">Пользователь (User)</option>
				<option value="verification">Верификация (Verification)</option>
				<option value="report">Жалоба (Report)</option>
			</select>
		</div>
	</div>

	<!-- Table -->
	<div class="rounded-2xl border border-border bg-card overflow-hidden shadow-xs">
		<div class="overflow-x-auto">
			<table class="w-full text-left border-collapse">
				<thead>
					<tr class="border-b border-border bg-muted/40 text-[11px] font-bold text-muted-foreground uppercase">
						<th class="py-3 px-4">Время</th>
						<th class="py-3 px-4">Администратор</th>
						<th class="py-3 px-4">Действие</th>
						<th class="py-3 px-4">Объект</th>
						<th class="py-3 px-4">Причина / Детали</th>
						<th class="py-3 px-4">Статусы (Старый $\to$ Новый)</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border text-xs">
					{#if filteredLogs.length === 0}
						<tr>
							<td colspan={6} class="py-8 text-center text-muted-foreground">
								Записи аудита не найдены.
							</td>
						</tr>
					{:else}
						{#each filteredLogs as item}
							<tr class="hover:bg-muted/20 transition-colors">
								<td class="py-3 px-4 font-mono text-[11px] text-muted-foreground whitespace-nowrap">
									{formatDate(item.created_at)}
								</td>
								<td class="py-3 px-4 font-medium text-foreground whitespace-nowrap">
									{item.admin_name || item.admin_id.slice(0, 8)}
								</td>
								<td class="py-3 px-4">
									<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-semibold bg-primary/10 text-primary">
										{formatActionLabel(item.action)}
									</span>
								</td>
								<td class="py-3 px-4 font-mono text-[11px] text-muted-foreground whitespace-nowrap">
									{item.target_type}: {item.target_id.slice(0, 8)}...
								</td>
								<td class="py-3 px-4 text-foreground max-w-xs truncate">
									{item.reason || item.note || '—'}
								</td>
								<td class="py-3 px-4 font-mono text-[11px] whitespace-nowrap">
									{#if item.old_status || item.new_status}
										<span class="text-muted-foreground">{item.old_status || 'null'}</span>
										<span class="text-primary font-bold"> → </span>
										<span class="text-foreground font-semibold">{item.new_status || 'null'}</span>
									{:else}
										<span class="text-muted-foreground">—</span>
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
