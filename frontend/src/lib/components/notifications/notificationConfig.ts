// src/lib/components/notifications/notificationConfig.ts
import {
	CheckCircle2,
	XCircle,
	AlertTriangle,
	Calendar,
	CalendarCheck,
	MessageSquare,
	ShieldCheck,
	ShieldAlert,
	Bell,
	type Icon
} from 'lucide-svelte';
import type { Notification, NotificationType } from '$lib/types/notification';

export interface NotificationMeta {
	icon: typeof Icon;
	iconClass: string;
	badgeClass: string;
	bgLight: string;
	defaultTitle: string;
	getLink?: (payload: Record<string, any>) => string | null;
}

export const NOTIFICATION_CONFIG: Record<string, NotificationMeta> = {
	listing_approved: {
		icon: CheckCircle2,
		iconClass: 'text-emerald-600',
		badgeClass: 'bg-emerald-500',
		bgLight: 'bg-emerald-50 text-emerald-700 border-emerald-100',
		defaultTitle: 'Объявление опубликовано',
		getLink: (payload) => (payload?.listing_id ? `/listings/${payload.listing_id}` : '/host/listings')
	},
	listing_rejected: {
		icon: XCircle,
		iconClass: 'text-rose-600',
		badgeClass: 'bg-rose-500',
		bgLight: 'bg-rose-50 text-rose-700 border-rose-100',
		defaultTitle: 'Объявление отклонено',
		getLink: () => '/host/listings'
	},
	enforcement_issued: {
		icon: AlertTriangle,
		iconClass: 'text-amber-600',
		badgeClass: 'bg-amber-500',
		bgLight: 'bg-amber-50 text-amber-700 border-amber-100',
		defaultTitle: 'Предупреждение модерации',
		getLink: () => '/profile'
	},
	booking_created: {
		icon: Calendar,
		iconClass: 'text-blue-600',
		badgeClass: 'bg-blue-500',
		bgLight: 'bg-blue-50 text-blue-700 border-blue-100',
		defaultTitle: 'Новое бронирование',
		getLink: () => '/host/calendar'
	},
	booking_status_changed: {
		icon: CalendarCheck,
		iconClass: 'text-indigo-600',
		badgeClass: 'bg-indigo-500',
		bgLight: 'bg-indigo-50 text-indigo-700 border-indigo-100',
		defaultTitle: 'Статус бронирования изменен',
		getLink: () => '/host/calendar'
	},
	new_message: {
		icon: MessageSquare,
		iconClass: 'text-violet-600',
		badgeClass: 'bg-violet-500',
		bgLight: 'bg-violet-50 text-violet-700 border-violet-100',
		defaultTitle: 'Новое сообщение',
		getLink: () => '/profile'
	},
	verification_approved: {
		icon: ShieldCheck,
		iconClass: 'text-emerald-600',
		badgeClass: 'bg-emerald-500',
		bgLight: 'bg-emerald-50 text-emerald-700 border-emerald-100',
		defaultTitle: 'Верификация подтверждена',
		getLink: () => '/host/verification'
	},
	verification_rejected: {
		icon: ShieldAlert,
		iconClass: 'text-rose-600',
		badgeClass: 'bg-rose-500',
		bgLight: 'bg-rose-50 text-rose-700 border-rose-100',
		defaultTitle: 'Верификация отклонена',
		getLink: () => '/host/verification'
	},
	verification_changes_requested: {
		icon: AlertTriangle,
		iconClass: 'text-amber-600',
		badgeClass: 'bg-amber-500',
		bgLight: 'bg-amber-50 text-amber-700 border-amber-100',
		defaultTitle: 'Требуются изменения в заявке',
		getLink: () => '/host/verification'
	}
};

const DEFAULT_META: NotificationMeta = {
	icon: Bell,
	iconClass: 'text-zinc-600',
	badgeClass: 'bg-zinc-500',
	bgLight: 'bg-zinc-100 text-zinc-700 border-zinc-200',
	defaultTitle: 'Уведомление',
	getLink: () => null
};

export function getNotificationMeta(type: NotificationType | string): NotificationMeta {
	return NOTIFICATION_CONFIG[type] || DEFAULT_META;
}

export function getNotificationLink(notification: Notification): string | null {
	const meta = getNotificationMeta(notification.type);
	return meta.getLink ? meta.getLink(notification.payload || {}) : null;
}
