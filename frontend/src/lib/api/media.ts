import { apiRequest } from './client';
import type {
	PresignMediaRequest,
	PresignMediaResponse,
	CompleteMediaResponse
} from '$lib/types/media';

export const mediaApi = {
	/**
	 * Step 1: Request S3 / Selectel presigned PUT upload URL.
	 */
	presign(payload: PresignMediaRequest): Promise<PresignMediaResponse> {
		return apiRequest<PresignMediaResponse>('/media/presign', {
			method: 'POST',
			body: payload
		});
	},

	/**
	 * Step 2: Upload file directly to S3/Storage via presigned URL.
	 */
	async uploadDirect(uploadUrl: string, file: File | Blob, contentType: string): Promise<void> {
		const res = await fetch(uploadUrl, {
			method: 'PUT',
			headers: {
				'Content-Type': contentType
			},
			body: file
		});

		if (!res.ok) {
			throw new Error(`Direct storage upload failed with status ${res.status}: ${res.statusText}`);
		}
	},

	/**
	 * Step 3: Complete upload and trigger server-side magic byte inspection.
	 */
	complete(mediaId: string): Promise<CompleteMediaResponse> {
		return apiRequest<CompleteMediaResponse>(`/media/${mediaId}/complete`, {
			method: 'POST'
		});
	}
};
