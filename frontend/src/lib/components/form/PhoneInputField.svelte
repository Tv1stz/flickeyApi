<!-- src/lib/components/form/PhoneInputField.svelte -->
<script lang="ts">
	import { tick, onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { Check, AlertCircle, ChevronDown, Search } from 'lucide-svelte';
	import {
		countries,
		popularCountries,
		getDefaultCountry,
		parsePhoneNumber,
		type Country
	} from '$lib/config/countries';

	interface Props {
		id: string;
		label?: string;
		required?: boolean;
		value?: string;
		country?: Country;
		error?: string;
		touched?: boolean;
		disabled?: boolean;
		helpText?: string;
		valid?: boolean;
		onInput: (phone: string, country: Country) => void;
		onBlur?: () => void;
	}

	let {
		id,
		label = 'Номер телефона',
		required = false,
		value = '',
		country = getDefaultCountry(),
		error = '',
		touched = false,
		disabled = false,
		helpText = '',
		valid = false,
		onInput,
		onBlur
	}: Props = $props();

	let showCountryPicker = $state(false);
	let searchQuery = $state('');
	let activeIndex = $state(-1);
	let inputEl = $state<HTMLInputElement | null>(null);
	let containerEl = $state<HTMLDivElement | null>(null);
	let dropdownEl = $state<HTMLDivElement | null>(null);
	let searchInputEl = $state<HTMLInputElement | null>(null);
	let buttonEl = $state<HTMLButtonElement | null>(null);

	const hasError = $derived(Boolean(touched && error));
	const hasSuccess = $derived(Boolean(touched && !hasError && valid));

	// ═══════════════════════════════════════════════════════════════
	// COUNTRY FILTERING
	// ═══════════════════════════════════════════════════════════════

	const filteredCountries = $derived.by(() => {
		if (!searchQuery.trim()) {
			const popular = countries.filter((c) => popularCountries.includes(c.code));
			const others = countries.filter((c) => !popularCountries.includes(c.code));
			return [...popular, ...others];
		}

		const query = searchQuery.toLowerCase().trim();
		return countries.filter(
			(c) =>
				c.name.toLowerCase().includes(query) ||
				c.dialCode.includes(query) ||
				c.code.toLowerCase().includes(query)
		);
	});

	// ═══════════════════════════════════════════════════════════════
	// PHONE FORMATTING
	// ═══════════════════════════════════════════════════════════════

	function formatWithMask(digits: string, format: string): string {
		let result = '';
		let digitIndex = 0;

		for (const char of format) {
			if (digitIndex >= digits.length) break;

			if (char === '#') {
				result += digits[digitIndex];
				digitIndex++;
			} else {
				result += char;
			}
		}

		return result;
	}

	const displayValue = $derived(formatWithMask(value, country.format));
	const placeholder = $derived(country.format.replace(/#/g, '0'));

	// ═══════════════════════════════════════════════════════════════
	// INPUT HANDLERS
	// ═══════════════════════════════════════════════════════════════

	function handleInput(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const rawValue = input.value;

		let digits = rawValue.replace(/\D/g, '');
		digits = digits.slice(0, country.maxLength);

		onInput(digits, country);

		tick().then(() => {
			if (inputEl) {
				const formatted = formatWithMask(digits, country.format);
				inputEl.setSelectionRange(formatted.length, formatted.length);
			}
		});
	}

	function handleKeydown(e: KeyboardEvent) {
		const allowedKeys = [
			'Backspace',
			'Delete',
			'Tab',
			'Escape',
			'Enter',
			'ArrowLeft',
			'ArrowRight',
			'Home',
			'End'
		];

		if (allowedKeys.includes(e.key)) return;
		if (e.ctrlKey || e.metaKey) return; // ✅ Исправлено: metaKey вместо metaCmd

		if (!/^\d$/.test(e.key)) {
			e.preventDefault();
			return;
		}

		if (value.length >= country.maxLength) {
			e.preventDefault();
		}
	}

	function handlePaste(e: ClipboardEvent) {
		e.preventDefault();

		const text = e.clipboardData?.getData('text') ?? '';
		if (!text) return;

		const parsed = parsePhoneNumber(text, country);

		onInput(parsed.localNumber, parsed.country);

		tick().then(() => {
			if (inputEl) {
				const formatted = formatWithMask(parsed.localNumber, parsed.country.format);
				inputEl.setSelectionRange(formatted.length, formatted.length);
			}
		});
	}

	// ═══════════════════════════════════════════════════════════════
	// COUNTRY PICKER
	// ═══════════════════════════════════════════════════════════════

	function toggleCountryPicker(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();

		if (disabled) return;

		if (showCountryPicker) {
			closeCountryPicker();
		} else {
			openCountryPicker();
		}
	}

	function openCountryPicker() {
		showCountryPicker = true;
		searchQuery = '';
		activeIndex = -1;

		tick().then(() => {
			searchInputEl?.focus();
		});
	}

	function closeCountryPicker() {
		showCountryPicker = false;
		searchQuery = '';
		activeIndex = -1;
	}

	function selectCountry(c: Country) {
		const newValue = value.slice(0, c.maxLength);
		onInput(newValue, c);
		closeCountryPicker();
		tick().then(() => inputEl?.focus());
	}

	function handlePickerKeydown(e: KeyboardEvent) {
		if (!showCountryPicker) return;

		switch (e.key) {
			case 'Escape':
				e.preventDefault();
				closeCountryPicker();
				buttonEl?.focus();
				break;
			case 'ArrowDown':
				e.preventDefault();
				activeIndex = Math.min(activeIndex + 1, filteredCountries.length - 1);
				scrollActiveIntoView();
				break;
			case 'ArrowUp':
				e.preventDefault();
				activeIndex = Math.max(activeIndex - 1, 0);
				scrollActiveIntoView();
				break;
			case 'Enter':
				e.preventDefault();
				if (activeIndex >= 0 && filteredCountries[activeIndex]) {
					selectCountry(filteredCountries[activeIndex]);
				}
				break;
		}
	}

	function scrollActiveIntoView() {
		if (!browser) return; // ✅ SSR-safe check

		tick().then(() => {
			const el = document.getElementById(`${id}-country-${activeIndex}`);
			el?.scrollIntoView({ block: 'nearest' });
		});
	}

	function handleClickOutside(e: MouseEvent) {
		if (!showCountryPicker) return;

		const target = e.target as Node;

		if (dropdownEl && dropdownEl.contains(target)) return;
		if (buttonEl && buttonEl.contains(target)) return;

		closeCountryPicker();
	}

	// ✅ SSR-safe: onMount only runs on client
	onMount(() => {
		document.addEventListener('click', handleClickOutside, true);

		return () => {
			document.removeEventListener('click', handleClickOutside, true);
		};
	});

	// ═══════════════════════════════════════════════════════════════
	// STYLES
	// ═══════════════════════════════════════════════════════════════

	const containerBaseClasses =
		'relative flex items-stretch min-h-14 border rounded-2xl ' +
		'bg-white transition-all duration-200';

	const containerOkClasses =
		'border-zinc-300 hover:border-zinc-400 hover:shadow-sm ' +
		'focus-within:ring-2 focus-within:ring-zinc-900 focus-within:ring-offset-1 focus-within:border-zinc-900';

	const containerSuccessClasses =
		'border-green-500 ' +
		'focus-within:ring-2 focus-within:ring-green-500 focus-within:ring-offset-1 focus-within:border-green-500';

	const containerErrClasses =
		'border-red-500 bg-red-50/50 ' +
		'focus-within:ring-2 focus-within:ring-red-500 focus-within:ring-offset-1 focus-within:border-red-600';

	const containerStateClasses = $derived(
		hasError ? containerErrClasses : hasSuccess ? containerSuccessClasses : containerOkClasses
	);

	const containerClasses = $derived(
		`${containerBaseClasses} ${containerStateClasses} ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`
	);

	const labelColorClass = $derived(
		hasError ? 'text-red-600' : hasSuccess ? 'text-green-600' : 'text-zinc-600'
	);

	const errorId = $derived(`${id}-error`);
	const helpId = $derived(`${id}-help`);
</script>

<div class="space-y-2" bind:this={containerEl}>
	<div class="relative">
		<div class={containerClasses}>
			<!-- Country selector button -->
			<button
				bind:this={buttonEl}
				type="button"
				class="flex shrink-0 touch-manipulation items-center gap-1.5 rounded-l-2xl border-r
                           border-zinc-200 px-3 transition-colors
                           hover:bg-zinc-50 active:bg-zinc-100 disabled:pointer-events-none"
				onclick={toggleCountryPicker}
				{disabled}
				aria-label="Выбрать страну"
				aria-haspopup="listbox"
				aria-expanded={showCountryPicker}
			>
				<span class="text-lg leading-none">{country.flag}</span>
				<span class="text-sm font-medium text-zinc-700 tabular-nums">
					{country.dialCode}
				</span>
				<ChevronDown
					class="h-4 w-4 text-zinc-400 transition-transform duration-200 {showCountryPicker
						? 'rotate-180'
						: ''}"
				/>
			</button>

			<!-- Phone input -->
			<div class="relative min-w-0 flex-1">
				<input
					bind:this={inputEl}
					{id}
					type="tel"
					inputmode="numeric"
					autocomplete="tel-national"
					class="h-full w-full bg-transparent pt-6 pr-10 pb-2 pl-3
                               text-base text-zinc-900 placeholder:text-zinc-400
                               focus:outline-none disabled:cursor-not-allowed"
					{placeholder}
					value={displayValue}
					{disabled}
					aria-invalid={hasError}
					aria-required={required}
					aria-describedby={helpText ? helpId : touched && error ? errorId : undefined}
					oninput={handleInput}
					onkeydown={handleKeydown}
					onpaste={handlePaste}
					onblur={() => onBlur?.()}
				/>

				<!-- Floating label -->
				<label
					for={id}
					class="pointer-events-none absolute top-1.5 left-3 bg-transparent
                               {labelColorClass} text-[11px] leading-none font-medium
                               transition-colors duration-200"
				>
					{label}{#if required}<span class="ml-0.5 text-red-500">*</span>{/if}
				</label>

				<!-- Status icons -->
				{#if hasSuccess}
					<div
						class="pointer-events-none absolute top-1/2 right-3 flex
                                   h-5 w-5 -translate-y-1/2 items-center justify-center rounded-full bg-green-500"
						aria-hidden="true"
					>
						<Check class="h-3.5 w-3.5 text-white" strokeWidth={3} />
					</div>
				{:else if hasError}
					<div
						class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-red-500"
						aria-hidden="true"
					>
						<AlertCircle class="h-5 w-5" />
					</div>
				{/if}
			</div>
		</div>

		<!-- Country picker dropdown -->
		{#if showCountryPicker}
			<div
				bind:this={dropdownEl}
				class="dropdown-surface animate-in fade-in slide-in-from-top-2 absolute top-full
				   left-0 z-50 mt-2 w-full duration-200"
				onkeydown={handlePickerKeydown}
				role="listbox"
				tabindex="-1"
				aria-label="Выберите страну"
			>
				<!-- Search -->
				<div class="border-b border-zinc-100 p-3">
					<div class="relative">
						<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-zinc-400" />
						<input
							bind:this={searchInputEl}
							type="text"
							class="h-10 w-full rounded-xl border border-zinc-300 pr-4 pl-10
                                       text-sm placeholder:text-zinc-400
                                       focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900
                                       focus:ring-offset-1 focus:outline-none"
							placeholder="Поиск страны..."
							bind:value={searchQuery}
						/>
					</div>
				</div>

				<!-- Countries list -->
				<div class="max-h-64 overflow-y-auto overscroll-contain p-1.5">
					{#if filteredCountries.length === 0}
						<div class="px-4 py-8 text-center text-sm text-zinc-500">Страна не найдена</div>
					{:else}
						{#each filteredCountries as c, i (c.code)}
							{@const isSelected = c.code === country.code}
							{@const isActive = i === activeIndex}

							{#if i === popularCountries.length && !searchQuery.trim()}
								<div class="mx-3 my-1.5 border-t border-zinc-100"></div>
							{/if}

							<button
								id={`${id}-country-${i}`}
								type="button"
								class="flex w-full touch-manipulation items-center gap-3 rounded-xl px-3 py-2.5
                                           text-left transition-colors duration-150
                                           {isActive || isSelected
									? 'bg-zinc-100'
									: 'hover:bg-zinc-50 active:bg-zinc-100'}"
								onclick={() => selectCountry(c)}
								onpointerenter={() => (activeIndex = i)}
								role="option"
								aria-selected={isSelected}
							>
								<span class="text-xl">{c.flag}</span>
								<span class="min-w-0 flex-1 truncate text-sm font-medium text-zinc-900">
									{c.name}
								</span>
								<span class="text-sm font-medium text-zinc-500 tabular-nums">
									{c.dialCode}
								</span>
								{#if isSelected}
									<div class="flex h-5 w-5 items-center justify-center rounded-full bg-zinc-900">
										<Check class="h-3 w-3 text-white" strokeWidth={3} />
									</div>
								{/if}
							</button>
						{/each}
					{/if}
				</div>
			</div>
		{/if}
	</div>

	<!-- Help text -->
	{#if helpText && !hasError}
		<p id={helpId} class="px-0.5 text-xs leading-relaxed text-zinc-500">
			{helpText}
		</p>
	{/if}

	<!-- Error message -->
	{#if touched && error}
		<div
			id={errorId}
			class="flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3 py-2.5"
		>
			<AlertCircle class="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
			<p class="text-xs leading-relaxed font-medium text-red-700">{error}</p>
		</div>
	{/if}
</div>
