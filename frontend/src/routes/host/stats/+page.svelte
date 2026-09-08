<!-- src/routes/host/stats/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		BarChart3,
		Eye,
		Heart,
		TrendingUp,
		Wallet,
		CalendarCheck,
		ArrowUpRight,
		Sparkles,
		Building2,
		Info,
		Layers,
		CheckCircle2,
		ShieldCheck,
		Calendar,
		LayoutGrid
	} from 'lucide-svelte';

	import PageShell from '$lib/components/ui/page/PageShell.svelte';
	import PageHeader from '$lib/components/ui/page/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/page/EmptyState.svelte';

	import { hostListingsStore } from '$lib/stores/listingsStore.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { canAccessHostArea } from '$lib/auth/permissions';
	import { formatBYN, pluralRu } from '$lib/utils/format';
	import { formatListingAddressFull } from '$lib/utils/location';
	import type { Listing } from '$lib/components/card/types';

	let guardChecked = $state(false);
	let listings = $derived(hostListingsStore.items);

	type Period = '30d' | 'year' | 'all';
	let activePeriod = $state<Period>('30d');

	$effect(() => {
		if (authStore.initialized) {
			if (!authStore.user) {
				authStore.setPendingAction(null, '/host/stats');
				goto(resolve('/auth'));
				return;
			}
			if (!guardChecked) {
				if (!canAccessHostArea(authStore.user)) {
					goto(resolve('/profile'));
					return;
				}
				if (authStore.viewMode !== 'host') authStore.setViewMode('host');
				guardChecked = true;
				hostListingsStore.refresh();
			}
		}
	});

	onMount(() => {
		hostListingsStore.initialize();
	});

	// Mock statistics tailored to host's actual listings count
	const stats = $derived.by(() => {
		const count = listings.length;
		const multiplier = activePeriod === 'year' ? 12 : activePeriod === 'all' ? 24 : 1;

		const avgPrice = count > 0 ? listings.reduce((acc, l) => acc + l.pricePerNight, 0) / count : 120;
		const estNightsPerMonth = Math.round(count * 20); // ~67% occupancy
		const estRevenue = Math.round(avgPrice * estNightsPerMonth * multiplier);
		const views = count * 240 * multiplier;
		const favorites = count * 18 * multiplier;

		return {
			count,
			estRevenue,
			occupancyRate: count > 0 ? 74 : 0,
			views,
			favorites,
			confirmedBookings: count * 4 * multiplier,
			avgRating: 4.96
		};
	});

	// Month bars for Airbnb-like performance chart
	const MONTHS = [
		{ label: 'Окт', height: 45, val: '2 850 BYN' },
		{ label: 'Ноя', height: 55, val: '3 400 BYN' },
		{ label: 'Дек', height: 85, val: '5 200 BYN' },
		{ label: 'Янв', height: 70, val: '4 100 BYN' },
		{ label: 'Фев', height: 50, val: '3 100 BYN' },
		{ label: 'Мар', height: 60, val: '3 700 BYN' },
		{ label: 'Апр', height: 65, val: '4 000 BYN' },
		{ label: 'Май', height: 80, val: '4 900 BYN' },
		{ label: 'Июн', height: 95, val: '6 100 BYN' },
		{ label: 'Июл', height: 100, val: '6 500 BYN' },
		{ label: 'Авг', height: 90, val: '5 800 BYN' },
		{ label: 'Сен', height: 75, val: '4 600 BYN' }
	];
</script>

<svelte:head>
	<title>Статистика и доходы — Flickey</title>
</svelte:head>

{#if guardChecked}
	<PageShell maxWidth="wide" background="white">
		<PageHeader
			title="Статистика и доходы"
			subtitle="Аналитика просмотров, заполняемости и финансовых показателей объектов"
		>
			{#snippet actions()}
				<div class="flex flex-wrap items-center gap-2">
					<Button
						variant="outline"
						size="sm"
						radius="xl"
						href={resolve('/host/listings')}
						as="a"
					>
						<LayoutGrid class="h-3.5 w-3.5 mr-1.5" />
						Мои объявления
					</Button>
					<Button
						variant="outline"
						size="sm"
						radius="xl"
						href={resolve('/host/calendar')}
						as="a"
					>
						<Calendar class="h-3.5 w-3.5 mr-1.5" />
						Календарь
					</Button>

					<!-- Period Selector Pills -->
					<div class="flex items-center gap-1 rounded-2xl bg-zinc-100 dark:bg-zinc-800/80 p-1 border border-zinc-200/80 dark:border-zinc-700">
					<button
						type="button"
						onclick={() => { activePeriod = '30d'; }}
						class="rounded-xl px-3 py-1.5 text-xs font-bold transition-all cursor-pointer
                               {activePeriod === '30d'
								? 'bg-white text-zinc-900 shadow-xs dark:bg-zinc-900 dark:text-white'
								: 'text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-white'}"
					>
						За 30 дней
					</button>
					<button
						type="button"
						onclick={() => { activePeriod = 'year'; }}
						class="rounded-xl px-3 py-1.5 text-xs font-bold transition-all cursor-pointer
                               {activePeriod === 'year'
								? 'bg-white text-zinc-900 shadow-xs dark:bg-white dark:text-zinc-900'
								: 'text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-white'}"
					>
						Этот год
					</button>
					<button
						type="button"
						onclick={() => { activePeriod = 'all'; }}
						class="rounded-xl px-3 py-1.5 text-xs font-bold transition-all cursor-pointer
                               {activePeriod === 'all'
								? 'bg-white text-zinc-900 shadow-xs dark:bg-white dark:text-zinc-900'
								: 'text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-white'}"
					>
						За всё время
					</button>
				</div>
			</div>
		{/snippet}
		</PageHeader>

		{#if listings.length === 0}
			<EmptyState
				icon={BarChart3}
				title="Пока нет данных для статистики"
				description="Создайте и опубликуйте своё первое объявление, чтобы отслеживать аналитику просмотров, бронирований и доходов."
				actionLabel="Создать объявление"
				actionHref={resolve('/listings/new')}
			/>
		{:else}
			<!-- ═══════════════════ AIRBNB-STYLE KPI METRICS GRID ═══════════════════ -->
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
				<!-- Revenue Card -->
				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-zinc-500 uppercase tracking-wider">Прогноз дохода</span>
						<div class="flex h-9 w-9 items-center justify-center rounded-2xl bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400">
							<Wallet class="h-4 w-4" />
						</div>
					</div>
					<div>
						<p class="text-2xl sm:text-3xl font-extrabold text-zinc-900 dark:text-zinc-100">
							{formatBYN(stats.estRevenue)}
						</p>
						<p class="mt-1 flex items-center gap-1 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
							<TrendingUp class="h-3.5 w-3.5" />
							+12.4% к прошлому периоду
						</p>
					</div>
				</div>

				<!-- Occupancy Rate -->
				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-zinc-500 uppercase tracking-wider">Заполняемость</span>
						<div class="flex h-9 w-9 items-center justify-center rounded-2xl bg-sky-50 dark:bg-sky-950/40 text-sky-600 dark:text-sky-400">
							<CalendarCheck class="h-4 w-4" />
						</div>
					</div>
					<div>
						<p class="text-2xl sm:text-3xl font-extrabold text-zinc-900 dark:text-zinc-100">
							{stats.occupancyRate}%
						</p>
						<p class="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
							~22 забронированные ночи / месяц
						</p>
					</div>
				</div>

				<!-- Views & Discovery -->
				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-zinc-500 uppercase tracking-wider">Просмотры</span>
						<div class="flex h-9 w-9 items-center justify-center rounded-2xl bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400">
							<Eye class="h-4 w-4" />
						</div>
					</div>
					<div>
						<p class="text-2xl sm:text-3xl font-extrabold text-zinc-900 dark:text-zinc-100">
							{stats.views.toLocaleString()}
						</p>
						<p class="mt-1 flex items-center gap-1 text-xs text-zinc-500 dark:text-zinc-400">
							<Heart class="h-3.5 w-3.5 text-rose-500" />
							{stats.favorites} сохранений в избранное
						</p>
					</div>
				</div>

				<!-- Quality Score -->
				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-zinc-500 uppercase tracking-wider">Оценка качества</span>
						<div class="flex h-9 w-9 items-center justify-center rounded-2xl bg-amber-50 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400">
							<Sparkles class="h-4 w-4" />
						</div>
					</div>
					<div>
						<p class="text-2xl sm:text-3xl font-extrabold text-zinc-900 dark:text-zinc-100">
							{stats.avgRating} <span class="text-sm font-normal text-zinc-400">/ 5.0</span>
						</p>
						<p class="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
							100% подтверждённых заездов
						</p>
					</div>
				</div>
			</div>

			<!-- ═══════════════════ AIRBNB-STYLE REVENUE CHART PLACEHOLDER ═══════════════════ -->
			<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-6 sm:p-8 shadow-sm mb-8 space-y-6">
				<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
					<div>
						<div class="flex items-center gap-2.5">
							<h3 class="text-base sm:text-lg font-extrabold text-zinc-900 dark:text-zinc-100">
								Динамика доходов и бронирований
							</h3>
							<span class="rounded-full bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-800 px-2.5 py-0.5 text-[10px] font-bold text-amber-700 dark:text-amber-300">
								Демо-аналитика
							</span>
						</div>
						<p class="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
							Помесячное распределение дохода и загруженности вашего номерного фонда
						</p>
					</div>

					<div class="flex items-center gap-3 text-xs text-zinc-500">
						<div class="flex items-center gap-1.5">
							<span class="h-3 w-3 rounded-full bg-zinc-900 dark:bg-white"></span>
							<span>Доход (BYN)</span>
						</div>
					</div>
				</div>

				<!-- Visual Bar Chart -->
				<div class="pt-6 pb-2">
					<div class="flex items-end justify-between gap-2 sm:gap-4 h-56 border-b border-zinc-100 dark:border-zinc-800 pb-2">
						{#each MONTHS as m}
							<div class="group relative flex-1 flex flex-col items-center h-full justify-end">
								<!-- Tooltip on hover -->
								<div class="opacity-0 group-hover:opacity-100 transition-opacity absolute -top-8 bg-zinc-900 dark:bg-zinc-800 text-white rounded-lg px-2 py-1 text-[10px] font-bold pointer-events-none whitespace-nowrap shadow-md z-10">
									{m.val}
								</div>

								<!-- Bar element -->
								<div
									class="w-full max-w-[36px] rounded-xl bg-gradient-to-t from-zinc-800 to-zinc-900 dark:from-zinc-200 dark:to-white group-hover:opacity-80 transition-all duration-300"
									style="height: {m.height}%;"
								></div>

								<span class="mt-2 text-[11px] font-semibold text-zinc-400 group-hover:text-zinc-900 dark:group-hover:text-white transition-colors">
									{m.label}
								</span>
							</div>
						{/each}
					</div>
				</div>

				<div class="rounded-2xl bg-zinc-50 dark:bg-zinc-800/40 p-4 border border-zinc-100 dark:border-zinc-800 flex items-center gap-3 text-xs text-zinc-500 dark:text-zinc-400">
					<Info class="h-4 w-4 shrink-0 text-zinc-400" />
					<span>
						Детальная финансовая выписка, акты выполненных работ и интеграция с расчетным счетом ИП/юрлица по закону РБ будут активированы на этапе подключения онлайн-эквайринга.
					</span>
				</div>
			</div>

			<!-- ═══════════════════ PROPERTY BREAKDOWN TABLE ═══════════════════ -->
			<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden shadow-sm mb-8">
				<div class="p-5 border-b border-zinc-100 dark:border-zinc-800">
					<h3 class="text-base font-bold text-zinc-900 dark:text-zinc-100">
						Доходность по объектам
					</h3>
					<p class="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5">
						Сравнение показателей и загрузки ваших квартир
					</p>
				</div>

				<div class="divide-y divide-zinc-100 dark:divide-zinc-800">
					{#each listings as item (item.id)}
						{@const fullAddr = formatListingAddressFull(item.address, item.location)}
						<div class="flex flex-col sm:flex-row sm:items-center justify-between p-4 sm:px-6 gap-4 hover:bg-zinc-50/50 dark:hover:bg-zinc-800/30 transition">
							<div class="flex items-center gap-3.5 min-w-0">
								<div class="relative h-12 w-16 shrink-0 overflow-hidden rounded-2xl bg-zinc-100 dark:bg-zinc-800">
									{#if item.images?.[0]}
										<img src={item.images[0]} alt="" class="h-full w-full object-cover" />
									{/if}
								</div>
								<div class="min-w-0">
									<p class="font-bold text-sm text-zinc-900 dark:text-zinc-100 truncate">{item.title}</p>
									<p class="text-xs text-zinc-500 dark:text-zinc-400 truncate mt-0.5">{fullAddr}</p>
								</div>
							</div>

							<div class="flex items-center gap-6 sm:gap-8 self-end sm:self-center">
								<div class="text-right">
									<p class="text-xs text-zinc-400">Тариф / ночь</p>
									<p class="text-xs font-bold text-zinc-900 dark:text-zinc-100 mt-0.5">{formatBYN(item.pricePerNight)}</p>
								</div>
								<div class="text-right">
									<p class="text-xs text-zinc-400">Загрузка</p>
									<p class="text-xs font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">~76%</p>
								</div>
								<Button
									variant="outline"
									size="sm"
									radius="xl"
									href={`/host/calendar?listing_id=${item.id}`}
									as="a"
								>
									Календарь
								</Button>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- ═══════════════════ AIRBNB-STYLE HOST TIPS ═══════════════════ -->
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm space-y-2">
					<p class="text-xs font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
						<span class="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 text-[11px] font-extrabold">1</span>
						Синхронизируйте iCal
					</p>
					<p class="text-xs text-zinc-500 dark:text-zinc-400 leading-relaxed">
						Подключите календари Airbnb, Booking и Авито, чтобы избежать двойных броней и повысить конверсию.
					</p>
				</div>

				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm space-y-2">
					<p class="text-xs font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
						<span class="flex h-5 w-5 items-center justify-center rounded-full bg-blue-100 text-blue-700 text-[11px] font-extrabold">2</span>
						Качественные фото
					</p>
					<p class="text-xs text-zinc-500 dark:text-zinc-400 leading-relaxed">
						Объявления с 10+ яркими фотографиями и видеоверификацией получают на 35% больше прямых откликов гостей.
					</p>
				</div>

				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm space-y-2">
					<p class="text-xs font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
						<span class="flex h-5 w-5 items-center justify-center rounded-full bg-purple-100 text-purple-700 text-[11px] font-extrabold">3</span>
						Верификация партнера
					</p>
					<p class="text-xs text-zinc-500 dark:text-zinc-400 leading-relaxed">
						Подтвердите статус самозанятого или юрлица в РБ для легальной сдачи и гарантированного вывода средств.
					</p>
				</div>
			</div>
		{/if}
	</PageShell>
{/if}
