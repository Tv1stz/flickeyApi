<script lang="ts">
	interface Props {
		description: string;
	}

	let { description = '' }: Props = $props();
</script>

<div class="space-y-3">
	<h3 class="text-lg font-bold text-foreground">Об объекте</h3>
	<div class="text-sm text-muted-foreground leading-relaxed whitespace-pre-line">
		{description}
	</div>
</div>
