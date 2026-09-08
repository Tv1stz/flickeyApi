<!-- src/lib/components/form/TextareaField.svelte -->
<script lang="ts">
	import { AlertCircle } from 'lucide-svelte';

	interface Props {
		id: string;
		label: string;
		placeholder?: string;
		required?: boolean;
		rows?: number;
		value?: string;
		error?: string;
		touched?: boolean;
		disabled?: boolean;
		helpText?: string;
		onInput: (value: string) => void;
		onBlur: () => void;
	}

	let {
		id,
		label,
		placeholder = '',
		required = false,
		rows = 4,
		value = '',
		error = '',
		touched = false,
		disabled = false,
		helpText = '',
		onInput,
		onBlur
	}: Props = $props();

	const hasError = $derived(Boolean(touched && error));

	const baseInputClasses =
		'block w-full rounded-2xl border px-4 py-3.5 text-base bg-white text-zinc-900 placeholder:text-zinc-400 ' +
		'transition-all duration-200 resize-y touch-manipulation ' +
		'disabled:bg-zinc-50 disabled:text-zinc-400 disabled:cursor-not-allowed disabled:opacity-60';

	const okClasses =
		'border-zinc-300 hover:border-zinc-400 hover:shadow-sm ' +
		'focus:outline-none focus:ring-2 focus:ring-zinc-900 focus:ring-offset-1 focus:border-zinc-900';

	const errClasses =
		'border-red-500 bg-red-50/50 ' +
		'focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1 focus:border-red-600 focus:bg-white';

	const inputClasses = $derived(`${baseInputClasses} ${hasError ? errClasses : okClasses}`);
	const labelColorClass = $derived(hasError ? 'text-red-600' : 'text-zinc-900');

	const errorId = $derived(`${id}-error`);
	const helpId = $derived(`${id}-help`);
	const describedBy = $derived(
		[helpText && !hasError ? helpId : null, hasError ? errorId : null].filter(Boolean).join(' ') ||
			undefined
	);
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

	<textarea
		{id}
		{rows}
		{disabled}
		class={inputClasses}
		{placeholder}
		{value}
		aria-invalid={hasError}
		aria-required={required}
		aria-describedby={describedBy}
		oninput={(e) => onInput((e.currentTarget as HTMLTextAreaElement).value)}
		onblur={onBlur}
	></textarea>

	{#if helpText && !hasError}
		<p id={helpId} class="px-0.5 text-xs leading-relaxed whitespace-pre-line text-zinc-500">
			{helpText}
		</p>
	{/if}

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
