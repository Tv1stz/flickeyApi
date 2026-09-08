<script lang="ts">
	import type { Snippet } from 'svelte';

	type MaxWidth = 'narrow' | 'content' | 'wide';
	type Background = 'white' | 'zinc';

	interface Props {
		children?: Snippet;
		maxWidth?: MaxWidth;
		background?: Background;
		includeBottomNavSpace?: boolean;
		class?: string;
		containerClass?: string;
	}

	let {
		children,
		maxWidth = 'wide',
		background = 'white',
		includeBottomNavSpace = true,
		class: className = '',
		containerClass = ''
	}: Props = $props();

	const maxWidthClass = $derived.by(() => {
		if (maxWidth === 'narrow') return 'max-w-4xl';
		if (maxWidth === 'content') return 'max-w-5xl';
		return 'max-w-[1600px]';
	});
</script>

<div
	class="min-h-screen {background === 'white' ? 'bg-white' : 'bg-zinc-50'}
        {includeBottomNavSpace ? 'pb-24 lg:pb-8' : ''}
        {className}"
>
	<div class="mx-auto {maxWidthClass} px-4 py-8 sm:px-6 lg:px-8 lg:py-12 {containerClass}">
		{@render children?.()}
	</div>
</div>
