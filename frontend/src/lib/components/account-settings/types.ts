import type { ComponentType } from 'svelte';

export type AccountSectionId =
	| 'personal'
	| 'security'
	| 'payments'
	| 'notifications'
	| 'help'
	| 'admin';

export type AccountSectionGroup = 'main' | 'admin';
export type AccountSettingsRoute = '/help' | '/admin' | '/host/listings' | '/host/verification';

export type AccountSettingAction =
	| {
			type: 'route';
			path: AccountSettingsRoute;
			label: string;
	  }
	| {
			type: 'soon';
			feature: string;
			label: string;
	  }
	| {
			type: 'none';
	  };

export interface AccountSettingRow {
	id: string;
	label: string;
	value: string;
	details?: string;
	action?: AccountSettingAction;
}

export interface AccountNavItem {
	id: AccountSectionId;
	title: string;
	group: AccountSectionGroup;
	icon: ComponentType;
}
