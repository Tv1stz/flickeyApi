<script lang="ts">
	import '../app.css';
	import type { Snippet } from 'svelte';
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth/auth.svelte';
	import Header from '$lib/components/layout/Header.svelte';
	import Footer from '$lib/components/layout/Footer.svelte';
	import AuthDialog from '$lib/components/auth/AuthDialog.svelte';
	import Toaster from '$lib/components/ui/Toaster.svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();

	onMount(() => {
		auth.init();
	});
</script>

<div class="min-h-screen flex flex-col bg-background text-foreground selection:bg-primary/20 selection:text-primary">
	<Header />

	<main class="flex-1">
		{@render children()}
	</main>

	<Footer />

	<!-- Global Auth Dialog Modal -->
	<AuthDialog />

	<!-- Global Toast Notifications Container -->
	<Toaster />
</div>
