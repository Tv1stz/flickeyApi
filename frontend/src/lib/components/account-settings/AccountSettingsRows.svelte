<script lang="ts">
	import type { AccountSettingAction, AccountSettingRow } from './types';

	interface Props {
		rows: AccountSettingRow[];
		onAction: (action: Exclude<AccountSettingAction, { type: 'none' }>) => void;
		variant?: 'desktop' | 'mobile';
	}

	let { rows, onAction, variant = 'desktop' }: Props = $props();

	function isActionable(
		action: AccountSettingAction | undefined
	): action is Exclude<AccountSettingAction, { type: 'none' }> {
		return Boolean(action && action.type !== 'none');
	}

	function toAction(
		action: AccountSettingAction | undefined
	): Exclude<AccountSettingAction, { type: 'none' }> | null {
		return isActionable(action) ? action : null;
	}

	const isMobile = $derived(variant === 'mobile');
	const rowClass = $derived(isMobile ? 'py-4' : 'py-4 sm:py-5');
</script>

<div class="divide-y divide-zinc-100">
	{#each rows as row (row.id)}
		{@const action = toAction(row.action)}
		<div class={`group flex items-start justify-between gap-4 ${rowClass}`}>
			<div class="min-w-0 flex-1">
				<p class="text-[15px] font-medium text-zinc-900">{row.label}</p>
				<p class="mt-0.5 text-sm text-zinc-600">{row.value}</p>
				{#if row.details}
					<p class="mt-1 text-xs text-zinc-400">{row.details}</p>
				{/if}
			</div>

			{#if action}
				<button
					type="button"
					class="shrink-0 text-sm font-medium text-zinc-900 underline underline-offset-2 decoration-zinc-300 transition-all hover:decoration-zinc-900 active:opacity-70"
					onclick={() => onAction(action)}
				>
					{action.label}
				</button>
			{/if}
		</div>
	{/each}
</div>
