<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth/auth.svelte';
	import { listingsApi } from '$lib/api/listings';
	import type { ListingHost } from '$lib/types/listings';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import { formatCurrency, translateHousingType, translateListingStatus, formatDate, cn } from '$lib/utils';
	import {
		Building2,
		PlusCircle,
		LayoutDashboard,
		ShieldCheck,
		ExternalLink,
		Clock,
		AlertCircle,
		CheckCircle2,
		Sparkles
	} from 'lucide-svelte';

	let myListings = $state<ListingHost[]>([]);
	let loading = $state(true);
	let errorMessage = $state<string | null>(null);

	async function loadHostListings() {
		if (!auth.isAuthenticated) {
			loading = false;
			return;
		}

		loading = true;
		errorMessage = null;
		try {
			const res = await listingsApi.getMyListings();
			myListings = res || [];
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Не удалось загрузить объекты';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadHostListings();
	});

	$effect(() => {
		const userId = auth.user?.id;
		if (userId) {
			loadHostListings();
		} else {
			myListings = [];
			loading = false;
		}
	});

	let stats = $derived({
		total: myListings.length,
		published: myListings.filter((l) => l.status === 'published').length,
		pending: myListings.filter(
			(l) => l.status === 'pending_review' || l.status === 'awaiting_company_verification'
		).length
	});
</script>

<svelte:head>
	<title>Кабинет хоста — Flickey</title>
</svelte:head>

<div class="container mx-auto px-4 sm:px-6 py-8 space-y-8 max-w-6xl">
	<!-- Top Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl sm:text-3xl font-extrabold text-foreground tracking-tight">
				Панель управления жильем
			</h1>
			<p class="text-xs sm:text-sm text-muted-foreground mt-1">
				Управляйте своими объектами, проверяйте статусы модерации и создавайте новые объявления.
			</p>
		</div>

		<a href="/host/new">
			<Button class="gap-2 font-semibold shadow-md">
				<PlusCircle class="h-4 w-4" /> Добавить жилье
			</Button>
		</a>
	</div>

	{#if !auth.isAuthenticated}
		<!-- Unauthenticated State -->
		<div class="rounded-3xl border border-border bg-card p-8 sm:p-12 text-center max-w-lg mx-auto space-y-4 shadow-sm">
			<div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
				<LayoutDashboard class="h-7 w-7" />
			</div>
			<h3 class="text-lg font-bold text-foreground">Войдите в систему для доступа к кабинету</h3>
			<p class="text-xs text-muted-foreground leading-relaxed">
				Авторизуйтесь по номеру телефона, чтобы просматривать свои объекты и управлять бронированиями.
			</p>
			<Button onclick={() => auth.openAuthModal('phone')} class="font-semibold">
				Войти в аккаунт
			</Button>
		</div>
	{:else if !auth.isHost && !loading && myListings.length === 0}
		<!-- Non-host user state: Promotion CTA -->
		<div class="rounded-3xl border border-primary/30 bg-primary/5 p-8 sm:p-12 text-center max-w-xl mx-auto space-y-4">
			<div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-md">
				<Sparkles class="h-7 w-7" />
			</div>
			<h2 class="text-xl sm:text-2xl font-extrabold text-foreground">
				Станьте Хостом на Flickey
			</h2>
			<p class="text-xs sm:text-sm text-muted-foreground leading-relaxed">
				У вас пока нет активных объектов. Создайте свое первое объявление — это займет всего пару минут, и вы автоматически получите статус Хоста.
			</p>
			<a href="/host/new" class="inline-block pt-2">
				<Button size="lg" class="gap-2 font-bold shadow-lg">
					<PlusCircle class="h-5 w-5" /> Создать первое объявление
				</Button>
			</a>
		</div>
	{:else}
		<!-- Host Stats Overview -->
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
			<div class="rounded-2xl border border-border bg-card p-5 shadow-sm">
				<div class="text-xs text-muted-foreground font-medium">Всего объектов</div>
				<div class="text-2xl font-extrabold text-foreground mt-1">{stats.total}</div>
			</div>

			<div class="rounded-2xl border border-border bg-card p-5 shadow-sm">
				<div class="text-xs text-muted-foreground font-medium">Опубликовано</div>
				<div class="text-2xl font-extrabold text-emerald-600 mt-1">{stats.published}</div>
			</div>

			<div class="rounded-2xl border border-border bg-card p-5 shadow-sm">
				<div class="text-xs text-muted-foreground font-medium">На проверке / верификации</div>
				<div class="text-2xl font-extrabold text-amber-600 mt-1">{stats.pending}</div>
			</div>
		</div>

		<!-- Listings Content Table/Grid -->
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="text-base font-bold text-foreground">Список ваших объектов</h3>
				<Button variant="outline" size="sm" onclick={loadHostListings} disabled={loading} class="text-xs">
					Обновить
				</Button>
			</div>

			{#if loading}
				<div class="space-y-3">
					{#each Array(3) as _}
						<Skeleton class="h-20 w-full rounded-2xl" />
					{/each}
				</div>
			{:else if myListings.length > 0}
				<!-- Desktop Table -->
				<div class="hidden md:block overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
					<table class="w-full text-left text-xs">
						<thead class="border-b border-border bg-muted/40 font-semibold text-muted-foreground">
							<tr>
								<th class="p-4">Объект</th>
								<th class="p-4">Категория</th>
								<th class="p-4">Цена / сутки</th>
								<th class="p-4">Статус</th>
								<th class="p-4">Дата создания</th>
								<th class="p-4 text-right">Действия</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-border">
							{#each myListings as item (item.id)}
								{@const statusInfo = translateListingStatus(item.status)}
								<tr class="hover:bg-muted/20 transition-colors">
									<td class="p-4">
										<div class="font-bold text-foreground text-sm line-clamp-1">{item.name}</div>
										<div class="text-[11px] text-muted-foreground font-mono mt-0.5">ID: {item.id.slice(0, 8)}</div>
									</td>
									<td class="p-4">
										<span class="rounded-lg bg-muted px-2.5 py-1 font-medium">{translateHousingType(item.type)}</span>
									</td>
									<td class="p-4 font-bold text-foreground">
										{formatCurrency(item.price_per_night, item.currency)}
									</td>
									<td class="p-4">
										<span class={cn('inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-semibold', statusInfo.color)}>
											{statusInfo.label}
										</span>
									</td>
									<td class="p-4 text-muted-foreground">
										{formatDate(item.created_at)}
									</td>
									<td class="p-4 text-right">
										<a href={`/listings/${item.id}`} class="inline-flex">
											<Button variant="ghost" size="sm" class="h-8 gap-1 text-xs">
												<ExternalLink class="h-3.5 w-3.5" /> Просмотр
											</Button>
										</a>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				<!-- Mobile Cards Grid -->
				<div class="grid grid-cols-1 gap-4 md:hidden">
					{#each myListings as item (item.id)}
						{@const statusInfo = translateListingStatus(item.status)}
						<div class="rounded-2xl border border-border bg-card p-4 space-y-3 shadow-sm">
							<div class="flex items-start justify-between gap-2">
								<div>
									<h4 class="font-bold text-sm text-foreground line-clamp-1">{item.name}</h4>
									<p class="text-[11px] text-muted-foreground font-mono mt-0.5">ID: {item.id.slice(0, 8)}</p>
								</div>
								<span class={cn('inline-flex items-center rounded-full border px-2.5 py-0.5 text-[10px] font-semibold shrink-0', statusInfo.color)}>
									{statusInfo.label}
								</span>
							</div>

							<div class="flex items-center justify-between text-xs pt-2 border-t border-border">
								<span class="text-muted-foreground">{translateHousingType(item.type)}</span>
								<span class="font-bold text-foreground">{formatCurrency(item.price_per_night, item.currency)} / сут.</span>
							</div>

							<div class="pt-2 flex justify-end">
								<a href={`/listings/${item.id}`} class="w-full">
									<Button variant="outline" size="sm" class="w-full text-xs gap-1">
										<ExternalLink class="h-3.5 w-3.5" /> Просмотр на сайте
									</Button>
								</a>
							</div>
						</div>
					{/each}
				</div>
			{:else}
				<div class="rounded-2xl border border-dashed border-border p-8 text-center text-xs text-muted-foreground space-y-2">
					<p>У вас пока нет созданных объектов недвижимости.</p>
					<a href="/host/new" class="inline-block pt-1">
						<Button size="sm">Создать объявление</Button>
					</a>
				</div>
			{/if}
		</div>
	{/if}
</div>
