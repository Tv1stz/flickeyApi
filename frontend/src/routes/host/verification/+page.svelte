<!-- src/routes/host/verification/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		ShieldCheck,
		Clock,
		AlertCircle,
		CheckCircle2,
		XCircle,
		Building2,
		User,
		ArrowLeft,
		Loader2
	} from 'lucide-svelte';
	import PageShell from '$lib/components/ui/page/PageShell.svelte';
	import PageHeader from '$lib/components/ui/page/PageHeader.svelte';
	import PartnerRequisitesForm from '$lib/components/verification/PartnerRequisitesForm.svelte';
	import { verificationApi } from '$lib/api/verification';
	import type { VerificationResponse } from '$lib/types/verification';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { toast } from '$lib/stores/toastStore';

	let verification = $state<VerificationResponse | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	async function loadVerification() {
		isLoading = true;
		error = null;
		try {
			verification = await verificationApi.getMyVerification();
		} catch (err: any) {
			console.error('Ошибка загрузки верификации:', err);
			error = err?.message || 'Не удалось загрузить статус верификации';
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		if (authStore.initialized && !authStore.user) {
			goto(resolve('/auth'));
			return;
		}
		loadVerification();
	});

	function handleSuccess(res: VerificationResponse) {
		verification = res;
		toast.success('Заявка отправлена', 'Документы успешно переданы на рассмотрение.');
	}
</script>

<svelte:head>
	<title>Верификация партнера (РБ) — Flickey</title>
</svelte:head>

<PageShell>
	<div class="mb-6 flex items-center justify-between">
		<button
			type="button"
			onclick={() => goto(resolve('/host/listings'))}
			class="inline-flex items-center gap-1.5 text-xs font-semibold text-zinc-500 hover:text-zinc-900 transition dark:text-muted-foreground dark:hover:text-foreground cursor-pointer"
		>
			<ArrowLeft class="h-4 w-4" />
			<span>Назад к объявлениям</span>
		</button>
	</div>

	<div class="max-w-3xl mx-auto space-y-6">
		<div class="space-y-1">
			<h1 class="text-xl sm:text-2xl font-black tracking-tight text-zinc-900 dark:text-foreground">
				Верификация партнера (Беларусь)
			</h1>
			<p class="text-xs sm:text-sm text-zinc-500 dark:text-muted-foreground leading-relaxed">
				В соответствии с законодательством Республики Беларусь для сдачи жилья в аренду необходимо предоставить реквизиты юридического лица, ИП или самозанятого (плательщика налога на проф. доход).
			</p>
		</div>

		{#if isLoading}
			<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-xs dark:border-border dark:bg-card">
				<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
				<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка статуса верификации...</p>
			</div>
		{:else if error}
			<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-xs text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
				<span>{error}</span>
				<button
					type="button"
					onclick={loadVerification}
					class="font-semibold underline hover:no-underline cursor-pointer"
				>
					Повторить
				</button>
			</div>
		{:else if verification?.status === 'approved'}
			<!-- APPROVED STATUS BANNER -->
			<div class="rounded-3xl border border-emerald-200 bg-emerald-50/70 p-6 sm:p-8 dark:border-emerald-900/60 dark:bg-emerald-950/40 space-y-4">
				<div class="flex items-start gap-4">
					<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-emerald-600 text-white shadow-xs">
						<CheckCircle2 class="h-6 w-6" />
					</div>
					<div>
						<h2 class="text-base font-bold text-emerald-900 dark:text-emerald-200">
							Верификация партнера подтверждена
						</h2>
						<p class="text-xs text-emerald-700 dark:text-emerald-300 mt-1 leading-relaxed">
							Ваши реквизиты проверены и одобрены администрацией. Все созданные вами объявления после загрузки подтверждающего видео автоматически отправляются на модерацию.
						</p>
					</div>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-4 border-t border-emerald-200/60 dark:border-emerald-900/40 text-xs">
					<div>
						<span class="text-emerald-800/70 dark:text-emerald-400">Наименование:</span>
						<p class="font-bold text-emerald-950 dark:text-emerald-100 mt-0.5">{verification.legal_name}</p>
					</div>
					<div>
						<span class="text-emerald-800/70 dark:text-emerald-400">УНП:</span>
						<p class="font-mono font-bold text-emerald-950 dark:text-emerald-100 mt-0.5">{verification.unp}</p>
					</div>
				</div>
			</div>

		{:else if verification?.status === 'pending'}
			<!-- PENDING STATUS BANNER -->
			<div class="rounded-3xl border border-amber-200 bg-amber-50/70 p-6 sm:p-8 dark:border-amber-900/60 dark:bg-amber-950/40 space-y-4">
				<div class="flex items-start gap-4">
					<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-amber-500 text-white shadow-xs">
						<Clock class="h-6 w-6" />
					</div>
					<div>
						<h2 class="text-base font-bold text-amber-900 dark:text-amber-200">
							Документы находятся на проверке
						</h2>
						<p class="text-xs text-amber-700 dark:text-amber-300 mt-1 leading-relaxed">
							Администрация сервиса проверяет предоставленные вами данные и скан-копии. Обычно рассмотрение занимает от 1 до 24 часов. Как только заявка будет одобрена, ваши объявления с загруженным видео объекта перейдут на публикацию.
						</p>
					</div>
				</div>

				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-4 border-t border-amber-200/60 dark:border-amber-900/40 text-xs">
					<div>
						<span class="text-amber-800/70 dark:text-amber-400">Наименование заявителя:</span>
						<p class="font-bold text-amber-950 dark:text-amber-100 mt-0.5">{verification.legal_name}</p>
					</div>
					<div>
						<span class="text-amber-800/70 dark:text-amber-400">УНП:</span>
						<p class="font-mono font-bold text-amber-950 dark:text-amber-100 mt-0.5">{verification.unp}</p>
					</div>
				</div>
			</div>

		{:else}
			<!-- CHANGES REQUESTED OR REJECTED BANNER -->
			{#if verification?.status === 'changes_requested'}
				<div class="rounded-3xl border border-amber-300 bg-amber-50 p-5 text-xs text-amber-900 dark:border-amber-900 dark:bg-amber-950/70 dark:text-amber-200 flex items-start gap-3">
					<AlertCircle class="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
					<div>
						<p class="font-bold">Модератор запросил исправление данных:</p>
						<p class="mt-1">{verification.rejection_reason || 'Пожалуйста, проверьте правильность введенных реквизитов и прикрепленных документов.'}</p>
					</div>
				</div>
			{:else if verification?.status === 'rejected'}
				<div class="rounded-3xl border border-rose-300 bg-rose-50 p-5 text-xs text-rose-900 dark:border-rose-900 dark:bg-rose-950/70 dark:text-rose-200 flex items-start gap-3">
					<XCircle class="h-5 w-5 text-rose-600 shrink-0 mt-0.5" />
					<div>
						<p class="font-bold">Предыдущая заявка была отклонена:</p>
						<p class="mt-1">{verification.rejection_reason || 'Пожалуйста, предоставьте корректные реквизиты и документы.'}</p>
					</div>
				</div>
			{/if}

			<!-- REQUISITES FORM -->
			<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-8 shadow-xs dark:border-border dark:bg-card">
				<PartnerRequisitesForm
					initialData={verification}
					onSuccess={handleSuccess}
				/>
			</div>
		{/if}
	</div>
</PageShell>
