<script lang="ts">
	import { Share2, Heart } from 'lucide-svelte';
	import type { HousingType } from '$lib/types/listings';

	interface Props {
		name: string;
		address?: string;
		city?: string;
		street?: string;
		type?: HousingType;
		createdAt?: string;
		id?: string;
	}

	let { name, address, city, street }: Props = $props();

	let displayAddress = $derived(
		address ||
		(city && street ? `${city}, ${street}` : city || '')
	);

	let isSaved = $state(false);
	let isCopied = $state(false);

	async function handleShare() {
		if (typeof navigator !== 'undefined' && navigator.clipboard) {
			try {
				await navigator.clipboard.writeText(window.location.href);
				isCopied = true;
				setTimeout(() => (isCopied = false), 2000);
			} catch {
				// clipboard fallback
			}
		}
	}

	function handleSave() {
		isSaved = !isSaved;
	}
</script>

<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
	<!-- Title & Address Header -->
	<div class="space-y-1">
		<h1 class="text-xl sm:text-2xl lg:text-3xl font-bold text-foreground tracking-tight leading-snug">
			{name}
		</h1>
		{#if displayAddress}
			<p class="text-xs text-muted-foreground font-normal">
				{displayAddress}
			</p>
		{/if}
	</div>

	<!-- Top Actions: Share & Save -->
	<div class="flex items-center gap-4 shrink-0 text-xs font-medium text-foreground/80">
		<button
			type="button"
			onclick={handleShare}
			class="flex items-center gap-1.5 hover:text-foreground transition-colors cursor-pointer py-1"
			title="Поделиться ссылкой"
		>
			<Share2 class="h-3.5 w-3.5" />
			<span>{isCopied ? 'Скопировано!' : 'Поделиться'}</span>
		</button>

		<button
			type="button"
			onclick={handleSave}
			class="flex items-center gap-1.5 hover:text-foreground transition-colors cursor-pointer py-1"
			title={isSaved ? 'Удалить из сохранённых' : 'Сохранить'}
		>
			<Heart class={`h-3.5 w-3.5 transition-colors ${isSaved ? 'fill-rose-500 text-rose-500' : ''}`} />
			<span>{isSaved ? 'Сохранено' : 'Сохранить'}</span>
		</button>
	</div>
</div>
