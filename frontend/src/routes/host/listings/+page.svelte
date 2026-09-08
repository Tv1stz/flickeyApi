<!-- src/routes/host/listings/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { Plus, PencilLine, FileText, Clock, AlertCircle, Building2, Video, CheckCircle2, Calendar, BarChart3 } from 'lucide-svelte';

	import Button from '$lib/components/ui/Button.svelte';
	import HostListingCard from '$lib/components/listing/HostListingCard.svelte';
	import ListingVerificationModal from '$lib/components/verification/ListingVerificationModal.svelte';
	import ListingModerationModal from '$lib/components/listing/ListingModerationModal.svelte';
	import PageShell from '$lib/components/ui/page/PageShell.svelte';
	import PageHeader from '$lib/components/ui/page/PageHeader.svelte';
	import ViewModeToggle from '$lib/components/ui/page/ViewModeToggle.svelte';
	import EmptyState from '$lib/components/ui/page/EmptyState.svelte';

	import { hostListingsStore } from '$lib/stores/listingsStore.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { canAccessHostArea } from '$lib/auth/permissions';
	import { toast } from '$lib/stores/toastStore';
	import { clearDraft, getDraftAge, loadDraft, saveDraft } from '$lib/utils/formStorage';
	import { pluralRu } from '$lib/utils/format';
	import { verificationApi } from '$lib/api/verification';
	import type { VerificationResponse } from '$lib/types/verification';
	import type { Listing } from '$lib/components/card/types';
	import type { DraftMeta } from '$lib/utils/formStorage';
	import type { ListingFormValues } from '$lib/validation/listingValidation';

	let viewMode = $state<'grid' | 'list'>('list');
	let listings = $derived(hostListingsStore.items);
	
	let guardChecked = $state(false);
	let currentUserId = $state<string | null>(null);
	let currentUserListingsCount = $state(0);
	let draftAge = $state<string | null>(null);
	let draftMeta = $state<DraftMeta | null>(null);
	let hasDraft = $state(false);

	let hostVerification = $state<VerificationResponse | null>(null);

	let verificationModalOpen = $state(false);
	let selectedListingForVerification = $state<Listing | null>(null);

	let feedbackModalOpen = $state(false);
	let selectedListingForFeedback = $state<Listing | null>(null);

	function openFeedbackModal(e: MouseEvent, listing: Listing) {
		e.preventDefault();
		e.stopPropagation();
		selectedListingForFeedback = listing;
		feedbackModalOpen = true;
	}

	function closeFeedbackModal() {
		feedbackModalOpen = false;
		selectedListingForFeedback = null;
	}

	async function loadHostVerificationStatus() {
		try {
			hostVerification = await verificationApi.getMyVerification();
		} catch (e) {
			console.warn('Failed to load host verification status:', e);
		}
	}

	function openVerificationModal(e: MouseEvent, listing: Listing) {
		e.preventDefault();
		e.stopPropagation();
		selectedListingForVerification = listing;
		verificationModalOpen = true;
	}

	function closeVerificationModal() {
		verificationModalOpen = false;
		selectedListingForVerification = null;
	}

	$effect(() => {
		currentUserId = authStore.user?.id ?? null;
		currentUserListingsCount = authStore.user?.hostProfile?.listingsCount || 0;
		updateDraftState();
	});

	$effect(() => {
		if (authStore.initialized) {
			if (!authStore.user) {
				authStore.setPendingAction(null, '/host/listings');
				goto(resolve('/auth'));
				return;
			}
			if (!guardChecked) {
				if (!canAccessHostArea(authStore.user)) {
					goto(resolve('/profile'));
					return;
				}
				if (authStore.viewMode !== 'host') authStore.setViewMode('host');
				guardChecked = true;
				hostListingsStore.refresh();
			}
		}
	});

	onMount(() => {
		hostListingsStore.initialize();
		loadHostVerificationStatus();
		if (authStore.initialized && !guardChecked) {
			if (!authStore.user) {
				authStore.setPendingAction(null, '/host/listings');
				goto(resolve('/auth'));
				return;
			}
			if (!canAccessHostArea(authStore.user)) {
				goto(resolve('/profile'));
				return;
			}
			if (authStore.viewMode !== 'host') authStore.setViewMode('host');
			guardChecked = true;
		}
		updateDraftState();
	});


	const listingsMeta = $derived.by(() => {
		if (listings.length === 0) return '';
		return `${listings.length} ${pluralRu(listings.length, ['объявление', 'объявления', 'объявлений'])}`;
	});

	import { listingsApi } from '$lib/api/listings';

	function openListing(id: string) {
		goto(resolve('/listings/[id]', { id }));
	}

	function handleEditListing(eOrListing?: MouseEvent | Listing, maybeListing?: Listing) {
		if (eOrListing && 'preventDefault' in eOrListing) {
			eOrListing.preventDefault();
			eOrListing.stopPropagation();
		}
		const target = (maybeListing || ((eOrListing as Listing)?.id ? eOrListing : selectedListingForFeedback)) as Listing | null;
		closeFeedbackModal();
		if (target?.id) {
			goto(`/host/listings/${target.id}`);
		}
	}

	async function handleArchiveListing(e: MouseEvent, listing: Listing) {
		e.preventDefault();
		e.stopPropagation();
		toast.confirm({
			title: 'Архивировать объявление?',
			message: `«${listing.title}» будет скрыто из публичного поиска. Вы сможете разархивировать его в любой момент.`,
			confirmText: 'В архив',
			cancelText: 'Отмена',
			type: 'danger',
			onConfirm: async () => {
				try {
					await hostListingsStore.archive(listing.id);
					toast.success('Объявление перемещено в архив');
				} catch (err: any) {
					toast.error('Ошибка', err?.message || 'Не удалось архивировать объявление');
				}
			}
		});
	}

	async function handleUnarchiveListing(e: MouseEvent, listing: Listing) {
		e.preventDefault();
		e.stopPropagation();
		try {
			const updated = await hostListingsStore.unarchive(listing.id);
			if (updated?.status === 'pending_review' || updated?.status === 'awaiting_company_verification') {
				toast.success('Объявление отправлено на модерацию');
			} else {
				toast.success('Объявление опубликовано');
			}
		} catch (err: any) {
			toast.error('Ошибка', err?.message || 'Не удалось восстановить объявление');
		}
	}

	async function deleteListingById(id: string) {
		try {
			await hostListingsStore.remove(id);
			authStore.decrementListingsCount();
			toast.success('Объявление удалено');
		} catch {
			toast.error('Не удалось удалить объявление');
		}
	}


	function requestDelete(e: MouseEvent, listing: Listing) {
		e.preventDefault();
		e.stopPropagation();
		toast.confirm({
			title: 'Удалить объявление?',
			message: `«${listing.title}» будет удалено без возможности восстановления.`,
			confirmText: 'Удалить',
			cancelText: 'Отмена',
			type: 'danger',
			onConfirm: () => deleteListingById(listing.id)
		});
	}

	function updateDraftState() {
		if (!currentUserId || currentUserListingsCount === 0) {
			if (currentUserId) clearDraft(currentUserId);
			hasDraft = false;
			draftMeta = null;
			draftAge = null;
			return;
		}
		const draft = loadDraft(currentUserId);
		hasDraft = Boolean(draft);
		draftMeta = draft?.meta ?? null;
		draftAge = getDraftAge(currentUserId);
	}

	function continueDraft() {
		goto(resolve('/listings/new'), { state: { resumeDraft: true } });
	}

	function discardDraft() {
		if (!currentUserId) return;
		clearDraft(currentUserId);
		updateDraftState();
		toast.success('Черновик удалён');
	}
</script>

<svelte:head>
	<title>Мои объявления — Flickey</title>
</svelte:head>

{#if guardChecked}
	<PageShell maxWidth="wide" background="white">
		<PageHeader title="Мои объявления" meta={listingsMeta}>
			{#snippet actions()}
				{#if listings.length > 0}
					<ViewModeToggle value={viewMode} onChange={(v) => (viewMode = v)} />
					<Button
						variant="outline"
						tone="neutral"
						size="lg"
						radius="xl"
						href={resolve('/host/calendar')}
						as="a"
					>
						<Calendar size={16} class="mr-1.5" />
						Календарь
					</Button>
					<Button
						variant="outline"
						tone="neutral"
						size="lg"
						radius="xl"
						href={resolve('/host/stats')}
						as="a"
					>
						<BarChart3 size={16} class="mr-1.5" />
						Статистика
					</Button>
					<Button
						variant="solid"
						tone="primary"
						size="lg"
						radius="xl"
						href={resolve('/listings/new')}
						as="a"
					>
						<Plus size={16} class="mr-1.5" />
						Создать
					</Button>
				{/if}
			{/snippet}
		</PageHeader>

		<!-- Баннер верификации партнера (Беларусь) -->
		{#if hostVerification && hostVerification.status !== 'approved'}
			{#if hostVerification.status === 'pending'}
				<div class="mb-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-2xl border border-amber-200 bg-amber-50/80 px-5 py-4 dark:border-amber-900/60 dark:bg-amber-950/30">
					<div class="flex items-start sm:items-center gap-3.5">
						<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300">
							<Clock class="h-5 w-5" />
						</div>
						<div class="text-xs">
							<p class="font-bold text-amber-950 dark:text-amber-200">
								Документы партнера на проверке
							</p>
							<p class="text-amber-700 dark:text-amber-300 mt-0.5 leading-relaxed">
								Заявка на верификацию рассматривается модератором ({hostVerification.legal_name}). Как только документы будут одобрены, ваши объявления с прикрепленным видео автоматически перейдут на модерацию.
							</p>
						</div>
					</div>
					<Button
						variant="outline"
						size="sm"
						radius="lg"
						href={resolve('/host/verification')}
						as="a"
						class="shrink-0 self-start sm:self-auto"
					>
						Статус заявки
					</Button>
				</div>
			{:else if hostVerification.status === 'rejected'}
				<div class="mb-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-2xl border border-rose-200 bg-rose-50/80 px-5 py-4 dark:border-rose-900/60 dark:bg-rose-950/30">
					<div class="flex items-start sm:items-center gap-3.5">
						<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-rose-100 text-rose-700 dark:bg-rose-900 dark:text-rose-300">
							<AlertCircle class="h-5 w-5" />
						</div>
						<div class="text-xs">
							<p class="font-bold text-rose-950 dark:text-rose-200">
								Требуется исправление документов партнера
							</p>
							<p class="text-rose-700 dark:text-rose-300 mt-0.5 leading-relaxed">
								{hostVerification.rejection_reason || 'Модератор запросил уточнение данных для соответствия законодательству РБ.'}
							</p>
						</div>
					</div>
					<Button
						variant="solid"
						tone="danger"
						size="sm"
						radius="lg"
						href={resolve('/host/verification')}
						as="a"
						class="shrink-0 self-start sm:self-auto"
					>
						Исправить данные
					</Button>
				</div>
			{:else}
				<div class="mb-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-2xl border border-blue-200/80 bg-blue-50/70 px-5 py-4 dark:border-blue-900/60 dark:bg-blue-950/30">
					<div class="flex items-start sm:items-center gap-3.5">
						<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
							<Building2 class="h-5 w-5" />
						</div>
						<div class="text-xs">
							<p class="font-bold text-blue-950 dark:text-blue-200">
								Подтвердите статус партнера по законодательству РБ
							</p>
							<p class="text-blue-700 dark:text-blue-300 mt-0.5 leading-relaxed">
								Для публикации объявлений и сдачи жилья в аренду предоставьте реквизиты (физлицо/самозанятый или компания/ИП). Заполняется 1 раз для всех ваших объектов.
							</p>
						</div>
					</div>
					<Button
						variant="solid"
						tone="primary"
						size="sm"
						radius="lg"
						href={resolve('/host/verification')}
						as="a"
						class="shrink-0 self-start sm:self-auto"
					>
						Пройти верификацию
					</Button>
				</div>
			{/if}
		{/if}

		<!-- Черновик -->
		{#if hasDraft}
			<div
				class="mb-8 flex items-center gap-4 rounded-2xl border border-zinc-200 bg-zinc-50 px-5 py-4"
			>
				<div
					class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white shadow-sm ring-1 ring-black/5"
				>
					<FileText size={18} class="text-zinc-400" />
				</div>
				<div class="min-w-0 flex-1">
					<p class="text-sm font-semibold text-zinc-900">Незавершённый черновик</p>
					<p class="mt-0.5 text-xs text-zinc-500">
						{#if draftAge}Обновлён {draftAge}{:else}Готов к продолжению{/if}{#if draftMeta?.step}&ensp;·&ensp;шаг
							{draftMeta.step}{/if}{#if draftMeta?.photoCount}&ensp;·&ensp;{draftMeta.photoCount} фото{/if}
					</p>
				</div>
				<div class="flex shrink-0 items-center gap-2">
					<Button variant="solid" tone="primary" size="sm" radius="lg" onclick={continueDraft}>
						Продолжить
					</Button>
					<Button variant="ghost" tone="neutral" size="sm" radius="lg" onclick={discardDraft}>
						Удалить
					</Button>
				</div>
			</div>
		{/if}

		{#if listings.length > 0}
			<!-- ═══════════════════════════════ LIST ═══════════════════════════════ -->
			{#if viewMode === 'list'}
				<div class="overflow-hidden rounded-2xl border border-zinc-200">
					<!-- Заголовки -->
					<div
						class="hidden items-center gap-4 border-b border-zinc-100 bg-zinc-50 px-6 py-3 sm:grid"
						style="grid-template-columns: 1fr 110px 160px 240px 140px;"
					>
						{#each ['Объявление', 'Тип', 'Местоположение', 'Статус', ''] as col (col)}
							<span class="text-[11px] font-semibold tracking-wider text-zinc-400 uppercase"
								>{col}</span
							>
						{/each}
					</div>

					<!-- Строки -->
					<div class="divide-y divide-zinc-100">
						{#each listings as listing (listing.id)}
							<HostListingCard
								{listing}
								variant="list"
								onOpen={openListing}
								onDelete={requestDelete}
								onEdit={handleEditListing}
								onArchive={handleArchiveListing}
								onUnarchive={handleUnarchiveListing}
								onVerify={openVerificationModal}
								onShowFeedback={openFeedbackModal}
							/>
						{/each}
					</div>
				</div>

				<!-- ═══════════════════════════════ GRID ═══════════════════════════════ -->
			{:else}
				<div class="grid grid-cols-1 gap-x-5 gap-y-10 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each listings as listing (listing.id)}
						<HostListingCard
							{listing}
							variant="grid"
							onOpen={openListing}
							onDelete={requestDelete}
							onEdit={handleEditListing}
							onArchive={handleArchiveListing}
							onUnarchive={handleUnarchiveListing}
							onVerify={openVerificationModal}
							onShowFeedback={openFeedbackModal}
						/>
					{/each}
				</div>
			{/if}
		{:else}
			<EmptyState
				icon={PencilLine}
				title="Нет объявлений"
				description="Создайте своё первое объявление"
				actionLabel="Создать объявление"
				actionHref={resolve('/listings/new')}
			/>
		{/if}
	</PageShell>

	{#if selectedListingForVerification}
		<ListingVerificationModal
			open={verificationModalOpen}
			listing={{
				...selectedListingForVerification,
				host_id: selectedListingForVerification.ownerId,
				verification_video_url: selectedListingForVerification.verificationVideoUrl,
				verification_video_id: selectedListingForVerification.verificationVideoId,
				status: (selectedListingForVerification.status as any) || 'draft',
				type: selectedListingForVerification.propertyType === 'estate' ? 'manor' : selectedListingForVerification.propertyType,
				name: selectedListingForVerification.title,
				square: selectedListingForVerification.area,
				floor: selectedListingForVerification.floor || 1,
				total_floors: selectedListingForVerification.totalFloors,
				max_guests: selectedListingForVerification.maxGuests,
				rooms_count: selectedListingForVerification.bedrooms,
				beds_count: selectedListingForVerification.beds,
				bathrooms_count: selectedListingForVerification.bathrooms,
				price_per_night: selectedListingForVerification.pricePerNight,
				currency: selectedListingForVerification.currency,
				min_nights: selectedListingForVerification.rules?.minNights || 1,
				checkin_from: selectedListingForVerification.rules?.checkIn || '14:00',
				checkout_until: selectedListingForVerification.rules?.checkOut || '12:00',
				media: selectedListingForVerification.images,
				amenities: selectedListingForVerification.amenities,
				rules: selectedListingForVerification.rules as any,
				description: selectedListingForVerification.description,
				created_at: selectedListingForVerification.createdAt,
				updated_at: selectedListingForVerification.updatedAt
			}}
			onClose={closeVerificationModal}
			onUpdated={async () => {
				await hostListingsStore.refresh();
				await loadHostVerificationStatus();
			}}
		/>
	{/if}

	{#if selectedListingForFeedback}
		<ListingModerationModal
			open={feedbackModalOpen}
			listing={selectedListingForFeedback}
			onClose={closeFeedbackModal}
			onEdit={(listing) => {
				const target = listing || selectedListingForFeedback;
				closeFeedbackModal();
				if (target?.id) {
					goto(`/host/listings/${target.id}`);
				}
			}}
		/>
	{/if}
{/if}
