<!-- src/lib/components/form/TextInputField.svelte -->
<script lang="ts">
	import { Check, AlertCircle } from 'lucide-svelte';

	interface Props {
		id: string;
		label: string;
		placeholder?: string;
		required?: boolean;
		type?: 'text' | 'number' | 'email';
		value?: string;
		error?: string;
		touched?: boolean;
		disabled?: boolean;
		min?: number;
		max?: number;
		step?: number | string;
		helpText?: string;
		onInput: (value: string) => void;
		onBlur: () => void;
		prefix?: string;
		suffix?: string;
		valid?: boolean;
		numeric?: boolean;
		allowDecimals?: boolean;
		hideErrorMessage?: boolean;
	}

	let {
		id,
		label,
		placeholder = '',
		required = false,
		type = 'text',
		value = '',
		error = '',
		touched = false,
		disabled = false,
		min = undefined,
		max = undefined,
		step = undefined,
		helpText = '',
		onInput,
		onBlur,
		prefix = undefined,
		suffix = undefined,
		valid = false,
		numeric = false,
		allowDecimals = false,
		hideErrorMessage = false
	}: Props = $props();

	const hasError = $derived(Boolean(touched && error));
	const hasSuccess = $derived(Boolean(touched && !hasError && valid));

	function normalizeNumericInput(raw: string): string {
		let cleaned = allowDecimals ? raw.replace(/[^\d.]/g, '') : raw.replace(/\D/g, '');

		if (allowDecimals) {
			const parts = cleaned.split('.');
			if (parts.length > 2) {
				cleaned = parts[0] + '.' + parts.slice(1).join('');
			}
		}

		// Remove leading zeros
		cleaned = cleaned.replace(/^0+(?=\d)/, '');

		return cleaned;
	}

	function handleInput(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		let newValue = input.value;

		if (numeric) {
			newValue = normalizeNumericInput(newValue);

			if (max !== undefined && newValue !== '') {
				const num = parseFloat(newValue);
				if (!isNaN(num) && num > max) {
					newValue = String(max);
				}
			}

			input.value = newValue;
		}

		onInput(newValue);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!numeric) return;

		const allowedKeys = [
			'Backspace',
			'Delete',
			'Tab',
			'Escape',
			'Enter',
			'ArrowLeft',
			'ArrowRight',
			'ArrowUp',
			'ArrowDown',
			'Home',
			'End'
		];

		if (allowedKeys.includes(e.key)) return;
		if (e.ctrlKey || e.metaKey) return;
		if (allowDecimals && e.key === '.' && !value.includes('.')) return;

		if (!/^\d$/.test(e.key)) {
			e.preventDefault();
		}
	}

	function handlePaste(e: ClipboardEvent) {
		if (!numeric) return;

		e.preventDefault();
		const input = e.currentTarget as HTMLInputElement;
		const text = e.clipboardData?.getData('text') ?? '';
		let normalized = normalizeNumericInput(text);

		if (max !== undefined && normalized !== '') {
			const num = parseFloat(normalized);
			if (!isNaN(num) && num > max) {
				normalized = String(max);
			}
		}

		input.value = normalized;
		onInput(normalized);
	}

	const padLeft = $derived(prefix ? 'pl-12' : 'pl-4');
	const padRight = $derived(suffix || hasError || hasSuccess ? 'pr-12' : 'pr-4');

	const baseInputClasses =
		'block w-full min-h-[56px] border rounded-2xl px-4 pt-7 pb-2 text-base ' +
		'bg-white text-zinc-900 placeholder:text-zinc-400 touch-manipulation ' +
		'transition-all duration-200 ' +
		'disabled:bg-zinc-50 disabled:text-zinc-400 disabled:cursor-not-allowed disabled:border-zinc-200';

	const okClasses =
		'border-zinc-300 hover:border-zinc-400 hover:shadow-sm ' +
		'focus:outline-none focus:ring-2 focus:ring-zinc-900 focus:ring-offset-1 focus:border-zinc-900';

	const successClasses =
		'border-green-500 ' +
		'focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-1 focus:border-green-500';

	const errClasses =
		'border-red-500 bg-red-50/50 ' +
		'focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1 focus:border-red-600 focus:bg-white';

	const stateClasses = $derived(hasError ? errClasses : hasSuccess ? successClasses : okClasses);
	const inputClasses = $derived(`${baseInputClasses} ${padLeft} ${padRight} ${stateClasses}`);

	const labelLeftClass = $derived(prefix ? 'left-12' : 'left-4');
	const labelColorClass = $derived(
		hasError ? 'text-red-600' : hasSuccess ? 'text-green-600' : 'text-zinc-600'
	);
	const affixColorClass = $derived(
		hasError ? 'text-red-600' : hasSuccess ? 'text-green-600' : 'text-zinc-500'
	);

	const inputMode = $derived(numeric ? (allowDecimals ? 'decimal' : 'numeric') : undefined);

	const errorId = $derived(`${id}-error`);
	const helpId = $derived(`${id}-help`);
	const describedBy = $derived(
		[helpText ? helpId : '', touched && error ? errorId : ''].filter(Boolean).join(' ') || undefined
	);
</script>

<div class="space-y-2">
	<div class="relative">
		{#if prefix}
			<span
				class="pointer-events-none absolute top-1/2 left-4 -translate-y-1/2 text-base font-medium {affixColorClass} transition-colors duration-200"
				aria-hidden="true"
			>
				{prefix}
			</span>
		{/if}

		<input
			{id}
			type={numeric ? 'text' : type}
			inputmode={inputMode}
			{disabled}
			class={inputClasses}
			{placeholder}
			{value}
			{min}
			{step}
			aria-invalid={hasError}
			aria-required={required}
			aria-describedby={describedBy}
			oninput={handleInput}
			onkeydown={handleKeydown}
			onpaste={handlePaste}
			onblur={onBlur}
		/>

		<label
			for={id}
			class="pointer-events-none absolute {labelLeftClass} top-2 bg-white {labelColorClass}
                -ml-1 rounded px-1 text-[11px] leading-none font-medium transition-colors duration-200"
		>
			{label}{#if required}<span class="ml-0.5 text-red-500">*</span>{/if}
		</label>

		{#if suffix && !hasError && !hasSuccess}
			<span
				class="pointer-events-none absolute top-1/2 right-4 -translate-y-1/2 text-base font-medium {affixColorClass} transition-colors duration-200"
				aria-hidden="true"
			>
				{suffix}
			</span>
		{/if}

		{#if hasSuccess}
			<div
				class="pointer-events-none absolute top-1/2 right-4 flex h-5 w-5 -translate-y-1/2 items-center justify-center rounded-full bg-green-500"
				aria-hidden="true"
			>
				<Check class="h-3.5 w-3.5 text-white" strokeWidth={3} />
			</div>
		{/if}

		{#if hasError}
			<div
				class="pointer-events-none absolute top-1/2 right-4 -translate-y-1/2 text-red-500"
				aria-hidden="true"
			>
				<AlertCircle class="h-5 w-5" />
			</div>
		{/if}
	</div>

	{#if helpText && !hasError}
		<p id={helpId} class="px-0.5 text-xs leading-relaxed text-zinc-500">
			{helpText}
		</p>
	{/if}

	{#if touched && error && !hideErrorMessage}
		<div
			id={errorId}
			class="flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3 py-2.5"
		>
			<AlertCircle class="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
			<p class="text-xs leading-relaxed font-medium text-red-700">{error}</p>
		</div>
	{/if}
</div>

<style>
	input[type='number']::-webkit-outer-spin-button,
	input[type='number']::-webkit-inner-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}

	input[type='number'] {
		-moz-appearance: textfield;
		appearance: textfield;
	}
</style>
