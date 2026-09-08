<!-- src/lib/components/verification/ListingVerificationModal.svelte -->
<script lang="ts">
	import { X, Video } from 'lucide-svelte';
	import type { ListingHost } from '$lib/types/listings';
	import { listingsApi } from '$lib/api/listings';
	import { toast } from '$lib/stores/toastStore';
	import VerificationVideoUploader from './VerificationVideoUploader.svelte';

	interface Props {
		open: boolean;
		listing: ListingHost;
		onClose: () => void;
		onUpdated?: (listing: ListingHost) => void;
	}

	let { open, listing, onClose, onUpdated }: Props = $props();

	let localVideoUrl = $state<string | null>(null);
	let localVideoId = $state<string | null>(null);
	let localStatus = $state<string | null>(null);

	const activeListing = $derived<ListingHost>({
		...listing,
		status: (localStatus || listing.status) as any,
		verification_video_url: localVideoUrl || listing.verification_video_url,
		verification_video_id: localVideoId || listing.verification_video_id
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open) {
			onClose();
		}
	}

	async function handleVideoAttached(mediaId: string, url: string) {
		try {
			const updated = await listingsApi.attachVerificationVideo(activeListing.id, mediaId);
			localStatus = updated.status;
			localVideoUrl = url;
			localVideoId = mediaId;
			onUpdated?.(activeListing);

			if (updated.status === 'pending_review') {
				toast.success('Объявление отправлено на модерацию!', 'Видео прикреплено, реквизиты подтверждены.');
			} else if (updated.status === 'awaiting_company_verification') {
				toast.info('Видео прикреплено', 'Ожидается подтверждение документов партнера.');
			} else {
				toast.info('Видео прикреплено', 'Для публикации подтвердите реквизиты партнера в панели управления.');
			}
		} catch (err: any) {
			console.error('Ошибка привязки видео:', err);
			toast.error('Ошибка привязки видео', err?.message);
		}
	}

	const hasVideo = $derived(Boolean(activeListing.verification_video_url || activeListing.verification_video_id));
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6"
		role="dialog"
		aria-modal="true"
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity duration-200"
			onclick={onClose}
		></div>

		<!-- Dialog Container -->
		<div
			class="relative w-full max-w-xl max-h-[90vh] flex flex-col rounded-3xl bg-white shadow-2xl overflow-hidden z-10 dark:bg-card border border-zinc-200/80 dark:border-border animate-in fade-in zoom-in-95 duration-200"
		>
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-zinc-100 dark:border-border">
				<div class="flex items-center gap-3 min-w-0">
					<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-amber-50 text-amber-600 dark:bg-amber-950 dark:text-amber-300">
						<Video class="h-5 w-5" />
					</div>
					<div class="min-w-0">
						<h2 class="text-base font-bold text-zinc-900 truncate dark:text-foreground">
							Видео верификации жилья
						</h2>
						<p class="text-xs text-zinc-500 dark:text-muted-foreground truncate max-w-xs sm:max-w-md mt-0.5">
							{activeListing.name}
						</p>
					</div>
				</div>
				<button
					type="button"
					onclick={onClose}
					class="rounded-full p-2 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-600 transition dark:hover:bg-muted dark:hover:text-foreground cursor-pointer"
					aria-label="Закрыть"
				>
					<X class="h-5 w-5" />
				</button>
			</div>

			<!-- Body (scrollable) -->
			<div class="flex-1 overflow-y-auto p-6 space-y-5">
				<!-- Explanation notice -->
				<div class="rounded-2xl border border-zinc-200/80 bg-zinc-50/70 p-4 text-xs dark:border-border dark:bg-muted/30">
					<p class="font-semibold text-zinc-900 dark:text-foreground">
						Подтверждение владения и реального состояния
					</p>
					<p class="text-zinc-600 dark:text-muted-foreground mt-1 leading-relaxed">
						Прикрепите короткое видео-обход квартиры (до 100 МБ). Видео строго конфиденциально и доступно <strong>только вам и модератору</strong> платформы.
					</p>
				</div>

				<!-- Video Uploader -->
				<VerificationVideoUploader
					listingId={activeListing.id}
					videoUrl={activeListing.verification_video_url}
					onVideoAttached={handleVideoAttached}
				/>
			</div>

			<!-- Footer -->
			<div class="flex items-center justify-end gap-3 px-6 py-4 bg-zinc-50 border-t border-zinc-100 dark:bg-muted/30 dark:border-border">
				<button
					type="button"
					onclick={onClose}
					class="rounded-full bg-zinc-900 px-5 py-2 text-xs font-semibold text-white hover:bg-zinc-800 transition active:scale-95 cursor-pointer dark:bg-primary dark:text-primary-foreground"
				>
					{hasVideo ? 'Готово' : 'Закрыть'}
				</button>
			</div>
		</div>
	</div>
{/if}
