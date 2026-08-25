<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';
	import { AlertCircle, CheckCircle2, Info, AlertTriangle } from 'lucide-svelte';

	interface Props {
		variant?: 'default' | 'destructive' | 'success' | 'warning';
		title?: string;
		class?: string;
		children?: Snippet;
	}

	let {
		variant = 'default',
		title = '',
		class: className = '',
		children
	}: Props = $props();

	const variants: Record<string, string> = {
		default: 'bg-background text-foreground border-border',
		destructive: 'border-destructive/50 text-destructive bg-destructive/10 dark:border-destructive [&>svg]:text-destructive',
		success: 'border-emerald-500/50 text-emerald-800 bg-emerald-50 dark:bg-emerald-950/20 dark:text-emerald-300 [&>svg]:text-emerald-600',
		warning: 'border-amber-500/50 text-amber-800 bg-amber-50 dark:bg-amber-950/20 dark:text-amber-300 [&>svg]:text-amber-600'
	};
</script>

<div
	role="alert"
	class={cn(
		'relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4',
		variants[variant] || variants.default,
		className
	)}
>
	{#if variant === 'destructive'}
		<AlertCircle class="h-4 w-4" />
	{:else if variant === 'success'}
		<CheckCircle2 class="h-4 w-4" />
	{:else if variant === 'warning'}
		<AlertTriangle class="h-4 w-4" />
	{:else}
		<Info class="h-4 w-4 text-primary" />
	{/if}

	{#if title}
		<h5 class="mb-1 font-semibold leading-none tracking-tight">{title}</h5>
	{/if}
	<div class="text-sm [&_p]:leading-relaxed">
		{@render children?.()}
	</div>
</div>
