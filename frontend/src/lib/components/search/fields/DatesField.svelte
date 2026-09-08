<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { pluralRu } from '$lib/utils/format';
	import { ChevronLeft, ChevronRight, X } from 'lucide-svelte';

	interface Props {
		checkin: string | null; // YYYY-MM-DD
		checkout: string | null; // YYYY-MM-DD
		isOpen: boolean;
		onSelect: (checkin: string | null, checkout: string | null) => void;
		onFocus: () => void;
		onComplete?: () => void;
		compact?: boolean;
		dropdownOffsetLeft?: number;
	}

	let {
		checkin,
		checkout,
		isOpen,
		onSelect,
		onFocus,
		onComplete,
		compact = false,
		dropdownOffsetLeft = 0
	}: Props = $props();

	// Navigation month state (0 = current month)
	let viewMonthOffset = $state(0);
	let hoveredDate = $state<string | null>(null);

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

	const todayISO = $derived.by(() => {
		const now = new Date();
		return toISODate(now);
	});

	function formatDisplayDate(iso: string | null): string {
		if (!iso) return '';
		const [y, m, d] = iso.split('-').map(Number);
		return `${d} ${RU_SHORT_MONTHS[m - 1]}`;
	}

	const nightsCount = $derived.by(() => {
		if (!checkin || !checkout) return 0;
		const d1 = new Date(checkin);
		const d2 = new Date(checkout);
		const diff = Math.round((d2.getTime() - d1.getTime()) / (1000 * 60 * 60 * 24));
		return diff > 0 ? diff : 0;
	});

	const label = $derived.by(() => {
		if (checkin && checkout) {
			return `${formatDisplayDate(checkin)} – ${formatDisplayDate(checkout)}`;
		}
		if (checkin) {
			return `${formatDisplayDate(checkin)} – ...`;
		}
		return compact ? 'Когда' : 'Укажите даты';
	});

	interface MonthData {
		year: number;
		month: number; // 0-11
		title: string;
		days: Array<{
			date: string; // YYYY-MM-DD
			dayNum: number;
			isCurrentMonth: boolean;
			isPast: boolean;
		}>;
	}

	function getMonthData(offset: number): MonthData {
		const base = new Date();
		base.setDate(1);
		base.setMonth(base.getMonth() + offset);

		const year = base.getFullYear();
		const month = base.getMonth();
		const title = `${RU_MONTHS[month]} ${year}`;

		// First day of month (0 = Sunday, 1 = Monday... convert to Mon=0, Sun=6)
		const firstDayIndex = (new Date(year, month, 1).getDay() + 6) % 7;
		const daysInMonth = new Date(year, month + 1, 0).getDate();

		const days: MonthData['days'] = [];

		// Leading empty slots from prev month
		const prevMonthDays = new Date(year, month, 0).getDate();
		for (let i = firstDayIndex - 1; i >= 0; i--) {
			const dayNum = prevMonthDays - i;
			const d = new Date(year, month - 1, dayNum);
			const iso = toISODate(d);
			days.push({
				date: iso,
				dayNum,
				isCurrentMonth: false,
				isPast: true
			});
		}

		// Days of current month
		for (let dayNum = 1; dayNum <= daysInMonth; dayNum++) {
			const d = new Date(year, month, dayNum);
			const iso = toISODate(d);
			days.push({
				date: iso,
				dayNum,
				isCurrentMonth: true,
				isPast: iso < todayISO
			});
		}

		// Trailing slots to complete rows of 7
		const remaining = (7 - (days.length % 7)) % 7;
		for (let dayNum = 1; dayNum <= remaining; dayNum++) {
			const d = new Date(year, month + 1, dayNum);
			const iso = toISODate(d);
			days.push({
				date: iso,
				dayNum,
				isCurrentMonth: false,
				isPast: true
			});
		}

		return { year, month, title, days };
	}

	const month1 = $derived(getMonthData(viewMonthOffset));
	const month2 = $derived(getMonthData(viewMonthOffset + 1));

	function handleDateClick(dateISO: string, isPast: boolean) {
		if (isPast) return;

		if (!checkin || (checkin && checkout)) {
			// Start new range selection
			onSelect(dateISO, null);
		} else {
			// checkin is chosen, picking checkout
			if (dateISO > checkin) {
				onSelect(checkin, dateISO);
				if (onComplete) onComplete();
			} else if (dateISO === checkin) {
				// Clicked same date: reset checkout
				onSelect(checkin, null);
			} else {
				// Clicked earlier date: make it the new checkin
				onSelect(dateISO, null);
			}
		}
	}

	function handleClear(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		onSelect(null, null);
	}

	function prevMonth(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		if (viewMonthOffset > 0) {
			viewMonthOffset -= 1;
		}
	}

	function nextMonth(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		viewMonthOffset += 1;
	}

	function isRangeActive(dateISO: string): boolean {
		if (!checkin) return false;
		const end = checkout || hoveredDate;
		if (!end) return false;
		return dateISO > checkin && dateISO < end;
	}
</script>

<div class="relative min-w-0 flex-1">
	<button
		type="button"
		onclick={onFocus}
		class="flex w-full cursor-pointer flex-col justify-center text-left focus:outline-none
               {compact
					? checkin
						? 'pl-3 pr-7 py-1'
						: 'px-3 py-1'
					: checkin
						? 'pl-6 pr-10 py-2.5'
						: 'px-6 py-2.5'}"
	>
		{#if !compact}
			<span class="mb-0.5 text-[14px] font-medium text-zinc-800 select-none"> Когда </span>
		{/if}

		<span
			class="w-full truncate font-medium
                   {!checkin ? 'text-zinc-400' : 'text-zinc-900'}
                   {compact ? 'text-[12px]' : 'text-[14.5px]'}"
		>
			{label}
		</span>
	</button>

	{#if checkin}
		<button
			type="button"
			onclick={handleClear}
			aria-label="Очистить даты"
			class="absolute top-1/2 right-2 z-20 flex h-7 w-7 -translate-y-1/2 touch-manipulation items-center justify-center
                   rounded-full text-zinc-900 transition-all duration-200
                   hover:bg-zinc-300 hover:text-zinc-800 active:scale-[0.94]
                   {compact ? 'right-2 h-5 w-5' : 'right-3 h-7 w-7'}"
		>
			<X class={compact ? 'h-3.5 w-3.5' : 'h-4 w-4'} />
		</button>
	{/if}

	{#if isOpen}
		<div
			class="dropdown-surface-lg absolute top-full z-[100] mt-3 w-[660px] p-6 shadow-2xl"
			style={`left: ${dropdownOffsetLeft}px;`}
			transition:fly={{ y: -8, duration: 220, easing: cubicOut }}
			role="dialog"
			tabindex="-1"
			aria-label="Выбор дат поездки"
			onpointerdown={(e) => {
				e.preventDefault();
				e.stopPropagation();
			}}
		>
			<!-- Header Navigation -->
			<div class="mb-4 flex items-center justify-between">
				<button
					type="button"
					onclick={prevMonth}
					disabled={viewMonthOffset <= 0}
					class="flex h-9 w-9 items-center justify-center rounded-full border border-zinc-200 text-zinc-700 transition hover:border-zinc-900 hover:text-zinc-900 disabled:opacity-30 disabled:cursor-not-allowed"
					aria-label="Предыдущий месяц"
				>
					<ChevronLeft class="h-4 w-4" />
				</button>
				<span class="text-sm font-semibold text-zinc-800">
					{nightsCount > 0
						? `${nightsCount} ${pluralRu(nightsCount, ['ночь', 'ночи', 'ночей'])}`
						: checkin
							? 'Выберите дату отъезда'
							: 'Выберите дату заезда'}
				</span>
				<button
					type="button"
					onclick={nextMonth}
					class="flex h-9 w-9 items-center justify-center rounded-full border border-zinc-200 text-zinc-700 transition hover:border-zinc-900 hover:text-zinc-900"
					aria-label="Следующий месяц"
				>
					<ChevronRight class="h-4 w-4" />
				</button>
			</div>

			<!-- 2 Month Calendars Side-by-Side -->
			<div class="grid grid-cols-2 gap-8">
				{#each [month1, month2] as m (m.title)}
					<div>
						<div class="mb-3 text-center text-sm font-semibold text-zinc-900">
							{m.title}
						</div>
						<div class="grid grid-cols-7 gap-1 text-center text-xs text-zinc-400 font-medium mb-1">
							{#each WEEKDAYS as wd}
								<div class="py-1">{wd}</div>
							{/each}
						</div>
						<div class="grid grid-cols-7 gap-y-1">
							{#each m.days as day (day.date)}
								{@const isSelectedStart = day.date === checkin}
								{@const isSelectedEnd = day.date === checkout}
								{@const isInRange = isRangeActive(day.date)}
								<div
									class="relative flex items-center justify-center py-1
                                           {isInRange ? 'bg-zinc-100' : ''}
                                           {isSelectedStart && checkout ? 'bg-gradient-to-r from-transparent to-zinc-100' : ''}
                                           {isSelectedEnd ? 'bg-gradient-to-l from-transparent to-zinc-100' : ''}"
								>
									<button
										type="button"
										onclick={() => handleDateClick(day.date, day.isPast || !day.isCurrentMonth)}
										onmouseenter={() => { if (checkin && !checkout) hoveredDate = day.date; }}
										onmouseleave={() => { hoveredDate = null; }}
										disabled={day.isPast || !day.isCurrentMonth}
										class="relative z-10 flex h-9 w-9 items-center justify-center rounded-full text-xs font-semibold transition-all
                                               {isSelectedStart || isSelectedEnd
                                                   ? 'bg-zinc-900 text-white shadow-md'
                                                   : isInRange
                                                       ? 'text-zinc-900 hover:bg-zinc-200'
                                                       : day.isPast || !day.isCurrentMonth
                                                           ? 'text-zinc-300 cursor-not-allowed'
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

			<!-- Footer info & clear -->
			<div class="mt-6 flex items-center justify-between border-t border-zinc-100 pt-4">
				<button
					type="button"
					onclick={handleClear}
					class="text-xs font-semibold text-zinc-600 underline hover:text-zinc-900"
				>
					Сбросить даты
				</button>
				{#if checkin && checkout}
					<div class="text-xs font-medium text-zinc-500">
						{formatDisplayDate(checkin)} – {formatDisplayDate(checkout)} ({nightsCount} {pluralRu(nightsCount, ['ночь', 'ночи', 'ночей'])})
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
