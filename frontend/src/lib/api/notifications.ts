import { apiRequest } from './client';
import type { Notification, NotificationListResponse, UnreadCountResponse } from '$lib/types/notification';

export interface GetNotificationsParams {
	limit?: number;
	offset?: number;
	unread_only?: boolean;
}

export const notificationsApi = {
	/**
	 * Get paginated list of user notifications.
	 */
	getNotifications(params: GetNotificationsParams = {}): Promise<NotificationListResponse> {
		const searchParams = new URLSearchParams();
		if (params.limit !== undefined) searchParams.set('limit', params.limit.toString());
		if (params.offset !== undefined) searchParams.set('offset', params.offset.toString());
		if (params.unread_only !== undefined) searchParams.set('unread_only', params.unread_only.toString());

		const query = searchParams.toString() ? `?${searchParams.toString()}` : '';
		return apiRequest<NotificationListResponse>(`/notifications${query}`, {
			method: 'GET'
		});
	},

	/**
	 * Get unread notification count.
	 */
	getUnreadCount(): Promise<UnreadCountResponse> {
		return apiRequest<UnreadCountResponse>('/notifications/unread-count', {
			method: 'GET'
		});
	},

	/**
	 * Mark a single notification as read.
	 */
	markAsRead(id: string): Promise<{ success: boolean }> {
		return apiRequest<{ success: boolean }>(`/notifications/${id}/read`, {
			method: 'PATCH'
		});
	},

	/**
	 * Mark all unread notifications as read.
	 */
	markAllAsRead(): Promise<{ success: boolean }> {
		return apiRequest<{ success: boolean }>('/notifications/read-all', {
			method: 'POST'
		});
	},

	/**
	 * Send test notification (development only).
	 */
	sendTestNotification(type?: string): Promise<Notification> {
		return apiRequest<Notification>('/notifications/test', {
			method: 'POST',
			body: JSON.stringify(type ? { type } : {})
		});
	}
};
