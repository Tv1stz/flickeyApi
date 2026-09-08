<script lang="ts">
	import { CheckCircle, XCircle, AlertCircle, ShieldAlert, Loader2, X } from 'lucide-svelte';
	import type { ModerationAction } from '$lib/types/admin';

	interface Props {
		open: boolean;
		action: ModerationAction;
		listingTitle: string;
		isSubmitting?: boolean;
		onConfirm: (data: { reason: string; note: string }) => void;
		onCancel: () => void;
	}

	let { open, action, listingTitle, isSubmitting = false, onConfirm, onCancel }: Props = $props();

	let reason = $state('');
	let note = $state('');
	let validationError = $state('');

	const reasonTaxonomy: Record<string, string> = {
		incorrect_information: 'Недостоверная информация об объекте',
		prohibited_property: 'Запрещенный тип объекта или локация',
		misleading_description: 'Вводящее в заблуждение описание',
		incorrect_photos: 'Некачественные или несоответствующие фотографии',
		incorrect_price: 'Заниженная / некорректная цена за сутки',
		duplicate_listing: 'Дубликат существующего объявления',
		legal_compliance_issue: 'Нарушение законодательства или отсутствие документов',
		safety_issue: 'Угроза безопасности гостей',
		rules_violation: 'Нарушение правил сервиса Flickey',
		other: 'Другая причина (указана в комментарии)'
	};

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if ((action === 'reject' || action === 'request_changes' || action === 'suspend') && !reason) {
			validationError = 'Пожалуйста, укажите причину принятого решения.';
			return;
		}
		if (action !== 'approve' && reason === 'other' && !note.trim()) {
			validationError = 'Для варианта «Другая причина» заполните подробный комментарий.';
			return;
		}
		validationError = '';
		onConfirm({
			reason: reason ? (reasonTaxonomy[reason] || reason) : 'Соответствует стандартам качества Flickey',
			note: note.trim()
		});
	}

	function getActionConfig(act: ModerationAction) {
		switch (act) {
			case 'approve':
				return {
					title: 'Одобрить и опубликовать',
					description: 'Объявление станет доступным для бронирования в каталоге Flickey.',
					badge: 'Одобрение',
					badgeClass: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300',
					btnText: 'Одобрить публикацию',
					btnClass: 'bg-emerald-600 hover:bg-emerald-700 text-white',
					icon: CheckCircle,
					iconClass: 'text-emerald-600 bg-emerald-50 dark:bg-emerald-950 dark:text-emerald-400'
				};
			case 'reject':
				return {
					title: 'Отклонить объявление',
					description: 'Объявление не будет опубликовано. Хозяин получит уведомление с причиной отклонения.',
					badge: 'Отказ',
					badgeClass: 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300',
					btnText: 'Отклонить объявление',
					btnClass: 'bg-rose-600 hover:bg-rose-700 text-white',
					icon: XCircle,
					iconClass: 'text-rose-600 bg-rose-50 dark:bg-rose-950 dark:text-rose-400'
				};
			case 'request_changes':
				return {
					title: 'Запросить доработку',
					description: 'Объявление вернется автору как черновик с подробными инструкциями по исправлению.',
					badge: 'Запрос правок',
					badgeClass: 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300',
					btnText: 'Отправить замечания',
					btnClass: 'bg-amber-600 hover:bg-amber-700 text-white',
					icon: AlertCircle,
					iconClass: 'text-amber-600 bg-amber-50 dark:bg-amber-950 dark:text-amber-400'
				};
			case 'suspend':
				return {
					title: 'Приостановить публикацию',
					description: 'Объявление временно скрывается из поиска каталога до устранения претензий.',
					badge: 'Приостановка',
					badgeClass: 'bg-zinc-800 text-white dark:bg-zinc-700',
					btnText: 'Приостановить объявление',
					btnClass: 'bg-zinc-900 hover:bg-zinc-800 text-white',
					icon: ShieldAlert,
					iconClass: 'text-zinc-700 bg-zinc-100 dark:bg-zinc-800 dark:text-zinc-300'
				};
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open && !isSubmitting) {
			onCancel();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	{@const cfg = getActionConfig(action)}
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

		<!-- Dialog Panel -->
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
						<p class="mt-1 text-xs text-zinc-500 dark:text-muted-foreground truncate max-w-xs sm:max-w-sm">
							{listingTitle}
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
						<label for="reason-select" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
							Причина решения <span class="text-rose-500">*</span>
						</label>
						<select
							id="reason-select"
							bind:value={reason}
							disabled={isSubmitting}
							class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground cursor-pointer"
						>
							<option value="">-- Выберите основание из классификатора --</option>
							{#each Object.entries(reasonTaxonomy) as [key, label]}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="space-y-1.5">
					<label for="moderation-note" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
						{action === 'approve' ? 'Служебная заметка (необязательно)' : 'Комментарий для автора объявления'}
						{#if action !== 'approve' && reason === 'other'}
							<span class="text-rose-500">*</span>
						{/if}
					</label>
					<textarea
						id="moderation-note"
						bind:value={note}
						disabled={isSubmitting}
						rows={3}
						placeholder={action === 'approve' ? 'Внутренняя отметка модератора...' : 'Опишите конкретные замечания, которые автор должен исправить...'}
						class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground placeholder:text-zinc-400 resize-none"
					></textarea>
				</div>

				<!-- Footer -->
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
						class="rounded-xl px-4 py-2 text-xs font-semibold shadow-xs transition-colors flex items-center gap-1.5 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed {cfg.btnClass}"
					>
						{#if isSubmitting}
							<Loader2 class="h-3.5 w-3.5 animate-spin" />
							<span>Обработка...</span>
						{:else}
							<Icon class="h-3.5 w-3.5" />
							<span>{cfg.btnText}</span>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}