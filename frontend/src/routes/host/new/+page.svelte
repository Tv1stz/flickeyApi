<!-- src/routes/host/new/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';

	onMount(() => {
		const search = $page.url.search || '';
		goto(resolve(('/listings/new' + search) as '/'), { replaceState: true });
	});
</script>

<div class="flex min-h-screen items-center justify-center bg-white">
	<div class="h-8 w-8 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"></div>
</div>
