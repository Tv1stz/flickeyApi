<script lang="ts">
	import { notifications } from '$lib/stores/notifications.svelte';
	import NotificationDropdown from './NotificationDropdown.svelte';
	import { Bell } from 'lucide-svelte';

	let isDropdownOpen = $state(false);

	function toggleDropdown() {
		isDropdownOpen = !isDropdownOpen;
	}

	function closeDropdown() {
		isDropdownOpen = false;
	}
</script>

<div class="relative">
	<button
		type="button"
		onclick={toggleDropdown}
		class="relative flex h-9 w-9 items-center justify-center rounded-xl border border-border bg-card text-muted-foreground hover:bg-accent hover:text-foreground transition-colors cursor-pointer"
		aria-label="Уведомления"
		aria-expanded={isDropdownOpen}
	>
		<Bell class="h-4 w-4" />

		{#if notifications.unreadCount > 0}
			<span
				class="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-rose-600 px-1 text-[10px] font-black text-white shadow-xs animate-in zoom-in"
			>
				{notifications.unreadCount > 99 ? '99+' : notifications.unreadCount}
			</span>
		{/if}
	</button>

	<NotificationDropdown open={isDropdownOpen} onclose={closeDropdown} />
</div>
