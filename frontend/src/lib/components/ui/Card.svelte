<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		class?: string;
		children?: Snippet;
	}

	let { class: className = '', children, ...restProps }: Props = $props();
</script>

<div
	class={cn('rounded-xl border border-border bg-card text-card-foreground shadow-sm transition-all', className)}
	{...restProps}
>
	{@render children?.()}
</div>
