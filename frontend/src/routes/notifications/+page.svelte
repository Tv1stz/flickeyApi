<!-- src/routes/notifications/+page.svelte -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authStore } from '$lib/stores/authStore.svelte';
	import { notifications } from '$lib/stores/notifications.svelte';
	import NotificationItem from '$lib/components/notifications/NotificationItem.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Bell, CheckCheck, Sparkles, Inbox, ArrowLeft } from 'lucide-svelte';
	import { fade } from 'svelte/transition';

	let authInitialized = $derived(authStore.initialized);
	let authenticated = $derived(authStore.isAuthenticated);
	let redirectingToAuth = $state(false);
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

	$effect(() => {
		if (!authStore.initialized || authStore.user || redirectingToAuth) return;
		redirectingToAuth = true;
		authStore.setPendingAction(null, '/notifications');
		goto(resolve('/auth'));
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
</script>

<svelte:head>
	<title>Уведомления — Flickey</title>
</svelte:head>

{#if !authInitialized}
	<div class="flex min-h-[60vh] items-center justify-center bg-white">
		<div class="flex flex-col items-center gap-3">
			<div class="h-8 w-8 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"></div>
			<p class="text-sm text-zinc-500">Загрузка уведомлений...</p>
		</div>
	</div>
{:else if !authenticated}
	<div class="flex min-h-[60vh] items-center justify-center bg-white">
		<p class="text-sm text-zinc-500">Перенаправляем на страницу входа...</p>
	</div>
{:else}
	<div class="min-h-screen bg-zinc-50/40">
		<div class="mx-auto max-w-4xl px-4 pt-6 pb-32 sm:px-6 lg:px-8 lg:pt-10 lg:pb-20">
			<!-- Breadcrumb / Back button on mobile -->
			<div class="mb-4 flex items-center justify-between">
				<a
					href={resolve('/profile')}
					class="inline-flex items-center gap-1.5 text-xs font-medium text-zinc-500 transition-colors hover:text-zinc-900"
				>
					<ArrowLeft class="h-3.5 w-3.5" strokeWidth={2} />
					<span>В профиль</span>
				</a>

				<button
					type="button"
					onclick={handleSendTest}
					disabled={isSendingTest}
					class="inline-flex items-center gap-1.5 rounded-full border border-amber-200 bg-amber-50/80 px-3 py-1 text-xs font-semibold text-amber-800 transition-all hover:bg-amber-100 disabled:opacity-50"
					title="Отправить тестовое уведомление через SSE"
				>
					<Sparkles class="h-3.5 w-3.5 text-amber-600" strokeWidth={2} />
					<span>{isSendingTest ? 'Отправка...' : 'Тест SSE'}</span>
				</button>
			</div>

			<!-- Header -->
			<div class="rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-xs sm:p-8">
				<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-900">
							<Bell class="h-6 w-6" strokeWidth={1.75} />
						</div>
						<div>
							<div class="flex items-center gap-2.5">
								<h1 class="text-2xl font-bold tracking-tight text-zinc-900 sm:text-3xl">Уведомления</h1>
								{#if unreadCount > 0}
									<span class="rounded-full bg-rose-500 px-2.5 py-0.5 text-xs font-bold text-white shadow-xs">
										{unreadCount}
									</span>
								{/if}
							</div>
							<p class="mt-0.5 text-xs text-zinc-500 sm:text-sm">
								История оповещений, бронирований и сообщений
							</p>
						</div>
					</div>

					{#if hasUnread}
						<Button
							variant="outline"
							tone="neutral"
							size="md"
							radius="xl"
							iconLeft={CheckCheck}
							onclick={handleMarkAllAsRead}
						>
							Прочитать все
						</Button>
					{/if}
				</div>

				<!-- Tabs Bar -->
				<div class="mt-6 flex border-b border-zinc-100">
					<button
						type="button"
						onclick={() => (activeTab = 'all')}
						class="relative px-4 py-2.5 text-sm font-semibold transition-colors
						       {activeTab === 'all' ? 'text-zinc-900' : 'text-zinc-500 hover:text-zinc-800'}"
					>
						Все ({items.length})
						{#if activeTab === 'all'}
							<span class="absolute bottom-0 left-0 right-0 h-0.5 bg-zinc-900"></span>
						{/if}
					</button>

					<button
						type="button"
						onclick={() => (activeTab = 'unread')}
						class="relative px-4 py-2.5 text-sm font-semibold transition-colors
						       {activeTab === 'unread' ? 'text-zinc-900' : 'text-zinc-500 hover:text-zinc-800'}"
					>
						Непрочитанные ({unreadCount})
						{#if activeTab === 'unread'}
							<span class="absolute bottom-0 left-0 right-0 h-0.5 bg-zinc-900"></span>
						{/if}
					</button>
				</div>

				<!-- Notifications Feed -->
				<div class="mt-6 space-y-3">
					{#if isLoading && items.length === 0}
						<div class="flex flex-col items-center justify-center py-16 text-center text-zinc-400">
							<div class="h-8 w-8 animate-spin rounded-full border-2 border-zinc-900 border-t-transparent"></div>
							<p class="mt-3 text-sm">Загрузка уведомлений...</p>
						</div>
					{:else if displayedItems.length === 0}
						<div class="flex flex-col items-center justify-center py-16 text-center" in:fade={{ duration: 200 }}>
							<div class="mb-4 flex h-16 w-16 items-center justify-center rounded-3xl bg-zinc-100 text-zinc-400">
								<Inbox class="h-8 w-8" strokeWidth={1.5} />
							</div>
							<h3 class="text-base font-semibold text-zinc-900">
								{activeTab === 'unread' ? 'Все уведомления прочитаны' : 'У вас пока нет уведомлений'}
							</h3>
							<p class="mt-1 max-w-sm text-xs leading-relaxed text-zinc-500 sm:text-sm">
								{activeTab === 'unread'
									? 'Здесь появятся новые события и оповещения, требующие вашего внимания'
									: 'Здесь будут появляться оповещения о бронированиях, проверке документов и статусе объявлений'}
							</p>
						</div>
					{:else}
						{#each displayedItems as item (item.id)}
							<NotificationItem notification={item} compact={false} />
						{/each}
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}
