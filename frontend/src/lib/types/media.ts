export interface PresignMediaRequest {
	content_type: 'image/jpeg' | 'image/png' | 'image/webp' | 'video/mp4' | 'video/quicktime' | 'video/webm' | 'application/pdf' | string;
	file_size_bytes: number;
}

export interface PresignMediaResponse {
	media_id: string;
	upload_url: string;
	file_key: string;
	expires_in: number;
}

export interface CompleteMediaResponse {
	media_id: string;
	status: 'uploaded' | 'attached';
	file_key: string;
}

export interface UploadedPhoto {
	id?: string;
	mediaId?: string;
	file?: File;
	previewUrl: string;
	status: 'idle' | 'uploading' | 'uploaded' | 'failed';
	progress?: number;
	errorMessage?: string;
}
