export type UserRole = 'guest' | 'host' | 'admin';

export type UserStatus = 'pending_profile' | 'active' | 'suspended' | 'banned';

export interface User {
	id: string;
	phone: string;
	first_name: string | null;
	last_name: string | null;
	email: string | null;
	role: UserRole;
	status: UserStatus;
	created_at: string;
	updated_at: string;
}

export interface RequestOTPRequest {
	phone: string;
	return_url?: string;
}

export interface RequestOTPResponse {
	challenge_id: string;
	expires_in: number;
}

export interface VerifyOTPRequest {
	challenge_id: string;
	code: string;
}

export interface VerifyOTPResponse {
	access_token: string;
	token_type: string;
	is_new_user: boolean;
	return_url?: string;
}

export interface CompleteProfileRequest {
	first_name: string;
	last_name: string;
	email?: string;
}

export interface DevLoginRequest {
	phone: string;
	first_name?: string;
	last_name?: string;
	role?: UserRole;
}

export interface TokenRefreshResponse {
	access_token: string;
	token_type: string;
}
