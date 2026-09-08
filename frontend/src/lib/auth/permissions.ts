import type { User, UserRole, ViewMode } from '$lib/stores/authStore.svelte';

export function hasRole(user: User | null | undefined, role: UserRole): boolean {
	return Boolean(user?.roles.includes(role));
}

export function isHostUser(user: User | null | undefined): boolean {
	return (hasRole(user, 'host') || hasRole(user, 'admin')) && Boolean(user?.hostProfile);
}

export function isAdminUser(user: User | null | undefined): boolean {
	return hasRole(user, 'admin');
}

export function canAccessHostArea(user: User | null | undefined): boolean {
	return isHostUser(user) || isAdminUser(user);
}

export function normalizeViewModeForUser(
	user: User | null | undefined,
	requestedMode: ViewMode | null | undefined
): ViewMode {
	if (requestedMode === 'host' && canAccessHostArea(user)) {
		return 'host';
	}

	return 'guest';
}
