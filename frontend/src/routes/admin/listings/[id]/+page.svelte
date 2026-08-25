<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { adminListingsApi } from '$lib/api/admin/listings';
	import type { ListingPublic } from '$lib/types/listings';
	import type { ModerationAction } from '$lib/types/admin';
	import ListingView from '$lib/components/listings/ListingView.svelte';
	import ModerationActionDialog from '$lib/components/admin/ModerationActionDialog.svelte';
	import { ArrowLeft, CheckCircle, XCircle, AlertCircle, ShieldAlert, Loader2 } from 'lucide-svelte';

	let listingId = $derived(page.params.id);
	let listing = $state<ListingPublic | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Action dialog state
	let isDialogOpen = $state(false);
	let dialogAction = $state<ModerationAction>('approve');
	let isSubmitting = $state(false);

	function loadListing(id: string) {
		if (!id) return;
		isLoading = true;
		error = null;
		adminListingsApi
			.getListing(id)
			.then((res) => {
				listing = res;
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки данных жилья';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	$effect(() => {
		if (listingId) {
			loadListing(listingId);
		}
	});

	function openDialog(action: ModerationAction) {
		dialogAction = action;
		isDialogOpen = true;
	}

	function handleConfirmModeration(data: { reason: string; note: string }) {
		if (!listing) return;
		isSubmitting = true;
		adminListingsApi
			.moderateListing(listing.id, {
				action: dialogAction,
				reason: data.reason,
				note: data.note
			})
			.then((updated) => {
				listing = updated;
				isDialogOpen = false;
			})
			.catch((err) => {
				alert(err.message || 'Ошибка обработки модерации');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<a
			href="/admin/listings"
			class="inline-flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
		>
			<ArrowLeft class="h-4 w-4" />
			<span>Назад в очередь модерации</span>
		</a>

		{#if listing}
			<!-- Moderation Decision Toolbar -->
			<div class="flex items-center gap-2">
				<button
					type="button"
					onclick={() => openDialog('approve')}
					class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs"
				>
					<CheckCircle class="h-4 w-4" />
					<span>Одобрить и опубликовать</span>
				</button>
				<button
					type="button"
					onclick={() => openDialog('request_changes')}
					class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs"
				>
					<AlertCircle class="h-4 w-4" />
					<span>Запросить правки</span>
				</button>
				<button
					type="button"
					onclick={() => openDialog('reject')}
					class="px-4 py-2 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs"
				>
					<XCircle class="h-4 w-4" />
					<span>Отклонить</span>
				</button>
				<button
					type="button"
					onclick={() => openDialog('suspend')}
					class="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-900 text-white font-semibold text-xs transition-colors flex items-center gap-1.5 shadow-xs"
				>
					<ShieldAlert class="h-4 w-4" />
					<span>Приостановить</span>
				</button>
			</div>
		{/if}
	</div>

	{#if isLoading}
		<div class="p-12 text-center">
			<Loader2 class="h-8 w-8 text-primary animate-spin mx-auto mb-2" />
			<p class="text-xs text-muted-foreground">Загрузка карточки модерации...</p>
		</div>
	{:else if error}
		<div class="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-sm">
			{error}
		</div>
	{:else if listing}
		<!-- Dedicated Inspection Notice Banner -->
		<div class="p-4 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div class="p-2 rounded-xl bg-primary text-white font-bold text-xs">INSPECT</div>
				<div>
					<p class="text-xs font-bold text-foreground">Инспекционный режим модератора Flickey Compliance</p>
					<p class="text-[11px] text-muted-foreground">ID: {listing.id} • Статус: {listing.status}</p>
				</div>
			</div>
		</div>

		<!-- Reusable Composite Presentation View -->
		<div class="bg-card border border-border rounded-3xl p-6 shadow-xs">
			<ListingView {listing} mode="public" />
		</div>
	{/if}
</div>

{#if listing}
	<ModerationActionDialog
		open={isDialogOpen}
		action={dialogAction}
		listingTitle={listing.name}
		{isSubmitting}
		onConfirm={handleConfirmModeration}
		onCancel={() => (isDialogOpen = false)}
	/>
{/if}
