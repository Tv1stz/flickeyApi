<!-- src/lib/components/form/OTPInput.svelte -->
<script lang="ts">
	import { tick } from 'svelte';
	import { AlertCircle, Check } from 'lucide-svelte';

	interface Props {
		id?: string;
		length?: number;
		disabled?: boolean;
		error?: string;
		valid?: boolean;
		label?: string;
		onComplete?: (code: string) => void;
		onChange?: (code: string) => void;
	}

	let {
		id = 'otp',
		length = 6,
		disabled = false,
		error = '',
		valid = false,
		label = '',
		onComplete,
		onChange
	}: Props = $props();

	let inputs: HTMLInputElement[] = $state([]);
	let values: string[] = $state(Array.from({ length: 6 }, () => ''));
	let focusedIndex = $state(-1);

	// Keep values array in sync with length prop
	$effect(() => {
		if (values.length !== length) {
			values = Array(length).fill('');
		}
	});
	const hasError = $derived(Boolean(error));
	const isFilled = $derived(values.every((v) => v !== ''));

	function getCode(): string {
		return values.join('');
	}

	function focusInput(index: number) {
		if (index >= 0 && index < length) {
			tick().then(() => {
				inputs[index]?.focus();
				inputs[index]?.select();
			});
		}
	}

	function handleInput(index: number, e: Event) {
		const input = e.target as HTMLInputElement;
		const inputValue = input.value;

		if (inputValue.length > 1) {
			const digits = inputValue.replace(/\D/g, '').split('').slice(0, length);
			const newValues = [...values];

			digits.forEach((digit, i) => {
				if (index + i < length) {
					newValues[index + i] = digit;
				}
			});

			values = newValues;

			const nextIndex = Math.min(index + digits.length, length - 1);
			focusInput(digits.length === length ? length - 1 : nextIndex);
		} else {
			const digit = inputValue.replace(/\D/g, '');
			const newValues = [...values];
			newValues[index] = digit;
			values = newValues;

			if (digit && index < length - 1) {
				focusInput(index + 1);
			}
		}

		const code = getCode();
		onChange?.(code);

		if (code.length === length && !values.includes('')) {
			onComplete?.(code);
		}
	}

	function handleKeydown(index: number, e: KeyboardEvent) {
		if (e.key === 'Backspace') {
			e.preventDefault();
			const newValues = [...values];

			if (values[index]) {
				newValues[index] = '';
				values = newValues;
				onChange?.(getCode());
			} else if (index > 0) {
				newValues[index - 1] = '';
				values = newValues;
				onChange?.(getCode());
				focusInput(index - 1);
			}
		} else if (e.key === 'ArrowLeft' && index > 0) {
			e.preventDefault();
			focusInput(index - 1);
		} else if (e.key === 'ArrowRight' && index < length - 1) {
			e.preventDefault();
			focusInput(index + 1);
		} else if (e.key === 'Delete') {
			e.preventDefault();
			const newValues = [...values];
			newValues[index] = '';
			values = newValues;
			onChange?.(getCode());
		} else if (
			!/^\d$/.test(e.key) &&
			!['Tab', 'Enter'].includes(e.key) &&
			!e.ctrlKey &&
			!e.metaKey
		) {
			e.preventDefault();
		}
	}

	function handlePaste(e: ClipboardEvent) {
		e.preventDefault();
		const text = e.clipboardData?.getData('text') ?? '';
		const digits = text.replace(/\D/g, '').split('').slice(0, length);

		if (digits.length > 0) {
			const newValues = Array(length).fill('');
			digits.forEach((digit, i) => {
				newValues[i] = digit;
			});
			values = newValues;

			const code = getCode();
			onChange?.(code);

			focusInput(Math.min(digits.length, length - 1));

			if (digits.length === length) {
				onComplete?.(code);
			}
		}
	}

	function handleFocus(index: number) {
		focusedIndex = index;
		inputs[index]?.select();
	}

	function handleBlur() {
		focusedIndex = -1;
	}

	// Public methods
	export function focus() {
		const firstEmpty = values.findIndex((v) => !v);
		focusInput(firstEmpty >= 0 ? firstEmpty : 0);
	}

	export function clear() {
		values = Array(length).fill('');
		onChange?.('');
		focusInput(0);
	}

	export function setValue(code: string) {
		const digits = code.replace(/\D/g, '').split('').slice(0, length);
		const newValues = Array(length).fill('');
		digits.forEach((digit, i) => {
			newValues[i] = digit;
		});
		values = newValues;
		onChange?.(getCode());
	}

	// ═══════════════════════════════════════════════════════════════
	// STYLES (matching TextInputField patterns)
	// ═══════════════════════════════════════════════════════════════

	function getInputClasses(index: number): string {
		const base =
			'w-11 h-14 sm:w-12 sm:h-16 text-center text-xl sm:text-2xl font-semibold ' +
			'rounded-xl sm:rounded-2xl border bg-white text-zinc-900 ' +
			'transition-all duration-200 outline-none touch-manipulation ' +
			'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-zinc-50 ' +
			'caret-transparent';

		const isFocused = focusedIndex === index;
		const hasValue = values[index] !== '';

		if (hasError) {
			return `${base} border-red-500 bg-red-50/50 ${isFocused ? 'ring-2 ring-red-500 ring-offset-1' : ''}`;
		}

		if (isFocused) {
			return `${base} border-zinc-900 ring-2 ring-zinc-900 ring-offset-1`;
		}

		if (hasValue) {
			return `${base} border-zinc-400`;
		}

		return `${base} border-zinc-300 hover:border-zinc-400`;
	}
</script>

<div class="space-y-3">
	<!-- Label -->
	{#if label}
		<label for={`${id}-0`} class="block text-center text-sm font-medium text-zinc-700">
			{label}
		</label>
	{/if}

	<!-- OTP inputs -->
	<div class="flex justify-center gap-2 sm:gap-3">
		{#each Array.from({ length }, (_, i) => i) as i (i)}
			{@const isFocused = focusedIndex === i}
			{@const hasValueInCell = values[i] !== ''}

			<div class="relative">
				<input
					bind:this={inputs[i]}
					id={`${id}-${i}`}
					type="text"
					inputmode="numeric"
					autocomplete={i === 0 ? 'one-time-code' : 'off'}
					maxlength={length}
					class={getInputClasses(i)}
					value={values[i]}
					{disabled}
					oninput={(e) => handleInput(i, e)}
					onkeydown={(e) => handleKeydown(i, e)}
					onpaste={handlePaste}
					onfocus={() => handleFocus(i)}
					onblur={handleBlur}
					aria-label={`Цифра ${i + 1} из ${length}`}
				/>

				<!-- Cursor indicator -->
				{#if isFocused && !hasValueInCell && !disabled}
					<div
						class="absolute top-1/2 left-1/2 h-6 w-0.5
                               -translate-x-1/2 -translate-y-1/2 animate-pulse rounded-full bg-zinc-900"
					></div>
				{/if}
			</div>
		{/each}
	</div>

	<!-- Success indicator -->
	{#if valid && isFilled}
		<div class="flex items-center justify-center gap-2 text-green-600">
			<div class="flex h-5 w-5 items-center justify-center rounded-full bg-green-500">
				<Check class="h-3 w-3 text-white" strokeWidth={3} />
			</div>
			<span class="text-sm font-medium">Код введён</span>
		</div>
	{/if}

	<!-- Error message -->
	{#if error}
		<div
			class="mx-auto flex max-w-xs items-center justify-center gap-2 rounded-xl border
                    border-red-200 bg-red-50 px-4 py-2.5"
		>
			<AlertCircle class="h-4 w-4 shrink-0 text-red-500" />
			<p class="text-xs font-medium text-red-700">{error}</p>
		</div>
	{/if}
</div>
