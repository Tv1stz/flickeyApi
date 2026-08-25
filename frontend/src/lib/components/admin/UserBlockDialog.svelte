<script lang="ts">
	import { ShieldAlert, AlertTriangle, UserCheck, Lock, Loader2 } from 'lucide-svelte';
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
		'Жалобы на мошенничество или недобросовестность',
		'Предоставление недостоверных регистрационных данных',
		'Запрещенный контент или фото с нарушениями',
		'Спам, рассылка неприемлемых сообщений',
		'Нарушение правил безопасного проживания и бронирования',
		'Систематическое несоблюдение требований платформы',
		'Попытки обхода комиссии платформы Flickey',
		'Иная причина'
	];

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (enforcementType !== 'unblock' && !reason) {
			validationError = 'Пожалуйста, выберите причину принудительного воздействия.';
			return;
		}
		validationError = '';
		onConfirm({
			enforcement_type: enforcementType,
			reason: reason || 'Административное решение',
			note,
			duration_days: enforcementType === 'temporary_block' || enforcementType === 'temporary_restriction' ? durationDays : undefined
		});
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
		<div class="bg-card border border-border rounded-2xl shadow-2xl max-w-lg w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
			<div class="p-6 border-b border-border flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="p-2.5 rounded-xl bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300">
						<ShieldAlert class="h-5 w-5" />
					</div>
					<div>
						<h3 class="text-lg font-bold text-foreground">Применение дисциплинарных мер</h3>
						<p class="text-xs text-muted-foreground truncate max-w-xs">{userName}</p>
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

				<div class="space-y-1.5">
					<label for="enforce-type" class="text-xs font-semibold text-foreground">Тип дисциплинарного воздействия <span class="text-rose-500">*</span></label>
					<select
						id="enforce-type"
						bind:value={enforcementType}
						class="w-full p-2.5 rounded-xl border border-input bg-background text-sm text-foreground focus:ring-2 focus:ring-primary focus:outline-none"
					>
						<option value="warning">Предупреждение (официальное уведомление)</option>
						<option value="temporary_restriction">Временное ограничение (ограничение создания объявлений)</option>
						<option value="temporary_block">Временная блокировка аккаунта</option>
						<option value="permanent_block">Персистентная блокировка (бессрочный бан)</option>
						<option value="unblock">Снятие всех ограничений (разблокировка)</option>
					</select>
				</div>

				{#if enforcementType === 'temporary_block' || enforcementType === 'temporary_restriction'}
					<div class="space-y-1.5">
						<label for="duration-select" class="text-xs font-semibold text-foreground">Срок действия (в днях)</label>
						<select
							id="duration-select"
							bind:value={durationDays}
							class="w-full p-2.5 rounded-xl border border-input bg-background text-sm text-foreground focus:ring-2 focus:ring-primary focus:outline-none"
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
						<label for="reason-select" class="text-xs font-semibold text-foreground">Причина нарушения <span class="text-rose-500">*</span></label>
						<select
							id="reason-select"
							bind:value={reason}
							class="w-full p-2.5 rounded-xl border border-input bg-background text-sm text-foreground focus:ring-2 focus:ring-primary focus:outline-none"
						>
							<option value="">-- Выберите причину из реестра --</option>
							{#each violationReasons as r}
								<option value={r}>{r}</option>
							{/each}
						</select>
					</div>
				{/if}

				<div class="space-y-1.5">
					<label for="note-text" class="text-xs font-semibold text-foreground">
						Официальное примечание администратора
					</label>
					<textarea
						id="note-text"
						bind:value={note}
						rows={3}
						placeholder="Детали решения, ссылки на обращения или расследование..."
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
						class="px-5 py-2.5 rounded-xl text-sm font-semibold text-white transition-all shadow-sm flex items-center gap-2 disabled:opacity-50 {enforcementType === 'unblock' ? 'bg-emerald-600 hover:bg-emerald-700' : enforcementType === 'permanent_block' ? 'bg-rose-600 hover:bg-rose-700' : 'bg-amber-600 hover:bg-amber-700'}"
					>
						{#if isSubmitting}
							<Loader2 class="h-4 w-4 animate-spin" />
						{/if}
						<span>Подтвердить решение</span>
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
