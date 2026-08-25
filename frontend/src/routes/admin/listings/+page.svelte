<script lang="ts">
	import { onMount } from 'svelte';
	import { adminListingsApi } from '$lib/api/admin/listings';
	import type { ListingPublic } from '$lib/types/listings';
	import type { ModerationAction } from '$lib/types/admin';
	import ModerationActionDialog from '$lib/components/admin/ModerationActionDialog.svelte';
	import { translateListingStatus, translateHousingType, formatCurrency, formatDate } from '$lib/utils';
	import { Eye, CheckCircle, XCircle, AlertCircle, ShieldAlert, Loader2, Filter } from 'lucide-svelte';

	let listings = $state<ListingPublic[]>([]);
	let selectedStatus = $state('pending_review');
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Action Modal state
	let isDialogOpen = $state(false);
	let targetListing = $state<ListingPublic | null>(null);
	let dialogAction = $state<ModerationAction>('approve');
	let isSubmitting = $state(false);

	function loadListings(statusFilter: string) {
		isLoading = true;
		error = null;
		adminListingsApi
			.getListings(statusFilter)
			.then((res) => {
				listings = res || [];
			})
			.catch((err) => {
				error = err.message || 'Ошибка загрузки объявлений';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	onMount(() => {
		loadListings(selectedStatus);
	});

	function selectTab(status: string) {
		selectedStatus = status;
		loadListings(status);
	}

	function openDialog(listing: ListingPublic, action: ModerationAction) {
		targetListing = listing;
		dialogAction = action;
		isDialogOpen = true;
	}

	function handleConfirmModeration(data: { reason: string; note: string }) {
		if (!targetListing) return;
		isSubmitting = true;
		adminListingsApi
			.moderateListing(targetListing.id, {
				action: dialogAction,
				reason: data.reason,
				note: data.note
			})
			.then(() => {
				isDialogOpen = false;
				targetListing = null;
				loadListings(selectedStatus);
			})
			.catch((err) => {
				alert(err.message || 'Ошибка выполнения операции модерации');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}
</script>

<div class="space-y-6">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-foreground tracking-tight">Модерация объектов недвижимости</h1>
			<p class="text-sm text-muted-foreground">Проверка поступающих объявлений, проверка юридических требований и вынесение решений.</p>
		</div>

		<!-- Status Filter Tabs -->
		<div class="flex items-center space-x-1 bg-muted/40 p-1 rounded-2xl border border-border overflow-x-auto scrollbar-none">
			{#each [
				{ id: 'pending_review', label: 'На модерации' },
				{ id: 'published', label: 'Опубликовано' },
				{ id: 'changes_requested', label: 'Правки' },
				{ id: 'rejected', label: 'Отклонено' },
				{ id: 'suspended', label: 'Приостановлено' },
				{ id: '', label: 'Все объекты' }
			] as tab}
				<button
					type="button"
					onclick={() => selectTab(tab.id)}
					class="px-3 py-1.5 rounded-xl text-xs font-semibold whitespace-nowrap transition-colors {selectedStatus === tab.id
						? 'bg-card text-foreground shadow-xs'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					{tab.label}
				</button>
			{/each}
		</div>
	</div>

	{#if isLoading}
		<div class="p-12 text-center">
			<Loader2 class="h-8 w-8 text-primary animate-spin mx-auto mb-2" />
			<p class="text-xs text-muted-foreground">Загрузка очереди модерации...</p>
		</div>
	{:else if error}
		<div class="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 dark:bg-rose-950 dark:border-rose-900 text-sm">
			{error}
		</div>
	{:else if listings.length === 0}
		<div class="p-12 rounded-2xl border border-dashed border-border bg-card text-center space-y-2">
			<CheckCircle class="h-10 w-10 text-emerald-500 mx-auto" />
			<h3 class="text-base font-bold text-foreground">Очередь пуста</h3>
			<p class="text-xs text-muted-foreground max-w-sm mx-auto">В данной категории нет объявлений, требующих внимания модератора.</p>
		</div>
	{:else}
		<!-- Listings Table -->
		<div class="rounded-2xl border border-border bg-card overflow-hidden shadow-xs">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-border bg-muted/40 text-[11px] font-bold text-muted-foreground uppercase">
							<th class="py-3.5 px-4">Объект</th>
							<th class="py-3.5 px-4">Тип</th>
							<th class="py-3.5 px-4">Параметры</th>
							<th class="py-3.5 px-4">Цена / сутки</th>
							<th class="py-3.5 px-4">Статус</th>
							<th class="py-3.5 px-4">Дата создания</th>
							<th class="py-3.5 px-4 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-border text-xs">
						{#each listings as listing}
							{@const statusInfo = translateListingStatus(listing.status || '')}
							<tr class="hover:bg-muted/20 transition-colors">
								<td class="py-3.5 px-4">
									<div class="flex items-center gap-3">
										{#if listing.media && listing.media.length > 0}
											<img src={listing.media[0]} alt={listing.name} class="w-12 h-12 rounded-xl object-cover border border-border shrink-0" />
										{:else}
											<div class="w-12 h-12 rounded-xl bg-muted border border-border shrink-0 flex items-center justify-center text-muted-foreground text-[10px]">
												Нет фото
											</div>
										{/if}
										<div>
											<a href="/admin/listings/{listing.id}" class="font-bold text-foreground hover:text-primary transition-colors line-clamp-1">
												{listing.name}
											</a>
											<p class="text-[11px] text-muted-foreground font-mono">ID: {listing.id.slice(0, 8)}...</p>
										</div>
									</div>
								</td>
								<td class="py-3.5 px-4 font-medium text-foreground">
									{translateHousingType(listing.type)}
								</td>
								<td class="py-3.5 px-4 text-muted-foreground">
									{listing.square} м² • {listing.rooms_count} комн. • до {listing.max_guests} гостей
								</td>
								<td class="py-3.5 px-4 font-bold text-foreground">
									{formatCurrency(listing.price_per_night, listing.currency)}
								</td>
								<td class="py-3.5 px-4">
									<span class="px-2.5 py-1 rounded-full text-[11px] font-semibold border {statusInfo.color}">
										{statusInfo.label}
									</span>
								</td>
								<td class="py-3.5 px-4 text-muted-foreground whitespace-nowrap">
									{formatDate(listing.created_at)}
								</td>
								<td class="py-3.5 px-4 text-right">
									<div class="flex items-center justify-end gap-1.5">
										<a
											href="/admin/listings/{listing.id}"
											class="p-2 rounded-xl border border-border hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
											title="Детальный инспектор"
										>
											<Eye class="h-4 w-4" />
										</a>
										<button
											type="button"
											onclick={() => openDialog(listing, 'approve')}
											class="p-2 rounded-xl bg-emerald-100 text-emerald-800 hover:bg-emerald-200 dark:bg-emerald-950 dark:text-emerald-300 transition-colors"
											title="Одобрить"
										>
											<CheckCircle class="h-4 w-4" />
										</button>
										<button
											type="button"
											onclick={() => openDialog(listing, 'request_changes')}
											class="p-2 rounded-xl bg-amber-100 text-amber-800 hover:bg-amber-200 dark:bg-amber-950 dark:text-amber-300 transition-colors"
											title="Запросить правки"
										>
											<AlertCircle class="h-4 w-4" />
										</button>
										<button
											type="button"
											onclick={() => openDialog(listing, 'reject')}
											class="p-2 rounded-xl bg-rose-100 text-rose-800 hover:bg-rose-200 dark:bg-rose-950 dark:text-rose-300 transition-colors"
											title="Отклонить"
										>
											<XCircle class="h-4 w-4" />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

{#if targetListing}
	<ModerationActionDialog
		open={isDialogOpen}
		action={dialogAction}
		listingTitle={targetListing.name}
		{isSubmitting}
		onConfirm={handleConfirmModeration}
		onCancel={() => (isDialogOpen = false)}
	/>
{/if}
