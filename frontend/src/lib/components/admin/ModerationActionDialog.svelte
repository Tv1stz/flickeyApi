<script lang="ts">
	import { CheckCircle, XCircle, AlertCircle, ShieldAlert, Loader2 } from 'lucide-svelte';
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
		incorrect_information: 'Некорректная информация об объекте',
		prohibited_property: 'Запрещенный тип недвижимого имущества',
		misleading_description: 'Вводящее в заблуждение описание',
		incorrect_photos: 'Некачественные или несоответствующие фото',
		incorrect_price: 'Заниженная / завышенная нереалистичная цена',
		duplicate_listing: 'Дубликат существующего объявления',
		legal_compliance_issue: 'Нарушение законодательства РБ',
		safety_issue: 'Угроза безопасности или санитарным нормам',
		rules_violation: 'Нарушение правил платформы Flickey',
		other: 'Иная причина (см. примечание)'
	};

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if ((action === 'reject' || action === 'request_changes' || action === 'suspend') && !reason) {
			validationError = 'Пожалуйста, выберите официальную причину решения.';
			return;
		}
		if (action !== 'approve' && !note.trim()) {
			validationError = 'Пожалуйста, укажите пояснительное примечание для модерации.';
			return;
		}
		validationError = '';
		onConfirm({ reason, note });
	}

	function getActionConfig(act: ModerationAction) {
		switch (act) {
			case 'approve':
				return { title: 'Одобрить объявление', btnText: 'Подтвердить публикацию', bg: 'bg-emerald-600 hover:bg-emerald-700', icon: CheckCircle };
			case 'reject':
				return { title: 'Отклонить объявление', btnText: 'Подтвердить отклонение', bg: 'bg-rose-600 hover:bg-rose-700', icon: XCircle };
			case 'request_changes':
				return { title: 'Запросить исправления', btnText: 'Отправить замечания', bg: 'bg-amber-600 hover:bg-amber-700', icon: AlertCircle };
			case 'suspend':
				return { title: 'Приостановить показ', btnText: 'Приостановить объявление', bg: 'bg-slate-800 hover:bg-slate-900', icon: ShieldAlert };
		}
	}
</script>

{#if open}
	{@const cfg = getActionConfig(action)}
	{@const Icon = cfg.icon}
	<div class="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
		<div class="bg-card border border-border rounded-2xl shadow-2xl max-w-lg w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
			<div class="p-6 border-b border-border flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="p-2.5 rounded-xl bg-muted text-foreground">
						<Icon class="h-5 w-5" />
					</div>
					<div>
						<h3 class="text-lg font-bold text-foreground">{cfg.title}</h3>
						<p class="text-xs text-muted-foreground truncate max-w-xs">{listingTitle}</p>
					</div>
				</div>
				<button type="button" onclick={onCancel} class="text-muted-foreground hover:text-foreground text-xl font-bold">×</button>
			</div>

			<form onsubmit={handleSubmit} class="p-6 space-y-4">
				{#if validationError}
					<div class="p-3 text-xs rounded-xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 dark:text-rose-200">
						{validationError}
					</div>
				{/if}

				{#if action !== 'approve'}
					<div class="space-y-1.5">
						<label for="reason-select" class="text-xs font-semibold text-foreground">Официальная причина решения <span class="text-rose-500">*</span></label>
						<select
							id="reason-select"
							bind:value={reason}
							class="w-full p-2.5 rounded-xl border border-input bg-background text-sm text-foreground focus:ring-2 focus:ring-primary focus:outline-none"
						>
							<option value="">-- Выберите причину из классификатора --</option>
							{#each Object.entries(reasonTaxonomy) as [key, label]}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="space-y-1.5">
					<label for="note-text" class="text-xs font-semibold text-foreground">
						Примечание модератора {action === 'approve' ? '(опционально)' : '*'}
					</label>
					<textarea
						id="note-text"
						bind:value={note}
						rows={3}
						placeholder={action === 'approve' ? 'Дополнительные комментарии...' : 'Подробное объяснение решения для владельца...'}
						class="w-full p-3 rounded-xl border border-input bg-background text-sm text-foreground focus:ring-2 focus:ring-primary focus:outline-none resize-none"
					></textarea>
				</div>

				<div class="flex items-center justify-end gap-3 pt-3 border-t border-border">
					<button
						type="button"
						onclick={onCancel}
						disabled={isSubmitting}
						class="px-4 py-2.5 rounded-xl border border-border text-sm font-medium hover:bg-muted transition-colors disabled:opacity-50"
					>
						Отмена
					</button>
					<button
						type="submit"
						disabled={isSubmitting}
						class="px-5 py-2.5 rounded-xl text-sm font-semibold text-white transition-all shadow-sm flex items-center gap-2 {cfg.bg} disabled:opacity-50"
					>
						{#if isSubmitting}
							<Loader2 class="h-4 w-4 animate-spin" />
						{/if}
						<span>{cfg.btnText}</span>
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
