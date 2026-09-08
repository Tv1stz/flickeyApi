<script lang="ts">
	import type { ComponentType } from 'svelte';

	type Variant = 'surface' | 'floating';
	type Tone = 'neutral' | 'danger' | 'primary';
	type TooltipAlign = 'center' | 'left' | 'right';

	interface Props {
		icon: ComponentType;
		ariaLabel: string;
		tooltip?: string;
		tooltipAlign?: TooltipAlign;
		onclick?: (e: MouseEvent) => void;
		variant?: Variant;
		tone?: Tone;
		elevated?: boolean;
		iconSize?: number;
		disabled?: boolean;
		class?: string;
	}

	let {
		icon,
		ariaLabel,
		tooltip,
		tooltipAlign = 'center',
		onclick,
		variant = 'surface',
		tone = 'neutral',
		elevated = false,
		iconSize = 15,
		disabled = false,
		class: className = ''
	}: Props = $props();

	const cx = (...classes: Array<string | false | null | undefined>) =>
		classes.filter(Boolean).join(' ');

	const variantClasses: Record<Variant, string> = {
		surface: 'border border-zinc-200 bg-white',
		floating: 'bg-white/90 shadow-sm backdrop-blur-sm'
	};

	const toneClasses: Record<Variant, Record<Tone, string>> = {
		surface: {
			neutral: 'text-zinc-500 hover:border-zinc-300 hover:text-zinc-900',
			danger: 'text-zinc-500 hover:border-red-200 hover:text-red-500',
			primary: 'text-amber-600 border-amber-200 bg-amber-50/50 hover:bg-amber-100 hover:text-amber-700'
		},
		floating: {
			neutral: 'text-zinc-600 hover:bg-white hover:text-zinc-900 hover:shadow',
			danger: 'text-zinc-600 hover:bg-white hover:text-red-500 hover:shadow',
			primary: 'text-amber-600 bg-amber-50 hover:bg-white hover:text-amber-700 hover:shadow'
		}
	};

	const buttonClasses = $derived(
		cx(
			'flex h-9 w-9 items-center justify-center rounded-xl transition-all',
			'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-900/20 focus-visible:ring-offset-1',
			'disabled:pointer-events-none disabled:opacity-50',
			variantClasses[variant],
			toneClasses[variant][tone],
			elevated && variant === 'surface' && 'shadow-sm hover:shadow',
			className
		)
	);
	const Icon = $derived(icon);
	const tooltipText = $derived(tooltip || ariaLabel);

	const alignClasses: Record<TooltipAlign, { bubble: string; arrow: string }> = {
		center: {
			bubble: 'left-1/2 -translate-x-1/2 origin-bottom',
			arrow: 'left-1/2 -translate-x-1/2'
		},
		right: {
			bubble: 'right-0 origin-bottom-right',
			arrow: 'right-[14px]'
		},
		left: {
			bubble: 'left-0 origin-bottom-left',
			arrow: 'left-[14px]'
		}
	};

	function handleClick(e: MouseEvent) {
		e.stopPropagation();
		if (disabled) {
			e.preventDefault();
			return;
		}
		onclick?.(e);
	}
</script>

<div class="relative group/action-btn inline-flex">
	<button
		type="button"
		class={buttonClasses}
		aria-label={ariaLabel}
		{disabled}
		onclick={handleClick}
	>
		<Icon size={iconSize} />
	</button>

	{#if tooltipText}
		<div
			role="tooltip"
			class={cx(
				'pointer-events-none absolute bottom-full mb-2 z-50 whitespace-nowrap rounded-lg bg-zinc-900/95 px-2.5 py-1 text-[11px] font-medium text-white shadow-lg backdrop-blur-xs transition-all duration-150 opacity-0 scale-95 group-hover/action-btn:opacity-100 group-hover/action-btn:scale-100 group-focus-within/action-btn:opacity-100 group-focus-within/action-btn:scale-100 select-none',
				alignClasses[tooltipAlign].bubble
			)}
		>
			{tooltipText}
			<div
				class={cx(
					'absolute top-full -mt-1 border-4 border-transparent border-t-zinc-900/95',
					alignClasses[tooltipAlign].arrow
				)}
			></div>
		</div>
	{/if}
</div>
