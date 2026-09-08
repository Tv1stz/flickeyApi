<script lang="ts">
	import { onMount } from 'svelte';
	import { adminAuditApi } from '$lib/api/admin/audit';
	import type { AuditLogEntry } from '$lib/types/admin';
	import AuditTable from '$lib/components/admin/AuditTable.svelte';
	import { FileText, RotateCw, Loader2 } from 'lucide-svelte';

	let logs = $state<AuditLogEntry[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let isRefreshing = $state(false);

	function loadAudit() {
		isRefreshing = true;
		error = null;
		adminAuditApi
			.getAuditLogs()
			.then((res) => {
				logs = res || [];
			})
			.catch((err) => {
				error = err.message || 'Не удалось загрузить журнал аудита';
			})
			.finally(() => {
				isLoading = false;
				isRefreshing = false;
			});
	}

	onMount(() => {
		loadAudit();
	});
</script>

<svelte:head>
	<title>Журнал аудита — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<div class="flex items-center gap-2.5">
				<h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-foreground">
					Журнал аудита действий
				</h1>
				<span class="flex h-6 min-w-[24px] items-center justify-center rounded-full bg-zinc-900 px-2 text-xs font-bold text-white dark:bg-white dark:text-zinc-900">
					{logs.length}
				</span>
			</div>
			<p class="mt-1 text-sm text-zinc-500 dark:text-muted-foreground">
				Неизменяемый хронологический лог всех действий модераторов: проверки объявлений, санкции и верификации.
			</p>
		</div>

		<button
			type="button"
			disabled={isRefreshing}
			onclick={loadAudit}
			class="inline-flex items-center gap-2 rounded-full border border-zinc-200/80 bg-white px-4 py-2 text-xs font-semibold text-zinc-700 shadow-2xs transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground cursor-pointer"
		>
			<RotateCw class="h-3.5 w-3.5 {isRefreshing ? 'animate-spin text-zinc-900 dark:text-foreground' : ''}" />
			<span>Обновить журнал</span>
		</button>
	</div>

	{#if isLoading}
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card">
			<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка записей аудита...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={loadAudit}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else}
		<AuditTable {logs} />
	{/if}
</div>