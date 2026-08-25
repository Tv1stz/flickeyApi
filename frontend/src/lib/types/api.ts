export interface ApiErrorResponse {
	code: string;
	message: string;
	details?: Record<string, unknown>;
}

export class ApiError extends Error {
	constructor(
		public code: string,
		message: string,
		public status: number,
		public details?: Record<string, unknown>
	) {
		super(message);
		this.name = 'ApiError';
	}
}

export type Result<T> =
	| { ok: true; data: T; status: number }
	| { ok: false; error: ApiError; status: number };
