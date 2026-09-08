<!-- src/lib/components/host-listing-edit/ListingEditHeader.svelte -->
<script lang="ts">
	import { ArrowLeft, Send, Loader2 } from 'lucide-svelte';

	interface Props {
		listingTitle?: string;
		listingStatus?: string;
		isDirty?: boolean;
		isSaving?: boolean;
		onBack?: () => void;
		onSubmitForModeration?: () => void;
	}

	let { 
		listingTitle = '', 
		listingStatus = '',
		isDirty = false, 
		isSaving = false,
		onBack,
		onSubmitForModeration
	}: Props = $props();

	const canSubmitToModeration = $derived(
		listingStatus === 'draft' || 
		listingStatus === 'archived' || 
		listingStatus === 'changes_requested' ||
		listingStatus === 'rejected'
	);
</script>

<header class="sticky top-0 z-40 border-b border-zinc-200 bg-white/95 backdrop-blur-md">
	<div class="flex items-center justify-between gap-4 px-4 py-3 sm:px-6 lg:px-8">
		<!-- Left: Back Button + Title -->
		<div class="flex items-center gap-3.5 min-w-0">
			{#if onBack}
				<button
					type="button"
					class="shrink-0 flex items-center gap-2 rounded-xl border border-zinc-200 bg-white px-3.5 py-2 text-xs sm:text-sm font-semibold text-zinc-700 shadow-xs transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 cursor-pointer"
					onclick={onBack}
					aria-label="Вернуться назад"
				>
					<ArrowLeft class="h-4 w-4" strokeWidth={2} />
					<span>Вернуться назад</span>
				</button>
			{/if}

			<div class="hidden sm:block h-5 w-px bg-zinc-200"></div>

			<div class="min-w-0">
				<h1 class="truncate text-sm sm:text-base font-semibold text-zinc-900 leading-tight">
					Редактирование {listingTitle ? `«${listingTitle}»` : 'объявления'}
				</h1>
			</div>
		</div>

		<!-- Right: Actions + Status indicator badge -->
		<div class="flex shrink-0 items-center gap-3">
			{#if canSubmitToModeration && onSubmitForModeration}
				<button
					type="button"
					class="flex items-center gap-2 rounded-xl bg-emerald-600 px-3.5 py-2 text-xs sm:text-sm font-semibold text-white shadow-sm transition hover:bg-emerald-700 active:scale-95 disabled:opacity-50 cursor-pointer"
					onclick={onSubmitForModeration}
					disabled={isSaving}
					title="Отправить объявление на проверку модераторам"
				>
					{#if isSaving}
						<Loader2 class="h-4 w-4 animate-spin" />
						<span>Отправка...</span>
					{:else}
						<Send class="h-4 w-4" />
						<span>Отправить на модерацию</span>
					{/if}
				</button>
			{/if}

			{#if isDirty}
				<div class="hidden sm:inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-800 ring-1 ring-amber-200/70">
					<span class="h-1.5 w-1.5 rounded-full bg-amber-500 animate-pulse"></span>
					<span>Несохранённые правки</span>
				</div>
			{:else}
				<span class="hidden sm:inline text-xs text-zinc-400">Все изменения сохранены</span>
			{/if}
		</div>
	</div>
</header>
