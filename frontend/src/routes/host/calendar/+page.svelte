<!-- src/routes/host/calendar/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import {
		Calendar as CalendarIcon,
		RefreshCw,
		Copy,
		Check,
		Trash2,
		Plus,
		Lock,
		Globe,
		ChevronLeft,
		ChevronRight,
		Info,
		CheckCircle2,
		AlertTriangle,
		Building2,
		Search,
		ExternalLink,
		SlidersHorizontal,
		Palette,
		Eye,
		Layers,
		ArrowUpRight,
		Phone,
		User,
		BarChart3,
		LayoutGrid
	} from 'lucide-svelte';

	import PageShell from '$lib/components/ui/page/PageShell.svelte';
	import PageHeader from '$lib/components/ui/page/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EmptyState from '$lib/components/ui/page/EmptyState.svelte';

	import { hostListingsStore } from '$lib/stores/listingsStore.svelte';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { canAccessHostArea } from '$lib/auth/permissions';
	import { calendarApi } from '$lib/api/calendar';
	import { toast } from '$lib/stores/toastStore';
	import { formatBYN, pluralRu } from '$lib/utils/format';
	import { formatListingAddressFull } from '$lib/utils/location';
	import type { Listing } from '$lib/components/card/types';
	import type {
		HostCalendarResponse,
		ListingReservation,
		ListingCalendarSync
	} from '$lib/types/calendar';

	// Available color presets for iCal feeds
	interface ColorPreset {
		key: string;
		label: string;
		dotClass: string;
		badgeBg: string;
		badgeText: string;
		badgeBorder: string;
		calendarCellBg: string;
		calendarCellBorder: string;
	}

	const COLOR_PRESETS: ColorPreset[] = [
		{
			key: 'rose',
			label: 'Красный (Airbnb)',
			dotClass: 'bg-rose-500',
			badgeBg: 'bg-rose-50 dark:bg-rose-950/40',
			badgeText: 'text-rose-700 dark:text-rose-300',
			badgeBorder: 'border-rose-200 dark:border-rose-900',
			calendarCellBg: '!bg-rose-50 dark:!bg-rose-950/40',
			calendarCellBorder: '!border-rose-200 dark:!border-rose-800'
		},
		{
			key: 'blue',
			label: 'Синий (Booking.com)',
			dotClass: 'bg-sky-500',
			badgeBg: 'bg-sky-50 dark:bg-sky-950/40',
			badgeText: 'text-sky-700 dark:text-sky-300',
			badgeBorder: 'border-sky-200 dark:border-sky-900',
			calendarCellBg: '!bg-sky-50 dark:!bg-sky-950/40',
			calendarCellBorder: '!border-sky-200 dark:!border-sky-800'
		},
		{
			key: 'purple',
			label: 'Фиолетовый (Авито)',
			dotClass: 'bg-purple-500',
			badgeBg: 'bg-purple-50 dark:bg-purple-950/40',
			badgeText: 'text-purple-700 dark:text-purple-300',
			badgeBorder: 'border-purple-200 dark:border-purple-900',
			calendarCellBg: '!bg-purple-50 dark:!bg-purple-950/40',
			calendarCellBorder: '!border-purple-200 dark:!border-purple-800'
		},
		{
			key: 'amber',
			label: 'Оранжевый (Суточно.ру)',
			dotClass: 'bg-amber-500',
			badgeBg: 'bg-amber-50 dark:bg-amber-950/40',
			badgeText: 'text-amber-700 dark:text-amber-300',
			badgeBorder: 'border-amber-200 dark:border-amber-900',
			calendarCellBg: '!bg-amber-50 dark:!bg-amber-950/40',
			calendarCellBorder: '!border-amber-200 dark:!border-amber-800'
		},
		{
			key: 'emerald',
			label: 'Изумрудный (Сайт/Прямые)',
			dotClass: 'bg-emerald-500',
			badgeBg: 'bg-emerald-50 dark:bg-emerald-950/40',
			badgeText: 'text-emerald-700 dark:text-emerald-300',
			badgeBorder: 'border-emerald-200 dark:border-emerald-900',
			calendarCellBg: '!bg-emerald-50 dark:!bg-emerald-950/40',
			calendarCellBorder: '!border-emerald-200 dark:!border-emerald-800'
		},
		{
			key: 'teal',
			label: 'Бирюзовый (Островок)',
			dotClass: 'bg-teal-500',
			badgeBg: 'bg-teal-50 dark:bg-teal-950/40',
			badgeText: 'text-teal-700 dark:text-teal-300',
			badgeBorder: 'border-teal-200 dark:border-teal-900',
			calendarCellBg: '!bg-teal-50 dark:!bg-teal-950/40',
			calendarCellBorder: '!border-teal-200 dark:!border-teal-800'
		},
		{
			key: 'indigo',
			label: 'Индиго (Стандарт)',
			dotClass: 'bg-indigo-500',
			badgeBg: 'bg-indigo-50 dark:bg-indigo-950/40',
			badgeText: 'text-indigo-700 dark:text-indigo-300',
			badgeBorder: 'border-indigo-200 dark:border-indigo-900',
			calendarCellBg: '!bg-indigo-50 dark:!bg-indigo-950/40',
			calendarCellBorder: '!border-indigo-200 dark:!border-indigo-800'
		}
	];

	function getPresetTheme(colorKey?: string): ColorPreset {
		if (!colorKey) return COLOR_PRESETS[6]; // indigo
		return COLOR_PRESETS.find((p) => p.key === colorKey) || COLOR_PRESETS[6];
	}

	let guardChecked = $state(false);
	let listings = $derived(hostListingsStore.items);

	// Property search filter for property switcher
	let propertySearch = $state('');
	let filteredListings = $derived(
		propertySearch.trim()
			? listings.filter((l) =>
					l.title.toLowerCase().includes(propertySearch.toLowerCase()) ||
					formatListingAddressFull(l.address, l.location).toLowerCase().includes(propertySearch.toLowerCase())
				)
			: listings
	);

	// Active selected listing
	let selectedListingId = $state<string | null>(null);
	let selectedListing = $derived(listings.find((l) => l.id === selectedListingId) ?? listings[0] ?? null);

	// Active tab
	type Tab = 'calendar' | 'sync' | 'overview';
	let activeTab = $state<Tab>('calendar');

	// Calendar state for selected listing
	let loading = $state(false);
	let syncing = $state(false);
	let calendarData = $state<HostCalendarResponse | null>(null);

	// Manual block form state
	let blockStartDate = $state('');
	let blockEndDate = $state('');
	let blockNote = $state('');
	let blockGuestName = $state('');
	let blockGuestPhone = $state('');
	let blockSubmitting = $state(false);

	// Add feed form state
	let newFeedName = $state('');
	let newFeedUrl = $state('');
	let newFeedColor = $state('rose');
	let feedSubmitting = $state(false);

	// Inline feed color editing state
	let editingFeedColorId = $state<string | null>(null);

	// Copy URL state
	let copied = $state(false);

	// Month view offset (0 = current month)
	let viewMonthOffset = $state(0);

	const RU_MONTHS = [
		'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
		'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
	];
	const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

	function toISODate(d: Date): string {
		const year = d.getFullYear();
		const month = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	const todayISO = $derived(toISODate(new Date()));

	// Auth Guard & Store initialization
	$effect(() => {
		if (authStore.initialized) {
			if (!authStore.user) {
				authStore.setPendingAction(null, '/host/calendar');
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

	onMount(async () => {
		await hostListingsStore.initialize();
		// Read initial listing_id from URL query if provided
		const paramId = $page.url.searchParams.get('listing_id');
		if (paramId) {
			selectedListingId = paramId;
		}
	});

	// Synchronize selected listing ID once listings are loaded
	$effect(() => {
		if (listings.length > 0) {
			if (!selectedListingId || !listings.some((l) => l.id === selectedListingId)) {
				const paramId = $page.url.searchParams.get('listing_id');
				if (paramId && listings.some((l) => l.id === paramId)) {
					selectedListingId = paramId;
				} else {
					selectedListingId = listings[0].id;
				}
			}
		}
	});

	// Load calendar data whenever the selected listing changes
	$effect(() => {
		if (selectedListingId) {
			loadCalendarData(selectedListingId);
		}
	});

	function selectListing(id: string) {
		selectedListingId = id;
		// Update URL without page reload
		const url = new URL(window.location.href);
		url.searchParams.set('listing_id', id);
		window.history.replaceState({}, '', url.toString());
	}

	async function loadCalendarData(listingId: string) {
		try {
			loading = true;
			calendarData = await calendarApi.getHostCalendar(listingId);
		} catch (err: any) {
			toast.error('Ошибка загрузки календаря', err?.message || 'Попробуйте позже');
		} finally {
			loading = false;
		}
	}

	async function handleBlockDates(e: SubmitEvent) {
		e.preventDefault();
		if (!selectedListingId) return;

		if (!blockStartDate || !blockEndDate) {
			toast.error('Укажите даты', 'Выберите дату заезда и выезда');
			return;
		}
		if (blockEndDate <= blockStartDate) {
			toast.error('Неверный период', 'Дата выезда должна быть позже даты заезда');
			return;
		}

		try {
			blockSubmitting = true;
			await calendarApi.blockDates(selectedListingId, {
				start_date: blockStartDate,
				end_date: blockEndDate,
				note: blockNote.trim() || undefined,
				guest_name: blockGuestName.trim() || undefined,
				guest_phone: blockGuestPhone.trim() || undefined
			});
			toast.success('Даты заблокированы', `${blockStartDate} — ${blockEndDate}`);
			blockStartDate = '';
			blockEndDate = '';
			blockNote = '';
			blockGuestName = '';
			blockGuestPhone = '';
			await loadCalendarData(selectedListingId);
		} catch (err: any) {
			toast.error('Не удалось заблокировать', err?.message || 'Даты конфликтуют с существующей бронью');
		} finally {
			blockSubmitting = false;
		}
	}

	async function handleUnblock(resId: string) {
		if (!selectedListingId) return;
		try {
			await calendarApi.unblockDates(selectedListingId, resId);
			toast.success('Даты разблокированы');
			await loadCalendarData(selectedListingId);
		} catch (err: any) {
			toast.error('Ошибка', err?.message || 'Не удалось разблокировать даты');
		}
	}

	async function handleAddFeed(e: SubmitEvent) {
		e.preventDefault();
		if (!selectedListingId) return;

		if (!newFeedName.trim() || !newFeedUrl.trim()) {
			toast.error('Заполните поля', 'Укажите название и URL iCal фида');
			return;
		}

		try {
			feedSubmitting = true;
			await calendarApi.addSyncFeed(selectedListingId, {
				name: newFeedName.trim(),
				feed_url: newFeedUrl.trim(),
				color: newFeedColor
			});
			toast.success('Календарь подключён', 'Синхронизация запущена в фоне');
			newFeedName = '';
			newFeedUrl = '';
			await loadCalendarData(selectedListingId);
		} catch (err: any) {
			toast.error('Ошибка подключения фида', err?.message || 'Проверьте корректность URL');
		} finally {
			feedSubmitting = false;
		}
	}

	async function handleUpdateFeedColor(feedId: string, colorKey: string) {
		if (!selectedListingId) return;
		try {
			await calendarApi.updateSyncFeed(selectedListingId, feedId, { color: colorKey });
			if (calendarData?.sync_feeds) {
				calendarData.sync_feeds = calendarData.sync_feeds.map((f) =>
					f.id === feedId ? { ...f, color: colorKey } : f
				);
			}
			toast.success('Цвет изменён');
			editingFeedColorId = null;
		} catch (err: any) {
			toast.error('Ошибка смены цвета', err?.message || 'Попробуйте позже');
		}
	}

	async function handleDeleteFeed(syncId: string) {
		if (!selectedListingId) return;
		toast.confirm({
			title: 'Удалить календарь?',
			message: 'Синхронизация прекратится, а все импортированные этим фидом бронирования будут сняты.',
			confirmText: 'Удалить',
			cancelText: 'Отмена',
			type: 'danger',
			onConfirm: async () => {
				try {
					await calendarApi.deleteSyncFeed(selectedListingId!, syncId);
					toast.success('Календарь удалён');
					await loadCalendarData(selectedListingId!);
				} catch (err: any) {
					toast.error('Ошибка', err?.message || 'Не удалось удалить фид');
				}
			}
		});
	}

	async function handleSyncNow() {
		if (!selectedListingId) return;
		try {
			syncing = true;
			const res = await calendarApi.syncNow(selectedListingId);
			toast.success('Синхронизация запущена', res.message || 'Календарь обновляется');
			setTimeout(async () => {
				if (selectedListingId) await loadCalendarData(selectedListingId);
				syncing = false;
			}, 2500);
		} catch (err: any) {
			syncing = false;
			toast.error('Ошибка синхронизации', err?.message || 'Попробуйте позже');
		}
	}

	async function copyExportUrl(url?: string) {
		const targetUrl = url || calendarData?.export_url;
		if (!targetUrl) return;
		try {
			await navigator.clipboard.writeText(targetUrl);
			copied = true;
			toast.success('Ссылка скопирована', 'Вставьте её на внешнем сервисе');
			setTimeout(() => { copied = false; }, 3000);
		} catch {
			toast.error('Не удалось скопировать ссылку');
		}
	}

	function handleDayClick(dayDate: string) {
		if (!blockStartDate || (blockStartDate && blockEndDate)) {
			blockStartDate = dayDate;
			blockEndDate = '';
		} else if (blockStartDate && !blockEndDate) {
			if (dayDate > blockStartDate) {
				blockEndDate = dayDate;
			} else {
				blockStartDate = dayDate;
				blockEndDate = '';
			}
		}
	}

	// Month grid calculations
	function getReservationsForDate(iso: string): ListingReservation[] {
		if (!calendarData?.reservations) return [];
		return calendarData.reservations.filter((r) => iso >= r.start_date && iso < r.end_date && r.status !== 'cancelled');
	}

	interface MonthGrid {
		year: number;
		month: number;
		title: string;
		days: Array<{
			date: string;
			dayNum: number;
			isCurrentMonth: boolean;
			reservations: ListingReservation[];
		}>;
	}

	function getMonthGrid(offset: number): MonthGrid {
		const base = new Date();
		base.setDate(1);
		base.setMonth(base.getMonth() + offset);

		const year = base.getFullYear();
		const month = base.getMonth();
		const title = `${RU_MONTHS[month]} ${year}`;

		const firstDayIndex = (new Date(year, month, 1).getDay() + 6) % 7;
		const daysInMonth = new Date(year, month + 1, 0).getDate();

		const days: MonthGrid['days'] = [];

		const prevMonthDays = new Date(year, month, 0).getDate();
		for (let i = firstDayIndex - 1; i >= 0; i--) {
			const dayNum = prevMonthDays - i;
			const d = new Date(year, month - 1, dayNum);
			const iso = toISODate(d);
			days.push({
				date: iso,
				dayNum,
				isCurrentMonth: false,
				reservations: []
			});
		}

		for (let dayNum = 1; dayNum <= daysInMonth; dayNum++) {
			const d = new Date(year, month, dayNum);
			const iso = toISODate(d);
			days.push({
				date: iso,
				dayNum,
				isCurrentMonth: true,
				reservations: getReservationsForDate(iso)
			});
		}

		const remaining = (7 - (days.length % 7)) % 7;
		for (let dayNum = 1; dayNum <= remaining; dayNum++) {
			const d = new Date(year, month + 1, dayNum);
			const iso = toISODate(d);
			days.push({
				date: iso,
				dayNum,
				isCurrentMonth: false,
				reservations: []
			});
		}

		return { year, month, title, days };
	}

	const monthGrid = $derived(getMonthGrid(viewMonthOffset));

	// Helper to get feed-specific styling for an iCal reservation
	function getReservationStyle(res: ListingReservation) {
		if (res.source === 'flickey') {
			return {
				label: 'Flickey',
				bg: 'bg-emerald-50 text-emerald-800 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800',
				cellBg: '!bg-emerald-50/70 dark:!bg-emerald-950/30 !border-emerald-200 dark:!border-emerald-800',
				dot: 'bg-emerald-500'
			};
		}
		if (res.source === 'manual_block') {
			return {
				label: 'Блокировка',
				bg: 'bg-amber-50 text-amber-800 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800',
				cellBg: '!bg-amber-50/70 dark:!bg-amber-950/30 !border-amber-200 dark:!border-amber-800',
				dot: 'bg-amber-500'
			};
		}

		// iCal import with custom feed color!
		const feed = calendarData?.sync_feeds?.find((f) => f.id === res.sync_feed_id);
		const theme = getPresetTheme(feed?.color);
		return {
			label: feed?.name || 'iCal фид',
			bg: `${theme.badgeBg} ${theme.badgeText} border ${theme.badgeBorder}`,
			cellBg: `${theme.calendarCellBg} ${theme.calendarCellBorder}`,
			dot: theme.dotClass
		};
	}
</script>

<svelte:head>
	<title>Календарь и iCal синхронизация — Flickey</title>
</svelte:head>

{#if guardChecked}
	<PageShell maxWidth="wide" background="white">
		<PageHeader
			title="Календарь и iCal"
			meta={listings.length > 0 ? `${listings.length} ${pluralRu(listings.length, ['объект', 'объекта', 'объектов'])}` : ''}
		>
			{#snippet actions()}
				<div class="flex items-center gap-2">
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
						href={resolve('/host/stats')}
						as="a"
					>
						<BarChart3 class="h-3.5 w-3.5 mr-1.5" />
						Статистика
					</Button>
					{#if selectedListingId}
						<Button
							variant="solid"
							tone="neutral"
							size="sm"
							radius="xl"
							onclick={handleSyncNow}
							disabled={syncing}
						>
							<RefreshCw class="h-3.5 w-3.5 mr-1.5 {syncing ? 'animate-spin' : ''}" />
							{syncing ? 'Синхронизация...' : 'Обновить сейчас'}
						</Button>
					{/if}
				</div>
			{/snippet}
		</PageHeader>

		{#if listings.length === 0}
			<EmptyState
				icon={Building2}
				title="У вас пока нет объектов"
				description="Создайте первое объявление, чтобы настраивать календарь доступности и синхронизацию с Airbnb, Booking и Авито"
				actionLabel="Создать объявление"
				actionHref={resolve('/listings/new')}
			/>
		{:else}
			<!-- ═══════════════════ PROPERTY SELECTOR BAR ═══════════════════ -->
			<div class="mb-6 space-y-3">
				<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
					<div class="flex items-center gap-2">
						<span class="text-xs font-bold text-zinc-500 uppercase tracking-wider">Выберите объект:</span>
						<span class="rounded-full bg-zinc-100 dark:bg-zinc-800 px-2 py-0.5 text-xs font-semibold text-zinc-600 dark:text-zinc-300">
							{filteredListings.length}
						</span>
					</div>

					{#if listings.length > 3}
						<div class="relative w-full sm:w-64">
							<Search class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-400" />
							<input
								type="text"
								placeholder="Поиск по адресу или названию..."
								bind:value={propertySearch}
								class="w-full rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/60 dark:bg-zinc-900 pl-8 pr-3 py-1.5 text-xs text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 focus:border-zinc-900 dark:focus:border-zinc-100 focus:bg-white focus:outline-none transition"
							/>
						</div>
					{/if}
				</div>

				<!-- Property Cards Carousel / Grid -->
				<div class="flex gap-3 overflow-x-auto pb-2 scrollbar-thin">
					{#each filteredListings as item (item.id)}
						{@const isSelected = item.id === selectedListingId}
						{@const fullAddr = formatListingAddressFull(item.address, item.location)}
						<button
							type="button"
							onclick={() => selectListing(item.id)}
							class="group flex shrink-0 items-center gap-3.5 rounded-2xl border p-2.5 pr-4 text-left transition-all duration-200 cursor-pointer
                                   {isSelected
									? 'border-zinc-900 bg-zinc-900 text-white shadow-md dark:border-white dark:bg-white dark:text-zinc-900'
									: 'border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900/60 text-zinc-900 dark:text-zinc-100 hover:border-zinc-300 dark:hover:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-800/40'}"
						>
							<div class="relative h-12 w-16 shrink-0 overflow-hidden rounded-xl bg-zinc-100 dark:bg-zinc-800">
								{#if item.images?.[0]}
									<img
										src={item.images[0]}
										alt=""
										class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
										loading="lazy"
									/>
								{:else}
									<div class="flex h-full w-full items-center justify-center text-zinc-400">
										<Building2 class="h-5 w-5" />
									</div>
								{/if}
							</div>
							<div class="min-w-0 max-w-[200px]">
								<p class="truncate text-xs font-bold leading-tight">{item.title}</p>
								<p class="mt-0.5 truncate text-[11px] {isSelected ? 'text-zinc-300 dark:text-zinc-600' : 'text-zinc-500 dark:text-zinc-400'}">
									{fullAddr}
								</p>
								<p class="mt-1 text-[11px] font-semibold {isSelected ? 'text-zinc-200 dark:text-zinc-700' : 'text-zinc-700 dark:text-zinc-300'}">
									{formatBYN(item.pricePerNight)}<span class="text-[10px] font-normal opacity-70"> / ночь</span>
								</p>
							</div>
						</button>
					{/each}
				</div>
			</div>

			<!-- ═══════════════════ SELECTED PROPERTY HEADER & TABS ═══════════════════ -->
			{#if selectedListing}
				{@const activeAddr = formatListingAddressFull(selectedListing.address, selectedListing.location)}
				<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/40 p-5 mb-6">
					<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
						<div class="flex items-center gap-3">
							<div class="relative h-14 w-20 shrink-0 overflow-hidden rounded-2xl bg-zinc-200 dark:bg-zinc-800 shadow-xs">
								{#if selectedListing.images?.[0]}
									<img src={selectedListing.images[0]} alt="" class="h-full w-full object-cover" />
								{/if}
							</div>
							<div class="min-w-0">
								<div class="flex items-center gap-2">
									<h2 class="text-base font-extrabold text-zinc-900 dark:text-zinc-100 truncate">
										{selectedListing.title}
									</h2>
									<a
										href={`/listings/${selectedListing.id}`}
										target="_blank"
										class="text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-200 transition-colors p-1"
										title="Открыть объявление"
									>
										<ArrowUpRight class="h-4 w-4" />
									</a>
								</div>
								<p class="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5 truncate">
									{activeAddr}
								</p>
								{#if calendarData?.sync_feeds}
									<div class="mt-1 flex items-center gap-2 text-[11px] text-zinc-500">
										<span class="inline-flex items-center gap-1 font-semibold text-zinc-700 dark:text-zinc-300">
											<Globe class="h-3 w-3" />
											{calendarData.sync_feeds.length} {pluralRu(calendarData.sync_feeds.length, ['фид', 'фида', 'фидов'])} синхронизации
										</span>
									</div>
								{/if}
							</div>
						</div>

						<!-- Tabs switcher -->
						<div class="flex items-center gap-1.5 rounded-2xl bg-white dark:bg-zinc-800/80 p-1 border border-zinc-200/80 dark:border-zinc-700 shadow-2xs self-start sm:self-auto">
							<button
								type="button"
								onclick={() => { activeTab = 'calendar'; }}
								class="flex items-center gap-1.5 rounded-xl px-3.5 py-1.5 text-xs font-bold transition-all cursor-pointer
                                       {activeTab === 'calendar'
										? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
										: 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200'}"
							>
								<CalendarIcon class="h-3.5 w-3.5" />
								Календарь
							</button>
							<button
								type="button"
								onclick={() => { activeTab = 'sync'; }}
								class="flex items-center gap-1.5 rounded-xl px-3.5 py-1.5 text-xs font-bold transition-all cursor-pointer
                                       {activeTab === 'sync'
										? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
										: 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200'}"
							>
								<Globe class="h-3.5 w-3.5" />
								iCal Синхронизация
								{#if calendarData?.sync_feeds && calendarData.sync_feeds.length > 0}
									<span class="ml-0.5 rounded-full px-1.5 py-0.2 text-[10px] {activeTab === 'sync' ? 'bg-white/20 text-white dark:bg-zinc-900/20 dark:text-zinc-900' : 'bg-zinc-100 dark:bg-zinc-700 text-zinc-600 dark:text-zinc-300'}">
										{calendarData.sync_feeds.length}
									</span>
								{/if}
							</button>
							<button
								type="button"
								onclick={() => { activeTab = 'overview'; }}
								class="flex items-center gap-1.5 rounded-xl px-3.5 py-1.5 text-xs font-bold transition-all cursor-pointer
                                       {activeTab === 'overview'
										? 'bg-zinc-900 text-white shadow-xs dark:bg-white dark:text-zinc-900'
										: 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200'}"
							>
								<Layers class="h-3.5 w-3.5" />
								Все объекты
							</button>
						</div>
					</div>
				</div>

				<!-- ═══════════════════ TAB 1: CALENDAR & BOOKINGS ═══════════════════ -->
				{#if activeTab === 'calendar'}
					{#if loading && !calendarData}
						<div class="flex h-72 items-center justify-center">
							<RefreshCw class="h-7 w-7 animate-spin text-zinc-400" />
						</div>
					{:else}
						<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
							<!-- Left Column: Visual Month Calendar -->
							<div class="lg:col-span-2 rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-6 shadow-sm">
								<!-- Month Navigation Header -->
								<div class="mb-5 flex items-center justify-between">
									<div class="flex items-center gap-3">
										<h3 class="text-base sm:text-lg font-bold text-zinc-900 dark:text-zinc-100">
											{monthGrid.title}
										</h3>
										{#if viewMonthOffset !== 0}
											<button
												type="button"
												onclick={() => { viewMonthOffset = 0; }}
												class="rounded-lg border border-zinc-200 dark:border-zinc-700 px-2 py-0.5 text-xs font-medium text-zinc-600 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition"
											>
												Сегодня
											</button>
										{/if}
									</div>

									<div class="flex items-center gap-1.5">
										<button
											type="button"
											onclick={() => { viewMonthOffset -= 1; }}
											class="flex h-8 w-8 items-center justify-center rounded-xl border border-zinc-200 dark:border-zinc-700 text-zinc-600 dark:text-zinc-300 hover:border-zinc-900 dark:hover:border-zinc-100 hover:text-zinc-900 dark:hover:text-white transition cursor-pointer"
											aria-label="Предыдущий месяц"
										>
											<ChevronLeft class="h-4 w-4" />
										</button>
										<button
											type="button"
											onclick={() => { viewMonthOffset += 1; }}
											class="flex h-8 w-8 items-center justify-center rounded-xl border border-zinc-200 dark:border-zinc-700 text-zinc-600 dark:text-zinc-300 hover:border-zinc-900 dark:hover:border-zinc-100 hover:text-zinc-900 dark:hover:text-white transition cursor-pointer"
											aria-label="Следующий месяц"
										>
											<ChevronRight class="h-4 w-4" />
										</button>
									</div>
								</div>

								<!-- Weekday names -->
								<div class="grid grid-cols-7 gap-1 text-center text-xs font-bold text-zinc-400 mb-2">
									{#each WEEKDAYS as wd}
										<div class="py-1">{wd}</div>
									{/each}
								</div>

								<!-- Calendar Days Grid -->
								<div class="grid grid-cols-7 gap-1.5">
									{#each monthGrid.days as day (day.date)}
										{@const hasRes = day.reservations.length > 0}
										{@const primaryRes = day.reservations[0]}
										{@const resStyle = hasRes ? getReservationStyle(primaryRes) : null}
										{@const isToday = day.date === todayISO}
										{@const isSelectedRange = (blockStartDate && day.date === blockStartDate) || (blockEndDate && day.date === blockEndDate) || (blockStartDate && blockEndDate && day.date > blockStartDate && day.date < blockEndDate)}

										<button
											type="button"
											onclick={() => handleDayClick(day.date)}
											class="min-h-[64px] sm:min-h-[72px] rounded-2xl p-1.5 text-left transition-all relative border flex flex-col justify-between cursor-pointer
                                                   {day.isCurrentMonth
													? 'bg-zinc-50/50 dark:bg-zinc-800/30 border-zinc-100 dark:border-zinc-800/80 hover:border-zinc-300 dark:hover:border-zinc-700'
													: 'bg-zinc-100/30 dark:bg-zinc-900/30 border-transparent opacity-35'}
                                                   {isSelectedRange ? '!ring-2 !ring-zinc-900 dark:!ring-zinc-100' : ''}
                                                   {resStyle ? resStyle.cellBg : ''}"
										>
											<div class="flex items-center justify-between w-full">
												<span class="text-xs font-bold {isToday ? 'rounded-full bg-zinc-900 text-white dark:bg-white dark:text-zinc-900 px-1.5 py-0.5' : 'text-zinc-700 dark:text-zinc-300'}">
													{day.dayNum}
												</span>
												{#if hasRes}
													<span class="h-2 w-2 rounded-full {resStyle?.dot}"></span>
												{/if}
											</div>

											{#if hasRes}
												<div class="mt-1 flex flex-col gap-0.5 w-full">
													{#each day.reservations as r}
														{@const style = getReservationStyle(r)}
														<span
															class="truncate rounded px-1 py-0.5 text-[9px] font-semibold {style.bg}"
															title="{r.guest_name ? `${r.guest_name}: ` : ''}{r.note || style.label}"
														>
															{r.guest_name || r.note || style.label}
														</span>
													{/each}
												</div>
											{/if}
										</button>
									{/each}
								</div>

								<!-- Dynamic Legend with Custom iCal Colors -->
								<div class="mt-6 border-t border-zinc-100 dark:border-zinc-800 pt-4">
									<p class="text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-2.5">Обозначения в календаре:</p>
									<div class="flex flex-wrap items-center gap-4 text-xs text-zinc-600 dark:text-zinc-300">
										<div class="flex items-center gap-1.5">
											<span class="h-3 w-3 rounded-full bg-emerald-500"></span>
											<span>Flickey бронь</span>
										</div>
										<div class="flex items-center gap-1.5">
											<span class="h-3 w-3 rounded-full bg-amber-500"></span>
											<span>Ручная блокировка</span>
										</div>

										<!-- Dynamic connected feeds in legend -->
										{#if calendarData?.sync_feeds && calendarData.sync_feeds.length > 0}
											{#each calendarData.sync_feeds as feed (feed.id)}
												{@const preset = getPresetTheme(feed.color)}
												<div class="flex items-center gap-1.5">
													<span class="h-3 w-3 rounded-full {preset.dotClass}"></span>
													<span class="font-medium">{feed.name}</span>
												</div>
											{/each}
										{:else}
											<div class="flex items-center gap-1.5 text-zinc-400">
												<span class="h-3 w-3 rounded-full bg-indigo-500"></span>
												<span>Внешний iCal</span>
											</div>
										{/if}
									</div>
								</div>
							</div>

							<!-- Right Column: Manual Block Form & Active Bookings -->
							<div class="space-y-6">
								<!-- Manual Block Form -->
								<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm">
									<h4 class="text-sm font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-2 mb-1">
										<Lock class="h-4 w-4 text-zinc-700 dark:text-zinc-300" />
										Заблокировать даты
									</h4>
									<p class="text-xs text-zinc-500 dark:text-zinc-400 mb-4">
										Закройте даты для ремонта, личных нужд или бронирования вне сайта. Кликните по датам в календаре для автозаполнения.
									</p>

									<form onsubmit={handleBlockDates} class="space-y-3">
										<div class="grid grid-cols-2 gap-2.5">
											<div>
												<label for="block-start" class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1">
													Заезд
												</label>
												<input
													id="block-start"
													type="date"
													bind:value={blockStartDate}
													min={todayISO}
													required
													class="w-full rounded-xl border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800 px-2.5 py-1.5 text-xs text-zinc-900 dark:text-zinc-100 focus:border-zinc-900 dark:focus:border-zinc-100 focus:bg-white focus:outline-none"
												/>
											</div>
											<div>
												<label for="block-end" class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1">
													Выезд
												</label>
												<input
													id="block-end"
													type="date"
													bind:value={blockEndDate}
													min={blockStartDate || todayISO}
													required
													class="w-full rounded-xl border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800 px-2.5 py-1.5 text-xs text-zinc-900 dark:text-zinc-100 focus:border-zinc-900 dark:focus:border-zinc-100 focus:bg-white focus:outline-none"
												/>
											</div>
										</div>

										<div>
											<label for="block-note" class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1">
												Причина / Примечание
											</label>
											<input
												id="block-note"
												type="text"
												placeholder="Ремонт, бронь по телефону..."
												bind:value={blockNote}
												class="w-full rounded-xl border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800 px-3 py-1.5 text-xs text-zinc-900 dark:text-zinc-100 focus:border-zinc-900 dark:focus:border-zinc-100 focus:bg-white focus:outline-none"
											/>
										</div>

										<div class="grid grid-cols-2 gap-2.5">
											<div>
												<label for="block-guest" class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1">
													Имя гостя (опц.)
												</label>
												<input
													id="block-guest"
													type="text"
													placeholder="Алексей"
													bind:value={blockGuestName}
													class="w-full rounded-xl border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800 px-2.5 py-1.5 text-xs text-zinc-900 dark:text-zinc-100 focus:border-zinc-900 dark:focus:border-zinc-100 focus:bg-white focus:outline-none"
												/>
											</div>
											<div>
												<label for="block-phone" class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1">
													Телефон (опц.)
												</label>
												<input
													id="block-phone"
													type="tel"
													placeholder="+375 29..."
													bind:value={blockGuestPhone}
													class="w-full rounded-xl border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800 px-2.5 py-1.5 text-xs text-zinc-900 dark:text-zinc-100 focus:border-zinc-900 dark:focus:border-zinc-100 focus:bg-white focus:outline-none"
												/>
											</div>
										</div>

										<div class="flex justify-end pt-2">
											<Button type="submit" size="sm" radius="lg" disabled={blockSubmitting}>
												{blockSubmitting ? 'Блокировка...' : 'Заблокировать'}
											</Button>
										</div>
									</form>
								</div>

								<!-- Active bookings & manual blocks list -->
								<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-5 shadow-sm">
									<div class="flex items-center justify-between mb-3">
										<h4 class="text-sm font-bold text-zinc-900 dark:text-zinc-100">
											Ближайшие брони
										</h4>
										{#if calendarData?.reservations}
											<span class="text-xs text-zinc-400 font-semibold">
												{calendarData.reservations.length}
											</span>
										{/if}
									</div>

									{#if !calendarData?.reservations || calendarData.reservations.length === 0}
										<div class="py-8 text-center text-xs text-zinc-400">
											Нет активных бронирований
										</div>
									{:else}
										<div class="divide-y divide-zinc-100 dark:divide-zinc-800 max-h-72 overflow-y-auto pr-1 scrollbar-thin">
											{#each calendarData.reservations as res (res.id)}
												{@const style = getReservationStyle(res)}
												<div class="py-2.5 text-xs flex items-center justify-between gap-3">
													<div class="min-w-0 flex-1">
														<div class="flex items-center gap-2">
															<span class="rounded-md px-1.5 py-0.5 text-[10px] font-bold {style.bg}">
																{style.label}
															</span>
															<span class="font-bold text-zinc-900 dark:text-zinc-100">
																{res.start_date} – {res.end_date}
															</span>
														</div>
														{#if res.guest_name || res.note || res.guest_phone}
															<p class="mt-1 text-zinc-500 dark:text-zinc-400 truncate text-[11px]">
																{#if res.guest_name}<span>{res.guest_name}</span>{/if}
																{#if res.guest_phone}<span> ({res.guest_phone})</span>{/if}
																{#if res.note}<span> · {res.note}</span>{/if}
															</p>
														{/if}
													</div>

													{#if res.source === 'manual_block'}
														<button
															type="button"
															onclick={() => handleUnblock(res.id)}
															class="text-rose-600 dark:text-rose-400 hover:text-rose-700 font-semibold text-[11px] underline shrink-0 cursor-pointer"
														>
															Разблокировать
														</button>
													{/if}
												</div>
											{/each}
										</div>
									{/if}
								</div>
							</div>
						</div>
					{/if}

				<!-- ═══════════════════ TAB 2: ICAL SYNCHRONIZATION ═══════════════════ -->
				{:else if activeTab === 'sync'}
					<div class="space-y-6 max-w-4xl">
						<!-- Outbound Feed (Export) -->
						<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-6 shadow-sm space-y-4">
							<div class="flex items-start justify-between gap-4">
								<div>
									<h4 class="text-base font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
										<Globe class="h-4 w-4 text-zinc-700 dark:text-zinc-300" />
										Экспорт календаря Flickey (RFC 5545)
									</h4>
									<p class="mt-1 text-xs text-zinc-500 dark:text-zinc-400 leading-relaxed">
										Скопируйте эту уникальную ссылку и добавьте её в настройки внешних площадок (Airbnb, Booking.com, Авито, Суточно.ру). При бронировании на Flickey даты автоматически закроются на других сайтах.
									</p>
								</div>
							</div>

							<div class="flex items-center gap-2">
								<input
									type="text"
									readonly
									value={calendarData?.export_url ?? ''}
									class="flex-1 rounded-2xl border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800 px-3.5 py-2.5 text-xs font-mono text-zinc-700 dark:text-zinc-200 select-all focus:outline-none"
								/>
								<Button size="sm" radius="xl" onclick={() => copyExportUrl()} variant="outline">
									{#if copied}
										<Check class="h-3.5 w-3.5 mr-1 text-emerald-600" />
										Скопировано
									{:else}
										<Copy class="h-3.5 w-3.5 mr-1" />
										Копировать ссылку
									{/if}
								</Button>
							</div>

							<!-- How-to Guide Cards -->
							<div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2">
								<div class="rounded-2xl border border-zinc-100 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-800/40 p-3.5 text-xs">
									<p class="font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-1.5">
										<span class="h-2 w-2 rounded-full bg-rose-500"></span> Airbnb
									</p>
									<p class="mt-1 text-[11px] text-zinc-500 dark:text-zinc-400">
										Объявления → Цены и доступность → Синхронизация календарей → Импортировать календарь
									</p>
								</div>
								<div class="rounded-2xl border border-zinc-100 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-800/40 p-3.5 text-xs">
									<p class="font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-1.5">
										<span class="h-2 w-2 rounded-full bg-sky-500"></span> Booking.com
									</p>
									<p class="mt-1 text-[11px] text-zinc-500 dark:text-zinc-400">
										Экстранет → Тарифы и номера → Синхронизировать календари → Добавить подключение
									</p>
								</div>
								<div class="rounded-2xl border border-zinc-100 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-800/40 p-3.5 text-xs">
									<p class="font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-1.5">
										<span class="h-2 w-2 rounded-full bg-purple-500"></span> Авито / Суточно
									</p>
									<p class="mt-1 text-[11px] text-zinc-500 dark:text-zinc-400">
										Управление объявлением → Календарь занятости → Синхронизация iCal → Добавить ссылку
									</p>
								</div>
							</div>
						</div>

						<!-- Inbound Feeds (Import) -->
						<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-6 shadow-sm space-y-6">
							<div class="flex items-center justify-between">
								<div>
									<h4 class="text-base font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
										<RefreshCw class="h-4 w-4 text-zinc-700 dark:text-zinc-300" />
										Входящие календари (Импорт в Flickey)
									</h4>
									<p class="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
										Flickey автоматически опрашивает внешние календари каждые 15 минут и блокирует даты в календаре.
									</p>
								</div>
								<Button size="sm" radius="xl" variant="outline" onclick={handleSyncNow} disabled={syncing}>
									<RefreshCw class="h-3.5 w-3.5 mr-1.5 {syncing ? 'animate-spin' : ''}" />
									{syncing ? 'Синхронизация...' : 'Обновить сейчас'}
								</Button>
							</div>

							<!-- Connected feeds list -->
							{#if !calendarData?.sync_feeds || calendarData.sync_feeds.length === 0}
								<div class="rounded-2xl border border-dashed border-zinc-200 dark:border-zinc-800 py-8 text-center text-xs text-zinc-400">
									Нет подключённых внешних календарей. Добавьте первый фид ниже.
								</div>
							{:else}
								<div class="divide-y divide-zinc-100 dark:divide-zinc-800">
									{#each calendarData.sync_feeds as feed (feed.id)}
										{@const preset = getPresetTheme(feed.color)}
										<div class="flex flex-col sm:flex-row sm:items-center justify-between py-3.5 gap-3">
											<div class="min-w-0 flex-1">
												<div class="flex items-center gap-2.5">
													<!-- Color Dot / Color Picker Trigger -->
													<div class="relative">
														<button
															type="button"
															onclick={() => {
																editingFeedColorId = editingFeedColorId === feed.id ? null : feed.id;
															}}
															class="flex items-center gap-1 rounded-full p-1 border border-zinc-200 dark:border-zinc-700 hover:scale-110 transition cursor-pointer"
															title="Изменить цвет календаря"
														>
															<span class="h-3.5 w-3.5 rounded-full {preset.dotClass}"></span>
														</button>

														<!-- Inline Color Picker Dropdown -->
														{#if editingFeedColorId === feed.id}
															<div class="absolute left-0 top-full z-20 mt-1.5 flex gap-1.5 rounded-2xl border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 p-2 shadow-lg">
																{#each COLOR_PRESETS as p}
																	<button
																		type="button"
																		onclick={() => handleUpdateFeedColor(feed.id, p.key)}
																		class="h-5 w-5 rounded-full {p.dotClass} hover:scale-125 transition-transform cursor-pointer ring-offset-1 {feed.color === p.key ? 'ring-2 ring-zinc-900 dark:ring-white' : ''}"
																		title={p.label}
																	></button>
																{/each}
															</div>
														{/if}
													</div>

													<span class="font-bold text-sm text-zinc-900 dark:text-zinc-100">{feed.name}</span>

													{#if feed.sync_status === 'success'}
														<span class="flex items-center gap-1 rounded-md bg-emerald-50 dark:bg-emerald-950/40 px-2 py-0.5 text-[10px] font-semibold text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800">
															<CheckCircle2 class="h-3 w-3" /> Успешно
														</span>
													{:else if feed.sync_status === 'failed'}
														<span class="flex items-center gap-1 rounded-md bg-rose-50 dark:bg-rose-950/40 px-2 py-0.5 text-[10px] font-semibold text-rose-700 dark:text-rose-300 border border-rose-200 dark:border-rose-800">
															<AlertTriangle class="h-3 w-3" /> Ошибка
														</span>
													{:else if feed.sync_status === 'syncing'}
														<span class="flex items-center gap-1 rounded-md bg-sky-50 dark:bg-sky-950/40 px-2 py-0.5 text-[10px] font-semibold text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800">
															<RefreshCw class="h-3 w-3 animate-spin" /> Синхронизация
														</span>
													{:else}
														<span class="rounded-md bg-zinc-100 dark:bg-zinc-800 px-2 py-0.5 text-[10px] text-zinc-600 dark:text-zinc-400">
															Ожидание
														</span>
													{/if}
												</div>

												<p class="mt-1 font-mono text-[11px] text-zinc-400 truncate">{feed.feed_url}</p>

												{#if feed.error_message}
													<p class="mt-0.5 text-[11px] text-rose-600 dark:text-rose-400 truncate">{feed.error_message}</p>
												{/if}

												{#if feed.last_synced_at}
													<p class="mt-1 text-[11px] text-zinc-400">
														Последняя синхронизация: {new Date(feed.last_synced_at).toLocaleString('ru')}
													</p>
												{/if}
											</div>

											<div class="flex items-center gap-2 self-end sm:self-center">
												<button
													type="button"
													onclick={() => handleDeleteFeed(feed.id)}
													aria-label="Удалить фид"
													class="p-2 text-zinc-400 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/30 rounded-xl transition cursor-pointer"
													title="Удалить фид"
												>
													<Trash2 class="h-4 w-4" />
												</button>
											</div>
										</div>
									{/each}
								</div>
							{/if}

							<!-- Add new feed form with custom color picker -->
							<div class="rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/70 dark:bg-zinc-800/40 p-5 space-y-3">
								<h5 class="text-xs font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-1.5">
									<Plus class="h-3.5 w-3.5" /> Подключить новый календарь
								</h5>

								<form onsubmit={handleAddFeed} class="space-y-3">
									<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
										<div>
											<label for="new-feed-name" class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1">
												Название
											</label>
											<input
												id="new-feed-name"
												type="text"
												placeholder="Напр. Airbnb или Авито"
												bind:value={newFeedName}
												required
												class="w-full rounded-xl border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 px-3 py-2 text-xs text-zinc-900 dark:text-zinc-100 focus:border-zinc-900 dark:focus:border-zinc-100 focus:outline-none"
											/>
										</div>

										<div class="sm:col-span-2">
											<label for="new-feed-url" class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1">
												Ссылка на iCal (.ics)
											</label>
											<input
												id="new-feed-url"
												type="url"
												placeholder="https://.../calendar.ics"
												bind:value={newFeedUrl}
												required
												class="w-full rounded-xl border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 px-3 py-2 text-xs text-zinc-900 dark:text-zinc-100 focus:border-zinc-900 dark:focus:border-zinc-100 focus:outline-none"
											/>
										</div>
									</div>

									<!-- Color Preset Picker -->
									<div>
										<span class="block text-[11px] font-semibold text-zinc-600 dark:text-zinc-400 mb-1.5">
											Цвет отображения в календаре:
										</span>
										<div class="flex flex-wrap items-center gap-2">
											{#each COLOR_PRESETS as preset}
												<button
													type="button"
													onclick={() => { newFeedColor = preset.key; }}
													class="flex items-center gap-1.5 rounded-xl border px-2.5 py-1 text-xs font-semibold transition cursor-pointer
                                                           {newFeedColor === preset.key
															? 'border-zinc-900 dark:border-white bg-white dark:bg-zinc-900 shadow-xs scale-105'
															: 'border-zinc-200 dark:border-zinc-700 bg-transparent text-zinc-500 hover:border-zinc-400'}"
												>
													<span class="h-3 w-3 rounded-full {preset.dotClass}"></span>
													<span>{preset.label}</span>
												</button>
											{/each}
										</div>
									</div>

									<div class="flex justify-end pt-2">
										<Button type="submit" size="sm" radius="lg" disabled={feedSubmitting}>
											{feedSubmitting ? 'Подключение...' : 'Подключить календарь'}
										</Button>
									</div>
								</form>
							</div>
						</div>
					</div>

				<!-- ═══════════════════ TAB 3: ALL PROPERTIES SUMMARY ═══════════════════ -->
				{:else if activeTab === 'overview'}
					<div class="rounded-3xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden shadow-sm">
						<div class="p-5 border-b border-zinc-100 dark:border-zinc-800">
							<h3 class="text-base font-bold text-zinc-900 dark:text-zinc-100">
								Сводка по всем объектам
							</h3>
							<p class="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5">
								Быстрый обзор и управление iCal-синхронизацией по всей вашей недвижимости
							</p>
						</div>

						<div class="divide-y divide-zinc-100 dark:divide-zinc-800">
							{#each listings as item (item.id)}
								{@const fullAddr = formatListingAddressFull(item.address, item.location)}
								{@const isCurrent = item.id === selectedListingId}
								<div class="flex flex-col sm:flex-row sm:items-center justify-between p-4 sm:px-6 gap-4 hover:bg-zinc-50/60 dark:hover:bg-zinc-800/30 transition">
									<div class="flex items-center gap-3.5 min-w-0">
										<div class="relative h-14 w-20 shrink-0 overflow-hidden rounded-2xl bg-zinc-100 dark:bg-zinc-800">
											{#if item.images?.[0]}
												<img src={item.images[0]} alt="" class="h-full w-full object-cover" />
											{/if}
										</div>
										<div class="min-w-0">
											<p class="font-bold text-sm text-zinc-900 dark:text-zinc-100 truncate">{item.title}</p>
											<p class="text-xs text-zinc-500 dark:text-zinc-400 truncate mt-0.5">{fullAddr}</p>
											<p class="text-xs font-semibold text-zinc-700 dark:text-zinc-300 mt-1">
												{formatBYN(item.pricePerNight)} <span class="text-zinc-400 font-normal">/ ночь</span>
											</p>
										</div>
									</div>

									<div class="flex items-center gap-2 shrink-0">
										<Button
											variant="outline"
											size="sm"
											radius="xl"
											onclick={() => copyExportUrl(calendarApi.getExportFeedUrl(item.id))}
										>
											<Copy class="h-3.5 w-3.5 mr-1" />
											Копировать iCal
										</Button>
										<Button
											variant={isCurrent ? 'solid' : 'outline'}
											tone={isCurrent ? 'primary' : 'neutral'}
											size="sm"
											radius="xl"
											onclick={() => {
												selectListing(item.id);
												activeTab = 'calendar';
											}}
										>
											<CalendarIcon class="h-3.5 w-3.5 mr-1" />
											Открыть календарь
										</Button>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			{/if}
		{/if}
	</PageShell>
{/if}
