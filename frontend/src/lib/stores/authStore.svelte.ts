// src/lib/stores/authStore.svelte.ts
import { browser } from '$app/environment';
import { isHostUser, normalizeViewModeForUser } from '$lib/auth/permissions';

export type UserRole = 'guest' | 'host' | 'admin';
export type IsoDateString = string;

export interface HostProfile {
	verificationStatus: 'pending' | 'verified' | 'rejected';
	verifiedAt?: IsoDateString;
	listingsCount: number;
}

export interface User {
	id: string;
	phone: string;
	name: string;
	avatar?: string;
	createdAt: IsoDateString;
	roles: UserRole[];
	hostProfile?: HostProfile;
}

export type ViewMode = 'guest' | 'host';
export type PendingAction = 'favorite' | 'create-listing' | 'contact-host' | null;

export interface PendingContext {
	listingId?: string;
	returnPath?: string;
}

let onLogoutCallback: (() => void) | null = null;

function createAuthStore() {
	let user = $state<User | null>(null);
	let viewMode = $state<ViewMode>('guest');
	let pendingAction = $state<PendingAction>(null);
	let pendingRedirect = $state<string | null>(null);
	let pendingContext = $state<PendingContext | null>(null);
	let initialized = $state(false);

	return {
		onLogout(callback: () => void) {
			onLogoutCallback = callback;
		},

		initialize() {
			if (!browser || initialized) return;

			try {
				const savedUser = localStorage.getItem('user');
				const savedViewMode = localStorage.getItem('viewMode') as ViewMode | null;
				const savedPendingAction = sessionStorage.getItem('pendingAction') as PendingAction;
				const savedPendingRedirect = sessionStorage.getItem('pendingRedirect');
				const savedPendingContext = sessionStorage.getItem('pendingContext');

				let pc: PendingContext | null = null;
				if (savedPendingContext) {
					try {
						pc = JSON.parse(savedPendingContext);
					} catch {}
				}

				if (savedUser) {
					user = JSON.parse(savedUser);
					viewMode = normalizeViewModeForUser(user, savedViewMode);
				}
				pendingAction = savedPendingAction || null;
				pendingRedirect = savedPendingRedirect || null;
				pendingContext = pc;
				initialized = true;
			} catch (e) {
				console.error('Failed to restore auth state:', e);
				initialized = true;
			}
		},

		setUser(newUser: User) {
			const nextViewMode = normalizeViewModeForUser(newUser, viewMode);
			user = newUser;
			viewMode = nextViewMode;

			if (browser) {
				localStorage.setItem('user', JSON.stringify(user));
				localStorage.setItem('viewMode', nextViewMode);
			}
		},

		logout() {
			onLogoutCallback?.();
			user = null;
			viewMode = 'guest';
			pendingAction = null;
			pendingRedirect = null;
			pendingContext = null;

			if (browser) {
				localStorage.removeItem('user');
				localStorage.removeItem('viewMode');
				sessionStorage.removeItem('pendingAction');
				sessionStorage.removeItem('pendingRedirect');
				sessionStorage.removeItem('pendingContext');
			}
		},

		setViewMode(mode: ViewMode) {
            if (!user) return;
			const nextMode = normalizeViewModeForUser(user, mode);
			if (viewMode === nextMode) {
				return;
			}

            viewMode = nextMode;
			if (browser) {
				localStorage.setItem('viewMode', nextMode);
			}
		},

		becomeHost() {
			if (!user) return;

			const updatedUser: User = {
				...user,
				roles: user.roles.includes('host') ? user.roles : [...user.roles, 'host'],
				hostProfile: user.hostProfile || {
					verificationStatus: 'pending',
					listingsCount: 0
				}
			};

            user = updatedUser;
			if (browser) {
				localStorage.setItem('user', JSON.stringify(updatedUser));
			}
		},

		incrementListingsCount() {
			if (!user) return;

			const currentCount = user.hostProfile?.listingsCount || 0;
			const newCount = currentCount + 1;
			const isFirstListing = currentCount === 0;

			const updatedUser: User = {
				...user,
				roles: user.roles.includes('host') ? user.roles : [...user.roles, 'host'],
				hostProfile: {
					...user.hostProfile,
					verificationStatus: user.hostProfile?.verificationStatus || 'pending',
					listingsCount: newCount
				}
			};

            user = updatedUser;
            if (isFirstListing) {
                viewMode = 'host';
            }

			if (browser) {
				localStorage.setItem('user', JSON.stringify(updatedUser));
				if (isFirstListing) {
					localStorage.setItem('viewMode', 'host');
				}
			}
		},

		decrementListingsCount() {
			if (!user?.hostProfile) return;

			const newCount = Math.max(0, user.hostProfile.listingsCount - 1);
			const updatedUser: User = {
				...user,
				hostProfile: {
					...user.hostProfile,
					listingsCount: newCount
				}
			};

            user = updatedUser;
            if (newCount === 0 && viewMode === 'host') {
                viewMode = 'guest';
            }

			if (browser) {
				localStorage.setItem('user', JSON.stringify(updatedUser));
				if (newCount === 0 && localStorage.getItem('viewMode') === 'host') {
					localStorage.setItem('viewMode', 'guest');
				}
			}
		},

		updateProfile(updates: Partial<Pick<User, 'name' | 'avatar'>>) {
			if (!user) return;

			const updatedUser = { ...user, ...updates };
            user = updatedUser;

			if (browser) {
				localStorage.setItem('user', JSON.stringify(updatedUser));
			}
		},

		setPendingAction(action: PendingAction, redirect?: string, context?: PendingContext) {
            pendingAction = action;
            pendingRedirect = redirect || null;
            pendingContext = context || null;

			if (browser) {
				if (action) {
					sessionStorage.setItem('pendingAction', action);
				} else {
					sessionStorage.removeItem('pendingAction');
				}

				if (redirect) {
					sessionStorage.setItem('pendingRedirect', redirect);
				} else {
					sessionStorage.removeItem('pendingRedirect');
				}

				if (context) {
					sessionStorage.setItem('pendingContext', JSON.stringify(context));
				} else {
					sessionStorage.removeItem('pendingContext');
				}
			}
		},

		consumePendingAction(): {
			action: PendingAction;
			redirect: string | null;
			context: PendingContext | null;
		} {
			const action = pendingAction;
			const redirect = pendingRedirect;
			const context = pendingContext;

            pendingAction = null;
            pendingRedirect = null;
            pendingContext = null;

			if (browser) {
				sessionStorage.removeItem('pendingAction');
				sessionStorage.removeItem('pendingRedirect');
				sessionStorage.removeItem('pendingContext');
			}

			return { action, redirect, context };
		},

		clearPendingAction() {
            pendingAction = null;
            pendingRedirect = null;
            pendingContext = null;

			if (browser) {
				sessionStorage.removeItem('pendingAction');
				sessionStorage.removeItem('pendingRedirect');
				sessionStorage.removeItem('pendingContext');
			}
		},

        // ═══════════════════════════════════════════════════════════════════════════
        // REACTIVE GETTERS
        // ═══════════════════════════════════════════════════════════════════════════

		get user() {
			return user;
		},
        get viewMode() {
            return viewMode;
        },
        get pendingAction() {
            return pendingAction;
        },
        get pendingRedirect() {
            return pendingRedirect;
        },
        get pendingContext() {
            return pendingContext;
        },
        get initialized() {
            return initialized;
        },

		get isAuthenticated() {
			return user !== null;
		},
		get currentUser() {
			return user; // Kept for exact compatibility with previous logic
		},
		get userInitials() {
			if (!user?.name) return '';
			const parts = user.name.split(' ').filter(Boolean);
			if (parts.length >= 2) {
				return (parts[0][0] + parts[1][0]).toUpperCase();
			}
			return parts[0]?.substring(0, 2).toUpperCase() || '';
		},
		get isHost() {
			return isHostUser(user);
		},
		get isVerifiedHost() {
			return user?.hostProfile?.verificationStatus === 'verified';
		},
		get isHostMode() {
			return viewMode === 'host';
		},
		get canSwitchModes() {
			const listingsCount = user?.hostProfile?.listingsCount || 0;
			return isHostUser(user) && listingsCount > 0;
		},
		get hostListingsCount() {
			return user?.hostProfile?.listingsCount ?? 0;
		},
		get showCreateButtonInHeader() {
			if (!user) return true;
			const listingsCount = user.hostProfile?.listingsCount || 0;
			return listingsCount === 0;
		}
	};
}

export const authStore = createAuthStore();
