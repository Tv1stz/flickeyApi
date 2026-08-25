import { apiRequest } from '../client';
import type { User } from '$lib/types/auth';
import type { ListingPublic } from '$lib/types/listings';
import type { EnforcementPayload, UserRestrictionRecord } from '$lib/types/admin';

export interface UserDetailResponse {
	user: User;
	listings: ListingPublic[];
	restrictions: UserRestrictionRecord[];
}

export const adminUsersApi = {
	/**
	 * Fetch user directory.
	 */
	getUsers(role?: string, status?: string): Promise<User[]> {
		const params = new URLSearchParams();
		if (role) params.set('role', role);
		if (status) params.set('status', status);
		const q = params.toString() ? `?${params.toString()}` : '';
		return apiRequest<User[]>(`/admin/users${q}`, {
			method: 'GET'
		});
	},

	/**
	 * Fetch detailed user profile and restriction history.
	 */
	getUserDetail(id: string): Promise<UserDetailResponse> {
		return apiRequest<UserDetailResponse>(`/admin/users/${id}`, {
			method: 'GET'
		});
	},

	/**
	 * Apply administrative enforcement action (warn, restrict, block, unblock).
	 */
	enforceUser(id: string, payload: EnforcementPayload): Promise<User> {
		return apiRequest<User>(`/admin/users/${id}/enforce`, {
			method: 'POST',
			body: payload
		});
	},

	/**
	 * Helper: Warn user.
	 */
	warnUser(id: string, reason: string, note?: string): Promise<User> {
		return this.enforceUser(id, {
			enforcement_type: 'warning',
			reason,
			note
		});
	},

	/**
	 * Helper: Temporarily restrict user.
	 */
	restrictUser(id: string, reason: string, note?: string, durationDays: number = 7): Promise<User> {
		return this.enforceUser(id, {
			enforcement_type: 'temporary_restriction',
			reason,
			note,
			duration_days: durationDays
		});
	},

	/**
	 * Helper: Block user (temporary or permanent).
	 */
	blockUser(
		id: string,
		isPermanent: boolean,
		reason: string,
		note?: string,
		durationDays: number = 30
	): Promise<User> {
		return this.enforceUser(id, {
			enforcement_type: isPermanent ? 'permanent_block' : 'temporary_block',
			reason,
			note,
			duration_days: isPermanent ? undefined : durationDays
		});
	},

	/**
	 * Helper: Unblock / revoke restrictions.
	 */
	unblockUser(id: string, reason: string = 'Административная отмена ограничений', note?: string): Promise<User> {
		return this.enforceUser(id, {
			enforcement_type: 'unblock',
			reason,
			note
		});
	},

	/**
	 * Update user role (backed by backend authority).
	 */
	updateUserRole(id: string, role: string): Promise<User> {
		return apiRequest<User>(`/admin/users/${id}/role`, {
			method: 'POST',
			body: { role }
		});
	}
};
