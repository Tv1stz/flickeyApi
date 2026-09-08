<script lang="ts">
	import type { PageData } from './$types';
	import type { Listing } from '$lib/components/card/types';
	import ListingGrid from '$lib/components/card/ListingGrid.svelte';
	import { favoritesStore } from '$lib/stores/favoritesStore.svelte';
	import { pluralRu } from '$lib/utils/format';
	import { ArrowLeft, Share2, ShieldCheck, CheckCircle, Sparkles, Home, Calendar, BadgeCheck } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { toast } from '$lib/stores/toastStore';

	let { data }: { data: PageData } = $props();

	let user = $derived(data.user);
	let listings = $derived(data.listings || []);
	let favorites = $derived(favoritesStore.favoriteIds);

	const hostName = $derived(
		user?.name ||
		(user?.first_name ? `${user.first_name}${user.last_name ? ' ' + user.last_name : ''}` : '') ||
		(listings[0]?.owner?.name ?? 'Хозяин жилья')
	);

	const hostFirstName = $derived(
		user?.first_name ||
		hostName.split(' ')[0] ||
		'Хозяин'
	);

	const initials = $derived(
		hostName
			.split(' ')
			.filter(Boolean)
			.map((word: string) => word[0])
			.join('')
			.toUpperCase()
			.slice(0, 2) || 'Х'
	);

	const memberSinceYear = $derived(
		user?.created_at
			? new Date(user.created_at).getFullYear()
			: listings[0]?.owner?.createdAt
				? new Date(listings[0].owner.createdAt).getFullYear()
				: 2026
	);

	const yearsOnPlatform = $derived(Math.max(0, new Date().getFullYear() - memberSinceYear));
	const listingsCount = $derived(listings.length || user?.listings_count || 0);
	const isVerified = $derived(user?.is_verified ?? true);

	function handleBack() {
		if (typeof history !== 'undefined' && history.length > 1) {
			history.back();
			return;
		}
		goto(resolve('/search'));
	}

	async function handleShare() {
		if (typeof navigator !== 'undefined' && navigator.clipboard) {
			try {
				await navigator.clipboard.writeText(window.location.href);
				toast.success('Ссылка на профиль скопирована');
			} catch {
				toast.info('Скопируйте URL из адресной строки');
			}
		}
	}

	function handleFavoriteToggle(id: string) {
		const targetListing = listings.find((l: Listing) => l.id === id);
		if (targetListing) {
			favoritesStore.toggle(targetListing);
		}
	}
</script>

<svelte:head>
	<title>{hostName} — Профиль хозяина жилья на Flickey</title>
	<meta name="description" content="Объявления посуточной аренды от хозяина {hostName} в сервисе Flickey." />
</svelte:head>

<div class="min-h-screen bg-zinc-50/50 pb-20 pt-6 sm:pt-8 lg:pt-10">
	<div class="mx-auto max-w-[1440px] px-4 sm:px-6 lg:px-8">
		<!-- Навигация и действия -->
		<div class="mb-6 flex items-center justify-between gap-4">
			<button
				type="button"
				onclick={handleBack}
				class="group inline-flex items-center gap-2 rounded-full border border-zinc-200/80 bg-white px-4 py-2 text-sm font-semibold text-zinc-700 shadow-sm transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95"
			>
				<ArrowLeft size={16} class="transition group-hover:-translate-x-0.5" />
				Назад
			</button>

			<button
				type="button"
				onclick={handleShare}
				class="inline-flex items-center gap-2 rounded-full border border-zinc-200/80 bg-white px-4 py-2 text-sm font-semibold text-zinc-700 shadow-sm transition hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900 active:scale-95"
			>
				<Share2 size={16} />
				Поделиться
			</button>
		</div>

		<!-- Основной 2-колоночный layout -->
		<div class="grid grid-cols-1 gap-8 lg:grid-cols-[380px_1fr] lg:gap-12 items-start">
			<!-- Левая колонка: Карточка профиля хозяина -->
			<aside class="w-full lg:sticky lg:top-24">
				<div class="overflow-hidden rounded-3xl border border-zinc-200/80 bg-white p-6 sm:p-8 shadow-sm">
					<!-- Аватар и имя -->
					<div class="flex flex-col items-center text-center">
						<div
							class="flex h-28 w-28 items-center justify-center rounded-full bg-gradient-to-br from-zinc-800 to-zinc-950 text-[34px] font-bold text-white shadow-md ring-4 ring-white"
						>
							{initials}
						</div>

						<h1 class="mt-4 text-2xl font-bold tracking-tight text-zinc-900">
							{hostName}
						</h1>

						<div class="mt-1.5 inline-flex items-center gap-1.5 rounded-full bg-zinc-100 px-3 py-1 text-xs font-semibold text-zinc-700">
							<Home size={13} class="text-zinc-500" />
							Хозяин жилья
						</div>

						{#if isVerified}
							<div class="mt-3 flex items-center justify-center gap-1.5 text-xs font-semibold text-emerald-600">
								<BadgeCheck size={16} class="text-emerald-500" />
								Личность подтверждена
							</div>
						{:else}
							<div class="mt-3 text-xs text-zinc-400">
								Профиль не подтвержден
							</div>
						{/if}

						<div class="mt-2 flex items-center justify-center gap-1.5 text-xs text-zinc-500">
							<Calendar size={14} class="text-zinc-400" />
							На платформе с {memberSinceYear} года
						</div>
					</div>

					<!-- Статистика -->
					<div class="mt-6 grid grid-cols-2 divide-x divide-zinc-100 rounded-2xl bg-zinc-50/80 p-4">
						<div class="flex flex-col items-center justify-center text-center">
							<span class="text-2xl font-extrabold tracking-tight text-zinc-900">
								{listingsCount}
							</span>
							<span class="mt-0.5 text-xs font-medium text-zinc-500">
								{pluralRu(listingsCount, ['объявление', 'объявления', 'объявлений'])}
							</span>
						</div>

						<div class="flex flex-col items-center justify-center text-center">
							<span class="text-2xl font-extrabold tracking-tight text-zinc-900">
								{yearsOnPlatform > 0 ? yearsOnPlatform : '1-й'}
							</span>
							<span class="mt-0.5 text-xs font-medium text-zinc-500">
								{yearsOnPlatform > 0 ? pluralRu(yearsOnPlatform, ['год', 'года', 'лет']) : 'год на платформе'}
							</span>
						</div>
					</div>

					<!-- Дополнительная информация о доверии -->
					<div class="mt-6 space-y-3.5 border-t border-zinc-100 pt-6">
						<div class="flex items-center gap-3 text-xs text-zinc-700">
							<ShieldCheck size={18} class="shrink-0 text-zinc-900" />
							<span>Безопасная аренда через Flickey</span>
						</div>
						<div class="flex items-center gap-3 text-xs text-zinc-700">
							<CheckCircle size={18} class="shrink-0 text-zinc-900" />
							<span>Проверенные фотографии и описание объекта</span>
						</div>
						<div class="flex items-center gap-3 text-xs text-zinc-700">
							<Sparkles size={18} class="shrink-0 text-zinc-900" />
							<span>Быстрое подтверждение бронирований</span>
						</div>
					</div>
				</div>
			</aside>

			<!-- Правая колонка: Список объявлений -->
			<main class="w-full">
				<!-- Заголовок секции -->
				<div class="mb-6 flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<div class="flex items-center gap-3">
							<h2 class="text-2xl font-bold tracking-tight text-zinc-900">
								Объявления хозяина
							</h2>
							<span class="flex h-6 min-w-[24px] items-center justify-center rounded-full bg-zinc-900 px-2 text-xs font-bold text-white">
								{listings.length}
							</span>
						</div>
						<p class="mt-1 text-sm text-zinc-500">
							Доступно для безопасного бронирования от {hostFirstName}
						</p>
					</div>
				</div>

				<!-- Сетка объявлений -->
				{#if listings.length > 0}
					<ListingGrid
						{listings}
						{favorites}
						onFavoriteToggle={handleFavoriteToggle}
						layout="default"
					/>
				{:else}
					<div class="flex flex-col items-center justify-center rounded-3xl border border-zinc-200/80 bg-white p-12 text-center shadow-sm">
						<div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400">
							<Home size={32} strokeWidth={1.5} />
						</div>
						<h3 class="mt-4 text-lg font-bold text-zinc-900">
							У хозяина пока нет активных объявлений
						</h3>
						<p class="mt-1 max-w-sm text-sm text-zinc-500">
							Возможно, сейчас все объекты хозяина находятся на публикации или сданы. Попробуйте поискать другие предложения в каталоге.
						</p>
						<a
							href={resolve('/search')}
							class="mt-6 inline-flex items-center gap-2 rounded-full bg-zinc-900 px-6 py-2.5 text-sm font-semibold text-white transition hover:bg-zinc-800 active:scale-95"
						>
							Перейти к поиску
						</a>
					</div>
				{/if}
			</main>
		</div>
	</div>
</div>