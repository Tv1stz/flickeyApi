import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { ApiError } from '$lib/types/api';

/**
 * Base API URL derived from SvelteKit native public environment configuration.
 */
export const API_BASE = PUBLIC_API_BASE_URL ? `${PUBLIC_API_BASE_URL.replace(/\/$/, '')}/api/v1` : '/api/v1';

let currentAccessToken: string | null = null;
let refreshPromise: Promise<string | null> | null = null;

export function setAccessToken(token: string | null): void {
	currentAccessToken = token;
	if (typeof localStorage !== 'undefined') {
		if (token) {
			localStorage.setItem('flickey_access_token', token);
		} else {
			localStorage.removeItem('flickey_access_token');
		}
	}
}

export function getAccessToken(): string | null {
	if (!currentAccessToken && typeof localStorage !== 'undefined') {
		currentAccessToken = localStorage.getItem('flickey_access_token');
	}
	return currentAccessToken;
}

export interface RequestOptions extends Omit<RequestInit, 'body'> {
	body?: unknown;
	skipAuth?: boolean;
	headers?: Record<string, string>;
}

/**
 * Centralized, typed API request function with automatic 401 token refresh queue.
 */
export async function apiRequest<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
	const url = endpoint.startsWith('http') ? endpoint : `${API_BASE}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;

	const headers: Record<string, string> = {
		...options.headers
	};

	if (options.body && !(options.body instanceof FormData)) {
		headers['Content-Type'] = 'application/json';
	}

	const token = getAccessToken();
	if (!options.skipAuth && token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const fetchOptions: RequestInit = {
		method: options.method || 'GET',
		headers,
		credentials: 'include' // Always include HttpOnly cookies for refresh token rotation
	};

	if (options.body !== undefined) {
		fetchOptions.body = options.body instanceof FormData ? options.body : JSON.stringify(options.body);
	}

	let response: Response;
	try {
		response = await fetch(url, fetchOptions);
	} catch (err) {
		throw new ApiError('NETWORK_ERROR', err instanceof Error ? err.message : 'Network request failed', 0);
	}

	// 401 Handling: Attempt automatic token refresh once (if not already refreshing or calling auth endpoints)
	if (response.status === 401 && !options.skipAuth && !endpoint.includes('/auth/refresh') && !endpoint.includes('/auth/verify-otp') && !endpoint.includes('/auth/dev-login')) {
		const newToken = await attemptTokenRefresh();
		if (newToken) {
			headers['Authorization'] = `Bearer ${newToken}`;
			fetchOptions.headers = headers;
			try {
				response = await fetch(url, fetchOptions);
			} catch (err) {
				throw new ApiError('NETWORK_ERROR', err instanceof Error ? err.message : 'Network request failed on retry', 0);
			}
		}
	}

	// 204 No Content
	if (response.status === 204) {
		return null as T;
	}

	const contentType = response.headers.get('content-type') || '';
	const isJson = contentType.includes('application/json');

	let data: unknown;
	if (isJson) {
		try {
			data = await response.json();
		} catch {
			data = null;
		}
	} else {
		data = await response.text();
	}

	if (!response.ok) {
		if (response.status === 401) {
			setAccessToken(null);
		}
		const errObj = typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : {};
		const code = (errObj.code as string) || `HTTP_${response.status}`;
		const message = (errObj.message as string) || (typeof data === 'string' && data ? data : response.statusText || 'An error occurred');
		throw new ApiError(code, message, response.status, errObj.details as Record<string, unknown> | undefined);
	}

	return data as T;
}

/**
 * Single-flight refresh token lock to prevent concurrent duplicate refresh requests.
 */
async function attemptTokenRefresh(): Promise<string | null> {
	if (refreshPromise) {
		return refreshPromise;
	}

	refreshPromise = (async () => {
		try {
			const res = await fetch(`${API_BASE}/auth/refresh`, {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' }
			});

			if (!res.ok) {
				setAccessToken(null);
				return null;
			}

			const data = (await res.json()) as { access_token: string };
			if (data && data.access_token) {
				setAccessToken(data.access_token);
				return data.access_token;
			}

			setAccessToken(null);
			return null;
		} catch {
			setAccessToken(null);
			return null;
		} finally {
			refreshPromise = null;
		}
	})();

	return refreshPromise;
}
