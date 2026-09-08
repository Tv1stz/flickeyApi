<!-- src/lib/components/notifications/NotificationItem.svelte -->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { notifications } from '$lib/stores/notifications.svelte';
	import { formatRelativeTime } from '$lib/utils/format';
	import { getNotificationMeta, getNotificationLink } from './notificationConfig';
	import type { Notification } from '$lib/types/notification';
	import { Check, ChevronRight } from 'lucide-svelte';

	interface Props {
		notification: Notification;
		compact?: boolean;
		onSelect?: (notification: Notification) => void;
	}

	let { notification, compact = false, onSelect }: Props = $props();

	let meta = $derived(getNotificationMeta(notification.type));
	let link = $derived(getNotificationLink(notification));
	let relativeTime = $derived(formatRelativeTime(notification.created_at));
	let title = $derived(notification.title || meta.defaultTitle);
	let IconComponent = $derived(meta.icon);

	async function handleClick() {
		if (!notification.is_read) {
			notifications.markAsRead(notification.id);
		}

		onSelect?.(notification);

		if (link) {
			goto(resolve(link as any));
		}
	}

	async function handleMarkRead(e: MouseEvent) {
		e.stopPropagation();
		if (!notification.is_read) {
			await notifications.markAsRead(notification.id);
		}
	}
</script>

<div
	role="button"
	tabindex="0"
	onclick={handleClick}
	onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleClick()}
	class="group relative flex w-full cursor-pointer items-start gap-3.5 text-left transition-all duration-200 outline-none
	       {compact ? 'rounded-2xl p-3' : 'rounded-3xl p-4 sm:p-5 border border-zinc-100/80 shadow-xs'}
	       {notification.is_read ? 'bg-white hover:bg-zinc-50' : 'bg-blue-50/40 hover:bg-blue-50/70'}"
>
	<!-- Type Icon -->
	<div
		class="relative flex shrink-0 items-center justify-center rounded-2xl border transition-transform duration-200 group-hover:scale-105
		       {compact ? 'h-10 w-10 text-sm' : 'h-12 w-12 text-base'} {meta.bgLight}"
	>
		<IconComponent class="{compact ? 'h-5 w-5' : 'h-6 w-6'} {meta.iconClass}" strokeWidth={1.75} />
		{#if !notification.is_read}
			<span
				class="absolute -top-1 -right-1 flex h-3 w-3 items-center justify-center rounded-full bg-blue-500 ring-2 ring-white"
				title="Не прочитано"
			></span>
		{/if}
	</div>

	<!-- Content -->
	<div class="min-w-0 flex-1 pt-0.5">
		<div class="flex items-center justify-between gap-2">
			<h4
				class="truncate font-semibold tracking-tight text-zinc-900 transition-colors group-hover:text-zinc-950
				       {compact ? 'text-[13px] sm:text-[14px]' : 'text-[15px] sm:text-[16px]'}"
			>
				{title}
			</h4>
			<span class="shrink-0 text-[11px] font-medium text-zinc-400">
				{relativeTime}
			</span>
		</div>

		<p
			class="mt-1 line-clamp-2 leading-relaxed text-zinc-600
			       {compact ? 'text-[12px] sm:text-[13px]' : 'text-[13px] sm:text-[14px]'}"
		>
			{notification.message}
		</p>

		{#if link}
			<div class="mt-2 flex items-center gap-1 text-[11px] sm:text-xs font-semibold text-zinc-800 transition-colors group-hover:text-blue-600">
				<span>Перейти</span>
				<ChevronRight class="h-3.5 w-3.5 transition-transform duration-200 group-hover:translate-x-0.5" strokeWidth={2} />
			</div>
		{/if}
	</div>

	<!-- Action: Mark Read Button (hover on desktop or visible if unread) -->
	{#if !notification.is_read}
		<button
			type="button"
			class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-zinc-400 transition-all hover:bg-white hover:text-zinc-800 hover:shadow-xs"
			title="Отметить как прочитанное"
			onclick={handleMarkRead}
		>
			<Check class="h-4 w-4" strokeWidth={2} />
		</button>
	{/if}
</div>
