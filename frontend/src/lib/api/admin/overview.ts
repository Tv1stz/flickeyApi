import { apiRequest } from '../client';
import type { AdminOverviewMetrics } from '$lib/types/admin';

export const adminOverviewApi = {
	/**
	 * Get overview stats & counters.
	 */
	getOverview(): Promise<AdminOverviewMetrics> {
		return apiRequest<AdminOverviewMetrics>('/admin/overview', {
			method: 'GET'
		});
	}
};
