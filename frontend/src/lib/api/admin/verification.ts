import { apiRequest } from '../client';
import type { VerificationRequestItem } from '$lib/types/admin';

export interface VerificationReviewPayload {
	action: 'approve' | 'reject' | 'request_changes';
	reason?: string;
	note?: string;
}

export const adminVerificationApi = {
	/**
	 * Fetch verification request queue.
	 */
	getVerificationQueue(status?: string): Promise<VerificationRequestItem[]> {
		const query = status ? `?status=${encodeURIComponent(status)}` : '';
		return apiRequest<VerificationRequestItem[]>(`/admin/verification${query}`, {
			method: 'GET'
		});
	},

	/**
	 * Fetch verification detail.
	 */
	getVerificationDetail(id: string): Promise<VerificationRequestItem> {
		return apiRequest<VerificationRequestItem>(`/admin/verification/${id}`, {
			method: 'GET'
		});
	},

	/**
	 * Review verification application.
	 */
	reviewVerification(id: string, payload: VerificationReviewPayload): Promise<VerificationRequestItem> {
		return apiRequest<VerificationRequestItem>(`/admin/verification/${id}/review`, {
			method: 'POST',
			body: payload
		});
	},

	/**
	 * Helper: Approve verification.
	 */
	approveVerification(id: string, note?: string): Promise<VerificationRequestItem> {
		return this.reviewVerification(id, { action: 'approve', note });
	},

	/**
	 * Helper: Reject verification.
	 */
	rejectVerification(id: string, reason: string, note?: string): Promise<VerificationRequestItem> {
		return this.reviewVerification(id, { action: 'reject', reason, note });
	},

	/**
	 * Helper: Request additional information/corrections for verification.
	 */
	requestVerificationChanges(id: string, reason: string, note?: string): Promise<VerificationRequestItem> {
		return this.reviewVerification(id, { action: 'request_changes', reason, note });
	}
};
