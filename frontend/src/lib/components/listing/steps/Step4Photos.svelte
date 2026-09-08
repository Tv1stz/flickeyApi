<!-- Step4Photos.svelte -->
<script lang="ts">
	import { Upload, Star, Trash2, Camera, AlertCircle, Plus } from 'lucide-svelte';
	import { flip } from 'svelte/animate';
	import { fade } from 'svelte/transition';
	import { dndzone } from 'svelte-dnd-action';
	import { cardStyles } from '$lib/config/styles';
	import Button from '$lib/components/ui/Button.svelte';
	import type { ListingErrors, ListingTouched } from '$lib/validation/listingValidation';

	interface PhotoItem {
		id: string;
		mediaId?: string;
		file?: File;
		url: string;
		isCover: boolean;
	}

	interface Props {
		errors: ListingErrors;
		touched: ListingTouched;
		photoItems?: PhotoItem[];
		maxPhotos?: number;
		onAddFiles: (files: FileList | File[]) => void;
		onRemovePhoto: (id: string) => void;
		onClearAll: () => void;
		onSetCover: (id: string) => void;
		onReorder: (items: PhotoItem[]) => void;
	}

	let {
		errors,
		touched,
		photoItems = [],
		maxPhotos = 25,
		onAddFiles,
		onRemovePhoto,
		onClearAll,
		onSetCover,
		onReorder
	}: Props = $props();

	let isOver = $state(false);
	let fileInput: HTMLInputElement | undefined = $state();

	const hasPhotos = $derived(photoItems.length > 0);
	const hasError = $derived(touched.photos && Boolean(errors.photos));
	const canAddMore = $derived(photoItems.length < maxPhotos);
	const minPhotos = 5;
	const needMore = $derived(Math.max(0, minPhotos - photoItems.length));
	const coverPhoto = $derived(photoItems[0] || null);
	let localGridPhotos = $state<PhotoItem[]>([]);

	$effect(() => {
		localGridPhotos = photoItems.slice(1);
	});

	function handleConsider(e: CustomEvent<{ items: PhotoItem[] }>) {
		localGridPhotos = e.detail.items;
	}

	function handleFinalize(e: CustomEvent<{ items: PhotoItem[] }>) {
		localGridPhotos = e.detail.items;
		if (coverPhoto) {
			const updated = [coverPhoto, ...e.detail.items].map((item, i) => ({
				...item,
				isCover: i === 0
			}));
			onReorder(updated);
		}
	}

	function prevent(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
	}

	function handleDrop(e: DragEvent) {
		prevent(e);
		isOver = false;
		if (e.dataTransfer?.files?.length) {
			onAddFiles(e.dataTransfer.files);
		}
	}

	function handleChange(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		if (input.files?.length) {
			onAddFiles(input.files);
			input.value = '';
		}
	}
</script>

<section class={cardStyles.section}>
	<h2 class={cardStyles.title}>Добавьте фото вашего жилья</h2>
	<p class={cardStyles.subtitle}>
		Первое фото станет обложкой. Перетащите фото, чтобы изменить порядок.
	</p>

	<div class={cardStyles.contentGap}>
		<input
			bind:this={fileInput}
			type="file"
			accept="image/*"
			multiple
			class="sr-only"
			onchange={handleChange}
		/>

		<!-- ── EMPTY STATE ── -->
		{#if !hasPhotos}
			<div
				role="button"
				tabindex="0"
				id="photos"
				onclick={() => fileInput?.click()}
				onkeydown={(e) => e.key === 'Enter' && fileInput?.click()}
				ondragenter={(e) => {
					prevent(e);
					isOver = true;
				}}
				ondragover={(e) => {
					prevent(e);
					isOver = true;
				}}
				ondragleave={(e) => {
					prevent(e);
					isOver = false;
				}}
				ondrop={handleDrop}
				class="flex w-full cursor-pointer flex-col items-center justify-center
                        rounded-2xl border-2 border-dashed px-6 py-16 text-center
                        transition-all duration-200 sm:py-24
                        {isOver
					? 'border-zinc-900 bg-zinc-100'
					: hasError
						? 'border-red-300 bg-red-50'
						: 'border-zinc-200 bg-zinc-50 hover:border-zinc-300 hover:bg-zinc-100'}"
			>
				<div
					class="mb-5 flex h-16 w-16 items-center justify-center rounded-2xl transition-colors duration-200
                            {isOver ? 'bg-zinc-900' : 'bg-zinc-200'}"
				>
					{#if isOver}
						<Upload class="h-8 w-8 text-white" />
					{:else}
						<Camera class="h-8 w-8 text-zinc-500" />
					{/if}
				</div>

				<p class="mb-1 text-base font-semibold text-zinc-900">
					{isOver ? 'Отпустите для загрузки' : 'Перетащите фото сюда'}
				</p>
				<p class="mb-6 text-sm text-zinc-400">или выберите с устройства</p>

				<Button variant="solid" tone="primary" size="md" radius="xl">Выбрать фото</Button>

				<p class="mt-5 text-xs text-zinc-400">Минимум 5 фото · JPG, PNG, WebP · до 15 МБ</p>
			</div>

			<!-- ── GALLERY STATE ── -->
		{:else}
			<div
				role="region"
				aria-label="Галерея фотографий"
				ondragenter={(e) => {
					prevent(e);
					isOver = true;
				}}
				ondragover={(e) => {
					prevent(e);
					isOver = true;
				}}
				ondragleave={(e) => {
					prevent(e);
					isOver = false;
				}}
				ondrop={handleDrop}
			>
				<!-- Обложка -->
				{#if coverPhoto}
					<div class="group relative overflow-hidden rounded-2xl bg-zinc-100">
						<div class="aspect-[16/9] sm:aspect-[21/9]">
							<img
								src={coverPhoto.url}
								alt="Обложка"
								class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.01]"
								draggable="false"
							/>
						</div>

						<div
							class="absolute inset-0 bg-black/0 transition-colors duration-200 group-hover:bg-black/20"
						></div>

						<div
							class="absolute inset-x-0 bottom-0 flex items-center justify-between
                                    gap-2 p-3 sm:p-4
                                    sm:opacity-0 sm:transition-opacity sm:duration-150 sm:group-hover:opacity-100"
						>
							<span
								class="inline-flex items-center gap-1.5 rounded-lg bg-white/90
                                    px-2.5 py-1.5 text-xs font-semibold text-zinc-900 shadow-sm backdrop-blur-sm"
							>
								<Star class="h-3.5 w-3.5 fill-zinc-900" />
								Обложка
							</span>

							<button
								type="button"
								onclick={() => onRemovePhoto(coverPhoto.id)}
								class="flex h-7 w-7 items-center justify-center rounded-lg
                                        bg-black/50 text-white backdrop-blur-sm
                                        transition-colors hover:bg-red-500 active:scale-95
                                        sm:bg-white/90 sm:text-zinc-700
                                        sm:hover:bg-red-50 sm:hover:text-red-600"
								aria-label="Удалить обложку"
							>
								<Trash2 class="h-3.5 w-3.5" />
							</button>
						</div>

						{#if isOver}
							<div
								class="absolute inset-0 flex items-center justify-center rounded-2xl bg-zinc-900/60"
								transition:fade={{ duration: 150 }}
							>
								<div class="text-center text-white">
									<Upload class="mx-auto mb-2 h-10 w-10" />
									<p class="text-base font-semibold">Отпустите для загрузки</p>
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Сетка остальных фото -->
				{#if localGridPhotos.length > 0}
					<div
						class="mt-2 grid grid-cols-3 gap-2 sm:grid-cols-4 sm:gap-2.5 md:grid-cols-5"
						use:dndzone={{ items: localGridPhotos, flipDurationMs: 200, dropTargetStyle: {} }}
						onconsider={handleConsider}
						onfinalize={handleFinalize}
					>
						{#each localGridPhotos as photo (photo.id)}
							<div
								class="group relative aspect-square cursor-grab overflow-hidden
                                        rounded-xl bg-zinc-100 active:cursor-grabbing"
								animate:flip={{ duration: 200 }}
							>
								<img
									src={photo.url}
									alt="Фото"
									class="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105"
									draggable="false"
								/>

								<div
									class="absolute inset-0 bg-black/0 transition-colors duration-150
                                            group-hover:bg-black/40"
								></div>

								<div
									class="absolute inset-x-0 bottom-0 flex items-center justify-between
                                            gap-1 p-1.5
                                            sm:inset-0 sm:flex-col sm:items-stretch sm:justify-between sm:p-1.5
                                            sm:opacity-0 sm:transition-opacity sm:duration-150 sm:group-hover:opacity-100"
								>
									<button
										type="button"
										onclick={(e) => {
											e.stopPropagation();
											onSetCover(photo.id);
										}}
										class="hidden h-7 w-full items-center justify-center gap-1
                                                rounded-lg bg-white/90 text-xs font-semibold
                                                text-zinc-800 shadow-sm backdrop-blur-sm
                                                transition-colors hover:bg-white active:scale-95
                                                sm:flex"
										aria-label="Сделать обложкой"
									>
										<Star class="h-3.5 w-3.5 shrink-0" />
										Обложка
									</button>

									<div class="flex w-full items-center justify-between gap-1 sm:justify-end">
										<button
											type="button"
											onclick={(e) => {
												e.stopPropagation();
												onSetCover(photo.id);
											}}
											class="flex h-7 w-7 items-center justify-center rounded-lg
                                                    bg-black/50 text-white backdrop-blur-sm
                                                    transition-colors hover:bg-black/70 active:scale-95
                                                    sm:hidden"
											aria-label="Сделать обложкой"
										>
											<Star class="h-3.5 w-3.5" />
										</button>

										<button
											type="button"
											onclick={(e) => {
												e.stopPropagation();
												onRemovePhoto(photo.id);
											}}
											class="flex h-7 w-7 items-center justify-center rounded-lg
                                                    bg-black/50 text-white backdrop-blur-sm
                                                    transition-colors hover:bg-red-500 active:scale-95
                                                    sm:bg-white/90 sm:text-zinc-700
                                                    sm:hover:bg-red-50 sm:hover:text-red-600"
											aria-label="Удалить фото"
										>
											<Trash2 class="h-3.5 w-3.5" />
										</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}

				<!-- Футер галереи -->
				<div class="mt-4 flex items-center justify-between gap-3 border-t border-zinc-100 pt-4">
					<p class="text-sm text-zinc-500">
						{#if needMore > 0}
							<span class="font-medium text-zinc-900">{photoItems.length}</span> из {minPhotos} минимум
							<span class="mx-1 text-zinc-300">·</span>
							<span class="text-amber-600">ещё {needMore}</span>
						{:else}
							<span class="font-medium text-zinc-900">{photoItems.length}</span> из {maxPhotos} фото
						{/if}
					</p>

					<div class="flex items-center gap-2">
						<Button variant="ghost" tone="danger" size="sm" radius="lg" onclick={onClearAll}>
							Удалить все
						</Button>

						{#if canAddMore}
							<Button
								variant="solid"
								tone="primary"
								size="sm"
								radius="lg"
								iconLeft={Plus}
								onclick={() => fileInput?.click()}
							>
								Добавить
							</Button>
						{/if}
					</div>
				</div>
			</div>
		{/if}

		<!-- Ошибка -->
		{#if hasError}
			<div
				class="mt-4 flex items-start gap-3 rounded-xl border border-red-100 bg-red-50 p-3.5"
				in:fade={{ duration: 200 }}
			>
				<AlertCircle class="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
				<p class="text-sm text-red-600">{errors.photos}</p>
			</div>
		{/if}
	</div>
</section>
