<!-- src/lib/components/host-listing-edit/ListingSaveChoiceModal.svelte -->
<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { X, FileText, Send, ShieldCheck, ArrowRight, Loader2 } from 'lucide-svelte';

	interface Props {
		open: boolean;
		isSaving: boolean;
		onClose: () => void;
		onSaveAsDraft: () => void;
		onSubmitForModeration: () => void;
	}

	let { open, isSaving, onClose, onSaveAsDraft, onSubmitForModeration }: Props = $props();

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open && !isSaving) {
			onClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<!-- Backdrop -->
	<button
		type="button"
		class="fixed inset-0 z-50 cursor-default border-none bg-black/50 p-0 backdrop-blur-sm transition-opacity"
		in:fade={{ duration: 180 }}
		out:fade={{ duration: 150 }}
		onclick={() => {
			if (!isSaving) onClose();
		}}
		aria-label="Закрыть модальное окно"
	></button>

	<!-- Modal Dialog -->
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 pointer-events-none">
		<div
			class="pointer-events-auto relative w-full max-w-lg overflow-hidden rounded-3xl border border-zinc-200 bg-white p-6 shadow-2xl sm:p-8"
			in:scale={{ start: 0.95, duration: 200 }}
			out:scale={{ start: 0.95, duration: 150 }}
			role="dialog"
			aria-modal="true"
			aria-labelledby="save-modal-title"
		>
			<!-- Close button -->
			<button
				type="button"
				class="absolute top-5 right-5 flex h-9 w-9 items-center justify-center rounded-xl text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700 disabled:opacity-40 cursor-pointer"
				onclick={onClose}
				disabled={isSaving}
				aria-label="Закрыть"
			>
				<X class="h-5 w-5" />
			</button>

			<!-- Header -->
			<div class="mb-6 pr-8">
				<h3 id="save-modal-title" class="text-xl font-bold text-zinc-900 sm:text-2xl">
					Сохранение изменений
				</h3>
				<p class="mt-1.5 text-xs sm:text-sm text-zinc-500">
					Выберите, как вы хотите сохранить отредактированное объявление
				</p>
			</div>

			<!-- Options Grid -->
			<div class="space-y-3.5">
				<!-- Option 1: Оставить в черновиках -->
				<button
					type="button"
					class="group flex w-full items-start gap-4 rounded-2xl border-2 border-zinc-200 p-4 sm:p-5 text-left transition-all hover:border-blue-500 hover:bg-blue-50/40 active:scale-[0.99] disabled:opacity-50 cursor-pointer"
					onclick={onSaveAsDraft}
					disabled={isSaving}
				>
					<div
						class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-blue-100 text-blue-700 transition-colors group-hover:bg-blue-200/70"
					>
						<FileText class="h-5 w-5" />
					</div>
					<div class="min-w-0 flex-1">
						<div class="flex items-center justify-between gap-2">
							<span class="text-sm sm:text-base font-bold text-zinc-900">
								Оставить в черновиках
							</span>
							<ArrowRight class="h-4 w-4 text-zinc-400 opacity-0 transition-opacity group-hover:opacity-100" />
						</div>
						<p class="mt-1 text-xs text-zinc-600 leading-relaxed">
							Изменения сохранятся, но объявление не пойдет на проверку. Вы сможете продолжить редактирование позже.
						</p>
					</div>
				</button>

				<!-- Option 2: Отправить на модерацию -->
				<button
					type="button"
					class="group flex w-full items-start gap-4 rounded-2xl border-2 border-emerald-500/80 bg-emerald-50/30 p-4 sm:p-5 text-left transition-all hover:border-emerald-600 hover:bg-emerald-50/70 active:scale-[0.99] disabled:opacity-50 cursor-pointer shadow-sm"
					onclick={onSubmitForModeration}
					disabled={isSaving}
				>
					<div
						class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-700 transition-colors group-hover:bg-emerald-200/80"
					>
						{#if isSaving}
							<Loader2 class="h-5 w-5 animate-spin" />
						{:else}
							<Send class="h-5 w-5" />
						{/if}
					</div>
					<div class="min-w-0 flex-1">
						<div class="flex items-center justify-between gap-2">
							<div class="flex items-center gap-2">
								<span class="text-sm sm:text-base font-bold text-emerald-950">
									Отправить на модерацию
								</span>
								<span class="rounded-full bg-emerald-100 px-2 py-0.5 text-[10px] font-semibold text-emerald-800">
									Рекомендуется
								</span>
							</div>
							<ArrowRight class="h-4 w-4 text-emerald-600 opacity-0 transition-opacity group-hover:opacity-100" />
						</div>
						<p class="mt-1 text-xs text-emerald-900/80 leading-relaxed">
							Изменения сохранятся и сразу будут переданы модераторам на проверку. После одобрения объект будет опубликован.
						</p>
					</div>
				</button>
			</div>

			<!-- Footer -->
			<div class="mt-6 flex items-center justify-end border-t border-zinc-100 pt-4">
				<button
					type="button"
					class="px-4 py-2 text-xs sm:text-sm font-medium text-zinc-500 hover:text-zinc-800 transition cursor-pointer"
					onclick={onClose}
					disabled={isSaving}
				>
					Вернуться к редактированию
				</button>
			</div>
		</div>
	</div>
{/if}
