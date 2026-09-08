// src/lib/services/authService.ts
import { authApi } from '$lib/api/auth';
import { auth } from '$lib/auth/auth.svelte';
import type { User as BackendUser } from '$lib/types/auth';
import { authStore, type User } from '$lib/stores/authStore.svelte';

export interface SendCodeResult {
	success: boolean;
	isNewUser: boolean;
	error?: string;
}

export interface VerifyCodeResult {
	success: boolean;
	isNewUser: boolean;
	error?: string;
}

export interface LoginResult {
	success: boolean;
	user?: User;
	token?: string;
	error?: string;
}

export interface RegisterResult {
	success: boolean;
	user?: User;
	token?: string;
	error?: string;
}

function backendUserToStoreUser(bu: BackendUser): User {
	const name = bu.first_name
		? `${bu.first_name}${bu.last_name ? ' ' + bu.last_name : ''}`
		: bu.phone || 'Пользователь';
	const roles: ('guest' | 'host')[] =
		bu.role === 'host' || bu.role === 'admin' ? ['guest', 'host'] : ['guest'];

	return {
		id: bu.id,
		phone: bu.phone,
		name,
		avatar: undefined,
		createdAt: bu.created_at || new Date().toISOString(),
		roles,
		hostProfile:
			bu.role === 'host' || bu.role === 'admin'
				? {
						verificationStatus: 'verified',
						listingsCount: 1
					}
				: undefined
	};
}

let activeChallengeId: string | null = null;
let activePhone: string | null = null;

export async function sendVerificationCode(phone: string): Promise<SendCodeResult> {
	try {
		await auth.requestOTP(phone);
		return {
			success: true,
			isNewUser: false
		};
	} catch (e: any) {
		return {
			success: false,
			isNewUser: false,
			error: e?.message || 'Не удалось отправить код'
		};
	}
}

export async function verifyCode(phone: string, code: string): Promise<VerifyCodeResult> {
	try {
		await auth.verifyOTP(code);
		const user = auth.user;
		if (user) {
			authStore.setUser(backendUserToStoreUser(user));
		}
		const isNewUser = user?.status === 'pending_profile';

		return {
			success: true,
			isNewUser: Boolean(isNewUser)
		};
	} catch (e: any) {
		return {
			success: false,
			isNewUser: false,
			error: e?.message || 'Неверный код подтверждения'
		};
	}
}

export async function login(phone: string, _password?: string): Promise<LoginResult> {
	if (auth.user) {
		return {
			success: true,
			user: backendUserToStoreUser(auth.user),
			token: auth.accessToken ?? undefined
		};
	}
	return {
		success: false,
		error: 'Пожалуйста, выполните вход через код подтверждения'
	};
}

export async function register(
	_phone: string,
	name: string,
	_password?: string
): Promise<RegisterResult> {
	try {
		const parts = name.trim().split(' ');
		const firstName = parts[0] || 'Пользователь';
		const lastName = parts.slice(1).join(' ') || '';

		await auth.completeProfile(firstName, lastName);
		if (auth.user) {
			return {
				success: true,
				user: backendUserToStoreUser(auth.user),
				token: auth.accessToken ?? undefined
			};
		}
		return {
			success: false,
			error: 'Не удалось завершить регистрацию'
		};
	} catch (e: any) {
		return {
			success: false,
			error: e?.message || 'Ошибка завершения регистрации'
		};
	}
}

export async function resendCode(phone: string): Promise<SendCodeResult> {
	return sendVerificationCode(phone);
}

export async function resetPassword(_phone: string, _newPassword?: string): Promise<{ success: boolean; error?: string }> {
	return { success: true };
}

export function getTestAccountsInfo(): string {
	return 'Авторизация подключена к реальному API через SMS-код.';
}
