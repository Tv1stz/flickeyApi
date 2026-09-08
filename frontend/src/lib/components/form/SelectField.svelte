<!-- src/lib/components/form/SelectField.svelte -->
<script lang="ts">
	import { ChevronDown, Check, AlertCircle } from 'lucide-svelte';
	import { tick, onMount } from 'svelte';

	export type SelectOption = {
		value: string;
		label: string;
		disabled?: boolean;
	};

	interface Props {
		id: string;
		label: string;
		required?: boolean;
		value?: string;
		options?: SelectOption[];
		placeholder?: string;
		disabled?: boolean;
		error?: string;
		touched?: boolean;
		helpText?: string;
		prefix?: string;
		suffix?: string;
		/** Компактный режим без floating label */
		compact?: boolean;
		onInput: (value: string) => void;
		onBlur: () => void;
	}

	let {
		id,
		label,
		required = false,
		value = '',
		options = [],
		placeholder = 'Выберите',
		disabled = false,
		error = '',
		touched = false,
		helpText = '',
		prefix = undefined,
		suffix = undefined,
		compact = false,
		onInput,
		onBlur
	}: Props = $props();

	let open = $state(false);
	let activeIndex = $state(-1);
	let rootEl = $state<HTMLDivElement | null>(null);
	let listEl = $state<HTMLUListElement | null>(null);
	let buttonEl = $state<HTMLButtonElement | null>(null);

	let justClosed = $state(false);

	const hasError = $derived(Boolean(touched && error));
	const selected = $derived(options.find((o) => o.value === value));

	// Стили для обычного режима
	const baseBtnClasses = $derived(
		compact
			? 'block w-full min-h-[56px] rounded-2xl border bg-white text-zinc-900 transition-all duration-200 ' +
					'px-4 py-4 text-base text-left relative touch-manipulation ' +
					'disabled:bg-zinc-50 disabled:text-zinc-400 disabled:cursor-not-allowed disabled:opacity-60'
			: 'block w-full min-h-[56px] rounded-2xl border bg-white text-zinc-900 transition-all duration-200 ' +
					'px-4 pt-7 pb-2 text-base text-left relative touch-manipulation ' +
					'disabled:bg-zinc-50 disabled:text-zinc-400 disabled:cursor-not-allowed disabled:opacity-60'
	);

	const okClasses =
		'border-zinc-300 hover:border-zinc-400 hover:shadow-sm ' +
		'focus:outline-none focus:ring-2 focus:ring-zinc-900 focus:ring-offset-1 focus:border-zinc-900';

	const errClasses =
		'border-red-500 bg-red-50/50 ' +
		'focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1 focus:border-red-600 focus:bg-white';

	const padLeft = $derived(prefix ? 'pl-12' : 'pl-4');
	const padRight = $derived(suffix ? 'pr-20' : 'pr-12');
	const buttonClasses = $derived(
		`${baseBtnClasses} ${padLeft} ${padRight} ${hasError ? errClasses : okClasses}`
	);
	const labelLeftClass = $derived(prefix ? 'left-12' : 'left-4');
	const labelColorClass = $derived(hasError ? 'text-red-600' : 'text-zinc-600');
	const affixColorClass = $derived(hasError ? 'text-red-600' : 'text-zinc-500');

	const describedBy = $derived(
		[helpText && !hasError ? `${id}-help` : null, hasError ? `${id}-error` : null]
			.filter(Boolean)
			.join(' ') || undefined
	);

	function scrollActiveIntoView() {
		if (typeof document === 'undefined') return;
		const el = document.getElementById(`${id}-opt-${activeIndex}`);
		el?.scrollIntoView({ block: 'nearest' });
	}

	function openMenu() {
		if (disabled || justClosed) return;

		open = true;

		const selectedIndex = options.findIndex((o) => o.value === value && !o.disabled);
		activeIndex =
			selectedIndex >= 0
				? selectedIndex
				: Math.max(
						0,
						options.findIndex((o) => !o.disabled)
					);

		tick().then(() => {
			listEl?.focus();
			scrollActiveIntoView();
		});
	}

	function closeMenu(returnFocus = false) {
		if (!open) return;

		open = false;
		activeIndex = -1;

		justClosed = true;
		setTimeout(() => {
			justClosed = false;
		}, 100);

		if (returnFocus) {
			buttonEl?.focus();
		}

		onBlur();
	}

	function toggleMenu() {
		if (open) {
			closeMenu(true);
		} else {
			openMenu();
		}
	}

	function selectOption(opt: SelectOption) {
		if (opt.disabled) return;
		onInput(opt.value);
		closeMenu(true);
	}

	function handleOutsideClick(e: PointerEvent | TouchEvent) {
		if (!open) return;

		const target = e.target as Node | null;
		if (rootEl && target && !rootEl.contains(target)) {
			closeMenu();
		}
	}

	function handleButtonKeydown(e: KeyboardEvent) {
		if (disabled) return;

		if (e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			if (!open) openMenu();
		}

		if (e.key === 'Escape' && open) {
			e.preventDefault();
			closeMenu(true);
		}
	}

	function handleListKeydown(e: KeyboardEvent) {
		if (!open) return;

		switch (e.key) {
			case 'Escape':
				e.preventDefault();
				closeMenu(true);
				break;
			case 'ArrowDown':
				e.preventDefault();
				moveActive(1);
				break;
			case 'ArrowUp':
				e.preventDefault();
				moveActive(-1);
				break;
			case 'Enter':
			case ' ':
				e.preventDefault();
				{
					const opt = options[activeIndex];
					if (opt) selectOption(opt);
				}
				break;
			case 'Tab':
				closeMenu();
				break;
		}
	}

	function moveActive(delta: number) {
		if (!options.length) return;

		let i = activeIndex;
		for (let step = 0; step < options.length; step++) {
			i = (i + delta + options.length) % options.length;
			if (!options[i]?.disabled) {
				activeIndex = i;
				scrollActiveIntoView();
				return;
			}
		}
	}

	onMount(() => {
		if (typeof document === 'undefined') return;

		document.addEventListener('pointerdown', handleOutsideClick, { capture: true });
		document.addEventListener('touchstart', handleOutsideClick, { capture: true, passive: true });

		return () => {
			document.removeEventListener('pointerdown', handleOutsideClick, { capture: true });
			document.removeEventListener('touchstart', handleOutsideClick, { capture: true });
		};
	});
</script>

<div class="space-y-2">
	<!-- Label (только для не-compact режима, внешний) -->
	{#if !compact}
		<label
			id={`${id}-label`}
			for={id}
			class="block text-sm font-medium transition-colors duration-200 {labelColorClass}"
		>
			{label}
			{#if required}
				<span class="ml-0.5 text-red-500">*</span>
			{/if}
		</label>
	{/if}

	<!-- Select wrapper -->
	<div class="relative" bind:this={rootEl}>
		{#if prefix}
			<span
				class="pointer-events-none absolute top-1/2 left-4 -translate-y-1/2 text-base font-medium {affixColorClass} z-10 transition-colors duration-200"
				aria-hidden="true"
			>
				{prefix}
			</span>
		{/if}

		<!-- Button trigger -->
		<button
			bind:this={buttonEl}
			{id}
			type="button"
			class={buttonClasses}
			aria-labelledby={`${id}-label`}
			aria-haspopup="listbox"
			aria-expanded={open}
			aria-describedby={describedBy}
			aria-controls={`${id}-listbox`}
			{disabled}
			onclick={toggleMenu}
			onkeydown={handleButtonKeydown}
		>
			<!-- Floating label (только для не-compact режима) -->
			{#if !compact}
				<span
					class="pointer-events-none absolute {labelLeftClass} top-2 bg-white {labelColorClass}
                        -ml-1 rounded px-1 text-[11px] leading-none font-medium transition-colors duration-200"
				>
					{label}{#if required}<span class="ml-0.5 text-red-500">*</span>{/if}
				</span>
			{/if}

			<!-- Selected value or placeholder -->
			<span class={selected ? 'text-zinc-900' : 'text-zinc-400'}>
				{selected ? selected.label : placeholder}
			</span>

			{#if suffix}
				<span
					class="pointer-events-none absolute top-1/2 right-12 -translate-y-1/2 text-base font-medium {affixColorClass} transition-colors duration-200"
					aria-hidden="true"
				>
					{suffix}
				</span>
			{/if}

			<!-- Chevron -->
			<span
				class="pointer-events-none absolute top-1/2 right-4 -translate-y-1/2 {affixColorClass} transition-transform duration-200 {open
					? 'rotate-180'
					: 'rotate-0'}"
			>
				<ChevronDown class="h-5 w-5" aria-hidden="true" />
			</span>
		</button>

		<!-- Dropdown -->
		{#if open}
			<ul
				id={`${id}-listbox`}
				bind:this={listEl}
				tabindex="0"
				role="listbox"
				aria-labelledby={`${id}-label`}
				class="dropdown-surface absolute z-50 mt-2 max-h-64 w-full overflow-x-hidden overflow-y-auto
                    overscroll-contain p-1.5
                    [-webkit-overflow-scrolling:touch]"
				style="touch-action: pan-y;"
				onkeydown={handleListKeydown}
				ontouchstart={(e) => e.stopPropagation()}
				ontouchmove={(e) => e.stopPropagation()}
			>
				{#each options as opt, i (opt.value)}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<li
						id={`${id}-opt-${i}`}
						role="option"
						aria-selected={opt.value === value}
						class="flex touch-manipulation items-center justify-between rounded-xl px-4 py-3 text-base transition-colors duration-150 select-none
                            {opt.disabled
							? 'cursor-not-allowed opacity-50'
							: 'cursor-pointer active:bg-zinc-200'}
                            {i === activeIndex && !opt.disabled ? 'bg-zinc-100' : 'bg-transparent'}
                            {opt.value === value
							? 'font-semibold text-zinc-900'
							: 'font-normal text-zinc-700'}"
						onclick={() => !opt.disabled && selectOption(opt)}
						onpointerenter={() => !opt.disabled && (activeIndex = i)}
					>
						<span>{opt.label}</span>
						{#if opt.value === value}
							<div class="flex h-5 w-5 items-center justify-center rounded-full bg-zinc-900">
								<Check class="h-3.5 w-3.5 text-white" strokeWidth={3} />
							</div>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	{#if helpText && !hasError}
		<p id={`${id}-help`} class="px-0.5 text-xs leading-relaxed text-zinc-500">
			{helpText}
		</p>
	{/if}

	{#if touched && error}
		<div
			id={`${id}-error`}
			class="flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3 py-2.5"
		>
			<AlertCircle class="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
			<p class="text-xs leading-relaxed font-medium text-red-700">{error}</p>
		</div>
	{/if}
</div>
