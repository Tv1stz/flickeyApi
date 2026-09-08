<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import { ArrowLeft, Search, SlidersHorizontal } from 'lucide-svelte';

	type Variant = 'home' | 'search';

	interface Props {
		variant?: Variant;
		title: string;
		subtitle: string;
		showFilters?: boolean;
		filtersCount?: number;
		onBack?: () => void;
		onOpenSearch: () => void;
		onOpenFilters?: () => void;
		class?: string;
	}

	let {
		variant = 'home',
		title,
		subtitle,
		showFilters = false,
		filtersCount = 0,
		onBack,
		onOpenSearch,
		onOpenFilters,
		class: className = ''
	}: Props = $props();

	const isSearch = $derived(variant === 'search');
	const containerClass = $derived(
		['sticky top-0 z-40 bg-white/90 px-4 pb-2 pt-2 backdrop-blur', className]
			.filter(Boolean)
			.join(' ')
	);
</script>

<div class={containerClass} style="padding-top: max(8px, env(safe-area-inset-top));">
	{#if isSearch}
		<div class="grid grid-cols-[40px_minmax(0,1fr)_40px] items-center gap-2">
			<Button
				variant="ghost"
				tone="neutral"
				size="icon"
				class="!bg-transparent text-zinc-800 hover:!bg-transparent active:!bg-transparent"
				aria-label="Назад"
				onclick={() => onBack?.()}
			>
				<ArrowLeft size={18} strokeWidth={2.5} />
			</Button>

			<Button
				variant="outline"
				tone="neutral"
				size="xl"
				radius="pill"
				class="!h-14 w-full !gap-3 !px-4 shadow-md"
				onclick={() => onOpenSearch()}
			>
				<div class="min-w-0 items-center justify-center">
					<p class="truncate text-sm font-semibold text-zinc-900">{title}</p>
					<p class="truncate text-xs text-zinc-400">{subtitle}</p>
				</div>
			</Button>

			{#if showFilters}
				<Button
					variant="ghost"
					tone="neutral"
					size="icon"
					class="relative !bg-transparent text-zinc-800 hover:!bg-transparent active:!bg-transparent"
					aria-label="Фильтры"
					onclick={() => onOpenFilters?.()}
				>
					<SlidersHorizontal size={18} strokeWidth={2.2} />
					{#if filtersCount > 0}
						<span
							class="absolute -top-0.5 -right-0.5 flex h-5 min-w-[20px] items-center justify-center rounded-full bg-zinc-900 px-1 text-[10px] font-semibold text-white shadow"
							aria-label="{filtersCount} активных фильтров"
						>
							{filtersCount}
						</span>
					{/if}
				</Button>
			{:else}
				<div aria-hidden="true"></div>
			{/if}
		</div>
	{:else}
		<Button
			variant="outline"
			tone="neutral"
			size="xl"
			radius="pill"
			class="!h-14 w-full !px-4 shadow-md"
			onclick={() => onOpenSearch()}
		>
			<div class="flex w-full items-center justify-center gap-2">
				<span class="text-md font-semibold text-zinc-900">
					{title}
				</span>

				<Search size={16} strokeWidth={3} />
			</div>
		</Button>
	{/if}
</div>
