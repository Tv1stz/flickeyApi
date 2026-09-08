<script lang="ts">
	import type { HostInfo } from '$lib/types/listings';
	import { ShieldCheck, Calendar, ChevronRight, User } from 'lucide-svelte';

	interface Props {
		host?: HostInfo | null;
		hostName?: string;
		isVerified?: boolean;
		listingsCount?: number;
		yearsOnPlatform?: number;
		joinedYear?: number;
	}

	let {
		host,
		hostName,
		isVerified = true,
		listingsCount,
		yearsOnPlatform,
		joinedYear
	}: Props = $props();

	let name = $derived(
		host?.name ||
		(host?.first_name ? `${host.first_name}${host.last_name ? ' ' + host.last_name : ''}` : null) ||
		hostName ||
		'Арендодатель'
	);

	let initials = $derived(name ? name[0].toUpperCase() : 'H');
	let displayJoinedYear = $derived(joinedYear || 2024);
</script>

<div class="rounded-3xl border border-border/80 bg-card p-6 shadow-xs space-y-4">
	<!-- Host Avatar & Info -->
	<div class="flex flex-col items-center text-center">
		<div class="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-primary font-bold text-xl border-2 border-primary/20 shadow-xs">
			{initials}
		</div>

		<h3 class="font-bold text-sm sm:text-base text-foreground mt-2.5">
			{name}
		</h3>

		{#if isVerified}
			<div class="flex items-center gap-1 text-[11px] text-emerald-600 font-semibold mt-0.5">
				<ShieldCheck class="h-3.5 w-3.5 shrink-0" />
				<span>Подтверждённый профиль</span>
			</div>
		{/if}

		<div class="flex items-center gap-1 text-[11px] text-muted-foreground mt-1">
			<Calendar class="h-3 w-3 shrink-0" />
			<span>На платформе с {displayJoinedYear} года</span>
		</div>
	</div>

	<!-- Stats Bar matching screenshot style -->
	{#if listingsCount !== undefined || yearsOnPlatform !== undefined}
		<div class="grid grid-cols-2 rounded-2xl border border-border/70 overflow-hidden bg-muted/15 divide-x divide-border/70 py-2.5 text-center">
			{#if listingsCount !== undefined}
				<div>
					<div class="text-base font-extrabold text-foreground">{listingsCount}</div>
					<div class="text-[10px] text-muted-foreground">объявления</div>
				</div>
			{/if}
			{#if yearsOnPlatform !== undefined}
				<div>
					<div class="text-base font-extrabold text-foreground">{yearsOnPlatform}</div>
					<div class="text-[10px] text-muted-foreground">года на платформе</div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Footer Link -->
	{#if host?.id}
		<a
			href={`/users/${host.id}`}
			class="flex items-center justify-between pt-3 border-t border-border/70 text-xs font-semibold text-foreground/80 hover:text-foreground transition-colors"
		>
			<span>Посмотреть профиль</span>
			<ChevronRight class="h-3.5 w-3.5 text-muted-foreground" />
		</a>
	{/if}
</div>
