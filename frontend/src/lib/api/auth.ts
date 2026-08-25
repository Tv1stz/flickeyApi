import { apiRequest } from './client';
import type {
	User,
	RequestOTPRequest,
	RequestOTPResponse,
	VerifyOTPRequest,
	VerifyOTPResponse,
	CompleteProfileRequest,
	DevLoginRequest,
	TokenRefreshResponse
} from '$lib/types/auth';

export const authApi = {
	/**
	 * Request a one-time login/registration SMS code.
	 */
	requestOTP(payload: RequestOTPRequest): Promise<RequestOTPResponse> {
		return apiRequest<RequestOTPResponse>('/auth/request-otp', {
			method: 'POST',
			body: payload,
			skipAuth: true
		});
	},

	/**
	 * Verify an OTP code and obtain an access token. Sets HttpOnly refresh token cookie.
	 */
	verifyOTP(payload: VerifyOTPRequest): Promise<VerifyOTPResponse> {
		return apiRequest<VerifyOTPResponse>('/auth/verify-otp', {
			method: 'POST',
			body: payload,
			skipAuth: true
		});
	},

	/**
	 * Rotate refresh token and obtain new access token.
	 */
	refresh(): Promise<TokenRefreshResponse> {
		return apiRequest<TokenRefreshResponse>('/auth/refresh', {
			method: 'POST',
			skipAuth: true
		});
	},

	/**
	 * Fetch current authenticated user profile.
	 */
	getMe(): Promise<User> {
		return apiRequest<User>('/auth/me', {
			method: 'GET'
		});
	},

	/**
	 * Transition from pending_profile to active status by providing personal details.
	 */
	completeProfile(payload: CompleteProfileRequest): Promise<User> {
		return apiRequest<User>('/auth/complete-profile', {
			method: 'POST',
			body: payload
		});
	},

	/**
	 * Dev only: Instant one-click login for testing (guest/host).
	 */
	devLogin(payload: DevLoginRequest): Promise<VerifyOTPResponse> {
		return apiRequest<VerifyOTPResponse>('/auth/dev-login', {
			method: 'POST',
			body: payload,
			skipAuth: true
		});
	},

	/**
	 * Logout current session and revoke refresh token.
	 */
	logout(): Promise<void> {
		return apiRequest<void>('/auth/logout', {
			method: 'POST'
		});
	},

	/**
	 * Revoke all active sessions across all devices for this user.
	 */
	logoutAll(): Promise<void> {
		return apiRequest<void>('/auth/logout-all', {
			method: 'POST'
		});
	}
};
