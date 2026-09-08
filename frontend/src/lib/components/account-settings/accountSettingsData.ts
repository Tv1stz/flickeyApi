import { Bell, CreditCard, HelpCircle, Shield, User, ShieldCheck } from 'lucide-svelte';
import type { User as AuthUser } from '$lib/stores/authStore.svelte';
import type { AccountNavItem, AccountSectionId, AccountSettingRow } from './types';

export const ACCOUNT_NAV_ITEMS: AccountNavItem[] = [
	{ id: 'personal', title: 'Личная информация', group: 'main', icon: User },
	{ id: 'security', title: 'Вход и безопасность', group: 'main', icon: Shield },
	{ id: 'payments', title: 'Платежи', group: 'main', icon: CreditCard },
	{ id: 'notifications', title: 'Уведомления', group: 'main', icon: Bell },
	{ id: 'help', title: 'Помощь', group: 'main', icon: HelpCircle }
];

export const ADMIN_NAV_ITEM: AccountNavItem = {
	id: 'admin',
	title: 'Администрирование',
	group: 'admin',
	icon: ShieldCheck
};

export function getNavItems(isAdmin = false): AccountNavItem[] {
	if (!isAdmin) return ACCOUNT_NAV_ITEMS;
	return [...ACCOUNT_NAV_ITEMS, ADMIN_NAV_ITEM];
}

const ACCOUNT_SECTION_IDS = new Set<AccountSectionId>([
	'personal',
	'security',
	'payments',
	'notifications',
	'help',
	'admin'
]);

export function isAccountSectionId(value: string): value is AccountSectionId {
	return ACCOUNT_SECTION_IDS.has(value as AccountSectionId);
}

export function formatMemberSince(createdAt?: string): string {
	if (!createdAt) return 'Недавно';
	const date = new Date(createdAt);
	if (Number.isNaN(date.getTime())) return 'Недавно';
	return new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric' }).format(date);
}

export function getUserInitials(name?: string): string {
	if (!name) return 'FL';
	const parts = name.split(' ').filter(Boolean);
	if (parts.length >= 2) return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
	return parts[0]?.slice(0, 2).toUpperCase() || 'FL';
}

export function buildRowsBySection(
	user: AuthUser | null,
	verified: boolean,
	memberSince: string
): Record<AccountSectionId, AccountSettingRow[]> {
	return {
		personal: [
			{
				id: 'name',
				label: 'Имя',
				value: user?.name || 'Не указано',
				action: { type: 'soon', feature: 'Имя', label: 'Редактировать' }
			},
			{
				id: 'phone',
				label: 'Номер телефона',
				value: user?.phone || 'Не указан',
				details: 'Подтверждение аккаунта',
				action: { type: 'soon', feature: 'Номер телефона', label: 'Редактировать' }
			},
			{
				id: 'identity',
				label: 'Подтверждение личности',
				value: verified ? 'Подтверждено' : 'Не начато',
				action: verified
					? { type: 'soon', feature: 'Подтверждение личности', label: 'Подробнее' }
					: { type: 'route', path: '/host/verification', label: 'Начать' }
			},
			{ id: 'memberSince', label: 'Дата регистрации', value: memberSince, action: { type: 'none' } }
		],
		security: [
			{
				id: 'password',
				label: 'Пароль',
				value: 'Вход по одноразовому коду / SMS',
				action: { type: 'soon', feature: 'Смена способа входа', label: 'Редактировать' }
			},
			{
				id: 'twoFactor',
				label: 'Двухфакторная аутентификация',
				value: 'Не настроена',
				action: { type: 'soon', feature: '2FA', label: 'Настроить' }
			},
			{
				id: 'sessions',
				label: 'История входов',
				value: 'Текущее устройство',
				action: { type: 'soon', feature: 'История входов', label: 'Управлять' }
			}
		],
		payments: [
			{
				id: 'payMethods',
				label: 'Способы оплаты',
				value: 'Карта не добавлена',
				action: { type: 'soon', feature: 'Способы оплаты', label: 'Добавить' }
			},
			{
				id: 'payouts',
				label: 'Реквизиты для выплат',
				value: 'Не настроены',
				action: { type: 'soon', feature: 'Реквизиты выплат', label: 'Редактировать' }
			}
		],
		notifications: [
			{
				id: 'notifPush',
				label: 'Push-уведомления',
				value: 'Включены',
				action: { type: 'soon', feature: 'Push-уведомления', label: 'Редактировать' }
			},
			{
				id: 'notifEmail',
				label: 'Email-уведомления',
				value: 'Основные оповещения о бронированиях',
				action: { type: 'soon', feature: 'Email-уведомления', label: 'Редактировать' }
			}
		],
		help: [
			{
				id: 'helpCenter',
				label: 'Центр помощи',
				value: 'Поддержка пользователей и ответы на частые вопросы',
				action: { type: 'route', path: '/help', label: 'Открыть' }
			}
		],
		admin: [
			{
				id: 'adminPanel',
				label: 'Панель администратора',
				value: 'Модерация объявлений, пользователи, верификация',
				action: { type: 'route', path: '/admin', label: 'Перейти' }
			}
		]
	};
}
