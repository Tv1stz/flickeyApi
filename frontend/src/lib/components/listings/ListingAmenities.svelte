<script lang="ts">
	import { CheckCircle2 } from 'lucide-svelte';
	import { translateAmenity } from '$lib/utils';

	interface Props {
		amenities: string[];
	}

	let { amenities = [] }: Props = $props();
</script>

{#if amenities.length > 0}
	<div class="space-y-3 pt-4 border-t border-border">
		<h3 class="text-lg font-bold text-foreground">Удобства и оснащение</h3>
		<div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
			{#each amenities as item}
				<div class="flex items-center gap-2 p-3 rounded-xl bg-muted/20 border border-border/80 text-xs font-medium">
					<CheckCircle2 class="h-4 w-4 text-emerald-600 shrink-0" />
					<span class="truncate">{translateAmenity(item)}</span>
				</div>
			{/each}
		</div>
	</div>
{/if}
