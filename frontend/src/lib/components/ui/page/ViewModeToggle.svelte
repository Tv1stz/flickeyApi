<script lang="ts">
	import { LayoutGrid, List } from 'lucide-svelte';

	type ViewMode = 'grid' | 'list';

	interface Props {
		value: ViewMode;
		onChange?: (value: ViewMode) => void;
		class?: string;
	}

	let { value, onChange, class: className = '' }: Props = $props();

	function setMode(next: ViewMode) {
		if (next === value) return;
		onChange?.(next);
	}
</script>

<div class="flex items-center rounded-xl border border-zinc-200 p-0.5 {className}">
	<button
		type="button"
		class="flex h-11 w-11 items-center justify-center rounded-lg transition-colors
            {value === 'grid' ? 'bg-zinc-900 text-white' : 'text-zinc-400 hover:text-zinc-600'}"
		onclick={() => setMode('grid')}
		aria-label="Карточками"
	>
		<LayoutGrid size={18} />
	</button>
	<button
		type="button"
		class="flex h-11 w-11 items-center justify-center rounded-lg transition-colors
            {value === 'list' ? 'bg-zinc-900 text-white' : 'text-zinc-400 hover:text-zinc-600'}"
		onclick={() => setMode('list')}
		aria-label="Списком"
	>
		<List size={18} />
	</button>
</div>
