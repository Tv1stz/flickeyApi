<!-- src/lib/components/form/PasswordField.svelte -->
<script lang="ts">
	import { Eye, EyeOff, Check, AlertCircle } from 'lucide-svelte';

	interface Props {
		id: string;
		label?: string;
		placeholder?: string;
		required?: boolean;
		value?: string;
		error?: string;
		touched?: boolean;
		disabled?: boolean;
		helpText?: string;
		valid?: boolean;
		showStrength?: boolean;
		autocomplete?: 'current-password' | 'new-password' | 'off';
		onInput: (value: string) => void;
		onBlur?: () => void;
	}

	let {
		id,
		label = 'Пароль',
		placeholder = '',
		required = false,
		value = '',
		error = '',
		touched = false,
		disabled = false,
		helpText = '',
		valid = false,
		showStrength = false,
		autocomplete = 'current-password',
		onInput,
		onBlur
	}: Props = $props();

	let showPassword = $state(false);

	const hasError = $derived(Boolean(touched && error));
	const hasSuccess = $derived(Boolean(touched && !hasError && valid));
	const hasValue = $derived(value.length > 0);

	// ═══════════════════════════════════════════════════════════════
	// PASSWORD STRENGTH
	// ═══════════════════════════════════════════════════════════════

	const strength = $derived.by(() => {
		if (!value || !showStrength) return { score: 0, label: '', color: '', width: '0%' };

		let score = 0;
		const checks = {
			length: value.length >= 8,
			lowercase: /[a-z]/.test(value),
			uppercase: /[A-Z]/.test(value),
			numbers: /\d/.test(value),
			special: /[^a-zA-Z0-9]/.test(value)
		};

		if (checks.length) score++;
		if (value.length >= 12) score++;
		if (checks.lowercase && checks.uppercase) score++;
		if (checks.numbers) score++;
		if (checks.special) score++;

		const levels = [
			{ label: 'Очень слабый', color: 'bg-red-500' },
			{ label: 'Слабый', color: 'bg-orange-500' },
			{ label: 'Средний', color: 'bg-amber-500' },
			{ label: 'Хороший', color: 'bg-lime-500' },
			{ label: 'Отличный', color: 'bg-green-500' }
		];

		const level = levels[Math.min(score, 4)];
		const width = `${(score / 5) * 100}%`;

		return { score, ...level, width };
	});

	// ═══════════════════════════════════════════════════════════════
	// HANDLERS
	// ═══════════════════════════════════════════════════════════════

	function handleInput(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		onInput(input.value);
	}

	function toggleVisibility() {
		showPassword = !showPassword;
	}

	// ═══════════════════════════════════════════════════════════════
	// STYLES (matching TextInputField)
	// ═══════════════════════════════════════════════════════════════

	const baseInputClasses =
		'block w-full min-h-[56px] border rounded-2xl pl-4 pr-20 pt-7 pb-2 text-base ' +
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
	const inputClasses = $derived(`${baseInputClasses} ${stateClasses}`);

	const labelColorClass = $derived(
		hasError ? 'text-red-600' : hasSuccess ? 'text-green-600' : 'text-zinc-600'
	);

	const errorId = $derived(`${id}-error`);
	const helpId = $derived(`${id}-help`);
	const describedBy = $derived(
		[helpText ? helpId : '', touched && error ? errorId : ''].filter(Boolean).join(' ') || undefined
	);
</script>

<div class="space-y-2">
	<div class="relative">
		<input
			{id}
			type={showPassword ? 'text' : 'password'}
			{autocomplete}
			class={inputClasses}
			{placeholder}
			{value}
			{disabled}
			aria-invalid={hasError}
			aria-required={required}
			aria-describedby={describedBy}
			oninput={handleInput}
			onblur={() => onBlur?.()}
		/>

		<!-- Floating label -->
		<label
			for={id}
			class="pointer-events-none absolute top-2 left-4 bg-white {labelColorClass}
                   -ml-1 rounded px-1 text-[11px] leading-none font-medium transition-colors duration-200"
		>
			{label}{#if required}<span class="ml-0.5 text-red-500">*</span>{/if}
		</label>

		<!-- Right side controls -->
		<div class="absolute top-1/2 right-2 flex -translate-y-1/2 items-center gap-1">
			{#if hasSuccess}
				<div
					class="flex h-5 w-5 items-center justify-center rounded-full bg-green-500"
					aria-hidden="true"
				>
					<Check class="h-3.5 w-3.5 text-white" strokeWidth={3} />
				</div>
			{:else if hasError}
				<div class="text-red-500" aria-hidden="true">
					<AlertCircle class="h-5 w-5" />
				</div>
			{/if}

			<button
				type="button"
				class="flex h-9 w-9 items-center justify-center rounded-xl
                       text-zinc-400 transition-colors duration-150
                       hover:bg-zinc-100 hover:text-zinc-600 active:bg-zinc-200
                       disabled:pointer-events-none disabled:opacity-50"
				onclick={toggleVisibility}
				{disabled}
				tabindex={-1}
				aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
			>
				{#if showPassword}
					<EyeOff class="h-5 w-5" />
				{:else}
					<Eye class="h-5 w-5" />
				{/if}
			</button>
		</div>
	</div>

	<!-- Strength indicator -->
	{#if showStrength && hasValue}
		<div class="space-y-1.5 px-0.5">
			<div class="h-1.5 overflow-hidden rounded-full bg-zinc-200">
				<div
					class="h-full rounded-full transition-all duration-500 ease-out {strength.color}"
					style:width={strength.width}
				></div>
			</div>
			<div class="flex items-center justify-between">
				<span class="text-xs text-zinc-500">{strength.label}</span>
				<span class="text-[10px] text-zinc-400 tabular-nums">{strength.score}/5</span>
			</div>
		</div>
	{/if}

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
