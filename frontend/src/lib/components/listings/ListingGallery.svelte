<script lang="ts">
	import { X, ChevronLeft, ChevronRight, Image as ImageIcon } from 'lucide-svelte';

	interface Props {
		photos: string[];
		name: string;
	}

	let { photos = [], name = '' }: Props = $props();

	let selectedIndex = $state<number | null>(null);

	let displayPhotos = $derived(photos || []);
	let extraCount = $derived(Math.max(0, displayPhotos.length - 4));

	function openModal(index: number) {
		selectedIndex = index;
	}

	function closeModal() {
		selectedIndex = null;
	}

	function prevPhoto() {
		if (selectedIndex === null || displayPhotos.length === 0) return;
		selectedIndex = (selectedIndex - 1 + displayPhotos.length) % displayPhotos.length;
	}

	function nextPhoto() {
		if (selectedIndex === null || displayPhotos.length === 0) return;
		selectedIndex = (selectedIndex + 1) % displayPhotos.length;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (selectedIndex === null) return;
		if (e.key === 'Escape') closeModal();
		if (e.key === 'ArrowLeft') prevPhoto();
		if (e.key === 'ArrowRight') nextPhoto();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if displayPhotos.length === 0}
	<div class="w-full aspect-[2.2/1] rounded-3xl bg-muted/40 border border-border/80 flex flex-col items-center justify-center text-muted-foreground gap-2">
		<ImageIcon class="h-8 w-8 text-muted-foreground/60" />
		<span class="text-xs font-medium">Фотографии отсутствуют</span>
	</div>
{:else if displayPhotos.length === 1}
	<div class="relative overflow-hidden rounded-3xl aspect-[2.2/1] bg-muted">
		<button
			type="button"
			onclick={() => openModal(0)}
			class="w-full h-full cursor-pointer block group"
		>
			<img
				src={displayPhotos[0]}
				alt={name}
				class="w-full h-full object-cover group-hover:scale-102 transition-transform duration-300"
			/>
		</button>
	</div>
{:else}
	<!-- Bento Grid 5-Photo Layout matching screenshot -->
	<div class="grid grid-cols-1 md:grid-cols-4 md:grid-rows-2 gap-2.5 rounded-3xl overflow-hidden aspect-[4/3] md:aspect-[2.2/1]">
		<!-- Photo 1: Large Main Hero (left column, spans 2 rows) -->
		<button
			type="button"
			onclick={() => openModal(0)}
			class="relative md:col-span-2 md:row-span-2 overflow-hidden rounded-2xl bg-muted cursor-pointer group block w-full h-full"
		>
			<img
				src={displayPhotos[0]}
				alt={`${name} - 1`}
				class="h-full w-full object-cover group-hover:scale-102 transition-transform duration-300"
			/>
		</button>

		<!-- Photo 2: Top Middle -->
		{#if displayPhotos[1]}
			<button
				type="button"
				onclick={() => openModal(1)}
				class="relative hidden md:block overflow-hidden rounded-2xl bg-muted cursor-pointer group w-full h-full"
			>
				<img
					src={displayPhotos[1]}
					alt={`${name} - 2`}
					class="h-full w-full object-cover group-hover:scale-103 transition-transform duration-300"
				/>
			</button>
		{/if}

		<!-- Photo 3: Top Right -->
		{#if displayPhotos[2]}
			<button
				type="button"
				onclick={() => openModal(2)}
				class="relative hidden md:block overflow-hidden rounded-2xl bg-muted cursor-pointer group w-full h-full"
			>
				<img
					src={displayPhotos[2]}
					alt={`${name} - 3`}
					class="h-full w-full object-cover group-hover:scale-103 transition-transform duration-300"
				/>
			</button>
		{/if}

		<!-- Photo 4: Bottom Middle -->
		{#if displayPhotos[3]}
			<button
				type="button"
				onclick={() => openModal(3)}
				class="relative hidden md:block overflow-hidden rounded-2xl bg-muted cursor-pointer group w-full h-full"
			>
				<img
					src={displayPhotos[3]}
					alt={`${name} - 4`}
					class="h-full w-full object-cover group-hover:scale-103 transition-transform duration-300"
				/>
			</button>
		{/if}

		<!-- Photo 5: Bottom Right with Overlay badge -->
		{#if displayPhotos[4]}
			<button
				type="button"
				onclick={() => openModal(4)}
				class="relative hidden md:block overflow-hidden rounded-2xl bg-muted cursor-pointer group w-full h-full"
			>
				<img
					src={displayPhotos[4]}
					alt={`${name} - 5`}
					class="h-full w-full object-cover group-hover:scale-103 transition-transform duration-300"
				/>
				{#if extraCount > 0}
					<div class="absolute inset-0 bg-black/45 backdrop-blur-[1px] flex items-center justify-center group-hover:bg-black/55 transition-colors">
						<span class="text-white font-medium text-xs sm:text-sm">
							ещё {extraCount} фото
						</span>
					</div>
				{/if}
			</button>
		{/if}
	</div>
{/if}

<!-- Modal Fullscreen Lightbox -->
{#if selectedIndex !== null && displayPhotos.length > 0}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-md p-4 animate-in fade-in"
		role="dialog"
		aria-modal="true"
	>
		<!-- Close Button -->
		<button
			type="button"
			onclick={closeModal}
			class="absolute top-6 right-6 rounded-full bg-white/10 p-2 text-white hover:bg-white/20 transition-colors cursor-pointer z-50"
			aria-label="Закрыть"
		>
			<X class="h-6 w-6" />
		</button>

		<!-- Counter -->
		<div class="absolute top-6 left-6 text-sm font-semibold text-white/80">
			{selectedIndex + 1} / {displayPhotos.length}
		</div>

		<!-- Prev -->
		<button
			type="button"
			onclick={prevPhoto}
			class="absolute left-4 md:left-8 rounded-full bg-white/10 p-3 text-white hover:bg-white/20 transition-colors cursor-pointer z-50"
			aria-label="Предыдущее фото"
		>
			<ChevronLeft class="h-6 w-6" />
		</button>

		<!-- Main Image -->
		<div class="relative max-h-[85vh] max-w-[90vw] overflow-hidden rounded-2xl">
			<img
				src={displayPhotos[selectedIndex]}
				alt={`${name} - ${selectedIndex + 1}`}
				class="max-h-[85vh] max-w-[90vw] object-contain rounded-2xl shadow-2xl"
			/>
		</div>

		<!-- Next -->
		<button
			type="button"
			onclick={nextPhoto}
			class="absolute right-4 md:right-8 rounded-full bg-white/10 p-3 text-white hover:bg-white/20 transition-colors cursor-pointer z-50"
			aria-label="Следующее фото"
		>
			<ChevronRight class="h-6 w-6" />
		</button>
	</div>
{/if}
