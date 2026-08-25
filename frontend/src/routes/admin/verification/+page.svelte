<script lang="ts">
	import { adminVerificationApi } from '$lib/api/admin/verification';
	import type { VerificationRequestItem } from '$lib/types/admin';
	import { formatDate } from '$lib/utils';
	import { ShieldCheck, CheckCircle2, XCircle, AlertCircle, Eye, Loader2 } from 'lucide-svelte';

	let items = $state<VerificationRequestItem[]>([]);
	let selectedStatus = $state('pending');
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	function loadQueue(statusFilter: string) {
		isLoading = true;
		error = null;
		adminVerificationApi
			.getVerificationQueue(statusFilter)
			.then((res) => {
				items = res || [];
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки очереди верификации';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	$effect(() => {
		loadQueue(selectedStatus);
	});

	function handleAction(reqId: string, action: 'approve' | 'reject' | 'request_changes') {
		const note = prompt('Введите комментарий модерации:');
		if (note === null) return;
		const reason = action === 'reject' || action === 'request_changes' ? prompt('Укажите официальную причину:') || 'Несоответствие документов' : undefined;

		adminVerificationApi
			.reviewVerification(reqId, { action, reason, note })
			.then(() => {
				loadQueue(selectedStatus);
			})
			.catch((err) => {
				alert(err.message || 'Ошибка обработки заявки верификации');
			});
	}
</script>

<div class="space-y-6">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-foreground tracking-tight">Верификация партнёров и юридических лиц</h1>
			<p class="text-sm text-muted-foreground">Проверка субъектов хозяйствования (ИП, Самозанятые, Юр. лица, Физ. лица).</p>
		</div>

		<!-- Status Filters -->
		<div class="flex items-center space-x-1 bg-muted/40 p-1 rounded-2xl border border-border overflow-x-auto scrollbar-none">
			{#each [
				{ id: 'pending', label: 'Ожидают' },
				{ id: 'approved', label: 'Одобрены' },
				{ id: 'changes_requested', label: 'Правки' },
				{ id: 'rejected', label: 'Отклонены' }
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
			<p class="text-xs text-muted-foreground">Загрузка заявок верификации...</p>
		</div>
	{:else if error}
		<div class="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-sm">
			{error}
		</div>
	{:else if items.length === 0}
		<div class="p-12 rounded-2xl border border-dashed border-border bg-card text-center space-y-2">
			<CheckCircle2 class="h-10 w-10 text-emerald-500 mx-auto" />
			<h3 class="text-base font-bold text-foreground">Заявок не найдено</h3>
			<p class="text-xs text-muted-foreground max-w-sm mx-auto">В данной категории нет нерассмотренных заявок верификации.</p>
		</div>
	{:else}
		<div class="rounded-2xl border border-border bg-card overflow-hidden shadow-xs">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-border bg-muted/40 text-[11px] font-bold text-muted-foreground uppercase">
							<th class="py-3.5 px-4">Субъект / Юр. наименование</th>
							<th class="py-3.5 px-4">Тип субъекта</th>
							<th class="py-3.5 px-4">УНП / Налоговый №</th>
							<th class="py-3.5 px-4">Статус</th>
							<th class="py-3.5 px-4">Дата подачи</th>
							<th class="py-3.5 px-4 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-border text-xs">
						{#each items as item}
							<tr class="hover:bg-muted/20 transition-colors">
								<td class="py-3.5 px-4 font-bold text-foreground">
									<a href="/admin/verification/{item.id}" class="hover:text-primary transition-colors">
										{item.legal_name}
									</a>
								</td>
								<td class="py-3.5 px-4 font-medium uppercase text-muted-foreground">
									{item.provider_type}
								</td>
								<td class="py-3.5 px-4 font-mono">
									{item.unp || '—'}
								</td>
								<td class="py-3.5 px-4">
									<span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-primary/10 text-primary border border-primary/20">
										{item.status}
									</span>
								</td>
								<td class="py-3.5 px-4 text-muted-foreground whitespace-nowrap">
									{formatDate(item.created_at)}
								</td>
								<td class="py-3.5 px-4 text-right">
									<div class="flex items-center justify-end gap-1.5">
										<a
											href="/admin/verification/{item.id}"
											class="p-2 rounded-xl border border-border hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
											title="Детальная проверка документов"
										>
											<Eye class="h-4 w-4" />
										</a>
										<button
											type="button"
											onclick={() => handleAction(item.id, 'approve')}
											class="p-2 rounded-xl bg-emerald-100 text-emerald-800 hover:bg-emerald-200 dark:bg-emerald-950 dark:text-emerald-300 transition-colors"
											title="Одобрить"
										>
											<CheckCircle2 class="h-4 w-4" />
										</button>
										<button
											type="button"
											onclick={() => handleAction(item.id, 'request_changes')}
											class="p-2 rounded-xl bg-amber-100 text-amber-800 hover:bg-amber-200 dark:bg-amber-950 dark:text-amber-300 transition-colors"
											title="Запросить исправления"
										>
											<AlertCircle class="h-4 w-4" />
										</button>
										<button
											type="button"
											onclick={() => handleAction(item.id, 'reject')}
											class="p-2 rounded-xl bg-rose-100 text-rose-800 hover:bg-rose-200 dark:bg-rose-950 dark:text-rose-300 transition-colors"
											title="Отклонить"
										>
											<XCircle class="h-4 w-4" />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
