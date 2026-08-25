import { goto } from '$app/navigation';
import { authApi } from '$lib/api/auth';
import { setAccessToken, getAccessToken } from '$lib/api/client';
import { notifications } from '$lib/stores/notifications.svelte';
import type { User, UserRole } from '$lib/types/auth';

class AuthState {
	// Svelte 5 Runes for reactive state
	user = $state<User | null>(null);
	accessToken = $state<string | null>(null);
	activeContext = $state<'guest' | 'host'>('guest');
	isLoading = $state<boolean>(true);
	isInitialized = $state<boolean>(false);

	// Modal / Flow state
	isAuthModalOpen = $state<boolean>(false);
	authModalView = $state<'phone' | 'otp' | 'complete_profile'>('phone');
	pendingChallengeId = $state<string | null>(null);
	pendingPhone = $state<string>('');

	// Svelte 5 Derived Runes
	isAuthenticated = $derived(!!this.user);
	isHost = $derived(this.user?.role === 'host' || this.user?.role === 'admin');
	isAdmin = $derived(this.user?.role === 'admin');
	isProfilePending = $derived(this.user?.status === 'pending_profile');
	displayName = $derived(
		this.user?.first_name
			? `${this.user.first_name}${this.user.last_name ? ' ' + this.user.last_name : ''}`
			: this.user?.phone || 'Пользователь'
	);

	/**
	 * Initialize authentication session from backend (via stored token or HttpOnly refresh cookie).
	 */
	async init(): Promise<void> {
		if (typeof window === 'undefined') return;

		this.isLoading = true;
		try {
			// 1. Try restoring session from persisted access token
			const token = getAccessToken();
			if (token) {
				this.accessToken = token;
				const user = await authApi.getMe().catch(() => null);
				if (user) {
					this.user = user;
					this.syncActiveContext();
					notifications.connect(token);
					return;
				} else {
					this.clearSession();
				}
			}

			// 2. Otherwise try refreshing access token via cookie
			const refreshRes = await authApi.refresh().catch(() => null);
			if (refreshRes?.access_token) {
				this.setSession(refreshRes.access_token);
				const user = await authApi.getMe().catch(() => null);
				if (user) {
					this.user = user;
					this.syncActiveContext();
				} else {
					this.clearSession();
				}
			} else {
				this.clearSession();
			}
		} catch {
			this.clearSession();
		} finally {
			this.isLoading = false;
			this.isInitialized = true;
		}
	}

	/**
	 * Request OTP code via SMS.
	 */
	async requestOTP(phone: string): Promise<string> {
		const res = await authApi.requestOTP({ phone });
		this.pendingChallengeId = res.challenge_id;
		this.pendingPhone = phone;
		this.authModalView = 'otp';
		return res.challenge_id;
	}

	/**
	 * Verify OTP and authenticate.
	 */
	async verifyOTP(code: string): Promise<void> {
		if (!this.pendingChallengeId) {
			throw new Error('No pending challenge ID. Please request OTP again.');
		}

		const res = await authApi.verifyOTP({
			challenge_id: this.pendingChallengeId,
			code
		});

		this.setSession(res.access_token);
		const user = await authApi.getMe();
		this.user = user;
		this.syncActiveContext();

		if (user.status === 'pending_profile') {
			this.authModalView = 'complete_profile';
		} else {
			this.closeAuthModal();
		}
	}

	/**
	 * Quick login for development / testing.
	 */
	async quickDevLogin(phone: string = '+375291112233', role: UserRole = 'guest'): Promise<void> {
		const res = await authApi.devLogin({
			phone,
			first_name: role === 'host' ? 'Super' : 'Test',
			last_name: role === 'host' ? 'Host' : 'Guest',
			role
		});

		this.setSession(res.access_token);
		const user = await authApi.getMe();
		this.user = user;
		this.activeContext = role === 'host' ? 'host' : 'guest';
		this.closeAuthModal();
	}

	/**
	 * Complete profile information.
	 */
	async completeProfile(firstName: string, lastName: string, email?: string): Promise<void> {
		const updatedUser = await authApi.completeProfile({
			first_name: firstName,
			last_name: lastName,
			email: email || undefined
		});

		this.user = updatedUser;
		this.syncActiveContext();
		this.closeAuthModal();
	}

	/**
	 * Switch active context between Guest and Host mode.
	 */
	switchContext(context: 'guest' | 'host'): void {
		if (context === 'host' && !this.isHost) {
			return; // Cannot switch to host if not a host
		}
		this.activeContext = context;
	}

	/**
	 * Synchronize active context with server-authoritative role.
	 */
	private syncActiveContext(): void {
		if (this.isHost && this.activeContext !== 'host') {
			// Keep user in current context if valid, otherwise adjust
		} else if (!this.isHost) {
			this.activeContext = 'guest';
		}
	}

	/**
	 * Refresh user profile from backend (e.g. after submitting a listing to detect auto-promotion).
	 */
	async refreshUser(): Promise<void> {
		if (!this.accessToken) return;
		try {
			const user = await authApi.getMe();
			this.user = user;
			this.syncActiveContext();
		} catch {
			// Ignored if unauthenticated
		}
	}

	/**
	 * Logout current session.
	 */
	async logout(): Promise<void> {
		try {
			await authApi.logout();
		} catch {
			// Always clean local session on logout
		} finally {
			this.clearSession();
			if (typeof window !== 'undefined') {
				const path = window.location.pathname;
				if (path.startsWith('/host') || path.startsWith('/profile') || path.startsWith('/admin')) {
					goto('/');
				}
			}
		}
	}

	openAuthModal(view: 'phone' | 'otp' | 'complete_profile' = 'phone'): void {
		this.authModalView = view;
		this.isAuthModalOpen = true;
	}

	closeAuthModal(): void {
		this.isAuthModalOpen = false;
		this.pendingChallengeId = null;
		this.pendingPhone = '';
		this.authModalView = 'phone';
	}

	private setSession(token: string): void {
		this.accessToken = token;
		setAccessToken(token);
		notifications.connect(token);
	}

	private clearSession(): void {
		this.user = null;
		this.accessToken = null;
		this.activeContext = 'guest';
		setAccessToken(null);
		notifications.disconnect();
	}
}

export const auth = new AuthState();
