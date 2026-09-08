<script lang="ts">
	import { CheckCircle2, XCircle, AlertCircle, Loader2, X } from 'lucide-svelte';

	interface Props {
		open: boolean;
		action: 'approve' | 'reject' | 'request_changes';
		legalName: string;
		providerType?: string;
		isSubmitting?: boolean;
		onConfirm: (data: { reason?: string; note?: string }) => void;
		onCancel: () => void;
	}

	let {
		open,
		action,
		legalName,
		providerType = '',
		isSubmitting = false,
		onConfirm,
		onCancel
	}: Props = $props();

	let reasonPreset = $state('');
	let note = $state('');
	let validationError = $state('');

	const rejectionPresets: Record<string, string> = {
		unp_mismatch: 'Несоответствие данных в государственном реестре (ЕГР/УНП)',
		invalid_documents: 'Нечитаемые или неполные сканы документов',
		missing_authorization: 'Отсутствуют документы, подтверждающие полномочия представителя',
		invalid_activity_type: 'Вид деятельности не соответствует правилам краткосрочной аренды',
		suspicious_data: 'Обнаружены недостоверные или сфальсифицированные сведения',
		other: 'Другая причина (указана в комментарии)'
	};

	const requestChangesPresets: Record<string, string> = {
		resubmit_quality: 'Пожалуйста, загрузите скан документа в более высоком разрешении',
		additional_doc: 'Требуется предоставить свидетельство о государственной регистрации',
		clarify_legal_name: 'Уточните полное юридическое наименование организации/ИП',
		attorney_proof: 'Требуется прикрепить доверенность на право заключения договоров',
		other: 'Другие изменения (указаны в комментарии)'
	};

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (action !== 'approve' && !reasonPreset) {
			validationError = 'Пожалуйста, выберите причину решения из списка.';
			return;
		}
		if (action !== 'approve' && reasonPreset === 'other' && !note.trim()) {
			validationError = 'Для варианта «Другая причина» обязательно укажите подробный комментарий.';
			return;
		}
		validationError = '';

		const finalReason = action === 'approve'
			? undefined
			: (reasonPreset === 'other' ? note.trim() : (rejectionPresets[reasonPreset] || requestChangesPresets[reasonPreset] || reasonPreset));

		onConfirm({
			reason: finalReason,
			note: note.trim() || undefined
		});
	}

	const dialogConfig = $derived.by(() => {
		switch (action) {
			case 'approve':
				return {
					title: 'Подтвердить верификацию',
					description: `Вы подтверждаете юридические данные контрагента ${legalName}. Ему будет присвоен статус проверенного партнера.`,
					badge: 'Одобрение',
					badgeClass: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300',
					submitText: 'Подтвердить и выдать статус',
					submitClass: 'bg-emerald-600 hover:bg-emerald-700 text-white',
					icon: CheckCircle2,
					iconClass: 'text-emerald-600 bg-emerald-50 dark:bg-emerald-950 dark:text-emerald-400'
				};
			case 'request_changes':
				return {
					title: 'Запросить доработку документов',
					description: `Отправить заявителю ${legalName} уведомление о необходимости исправить или дополнить предоставленные сведения.`,
					badge: 'Запрос правок',
					badgeClass: 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300',
					submitText: 'Отправить запрос заявителю',
					submitClass: 'bg-amber-600 hover:bg-amber-700 text-white',
					icon: AlertCircle,
					iconClass: 'text-amber-600 bg-amber-50 dark:bg-amber-950 dark:text-amber-400'
				};
			case 'reject':
				return {
					title: 'Отклонить заявку на верификацию',
					description: `Заявка ${legalName} будет отклонена. Заявитель получит официальное уведомление с указанной причиной.`,
					badge: 'Отказ',
					badgeClass: 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300',
					submitText: 'Отклонить заявку',
					submitClass: 'bg-rose-600 hover:bg-rose-700 text-white',
					icon: XCircle,
					iconClass: 'text-rose-600 bg-rose-50 dark:bg-rose-950 dark:text-rose-400'
				};
		}
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open && !isSubmitting) {
			onCancel();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	{@const cfg = dialogConfig}
	{@const Icon = cfg.icon}

	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto"
		role="dialog"
		aria-modal="true"
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="fixed inset-0 bg-black/50 backdrop-blur-xs transition-opacity animate-in fade-in"
			onclick={() => { if (!isSubmitting) onCancel(); }}
		></div>

		<!-- Modal Container -->
		<div
			class="relative z-50 w-full max-w-lg rounded-3xl bg-white p-6 sm:p-7 shadow-2xl border border-zinc-200/80 transition-all animate-in zoom-in-95 my-8 dark:bg-card dark:border-border"
		>
			<!-- Header -->
			<div class="flex items-start justify-between gap-4 pb-4 border-b border-zinc-100 dark:border-border">
				<div class="flex items-start gap-3.5">
					<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl {cfg.iconClass}">
						<Icon class="h-6 w-6" />
					</div>
					<div>
						<div class="flex items-center gap-2">
							<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">{cfg.title}</h3>
							<span class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase {cfg.badgeClass}">
								{cfg.badge}
							</span>
						</div>
						<p class="mt-1 text-xs text-zinc-500 dark:text-muted-foreground line-clamp-2">
							{legalName} {providerType ? `(${providerType})` : ''}
						</p>
					</div>
				</div>

				<button
					type="button"
					disabled={isSubmitting}
					onclick={onCancel}
					class="rounded-full p-1.5 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 transition-colors dark:hover:bg-muted cursor-pointer"
					aria-label="Закрыть"
				>
					<X class="h-4 w-4" />
				</button>
			</div>

			<!-- Form Body -->
			<form onsubmit={handleSubmit} class="mt-5 space-y-4">
				{#if validationError}
					<div class="p-3 text-xs rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950/60 dark:border-rose-900 dark:text-rose-300">
						{validationError}
					</div>
				{/if}

				{#if action !== 'approve'}
					<div class="space-y-1.5">
						<label for="reason-preset-select" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
							Причина решения <span class="text-rose-500">*</span>
						</label>
						<select
							id="reason-preset-select"
							bind:value={reasonPreset}
							disabled={isSubmitting}
							class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground cursor-pointer"
						>
							<option value="">-- Выберите типовое основание --</option>
							{#each Object.entries(action === 'reject' ? rejectionPresets : requestChangesPresets) as [key, label]}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="space-y-1.5">
					<label for="review-note-input" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
						{action === 'approve' ? 'Служебная заметка (необязательно)' : 'Разъяснение для заявителя'}
						{#if action !== 'approve' && reasonPreset === 'other'}
							<span class="text-rose-500">*</span>
						{/if}
					</label>
					<textarea
						id="review-note-input"
						bind:value={note}
						disabled={isSubmitting}
						rows={3}
						placeholder={action === 'approve' ? 'Внутренняя отметка модератора...' : 'Опишите конкретные замечания или отсутствующие документы...'}
						class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground placeholder:text-zinc-400 resize-none"
					></textarea>
				</div>

				<!-- Actions Footer -->
				<div class="pt-3 border-t border-zinc-100 flex items-center justify-end gap-2.5 dark:border-border">
					<button
						type="button"
						disabled={isSubmitting}
						onclick={onCancel}
						class="rounded-xl border border-zinc-200/80 bg-white px-4 py-2 text-xs font-semibold text-zinc-700 hover:bg-zinc-50 transition-colors dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground cursor-pointer"
					>
						Отмена
					</button>

					<button
						type="submit"
						disabled={isSubmitting}
						class="rounded-xl px-4 py-2 text-xs font-semibold shadow-xs transition-colors flex items-center gap-1.5 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed {cfg.submitClass}"
					>
						{#if isSubmitting}
							<Loader2 class="h-3.5 w-3.5 animate-spin" />
							<span>Обработка...</span>
						{:else}
							<Icon class="h-3.5 w-3.5" />
							<span>{cfg.submitText}</span>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}