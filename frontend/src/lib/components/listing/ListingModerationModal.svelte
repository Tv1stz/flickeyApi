<!-- src/lib/components/listing/ListingModerationModal.svelte -->
<script lang="ts">
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import { AlertTriangle, XCircle, ShieldAlert, X, Pencil, ArrowRight, CheckCircle2 } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import StatusIndicator from '$lib/components/ui/StatusIndicator.svelte';
	import type { Listing } from '$lib/components/card/types';

	interface Props {
		open: boolean;
		listing: Listing | null;
		onClose: () => void;
		onEdit?: (listing: Listing) => void;
	}

	let { open, listing, onClose, onEdit }: Props = $props();

	let isChangesRequested = $derived(listing?.status === 'changes_requested');
	let isRejected = $derived(listing?.status === 'rejected');
	let isSuspended = $derived(listing?.status === 'suspended');

	let modalTitle = $derived.by(() => {
		if (isChangesRequested) return 'Замечания модератора';
		if (isRejected) return 'Причина отклонения';
		if (isSuspended) return 'Приостановка объявления';
		return 'Решение модерации';
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open) {
			onClose();
		}
	}

	function handleEdit(e?: MouseEvent) {
		e?.preventDefault?.();
		e?.stopPropagation?.();
		if (!listing) return;
		const target = listing;
		onClose();
		if (onEdit) {
			onEdit(target);
		} else {
			goto(`/host/listings/${target.id}`);
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open && listing}
	<!-- Backdrop -->
	<button
		type="button"
		class="fixed inset-0 z-[80] cursor-default border-none bg-black/50 p-0 backdrop-blur-sm transition-opacity"
		in:fade={{ duration: 200 }}
		out:fade={{ duration: 150 }}
		onclick={onClose}
		aria-label="Закрыть модальное окно"
	></button>

	<!-- Modal Dialog Container -->
	<div
		class="pointer-events-none fixed inset-0 z-[81] flex items-end justify-center p-0 sm:items-center sm:p-4"
	>
		<div
			class="pointer-events-auto flex max-h-[90vh] w-full max-w-lg flex-col overflow-hidden rounded-t-3xl bg-white shadow-2xl sm:rounded-3xl"
			in:fly={{ y: 24, duration: 250, easing: cubicOut }}
			out:scale={{ start: 0.97, duration: 180, easing: cubicOut }}
			role="dialog"
			aria-modal="true"
			aria-labelledby="moderation-modal-title"
		>
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-zinc-100 px-6 py-4 sm:px-8">
				<div class="flex items-center gap-3">
					<div
						class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl
						       {isChangesRequested ? 'bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-300' : ''}
						       {isRejected ? 'bg-rose-100 text-rose-700 dark:bg-rose-950 dark:text-rose-300' : ''}
						       {isSuspended ? 'bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300' : ''}"
					>
						{#if isChangesRequested}
							<AlertTriangle class="h-5 w-5" strokeWidth={2} />
						{:else if isRejected}
							<XCircle class="h-5 w-5" strokeWidth={2} />
						{:else}
							<ShieldAlert class="h-5 w-5" strokeWidth={2} />
						{/if}
					</div>
					<div>
						<h3 id="moderation-modal-title" class="text-lg font-bold tracking-tight text-zinc-900">
							{modalTitle}
						</h3>
						<div class="mt-0.5">
							<StatusIndicator status={listing.status} size="xs" />
						</div>
					</div>
				</div>

				<button
					type="button"
					onclick={onClose}
					class="flex h-9 w-9 items-center justify-center rounded-full text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700"
					aria-label="Закрыть"
				>
					<X class="h-5 w-5" strokeWidth={1.75} />
				</button>
			</div>

			<!-- Scrollable Content -->
			<div class="overflow-y-auto p-6 sm:p-8 space-y-6">
				<!-- Listing Target Info -->
				<div class="rounded-2xl border border-zinc-100 bg-zinc-50/70 p-3.5">
					<p class="text-xs font-semibold text-zinc-400 uppercase tracking-wider">Объект</p>
					<p class="mt-0.5 truncate text-sm font-bold text-zinc-900">{listing.title}</p>
					{#if listing.address}
						<p class="truncate text-xs text-zinc-500 mt-0.5">{listing.address}</p>
					{/if}
				</div>

				<!-- Decision Card -->
				<div
					class="rounded-2xl border p-4 sm:p-5
					       {isChangesRequested ? 'border-amber-200 bg-amber-50/60 text-amber-950' : ''}
					       {isRejected ? 'border-rose-200 bg-rose-50/60 text-rose-950' : ''}
					       {isSuspended ? 'border-zinc-200 bg-zinc-50 text-zinc-950' : ''}"
				>
					<div class="flex items-start gap-3">
						<div class="min-w-0 flex-1">
							<span class="text-[11px] font-bold uppercase tracking-wider opacity-70">
								{isChangesRequested ? 'Требуется исправление' : 'Основание решения'}
							</span>
							<h4 class="mt-1 text-base font-bold leading-snug">
								{listing.rejectionReason || (isChangesRequested ? 'Требуется скорректировать данные' : 'Не соответствует требованиям')}
							</h4>

							{#if listing.moderationComment}
								<div class="mt-3 rounded-xl border border-black/5 bg-white/80 p-3.5 backdrop-blur-xs text-xs sm:text-sm leading-relaxed text-zinc-800">
									<p class="text-[11px] font-semibold text-zinc-400 mb-1 uppercase tracking-wider">Комментарий проверки:</p>
									<p class="whitespace-pre-line">{listing.moderationComment}</p>
								</div>
							{/if}
						</div>
					</div>
				</div>

				<!-- Guidance Checklist -->
				<div class="space-y-3">
					<h5 class="text-xs font-bold uppercase tracking-wider text-zinc-400">Что делать дальше</h5>
					<div class="space-y-2.5 text-xs sm:text-sm text-zinc-600">
						<div class="flex items-start gap-2.5">
							<div class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 font-bold text-xs mt-0.5">
								1
							</div>
							<p>Нажмите <strong>«Редактировать объявление»</strong>, чтобы внести необходимые исправления в черновик.</p>
						</div>
						<div class="flex items-start gap-2.5">
							<div class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 font-bold text-xs mt-0.5">
								2
							</div>
							<p>Скорректируйте спорные данные: проверьте фотографии, описание, правила проживания или реквизиты.</p>
						</div>
						<div class="flex items-start gap-2.5">
							<div class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 font-bold text-xs mt-0.5">
								3
							</div>
							<p>Отправьте объявление повторно — оно сразу вернется на модерацию в приоритетном порядке.</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Footer Actions -->
			<div class="flex flex-col-reverse sm:flex-row items-center justify-end gap-2.5 border-t border-zinc-100 bg-zinc-50/60 px-6 py-4 sm:px-8">
				<Button
					variant="outline"
					tone="neutral"
					size="lg"
					radius="xl"
					onclick={onClose}
					class="w-full sm:w-auto"
				>
					Закрыть
				</Button>

				{#if onEdit || listing}
					<Button
						variant="solid"
						tone="primary"
						size="lg"
						radius="xl"
						iconLeft={Pencil}
						href={`/host/listings/${listing.id}`}
						onclick={handleEdit}
						class="w-full sm:w-auto cursor-pointer"
					>
						Редактировать объявление
					</Button>
				{/if}
			</div>
		</div>
	</div>
{/if}
