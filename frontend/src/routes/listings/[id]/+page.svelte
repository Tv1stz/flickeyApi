<script lang="ts">
	import type { PageData } from './$types';
	import ListingView from '$lib/components/listings/ListingView.svelte';
	import { ChevronRight, Home } from 'lucide-svelte';

	let { data }: { data: PageData } = $props();
	let listing = $derived(data.listing);
</script>

<svelte:head>
	<title>{listing.name} — Flickey</title>
	<meta name="description" content={listing.description.slice(0, 160)} />
</svelte:head>

<div class="container mx-auto px-4 sm:px-6 py-6 space-y-6 max-w-6xl">
	<!-- Breadcrumb Navigation -->
	<nav class="flex items-center gap-2 text-xs text-muted-foreground">
		<a href="/" class="hover:text-foreground transition-colors flex items-center gap-1">
			<Home class="h-3.5 w-3.5" /> Главная
		</a>
		<ChevronRight class="h-3 w-3" />
		<a href="/" class="hover:text-foreground transition-colors">Каталог</a>
		<ChevronRight class="h-3 w-3" />
		<span class="text-foreground font-medium truncate max-w-[200px] sm:max-w-md">{listing.name}</span>
	</nav>

	<!-- Master Shared Listing Presentation -->
	<ListingView {listing} mode="public" />
</div>
