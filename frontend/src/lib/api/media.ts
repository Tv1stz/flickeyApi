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
	},

	/**
	 * Helper: Upload a file (image, video, document) through presign -> direct upload -> complete.
	 */
	async uploadFile(file: File): Promise<{ mediaId: string; fileKey: string }> {
		const contentType = file.type || 'application/octet-stream';
		const presign = await this.presign({
			content_type: contentType,
			file_size_bytes: file.size
		});
		await this.uploadDirect(presign.upload_url, file, contentType);
		const complete = await this.complete(presign.media_id);
		return {
			mediaId: complete.media_id,
			fileKey: complete.file_key
		};
	},

	/**
	 * Resolve full public URL for a media file key.
	 */
	getMediaUrl(fileKey?: string): string {
		if (!fileKey) return '';
		if (fileKey.startsWith('http://') || fileKey.startsWith('https://')) return fileKey;
		return `http://localhost:8000/api/v1/media/dev-upload/${fileKey.replace(/^\/+/, '')}`;
	}
};

