// src/lib/stores/authFlowStore.svelte.ts

import { browser } from '$app/environment';
import { getDefaultCountry, type Country } from '$lib/config/countries';

export type AuthStep =
	| 'phone'
	| 'code'
	| 'password'
	| 'register'
	| 'forgot-password'
	| 'reset-code'
	| 'new-password';

export interface AuthFlowState {
	step: AuthStep;
	phone: string;
	country: Country;
	code: string;
	password: string;
	confirmPassword: string;
	name: string;
	newPassword: string;
	confirmNewPassword: string;
	isNewUser: boolean;
	loading: boolean;
	resendTimer: number;
}

function createAuthFlowStore() {
	// ═══════════════════════════════════════════════════════════════
	// STATE
	// ═══════════════════════════════════════════════════════════════

	let step = $state<AuthStep>('phone');
	let phone = $state('');
	let country = $state<Country>(getDefaultCountry());
	let code = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let name = $state('');
	let newPassword = $state('');
	let confirmNewPassword = $state('');
	let isNewUser = $state(false);
	let loading = $state(false);
	let resendTimer = $state(0);

	// Touched states
	let phoneTouched = $state(false);
	let nameTouched = $state(false);
	let passwordTouched = $state(false);
	let confirmPasswordTouched = $state(false);
	let newPasswordTouched = $state(false);
	let confirmNewPasswordTouched = $state(false);

	let resendInterval: ReturnType<typeof setInterval> | null = null;

	// ═══════════════════════════════════════════════════════════════
	// DERIVED / COMPUTED
	// ═══════════════════════════════════════════════════════════════

	const isResetFlow = $derived(
		step === 'forgot-password' || step === 'reset-code' || step === 'new-password'
	);

	const stepIndex = $derived.by(() => {
		if (isResetFlow) {
			const idx = ['forgot-password', 'reset-code', 'new-password'].indexOf(step);
			return idx >= 0 ? idx : 0;
		}
		const mainSteps = ['phone', 'code', isNewUser ? 'register' : 'password'];
		const idx = mainSteps.indexOf(step);
		return idx >= 0 ? idx : 0;
	});

	const totalSteps = 3;
	const progressPercent = $derived(((stepIndex + 1) / totalSteps) * 100);

	// ═══════════════════════════════════════════════════════════════
	// ACTIONS
	// ═══════════════════════════════════════════════════════════════

	function setStep(newStep: AuthStep) {
		step = newStep;
	}

	function setPhone(value: string, selectedCountry: Country) {
		phone = value;
		country = selectedCountry;
	}

	function setCode(value: string) {
		code = value;
	}

	function setPassword(value: string) {
		password = value;
	}

	function setConfirmPassword(value: string) {
		confirmPassword = value;
	}

	function setName(value: string) {
		name = value;
	}

	function setNewPassword(value: string) {
		newPassword = value;
	}

	function setConfirmNewPassword(value: string) {
		confirmNewPassword = value;
	}

	function setIsNewUser(value: boolean) {
		isNewUser = value;
	}

	function setLoading(value: boolean) {
		loading = value;
	}

	function touchPhone() {
		phoneTouched = true;
	}

	function touchName() {
		nameTouched = true;
	}

	function touchPassword() {
		passwordTouched = true;
	}

	function touchConfirmPassword() {
		confirmPasswordTouched = true;
	}

	function touchNewPassword() {
		newPasswordTouched = true;
	}

	function touchConfirmNewPassword() {
		confirmNewPasswordTouched = true;
	}

	function startResendTimer() {
		if (!browser) return;

		// Очищаем предыдущий интервал
		if (resendInterval) {
			clearInterval(resendInterval);
			resendInterval = null;
		}

		resendTimer = 60;

		resendInterval = setInterval(() => {
			if (resendTimer > 0) {
				resendTimer--;
			}
			if (resendTimer <= 0 && resendInterval) {
				clearInterval(resendInterval);
				resendInterval = null;
			}
		}, 1000);
	}

	function resetCodeState() {
		code = '';
	}

	function resetPasswordState() {
		password = '';
		confirmPassword = '';
		passwordTouched = false;
		confirmPasswordTouched = false;
	}

	function resetNewPasswordState() {
		newPassword = '';
		confirmNewPassword = '';
		newPasswordTouched = false;
		confirmNewPasswordTouched = false;
	}

	function resetAll() {
		// Очищаем таймер перед сбросом
		if (resendInterval) {
			clearInterval(resendInterval);
			resendInterval = null;
		}

		step = 'phone';
		phone = '';
		country = getDefaultCountry();
		code = '';
		password = '';
		confirmPassword = '';
		name = '';
		newPassword = '';
		confirmNewPassword = '';
		isNewUser = false;
		loading = false;
		resendTimer = 0;
		phoneTouched = false;
		nameTouched = false;
		passwordTouched = false;
		confirmPasswordTouched = false;
		newPasswordTouched = false;
		confirmNewPasswordTouched = false;
	}

	function cleanup() {
		if (resendInterval) {
			clearInterval(resendInterval);
			resendInterval = null;
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// RETURN PUBLIC API
	// ═══════════════════════════════════════════════════════════════

	return {
		// State getters (reactive)
		get step() {
			return step;
		},
		get phone() {
			return phone;
		},
		get country() {
			return country;
		},
		get code() {
			return code;
		},
		get password() {
			return password;
		},
		get confirmPassword() {
			return confirmPassword;
		},
		get name() {
			return name;
		},
		get newPassword() {
			return newPassword;
		},
		get confirmNewPassword() {
			return confirmNewPassword;
		},
		get isNewUser() {
			return isNewUser;
		},
		get loading() {
			return loading;
		},
		get resendTimer() {
			return resendTimer;
		},

		// Touched getters
		get phoneTouched() {
			return phoneTouched;
		},
		get nameTouched() {
			return nameTouched;
		},
		get passwordTouched() {
			return passwordTouched;
		},
		get confirmPasswordTouched() {
			return confirmPasswordTouched;
		},
		get newPasswordTouched() {
			return newPasswordTouched;
		},
		get confirmNewPasswordTouched() {
			return confirmNewPasswordTouched;
		},

		// Derived
		get isResetFlow() {
			return isResetFlow;
		},
		get stepIndex() {
			return stepIndex;
		},
		get totalSteps() {
			return totalSteps;
		},
		get progressPercent() {
			return progressPercent;
		},

		// Actions
		setStep,
		setPhone,
		setCode,
		setPassword,
		setConfirmPassword,
		setName,
		setNewPassword,
		setConfirmNewPassword,
		setIsNewUser,
		setLoading,
		touchPhone,
		touchName,
		touchPassword,
		touchConfirmPassword,
		touchNewPassword,
		touchConfirmNewPassword,
		startResendTimer,
		resetCodeState,
		resetPasswordState,
		resetNewPasswordState,
		resetAll,
		cleanup
	};
}

export const authFlow = createAuthFlowStore();
