<!-- src/lib/components/listing-page/ListingDescription.svelte -->
<script lang="ts">
	import { ChevronDown } from 'lucide-svelte';

	interface Props {
		description: string;
	}

	let { description }: Props = $props();

	let expanded = $state(false);
	let contentEl: HTMLDivElement | undefined = $state();
	let naturalHeight = $state(0);

	const COLLAPSED_HEIGHT = 112;
	const needsTruncate = $derived(description.length > 250);

	$effect(() => {
		if (!contentEl) return;
		const measure = () => {
			naturalHeight = contentEl!.scrollHeight;
		};
		measure();
		const ro = new ResizeObserver(measure);
		ro.observe(contentEl);
		return () => ro.disconnect();
	});

	const currentHeight = $derived(
		!needsTruncate ? 'auto' : expanded ? `${naturalHeight}px` : `${COLLAPSED_HEIGHT}px`
	);

	function toggle() {
		expanded = !expanded;
	}
</script>

<div>
	<h2 class="mb-5 text-[20px] font-semibold tracking-tight text-zinc-900">Об этом месте</h2>

	<div class="relative">
		<div
			bind:this={contentEl}
			class="overflow-hidden transition-[max-height] duration-500 ease-[cubic-bezier(0.33,1,0.68,1)] will-change-[max-height]"
			style="max-height: {currentHeight};"
		>
			<p class="text-[15px] leading-relaxed break-words whitespace-pre-wrap text-zinc-600">
				{description}
			</p>
		</div>

		<!-- Gradient fade -->
		<div
			class="pointer-events-none absolute right-0 bottom-0 left-0 h-14
                   bg-gradient-to-t from-white to-transparent
                   transition-opacity duration-400 ease-out
                   {needsTruncate && !expanded ? 'opacity-100' : 'opacity-0'}"
			aria-hidden="true"
		></div>
	</div>

	{#if needsTruncate}
		<button
			type="button"
			onclick={toggle}
			class="mt-4 inline-flex items-center gap-1.5 text-[14px] font-semibold
                   text-zinc-900 underline decoration-zinc-300 underline-offset-4
                   transition-colors hover:decoration-zinc-600"
		>
			{expanded ? 'Свернуть' : 'Показать полностью'}
			<ChevronDown
				size={14}
				strokeWidth={2.5}
				class="transition-transform duration-400 ease-[cubic-bezier(0.33,1,0.68,1)]
                       {expanded ? 'rotate-180' : ''}"
			/>
		</button>
	{/if}
</div>
