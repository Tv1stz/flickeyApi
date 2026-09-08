<script lang="ts">
	interface Props {
		title: string;
		description: string;
		caption?: string;
	}

	let { title, description, caption = '' }: Props = $props();
</script>

<div class="app-surface p-6 sm:p-8">
	<h1 class="text-2xl font-bold text-zinc-900 sm:text-3xl">{title}</h1>
	<p class="mt-3 text-base text-zinc-600">{description}</p>
	{#if caption}
		<p class="mt-4 text-sm text-zinc-500">{caption}</p>
	{/if}
</div>
