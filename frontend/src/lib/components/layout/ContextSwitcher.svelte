<script lang="ts">
	import { auth } from '$lib/auth/auth.svelte';
	import { cn } from '$lib/utils';
	import { User, Home, Lock } from 'lucide-svelte';

	let { class: className = '' } = $props();
</script>

{#if auth.isAuthenticated}
	<div
		class={cn(
			'inline-flex items-center rounded-xl bg-muted/80 p-1 border border-border/80 text-xs font-medium',
			className
		)}
	>
		<!-- Guest Context Button -->
		<button
			type="button"
			onclick={() => auth.switchContext('guest')}
			class={cn(
				'flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-all select-none cursor-pointer',
				auth.activeContext === 'guest'
					? 'bg-card text-foreground font-semibold shadow-sm'
					: 'text-muted-foreground hover:text-foreground'
			)}
		>
			<User class="h-3.5 w-3.5" />
			<span>Гость</span>
		</button>

		<!-- Host Context Button -->
		<button
			type="button"
			disabled={!auth.isHost}
			onclick={() => auth.switchContext('host')}
			class={cn(
				'flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-all select-none',
				auth.activeContext === 'host'
					? 'bg-card text-primary font-semibold shadow-sm cursor-pointer'
					: auth.isHost
						? 'text-muted-foreground hover:text-foreground cursor-pointer'
						: 'opacity-60 cursor-not-allowed text-muted-foreground'
			)}
			title={!auth.isHost ? 'Создайте объявление, чтобы стать Хостом' : ''}
		>
			<Home class="h-3.5 w-3.5" />
			<span>Хост</span>
			{#if !auth.isHost}
				<span
					class="inline-flex items-center gap-0.5 rounded-full bg-primary/10 text-primary px-1.5 py-0.2 text-[10px] font-bold"
				>
					<Lock class="h-2.5 w-2.5" /> Стать
				</span>
			{/if}
		</button>
	</div>
{/if}
