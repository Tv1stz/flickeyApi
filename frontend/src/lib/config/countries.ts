// src/lib/config/countries.ts

export interface Country {
	code: string;
	name: string;
	dialCode: string;
	flag: string;
	format: string;
	maxLength: number;
}

export const countries: Country[] = [
	{
		code: 'BY',
		name: 'Беларусь',
		dialCode: '+375',
		flag: '🇧🇾',
		format: '## ###-##-##',
		maxLength: 9
	},
	{
		code: 'RU',
		name: 'Россия',
		dialCode: '+7',
		flag: '🇷🇺',
		format: '### ###-##-##',
		maxLength: 10
	},
	{
		code: 'UA',
		name: 'Украина',
		dialCode: '+380',
		flag: '🇺🇦',
		format: '## ###-##-##',
		maxLength: 9
	},
	{
		code: 'KZ',
		name: 'Казахстан',
		dialCode: '+7',
		flag: '🇰🇿',
		format: '### ###-##-##',
		maxLength: 10
	},
	{
		code: 'PL',
		name: 'Польша',
		dialCode: '+48',
		flag: '🇵🇱',
		format: '### ### ###',
		maxLength: 9
	},
	{
		code: 'LT',
		name: 'Литва',
		dialCode: '+370',
		flag: '🇱🇹',
		format: '### #####',
		maxLength: 8
	},
	{
		code: 'LV',
		name: 'Латвия',
		dialCode: '+371',
		flag: '🇱🇻',
		format: '## ### ###',
		maxLength: 8
	},
	{
		code: 'EE',
		name: 'Эстония',
		dialCode: '+372',
		flag: '🇪🇪',
		format: '#### ####',
		maxLength: 8
	},
	{
		code: 'GE',
		name: 'Грузия',
		dialCode: '+995',
		flag: '🇬🇪',
		format: '### ### ###',
		maxLength: 9
	},
	{
		code: 'AM',
		name: 'Армения',
		dialCode: '+374',
		flag: '🇦🇲',
		format: '## ### ###',
		maxLength: 8
	},
	{
		code: 'AZ',
		name: 'Азербайджан',
		dialCode: '+994',
		flag: '🇦🇿',
		format: '## ### ## ##',
		maxLength: 9
	},
	{
		code: 'MD',
		name: 'Молдова',
		dialCode: '+373',
		flag: '🇲🇩',
		format: '### ## ###',
		maxLength: 8
	},
	{
		code: 'UZ',
		name: 'Узбекистан',
		dialCode: '+998',
		flag: '🇺🇿',
		format: '## ### ## ##',
		maxLength: 9
	},
	{
		code: 'KG',
		name: 'Кыргызстан',
		dialCode: '+996',
		flag: '🇰🇬',
		format: '### ### ###',
		maxLength: 9
	},
	{
		code: 'TJ',
		name: 'Таджикистан',
		dialCode: '+992',
		flag: '🇹🇯',
		format: '## ### ####',
		maxLength: 9
	},
	{
		code: 'TM',
		name: 'Туркменистан',
		dialCode: '+993',
		flag: '🇹🇲',
		format: '## ######',
		maxLength: 8
	},
	{
		code: 'DE',
		name: 'Германия',
		dialCode: '+49',
		flag: '🇩🇪',
		format: '### #######',
		maxLength: 10
	},
	{
		code: 'US',
		name: 'США',
		dialCode: '+1',
		flag: '🇺🇸',
		format: '### ###-####',
		maxLength: 10
	},
	{
		code: 'GB',
		name: 'Великобритания',
		dialCode: '+44',
		flag: '🇬🇧',
		format: '#### ######',
		maxLength: 10
	},
	{
		code: 'IL',
		name: 'Израиль',
		dialCode: '+972',
		flag: '🇮🇱',
		format: '## ### ####',
		maxLength: 9
	},
	{
		code: 'TR',
		name: 'Турция',
		dialCode: '+90',
		flag: '🇹🇷',
		format: '### ### ## ##',
		maxLength: 10
	},
	{
		code: 'AE',
		name: 'ОАЭ',
		dialCode: '+971',
		flag: '🇦🇪',
		format: '## ### ####',
		maxLength: 9
	},
	{
		code: 'CZ',
		name: 'Чехия',
		dialCode: '+420',
		flag: '🇨🇿',
		format: '### ### ###',
		maxLength: 9
	},
	{
		code: 'RS',
		name: 'Сербия',
		dialCode: '+381',
		flag: '🇷🇸',
		format: '## ### ####',
		maxLength: 9
	},
	{
		code: 'ME',
		name: 'Черногория',
		dialCode: '+382',
		flag: '🇲🇪',
		format: '## ### ###',
		maxLength: 8
	}
];

// Популярные страны для показа вверху списка
export const popularCountries = ['BY', 'RU', 'UA', 'KZ', 'PL'];

// Страна по умолчанию
export function getDefaultCountry(): Country {
	return countries.find((c) => c.code === 'BY') ?? countries[0];
}

// Поиск страны по коду
export function getCountryByCode(code: string): Country | undefined {
	return countries.find((c) => c.code === code);
}

// Полный номер телефона
export function getFullPhoneNumber(phone: string, country: Country): string {
	return `${country.dialCode}${phone}`;
}

// Валидация длины номера
export function validatePhoneLength(phone: string, country: Country): boolean {
	return phone.length === country.maxLength;
}

// ═══════════════════════════════════════════════════════════════
// ПАРСИНГ НОМЕРА С ОПРЕДЕЛЕНИЕМ СТРАНЫ
// ═══════════════════════════════════════════════════════════════

export interface ParsedPhone {
	country: Country;
	localNumber: string;
	isAutoDetected: boolean;
}

/**
 * Парсит вставленный номер и определяет страну
 * Приоритет: более длинные коды (напр. +375 перед +37)
 */
export function parsePhoneNumber(rawInput: string, currentCountry: Country): ParsedPhone {
	// Очищаем от всего кроме цифр и +
	let cleaned = rawInput.replace(/[^\d+]/g, '');

	// Если начинается с 00, заменяем на +
	if (cleaned.startsWith('00')) {
		cleaned = '+' + cleaned.slice(2);
	}

	// Если начинается с 8 и длина > 10, скорее всего это +7
	if (cleaned.startsWith('8') && cleaned.length >= 11) {
		cleaned = '+7' + cleaned.slice(1);
	}

	// Если нет + в начале, это локальный номер
	if (!cleaned.startsWith('+')) {
		const digits = cleaned.replace(/\D/g, '');
		return {
			country: currentCountry,
			localNumber: digits.slice(0, currentCountry.maxLength),
			isAutoDetected: false
		};
	}

	// Сортируем страны по длине dialCode (от большего к меньшему)
	const sortedCountries = [...countries].sort((a, b) => b.dialCode.length - a.dialCode.length);

	// Ищем совпадение по коду страны
	for (const country of sortedCountries) {
		const dialCodeDigits = country.dialCode.replace(/\D/g, '');
		const inputDigits = cleaned.replace(/\D/g, '');

		if (inputDigits.startsWith(dialCodeDigits)) {
			const localNumber = inputDigits.slice(dialCodeDigits.length);
			return {
				country,
				localNumber: localNumber.slice(0, country.maxLength),
				isAutoDetected: true
			};
		}
	}

	// Если страна не определена, используем текущую
	const digits = cleaned.replace(/\D/g, '');
	return {
		country: currentCountry,
		localNumber: digits.slice(0, currentCountry.maxLength),
		isAutoDetected: false
	};
}

/**
 * Проверяет, содержит ли строка достаточно цифр для определения страны
 */
export function hasEnoughDigitsForCountryDetection(input: string): boolean {
	const digits = input.replace(/\D/g, '');
	// Минимум код страны + 1 цифра (обычно 2-4 цифры кода)
	return digits.length >= 3;
}
