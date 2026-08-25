<script lang="ts">
	import type { HousingType } from '$lib/types/listings';
	import { cn } from '$lib/utils';
	import { Search, Building, Home, Trees, SlidersHorizontal } from 'lucide-svelte';

	interface Props {
		selectedType?: HousingType | 'all';
		searchQuery?: string;
		onfilter?: (type: HousingType | 'all', query: string) => void;
	}

	let {
		selectedType = $bindable('all'),
		searchQuery = $bindable(''),
		onfilter
	}: Props = $props();

	const typeOptions: Array<{ id: HousingType | 'all'; label: string; icon: typeof Building }> = [
		{ id: 'all', label: 'Все категории', icon: SlidersHorizontal },
		{ id: 'apartment', label: 'Квартиры', icon: Building },
		{ id: 'house', label: 'Дома и коттеджи', icon: Home },
		{ id: 'manor', label: 'Усадьбы', icon: Trees }
	];

	function handleTypeSelect(type: HousingType | 'all') {
		selectedType = type;
		onfilter?.(selectedType, searchQuery);
	}

	function handleSearchInput() {
		onfilter?.(selectedType, searchQuery);
	}
</script>

<div class="w-full space-y-4">
	<!-- Search Input Bar -->
	<div class="relative w-full max-w-2xl mx-auto">
		<Search class="absolute left-4 top-3.5 h-5 w-5 text-muted-foreground" />
		<input
			type="text"
			placeholder="Поиск по названию или описанию..."
			bind:value={searchQuery}
			oninput={handleSearchInput}
			class="w-full rounded-2xl border border-border bg-card py-3.5 pl-12 pr-4 text-sm shadow-sm transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:border-transparent placeholder:text-muted-foreground"
		/>
	</div>

	<!-- Category Type Pills -->
	<div class="flex items-center justify-center gap-2 overflow-x-auto py-2 scrollbar-none">
		{#each typeOptions as opt}
			{@const Icon = opt.icon}
			<button
				type="button"
				onclick={() => handleTypeSelect(opt.id)}
				class={cn(
					'flex items-center gap-2 rounded-xl px-4 py-2.5 text-xs font-semibold transition-all select-none whitespace-nowrap cursor-pointer',
					selectedType === opt.id
						? 'bg-primary text-primary-foreground shadow-md'
						: 'bg-card border border-border text-muted-foreground hover:bg-accent hover:text-foreground'
				)}
			>
				<Icon class="h-4 w-4" />
				<span>{opt.label}</span>
			</button>
		{/each}
	</div>
</div>
