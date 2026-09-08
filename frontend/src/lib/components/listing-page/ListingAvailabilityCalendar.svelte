<!-- src/lib/components/listing-page/ListingAvailabilityCalendar.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { calendarApi } from '$lib/api/calendar';
	import type { AvailabilityRange } from '$lib/types/calendar';
	import { formatBYN, pluralRu } from '$lib/utils/format';
	import { ChevronLeft, ChevronRight, Calendar as CalendarIcon, Info } from 'lucide-svelte';

	interface Props {
		listingId: string;
		pricePerNight: number;
		minNights?: number;
		onDatesSelected?: (checkin: string | null, checkout: string | null, nights: number) => void;
	}

	let {
		listingId,
		pricePerNight,
		minNights = 1,
		onDatesSelected
	}: Props = $props();

	let ranges = $state<AvailabilityRange[]>([]);
	let loading = $state(true);
	let viewOffset = $state(0);
	let checkin = $state<string | null>(null);
	let checkout = $state<string | null>(null);
	let hoveredDate = $state<string | null>(null);
	let errorNotice = $state<string | null>(null);

	const RU_MONTHS = [
		'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
		'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
	];
	const RU_SHORT_MONTHS = [
		'янв', 'фев', 'мар', 'апр', 'мая', 'июн',
		'июл', 'авг', 'сен', 'окт', 'ноя', 'дек'
	];
	const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

	function toISODate(d: Date): string {
		const year = d.getFullYear();
		const month = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	const todayISO = $derived(toISODate(new Date()));

	onMount(async () => {
		try {
			loading = true;
			ranges = await calendarApi.getPublicAvailability(listingId);
		} catch (err) {
			console.error('Failed to load availability:', err);
		} finally {
			loading = false;
		}
	});

	function isDateBlocked(iso: string): boolean {
		if (iso < todayISO) return true;
		for (const r of ranges) {
			// RFC 5545 end_date is exclusive checkout date: [start_date, end_date)
			if (iso >= r.start_date && iso < r.end_date) {
				return true;
			}
		}
		return false;
	}

	function hasBlockedDateInRange(startISO: string, endISO: string): boolean {
		for (const r of ranges) {
			if (r.start_date < endISO && r.end_date > startISO) {
				return true;
			}
		}
		return false;
	}

	const nightsCount = $derived.by(() => {
		if (!checkin || !checkout) return 0;
		const d1 = new Date(checkin);
		const d2 = new Date(checkout);
		const diff = Math.round((d2.getTime() - d1.getTime()) / (1000 * 60 * 60 * 24));
		return diff > 0 ? diff : 0;
	});

	const totalPrice = $derived(nightsCount * pricePerNight);

	function formatDisplayDate(iso: string | null): string {
		if (!iso) return '';
		const [y, m, d] = iso.split('-').map(Number);
		return `${d} ${RU_SHORT_MONTHS[m - 1]}`;
	}

	function handleDateClick(dateISO: string, isBlocked: boolean) {
		if (isBlocked) return;
		errorNotice = null;

		if (!checkin || (checkin && checkout)) {
			checkin = dateISO;
			checkout = null;
			onDatesSelected?.(checkin, null, 0);
		} else {
			if (dateISO <= checkin) {
				checkin = dateISO;
				checkout = null;
				onDatesSelected?.(checkin, null, 0);
			} else {
				// Verify no booked dates inside [checkin, dateISO)
				if (hasBlockedDateInRange(checkin, dateISO)) {
					errorNotice = 'В выбранном диапазоне есть забронированные даты';
					return;
				}

				const d1 = new Date(checkin);
				const d2 = new Date(dateISO);
				const diff = Math.round((d2.getTime() - d1.getTime()) / (1000 * 60 * 60 * 24));
				if (diff < minNights) {
					errorNotice = `Минимальный срок бронирования — ${minNights} ${pluralRu(minNights, ['ночь', 'ночи', 'ночей'])}`;
					return;
				}

				checkout = dateISO;
				onDatesSelected?.(checkin, checkout, diff);
			}
		}
	}

	function handleClear() {
		checkin = null;
		checkout = null;
		errorNotice = null;
		onDatesSelected?.(null, null, 0);
	}

	interface MonthGrid {
		year: number;
		month: number;
		title: string;
		days: Array<{
			date: string;
			dayNum: number;
			isCurrentMonth: boolean;
			isBlocked: boolean;
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
				isBlocked: true
			});
		}

		for (let dayNum = 1; dayNum <= daysInMonth; dayNum++) {
			const d = new Date(year, month, dayNum);
			const iso = toISODate(d);
			days.push({
				date: iso,
				dayNum,
				isCurrentMonth: true,
				isBlocked: isDateBlocked(iso)
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
				isBlocked: true
			});
		}

		return { year, month, title, days };
	}

	const monthLeft = $derived(getMonthGrid(viewOffset));
	const monthRight = $derived(getMonthGrid(viewOffset + 1));

	function isRangeActive(dateISO: string): boolean {
		if (!checkin) return false;
		const end = checkout || hoveredDate;
		if (!end) return false;
		return dateISO > checkin && dateISO < end;
	}
</script>

<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 md:p-8 shadow-sm">
	<div class="mb-6 flex flex-wrap items-center justify-between gap-4">
		<div>
			<h3 class="text-xl font-bold tracking-tight text-zinc-900 flex items-center gap-2">
				<CalendarIcon class="h-5 w-5 text-zinc-700" />
				Доступность и бронирование
			</h3>
			<p class="mt-1 text-sm text-zinc-500">
				{checkin && checkout
					? `${formatDisplayDate(checkin)} – ${formatDisplayDate(checkout)} (${nightsCount} ${pluralRu(nightsCount, ['ночь', 'ночи', 'ночей'])})`
					: checkin
						? `Заезд: ${formatDisplayDate(checkin)}. Выберите дату отъезда`
						: 'Выберите даты заезда и отъезда для точного расчёта'}
			</p>
		</div>

		{#if checkin || checkout}
			<button
				type="button"
				onclick={handleClear}
				class="text-xs font-semibold text-zinc-600 underline hover:text-zinc-900"
			>
				Сбросить даты
			</button>
		{/if}
	</div>

	{#if errorNotice}
		<div class="mb-4 flex items-center gap-2 rounded-xl bg-amber-50 border border-amber-200 px-4 py-2.5 text-xs font-medium text-amber-800">
			<Info class="h-4 w-4 shrink-0 text-amber-600" />
			<span>{errorNotice}</span>
		</div>
	{/if}

	<!-- Month navigation bar -->
	<div class="mb-6 flex items-center justify-between">
		<button
			type="button"
			onclick={() => { if (viewOffset > 0) viewOffset -= 1; }}
			disabled={viewOffset <= 0}
			class="flex h-9 w-9 items-center justify-center rounded-full border border-zinc-200 text-zinc-700 hover:border-zinc-900 hover:text-zinc-900 disabled:opacity-30 disabled:cursor-not-allowed transition"
			aria-label="Предыдущий месяц"
		>
			<ChevronLeft class="h-4 w-4" />
		</button>
		<div class="text-sm font-semibold text-zinc-700">
			{monthLeft.title} — {monthRight.title}
		</div>
		<button
			type="button"
			onclick={() => { viewOffset += 1; }}
			class="flex h-9 w-9 items-center justify-center rounded-full border border-zinc-200 text-zinc-700 hover:border-zinc-900 hover:text-zinc-900 transition"
			aria-label="Следующий месяц"
		>
			<ChevronRight class="h-4 w-4" />
		</button>
	</div>

	<!-- Calendars side-by-side -->
	<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
		{#each [monthLeft, monthRight] as m (m.title)}
			<div>
				<div class="mb-3 text-center text-sm font-semibold text-zinc-900">
					{m.title}
				</div>
				<div class="grid grid-cols-7 gap-1 text-center text-xs text-zinc-400 font-medium mb-2">
					{#each WEEKDAYS as wd}
						<div class="py-1">{wd}</div>
					{/each}
				</div>
				<div class="grid grid-cols-7 gap-y-1">
					{#each m.days as day (day.date)}
						{@const isStart = day.date === checkin}
						{@const isEnd = day.date === checkout}
						{@const inRange = isRangeActive(day.date)}
						<div
							class="relative flex items-center justify-center py-1
                                   {inRange ? 'bg-zinc-100' : ''}
                                   {isStart && checkout ? 'bg-gradient-to-r from-transparent to-zinc-100' : ''}
                                   {isEnd ? 'bg-gradient-to-l from-transparent to-zinc-100' : ''}"
						>
							<button
								type="button"
								onclick={() => handleDateClick(day.date, day.isBlocked || !day.isCurrentMonth)}
								onmouseenter={() => { if (checkin && !checkout) hoveredDate = day.date; }}
								onmouseleave={() => { hoveredDate = null; }}
								disabled={day.isBlocked || !day.isCurrentMonth}
								class="relative z-10 flex h-9 w-9 items-center justify-center rounded-full text-xs font-semibold transition-all
                                       {isStart || isEnd
                                           ? 'bg-zinc-900 text-white shadow-md scale-105'
                                           : inRange
                                               ? 'text-zinc-900 hover:bg-zinc-200'
                                               : day.isBlocked || !day.isCurrentMonth
                                                   ? 'text-zinc-300 line-through cursor-not-allowed bg-zinc-50'
                                                   : 'text-zinc-800 hover:bg-zinc-100'}"
							>
								{day.dayNum}
							</button>
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</div>

	<!-- Price Calculation Breakdown -->
	{#if nightsCount > 0}
		<div class="mt-8 rounded-2xl bg-zinc-50 border border-zinc-200/80 p-5 flex flex-wrap items-center justify-between gap-4">
			<div>
				<div class="text-xs text-zinc-500 font-medium">Расчёт стоимости:</div>
				<div class="text-sm font-semibold text-zinc-800 mt-0.5">
					{formatBYN(pricePerNight)} × {nightsCount} {pluralRu(nightsCount, ['ночь', 'ночи', 'ночей'])}
				</div>
			</div>
			<div class="text-right">
				<div class="text-xs text-zinc-500 font-medium">Итого к оплате:</div>
				<div class="text-xl font-extrabold text-zinc-900">
					{formatBYN(totalPrice)}
				</div>
			</div>
		</div>
	{/if}
</div>
