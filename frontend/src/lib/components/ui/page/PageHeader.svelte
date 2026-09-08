<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		subtitle?: string;
		meta?: string;
		actions?: Snippet;
		class?: string;
	}

	let { title, subtitle = '', meta = '', actions, class: className = '' }: Props = $props();
</script>

<div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between {className}">
	<div>
		<h1 class="text-2xl font-bold text-zinc-900 sm:text-3xl">{title}</h1>
		{#if subtitle}
			<p class="mt-1 text-sm text-zinc-500">{subtitle}</p>
		{/if}
		{#if meta}
			<p class="mt-1 text-sm text-zinc-500">{meta}</p>
		{/if}
	</div>

	{#if actions}
		<div class="flex items-center gap-3">
			{@render actions()}
		</div>
	{/if}
</div>
