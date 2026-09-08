<script lang="ts">
	import { CheckCircle2, XCircle, AlertTriangle, Loader2, X } from 'lucide-svelte';

	interface Props {
		open: boolean;
		action: 'resolve' | 'dismiss';
		reportId: string;
		targetType: string;
		targetId: string;
		reportedReason: string;
		isSubmitting?: boolean;
		onConfirm: (data: { reason: string; note: string }) => void;
		onCancel: () => void;
	}

	let {
		open,
		action,
		reportId,
		targetType,
		targetId,
		reportedReason,
		isSubmitting = false,
		onConfirm,
		onCancel
	}: Props = $props();

	let reasonPreset = $state('');
	let note = $state('');
	let validationError = $state('');

	const resolvePresets: Record<string, string> = {
		violation_confirmed: 'Нарушение подтверждено: приняты меры в отношении контента/пользователя',
		content_removed: 'Контент удален или отправлен на обязательную доработку',
		user_sanctioned: 'К виновной стороне применены административные санкции',
		warning_issued: 'Пользователю вынесено официальное предупреждение',
		other: 'Другое решение (указано в комментарии)'
	};

	const dismissPresets: Record<string, string> = {
		no_violation: 'Нарушений правил сервиса не обнаружено',
		false_report: 'Жалоба необоснованна / ложное обращение',
		insufficient_evidence: 'Недостаточно доказательств или данных для принятия мер',
		duplicate_report: 'Дубликат ранее рассмотренной жалобы',
		other: 'Другая причина отклонения (указана в комментарии)'
	};

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!reasonPreset) {
			validationError = 'Пожалуйста, выберите основание принятого решения.';
			return;
		}
		if (reasonPreset === 'other' && !note.trim()) {
			validationError = 'Для варианта «Другое» укажите подробный комментарий.';
			return;
		}
		validationError = '';

		const finalReason = reasonPreset === 'other'
			? note.trim()
			: ((action === 'resolve' ? resolvePresets[reasonPreset] : dismissPresets[reasonPreset]) || reasonPreset);

		onConfirm({
			reason: finalReason,
			note: note.trim()
		});
	}

	const dialogConfig = $derived.by(() => {
		if (action === 'resolve') {
			return {
				title: 'Удовлетворить жалобу',
				description: `Жалоба будет помечена как решенная. Подтверждается факт нарушения правил сервиса.`,
				badge: 'Удовлетворить',
				badgeClass: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300',
				submitText: 'Подтвердить решение',
				submitClass: 'bg-emerald-600 hover:bg-emerald-700 text-white',
				icon: CheckCircle2,
				iconClass: 'text-emerald-600 bg-emerald-50 dark:bg-emerald-950 dark:text-emerald-400'
			};
		}
		return {
			title: 'Отклонить жалобу',
			description: `Жалоба будет закрыта без применения мер. Заявитель будет уведомлен о результатах рассмотрения.`,
			badge: 'Отклонить жалобу',
			badgeClass: 'bg-zinc-100 text-zinc-800 dark:bg-zinc-800 dark:text-zinc-300',
			submitText: 'Закрыть и отклонить',
			submitClass: 'bg-zinc-900 hover:bg-zinc-800 text-white dark:bg-zinc-100 dark:text-zinc-900',
			icon: XCircle,
			iconClass: 'text-zinc-600 bg-zinc-100 dark:bg-zinc-800 dark:text-zinc-400'
		};
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
						<p class="mt-1 text-xs text-zinc-500 dark:text-muted-foreground">
							Объект: <span class="font-mono font-semibold">{targetType} ({targetId.slice(0, 8)}...)</span> • Причина: {reportedReason}
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

				<div class="space-y-1.5">
					<label for="report-reason-select" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
						Основание вердикта <span class="text-rose-500">*</span>
					</label>
					<select
						id="report-reason-select"
						bind:value={reasonPreset}
						disabled={isSubmitting}
						class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground cursor-pointer"
					>
						<option value="">-- Выберите типовое основание --</option>
						{#each Object.entries(action === 'resolve' ? resolvePresets : dismissPresets) as [key, label]}
							<option value={key}>{label}</option>
						{/each}
					</select>
				</div>

				<div class="space-y-1.5">
					<label for="report-note-input" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
						Комментарий модератора
						{#if reasonPreset === 'other'}
							<span class="text-rose-500">*</span>
						{/if}
					</label>
					<textarea
						id="report-note-input"
						bind:value={note}
						disabled={isSubmitting}
						rows={3}
						placeholder="Укажите подробности принятого решения или действия в отношении нарушителя..."
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