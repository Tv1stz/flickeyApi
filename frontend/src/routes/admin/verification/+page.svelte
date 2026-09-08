<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { adminVerificationApi } from '$lib/api/admin/verification';
	import type { VerificationRequestItem } from '$lib/types/admin';
	import VerificationReviewDialog from '$lib/components/admin/VerificationReviewDialog.svelte';
	import AdminActionMenu from '$lib/components/admin/AdminActionMenu.svelte';
	import { toast } from '$lib/stores/toastStore';
	import { formatDate } from '$lib/utils';
	import {
		ShieldCheck,
		CheckCircle2,
		XCircle,
		AlertCircle,
		Eye,
		Loader2,
		Search,
		X,
		ArrowRight,
		FileText
	} from 'lucide-svelte';

	let items = $state<VerificationRequestItem[]>([]);
	let selectedStatus = $state('pending');
	let searchQuery = $state('');
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Dialog state
	let targetItem = $state<VerificationRequestItem | null>(null);
	let dialogAction = $state<'approve' | 'reject' | 'request_changes'>('approve');
	let isDialogOpen = $state(false);
	let isSubmitting = $state(false);

	function loadQueue(statusFilter: string) {
		isLoading = true;
		error = null;
		adminVerificationApi
			.getVerificationQueue(statusFilter)
			.then((res) => {
				items = res || [];
			})
			.catch((err) => {
				error = err.message || 'Не удалось загрузить очередь верификации';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	$effect(() => {
		loadQueue(selectedStatus);
	});

	function openReviewDialog(item: VerificationRequestItem, action: 'approve' | 'reject' | 'request_changes') {
		targetItem = item;
		dialogAction = action;
		isDialogOpen = true;
	}

	function handleConfirmReview(data: { reason?: string; note?: string }) {
		if (!targetItem) return;
		isSubmitting = true;
		const action = dialogAction;
		adminVerificationApi
			.reviewVerification(targetItem.id, {
				action,
				reason: data.reason,
				note: data.note
			})
			.then(() => {
				isDialogOpen = false;
				targetItem = null;
				toast.success(
					action === 'approve'
						? 'Верификация успешно подтверждена'
						: action === 'reject'
							? 'Заявка отклонена'
							: 'Запрос на исправление документов отправлен'
				);
				loadQueue(selectedStatus);
			})
			.catch((err) => {
				toast.error('Ошибка верификации', err.message || 'Не удалось сохранить решение');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	const filteredItems = $derived.by(() => {
		if (!searchQuery.trim()) return items;
		const q = searchQuery.toLowerCase().trim();
		return items.filter((item) => {
			const nameMatch = item.legal_name?.toLowerCase().includes(q);
			const unpMatch = item.unp?.toLowerCase().includes(q);
			const typeMatch = item.provider_type?.toLowerCase().includes(q);
			const idMatch = item.id?.toLowerCase().includes(q);
			return nameMatch || unpMatch || typeMatch || idMatch;
		});
	});

	function navigateToDetail(id: string) {
		goto(`/admin/verification/${id}`);
	}

	function translateProviderType(type: string): string {
		switch (type) {
			case 'individual': return 'Физлицо';
			case 'self_employed': return 'Самозанятый';
			case 'individual_entrepreneur': return 'ИП';
			case 'legal_entity': return 'Юрлицо';
			default: return type;
		}
	}
</script>

<svelte:head>
	<title>Верификация контрагентов — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page Header & Search -->
	<div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
		<div>
			<div class="flex items-center gap-2.5">
				<h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-foreground">
					Верификация контрагентов
				</h1>
				<span class="flex h-6 min-w-[24px] items-center justify-center rounded-full bg-zinc-900 px-2 text-xs font-bold text-white dark:bg-white dark:text-zinc-900">
					{filteredItems.length}
				</span>
			</div>
			<p class="mt-1 text-sm text-zinc-500 dark:text-muted-foreground">
				Проверка регистрационных документов, УНП, статуса самозанятых, ИП и юрлиц в налоговых реестрах.
			</p>
		</div>

		<!-- Search Input -->
		<div class="relative w-full md:w-80">
			<Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Поиск по названию, УНП, типу..."
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

	<!-- Status Filter Tabs -->
	<div class="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-none">
		{#each [
			{ id: 'pending', label: 'На рассмотрении' },
			{ id: 'approved', label: 'Одобренные' },
			{ id: 'changes_requested', label: 'Требуют правок' },
			{ id: 'rejected', label: 'Отклоненные' },
			{ id: '', label: 'Все заявки' }
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
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка очереди верификации...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={() => loadQueue(selectedStatus)}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if filteredItems.length === 0}
		<div class="rounded-3xl border border-dashed border-zinc-200 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card space-y-3">
			<div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400 mx-auto dark:bg-muted">
				<ShieldCheck class="h-7 w-7 text-emerald-500" />
			</div>
			<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">
				{searchQuery ? 'Заявки не найдены' : 'Очередь пуста'}
			</h3>
			<p class="text-xs text-zinc-500 max-w-sm mx-auto dark:text-muted-foreground">
				{searchQuery ? 'Попробуйте изменить поисковый запрос.' : 'В этой категории нет ожидающих рассмотрения заявок.'}
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
		<!-- Table of Verification Requests with Clean Actions -->
		<div class="rounded-3xl border border-zinc-200/80 bg-white overflow-hidden shadow-2xs dark:border-border dark:bg-card">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-zinc-100 bg-zinc-50/70 text-[11px] font-bold text-zinc-500 uppercase tracking-wider dark:border-border dark:bg-muted/40 dark:text-muted-foreground">
							<th class="py-4 px-5">Наименование / Заявитель</th>
							<th class="py-4 px-4">Тип деятельности</th>
							<th class="py-4 px-4">УНП / ИНН</th>
							<th class="py-4 px-4">Статус</th>
							<th class="py-4 px-4">Дата подачи</th>
							<th class="py-4 px-5 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-zinc-100 text-xs dark:divide-border">
						{#each filteredItems as item (item.id)}
							<tr
								onclick={() => navigateToDetail(item.id)}
								class="group cursor-pointer transition-colors hover:bg-zinc-50/80 dark:hover:bg-muted/30"
							>
								<!-- Legal Name -->
								<td class="py-4 px-5">
									<div class="flex items-center gap-3">
										<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
											<ShieldCheck class="h-4 w-4" />
										</div>
										<div class="min-w-0 max-w-xs">
											<div class="font-bold text-zinc-900 group-hover:text-zinc-600 transition-colors truncate dark:text-foreground">
												{item.legal_name}
											</div>
											<div class="text-[11px] text-zinc-400 font-mono mt-0.5">
												ID: {item.id.slice(0, 8)}
											</div>
										</div>
									</div>
								</td>

								<!-- Provider Type -->
								<td class="py-4 px-4 font-medium text-zinc-700 dark:text-zinc-300">
									<span class="inline-flex items-center rounded-full bg-zinc-100 px-2.5 py-0.5 text-[11px] font-semibold text-zinc-800 dark:bg-muted dark:text-foreground">
										{translateProviderType(item.provider_type)}
									</span>
								</td>

								<!-- UNP -->
								<td class="py-4 px-4 font-mono font-bold text-zinc-900 dark:text-foreground">
									{item.unp || '—'}
								</td>

								<!-- Status Badge -->
								<td class="py-4 px-4">
									{#if item.status === 'approved'}
										<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300">
											<CheckCircle2 class="h-3 w-3" /> Одобрено
										</span>
									{:else if item.status === 'pending' || item.status === 'under_review'}
										<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950 dark:text-amber-300">
											<AlertCircle class="h-3 w-3" /> На проверке
										</span>
									{:else if item.status === 'changes_requested'}
										<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950 dark:text-blue-300">
											Правки
										</span>
									{:else if item.status === 'rejected'}
										<span class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-950 dark:text-rose-300">
											<XCircle class="h-3 w-3" /> Отклонено
										</span>
									{:else}
										<span class="rounded-full px-2.5 py-0.5 text-[11px] font-bold bg-zinc-100 text-zinc-700">
											{item.status}
										</span>
									{/if}
								</td>

								<!-- Created Date -->
								<td class="py-4 px-4 text-zinc-500 whitespace-nowrap dark:text-muted-foreground">
									{formatDate(item.created_at)}
								</td>

								<!-- Actions: 1 primary review button + 1 kebab dropdown -->
								<td class="py-4 px-5 text-right" onclick={(e) => e.stopPropagation()}>
									<div class="flex items-center justify-end gap-2">
										{#if item.status === 'pending' || item.status === 'under_review'}
											<a
												href="/admin/verification/{item.id}"
												class="inline-flex items-center gap-1.5 rounded-full bg-zinc-900 px-3.5 py-1.5 text-xs font-semibold text-white shadow-2xs transition hover:bg-zinc-800 active:scale-95 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-100"
											>
												<span>Документы</span>
												<ArrowRight class="h-3 w-3" />
											</a>
										{/if}

										<AdminActionMenu
											items={[
												{
													label: 'Инспекция и документы',
													icon: Eye,
													variant: 'primary',
													onclick: () => navigateToDetail(item.id)
												},
												{
													label: 'Подтвердить верификацию',
													icon: CheckCircle2,
													variant: 'success',
													onclick: () => openReviewDialog(item, 'approve')
												},
												{
													label: 'Запросить доработку',
													icon: AlertCircle,
													variant: 'warning',
													onclick: () => openReviewDialog(item, 'request_changes')
												},
												{
													label: 'Отклонить заявку',
													icon: XCircle,
													variant: 'danger',
													onclick: () => openReviewDialog(item, 'reject')
												},
												{
													label: 'Профиль заявителя',
													icon: Eye,
													divider: true,
													onclick: () => goto(`/admin/users/${item.user_id}`)
												}
											]}
										/>
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

{#if targetItem}
	<VerificationReviewDialog
		open={isDialogOpen}
		action={dialogAction}
		legalName={targetItem.legal_name}
		providerType={translateProviderType(targetItem.provider_type)}
		{isSubmitting}
		onConfirm={handleConfirmReview}
		onCancel={() => (isDialogOpen = false)}
	/>
{/if}