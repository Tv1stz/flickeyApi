<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';
	import Button from '$lib/components/ui/Button.svelte';

	type IconComponent = ComponentType;

	interface Props {
		icon?: IconComponent;
		title: string;
		description?: string;
		actionLabel?: string;
		actionHref?: string;
		onAction?: () => void;
		children?: Snippet;
	}

	let {
		icon,
		title,
		description = '',
		actionLabel = '',
		actionHref,
		onAction,
		children
	}: Props = $props();
</script>

<div class="py-16 text-center">
	{#if icon}
		{@const Icon = icon}
		<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-100">
			<Icon class="h-8 w-8 text-zinc-400" />
		</div>
	{/if}

	<h3 class="mb-2 text-lg font-semibold text-zinc-900">{title}</h3>
	{#if description}
		<p class="mb-6 text-zinc-500">{description}</p>
	{/if}

	{#if actionLabel}
		<Button
			variant="solid"
			tone="primary"
			size="lg"
			radius="xl"
			as={actionHref ? 'a' : 'button'}
			href={actionHref}
			onclick={onAction}
		>
			{actionLabel}
		</Button>
	{/if}

	{#if children}
		<div class="mt-4">
			{@render children()}
		</div>
	{/if}
</div>
