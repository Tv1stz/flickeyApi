<script lang="ts">
	type Variant = 'inline' | 'pill';
	type Size = 'xs' | 'sm';

	interface Props {
		active?: boolean;
		status?: string;
		activeLabel?: string;
		inactiveLabel?: string;
		variant?: Variant;
		size?: Size;
		class?: string;
	}

	let {
		active = true,
		status,
		activeLabel,
		inactiveLabel = 'Неактивно',
		variant = 'inline',
		size = 'xs',
		class: className = ''
	}: Props = $props();

	const cx = (...classes: Array<string | false | null | undefined>) =>
		classes.filter(Boolean).join(' ');

	const resolvedStatus = $derived.by(() => {
		if (status) return status;
		return active ? 'active' : 'inactive';
	});

	const statusConfig = $derived.by(() => {
		switch (resolvedStatus) {
			case 'published':
			case 'active':
				return {
					label: activeLabel || 'Активно',
					dot: 'bg-emerald-500',
					text: 'text-emerald-700 dark:text-emerald-400'
				};
			case 'pending_review':
			case 'awaiting_company_verification':
				return {
					label: activeLabel || 'На модерации',
					dot: 'bg-amber-500',
					text: 'text-amber-700 dark:text-amber-400'
				};
			case 'rejected':
				return {
					label: activeLabel || 'Отклонено',
					dot: 'bg-rose-500',
					text: 'text-rose-700 dark:text-rose-400'
				};
			case 'draft_video_required':
				return {
					label: activeLabel || 'Черновик: требуется загрузить видео',
					dot: 'bg-rose-500',
					text: 'text-rose-700 dark:text-rose-400'
				};
			case 'draft':
				return {
					label: activeLabel || 'Черновик',
					dot: 'bg-blue-500',
					text: 'text-blue-700 dark:text-blue-400'
				};
			case 'archived':
				return {
					label: activeLabel || 'В архиве',
					dot: 'bg-zinc-400',
					text: 'text-zinc-500 dark:text-zinc-400'
				};
			case 'suspended':
			case 'inactive':
			default:
				return {
					label: inactiveLabel || 'Неактивно',
					dot: 'bg-zinc-400',
					text: 'text-zinc-500 dark:text-zinc-400'
				};
		}
	});

	const dotClass = $derived(size === 'sm' ? 'h-2 w-2' : 'h-1.5 w-1.5');
	const inlineTextClass = $derived(size === 'sm' ? 'text-sm' : 'text-xs');
</script>

{#if variant === 'pill'}
	<span
		class={cx(
			'inline-flex items-center gap-1.5 rounded-full bg-white/90 dark:bg-zinc-900/90 px-2.5 py-1 text-[11px] font-semibold shadow-xs backdrop-blur-sm',
			className
		)}
	>
		<span class={cx('rounded-full shrink-0', dotClass, statusConfig.dot)}></span>
		<span class={cx('font-medium', statusConfig.text)}>{statusConfig.label}</span>
	</span>
{:else}
	<span class={cx('inline-flex items-center gap-1.5', className)}>
		<span class={cx('rounded-full shrink-0', dotClass, statusConfig.dot)}></span>
		<span class={cx('font-medium', inlineTextClass, statusConfig.text)}>{statusConfig.label}</span>
	</span>
{/if}
