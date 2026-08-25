import { apiRequest } from '../client';
import type { ReportItem } from '$lib/types/admin';

export interface ResolveReportPayload {
	action: 'resolve' | 'dismiss';
	reason?: string;
	note?: string;
}

export const adminReportsApi = {
	/**
	 * Fetch reports queue.
	 */
	getReports(status?: string): Promise<ReportItem[]> {
		const query = status ? `?status=${encodeURIComponent(status)}` : '';
		return apiRequest<ReportItem[]>(`/admin/reports${query}`, {
			method: 'GET'
		});
	},

	/**
	 * Resolve or dismiss a complaint.
	 */
	resolveReport(id: string, payload: ResolveReportPayload): Promise<ReportItem> {
		return apiRequest<ReportItem>(`/admin/reports/${id}/resolve`, {
			method: 'POST',
			body: payload
		});
	},

	/**
	 * Helper: Resolve valid report.
	 */
	confirmReport(id: string, reason: string, note?: string): Promise<ReportItem> {
		return this.resolveReport(id, { action: 'resolve', reason, note });
	},

	/**
	 * Helper: Dismiss unsubstantiated report.
	 */
	dismissReport(id: string, reason: string, note?: string): Promise<ReportItem> {
		return this.resolveReport(id, { action: 'dismiss', reason, note });
	}
};
