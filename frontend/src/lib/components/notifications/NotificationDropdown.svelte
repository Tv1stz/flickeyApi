<!-- src/lib/components/notifications/NotificationDropdown.svelte -->
<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import { notifications } from '$lib/stores/notifications.svelte';
	import NotificationItem from './NotificationItem.svelte';
	import { CheckCheck, Bell, Sparkles, Inbox } from 'lucide-svelte';
	import { scale, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	interface Props {
		open: boolean;
		onClose: () => void;
	}

	let { open = false, onClose }: Props = $props();

	let activeTab = $state<'all' | 'unread'>('all');
	let isSendingTest = $state(false);

	let items = $derived(notifications.items);
	let unreadCount = $derived(notifications.unreadCount);
	let hasUnread = $derived(notifications.hasUnread);
	let isLoading = $derived(notifications.isLoading);

	let displayedItems = $derived.by(() => {
		if (activeTab === 'unread') {
			return items.filter((item) => !item.is_read);
		}
		return items;
	});

	async function handleMarkAllAsRead() {
		await notifications.markAllAsRead();
	}

	async function handleSendTest() {
		isSendingTest = true;
		try {
			await notifications.sendDevTestNotification();
		} finally {
			isSendingTest = false;
		}
	}

	function handleViewAll() {
		onClose();
		goto(resolve('/notifications'));
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && open) {
			onClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div
		class="menu-surface absolute top-full right-0 z-[70] mt-3 w-84 sm:w-96 origin-top-right overflow-hidden rounded-3xl border border-zinc-200/80 bg-white/95 p-0 shadow-2xl backdrop-blur-2xl transition-all"
		in:fly={{ y: -12, duration: 300, easing: cubicOut }}
		out:scale={{ duration: 200, start: 0.96, opacity: 0, easing: cubicOut }}
		role="dialog"
		aria-label="Окно уведомлений"
	>
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-zinc-100 px-4 py-3.5 sm:px-5">
			<div class="flex items-center gap-2">
				<h3 class="text-[16px] font-bold tracking-tight text-zinc-900">Уведомления</h3>
				{#if unreadCount > 0}
					<span class="flex h-5 items-center justify-center rounded-full bg-blue-50 px-2 text-xs font-semibold text-blue-600">
						{unreadCount}
					</span>
				{/if}
			</div>

			{#if hasUnread}
				<button
					type="button"
					onclick={handleMarkAllAsRead}
					class="flex items-center gap-1.5 rounded-xl px-2.5 py-1 text-xs font-medium text-zinc-500 transition-colors hover:bg-zinc-100 hover:text-zinc-900 active:scale-95"
					title="Отметить все как прочитанные"
				>
					<CheckCheck class="h-3.5 w-3.5" strokeWidth={2} />
					<span>Прочитать все</span>
				</button>
			{/if}
		</div>

		<!-- Tabs -->
		<div class="flex border-b border-zinc-100 bg-zinc-50/60 px-3 py-1.5">
			<button
				type="button"
				onclick={() => (activeTab = 'all')}
				class="flex-1 rounded-xl py-1.5 text-center text-xs font-semibold transition-all
				       {activeTab === 'all' ? 'bg-white text-zinc-900 shadow-xs' : 'text-zinc-500 hover:text-zinc-800'}"
			>
				Все ({items.length})
			</button>
			<button
				type="button"
				onclick={() => (activeTab = 'unread')}
				class="flex-1 rounded-xl py-1.5 text-center text-xs font-semibold transition-all
				       {activeTab === 'unread' ? 'bg-white text-zinc-900 shadow-xs' : 'text-zinc-500 hover:text-zinc-800'}"
			>
				Непрочитанные ({unreadCount})
			</button>
		</div>

		<!-- Scrollable List -->
		<div class="max-h-[380px] sm:max-h-[420px] overflow-y-auto overscroll-contain p-2 space-y-1">
			{#if isLoading && items.length === 0}
				<div class="flex flex-col items-center justify-center py-12 text-center text-zinc-400">
					<div class="h-6 w-6 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"></div>
					<p class="mt-2 text-xs">Загрузка...</p>
				</div>
			{:else if displayedItems.length === 0}
				<div class="flex flex-col items-center justify-center py-12 text-center text-zinc-400">
					<div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400 mb-2.5">
						<Inbox class="h-6 w-6" strokeWidth={1.5} />
					</div>
					<p class="text-xs font-semibold text-zinc-700">
						{activeTab === 'unread' ? 'Нет непрочитанных уведомлений' : 'У вас пока нет уведомлений'}
					</p>
					<p class="mt-0.5 text-[11px] text-zinc-400">
						{activeTab === 'unread' ? 'Все уведомления прочитаны' : 'Здесь будут отображаться важные события и оповещения'}
					</p>
				</div>
			{:else}
				{#each displayedItems as item (item.id)}
					<NotificationItem notification={item} compact={true} onSelect={onClose} />
				{/each}
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex items-center justify-between border-t border-zinc-100 bg-zinc-50/70 px-4 py-2.5 sm:px-5">
			<button
				type="button"
				onclick={handleSendTest}
				disabled={isSendingTest}
				class="flex items-center gap-1.5 rounded-xl px-2 py-1 text-[11px] font-medium text-zinc-500 transition-colors hover:bg-zinc-200/70 hover:text-zinc-800 disabled:opacity-50"
				title="Отправить тестовое уведомление через SSE"
			>
				<Sparkles class="h-3.5 w-3.5 text-amber-500" strokeWidth={2} />
				<span>{isSendingTest ? 'Отправка...' : 'Тест SSE'}</span>
			</button>

			<button
				type="button"
				onclick={handleViewAll}
				class="text-xs font-semibold text-zinc-900 transition-colors hover:text-blue-600"
			>
				Все уведомления &rarr;
			</button>
		</div>
	</div>
{/if}
