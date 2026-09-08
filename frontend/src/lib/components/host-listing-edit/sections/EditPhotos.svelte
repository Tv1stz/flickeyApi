<!-- src/lib/components/host-listing-edit/sections/EditPhotos.svelte -->
<script lang="ts">
	import { cardStyles } from '$lib/config/styles';
	import { mediaApi } from '$lib/api/media';
	import { toast } from '$lib/stores/toastStore';
	import { Upload, Star, Trash2, Camera, Loader2, Plus } from 'lucide-svelte';
	import type { ListingEditFormData } from '../types';

	interface Props {
		form: Partial<ListingEditFormData>;
		onUpdate: (updates: Partial<ListingEditFormData>) => void;
	}

	let { form, onUpdate }: Props = $props();

	let fileInput: HTMLInputElement | undefined = $state();
	let isUploading = $state(false);

	const images = $derived(form.images || []);

	async function handleFilesSelected(e: Event) {
		const target = e.currentTarget as HTMLInputElement;
		if (!target.files?.length) return;

		const files = Array.from(target.files);
		target.value = '';

		isUploading = true;
		const newUrls: string[] = [];

		for (const file of files) {
			try {
				const res = await mediaApi.uploadFile(file);
				const url = `/api/v1/media/dev-upload/${res.fileKey}`;
				newUrls.push(url);
			} catch (err: any) {
				console.error('Failed to upload photo:', err);
				toast.error('Ошибка загрузки', err?.message || `Не удалось загрузить ${file.name}`);
			}
		}

		if (newUrls.length > 0) {
			onUpdate({ images: [...images, ...newUrls] });
			toast.success(`Загружено фото: ${newUrls.length}`);
		}
		isUploading = false;
	}

	function handleSetCover(index: number) {
		if (index === 0 || index >= images.length) return;
		const updated = [...images];
		const [selected] = updated.splice(index, 1);
		updated.unshift(selected);
		onUpdate({ images: updated });
		toast.info('Обложка обновлена');
	}

	function handleRemovePhoto(index: number) {
		const updated = images.filter((_, i) => i !== index);
		onUpdate({ images: updated });
	}
</script>

<section class={cardStyles.section}>
	<div class="flex items-center justify-between">
		<div>
			<h2 class={cardStyles.title}>Фотографии</h2>
			<p class={cardStyles.subtitle}>Добавьте качественные фото вашего жилья</p>
		</div>
		<button
			type="button"
			class="flex items-center gap-1.5 rounded-xl bg-zinc-900 px-4 py-2 text-xs font-semibold text-white transition-all hover:bg-zinc-800 active:scale-95 disabled:opacity-50"
			onclick={() => fileInput?.click()}
			disabled={isUploading}
		>
			{#if isUploading}
				<Loader2 class="h-4 w-4 animate-spin" />
				<span>Загрузка...</span>
			{:else}
				<Plus class="h-4 w-4" />
				<span>Добавить фото</span>
			{/if}
		</button>
	</div>

	<input
		bind:this={fileInput}
		type="file"
		accept="image/jpeg,image/png,image/webp"
		multiple
		class="hidden"
		onchange={handleFilesSelected}
	/>

	<div class={cardStyles.contentGap}>
		{#if images.length === 0}
			<!-- Empty state / Dropzone -->
			<div
				class="flex flex-col items-center justify-center rounded-2xl border-2 border-dashed border-zinc-300 p-10 text-center transition-colors hover:border-zinc-400"
			>
				<div class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-500">
					<Camera class="h-7 w-7" />
				</div>
				<h3 class="text-base font-semibold text-zinc-900">Загрузите фотографии</h3>
				<p class="mt-1 max-w-sm text-xs text-zinc-500">
					Первое фото станет обложкой в поиске. Поддерживаются форматы JPG, PNG и WEBP.
				</p>
				<button
					type="button"
					class="mt-5 flex items-center gap-2 rounded-xl bg-zinc-900 px-5 py-2.5 text-xs font-semibold text-white transition-all hover:bg-zinc-800 active:scale-95"
					onclick={() => fileInput?.click()}
					disabled={isUploading}
				>
					{#if isUploading}
						<Loader2 class="h-4 w-4 animate-spin" />
						<span>Загрузка...</span>
					{:else}
						<Upload class="h-4 w-4" />
						<span>Выбрать файлы</span>
					{/if}
				</button>
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4">
				{#each images as photo, idx (photo + idx)}
					<div
						class="group relative aspect-square overflow-hidden rounded-2xl border border-zinc-200 bg-zinc-100 shadow-sm"
					>
						<img src={photo} alt={`Фото ${idx + 1}`} class="h-full w-full object-cover" />

						{#if idx === 0}
							<div
								class="absolute top-2 left-2 flex items-center gap-1 rounded-md bg-zinc-900/80 px-2 py-1 text-[11px] font-semibold text-white backdrop-blur-sm"
							>
								<Star class="h-3 w-3 fill-amber-400 text-amber-400" />
								<span>Обложка</span>
							</div>
						{/if}

						<!-- Hover Actions -->
						<div
							class="absolute inset-0 flex items-center justify-center gap-2 bg-black/40 opacity-0 backdrop-blur-[2px] transition-opacity duration-150 group-hover:opacity-100"
						>
							{#if idx !== 0}
								<button
									type="button"
									class="flex h-8 items-center gap-1 rounded-lg bg-white px-2.5 text-[11px] font-medium text-zinc-800 shadow transition-transform hover:scale-105"
									onclick={() => handleSetCover(idx)}
									title="Сделать обложкой"
								>
									<Star class="h-3.5 w-3.5 text-amber-500" />
									<span>Обложка</span>
								</button>
							{/if}
							<button
								type="button"
								class="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-600 text-white shadow transition-transform hover:scale-105 hover:bg-rose-700"
								onclick={() => handleRemovePhoto(idx)}
								title="Удалить фото"
							>
								<Trash2 class="h-3.5 w-3.5" />
							</button>
						</div>
					</div>
				{/each}

				<!-- Add Card -->
				<button
					type="button"
					class="flex aspect-square flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed border-zinc-200 bg-zinc-50/50 p-4 text-zinc-500 transition-all hover:border-zinc-400 hover:bg-zinc-50 hover:text-zinc-800 active:scale-95"
					onclick={() => fileInput?.click()}
					disabled={isUploading}
				>
					{#if isUploading}
						<Loader2 class="h-6 w-6 animate-spin text-zinc-400" />
						<span class="text-xs font-medium">Загрузка...</span>
					{:else}
						<Plus class="h-6 w-6" />
						<span class="text-xs font-medium">Добавить фото</span>
					{/if}
				</button>
			</div>
		{/if}
	</div>
</section>
