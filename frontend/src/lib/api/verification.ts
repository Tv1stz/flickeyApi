// src/lib/api/verification.ts
import { apiRequest } from './client';
import type {
	SubmitVerificationPayload,
	VerificationResponse
} from '$lib/types/verification';

export const verificationApi = {
	/**
	 * Retrieve the current host's verification request status and requisites.
	 */
	getMyVerification(): Promise<VerificationResponse> {
		return apiRequest<VerificationResponse>('/verification/my', {
			method: 'GET'
		});
	},

	/**
	 * Submit or update partner verification requisites and document scans.
	 */
	submitVerification(payload: SubmitVerificationPayload): Promise<VerificationResponse> {
		return apiRequest<VerificationResponse>('/verification', {
			method: 'POST',
			body: payload
		});
	}
};
