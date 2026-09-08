<!-- src/lib/components/verification/VerificationVideoUploader.svelte -->
<script lang="ts">
	import { Video, UploadCloud, CheckCircle2, AlertCircle, Loader2, Lock, RotateCcw } from 'lucide-svelte';
	import { mediaApi } from '$lib/api/media';
	import { toast } from '$lib/stores/toastStore';

	interface Props {
		videoUrl?: string | null;
		listingId?: string;
		onVideoAttached: (mediaId: string, url: string) => void;
		onVideoRemoved?: () => void;
		disabled?: boolean;
	}

	let {
		videoUrl = null,
		listingId,
		onVideoAttached,
		onVideoRemoved,
		disabled = false
	}: Props = $props();

	const MAX_VIDEO_SIZE_BYTES = 100 * 1024 * 1024; // 100 MB

	let isDragging = $state(false);
	let isUploading = $state(false);
	let uploadProgress = $state(0);
	let errorMessage = $state<string | null>(null);
	let currentVideoUrl = $state<string | null>(null);

	$effect(() => {
		currentVideoUrl = videoUrl ?? null;
	});

	async function handleFileSelect(files: FileList | null) {
		if (!files || files.length === 0 || disabled || isUploading) return;
		const file = files[0];

		// Check file type
		const validTypes = ['video/mp4', 'video/quicktime', 'video/webm'];
		if (!validTypes.includes(file.type) && !file.name.match(/\.(mp4|mov|webm)$/i)) {
			errorMessage = 'Поддерживаются только форматы MP4, QuickTime (MOV) и WebM.';
			toast.error('Неверный формат', errorMessage);
			return;
		}

		// Check file size
		if (file.size > MAX_VIDEO_SIZE_BYTES) {
			const sizeMb = (file.size / (1024 * 1024)).toFixed(1);
			errorMessage = `Размер видео (${sizeMb} МБ) превышает лимит в 100 МБ. Пожалуйста, сожмите видео или снимите более короткий ролик.`;
			toast.error('Файл слишком большой', errorMessage);
			return;
		}

		errorMessage = null;
		isUploading = true;
		uploadProgress = 10;

		try {
			uploadProgress = 30;
			const uploaded = await mediaApi.uploadFile(file);
			uploadProgress = 85;

			const resolvedUrl = mediaApi.getMediaUrl(uploaded.fileKey);
			currentVideoUrl = resolvedUrl;
			uploadProgress = 100;

			toast.success('Видео загружено', 'Видео объекта успешно загружено и прикреплено.');
			onVideoAttached(uploaded.mediaId, resolvedUrl);
		} catch (err: any) {
			console.error('Ошибка загрузки видео:', err);
			errorMessage = err?.message || 'Не удалось загрузить видео. Попробуйте еще раз.';
			toast.error('Ошибка загрузки', errorMessage ?? undefined);
		} finally {
			isUploading = false;
			uploadProgress = 0;
		}
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
		if (e.dataTransfer?.files) {
			handleFileSelect(e.dataTransfer.files);
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		if (!disabled && !isUploading) isDragging = true;
	}

	function handleDragLeave() {
		isDragging = false;
	}
</script>

<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-xs dark:border-border dark:bg-card">
	<!-- Header -->
	<div class="flex items-start justify-between gap-4 mb-4">
		<div class="flex items-center gap-3">
			<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-400">
				<Video class="h-5 w-5" />
			</div>
			<div>
				<div class="flex items-center gap-2">
					<h3 class="text-sm font-bold text-zinc-900 dark:text-foreground">Видео верификации квартиры</h3>
					<span class="inline-flex items-center gap-1 rounded-full bg-zinc-100 px-2 py-0.5 text-[10px] font-semibold text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
						<Lock class="h-3 w-3" />
						Конфиденциально
					</span>
				</div>
				<p class="text-xs text-zinc-500 dark:text-muted-foreground mt-0.5">
					Видно только вам и модератору сервиса
				</p>
			</div>
		</div>

		{#if currentVideoUrl}
			<span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300">
				<CheckCircle2 class="h-3.5 w-3.5" />
				Видео прикреплено
			</span>
		{/if}
	</div>

	<!-- Explanation Notice -->
	<div class="rounded-2xl bg-zinc-50 p-4 text-xs leading-relaxed text-zinc-600 dark:bg-muted/40 dark:text-muted-foreground mb-5 border border-zinc-100 dark:border-border/60">
		<p class="font-medium text-zinc-900 dark:text-foreground mb-1">Как снять подтверждающее видео:</p>
		<ul class="list-disc list-inside space-y-1 pl-1">
			<li>Начните запись снаружи у входной двери, чтобы был виден номер квартиры/дома.</li>
			<li>Откройте дверь своим ключом и пройдите по всем комнатам, кухне и санузлу.</li>
			<li>Покажите вид из окна на улицу для подтверждения реального адреса.</li>
			<li>Длительность: 1–2 минуты, формат: MP4, MOV или WebM до 100 МБ.</li>
		</ul>
	</div>

	<!-- Upload or Preview Area -->
	{#if currentVideoUrl}
		<div class="space-y-4">
			<div class="relative overflow-hidden rounded-2xl border border-zinc-200 bg-black aspect-video max-h-[380px] flex items-center justify-center dark:border-border">
				<!-- svelte-ignore a11y_media_has_caption -->
				<video
					src={currentVideoUrl}
					controls
					preload="metadata"
					class="h-full w-full object-contain"
				>
					Ваш браузер не поддерживает воспроизведение этого видео.
				</video>
			</div>

			<div class="flex items-center justify-between pt-1">
				<p class="text-xs text-zinc-500 dark:text-muted-foreground">
					Если хотите заменить видео, выберите новый файл:
				</p>
				<label class="inline-flex items-center gap-1.5 rounded-full border border-zinc-200 bg-white px-3.5 py-1.5 text-xs font-semibold text-zinc-700 shadow-2xs hover:bg-zinc-50 active:scale-95 cursor-pointer dark:border-border dark:bg-card dark:text-foreground">
					<RotateCcw class="h-3.5 w-3.5" />
					<span>Заменить видео</span>
					<input
						type="file"
						accept="video/mp4,video/quicktime,video/webm,.mov"
						class="sr-only"
						onchange={(e) => handleFileSelect(e.currentTarget.files)}
						disabled={disabled || isUploading}
					/>
				</label>
			</div>
		</div>
	{:else}
		<div
			class="relative flex flex-col items-center justify-center rounded-2xl border-2 border-dashed p-8 text-center transition-colors duration-150 {isDragging
				? 'border-blue-500 bg-blue-50/50 dark:bg-blue-950/20'
				: 'border-zinc-300 hover:border-zinc-400 bg-zinc-50/50 dark:border-border dark:bg-muted/20'}"
			ondrop={handleDrop}
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
			role="region"
			aria-label="Загрузка подтверждающего видео"
		>
			{#if isUploading}
				<div class="py-6 flex flex-col items-center gap-3">
					<Loader2 class="h-8 w-8 text-blue-600 animate-spin" />
					<div class="text-center">
						<p class="text-sm font-semibold text-zinc-900 dark:text-foreground">Загрузка видео...</p>
						<p class="text-xs text-zinc-500 dark:text-muted-foreground mt-0.5">Это может занять некоторое время в зависимости от размера файла</p>
					</div>
					{#if uploadProgress > 0}
						<div class="w-48 h-2 bg-zinc-200 rounded-full overflow-hidden mt-1 dark:bg-zinc-700">
							<div
								class="h-full bg-blue-600 transition-all duration-300 rounded-full"
								style="width: {uploadProgress}%"
							></div>
						</div>
					{/if}
				</div>
			{:else}
				<div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-white shadow-2xs border border-zinc-200/80 dark:bg-card dark:border-border mb-3 text-zinc-600 dark:text-zinc-300">
					<UploadCloud class="h-6 w-6" />
				</div>

				<p class="text-sm font-semibold text-zinc-900 dark:text-foreground">
					Перетащите файл видео сюда или
					<label class="text-blue-600 hover:underline cursor-pointer font-bold ml-1 dark:text-blue-400">
						выберите на устройстве
						<input
							type="file"
							accept="video/mp4,video/quicktime,video/webm,.mov"
							class="sr-only"
							onchange={(e) => handleFileSelect(e.currentTarget.files)}
							disabled={disabled || isUploading}
						/>
					</label>
				</p>
				<p class="mt-1 text-xs text-zinc-400 dark:text-muted-foreground">
					MP4, MOV или WebM до 100 МБ
				</p>
			{/if}
		</div>
	{/if}

	{#if errorMessage}
		<div class="mt-3 flex items-center gap-2 rounded-xl bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/60 dark:text-rose-300 border border-rose-200/60 dark:border-rose-900">
			<AlertCircle class="h-4 w-4 shrink-0" />
			<span>{errorMessage}</span>
		</div>
	{/if}
</div>
