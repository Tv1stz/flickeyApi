<script lang="ts">
	import { adminAuditApi } from '$lib/api/admin/audit';
	import type { AuditLogEntry } from '$lib/types/admin';
	import AuditTable from '$lib/components/admin/AuditTable.svelte';
	import { FileText, ShieldCheck, Loader2 } from 'lucide-svelte';

	let logs = $state<AuditLogEntry[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	$effect(() => {
		isLoading = true;
		error = null;
		adminAuditApi
			.getAuditLogs()
			.then((res) => {
				logs = res || [];
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки реестра аудита';
			})
			.finally(() => {
				isLoading = false;
			});
	});
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-foreground tracking-tight">Неизменяемый журнал аудита административных действий</h1>
		<p class="text-sm text-muted-foreground">Подробная хронология всех административных решений, модерационных действий и применения мер воздействия.</p>
	</div>

	{#if isLoading}
		<div class="p-12 text-center">
			<Loader2 class="h-8 w-8 text-primary animate-spin mx-auto mb-2" />
			<p class="text-xs text-muted-foreground">Загрузка журнала аудита...</p>
		</div>
	{:else if error}
		<div class="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-sm">
			{error}
		</div>
	{:else}
		<AuditTable {logs} />
	{/if}
</div>
