<script lang="ts">
	import { ShieldAlert, AlertTriangle, UserCheck, Lock, Loader2, X } from 'lucide-svelte';
	import type { EnforcementType } from '$lib/types/admin';

	interface Props {
		open: boolean;
		userName: string;
		isSubmitting?: boolean;
		onConfirm: (data: { enforcement_type: EnforcementType; reason: string; note: string; duration_days?: number }) => void;
		onCancel: () => void;
	}

	let { open, userName, isSubmitting = false, onConfirm, onCancel }: Props = $props();

	let enforcementType = $state<EnforcementType>('warning');
	let reason = $state('');
	let note = $state('');
	let durationDays = $state(7);
	let validationError = $state('');

	const violationReasons = [
		'Жалобы гостей на несоответствие жилья описанию',
		'Предоставление недостоверных персональных данных',
		'Запрещенный контент в описании или фото',
		'Спам, рассылка рекламных сообщений',
		'Попытка проведения расчетов в обход сервиса Flickey',
		'Систематический срыв подтвержденных бронирований',
		'Оскорбительное поведение или угрозы пользователям',
		'Грубое нарушение правил и условий сервиса',
		'Другая причина'
	];

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (enforcementType !== 'unblock' && !reason) {
			validationError = 'Пожалуйста, выберите основание наложения санкций.';
			return;
		}
		if (reason === 'Другая причина' && !note.trim()) {
			validationError = 'Для варианта «Другая причина» заполните подробный комментарий.';
			return;
		}
		validationError = '';
		onConfirm({
			enforcement_type: enforcementType,
			reason: reason || 'Административное решение',
			note: note.trim(),
			duration_days: enforcementType === 'temporary_block' || enforcementType === 'temporary_restriction' ? durationDays : undefined
		});
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open && !isSubmitting) {
			onCancel();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
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
					<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300">
						<ShieldAlert class="h-6 w-6" />
					</div>
					<div>
						<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">Административные меры</h3>
						<p class="mt-1 text-xs text-zinc-500 dark:text-muted-foreground truncate max-w-xs sm:max-w-sm">
							Пользователь: <span class="font-semibold text-zinc-900 dark:text-foreground">{userName}</span>
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
					<label for="enforce-type" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
						Вид административного воздействия <span class="text-rose-500">*</span>
					</label>
					<select
						id="enforce-type"
						bind:value={enforcementType}
						disabled={isSubmitting}
						class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground cursor-pointer"
					>
						<option value="warning">Предупреждение (официальное уведомление)</option>
						<option value="temporary_restriction">Временное ограничение (запрет публикации)</option>
						<option value="temporary_block">Временная блокировка аккаунта</option>
						<option value="permanent_block">Перманентная блокировка (бессрочно)</option>
						<option value="unblock">Снять все ограничения (разблокировать)</option>
					</select>
				</div>

				{#if enforcementType === 'temporary_block' || enforcementType === 'temporary_restriction'}
					<div class="space-y-1.5">
						<label for="duration-select" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
							Срок действия ограничения
						</label>
						<select
							id="duration-select"
							bind:value={durationDays}
							disabled={isSubmitting}
							class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground cursor-pointer"
						>
							<option value={3}>3 дня</option>
							<option value={7}>7 дней (1 неделя)</option>
							<option value={14}>14 дней (2 недели)</option>
							<option value={30}>30 дней (1 месяц)</option>
							<option value={90}>90 дней (3 месяца)</option>
						</select>
					</div>
				{/if}

				{#if enforcementType !== 'unblock'}
					<div class="space-y-1.5">
						<label for="reason-select" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
							Основание для применения мер <span class="text-rose-500">*</span>
						</label>
						<select
							id="reason-select"
							bind:value={reason}
							disabled={isSubmitting}
							class="w-full rounded-xl border border-zinc-200 bg-white p-2.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-input dark:bg-background dark:text-foreground cursor-pointer"
						>
							<option value="">-- Выберите причину из списка --</option>
							{#each violationReasons as r}
								<option value={r}>{r}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="space-y-1.5">
					<label for="enforce-note" class="text-xs font-semibold text-zinc-700 dark:text-foreground">
						Подробный комментарий (причина и внутренние заметки)
						{#if reason === 'Другая причина'}
							<span class="text-rose-500">*</span>
						{/if}
					</label>
					<textarea
						id="enforce-note"
						bind:value={note}
						disabled={isSubmitting}
						rows={3}
						placeholder="Укажите факты, ссылки на жалобы или конкретные нарушения..."
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
						class="rounded-xl px-4 py-2 text-xs font-semibold text-white shadow-xs transition-colors flex items-center gap-1.5 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed {enforcementType === 'unblock' ? 'bg-emerald-600 hover:bg-emerald-700' : enforcementType === 'permanent_block' ? 'bg-rose-600 hover:bg-rose-700' : 'bg-zinc-900 hover:bg-zinc-800'}"
					>
						{#if isSubmitting}
							<Loader2 class="h-3.5 w-3.5 animate-spin" />
							<span>Применение...</span>
						{:else if enforcementType === 'unblock'}
							<UserCheck class="h-3.5 w-3.5" />
							<span>Разблокировать пользователя</span>
						{:else}
							<ShieldAlert class="h-3.5 w-3.5" />
							<span>Применить меры</span>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}