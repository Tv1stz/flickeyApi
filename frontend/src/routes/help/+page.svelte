<script lang="ts">
	import PageShell from '$lib/components/ui/page/PageShell.svelte';
	import InfoPageCard from '$lib/components/ui/page/InfoPageCard.svelte';
</script>

<svelte:head>
	<title>Помощь — Flickey</title>
</svelte:head>

<PageShell maxWidth="narrow" background="white" containerClass="py-10 lg:py-14">
	<InfoPageCard
		title="Помощь"
		description="Раздел поддержки в разработке. Скоро здесь появятся ответы на частые вопросы и способы связи со службой поддержки."
	/>
</PageShell>
