import { apiRequest } from '../client';
import type { AuditLogEntry } from '$lib/types/admin';

export const adminAuditApi = {
	/**
	 * Fetch read-only immutable administrative audit logs.
	 */
	getAuditLogs(): Promise<AuditLogEntry[]> {
		return apiRequest<AuditLogEntry[]>('/admin/audit', {
			method: 'GET'
		});
	}
};
