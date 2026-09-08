<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { adminListingsApi } from '$lib/api/admin/listings';
	import type { ListingPublic } from '$lib/types/listings';
	import type { ModerationAction } from '$lib/types/admin';
	import ModerationActionDialog from '$lib/components/admin/ModerationActionDialog.svelte';
	import AdminActionMenu from '$lib/components/admin/AdminActionMenu.svelte';
	import { toast } from '$lib/stores/toastStore';
	import {
		formatDate,
		formatCurrency,
		translateHousingType,
		translateListingStatus
	} from '$lib/utils';
	import {
		Building2,
		Search,
		Eye,
		CheckCircle,
		XCircle,
		AlertCircle,
		ShieldAlert,
		ExternalLink,
		Loader2,
		X,
		ArrowRight,
		Home
	} from 'lucide-svelte';

	let listings = $state<ListingPublic[]>([]);
	let selectedStatus = $state<string>('pending_review');
	let searchQuery = $state('');
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Action dialog state
	let targetListing = $state<ListingPublic | null>(null);
	let dialogAction = $state<ModerationAction>('approve');
	let isDialogOpen = $state(false);
	let isSubmitting = $state(false);

	function loadListings(statusFilter?: string) {
		isLoading = true;
		error = null;
		adminListingsApi
			.getListings(statusFilter)
			.then((res) => {
				listings = res || [];
			})
			.catch((err) => {
				error = err.message || 'Не удалось загрузить список объявлений';
			})
			.finally(() => {
				isLoading = false;
			});
	}

	onMount(() => {
		const initialStatus = page.url.searchParams.get('status') ?? 'pending_review';
		selectedStatus = initialStatus;
		loadListings(selectedStatus);
	});

	function selectTab(status: string) {
		selectedStatus = status;
		const url = new URL(window.location.href);
		if (status) {
			url.searchParams.set('status', status);
		} else {
			url.searchParams.delete('status');
		}
		window.history.replaceState({}, '', url.toString());
		loadListings(status);
	}

	const filteredListings = $derived.by(() => {
		if (!searchQuery.trim()) return listings;
		const q = searchQuery.toLowerCase().trim();
		return listings.filter((item) => {
			const nameMatch = item.name?.toLowerCase().includes(q);
			const idMatch = item.id?.toLowerCase().includes(q);
			const cityMatch = item.city?.toLowerCase().includes(q);
			const hostMatch = item.host_id?.toLowerCase().includes(q);
			return nameMatch || idMatch || cityMatch || hostMatch;
		});
	});

	function openDecisionDialog(listing: ListingPublic, action: ModerationAction) {
		targetListing = listing;
		dialogAction = action;
		isDialogOpen = true;
	}

	function handleConfirmModeration(data: { reason: string; note: string }) {
		if (!targetListing) return;
		isSubmitting = true;
		const action = dialogAction;
		adminListingsApi
			.moderateListing(targetListing.id, {
				action,
				reason: data.reason,
				note: data.note
			})
			.then(() => {
				isDialogOpen = false;
				targetListing = null;
				toast.success(
					action === 'approve'
						? 'Объявление одобрено и опубликовано'
						: action === 'reject'
							? 'Объявление отклонено'
							: action === 'request_changes'
								? 'Замечания отправлены автору'
								: 'Публикация приостановлена'
				);
				loadListings(selectedStatus);
			})
			.catch((err) => {
				toast.error('Ошибка модерации', err.message || 'Не удалось применить решение');
			})
			.finally(() => {
				isSubmitting = false;
			});
	}

	function navigateToDetail(id: string) {
		goto(`/admin/listings/${id}`);
	}
</script>

<svelte:head>
	<title>Модерация жилья — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Title & Search Bar -->
	<div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
		<div>
			<div class="flex items-center gap-2.5">
				<h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-foreground">
					Модерация жилья
				</h1>
				<span class="flex h-6 min-w-[24px] items-center justify-center rounded-full bg-zinc-900 px-2 text-xs font-bold text-white dark:bg-white dark:text-zinc-900">
					{filteredListings.length}
				</span>
			</div>
			<p class="mt-1 text-sm text-zinc-500 dark:text-muted-foreground">
				Проверка объявлений на соответствие стандартам Flickey, подтверждение описания, цен и удобств.
			</p>
		</div>

		<!-- Search Input -->
		<div class="relative w-full md:w-80">
			<Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Поиск по названию, городу, ID..."
				class="w-full rounded-full border border-zinc-200/80 bg-white py-2 pl-9 pr-9 text-xs text-zinc-900 placeholder:text-zinc-400 focus:border-zinc-900 focus:outline-none focus:ring-1 focus:ring-zinc-900 dark:border-border dark:bg-card dark:text-foreground shadow-2xs"
			/>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200"
				>
					<X class="h-3.5 w-3.5" />
				</button>
			{/if}
		</div>
	</div>

	<!-- Status Tabs Bar -->
	<div class="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-none">
		{#each [
			{ id: 'pending_review', label: 'На проверке' },
			{ id: 'published', label: 'Опубликовано' },
			{ id: 'changes_requested', label: 'Требуют правок' },
			{ id: 'rejected', label: 'Отклонено' },
			{ id: 'suspended', label: 'Приостановлено' },
			{ id: '', label: 'Все статусы' }
		] as tab}
			{@const active = selectedStatus === tab.id}
			<button
				type="button"
				onclick={() => selectTab(tab.id)}
				class="whitespace-nowrap rounded-full px-4 py-2 text-xs font-semibold transition-all cursor-pointer {active
					? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
					: 'border border-zinc-200/80 bg-white text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900 hover:border-zinc-300 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:text-foreground'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>

	<!-- Content Area -->
	{#if isLoading}
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card">
			<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка очереди модерации...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={() => loadListings(selectedStatus)}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if filteredListings.length === 0}
		<div class="rounded-3xl border border-dashed border-zinc-200 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card space-y-3">
			<div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400 mx-auto dark:bg-muted">
				<CheckCircle class="h-7 w-7 text-emerald-500" />
			</div>
			<h3 class="text-base font-bold text-zinc-900 dark:text-foreground">
				{searchQuery ? 'Ничего не найдено' : 'В этой очереди пусто'}
			</h3>
			<p class="text-xs text-zinc-500 max-w-sm mx-auto dark:text-muted-foreground">
				{searchQuery
					? 'Попробуйте изменить поисковый запрос или сбросить фильтры.'
					: 'Все объекты в выбранной категории проверены либо еще не поступили.'}
			</p>
			{#if searchQuery}
				<button
					type="button"
					onclick={() => (searchQuery = '')}
					class="mt-2 inline-flex items-center gap-1.5 rounded-full bg-zinc-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-zinc-800 dark:bg-white dark:text-zinc-900"
				>
					Сбросить поиск
				</button>
			{/if}
		</div>
	{:else}
		<!-- Listings Table with Interactive Rows & Single Action Menu -->
		<div class="rounded-3xl border border-zinc-200/80 bg-white overflow-hidden shadow-2xs dark:border-border dark:bg-card">
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-zinc-100 bg-zinc-50/70 text-[11px] font-bold text-zinc-500 uppercase tracking-wider dark:border-border dark:bg-muted/40 dark:text-muted-foreground">
							<th class="py-4 px-5">Объект</th>
							<th class="py-4 px-4">Тип</th>
							<th class="py-4 px-4">Параметры</th>
							<th class="py-4 px-4">Цена / сутки</th>
							<th class="py-4 px-4">Статус</th>
							<th class="py-4 px-4">Подано</th>
							<th class="py-4 px-5 text-right">Действия</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-zinc-100 text-xs dark:divide-border">
						{#each filteredListings as item (item.id)}
							{@const statusInfo = translateListingStatus(item.status || '')}
							<tr
								onclick={() => navigateToDetail(item.id)}
								class="group cursor-pointer transition-colors hover:bg-zinc-50/80 dark:hover:bg-muted/30"
							>
								<!-- Property Info -->
								<td class="py-4 px-5">
									<div class="flex items-center gap-3.5">
										{#if item.media && item.media.length > 0}
											<img
												src={item.media[0]}
												alt={item.name}
												class="h-12 w-14 rounded-xl object-cover border border-zinc-200/80 shrink-0 dark:border-border"
											/>
										{:else}
											<div class="flex h-12 w-14 shrink-0 items-center justify-center rounded-xl bg-zinc-100 border border-zinc-200/80 text-[10px] text-zinc-400 font-medium dark:bg-muted dark:border-border">
												Без фото
											</div>
										{/if}

										<div class="min-w-0 max-w-xs">
											<div class="font-bold text-zinc-900 group-hover:text-zinc-600 transition-colors truncate dark:text-foreground dark:group-hover:text-primary">
												{item.name}
											</div>
											<div class="flex items-center gap-1.5 text-[11px] text-zinc-400 font-mono mt-0.5">
												<span>{item.city || 'Минск'}</span>
												<span>•</span>
												<span>ID: {item.id.slice(0, 8)}</span>
											</div>
										</div>
									</div>
								</td>

								<!-- Housing Type -->
								<td class="py-4 px-4 font-medium text-zinc-700 dark:text-zinc-300">
									{translateHousingType(item.type)}
								</td>

								<!-- Parameters -->
								<td class="py-4 px-4 text-zinc-500 dark:text-muted-foreground">
									<span>{item.square} м²</span>
									<span>•</span>
									<span>{item.rooms_count} комн.</span>
									<span>•</span>
									<span>до {item.max_guests} гостей</span>
								</td>

								<!-- Price -->
								<td class="py-4 px-4 font-bold text-zinc-900 dark:text-foreground">
									{formatCurrency(item.price_per_night, item.currency)}
								</td>

								<!-- Status Badge -->
								<td class="py-4 px-4">
									<span class="inline-flex items-center rounded-full px-2.5 py-1 text-[11px] font-bold border {statusInfo.color}">
										{statusInfo.label}
									</span>
								</td>

								<!-- Submission Date -->
								<td class="py-4 px-4 text-zinc-500 whitespace-nowrap dark:text-muted-foreground">
									{formatDate(item.created_at)}
								</td>

								<!-- Clean Action: 1 primary review button + 1 kebab dropdown -->
								<td class="py-4 px-5 text-right" onclick={(e) => e.stopPropagation()}>
									<div class="flex items-center justify-end gap-2">
										{#if item.status === 'pending_review'}
											<a
												href="/admin/listings/{item.id}"
												class="inline-flex items-center gap-1.5 rounded-full bg-zinc-900 px-3.5 py-1.5 text-xs font-semibold text-white shadow-2xs transition hover:bg-zinc-800 active:scale-95 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-100"
											>
												<span>Проверить</span>
												<ArrowRight class="h-3 w-3" />
											</a>
										{/if}

										<AdminActionMenu
											items={[
												{
													label: 'Инспекция объявления',
													icon: Eye,
													variant: 'primary',
													onclick: () => navigateToDetail(item.id)
												},
												{
													label: 'Одобрить публикацию',
													icon: CheckCircle,
													variant: 'success',
													onclick: () => openDecisionDialog(item, 'approve')
												},
												{
													label: 'Запросить правки',
													icon: AlertCircle,
													variant: 'warning',
													onclick: () => openDecisionDialog(item, 'request_changes')
												},
												{
													label: 'Отклонить объявление',
													icon: XCircle,
													variant: 'danger',
													onclick: () => openDecisionDialog(item, 'reject')
												},
												{
													label: 'Приостановить',
													icon: ShieldAlert,
													divider: true,
													onclick: () => openDecisionDialog(item, 'suspend')
												},
												{
													label: 'Просмотреть на сайте',
													icon: ExternalLink,
													divider: true,
													onclick: () => window.open(`/listings/${item.id}`, '_blank')
												}
											]}
										/>
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