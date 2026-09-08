<!-- src/lib/components/notifications/NotificationBell.svelte -->
<script lang="ts">
	import { Bell } from 'lucide-svelte';
	import { notifications } from '$lib/stores/notifications.svelte';
	import NotificationDropdown from './NotificationDropdown.svelte';
	import { scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	let isOpen = $state(false);
	let unreadCount = $derived(notifications.unreadCount);
	let hasUnread = $derived(notifications.hasUnread);

	let formattedCount = $derived(unreadCount > 99 ? '99+' : String(unreadCount));

	function handleOutsideClick(e: PointerEvent) {
		if (isOpen) {
			const target = e.target as HTMLElement;
			if (!target.closest('[data-notification-menu]')) {
				isOpen = false;
			}
		}
	}
</script>

<svelte:window onpointerdown={handleOutsideClick} />

<div class="relative" data-notification-menu>
	<button
		type="button"
		class="group relative flex h-11 w-11 items-center justify-center rounded-full border border-zinc-200 bg-white/80 text-zinc-700 backdrop-blur-2xl transition-all duration-300 hover:border-zinc-300 hover:bg-white hover:text-zinc-950 hover:shadow-md active:scale-95
		       {isOpen ? 'border-zinc-400 bg-white shadow-md text-zinc-950' : ''}"
		onclick={() => (isOpen = !isOpen)}
		aria-label="Уведомления"
		aria-haspopup="dialog"
		aria-expanded={isOpen}
	>
		<Bell
			class="h-5 w-5 transition-transform duration-300 group-hover:scale-105 {hasUnread ? 'text-zinc-900' : 'text-zinc-600'}"
			strokeWidth={1.75}
		/>

		{#if hasUnread}
			<span
				class="absolute -top-1 -right-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-rose-500 px-1 text-[10px] font-bold text-white shadow-sm ring-2 ring-white"
				in:scale={{ duration: 300, easing: cubicOut }}
			>
				{formattedCount}
			</span>
		{/if}
	</button>

	<NotificationDropdown open={isOpen} onClose={() => (isOpen = false)} />
</div>
