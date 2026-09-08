<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { adminVerificationApi } from '$lib/api/admin/verification';
	import type { VerificationRequestItem } from '$lib/types/admin';
	import DocumentViewer from '$lib/components/admin/DocumentViewer.svelte';
	import VerificationReviewDialog from '$lib/components/admin/VerificationReviewDialog.svelte';
	import AdminActionMenu from '$lib/components/admin/AdminActionMenu.svelte';
	import { toast } from '$lib/stores/toastStore';
	import { formatDate } from '$lib/utils';
	import {
		ArrowLeft,
		CheckCircle2,
		XCircle,
		AlertCircle,
		ShieldCheck,
		Loader2,
		User,
		Building2,
		ExternalLink
	} from 'lucide-svelte';

	let reqId = $derived(page.params.id);
	let item = $state<VerificationRequestItem | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Dialog state
	let isDialogOpen = $state(false);
	let dialogAction = $state<'approve' | 'reject' | 'request_changes'>('approve');
	let isSubmitting = $state(false);

	function loadDetail(id: string) {
		isLoading = true;
		error = null;
		adminVerificationApi
			.getVerificationDetail(id)
			.then((res) => {
				item = res;
			})
			.catch((err) => {
				error = err.message || 'Не удалось загрузить данные заявки на верификацию';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	$effect(() => {
		if (reqId) {
			loadDetail(reqId);
		}
	});

	function openReview(action: 'approve' | 'reject' | 'request_changes') {
		dialogAction = action;
		isDialogOpen = true;
	}

	function handleConfirmReview(data: { reason?: string; note?: string }) {
		if (!item) return;
		isSubmitting = true;
		const action = dialogAction;
		adminVerificationApi
			.reviewVerification(item.id, {
				action,
				reason: data.reason,
				note: data.note
			})
			.then((updated) => {
				item = updated;
				isDialogOpen = false;
				toast.success(
					action === 'approve'
						? 'Верификация успешно подтверждена'
						: action === 'reject'
							? 'Заявка отклонена'
							: 'Запрос на исправление документов отправлен заявителю'
				);
				goto('/admin/verification');
			})
			.catch((err) => {
				toast.error('Ошибка верификации', err.message || 'Не удалось применить решение');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	function translateProviderType(type?: string): string {
		switch (type) {
			case 'individual': return 'Физическое лицо';
			case 'self_employed': return 'Самозанятый гражданин';
			case 'individual_entrepreneur': return 'Индивидуальный предприниматель (ИП)';
			case 'legal_entity': return 'Юридическое лицо (Организация)';
			default: return type || 'Не указан';
		}
	}

	const allDocuments = $derived.by(() => {
		const list = [...(item?.documents || [])];
		if (item?.requisites) {
			const req = item.requisites;
			if (Array.isArray(req.passport_scan_pages)) {
				req.passport_scan_pages.forEach((url: string, index: number) => {
					if (url && !list.some((d) => d.url === url)) {
						const labels = ['Паспорт: стр. 31', 'Паспорт: стр. 32-33', 'Паспорт: прописка'];
						list.push({
							id: `passport-${index}`,
							name: labels[index] || `Скан паспорта #${index + 1}`,
							url,
							status: 'Загружен'
						});
					}
				});
			}
			if (req.state_reg_cert_scan && typeof req.state_reg_cert_scan === 'string') {
				if (!list.some((d) => d.url === req.state_reg_cert_scan)) {
					list.push({
						id: 'state-reg-cert',
						name: 'Свидетельство ЕГР / Устав',
						url: req.state_reg_cert_scan,
						status: 'Загружен'
					});
				}
			}
		}
		return list;
	});
</script>

<svelte:head>
	<title>{item ? `${item.legal_name} — Верификация` : 'Верификация контрагента'} — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Sticky Action Bar -->
	<div class="sticky top-20 z-30 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 rounded-3xl border border-zinc-200/80 bg-white/95 p-4 sm:px-6 shadow-sm backdrop-blur-md dark:border-border dark:bg-card/95">
		<!-- Left Breadcrumb & Title -->
		<div class="flex items-center gap-3 min-w-0">
			<a
				href="/admin/verification"
				class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-zinc-200/80 bg-white text-zinc-600 transition hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground"
				title="Назад в очередь"
			>
				<ArrowLeft class="h-4 w-4" />
			</a>

			<div class="min-w-0">
				<div class="flex items-center gap-2">
					<span class="text-xs font-semibold text-zinc-500 dark:text-muted-foreground">Верификация</span>
					{#if item}
						<span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold {item.status === 'approved' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : item.status === 'rejected' ? 'bg-rose-50 text-rose-700 border border-rose-200' : 'bg-amber-50 text-amber-700 border border-amber-200'}">
							{item.status === 'approved' ? 'Одобрено' : item.status === 'rejected' ? 'Отклонено' : 'На проверке'}
						</span>
					{/if}
				</div>
				<h1 class="text-sm sm:text-base font-bold text-zinc-900 truncate dark:text-foreground">
					{item?.legal_name || 'Загрузка заявки...'}
				</h1>
			</div>
		</div>

		<!-- Right Action Buttons Hierarchy -->
		{#if item}
			<div class="flex items-center gap-2 shrink-0">
				<button
					type="button"
					onclick={() => openReview('approve')}
					class="inline-flex items-center gap-1.5 rounded-full bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-emerald-700 active:scale-95 cursor-pointer"
				>
					<CheckCircle2 class="h-3.5 w-3.5" />
					<span>Одобрить</span>
				</button>

				<button
					type="button"
					onclick={() => openReview('request_changes')}
					class="inline-flex items-center gap-1.5 rounded-full border border-amber-300 bg-amber-50 px-3.5 py-2 text-xs font-semibold text-amber-800 transition hover:bg-amber-100 active:scale-95 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-300 cursor-pointer"
				>
					<AlertCircle class="h-3.5 w-3.5" />
					<span>Запросить правки</span>
				</button>

				<button
					type="button"
					onclick={() => openReview('reject')}
					class="inline-flex items-center gap-1.5 rounded-full border border-rose-200 bg-rose-50 px-3.5 py-2 text-xs font-semibold text-rose-700 transition hover:bg-rose-100 active:scale-95 dark:border-rose-900 dark:bg-rose-950 dark:text-rose-300 cursor-pointer"
				>
					<XCircle class="h-3.5 w-3.5" />
					<span>Отклонить</span>
				</button>

				<AdminActionMenu
					items={[
						{
							label: 'Профиль заявителя',
							icon: User,
							onclick: () => goto(`/admin/users/${item?.user_id}`)
						}
					]}
				/>
			</div>
		{/if}
	</div>

	<!-- Content Grid -->
	{#if isLoading}
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card">
			<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка данных верификации...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={() => { if (reqId) loadDetail(reqId); }}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if item}
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
			<!-- Legal Summary Card -->
			<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-7 shadow-2xs dark:border-border dark:bg-card space-y-6">
				<div class="flex items-center gap-3.5 pb-4 border-b border-zinc-100 dark:border-border">
					<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
						<ShieldCheck class="h-6 w-6" />
					</div>
					<div class="min-w-0">
						<h2 class="text-base font-bold text-zinc-900 truncate dark:text-foreground">{item.legal_name}</h2>
						<p class="text-xs text-zinc-500 font-medium dark:text-muted-foreground">{translateProviderType(item.provider_type)}</p>
					</div>
				</div>

				<div class="space-y-3 text-xs">
					<div class="flex items-center justify-between py-1 border-b border-zinc-50 dark:border-border/50">
						<span class="text-zinc-500 dark:text-muted-foreground">УНП / ИНН:</span>
						<span class="font-mono font-bold text-zinc-900 dark:text-foreground">{item.unp || '—'}</span>
					</div>

					<div class="flex items-center justify-between py-1 border-b border-zinc-50 dark:border-border/50">
						<span class="text-zinc-500 dark:text-muted-foreground">Статус проверки:</span>
						<span class="font-bold text-zinc-900 uppercase dark:text-foreground">{item.status}</span>
					</div>

					<div class="flex items-center justify-between py-1 border-b border-zinc-50 dark:border-border/50">
						<span class="text-zinc-500 dark:text-muted-foreground">Дата подачи:</span>
						<span class="font-mono text-zinc-700 dark:text-zinc-300">{formatDate(item.created_at)}</span>
					</div>

					<div class="flex items-center justify-between py-1">
						<span class="text-zinc-500 dark:text-muted-foreground">ID заявителя:</span>
						<a href="/admin/users/{item.user_id}" class="font-mono text-zinc-900 hover:underline inline-flex items-center gap-1 dark:text-foreground">
							<span>{item.user_id.slice(0, 8)}...</span>
							<ExternalLink class="h-3 w-3" />
						</a>
					</div>
				</div>

				{#if item.rejection_reason || item.admin_note}
					<div class="rounded-2xl border border-zinc-100 bg-zinc-50/80 p-4 space-y-2 dark:border-border dark:bg-muted/30 text-xs">
						{#if item.rejection_reason}
							<div>
								<span class="font-bold text-rose-700 dark:text-rose-400">Причина отказа:</span>
								<p class="mt-0.5 text-zinc-600 dark:text-muted-foreground">{item.rejection_reason}</p>
							</div>
						{/if}
						{#if item.admin_note}
							<div>
								<span class="font-bold text-zinc-700 dark:text-foreground">Заметка модератора:</span>
								<p class="mt-0.5 text-zinc-600 dark:text-muted-foreground">{item.admin_note}</p>
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Belarus Requisites and Documents Column -->
			<div class="lg:col-span-2 space-y-6">
				{#if item.requisites}
					{@const req = item.requisites}
					<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-7 shadow-2xs dark:border-border dark:bg-card space-y-6">
						<div class="flex items-center justify-between pb-3 border-b border-zinc-100 dark:border-border">
							<div class="flex items-center gap-2.5">
								<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300">
									<Building2 class="h-5 w-5" />
								</div>
								<div>
									<h3 class="text-sm sm:text-base font-bold text-zinc-900 dark:text-foreground">
										Реквизиты партнера (Республика Беларусь)
									</h3>
									<p class="text-xs text-zinc-500 dark:text-muted-foreground">
										{item.provider_type === 'individual' ? 'Физическое лицо (самозанятый / НПД)' : 'Юридическое лицо / ИП'}
									</p>
								</div>
							</div>
							<span class="inline-flex items-center rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-bold text-emerald-700 border border-emerald-200">
								Закон РБ № 99-З
							</span>
						</div>

						<!-- Bank details grid -->
						<div>
							<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 mb-3">
								Банковские реквизиты
							</h4>
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
								<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
									<span class="text-zinc-500 dark:text-muted-foreground">Банк:</span>
									<p class="font-semibold text-zinc-900 mt-0.5 dark:text-foreground">{req.bank_name || '—'}</p>
								</div>
								<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
									<span class="text-zinc-500 dark:text-muted-foreground">БИК банка (BIC):</span>
									<p class="font-mono font-bold text-zinc-900 mt-0.5 dark:text-foreground">{req.bic || '—'}</p>
								</div>
								<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20 sm:col-span-2">
									<span class="text-zinc-500 dark:text-muted-foreground">Расчетный счет (IBAN):</span>
									<p class="font-mono font-bold text-zinc-900 mt-0.5 tracking-wide dark:text-foreground">{req.iban || '—'}</p>
								</div>
								<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
									<span class="text-zinc-500 dark:text-muted-foreground">УНП:</span>
									<p class="font-mono font-bold text-zinc-900 mt-0.5 dark:text-foreground">{req.unp || item.unp || '—'}</p>
								</div>
								<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
									<span class="text-zinc-500 dark:text-muted-foreground">Валюта счета:</span>
									<p class="font-bold text-zinc-900 mt-0.5 dark:text-foreground">{req.currency || 'BYN'}</p>
								</div>
							</div>
						</div>

						<!-- Subject Details: Individual vs Company -->
						{#if item.provider_type === 'individual' || req.personal_id}
							<div>
								<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 mb-3">
									Данные физического лица
								</h4>
								<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
										<span class="text-zinc-500 dark:text-muted-foreground">ФИО:</span>
										<p class="font-semibold text-zinc-900 mt-0.5 dark:text-foreground">
											{req.last_name || ''} {req.first_name || ''}
										</p>
									</div>
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
										<span class="text-zinc-500 dark:text-muted-foreground">Идентификационный номер:</span>
										<p class="font-mono font-bold text-zinc-900 mt-0.5 dark:text-foreground">{req.personal_id || '—'}</p>
									</div>
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
										<span class="text-zinc-500 dark:text-muted-foreground">Дата рождения:</span>
										<p class="font-medium text-zinc-900 mt-0.5 dark:text-foreground">{req.birth_date || '—'}</p>
									</div>
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
										<span class="text-zinc-500 dark:text-muted-foreground">Контакты:</span>
										<p class="text-zinc-900 mt-0.5 dark:text-foreground">{req.phone || '—'} · {req.email || '—'}</p>
									</div>
								</div>
							</div>
						{:else}
							<div>
								<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 mb-3">
									Данные организации / ИП
								</h4>
								<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20 sm:col-span-2">
										<span class="text-zinc-500 dark:text-muted-foreground">Юридическое наименование:</span>
										<p class="font-semibold text-zinc-900 mt-0.5 dark:text-foreground">{req.legal_name || item.legal_name}</p>
									</div>
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
										<span class="text-zinc-500 dark:text-muted-foreground">Бренд / Торговая марка:</span>
										<p class="font-medium text-zinc-900 mt-0.5 dark:text-foreground">{req.brand_name || '—'}</p>
									</div>
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
										<span class="text-zinc-500 dark:text-muted-foreground">Номер в ЕГР:</span>
										<p class="font-mono font-bold text-zinc-900 mt-0.5 dark:text-foreground">{req.egr_number || '—'}</p>
									</div>
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20 sm:col-span-2">
										<span class="text-zinc-500 dark:text-muted-foreground">Юридический адрес:</span>
										<p class="text-zinc-900 mt-0.5 dark:text-foreground">
											{[req.postal_code, req.city, req.legal_address].filter(Boolean).join(', ') || '—'}
										</p>
									</div>
									{#if req.actual_address}
										<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20 sm:col-span-2">
											<span class="text-zinc-500 dark:text-muted-foreground">Фактический адрес:</span>
											<p class="text-zinc-900 mt-0.5 dark:text-foreground">{req.actual_address}</p>
										</div>
									{/if}
									{#if req.website}
										<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
											<span class="text-zinc-500 dark:text-muted-foreground">Сайт:</span>
											<p class="mt-0.5"><a href={req.website} target="_blank" rel="noreferrer" class="text-blue-600 hover:underline">{req.website}</a></p>
										</div>
									{/if}
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-3.5 dark:border-border dark:bg-muted/20">
										<span class="text-zinc-500 dark:text-muted-foreground">Контакты компании:</span>
										<p class="text-zinc-900 mt-0.5 dark:text-foreground">{req.phone || '—'} · {req.email || '—'}</p>
									</div>
								</div>
							</div>

							{#if req.contact_person}
								<div>
									<h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 mb-3">
										Контактное лицо
									</h4>
									<div class="rounded-2xl border border-zinc-100 bg-zinc-50/50 p-4 text-xs dark:border-border dark:bg-muted/20">
										<p class="font-bold text-zinc-900 dark:text-foreground text-sm">{req.contact_person.full_name || '—'}</p>
										<p class="text-zinc-500 mt-0.5 dark:text-muted-foreground">{req.contact_person.position || 'Представитель'}</p>
										<div class="flex items-center gap-4 mt-2 text-zinc-700 dark:text-zinc-300">
											{#if req.contact_person.phone}<span>Тел: {req.contact_person.phone}</span>{/if}
											{#if req.contact_person.email}<span>Email: {req.contact_person.email}</span>{/if}
										</div>
									</div>
								</div>
							{/if}
						{/if}
					</div>
				{/if}

				<!-- Documents Inspection Column -->
				<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-7 shadow-2xs dark:border-border dark:bg-card">
					<DocumentViewer documents={allDocuments} />
				</div>
			</div>
		</div>
	{/if}
</div>

{#if item}
	<VerificationReviewDialog
		open={isDialogOpen}
		action={dialogAction}
		legalName={item.legal_name}
		providerType={translateProviderType(item.provider_type)}
		{isSubmitting}
		onConfirm={handleConfirmReview}
		onCancel={() => (isDialogOpen = false)}
	/>
{/if}