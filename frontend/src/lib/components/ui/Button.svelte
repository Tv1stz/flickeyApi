<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	interface Props extends HTMLButtonAttributes {
		variant?: 'default' | 'secondary' | 'outline' | 'destructive' | 'ghost' | 'link';
		size?: 'sm' | 'default' | 'lg' | 'icon';
		class?: string;
		children?: Snippet;
		loading?: boolean;
	}

	let {
		variant = 'default',
		size = 'default',
		class: className = '',
		children,
		loading = false,
		disabled = false,
		type = 'button',
		...restProps
	}: Props = $props();

	const variants: Record<string, string> = {
		default: 'bg-primary text-primary-foreground hover:bg-primary-hover shadow-sm',
		secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
		outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
		destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-sm',
		ghost: 'hover:bg-accent hover:text-accent-foreground',
		link: 'text-primary underline-offset-4 hover:underline'
	};

	const sizes: Record<string, string> = {
		sm: 'h-8 rounded-md px-3 text-xs',
		default: 'h-10 px-4 py-2 text-sm',
		lg: 'h-12 rounded-lg px-6 text-base font-semibold',
		icon: 'h-10 w-10 p-0'
	};
</script>

<button
	{type}
	disabled={disabled || loading}
	class={cn(
		'inline-flex items-center justify-center whitespace-nowrap rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 select-none cursor-pointer',
		variants[variant] || variants.default,
		sizes[size] || sizes.default,
		className
	)}
	{...restProps}
>
	{#if loading}
		<svg
			class="mr-2 h-4 w-4 animate-spin"
			xmlns="http://www.w3.org/2000/svg"
			fill="none"
			viewBox="0 0 24 24"
		>
			<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
			></circle>
			<path
				class="opacity-75"
				fill="currentColor"
				d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
			></path>
		</svg>
	{/if}
	{@render children?.()}
</button>
