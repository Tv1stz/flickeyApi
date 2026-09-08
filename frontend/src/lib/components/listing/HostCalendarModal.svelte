<!-- src/lib/components/listing/HostCalendarModal.svelte -->
<script lang="ts">
	import Modal from '$lib/components/ui/Modal.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { calendarApi } from '$lib/api/calendar';
	import type { HostCalendarResponse, ListingReservation, ListingCalendarSync } from '$lib/types/calendar';
	import { toast } from '$lib/stores/toastStore';
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
		AlertTriangle
	} from 'lucide-svelte';

	interface Props {
		open: boolean;
		listingId: string;
		listingTitle: string;
		onClose: () => void;
	}

	let { open, listingId, listingTitle, onClose }: Props = $props();

	type Tab = 'calendar' | 'sync';
	let activeTab = $state<Tab>('calendar');

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
	let feedSubmitting = $state(false);

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

	$effect(() => {
		if (open && listingId) {
			loadData();
		}
	});

	async function loadData() {
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
		if (!blockStartDate || !blockEndDate) {
			toast.error('Укажите даты', 'Выберите дату начала и окончания блокировки');
			return;
		}
		if (blockEndDate <= blockStartDate) {
			toast.error('Неверный период', 'Дата выезда должна быть позже даты заезда');
			return;
		}

		try {
			blockSubmitting = true;
			await calendarApi.blockDates(listingId, {
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
			await loadData();
		} catch (err: any) {
			toast.error('Не удалось заблокировать', err?.message || 'Даты конфликтуют с существующей бронью');
		} finally {
			blockSubmitting = false;
		}
	}

	async function handleUnblock(resId: string) {
		try {
			await calendarApi.unblockDates(listingId, resId);
			toast.success('Даты разблокированы');
			await loadData();
		} catch (err: any) {
			toast.error('Ошибка', err?.message || 'Не удалось разблокировать даты');
		}
	}

	async function handleAddFeed(e: SubmitEvent) {
		e.preventDefault();
		if (!newFeedName.trim() || !newFeedUrl.trim()) {
			toast.error('Заполните поля', 'Укажите название и URL iCal фида');
			return;
		}

		try {
			feedSubmitting = true;
			await calendarApi.addSyncFeed(listingId, {
				name: newFeedName.trim(),
				feed_url: newFeedUrl.trim()
			});
			toast.success('Фид подключён', 'Синхронизация запущена в фоне');
			newFeedName = '';
			newFeedUrl = '';
			await loadData();
		} catch (err: any) {
			toast.error('Ошибка подключения фида', err?.message || 'Проверьте корректность URL');
		} finally {
			feedSubmitting = false;
		}
	}

	async function handleDeleteFeed(syncId: string) {
		try {
			await calendarApi.deleteSyncFeed(listingId, syncId);
			toast.success('Фид удалён');
			await loadData();
		} catch (err: any) {
			toast.error('Ошибка', err?.message || 'Не удалось удалить фид');
		}
	}

	async function handleSyncNow() {
		try {
			syncing = true;
			const res = await calendarApi.syncNow(listingId);
			toast.success('Синхронизация запущена', res.message || 'Календарь обновляется');
			setTimeout(async () => {
				await loadData();
				syncing = false;
			}, 2500);
		} catch (err: any) {
			syncing = false;
			toast.error('Ошибка синхронизации', err?.message || 'Попробуйте позже');
		}
	}

	async function copyExportUrl() {
		if (!calendarData?.export_url) return;
		try {
			await navigator.clipboard.writeText(calendarData.export_url);
			copied = true;
			toast.success('Ссылка скопирована', 'Вставьте её на внешнем сервисе');
			setTimeout(() => { copied = false; }, 3000);
		} catch {
			toast.error('Не удалось скопировать ссылку');
		}
	}

	// Month grid calculation
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

	function getSourceBadge(source: string) {
		switch (source) {
			case 'flickey':
				return { label: 'Flickey', bg: 'bg-emerald-100 text-emerald-800 border-emerald-200' };
			case 'manual_block':
				return { label: 'Блокировка', bg: 'bg-amber-100 text-amber-800 border-amber-200' };
			case 'ical_import':
				return { label: 'iCal Фид', bg: 'bg-indigo-100 text-indigo-800 border-indigo-200' };
			default:
				return { label: source, bg: 'bg-zinc-100 text-zinc-800 border-zinc-200' };
		}
	}
</script>

<Modal {open} {onClose} title="Календарь и синхронизация" maxWidth="lg">
	<div class="space-y-6">
		<!-- Subtitle -->
		<div class="border-b border-zinc-100 pb-3">
			<h4 class="text-sm font-semibold text-zinc-900 truncate">{listingTitle}</h4>
			<p class="text-xs text-zinc-500">Управление бронированиями, ручными блокировками и внешними iCal-фидами</p>
		</div>

		<!-- Tabs -->
		<div class="flex items-center gap-2 border-b border-zinc-200">
			<button
				type="button"
				onclick={() => { activeTab = 'calendar'; }}
				class="flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-semibold transition-colors
                       {activeTab === 'calendar'
						? 'border-zinc-900 text-zinc-900'
						: 'border-transparent text-zinc-500 hover:text-zinc-800'}"
			>
				<CalendarIcon class="h-4 w-4" />
				Календарь и брони
			</button>
			<button
				type="button"
				onclick={() => { activeTab = 'sync'; }}
				class="flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-semibold transition-colors
                       {activeTab === 'sync'
						? 'border-zinc-900 text-zinc-900'
						: 'border-transparent text-zinc-500 hover:text-zinc-800'}"
			>
				<Globe class="h-4 w-4" />
				Синхронизация iCal
				{#if calendarData?.sync_feeds && calendarData.sync_feeds.length > 0}
					<span class="ml-1 rounded-full bg-zinc-100 px-2 py-0.5 text-xs text-zinc-600 font-medium">
						{calendarData.sync_feeds.length}
					</span>
				{/if}
			</button>
		</div>

		{#if loading && !calendarData}
			<div class="flex h-64 items-center justify-center">
				<RefreshCw class="h-6 w-6 animate-spin text-zinc-400" />
			</div>
		{:else if activeTab === 'calendar'}
			<!-- ═══ TAB 1: CALENDAR & BOOKINGS ═══ -->
			<div class="space-y-6">
				<!-- Visual Month Calendar -->
				<div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
					<!-- Month Nav -->
					<div class="mb-4 flex items-center justify-between">
						<button
							type="button"
							onclick={() => { viewMonthOffset -= 1; }}
							class="flex h-8 w-8 items-center justify-center rounded-full border border-zinc-200 text-zinc-600 hover:border-zinc-900 hover:text-zinc-900 transition"
							aria-label="Предыдущий месяц"
						>
							<ChevronLeft class="h-4 w-4" />
						</button>
						<span class="text-sm font-bold text-zinc-800">{monthGrid.title}</span>
						<button
							type="button"
							onclick={() => { viewMonthOffset += 1; }}
							class="flex h-8 w-8 items-center justify-center rounded-full border border-zinc-200 text-zinc-600 hover:border-zinc-900 hover:text-zinc-900 transition"
							aria-label="Следующий месяц"
						>
							<ChevronRight class="h-4 w-4" />
						</button>
					</div>

					<!-- Days grid -->
					<div class="grid grid-cols-7 gap-1 text-center text-xs font-semibold text-zinc-400 mb-2">
						{#each WEEKDAYS as wd}
							<div class="py-1">{wd}</div>
						{/each}
					</div>

					<div class="grid grid-cols-7 gap-1">
						{#each monthGrid.days as day (day.date)}
							{@const hasRes = day.reservations.length > 0}
							{@const firstRes = day.reservations[0]}
							<div
								class="min-h-[52px] rounded-xl p-1 text-left text-xs transition-colors relative border
                                       {day.isCurrentMonth ? 'bg-zinc-50/50 border-zinc-100' : 'bg-zinc-100/40 border-transparent opacity-40'}
                                       {hasRes && firstRes.source === 'flickey' ? '!bg-emerald-50 !border-emerald-200' : ''}
                                       {hasRes && firstRes.source === 'manual_block' ? '!bg-amber-50 !border-amber-200' : ''}
                                       {hasRes && firstRes.source === 'ical_import' ? '!bg-indigo-50 !border-indigo-200' : ''}"
							>
								<span class="block font-semibold {day.date === todayISO ? 'text-zinc-900 underline' : 'text-zinc-700'}">
									{day.dayNum}
								</span>
								{#if hasRes}
									<div class="mt-1 flex flex-col gap-0.5">
										{#each day.reservations as r}
											{@const badge = getSourceBadge(r.source)}
											<span
												class="inline-block truncate rounded px-1 py-0.2 text-[9px] font-medium {badge.bg}"
												title="{r.guest_name ? `${r.guest_name}: ` : ''}{r.note || badge.label}"
											>
												{r.guest_name || r.note || badge.label}
											</span>
										{/each}
									</div>
								{/if}
							</div>
						{/each}
					</div>

					<!-- Legend -->
					<div class="mt-4 flex flex-wrap items-center gap-4 border-t border-zinc-100 pt-3 text-xs text-zinc-600">
						<div class="flex items-center gap-1.5">
							<span class="h-3 w-3 rounded-full bg-emerald-500"></span>
							<span>Flickey бронь</span>
						</div>
						<div class="flex items-center gap-1.5">
							<span class="h-3 w-3 rounded-full bg-amber-500"></span>
							<span>Ручная блокировка</span>
						</div>
						<div class="flex items-center gap-1.5">
							<span class="h-3 w-3 rounded-full bg-indigo-500"></span>
							<span>Внешний iCal</span>
						</div>
					</div>
				</div>

				<!-- Manual Block Form -->
				<div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
					<h5 class="text-sm font-bold text-zinc-900 flex items-center gap-2 mb-3">
						<Lock class="h-4 w-4 text-zinc-700" />
						Заблокировать даты вручную
					</h5>
					<form onsubmit={handleBlockDates} class="space-y-3">
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							<div>
								<label for="block-start" class="block text-xs font-medium text-zinc-600 mb-1">Дата заезда (включительно)</label>
								<input
									id="block-start"
									type="date"
									bind:value={blockStartDate}
									min={todayISO}
									required
									class="w-full rounded-xl border border-zinc-300 px-3 py-2 text-sm text-zinc-900 focus:border-zinc-900 focus:outline-none"
								/>
							</div>
							<div>
								<label for="block-end" class="block text-xs font-medium text-zinc-600 mb-1">Дата выезда (день освобождения)</label>
								<input
									id="block-end"
									type="date"
									bind:value={blockEndDate}
									min={blockStartDate || todayISO}
									required
									class="w-full rounded-xl border border-zinc-300 px-3 py-2 text-sm text-zinc-900 focus:border-zinc-900 focus:outline-none"
								/>
							</div>
						</div>
						<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
							<div>
								<label for="block-note" class="block text-xs font-medium text-zinc-600 mb-1">Причина / Примечание</label>
								<input
									id="block-note"
									type="text"
									placeholder="Ремонт, бронь по телефону"
									bind:value={blockNote}
									class="w-full rounded-xl border border-zinc-300 px-3 py-2 text-sm text-zinc-900 focus:border-zinc-900 focus:outline-none"
								/>
							</div>
							<div>
								<label for="block-guest" class="block text-xs font-medium text-zinc-600 mb-1">Имя гостя (опционально)</label>
								<input
									id="block-guest"
									type="text"
									placeholder="Иван"
									bind:value={blockGuestName}
									class="w-full rounded-xl border border-zinc-300 px-3 py-2 text-sm text-zinc-900 focus:border-zinc-900 focus:outline-none"
								/>
							</div>
							<div>
								<label for="block-phone" class="block text-xs font-medium text-zinc-600 mb-1">Телефон гостя (опц.)</label>
								<input
									id="block-phone"
									type="tel"
									placeholder="+375 29 000-00-00"
									bind:value={blockGuestPhone}
									class="w-full rounded-xl border border-zinc-300 px-3 py-2 text-sm text-zinc-900 focus:border-zinc-900 focus:outline-none"
								/>
							</div>
						</div>
						<div class="flex justify-end pt-1">
							<Button type="submit" disabled={blockSubmitting} size="sm">
								{blockSubmitting ? 'Блокировка...' : 'Заблокировать'}
							</Button>
						</div>
					</form>
				</div>

				<!-- Active reservations list -->
				<div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm">
					<h5 class="text-sm font-bold text-zinc-900 mb-3">Активные бронирования и блокировки</h5>
					{#if !calendarData?.reservations || calendarData.reservations.length === 0}
						<p class="text-xs text-zinc-400 py-3 text-center">Нет активных бронирований</p>
					{:else}
						<div class="divide-y divide-zinc-100 max-h-60 overflow-y-auto">
							{#each calendarData.reservations as res (res.id)}
								{@const badge = getSourceBadge(res.source)}
								<div class="flex items-center justify-between py-2.5 text-xs">
									<div class="min-w-0 flex-1">
										<div class="flex items-center gap-2">
											<span class="rounded px-1.5 py-0.5 font-semibold text-[10px] {badge.bg} border">
												{badge.label}
											</span>
											<span class="font-bold text-zinc-900">
												{res.start_date} – {res.end_date}
											</span>
										</div>
										{#if res.guest_name || res.note}
											<p class="mt-0.5 text-zinc-500 truncate">
												{res.guest_name ? `Гость: ${res.guest_name}` : ''}
												{res.guest_phone ? `(${res.guest_phone})` : ''}
												{res.note ? ` · ${res.note}` : ''}
											</p>
										{/if}
									</div>
									{#if res.source === 'manual_block'}
										<button
											type="button"
											onclick={() => handleUnblock(res.id)}
											class="ml-3 text-red-600 hover:text-red-700 font-semibold text-[11px] underline shrink-0"
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
		{:else}
			<!-- ═══ TAB 2: ICAL SYNCHRONIZATION ═══ -->
			<div class="space-y-6">
				<!-- Outbound Feed (Export) -->
				<div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm space-y-3">
					<div class="flex items-start justify-between gap-4">
						<div>
							<h5 class="text-sm font-bold text-zinc-900 flex items-center gap-2">
								<CalendarIcon class="h-4 w-4 text-zinc-700" />
								Экспорт календаря (RFC 5545)
							</h5>
							<p class="mt-1 text-xs text-zinc-500">
								Скопируйте эту ссылку и добавьте её на Avito, Суточно.ру или Airbnb. Даты, забронированные на Flickey, будут мгновенно закрываться на других площадках.
							</p>
						</div>
					</div>

					<div class="flex items-center gap-2">
						<input
							type="text"
							readonly
							value={calendarData?.export_url ?? ''}
							class="flex-1 rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-xs font-mono text-zinc-700 select-all focus:outline-none"
						/>
						<Button size="sm" onclick={copyExportUrl} variant="outline">
							{#if copied}
								<Check class="h-3.5 w-3.5 mr-1 text-emerald-600" />
								Скопировано
							{:else}
								<Copy class="h-3.5 w-3.5 mr-1" />
								Копировать
							{/if}
						</Button>
					</div>

					<div class="rounded-xl bg-zinc-50 p-3 text-[11px] text-zinc-500 flex items-center gap-2 border border-zinc-100">
						<Info class="h-4 w-4 shrink-0 text-zinc-400" />
						<span>Экспортный фид защищён от зацикливания и не передаёт персональные данные гостей.</span>
					</div>
				</div>

				<!-- Inbound Feeds (Import) -->
				<div class="rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm space-y-4">
					<div class="flex items-center justify-between">
						<div>
							<h5 class="text-sm font-bold text-zinc-900 flex items-center gap-2">
								<Globe class="h-4 w-4 text-zinc-700" />
								Входящие iCal фиды (Импорт)
							</h5>
							<p class="mt-0.5 text-xs text-zinc-500">
								Внешние календари автоматически опрашиваются каждые 15 минут.
							</p>
						</div>
						<Button size="sm" variant="outline" onclick={handleSyncNow} disabled={syncing}>
							<RefreshCw class="h-3.5 w-3.5 mr-1.5 {syncing ? 'animate-spin text-zinc-700' : ''}" />
							{syncing ? 'Синхронизация...' : 'Обновить сейчас'}
						</Button>
					</div>

					<!-- Connected feeds list -->
					{#if !calendarData?.sync_feeds || calendarData.sync_feeds.length === 0}
						<div class="rounded-xl border border-dashed border-zinc-200 py-6 text-center text-xs text-zinc-400">
							Нет подключённых фидов. Добавьте первый календарь ниже.
						</div>
					{:else}
						<div class="divide-y divide-zinc-100">
							{#each calendarData.sync_feeds as feed (feed.id)}
								<div class="flex items-center justify-between py-3 text-xs">
									<div class="min-w-0 flex-1 pr-4">
										<div class="flex items-center gap-2">
											<span class="font-bold text-zinc-900">{feed.name}</span>
											{#if feed.sync_status === 'success'}
												<span class="flex items-center gap-1 rounded bg-emerald-50 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700 border border-emerald-200">
													<CheckCircle2 class="h-3 w-3" /> Успешно
												</span>
											{:else if feed.sync_status === 'failed'}
												<span class="flex items-center gap-1 rounded bg-red-50 px-1.5 py-0.5 text-[10px] font-semibold text-red-700 border border-red-200">
													<AlertTriangle class="h-3 w-3" /> Ошибка
												</span>
											{:else if feed.sync_status === 'syncing'}
												<span class="flex items-center gap-1 rounded bg-blue-50 px-1.5 py-0.5 text-[10px] font-semibold text-blue-700 border border-blue-200">
													<RefreshCw class="h-3 w-3 animate-spin" /> Синхронизация
												</span>
											{:else}
												<span class="rounded bg-zinc-100 px-1.5 py-0.5 text-[10px] text-zinc-600">
													Ожидание
												</span>
											{/if}
										</div>
										<p class="mt-1 font-mono text-[10px] text-zinc-400 truncate">{feed.feed_url}</p>
										{#if feed.error_message}
											<p class="mt-0.5 text-[10px] text-red-600 truncate">{feed.error_message}</p>
										{/if}
										{#if feed.last_synced_at}
											<p class="mt-0.5 text-[10px] text-zinc-400">
												Посл. синхронизация: {new Date(feed.last_synced_at).toLocaleString('ru')}
											</p>
										{/if}
									</div>
									<button
										type="button"
										onclick={() => handleDeleteFeed(feed.id)}
										aria-label="Удалить фид"
										class="p-2 text-zinc-400 hover:text-red-600 transition-colors rounded-lg hover:bg-zinc-50"
									>
										<Trash2 class="h-4 w-4" />
									</button>
								</div>
							{/each}
						</div>
					{/if}

					<!-- Add feed form -->
					<div class="rounded-xl border border-zinc-200 bg-zinc-50/50 p-4">
						<h6 class="text-xs font-bold text-zinc-900 mb-2.5 flex items-center gap-1.5">
							<Plus class="h-3.5 w-3.5" /> Подключить новый календарь
						</h6>
						<form onsubmit={handleAddFeed} class="space-y-2.5">
							<div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
								<input
									type="text"
									placeholder="Название (напр. Авито)"
									bind:value={newFeedName}
									required
									class="rounded-xl border border-zinc-300 bg-white px-3 py-1.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none"
								/>
								<input
									type="url"
									placeholder="https://.../calendar.ics"
									bind:value={newFeedUrl}
									required
									class="sm:col-span-2 rounded-xl border border-zinc-300 bg-white px-3 py-1.5 text-xs text-zinc-900 focus:border-zinc-900 focus:outline-none"
								/>
							</div>
							<div class="flex justify-end pt-1">
								<Button type="submit" size="sm" disabled={feedSubmitting}>
									{feedSubmitting ? 'Подключение...' : 'Подключить фид'}
								</Button>
							</div>
						</form>
					</div>
				</div>
			</div>
		{/if}
	</div>
</Modal>
