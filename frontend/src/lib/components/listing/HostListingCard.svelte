<!-- src/lib/components/listing/HostListingCard.svelte -->
<script lang="ts">
	import { Trash2, Pencil, Archive, ArchiveRestore, ShieldCheck, Video, AlertTriangle } from 'lucide-svelte';
	import type { Listing } from '$lib/components/card/types';
	import { PROPERTY_TYPE_LABELS } from '$lib/components/card/types';
	import { formatBYN } from '$lib/utils/format';
	import { formatListingAddressFull } from '$lib/utils/location';
	import IconActionButton from '$lib/components/ui/IconActionButton.svelte';
	import StatusIndicator from '$lib/components/ui/StatusIndicator.svelte';

	type Variant = 'list' | 'grid';

	interface Props {
		listing: Listing;
		variant?: Variant;
		onOpen: (id: string) => void;
		onDelete: (e: MouseEvent, listing: Listing) => void;
		onEdit?: (e: MouseEvent, listing: Listing) => void;
		onArchive?: (e: MouseEvent, listing: Listing) => void;
		onUnarchive?: (e: MouseEvent, listing: Listing) => void;
		onVerify?: (e: MouseEvent, listing: Listing) => void;
		onShowFeedback?: (e: MouseEvent, listing: Listing) => void;
	}

	let {
		listing,
		variant = 'grid',
		onOpen,
		onDelete,
		onEdit,
		onArchive,
		onUnarchive,
		onVerify,
		onShowFeedback
	}: Props = $props();

	const fullAddress = $derived(formatListingAddressFull(listing.address, listing.location));
	const hasVideo = $derived(Boolean(listing.verificationVideoUrl || listing.verificationVideoId));
	const isDraftNeedsVideo = $derived(
		(listing.status === 'draft' || listing.status === 'draft_video_required') && !hasVideo
	);
	const effectiveStatus = $derived(isDraftNeedsVideo ? 'draft_video_required' : listing.status);
	const hasFeedback = $derived(
		listing.status === 'changes_requested' || listing.status === 'rejected' || listing.status === 'suspended'
	);
	const feedbackLabel = $derived.by(() => {
		if (listing.status === 'changes_requested') return 'Замечания модератора';
		if (listing.status === 'rejected') return 'Причина отклонения';
		if (listing.status === 'suspended') return 'Причина блокировки';
		return 'Замечания модератора';
	});

	function handleCardKeydown(e: KeyboardEvent): void {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onOpen(listing.id);
		}
	}
</script>

{#if variant === 'list'}
	<div
		class="group cursor-pointer transition-colors duration-150 hover:bg-zinc-50"
		role="button"
		tabindex="0"
		onclick={() => onOpen(listing.id)}
		onkeydown={handleCardKeydown}
	>
		<div class="flex items-center gap-4 px-4 py-3 sm:hidden">
			<div class="relative h-20 w-24 shrink-0 overflow-hidden rounded-xl bg-zinc-100">
				{#if listing.images?.[0]}
					<img
						src={listing.images[0]}
						alt={listing.title}
						class="h-full w-full object-cover"
						loading="lazy"
						decoding="async"
					/>
				{/if}
			</div>
			<div class="min-w-0 flex-1">
				<p class="truncate text-sm font-semibold text-zinc-900">{listing.title}</p>
				<p class="mt-0.5 truncate text-xs text-zinc-500">
					{PROPERTY_TYPE_LABELS[listing.propertyType]} · {fullAddress}
				</p>
				<div class="mt-1.5 flex flex-wrap items-center gap-2">
					<StatusIndicator status={effectiveStatus} active={listing.isActive} />
					{#if hasFeedback && onShowFeedback}
						<button
							type="button"
							class="inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-xs font-semibold transition-colors cursor-pointer
							       {listing.status === 'changes_requested'
								? 'border-amber-300 bg-amber-50 text-amber-800 hover:bg-amber-100'
								: 'border-rose-300 bg-rose-50 text-rose-800 hover:bg-rose-100'}"
							onclick={(e) => {
								e.stopPropagation();
								onShowFeedback?.(e, listing);
							}}
						>
							<AlertTriangle size={11} class={listing.status === 'changes_requested' ? 'text-amber-600' : 'text-rose-600'} />
							<span>{listing.status === 'changes_requested' ? 'Замечания' : 'Причина'}</span>
						</button>
					{/if}
					{#if (listing.status === 'draft' || listing.status === 'draft_video_required' || !hasVideo) && onVerify}
						<button
							type="button"
							class="inline-flex items-center gap-1 rounded-md border border-rose-200 bg-rose-50 px-2 py-0.5 text-xs font-semibold text-rose-700 hover:bg-rose-100 dark:bg-rose-950/40 dark:border-rose-900 dark:text-rose-300 transition-colors cursor-pointer"
							onclick={(e) => {
								e.stopPropagation();
								onVerify?.(e, listing);
							}}
						>
							<Video size={12} class="text-rose-600 dark:text-rose-400" />
							<span>{hasVideo ? 'Видео прикреплено' : 'Загрузить видео'}</span>
						</button>
					{/if}
				</div>
				<p class="mt-1 text-sm font-bold text-zinc-900">
					{formatBYN(listing.pricePerNight)}<span class="text-xs font-normal text-zinc-400">
						/ ночь</span
					>
				</p>
			</div>
			<div class="flex shrink-0 flex-col gap-1.5">
				{#if onEdit}
					<IconActionButton
						icon={Pencil}
						ariaLabel="Редактировать объявление"
						onclick={(e) => onEdit?.(e, listing)}
					/>
				{/if}
				{#if (listing.status === 'archived' || listing.status === 'draft') && onUnarchive}
					<IconActionButton
						icon={ArchiveRestore}
						ariaLabel={listing.status === 'draft' ? 'Отправить на модерацию' : 'Опубликовать (разархивировать)'}
						tooltip={listing.status === 'draft' ? 'Отправить на модерацию' : 'Опубликовать'}
						onclick={(e) => onUnarchive?.(e, listing)}
					/>
				{:else if listing.status !== 'draft' && onArchive}
					<IconActionButton
						icon={Archive}
						ariaLabel="В архив"
						onclick={(e) => onArchive?.(e, listing)}
					/>
				{/if}
				<IconActionButton
					icon={Trash2}
					tone="danger"
					ariaLabel="Удалить"
					onclick={(e) => onDelete(e, listing)}
				/>
			</div>
		</div>

		<div
			class="hidden items-center gap-4 px-6 py-4 sm:grid"
			style="grid-template-columns: 1fr 110px 160px 240px 140px;"
		>
			<div class="flex min-w-0 items-center gap-4">
				<div class="relative h-[72px] w-[96px] shrink-0 overflow-hidden rounded-2xl bg-zinc-100">
					{#if listing.images?.[0]}
						<img
							src={listing.images[0]}
							alt={listing.title}
							class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-[1.05]"
							loading="lazy"
							decoding="async"
						/>
					{/if}
				</div>
				<div class="min-w-0">
					<p class="truncate text-[15px] font-semibold text-zinc-900">{listing.title}</p>
					<p class="mt-1 text-sm font-semibold text-zinc-900">
						{formatBYN(listing.pricePerNight)}
						<span class="text-xs font-normal text-zinc-400"> / ночь</span>
					</p>
				</div>
			</div>

			<p class="truncate text-sm text-zinc-600">{PROPERTY_TYPE_LABELS[listing.propertyType]}</p>

			<p class="truncate text-sm text-zinc-600">
				{fullAddress}
			</p>

			<div class="flex flex-col items-start gap-1.5 min-w-0">
				<StatusIndicator status={effectiveStatus} active={listing.isActive} size="sm" />
				{#if hasFeedback && onShowFeedback}
					<button
						type="button"
						class="inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs font-semibold transition-colors cursor-pointer shadow-2xs
						       {listing.status === 'changes_requested'
							? 'border-amber-300 bg-amber-50 text-amber-800 hover:bg-amber-100 hover:border-amber-400'
							: 'border-rose-300 bg-rose-50 text-rose-800 hover:bg-rose-100 hover:border-rose-400'}"
						onclick={(e) => {
							e.stopPropagation();
							onShowFeedback?.(e, listing);
						}}
						title="Посмотреть решение модерации"
					>
						<AlertTriangle size={13} class={listing.status === 'changes_requested' ? 'text-amber-600 shrink-0' : 'text-rose-600 shrink-0'} />
						<span>{feedbackLabel}</span>
					</button>
				{/if}
				{#if (listing.status === 'draft' || listing.status === 'draft_video_required' || !hasVideo) && onVerify}
					<button
						type="button"
						class="inline-flex items-center gap-1.5 rounded-lg border border-rose-200 bg-rose-50 px-2.5 py-1 text-xs font-semibold text-rose-700 hover:bg-rose-100 hover:border-rose-300 dark:bg-rose-950/40 dark:border-rose-900/60 dark:text-rose-300 dark:hover:bg-rose-950/70 transition-colors cursor-pointer shadow-2xs"
						onclick={(e) => {
							e.stopPropagation();
							onVerify?.(e, listing);
						}}
						title="Видео верификации квартиры"
					>
						<Video size={13} class="text-rose-600 dark:text-rose-400 shrink-0" />
						<span>{hasVideo ? 'Видео прикреплено' : 'Загрузить видео'}</span>
					</button>
				{/if}
			</div>

			<div class="flex items-center justify-end gap-1.5">
				{#if onEdit}
					<IconActionButton
						icon={Pencil}
						ariaLabel="Редактировать"
						tooltip="Редактировать объявление"
						elevated={true}
						onclick={(e) => onEdit?.(e, listing)}
					/>
				{/if}
				{#if (listing.status === 'archived' || listing.status === 'draft') && onUnarchive}
					<IconActionButton
						icon={ArchiveRestore}
						ariaLabel={listing.status === 'draft' ? 'Отправить на модерацию' : 'Опубликовать'}
						tooltip={listing.status === 'draft' ? 'Отправить на модерацию' : 'Опубликовать из архива'}
						elevated={true}
						onclick={(e) => onUnarchive?.(e, listing)}
					/>
				{:else if (listing.status === 'published' || listing.status === 'active') && onArchive}
					<IconActionButton
						icon={Archive}
						ariaLabel="В архив"
						tooltip="Снять с публикации (в архив)"
						elevated={true}
						onclick={(e) => onArchive?.(e, listing)}
					/>
				{/if}
				<IconActionButton
					icon={Trash2}
					tone="danger"
					ariaLabel="Удалить"
					tooltip="Удалить объявление"
					tooltipAlign="right"
					elevated={true}
					onclick={(e) => onDelete(e, listing)}
				/>
			</div>
		</div>
	</div>
{:else}
	<div
		class="group cursor-pointer"
		role="button"
		tabindex="0"
		onclick={() => onOpen(listing.id)}
		onkeydown={handleCardKeydown}
	>
		<div class="relative aspect-[4/3] overflow-hidden rounded-3xl bg-zinc-100">
			{#if listing.images?.[0]}
				<img
					src={listing.images[0]}
					alt={listing.title}
					class="h-full w-full object-cover transition-transform duration-500 ease-out group-hover:scale-[1.04]"
					loading="lazy"
					decoding="async"
				/>
			{/if}

			<div
				class="absolute inset-0 bg-black/0 transition-colors duration-300 group-hover:bg-black/10"
			></div>

			<div class="absolute top-3 left-3">
				<StatusIndicator status={effectiveStatus} active={listing.isActive} variant="pill" />
			</div>

			<div class="absolute right-3 bottom-3 flex items-center gap-1.5">
				{#if onEdit}
					<IconActionButton
						icon={Pencil}
						variant="floating"
						ariaLabel="Редактировать"
						tooltip="Редактировать объявление"
						onclick={(e) => onEdit?.(e, listing)}
					/>
				{/if}
				{#if (listing.status === 'archived' || listing.status === 'draft') && onUnarchive}
					<IconActionButton
						icon={ArchiveRestore}
						variant="floating"
						ariaLabel={listing.status === 'draft' ? 'Отправить на модерацию' : 'Опубликовать'}
						tooltip={listing.status === 'draft' ? 'Отправить на модерацию' : 'Опубликовать из архива'}
						onclick={(e) => onUnarchive?.(e, listing)}
					/>
				{:else if (listing.status === 'published' || listing.status === 'active') && onArchive}
					<IconActionButton
						icon={Archive}
						variant="floating"
						ariaLabel="В архив"
						tooltip="Снять с публикации (в архив)"
						onclick={(e) => onArchive?.(e, listing)}
					/>
				{/if}
				<IconActionButton
					icon={Trash2}
					variant="floating"
					tone="danger"
					ariaLabel="Удалить"
					tooltip="Удалить объявление"
					tooltipAlign="right"
					onclick={(e) => onDelete(e, listing)}
				/>
			</div>
		</div>

		<div class="mt-3">
			<div class="flex items-start justify-between gap-2">
				<h3 class="line-clamp-1 text-[15px] font-semibold text-zinc-900">
					{listing.title}
				</h3>
			</div>
			<p class="mt-0.5 line-clamp-1 text-sm text-zinc-500">
				{fullAddress}
			</p>
			<p class="mt-0.5 text-sm text-zinc-400">{PROPERTY_TYPE_LABELS[listing.propertyType]}</p>
			<p class="mt-2 text-[15px] font-bold text-zinc-900">
				{formatBYN(listing.pricePerNight)}
				<span class="text-xs font-normal text-zinc-400"> / ночь</span>
			</p>

			{#if hasFeedback && onShowFeedback}
				<div class="mt-3">
					<button
						type="button"
						class="inline-flex w-full items-center justify-center gap-2 rounded-2xl border px-3.5 py-2.5 text-xs font-bold transition hover:shadow-xs active:scale-[0.98] cursor-pointer
						       {listing.status === 'changes_requested'
							? 'border-amber-300 bg-amber-50 text-amber-800 hover:bg-amber-100 hover:border-amber-400'
							: 'border-rose-300 bg-rose-50 text-rose-800 hover:bg-rose-100 hover:border-rose-400'}"
						onclick={(e) => {
							e.stopPropagation();
							onShowFeedback?.(e, listing);
						}}
					>
						<AlertTriangle size={14} class={listing.status === 'changes_requested' ? 'text-amber-600 shrink-0' : 'text-rose-600 shrink-0'} />
						<span>{feedbackLabel}</span>
					</button>
				</div>
			{/if}

			{#if (listing.status === 'draft' || listing.status === 'draft_video_required' || !hasVideo) && onVerify}
				<div class="mt-3">
					<button
						type="button"
						class="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-rose-200 bg-rose-50 px-3.5 py-2.5 text-xs font-bold text-rose-700 transition hover:bg-rose-100 hover:border-rose-300 active:scale-[0.98] dark:bg-rose-950/40 dark:border-rose-900/60 dark:text-rose-300 dark:hover:bg-rose-950/70 cursor-pointer shadow-2xs"
						onclick={(e) => {
							e.stopPropagation();
							onVerify?.(e, listing);
						}}
					>
						<Video size={14} class="text-rose-600 dark:text-rose-400 shrink-0" />
						<span>{hasVideo ? 'Видео квартиры прикреплено' : 'Загрузить видео квартиры'}</span>
					</button>
				</div>
			{/if}
		</div>
	</div>
{/if}
