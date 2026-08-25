import { API_BASE } from '$lib/api/client';
import { notificationsApi } from '$lib/api/notifications';
import { toast } from '$lib/components/ui/toast.svelte';
import type { Notification } from '$lib/types/notification';

class NotificationStore {
	items = $state<Notification[]>([]);
	unreadCount = $state<number>(0);
	total = $state<number>(0);
	isLoading = $state<boolean>(false);
	isConnected = $state<boolean>(false);

	private eventSource: EventSource | null = null;

	hasUnread = $derived(this.unreadCount > 0);

	/**
	 * Fetch initial list of notifications and unread count.
	 */
	async fetchInitial(): Promise<void> {
		this.isLoading = true;
		try {
			const res = await notificationsApi.getNotifications({ limit: 30 });
			this.items = res.items || [];
			this.total = res.total || 0;
			this.unreadCount = res.unread_count || 0;
		} catch (err) {
			console.error('Failed to fetch notifications:', err);
		} finally {
			this.isLoading = false;
		}
	}

	/**
	 * Connect to Server-Sent Events stream using access token.
	 */
	connect(token: string): void {
		if (typeof window === 'undefined' || !token) return;

		// If existing connection is already active, do not recreate
		if (this.eventSource && this.isConnected) {
			return;
		}

		this.disconnect();

		const streamUrl = `${API_BASE}/notifications/stream?token=${encodeURIComponent(token)}`;
		this.eventSource = new EventSource(streamUrl);

		this.eventSource.onopen = () => {
			this.isConnected = true;
		};

		this.eventSource.addEventListener('notification', (event: MessageEvent) => {
			try {
				const notif: Notification = JSON.parse(event.data);

				// Prevent duplicate insertions
				const exists = this.items.some((i) => i.id === notif.id);
				if (!exists) {
					this.items = [notif, ...this.items];
					this.unreadCount += 1;
					this.total += 1;
				}

				this.triggerToast(notif);
			} catch (err) {
				console.error('Failed to parse SSE notification:', err);
			}
		});

		this.eventSource.onerror = () => {
			this.isConnected = false;
			// Refetch on reconnection or error to sync missed notifications
			this.fetchInitial().catch(() => {});
		};

		// Initial data load
		this.fetchInitial();
	}

	/**
	 * Disconnect from the SSE stream and clear state.
	 */
	disconnect(): void {
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}
		this.isConnected = false;
		this.items = [];
		this.unreadCount = 0;
		this.total = 0;
	}

	/**
	 * Mark a specific notification as read with optimistic update.
	 */
	async markAsRead(id: string): Promise<void> {
		const target = this.items.find((i) => i.id === id);
		if (target && !target.is_read) {
			target.is_read = true;
			target.read_at = new Date().toISOString();
			if (this.unreadCount > 0) {
				this.unreadCount -= 1;
			}
		}

		try {
			await notificationsApi.markAsRead(id);
		} catch (err) {
			console.error('Failed to mark notification as read on server:', err);
		}
	}

	/**
	 * Mark all notifications as read with optimistic update.
	 */
	async markAllAsRead(): Promise<void> {
		const now = new Date().toISOString();
		for (const item of this.items) {
			if (!item.is_read) {
				item.is_read = true;
				item.read_at = now;
			}
		}
		this.unreadCount = 0;

		try {
			await notificationsApi.markAllAsRead();
		} catch (err) {
			console.error('Failed to mark all notifications as read on server:', err);
		}
	}

	/**
	 * Display an in-app toast based on the notification type.
	 */
	private triggerToast(notif: Notification): void {
		switch (notif.type) {
			case 'listing_approved':
			case 'verification_approved':
				toast.success(notif.message, notif.title || 'Одобрено');
				break;
			case 'listing_rejected':
			case 'verification_rejected':
				toast.error(notif.message, notif.title || 'Отклонено');
				break;
			case 'enforcement_issued':
			case 'verification_changes_requested':
				toast.warning(notif.message, notif.title || 'Внимание');
				break;
			default:
				toast.info(notif.message, notif.title || 'Уведомление');
				break;
		}
	}
}

export const notifications = new NotificationStore();
