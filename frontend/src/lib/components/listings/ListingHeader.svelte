<script lang="ts">
	import Badge from '$lib/components/ui/Badge.svelte';
	import { translateHousingType, formatDate } from '$lib/utils';
	import type { HousingType } from '$lib/types/listings';
	import { ShieldCheck } from 'lucide-svelte';

	interface Props {
		name: string;
		type: HousingType;
		createdAt?: string;
		id?: string;
	}

	let { name, type, createdAt, id }: Props = $props();
</script>

<div class="space-y-2">
	<div class="flex flex-wrap items-center gap-2">
		<Badge variant="default" class="font-bold">
			{translateHousingType(type)}
		</Badge>
		<Badge variant="success" class="flex items-center gap-1">
			<ShieldCheck class="h-3 w-3" /> Проверенный объект
		</Badge>
	</div>

	<h1 class="text-2xl sm:text-4xl font-extrabold text-foreground tracking-tight leading-tight">
		{name}
	</h1>

	<div class="flex flex-wrap items-center gap-3 text-xs text-muted-foreground pt-1">
		{#if createdAt}
			<span>Опубликовано: {formatDate(createdAt)}</span>
			<span>•</span>
		{/if}
		{#if id}
			<span>ID: {id.slice(0, 8)}</span>
		{/if}
	</div>
</div>
