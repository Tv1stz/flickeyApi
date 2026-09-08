<script lang="ts">
	interface Props {
		description?: string;
	}

	let { description = '' }: Props = $props();
</script>

{#if description && description.trim().length > 0}
	<div class="space-y-3 pt-6 border-t border-border/60">
		<h2 class="text-base sm:text-lg font-bold text-foreground">Об этом месте</h2>
		<div class="text-xs sm:text-sm text-foreground/80 leading-relaxed whitespace-pre-line">
			{description}
		</div>
	</div>
{/if}
