<script lang="ts">
	import { pluralRu } from '$lib/utils/format';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface Props {
		adults: number;
		children: number;
		isOpen: boolean;
		onSelect: (adults: number, children: number) => void;
		onFocus: () => void;
		compact?: boolean;
		dropdownOffsetRight?: number;
		reserveRight?: number;
	}

	let {
		adults = 0,
		children = 0,
		isOpen,
		onSelect,
		onFocus,
		compact = false,
		dropdownOffsetRight = 0,
		reserveRight = 0
	}: Props = $props();

	const validAdults = $derived(Number(adults) || 0);
	const validChildren = $derived(Number(children) || 0);
	const totalGuests = $derived(validAdults + validChildren);
	const adultsLabel = $derived(
		validAdults > 0 ? `${validAdults} ${pluralRu(validAdults, ['взрослый', 'взрослых', 'взрослых'])}` : ''
	);
	const childrenLabel = $derived(
		validChildren > 0 ? `${validChildren} ${pluralRu(validChildren, ['ребёнок', 'ребёнка', 'детей'])}` : ''
	);
	const label = $derived(
		totalGuests === 0
			? compact
				? 'Гости'
				: 'Добавьте гостей'
			: [adultsLabel, childrenLabel].filter(Boolean).join(', ')
	);

	function setAdults(next: number, e?: MouseEvent) {
		e?.preventDefault();
		e?.stopPropagation();
		if (next < 0) return;
		if (next + children > 20) return;
		if (next === 0 && children > 0) return;
		onSelect(next, children);
	}

	function setChildren(next: number, e?: MouseEvent) {
		e?.preventDefault();
		e?.stopPropagation();
		if (next < 0) return;
		if (adults + next > 20) return;
		if (next > 0 && adults === 0) return;
		onSelect(adults, next);
	}
</script>

<div class="relative min-w-0 flex-1">
	<button
		type="button"
		onclick={onFocus}
		style={reserveRight ? `padding-right:${reserveRight}px;` : undefined}
		class="flex w-full cursor-pointer flex-col items-start justify-center focus:outline-none
               {compact ? 'pl-2.5 pr-2 py-1' : 'pl-5 pr-4 py-2.5'}"
	>
		{#if !compact}
			<span class="mb-0.5 text-[14px] font-medium text-zinc-800 select-none"> Кто </span>
		{/if}
		<span
			class="w-full truncate text-left font-medium
                   {totalGuests === 0 ? 'text-zinc-400' : 'text-zinc-900'}
                   {compact ? 'text-[12px]' : 'text-[14.5px]'}"
		>
			{label}
		</span>
	</button>

	{#if isOpen}
		<div
			class="dropdown-surface-lg absolute top-full z-[100] mt-3 w-[320px] p-5"
			style={`right: ${-dropdownOffsetRight}px;`}
			transition:fly={{ y: -8, duration: 220, easing: cubicOut }}
			role="dialog"
			tabindex="-1"
			aria-label="Выбор количества гостей"
			aria-modal="false"
			onpointerdown={(e) => {
				e.preventDefault();
				e.stopPropagation();
			}}
		>
			<div class="space-y-4">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-[15px] font-semibold text-zinc-900">Взрослые</p>
						<p class="mt-0.5 text-[13px] text-zinc-500">От 18 лет</p>
					</div>

					<div class="flex items-center gap-3" role="group" aria-label="Количество взрослых">
						<button
							type="button"
							onclick={(e) => setAdults(adults - 1, e)}
							disabled={adults === 0 || (adults === 1 && children > 0)}
							aria-label="Уменьшить количество взрослых"
							class="flex h-[34px] w-[34px] touch-manipulation items-center justify-center rounded-full border border-zinc-300
                                   text-zinc-600 transition-colors hover:border-zinc-900 hover:text-zinc-900
                                   disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-zinc-300"
						>
							<svg width="12" height="2" viewBox="0 0 12 2" fill="none" aria-hidden="true">
								<path d="M1 1h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
							</svg>
						</button>

						<span
							class="w-6 text-center text-[15px] font-semibold text-zinc-900 tabular-nums"
							aria-live="polite"
							aria-atomic="true"
							aria-label="{adults} взрослых"
						>
							{adults}
						</span>

						<button
							type="button"
							onclick={(e) => setAdults(adults + 1, e)}
							disabled={adults + children >= 20}
							aria-label="Увеличить количество взрослых"
							class="flex h-[34px] w-[34px] touch-manipulation items-center justify-center rounded-full border border-zinc-300
                                   text-zinc-600 transition-colors hover:border-zinc-900 hover:text-zinc-900
                                   disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-zinc-300"
						>
							<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
								<path
									d="M6 1v10M1 6h10"
									stroke="currentColor"
									stroke-width="1.5"
									stroke-linecap="round"
								/>
							</svg>
						</button>
					</div>
				</div>

				<div class="h-px bg-zinc-200/60"></div>

				<div class="flex items-center justify-between">
					<div>
						<p class="text-[15px] font-semibold text-zinc-900">Дети</p>
						<p class="mt-0.5 text-[13px] text-zinc-500">До 18 лет</p>
					</div>

					<div class="flex items-center gap-3" role="group" aria-label="Количество детей">
						<button
							type="button"
							onclick={(e) => setChildren(children - 1, e)}
							disabled={children === 0}
							aria-label="Уменьшить количество детей"
							class="flex h-[34px] w-[34px] touch-manipulation items-center justify-center rounded-full border border-zinc-300
                                   text-zinc-600 transition-colors hover:border-zinc-900 hover:text-zinc-900
                                   disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-zinc-300"
						>
							<svg width="12" height="2" viewBox="0 0 12 2" fill="none" aria-hidden="true">
								<path d="M1 1h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
							</svg>
						</button>

						<span
							class="w-6 text-center text-[15px] font-semibold text-zinc-900 tabular-nums"
							aria-live="polite"
							aria-atomic="true"
							aria-label="{children} детей"
						>
							{children}
						</span>

						<button
							type="button"
							onclick={(e) => setChildren(children + 1, e)}
							disabled={adults + children >= 20 || adults === 0}
							aria-label="Увеличить количество детей"
							class="flex h-[34px] w-[34px] touch-manipulation items-center justify-center rounded-full border border-zinc-300
                                   text-zinc-600 transition-colors hover:border-zinc-900 hover:text-zinc-900
                                   disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-zinc-300"
						>
							<svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
								<path
									d="M6 1v10M1 6h10"
									stroke="currentColor"
									stroke-width="1.5"
									stroke-linecap="round"
								/>
							</svg>
						</button>
					</div>
				</div>
			</div>

			{#if totalGuests > 0}
				<button
					type="button"
					onclick={(e) => {
						e.stopPropagation();
						onSelect(0, 0);
					}}
					class="mt-5 w-full touch-manipulation rounded-2xl bg-zinc-100/80 py-2.5 text-[13px] font-semibold
                           text-zinc-800 transition-colors hover:bg-zinc-200 active:bg-zinc-300"
				>
					Очистить
				</button>
			{/if}
		</div>
	{/if}
</div>

<style></style>
