<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		variant?: 'default' | 'secondary' | 'outline' | 'destructive' | 'success' | 'warning' | 'info';
		class?: string;
		children?: Snippet;
	}

	let {
		variant = 'default',
		class: className = '',
		children,
		...restProps
	}: Props = $props();

	const variants: Record<string, string> = {
		default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
		secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
		outline: 'text-foreground border-border',
		destructive: 'border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80',
		success: 'border-emerald-200 bg-emerald-50 text-emerald-700',
		warning: 'border-amber-200 bg-amber-50 text-amber-700',
		info: 'border-sky-200 bg-sky-50 text-sky-700'
	};
</script>

<div
	class={cn(
		'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
		variants[variant] || variants.default,
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
