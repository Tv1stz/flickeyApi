<script lang="ts">
	import { Image as ImageIcon, X, ChevronLeft, ChevronRight } from 'lucide-svelte';

	interface Props {
		photos: string[];
		name: string;
	}

	let { photos = [], name = '' }: Props = $props();

	let selectedIndex = $state<number | null>(null);

	const fallbackImages = [
		'https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?w=1200&auto=format&fit=crop&q=80',
		'https://images.unsplash.com/photo-1502672260266-1c1ef2d93688?w=1200&auto=format&fit=crop&q=80',
		'https://images.unsplash.com/photo-1560448204-e02f11c3d0e2?w=1200&auto=format&fit=crop&q=80',
		'https://images.unsplash.com/photo-1484154218962-a197022b5858?w=1200&auto=format&fit=crop&q=80'
	];

	let displayPhotos = $derived(photos.length > 0 ? photos : fallbackImages);

	function openModal(index: number) {
		selectedIndex = index;
	}

	function closeModal() {
		selectedIndex = null;
	}

	function prevPhoto() {
		if (selectedIndex === null) return;
		selectedIndex = (selectedIndex - 1 + displayPhotos.length) % displayPhotos.length;
	}

	function nextPhoto() {
		if (selectedIndex === null) return;
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

<!-- Gallery Grid Layout -->
<div class="relative overflow-hidden rounded-2xl">
	{#if displayPhotos.length === 1}
		<button
			type="button"
			onclick={() => openModal(0)}
			class="relative aspect-[16/9] w-full overflow-hidden bg-muted cursor-pointer block"
		>
			<img src={displayPhotos[0]} alt={name} class="h-full w-full object-cover hover:scale-102 transition-transform duration-300" />
		</button>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-4 md:grid-rows-2 gap-2 aspect-[4/3] md:aspect-[21/9]">
			<!-- Main Large Photo -->
			<button
				type="button"
				onclick={() => openModal(0)}
				class="relative md:col-span-2 md:row-span-2 overflow-hidden bg-muted cursor-pointer group"
			>
				<img
					src={displayPhotos[0]}
					alt={`${name} - 1`}
					class="h-full w-full object-cover group-hover:scale-105 transition-transform duration-300"
				/>
			</button>

			<!-- 4 Secondary Photos -->
			{#each displayPhotos.slice(1, 5) as photo, idx}
				<button
					type="button"
					onclick={() => openModal(idx + 1)}
					class="relative hidden md:block overflow-hidden bg-muted cursor-pointer group"
				>
					<img
						src={photo}
						alt={`${name} - ${idx + 2}`}
						class="h-full w-full object-cover group-hover:scale-105 transition-transform duration-300"
					/>
				</button>
			{/each}
		</div>
	{/if}

	<!-- View All Photos Badge -->
	<button
		type="button"
		onclick={() => openModal(0)}
		class="absolute bottom-4 right-4 flex items-center gap-1.5 rounded-xl bg-background/90 backdrop-blur-md px-3.5 py-2 text-xs font-bold text-foreground shadow-lg hover:bg-background transition-all cursor-pointer"
	>
		<ImageIcon class="h-4 w-4 text-primary" />
		<span>Все фото ({displayPhotos.length})</span>
	</button>
</div>

<!-- Modal Fullscreen Lightbox -->
{#if selectedIndex !== null}
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
		<div class="relative max-h-[85vh] max-w-[90vw] overflow-hidden rounded-xl">
			<img
				src={displayPhotos[selectedIndex]}
				alt={`${name} - ${selectedIndex + 1}`}
				class="max-h-[85vh] max-w-[90vw] object-contain rounded-xl shadow-2xl"
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
