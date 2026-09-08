<!-- src/routes/auth/+layout.svelte -->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import '../../app.css';
	import ToastContainer from '$lib/components/ui/ToastContainer.svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<svelte:head>
	<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
	<meta name="theme-color" content="#fafafa" />
</svelte:head>

{@render children()}

<ToastContainer />
