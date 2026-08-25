<script lang="ts">
	import { notifications } from '$lib/stores/notifications.svelte';
	import { formatRelativeTime } from '$lib/utils';
	import type { Notification } from '$lib/types/notification';
	import {
		CheckCircle2,
		AlertCircle,
		ShieldAlert,
		Calendar,
		Bell,
		BellOff,
		CheckCheck,
		Loader2
	} from 'lucide-svelte';

	interface Props {
		open: boolean;
		onclose: () => void;
	}

	let { open, onclose }: Props = $props();

	function getNotificationIcon(type: string) {
		switch (type) {
			case 'listing_approved':
			case 'verification_approved':
				return { component: CheckCircle2, color: 'text-emerald-600 dark:text-emerald-400 bg-emerald-100 dark:bg-emerald-950/60' };
			case 'listing_rejected':
			case 'verification_rejected':
				return { component: AlertCircle, color: 'text-rose-600 dark:text-rose-400 bg-rose-100 dark:bg-rose-950/60' };
			case 'enforcement_issued':
			case 'verification_changes_requested':
				return { component: ShieldAlert, color: 'text-amber-600 dark:text-amber-400 bg-amber-100 dark:bg-amber-950/60' };
			case 'booking_created':
			case 'booking_status_changed':
				return { component: Calendar, color: 'text-sky-600 dark:text-sky-400 bg-sky-100 dark:bg-sky-950/60' };
			default:
				return { component: Bell, color: 'text-primary bg-primary/10' };
		}
	}

	async function handleItemClick(item: Notification) {
		if (!item.is_read) {
			await notifications.markAsRead(item.id);
		}
	}
</script>

{#if open}
	<!-- Backdrop -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="fixed inset-0 z-40" onclick={onclose}></div>

	<!-- Dropdown Panel -->
	<div
		class="absolute right-0 top-full mt-2 w-80 sm:w-96 z-50 rounded-2xl border border-border bg-card shadow-2xl animate-in zoom-in-95 overflow-hidden flex flex-col max-h-[520px]"
	>
		<!-- Header -->
		<div class="px-4 py-3 border-b border-border/80 flex items-center justify-between bg-card/95 backdrop-blur-xs shrink-0">
			<div class="flex items-center gap-2">
				<h3 class="text-xs font-extrabold text-foreground">Уведомления</h3>
				{#if notifications.unreadCount > 0}
					<span class="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-primary/15 text-primary">
						{notifications.unreadCount} новых
					</span>
				{/if}
			</div>

			{#if notifications.unreadCount > 0}
				<button
					type="button"
					onclick={() => notifications.markAllAsRead()}
					class="text-[11px] font-semibold text-primary hover:text-primary/80 transition-colors flex items-center gap-1 cursor-pointer"
				>
					<CheckCheck class="h-3 w-3" />
					<span>Прочитать все</span>
				</button>
			{/if}
		</div>

		<!-- List Content -->
		<div class="overflow-y-auto flex-1 divide-y divide-border/60">
			{#if notifications.isLoading && notifications.items.length === 0}
				<div class="py-12 text-center space-y-2">
					<Loader2 class="h-5 w-5 text-primary animate-spin mx-auto" />
					<p class="text-xs text-muted-foreground">Загрузка уведомлений...</p>
				</div>
			{:else if notifications.items.length === 0}
				<div class="py-12 px-4 text-center space-y-2">
					<div class="w-10 h-10 rounded-2xl bg-muted flex items-center justify-center mx-auto text-muted-foreground">
						<BellOff class="h-5 w-5" />
					</div>
					<h4 class="text-xs font-bold text-foreground">Уведомлений нет</h4>
					<p class="text-[11px] text-muted-foreground max-w-[200px] mx-auto">
						Здесь будут появляться статусы ваших бронирований, модерации и важные сообщения.
					</p>
				</div>
			{:else}
				{#each notifications.items as item (item.id)}
					{@const iconInfo = getNotificationIcon(item.type)}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						onclick={() => handleItemClick(item)}
						class="p-3.5 hover:bg-accent/60 transition-colors cursor-pointer flex gap-3 text-left relative {item.is_read ? 'opacity-80' : 'bg-primary/[0.03]'}"
					>
						<!-- Icon -->
						<div class="w-8 h-8 rounded-xl flex items-center justify-center shrink-0 mt-0.5 {iconInfo.color}">
							<iconInfo.component class="h-4 w-4" />
						</div>

						<!-- Content -->
						<div class="min-w-0 flex-1 space-y-0.5">
							<div class="flex items-start justify-between gap-2">
								<p class="text-xs font-bold text-foreground leading-snug {item.is_read ? 'font-medium' : 'font-bold'}">
									{item.title}
								</p>
								{#if !item.is_read}
									<span class="w-2 h-2 rounded-full bg-primary shrink-0 mt-1" title="Новое"></span>
								{/if}
							</div>
							<p class="text-[11px] text-muted-foreground leading-relaxed line-clamp-3">
								{item.message}
							</p>
							<p class="text-[10px] text-muted-foreground/70 font-mono pt-1">
								{formatRelativeTime(item.created_at)}
							</p>
						</div>
					</div>
				{/each}
			{/if}
		</div>
	</div>
{/if}
