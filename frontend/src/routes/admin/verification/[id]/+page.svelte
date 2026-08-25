<script lang="ts">
	import { page } from '$app/state';
	import { adminVerificationApi } from '$lib/api/admin/verification';
	import type { VerificationRequestItem } from '$lib/types/admin';
	import DocumentViewer from '$lib/components/admin/DocumentViewer.svelte';
	import { formatDate } from '$lib/utils';
	import { ArrowLeft, CheckCircle2, XCircle, AlertCircle, ShieldCheck, Loader2 } from 'lucide-svelte';

	let reqId = $derived(page.params.id);
	let item = $state<VerificationRequestItem | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	$effect(() => {
		if (!reqId) return;
		isLoading = true;
		error = null;
		adminVerificationApi
			.getVerificationDetail(reqId)
			.then((res) => {
				item = res;
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки заявки верификации';
			})
			.finally(() => {
				isLoading = false;
			});
	});

	function handleAction(action: 'approve' | 'reject' | 'request_changes') {
		if (!item) return;
		const note = prompt('Введите комментарий решения модератора:') || '';
		const reason = action === 'reject' || action === 'request_changes' ? prompt('Укажите официальную причину:') || 'Несоответствие нормам' : undefined;

		adminVerificationApi
			.reviewVerification(item.id, { action, reason, note })
			.then((updated) => {
				item = updated;
			})
			.catch((err) => {
				alert(err.message || 'Ошибка обновления статуса верификации');
			});
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<a
			href="/admin/verification"
			class="inline-flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
		>
			<ArrowLeft class="h-4 w-4" />
			<span>Назад в очередь верификации</span>
		</a>

		{#if item}
			<div class="flex items-center gap-2">
				<button
					type="button"
					onclick={() => handleAction('approve')}
					class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs"
				>
					<CheckCircle2 class="h-4 w-4" />
					<span>Одобрить субъект</span>
				</button>
				<button
					type="button"
					onclick={() => handleAction('request_changes')}
					class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs"
				>
					<AlertCircle class="h-4 w-4" />
					<span>Запросить исправления</span>
				</button>
				<button
					type="button"
					onclick={() => handleAction('reject')}
					class="px-4 py-2 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs"
				>
					<XCircle class="h-4 w-4" />
					<span>Отклонить заявку</span>
				</button>
			</div>
		{/if}
	</div>

	{#if isLoading}
		<div class="p-12 text-center">
			<Loader2 class="h-8 w-8 text-primary animate-spin mx-auto mb-2" />
			<p class="text-xs text-muted-foreground">Загрузка карточки верификации...</p>
		</div>
	{:else if error}
		<div class="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-sm">
			{error}
		</div>
	{:else if item}
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<div class="p-6 rounded-2xl bg-card border border-border space-y-4">
				<div class="flex items-center gap-3">
					<div class="p-3 rounded-2xl bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300">
						<ShieldCheck class="h-6 w-6" />
					</div>
					<div>
						<h2 class="text-base font-bold text-foreground">{item.legal_name}</h2>
						<p class="text-xs text-muted-foreground uppercase">{item.provider_type}</p>
					</div>
				</div>

				<div class="space-y-2 pt-3 border-t border-border text-xs">
					<div class="flex justify-between py-1">
						<span class="text-muted-foreground">УНП / Код:</span>
						<span class="font-mono font-bold text-foreground">{item.unp || '—'}</span>
					</div>
					<div class="flex justify-between py-1">
						<span class="text-muted-foreground">Текущий статус:</span>
						<span class="font-bold text-foreground uppercase">{item.status}</span>
					</div>
					<div class="flex justify-between py-1">
						<span class="text-muted-foreground">Дата подачи:</span>
						<span class="font-mono text-foreground">{formatDate(item.created_at)}</span>
					</div>
				</div>
			</div>

			<div class="lg:col-span-2 p-6 rounded-2xl bg-card border border-border space-y-6">
				<DocumentViewer documents={item.documents || []} />
			</div>
		</div>
	{/if}
</div>
