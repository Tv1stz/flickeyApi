<script lang="ts">
	import ListingWizard from '$lib/components/wizard/ListingWizard.svelte';
	import { auth } from '$lib/auth/auth.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Sparkles, ArrowLeft } from 'lucide-svelte';
</script>

<svelte:head>
	<title>Создание объявления — Flickey</title>
</svelte:head>

<div class="container mx-auto px-4 sm:px-6 py-6 space-y-6 max-w-4xl">
	<!-- Top Bar -->
	<div class="flex items-center justify-between">
		<a href="/host" class="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors">
			<ArrowLeft class="h-4 w-4" /> Назад в кабинет
		</a>
		<div class="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-3 py-0.5 text-xs font-bold text-primary">
			<Sparkles class="h-3.5 w-3.5" /> Мастер создания жилья
		</div>
	</div>

	{#if !auth.isAuthenticated}
		<div class="rounded-3xl border border-border bg-card p-8 sm:p-12 text-center max-w-md mx-auto space-y-4 shadow-sm">
			<h3 class="text-lg font-bold text-foreground">Требуется авторизация</h3>
			<p class="text-xs text-muted-foreground leading-relaxed">
				Для создания объявления необходимо войти в систему.
			</p>
			<Button onclick={() => auth.openAuthModal('phone')} class="font-semibold">
				Войти по номеру телефона
			</Button>
		</div>
	{:else}
		<!-- 6-step Listing Creation Wizard -->
		<ListingWizard />
	{/if}
</div>
