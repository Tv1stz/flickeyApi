<script lang="ts">
	import type { PageData } from './$types';
	import ListingPage from '$lib/components/listing-page/ListingPage.svelte';

	let { data }: { data: PageData } = $props();
	let listing = $derived(data.listing);
</script>

<svelte:head>
	{#if listing}
		<title>{listing.title} — Flickey</title>
		<meta name="description" content={listing.description?.slice(0, 160) ?? listing.address} />
	{:else}
		<title>Объявление — Flickey</title>
	{/if}
</svelte:head>

{#if listing}
	<ListingPage {listing} />
{/if}

