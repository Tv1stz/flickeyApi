<!-- src/lib/components/form/CounterField.svelte -->
<script lang="ts">
	import { Minus, Plus, AlertCircle } from 'lucide-svelte';

	interface Props {
		id: string;
		label: string;
		required?: boolean;
		value?: string;
		min?: number;
		max?: number;
		step?: number;
		helpText?: string;
		error?: string;
		touched?: boolean;
		disabled?: boolean;
		onInput: (value: string) => void;
		onBlur: () => void;
	}

	let {
		id,
		label,
		required = false,
		value = '',
		min = 0,
		max = undefined,
		step = 1,
		helpText = '',
		error = '',
		touched = false,
		disabled = false,
		onInput,
		onBlur
	}: Props = $props();

	let focused = $state(false);
	let draft = $state('');

	const hasError = $derived(Boolean(touched && error));

	function clamp(n: number): number {
		let result = n;
		if (result < min) result = min;
		if (max !== undefined && result > max) result = max;
		return result;
	}

	function parseValue(raw: string): { normalized: string; num: number } {
		const trimmed = raw.trim();

		if (!trimmed || !/^\d+$/.test(trimmed)) {
			const num = clamp(min);
			return { normalized: String(num), num };
		}

		const parsed = Math.trunc(Number(trimmed));
		const clamped = clamp(parsed);
		return { normalized: String(clamped), num: clamped };
	}

	function commit(raw: string) {
		const { normalized } = parseValue(raw);
		draft = normalized;
		onInput(normalized);
	}

	// Синхронизация draft с value когда не в фокусе
	$effect(() => {
		if (!focused) {
			draft = parseValue(value).normalized;
		}
	});

	// Нормализация значения при изменении извне
	$effect(() => {
		if (!focused) {
			const { normalized } = parseValue(value);
			if (value !== normalized) {
				onInput(normalized);
			}
		}
	});

	const current = $derived(parseValue(focused ? draft : value).num);
	const canDecrement = $derived(current > min);
	const canIncrement = $derived(max === undefined || current < max);

	function handleInput(e: Event) {
		commit((e.currentTarget as HTMLInputElement).value);
	}

	function handlePaste(e: ClipboardEvent) {
		const text = e.clipboardData?.getData('text') ?? '';
		if (!/^\d+$/.test(text.trim())) {
			e.preventDefault();
			commit(String(min));
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowUp') {
			e.preventDefault();
			if (canIncrement && !disabled) {
				onInput(String(clamp(current + step)));
			}
		}
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			if (canDecrement && !disabled) {
				onInput(String(clamp(current - step)));
			}
		}
	}

	function decrement() {
		if (canDecrement && !disabled) {
			onInput(String(clamp(current - step)));
		}
	}

	function increment() {
		if (canIncrement && !disabled) {
			onInput(String(clamp(current + step)));
		}
	}

	const wrapperClasses = $derived.by(() => {
		const base =
			'flex w-full items-center rounded-2xl border transition-all duration-200 overflow-hidden h-14 bg-white';
		const stateClasses = hasError
			? 'border-red-500 bg-red-50/50 focus-within:ring-2 focus-within:ring-red-500 focus-within:ring-offset-1 focus-within:border-red-600 focus-within:bg-white'
			: 'border-zinc-300 hover:border-zinc-400 hover:shadow-sm focus-within:ring-2 focus-within:ring-zinc-900 focus-within:ring-offset-1 focus-within:border-zinc-900';
		const disabledClass = disabled ? 'opacity-60 bg-zinc-50 cursor-not-allowed' : '';
		return `${base} ${stateClasses} ${disabledClass}`;
	});

	const buttonClasses = $derived.by(() => {
		const base =
			'h-14 w-[4.5rem] sm:w-20 grid place-items-center transition-all duration-200 select-none disabled:opacity-40 disabled:cursor-not-allowed';
		const colorClass = hasError
			? 'text-red-600 hover:bg-red-100 active:bg-red-200'
			: 'text-zinc-900 hover:bg-zinc-100 active:bg-zinc-200';
		return `${base} ${colorClass}`;
	});

	const labelColorClass = $derived(hasError ? 'text-red-600' : 'text-zinc-900');
</script>

<div class="space-y-2">
	<label
		for={id}
		class="block text-sm font-medium transition-colors duration-200 {labelColorClass}"
	>
		{label}
		{#if required}
			<span class="ml-0.5 text-red-500">*</span>
		{/if}
	</label>

	<div class={wrapperClasses}>
		<button
			type="button"
			class="{buttonClasses} border-r border-zinc-200"
			onclick={decrement}
			disabled={!canDecrement || disabled}
			aria-label="Уменьшить {label}"
		>
			<Minus class="h-4 w-4" strokeWidth={2.5} />
		</button>

		<input
			{id}
			type="text"
			inputmode="numeric"
			pattern="[0-9]*"
			autocomplete="off"
			{disabled}
			class="h-12 w-full bg-transparent px-3 text-center text-base font-semibold text-zinc-900 tabular-nums outline-none disabled:cursor-not-allowed disabled:text-zinc-400"
			bind:value={draft}
			oninput={handleInput}
			onpaste={handlePaste}
			onkeydown={handleKeydown}
			onfocus={() => (focused = true)}
			onblur={() => {
				focused = false;
				commit(draft);
				onBlur();
			}}
			aria-invalid={hasError}
		/>

		<button
			type="button"
			class="{buttonClasses} border-l border-zinc-200"
			onclick={increment}
			disabled={!canIncrement || disabled}
			aria-label="Увеличить {label}"
		>
			<Plus class="h-4 w-4" strokeWidth={2.5} />
		</button>
	</div>

	{#if helpText && !hasError}
		<p class="px-0.5 text-xs leading-relaxed text-zinc-500">
			{helpText}
		</p>
	{/if}

	{#if touched && error}
		<div class="flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-3 py-2.5">
			<AlertCircle class="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
			<p class="text-xs leading-relaxed font-medium text-red-700">{error}</p>
		</div>
	{/if}
</div>
