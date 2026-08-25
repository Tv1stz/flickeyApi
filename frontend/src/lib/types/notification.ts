export type NotificationType =
	| 'listing_rejected'
	| 'listing_approved'
	| 'enforcement_issued'
	| 'booking_created'
	| 'booking_status_changed'
	| 'new_message'
	| 'verification_approved'
	| 'verification_rejected'
	| 'verification_changes_requested'
	| (string & {});

export interface Notification {
	id: string;
	user_id: string;
	type: NotificationType;
	title: string;
	message: string;
	payload: Record<string, any>;
	is_read: boolean;
	read_at: string | null;
	created_at: string;
}

export interface NotificationListResponse {
	items: Notification[];
	total: number;
	unread_count: number;
}

export interface UnreadCountResponse {
	unread_count: number;
}
