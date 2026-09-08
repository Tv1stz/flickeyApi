<script lang="ts">
	import { CITY_OPTIONS, type CityOption } from '../taxonomy';
	import { matchesSearch } from '$lib/utils/searchNormalization';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface Props {
		value: string;
		isOpen: boolean;
		onSelect: (value: string) => void;
		onFocus: () => void;
		onComplete?: () => void;
		compact?: boolean;
		cities?: CityOption[];
		popularCities?: CityOption[];
		dropdownOffsetLeft?: number;
	}

	let {
		value,
		isOpen,
		onSelect,
		onFocus,
		onComplete,
		compact = false,
		cities,
		popularCities,
		dropdownOffsetLeft = 0
	}: Props = $props();

	let inputEl: HTMLInputElement | undefined;
	let inputValue = $derived.by(() => value);
	const availableCities = $derived.by(() => (cities ? cities : CITY_OPTIONS));
	const popularOptions = $derived.by(() => {
		if (popularCities && popularCities.length > 0) return popularCities;
		return availableCities.slice(0, 6);
	});
	const query = $derived(inputValue.trim());
	const filteredCities = $derived(
		query.length === 0
			? availableCities
			: availableCities.filter((c) => {
					if (matchesSearch(query, c.label)) return true;
					return c.region ? matchesSearch(query, c.region) : false;
				})
	);
	const showPopular = $derived(query.length === 0 && popularOptions.length > 0);
	const showNoMatches = $derived(query.length > 0 && filteredCities.length === 0);

	$effect(() => {
		if (isOpen && inputEl) inputEl.focus();
	});

	function handleInput(e: Event) {
		const nextValue = (e.target as HTMLInputElement).value;
		onSelect(nextValue);
	}

	function handleClear(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();

		onSelect('');

		inputEl?.focus();
	}
</script>

<div class="relative min-w-0 flex-1">
	<label
		class="flex cursor-pointer flex-col justify-center
               {compact
					? value.trim().length > 0
						? 'pl-3 pr-7 py-1'
						: 'px-3 py-1'
					: value.trim().length > 0
						? 'pl-6 pr-11 py-2.5'
						: 'px-6 py-2.5'}"
	>
		{#if !compact}
			<span class="mb-0.5 text-[14px] font-medium text-zinc-800 select-none"> Где </span>
		{/if}

		<input
			bind:this={inputEl}
			type="text"
			placeholder={compact ? 'Куда?' : 'Поиск направлений'}
			value={inputValue}
			oninput={handleInput}
			onfocus={onFocus}
			class="w-full truncate bg-transparent p-0 font-medium
                   text-zinc-900 placeholder:font-medium placeholder:text-zinc-400 focus:outline-none
                   {compact ? 'text-[12px]' : 'text-[14.5px]'}"
		/>
	</label>

	{#if value.trim().length > 0}
		<button
			type="button"
			onclick={handleClear}
			aria-label="Очистить локацию"
			class="absolute top-1/2 z-20 flex touch-manipulation items-center justify-center
                   rounded-full text-zinc-900 transition-all duration-200
                   hover:bg-zinc-300 hover:text-zinc-800 active:scale-[0.94]
                   {compact ? 'right-2 h-5 w-5 -translate-y-1/2' : 'right-4 h-7 w-7 -translate-y-1/2'}"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 20 20"
				fill="none"
				aria-hidden="true"
				class={compact ? 'h-3.5 w-3.5' : 'h-4 w-4'}
			>
				<path
					d="M5 5l10 10M15 5L5 15"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
				/>
			</svg>
		</button>
	{/if}

	{#if isOpen}
		<div
			class="dropdown-surface-lg absolute top-full z-[100] mt-3 w-[320px] py-3"
			style={`left: ${-dropdownOffsetLeft}px;`}
			transition:fly={{ y: -8, duration: 220, easing: cubicOut }}
			role="listbox"
			tabindex="0"
			aria-label="Выберите город"
			onpointerdown={(e) => {
				e.preventDefault();
				e.stopPropagation();
			}}
		>
			<div class="max-h-[320px] overflow-y-auto overscroll-contain px-2 py-1">
				{#if showPopular}
					<p class="px-4 pb-2 text-[12px] font-semibold tracking-wide text-zinc-400 uppercase">
						Популярные города
					</p>
					{#each popularOptions as city (city.id)}
						<button
							type="button"
							role="option"
							aria-selected={value === city.label}
							onclick={(e) => {
								e.stopPropagation();
								inputValue = city.label;
								onSelect(city.label);
								onComplete?.();
							}}
							class="flex w-full touch-manipulation items-center gap-4 rounded-2xl px-4 py-3
                               text-left transition-colors hover:bg-zinc-50 active:bg-zinc-100"
						>
							<div
								class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-zinc-100/80"
								aria-hidden="true"
							>
								<svg
									width="20"
									height="20"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									class="text-zinc-600"
									aria-hidden="true"
								>
									<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
									<circle cx="12" cy="10" r="3"></circle>
								</svg>
							</div>

							<div class="min-w-0">
								<p class="text-[15px] font-semibold text-zinc-900">{city.label}</p>
								{#if city.region}
									<p class="text-[13px] text-zinc-500">{city.region}</p>
								{/if}
							</div>
						</button>
					{/each}
				{:else if showNoMatches}
					<div class="px-4 py-6 text-center text-sm text-zinc-400">Нет совпадений</div>
				{:else}
					{#each filteredCities as city (city.id)}
						<button
							type="button"
							role="option"
							aria-selected={value === city.label}
							onclick={(e) => {
								e.stopPropagation();
								inputValue = city.label;
								onSelect(city.label);
								onComplete?.();
							}}
							class="flex w-full touch-manipulation items-center gap-4 rounded-2xl px-4 py-3
                               text-left transition-colors hover:bg-zinc-50 active:bg-zinc-100"
						>
							<div
								class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-zinc-100/80"
								aria-hidden="true"
							>
								<svg
									width="20"
									height="20"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									class="text-zinc-600"
									aria-hidden="true"
								>
									<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
									<circle cx="12" cy="10" r="3"></circle>
								</svg>
							</div>

							<div class="min-w-0">
								<p class="text-[15px] font-semibold text-zinc-900">{city.label}</p>
								{#if city.region}
									<p class="text-[13px] text-zinc-500">{city.region}</p>
								{/if}
							</div>
						</button>
					{/each}
				{/if}
			</div>
		</div>
	{/if}
</div>

<style></style>
