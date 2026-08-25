<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { listingsApi } from '$lib/api/listings';
	import type { ListingPublic, DraftDetail } from '$lib/types/listings';
	import ListingView from '$lib/components/listings/ListingView.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import Alert from '$lib/components/ui/Alert.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { auth } from '$lib/auth/auth.svelte';
	import { toast } from '$lib/components/ui/toast.svelte';
	import { ArrowLeft, Sparkles } from 'lucide-svelte';

	let draftId = $derived(page.url.searchParams.get('draft_id') || '');
	let draft = $state<DraftDetail | null>(null);
	let loading = $state(true);
	let errorMessage = $state<string | null>(null);
	let publishing = $state(false);

	async function loadDraftData() {
		if (!draftId) {
			loading = false;
			errorMessage = 'ID черновика не указан';
			return;
		}

		loading = true;
		errorMessage = null;
		try {
			draft = await listingsApi.getDraft(draftId);
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось загрузить черновик';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadDraftData();
	});

	// Transform DraftDetail into ListingPublic for ListingView
	let previewListing = $derived<ListingPublic | null>(
		draft
			? {
					id: draft.id,
					type: draft.type || 'apartment',
					name: draft.name || 'Без названия',
					square: draft.square || 0,
					floor: draft.floor || 1,
					total_floors: draft.total_floors || 1,
					max_guests: draft.max_guests || 1,
					rooms_count: draft.rooms_count || 1,
					beds_count: draft.beds_count || 1,
					bathrooms_count: draft.bathrooms_count || 1,
					price_per_night: draft.price_per_night || 0,
					currency: draft.currency || 'BYN',
					min_nights: draft.min_nights || 1,
					checkin_from: draft.checkin_from || '14:00',
					checkout_until: draft.checkout_until || '12:00',
					description: draft.description || '',
					amenities: draft.amenities || [],
					media: draft.media_ids || [],
					created_at: draft.created_at || new Date().toISOString()
				}
			: null
	);

	async function handlePublish() {
		if (!draftId) return;

		publishing = true;
		try {
			const idempotencyKey = crypto.randomUUID();
			await listingsApi.submitDraft(draftId, idempotencyKey);
			await auth.refreshUser();
			auth.switchContext('host');
			toast.success('Объявление успешно опубликовано!');
			window.location.href = '/host';
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Ошибка публикации');
		} finally {
			publishing = false;
		}
	}

	function handleEdit() {
		window.location.href = '/host/new';
	}
</script>

<svelte:head>
	<title>Предпросмотр черновика — Flickey</title>
</svelte:head>

<div class="container mx-auto px-4 sm:px-6 py-6 space-y-6 max-w-6xl">
	<!-- Top Bar -->
	<div class="flex items-center justify-between">
		<button
			type="button"
			onclick={handleEdit}
			class="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
		>
			<ArrowLeft class="h-4 w-4" /> Назад к редактированию
		</button>
		<div class="inline-flex items-center gap-1.5 rounded-full bg-amber-500/10 px-3 py-0.5 text-xs font-bold text-amber-700 dark:text-amber-400">
			<Sparkles class="h-3.5 w-3.5" /> Режим предпросмотра
		</div>
	</div>

	{#if loading}
		<div class="space-y-6">
			<Skeleton class="h-10 w-2/3" />
			<Skeleton class="aspect-[21/9] w-full rounded-2xl" />
			<div class="grid grid-cols-3 gap-8">
				<div class="col-span-2 space-y-4">
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-40 w-full" />
				</div>
				<div class="col-span-1">
					<Skeleton class="h-64 w-full" />
				</div>
			</div>
		</div>
	{:else if errorMessage}
		<Alert variant="destructive">{errorMessage}</Alert>
	{:else if previewListing}
		<ListingView
			listing={previewListing}
			mode="preview"
			onpublish={handlePublish}
			onedit={handleEdit}
		/>
	{/if}
</div>
