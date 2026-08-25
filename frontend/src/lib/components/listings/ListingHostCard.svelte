<script lang="ts">
	import type { HostInfo } from '$lib/types/listings';
	import { ShieldCheck, MessageCircle, Star, Phone } from 'lucide-svelte';
	import { maskPhone } from '$lib/utils';

	interface Props {
		host?: HostInfo | null;
		hostName?: string;
		isVerified?: boolean;
	}

	let { host, hostName, isVerified = true }: Props = $props();

	let name = $derived(
		host?.name ||
		(host?.first_name ? `${host.first_name}${host.last_name ? ' ' + host.last_name : ''}` : null) ||
		host?.phone ||
		hostName ||
		'Арендодатель'
	);
	let phone = $derived(host?.phone ? maskPhone(host.phone) : null);
	let roleLabel = $derived(host?.role === 'admin' ? 'Администратор' : host?.role === 'host' ? 'Хост' : 'Пользователь');
</script>

<div class="rounded-2xl border border-border bg-card p-5 shadow-sm space-y-4">
	<div class="flex items-center gap-3">
		<div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary font-bold text-lg">
			{name ? name[0].toUpperCase() : 'H'}
		</div>
		<div>
			<h4 class="font-bold text-sm text-foreground">{name}</h4>
			{#if phone}
				<p class="text-xs text-muted-foreground font-mono flex items-center gap-1 mt-0.5">
					<Phone class="h-3 w-3" /> {phone}
				</p>
			{/if}
			{#if isVerified}
				<div class="flex items-center gap-1 text-[11px] text-emerald-600 font-semibold mt-1">
					<ShieldCheck class="h-3.5 w-3.5" /> Проверенный хост ({roleLabel})
				</div>
			{/if}
		</div>
	</div>

	<div class="pt-2 border-t border-border/80 flex items-center justify-between text-xs text-muted-foreground">
		<span class="flex items-center gap-1 text-amber-500 font-semibold">
			<Star class="h-3.5 w-3.5 fill-amber-500" /> 5.0 (Новый объект)
		</span>
		<span class="flex items-center gap-1">
			<MessageCircle class="h-3.5 w-3.5" /> Быстрый ответ
		</span>
	</div>
</div>
