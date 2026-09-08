<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { adminListingsApi } from '$lib/api/admin/listings';
	import type { ListingPublic } from '$lib/types/listings';
	import type { ModerationAction } from '$lib/types/admin';
	import ListingView from '$lib/components/listings/ListingView.svelte';
	import ModerationActionDialog from '$lib/components/admin/ModerationActionDialog.svelte';
	import AdminActionMenu from '$lib/components/admin/AdminActionMenu.svelte';
	import { toast } from '$lib/stores/toastStore';
	import { formatApiError } from '$lib/utils/formatError';
	import { translateListingStatus } from '$lib/utils';
	import {
		ArrowLeft,
		CheckCircle,
		XCircle,
		AlertCircle,
		ShieldAlert,
		Loader2,
		ExternalLink,
		User,
		Building2,
		Video,
		Lock
	} from 'lucide-svelte';

	let listingId = $derived(page.params.id);
	let listing = $state<(ListingPublic & { verification_video_url?: string | null; verification_video_id?: string | null }) | null>(null);
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
				error = err.message || 'Не удалось загрузить данные объявления';
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
		const action = dialogAction;
		adminListingsApi
			.moderateListing(listing.id, {
				action,
				reason: data.reason,
				note: data.note
			})
			.then(() => {
				isDialogOpen = false;
				const actionMessages: Record<ModerationAction, string> = {
					approve: 'Объявление успешно одобрено и опубликовано',
					reject: 'Объявление отклонено',
					request_changes: 'Замечания отправлены автору объявления',
					suspend: 'Публикация объявления приостановлена'
				};
				toast.success(actionMessages[action] || 'Решение модерации зафиксировано');
				goto('/admin/listings');
			})
			.catch((err) => {
				toast.error('Ошибка модерации', formatApiError(err, 'Не удалось применить решение'));
			})
			.finally(() => {
				isSubmitting = false;
			});
	}
</script>

<svelte:head>
	<title>{listing ? `${listing.name} — Инспекция жилья` : 'Инспекция объявления'} — Flickey Admin</title>
</svelte:head>

<div class="space-y-6">
	<!-- Top Sticky Action Bar with Hierarchy -->
	<div class="sticky top-20 z-30 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 rounded-3xl border border-zinc-200/80 bg-white/95 p-4 sm:px-6 shadow-sm backdrop-blur-md dark:border-border dark:bg-card/95">
		<!-- Left: Breadcrumb & Title -->
		<div class="flex items-center gap-3 min-w-0">
			<a
				href="/admin/listings"
				class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-zinc-200/80 bg-white text-zinc-600 transition hover:bg-zinc-50 hover:text-zinc-900 active:scale-95 dark:border-border dark:bg-card dark:text-muted-foreground"
				title="Назад в список"
			>
				<ArrowLeft class="h-4 w-4" />
			</a>

			<div class="min-w-0">
				<div class="flex items-center gap-2">
					<span class="text-xs font-semibold text-zinc-500 dark:text-muted-foreground">Модерация жилья</span>
					{#if listing}
						{@const statusInfo = translateListingStatus(listing.status || '')}
						<span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold border {statusInfo.color}">
							{statusInfo.label}
						</span>
					{/if}
				</div>
				<h1 class="text-sm sm:text-base font-bold text-zinc-900 truncate dark:text-foreground">
					{listing?.name || 'Загрузка объекта...'}
				</h1>
			</div>
		</div>

		<!-- Right: Action Buttons Hierarchy -->
		{#if listing}
			<div class="flex items-center gap-2 shrink-0">
				<button
					type="button"
					onclick={() => openDialog('approve')}
					class="inline-flex items-center gap-1.5 rounded-full bg-emerald-600 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-emerald-700 active:scale-95 cursor-pointer"
				>
					<CheckCircle class="h-3.5 w-3.5" />
					<span>Одобрить</span>
				</button>

				<button
					type="button"
					onclick={() => openDialog('request_changes')}
					class="inline-flex items-center gap-1.5 rounded-full border border-amber-300 bg-amber-50 px-3.5 py-2 text-xs font-semibold text-amber-800 transition hover:bg-amber-100 active:scale-95 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-300 cursor-pointer"
				>
					<AlertCircle class="h-3.5 w-3.5" />
					<span>Правки</span>
				</button>

				<button
					type="button"
					onclick={() => openDialog('reject')}
					class="inline-flex items-center gap-1.5 rounded-full border border-rose-200 bg-rose-50 px-3.5 py-2 text-xs font-semibold text-rose-700 transition hover:bg-rose-100 active:scale-95 dark:border-rose-900 dark:bg-rose-950 dark:text-rose-300 cursor-pointer"
				>
					<XCircle class="h-3.5 w-3.5" />
					<span>Отклонить</span>
				</button>

				<AdminActionMenu
					items={[
						{
							label: 'Приостановить публикацию',
							icon: ShieldAlert,
							variant: 'warning',
							onclick: () => openDialog('suspend')
						},
						{
							label: 'Профиль хозяина',
							icon: User,
							onclick: () => goto(`/admin/users/${listing?.host_id}`)
						},
						{
							label: 'Открыть на сайте',
							icon: ExternalLink,
							divider: true,
							onclick: () => window.open(`/listings/${listing?.id}`, '_blank')
						}
					]}
				/>
			</div>
		{/if}
	</div>

	<!-- Main Detail Content -->
	{#if isLoading}
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-16 text-center shadow-2xs dark:border-border dark:bg-card">
			<Loader2 class="h-8 w-8 text-zinc-900 animate-spin mx-auto mb-3 dark:text-white" />
			<p class="text-xs font-medium text-zinc-500 dark:text-muted-foreground">Загрузка данных объекта...</p>
		</div>
	{:else if error}
		<div class="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-800 dark:border-rose-900 dark:bg-rose-950/60 dark:text-rose-200 flex items-center justify-between">
			<span>{error}</span>
			<button
				type="button"
				onclick={() => { if (listingId) loadListing(listingId); }}
				class="font-semibold underline hover:no-underline cursor-pointer"
			>
				Повторить
			</button>
		</div>
	{:else if listing}
		<!-- Private Verification Video Section -->
		{#if listing.verification_video_url}
			<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-8 shadow-2xs dark:border-border dark:bg-card">
				<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-4">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-violet-100 text-violet-700 dark:bg-violet-950 dark:text-violet-300">
							<Video class="h-5 w-5" />
						</div>
						<div>
							<div class="flex items-center gap-2">
								<h2 class="text-sm sm:text-base font-bold text-zinc-900 dark:text-foreground">
									Видео верификации квартиры
								</h2>
								<span class="inline-flex items-center gap-1 rounded-full bg-zinc-100 px-2 py-0.5 text-[10px] font-bold text-zinc-600 dark:bg-muted dark:text-muted-foreground">
									<Lock class="h-2.5 w-2.5" /> Только для админа
								</span>
							</div>
							<p class="text-xs text-zinc-500 dark:text-muted-foreground mt-0.5">
								Видео-обход квартиры от арендодателя для подтверждения владения и реального состояния жилья.
							</p>
						</div>
					</div>

					<a
						href={listing.verification_video_url}
						target="_blank"
						rel="noreferrer"
						class="inline-flex items-center gap-1.5 self-start sm:self-auto rounded-xl border border-zinc-200 bg-white px-3 py-1.5 text-xs font-semibold text-zinc-700 hover:bg-zinc-50 dark:border-border dark:bg-card dark:text-muted-foreground dark:hover:bg-muted"
					>
						<ExternalLink class="h-3.5 w-3.5" />
						Открыть видео
					</a>
				</div>

				<div class="overflow-hidden rounded-2xl bg-black aspect-video max-w-2xl mx-auto shadow-inner">
					<video
						src={listing.verification_video_url}
						controls
						playsinline
						class="h-full w-full object-contain"
					>
						<track kind="captions" />
						Ваш браузер не поддерживает воспроизведение видео.
					</video>
				</div>
			</div>
		{:else}
			<div class="rounded-3xl border border-amber-200 bg-amber-50/70 p-5 dark:border-amber-900/60 dark:bg-amber-950/30">
				<div class="flex items-center gap-3">
					<AlertCircle class="h-5 w-5 text-amber-600 shrink-0" />
					<div class="text-xs text-amber-800 dark:text-amber-300">
						<span class="font-semibold">Видео верификации не прикреплено.</span>
						Хозяин ещё не загрузил видео-обход для этого объекта.
					</div>
				</div>
			</div>
		{/if}

		<!-- Listing View Container -->
		<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-8 shadow-2xs dark:border-border dark:bg-card">
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