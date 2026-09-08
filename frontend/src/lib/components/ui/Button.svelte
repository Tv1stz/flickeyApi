<!-- src/lib/components/ui/Button.svelte -->
<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';

	type Variant = 'solid' | 'outline' | 'ghost' | 'soft' | 'link' | 'glass';
	type Tone = 'primary' | 'neutral' | 'danger' | 'success' | 'rose';
	type Size = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'icon-xs' | 'icon-sm' | 'icon' | 'icon-lg';
	type Radius = 'none' | 'md' | 'lg' | 'xl' | '2xl' | 'pill';
	type IconComponent = ComponentType;

	interface Props {
		children?: Snippet;
		as?: 'button' | 'a';
		href?: string;
		target?: string;
		rel?: string;
		type?: 'button' | 'submit' | 'reset';
		variant?: Variant;
		tone?: Tone;
		size?: Size;
		radius?: Radius;
		fullWidth?: boolean;
		disabled?: boolean;
		loading?: boolean;
		loadingText?: string;
		active?: boolean;
		iconLeft?: IconComponent;
		iconRight?: IconComponent;
		iconClass?: string;
		class?: string;
		onclick?: (e: MouseEvent) => void;
		[key: string]: unknown;
	}

	let {
		children,
		as,
		href,
		target,
		rel,
		type = 'button',
		variant = 'solid',
		tone = 'primary',
		size = 'md',
		radius = 'xl',
		fullWidth = false,
		disabled = false,
		loading = false,
		loadingText = '',
		active = false,
		iconLeft,
		iconRight,
		iconClass = '',
		class: className = '',
		onclick,
		...rest
	}: Props = $props();

	const cx = (...classes: (string | false | null | undefined)[]) =>
		classes.filter(Boolean).join(' ');

	const element = $derived(as ?? (href ? 'a' : 'button'));
	const isDisabled = $derived(disabled || loading);
	const isIconOnly = $derived(size.startsWith('icon') && !loadingText && !children);

	const radiusClasses: Record<Radius, string> = {
		none: 'rounded-none',
		md: 'rounded-md',
		lg: 'rounded-lg',
		xl: 'rounded-xl',
		'2xl': 'rounded-2xl',
		pill: 'rounded-full'
	};

	const sizeClasses: Record<Size, string> = {
		xs: 'h-8 min-h-[32px] px-3 text-xs gap-1.5',
		sm: 'h-9 min-h-[36px] px-4 text-sm gap-1.5',
		md: 'h-10 min-h-[40px] px-4 text-sm gap-2',
		lg: 'h-11 min-h-[44px] px-5 text-sm gap-2',
		xl: 'h-12 min-h-[48px] px-6 text-base gap-2.5',
		'icon-xs': 'h-7 w-7 min-h-[28px] min-w-[28px] p-0',
		'icon-sm': 'h-8 w-8 min-h-[32px] min-w-[32px] p-0',
		icon: 'h-10 w-10 min-h-[40px] min-w-[40px] p-0',
		'icon-lg': 'h-11 w-11 min-h-[44px] min-w-[44px] p-0'
	};

	const computedRadius = $derived(size.startsWith('icon') ? 'pill' : radius);

	const focusRing: Record<Tone, string> = {
		primary:
			'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-900 focus-visible:ring-offset-2',
		neutral:
			'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-400 focus-visible:ring-offset-2',
		danger:
			'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2',
		success:
			'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2',
		rose: 'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-rose-500 focus-visible:ring-offset-2'
	};

	const variantToneClasses = $derived.by(() => {
		const map: Record<Tone, Record<Variant, string>> = {
			primary: {
				solid: `bg-zinc-900 text-white hover:bg-zinc-950 active:bg-zinc-950 ${focusRing.primary}`,
				outline: `bg-white text-zinc-900 border border-zinc-900 hover:bg-zinc-900 hover:text-white ${focusRing.primary}`,
				ghost: `bg-transparent text-zinc-900 hover:bg-zinc-100 active:bg-zinc-200 ${focusRing.primary}`,
				soft: `bg-zinc-100 text-zinc-900 hover:bg-zinc-200 active:bg-zinc-300 ${focusRing.primary}`,
				link: `bg-transparent text-zinc-900 !px-0 !h-auto !min-h-0 underline underline-offset-4 decoration-zinc-300 hover:decoration-zinc-500 ${focusRing.primary}`,
				glass: `bg-white/80 backdrop-blur-xl text-zinc-700 border border-zinc-200/50 shadow-lg shadow-zinc-900/5 hover:bg-white hover:text-zinc-900 hover:shadow-xl hover:border-zinc-300 ${focusRing.primary}`
			},
			neutral: {
				solid: `bg-zinc-100 text-zinc-800 hover:bg-zinc-200 active:bg-zinc-300 ${focusRing.neutral}`,
				outline: `bg-white text-zinc-700 border border-zinc-300 hover:bg-zinc-50 hover:border-zinc-400 active:bg-zinc-100 ${focusRing.neutral}`,
				ghost: `bg-transparent text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 active:bg-zinc-200 ${focusRing.neutral}`,
				soft: `bg-zinc-50 text-zinc-700 hover:bg-zinc-100 active:bg-zinc-200 ${focusRing.neutral}`,
				link: `bg-transparent text-zinc-500 !px-0 !h-auto !min-h-0 hover:text-zinc-700 underline underline-offset-4 decoration-zinc-300 hover:decoration-zinc-400 ${focusRing.neutral}`,
				glass: `bg-white/80 backdrop-blur-xl text-zinc-600 border border-zinc-200/50 shadow-lg shadow-zinc-900/5 hover:bg-white hover:text-zinc-900 hover:shadow-xl hover:border-zinc-300 ${focusRing.neutral}`
			},
			danger: {
				solid: `bg-red-600 text-white hover:bg-red-700 active:bg-red-800 ${focusRing.danger}`,
				outline: `bg-white text-red-600 border border-red-300 hover:bg-red-50 hover:border-red-400 active:bg-red-100 ${focusRing.danger}`,
				ghost: `bg-transparent text-red-600 hover:bg-red-50 active:bg-red-100 ${focusRing.danger}`,
				soft: `bg-red-50 text-red-600 hover:bg-red-100 active:bg-red-200 ${focusRing.danger}`,
				link: `bg-transparent text-red-600 !px-0 !h-auto !min-h-0 hover:text-red-700 underline underline-offset-4 decoration-red-300 hover:decoration-red-400 ${focusRing.danger}`,
				glass: `bg-red-50/80 backdrop-blur-xl text-red-600 border border-red-200/50 shadow-lg shadow-red-900/5 hover:bg-red-50 hover:shadow-xl hover:border-red-300 ${focusRing.danger}`
			},
			success: {
				solid: `bg-emerald-600 text-white hover:bg-emerald-700 active:bg-emerald-800 ${focusRing.success}`,
				outline: `bg-white text-emerald-600 border border-emerald-300 hover:bg-emerald-50 hover:border-emerald-400 active:bg-emerald-100 ${focusRing.success}`,
				ghost: `bg-transparent text-emerald-600 hover:bg-emerald-50 active:bg-emerald-100 ${focusRing.success}`,
				soft: `bg-emerald-50 text-emerald-600 hover:bg-emerald-100 active:bg-emerald-200 ${focusRing.success}`,
				link: `bg-transparent text-emerald-600 !px-0 !h-auto !min-h-0 hover:text-emerald-700 underline underline-offset-4 decoration-emerald-300 hover:decoration-emerald-400 ${focusRing.success}`,
				glass: `bg-emerald-50/80 backdrop-blur-xl text-emerald-600 border border-emerald-200/50 shadow-lg shadow-emerald-900/5 hover:bg-emerald-50 hover:shadow-xl hover:border-emerald-300 ${focusRing.success}`
			},
			rose: {
				solid: `bg-rose-500 text-white hover:bg-rose-600 active:bg-rose-700 ${focusRing.rose}`,
				outline: `bg-white text-rose-600 border border-rose-300 hover:bg-rose-50 hover:border-rose-400 active:bg-rose-100 ${focusRing.rose}`,
				ghost: `bg-transparent text-rose-600 hover:bg-rose-50 active:bg-rose-100 ${focusRing.rose}`,
				soft: `bg-rose-50 text-rose-600 hover:bg-rose-100 active:bg-rose-200 ${focusRing.rose}`,
				link: `bg-transparent text-rose-600 !px-0 !h-auto !min-h-0 hover:text-rose-700 underline underline-offset-4 decoration-rose-300 hover:decoration-rose-400 ${focusRing.rose}`,
				glass: `bg-rose-50/90 backdrop-blur-xl text-rose-600 border border-rose-200 shadow-lg shadow-rose-900/5 hover:bg-rose-100 hover:shadow-xl hover:border-rose-300 ${focusRing.rose}`
			}
		};
		return map[tone][variant];
	});

	const iconSizeClass = $derived.by(() => {
		const sizes: Record<Size, string> = {
			xs: 'h-3.5 w-3.5',
			sm: 'h-4 w-4',
			md: 'h-4 w-4',
			lg: 'h-[18px] w-[18px]',
			xl: 'h-5 w-5',
			'icon-xs': 'h-3.5 w-3.5',
			'icon-sm': 'h-4 w-4',
			icon: 'h-[18px] w-[18px]',
			'icon-lg': 'h-5 w-5'
		};
		return sizes[size];
	});

	const buttonClasses = $derived(
		cx(
			'inline-flex items-center justify-center',
			'font-medium select-none whitespace-nowrap',
			'transition-all duration-200 ease-out',
			'outline-none touch-manipulation',
			'active:scale-[0.97]',
			'disabled:opacity-50 disabled:pointer-events-none disabled:scale-100',
			sizeClasses[size],
			radiusClasses[computedRadius],
			variantToneClasses,
			fullWidth && 'w-full',
			isDisabled && element !== 'button' && 'pointer-events-none opacity-50',
			className
		)
	);

	const computedIconClass = $derived(cx(iconSizeClass, 'shrink-0', iconClass));

	function handleClick(e: MouseEvent) {
		if (isDisabled) {
			e.preventDefault();
			e.stopPropagation();
			return;
		}
		onclick?.(e);
	}
</script>

<svelte:element
	this={element}
	{...rest}
	href={element === 'a' ? href : undefined}
	target={element === 'a' ? target : undefined}
	rel={element === 'a' && target === '_blank' ? rel || 'noopener noreferrer' : rel}
	type={element === 'button' ? type : undefined}
	disabled={element === 'button' ? isDisabled : undefined}
	aria-disabled={element !== 'button' && isDisabled ? 'true' : undefined}
	aria-busy={loading ? 'true' : undefined}
	aria-pressed={active ? 'true' : undefined}
	tabindex={element !== 'button' && isDisabled ? -1 : undefined}
	class={buttonClasses}
	onclick={handleClick}
>
	{#if loading}
		<svg class="h-4 w-4 shrink-0 animate-spin" viewBox="0 0 24 24" fill="none">
			<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" />
			<path
				class="opacity-75"
				d="M4 12a8 8 0 0 1 8-8"
				stroke="currentColor"
				stroke-width="3"
				stroke-linecap="round"
			/>
		</svg>
		{#if loadingText}
			<span>{loadingText}</span>
		{/if}
	{:else}
		{#if iconLeft}
			{@const IconLeft = iconLeft}
			<IconLeft class={computedIconClass} />
		{/if}

		{#if children && !isIconOnly}
			{@render children()}
		{/if}

		{#if iconRight}
			{@const IconRight = iconRight}
			<IconRight class={computedIconClass} />
		{/if}
	{/if}
</svelte:element>
