// src/lib/utils/contactFormatters.ts
import type { ContactMethodType } from '$lib/components/card/types';

interface ContactConfig {
	allowedChars: RegExp;
	maxLength: number;
	inputMode: 'email' | 'tel' | 'text';
	placeholder: string;
	hint: string;
}

const CONTACT_CONFIGS: Record<ContactMethodType, ContactConfig> = {
	email: {
		allowedChars: /[a-zA-Z0-9@.\-_+]/,
		maxLength: 254,
		inputMode: 'email',
		placeholder: 'example@mail.com',
		hint: 'Электронная почта'
	},
	telegram: {
		allowedChars: /[a-zA-Z0-9_@+\d]/,
		maxLength: 33,
		inputMode: 'text',
		placeholder: '@username',
		hint: '@username или номер телефона'
	},
	whatsapp: {
		allowedChars: /[\d+\s]/,
		maxLength: 20,
		inputMode: 'tel',
		placeholder: '+375 29 123 45 67',
		hint: 'Номер с кодом страны'
	},
	viber: {
		allowedChars: /[\d+\s]/,
		maxLength: 20,
		inputMode: 'tel',
		placeholder: '+375 29 123 45 67',
		hint: 'Номер с кодом страны'
	},
	signal: {
		allowedChars: /[\d+\s]/,
		maxLength: 20,
		inputMode: 'tel',
		placeholder: '+375 29 123 45 67',
		hint: 'Номер с кодом страны'
	}
};

// ═══════════════════════════════════════════════════════════════════
// CONFIG GETTERS
// ═══════════════════════════════════════════════════════════════════

export function getContactPlaceholder(type: ContactMethodType): string {
	return CONTACT_CONFIGS[type].placeholder;
}

export function getContactHint(type: ContactMethodType): string {
	return CONTACT_CONFIGS[type].hint;
}

export function getContactInputMode(type: ContactMethodType): 'email' | 'tel' | 'text' {
	return CONTACT_CONFIGS[type].inputMode;
}

export function getContactMaxLength(type: ContactMethodType): number {
	return CONTACT_CONFIGS[type].maxLength;
}

// ═══════════════════════════════════════════════════════════════════
// PHONE UTILITIES (оптимизированные)
// ═══════════════════════════════════════════════════════════════════

function filterPhone(value: string): string {
	let result = value.replace(/[^\d+]/g, '');
	const hasPlus = result.includes('+');
	result = result.replace(/\+/g, '');

	if (hasPlus) result = '+' + result;

	const digits = result.replace(/\D/g, '');
	if (digits.length > 15) {
		result = result.startsWith('+') ? '+' + digits.slice(0, 15) : digits.slice(0, 15);
	}

	return result;
}

// Паттерны форматирования для разных стран
const PHONE_PATTERNS: Record<string, number[]> = {
	'375': [3, 2, 3, 2, 2], // Belarus
	'7': [1, 3, 3, 2, 2], // Russia/Kazakhstan
	'380': [3, 2, 3, 2, 2], // Ukraine
	'48': [2, 3, 3, 3], // Poland
	'370': [3, 3, 5], // Lithuania
	'371': [3, 2, 3, 3], // Latvia
	'372': [3, 4, 4], // Estonia
	default: [3, 3, 3, 3, 3]
};

function formatWithPattern(digits: string, groups: number[]): string {
	const parts: string[] = [];
	let index = 0;

	for (const groupSize of groups) {
		if (index >= digits.length) break;
		parts.push(digits.slice(index, index + groupSize));
		index += groupSize;
	}

	if (index < digits.length) parts.push(digits.slice(index));
	return parts.join(' ').trim();
}

function formatPhone(value: string): string {
	const cleaned = value.replace(/[^\d+]/g, '');
	if (!cleaned) return '';

	const hasPlus = cleaned.startsWith('+');
	const digits = cleaned.replace(/\D/g, '');

	if (!digits) return hasPlus ? '+' : '';

	// Найти подходящий паттерн
	const pattern =
		Object.entries(PHONE_PATTERNS).find(
			([prefix]) => prefix !== 'default' && digits.startsWith(prefix)
		)?.[1] ?? PHONE_PATTERNS.default;

	const formatted = formatWithPattern(digits, pattern);
	return hasPlus ? '+' + formatted : formatted;
}

// ═══════════════════════════════════════════════════════════════════
// EMAIL UTILITIES
// ═══════════════════════════════════════════════════════════════════

function filterEmail(value: string): string {
	let result = value.toLowerCase().replace(/\s/g, '');
	result = result.replace(/[^a-z0-9@.\-_+]/g, '');

	const atIndex = result.indexOf('@');
	if (atIndex !== -1) {
		const before = result.slice(0, atIndex + 1);
		const after = result.slice(atIndex + 1).replace(/@/g, '');
		result = before + after;
	}

	return result;
}

// ═══════════════════════════════════════════════════════════════════
// TELEGRAM UTILITIES
// ═══════════════════════════════════════════════════════════════════

function filterTelegram(value: string): string {
	const trimmed = value.trim();
	const isUsername = trimmed.startsWith('@') || /^[a-zA-Z]/.test(trimmed);

	if (isUsername) {
		let result = trimmed.replace(/[^a-zA-Z0-9_@]/g, '');
		const hasAt = result.includes('@');
		result = result.replace(/@/g, '');

		if (hasAt || /^[a-zA-Z]/.test(result)) {
			result = '@' + result;
		}

		if (result.length > 33) {
			result = result.slice(0, 33);
		}

		return result === '@' ? '' : result;
	} else {
		return filterPhone(trimmed);
	}
}

// ═══════════════════════════════════════════════════════════════════
// MAIN FUNCTIONS
// ═══════════════════════════════════════════════════════════════════

export function filterContactInput(type: ContactMethodType, value: string): string {
	switch (type) {
		case 'email':
			return filterEmail(value);
		case 'telegram':
			return filterTelegram(value);
		case 'whatsapp':
		case 'viber':
		case 'signal':
			return filterPhone(value);
		default:
			return value;
	}
}

function formatTelegram(value: string): string {
	const trimmed = value.trim();

	if (/^[\d+]/.test(trimmed) && !trimmed.startsWith('@')) {
		return formatPhone(trimmed);
	}

	return trimmed;
}

export function formatContactDisplay(type: ContactMethodType, value: string): string {
	switch (type) {
		case 'email':
			return value.toLowerCase().trim();
		case 'telegram':
			return formatTelegram(value);
		case 'whatsapp':
		case 'viber':
		case 'signal':
			return formatPhone(value);
		default:
			return value;
	}
}

// ═══════════════════════════════════════════════════════════════════
// VALIDATION
// ═══════════════════════════════════════════════════════════════════

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

export function validateContactValue(type: ContactMethodType, value: string): string {
	const trimmed = value.trim();

	if (!trimmed) return ''; // Optional field

	switch (type) {
		case 'email': {
			if (!trimmed.includes('@')) return 'Добавьте @';

			const [local, domain] = trimmed.split('@');
			if (!local) return 'Введите имя до @';
			if (!domain) return 'Введите домен после @';
			if (!domain.includes('.')) return 'Домен должен содержать точку';

			const domainParts = domain.split('.');
			const tld = domainParts[domainParts.length - 1];
			if (tld.length < 2) return 'Некорректный домен';

			if (!EMAIL_REGEX.test(trimmed)) return 'Некорректный email';
			return '';
		}

		case 'telegram': {
			if (trimmed.startsWith('@')) {
				const username = trimmed.slice(1);
				if (username.length < 5) return 'Минимум 5 символов';
				if (username.length > 32) return 'Максимум 32 символа';
				if (!/^[a-zA-Z]/.test(username)) return 'Начните с буквы';
				if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(username)) return 'Только буквы, цифры и _';
				return '';
			}

			const digits = trimmed.replace(/\D/g, '');
			if (digits.length < 9) return 'Введите @username или номер';
			if (digits.length > 15) return 'Номер слишком длинный';
			return '';
		}

		case 'whatsapp':
		case 'viber':
		case 'signal': {
			const digits = trimmed.replace(/\D/g, '');
			if (digits.length < 9) return 'Минимум 9 цифр';
			if (digits.length > 15) return 'Максимум 15 цифр';
			if (!trimmed.startsWith('+')) return 'Начните с +';
			return '';
		}

		default:
			return '';
	}
}

export function isKeyAllowed(
	type: ContactMethodType,
	key: string,
	currentValue: string,
	cursorPosition: number
): boolean {
	const controlKeys = [
		'Backspace',
		'Delete',
		'Tab',
		'Escape',
		'Enter',
		'ArrowLeft',
		'ArrowRight',
		'ArrowUp',
		'ArrowDown',
		'Home',
		'End'
	];

	if (controlKeys.includes(key)) return true;
	if (key.length !== 1) return true;

	switch (type) {
		case 'email':
			return /^[a-zA-Z0-9@.\-_+]$/.test(key);

		case 'telegram': {
			const isUsernameMode = currentValue.includes('@') || /^[a-zA-Z]/.test(currentValue);

			if (isUsernameMode) {
				if (key === '@') return cursorPosition === 0 && !currentValue.includes('@');
				return /^[a-zA-Z0-9_]$/.test(key);
			} else {
				if (key === '@') return cursorPosition === 0 && currentValue.length === 0;
				if (/^[a-zA-Z]$/.test(key)) return currentValue.length === 0;
				if (key === '+') return cursorPosition === 0 && !currentValue.includes('+');
				return /^\d$/.test(key);
			}
		}

		case 'whatsapp':
		case 'viber':
		case 'signal': {
			if (key === '+') return cursorPosition === 0 && !currentValue.includes('+');
			return /^\d$/.test(key);
		}

		default:
			return true;
	}
}
