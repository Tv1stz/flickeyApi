// src/lib/utils/pluralize.ts

/**
 * Склонение слов в зависимости от числа
 * @param n - число
 * @param forms - массив из 3 форм: [для 1, для 2-4, для 5-20]
 * @example pluralize(5, ['гость', 'гостя', 'гостей']) // "гостей"
 */
export function pluralize(n: number, forms: [string, string, string]): string {
	const abs = Math.abs(n);
	const lastTwo = abs % 100;
	const lastOne = abs % 10;

	if (lastTwo >= 11 && lastTwo <= 19) {
		return forms[2];
	}

	if (lastOne === 1) {
		return forms[0];
	}

	if (lastOne >= 2 && lastOne <= 4) {
		return forms[1];
	}

	return forms[2];
}

/**
 * Число + слово в правильной форме
 * @example withPlural(5, ['гость', 'гостя', 'гостей']) // "5 гостей"
 */
export function withPlural(n: number, forms: [string, string, string]): string {
	return `${n} ${pluralize(n, forms)}`;
}

// Готовые словари для частых случаев
export const PLURAL_FORMS = {
	guests: ['гость', 'гостя', 'гостей'] as [string, string, string],
	bedrooms: ['спальня', 'спальни', 'спален'] as [string, string, string],
	beds: ['кровать', 'кровати', 'кроватей'] as [string, string, string],
	bathrooms: ['ванная', 'ванные', 'ванных'] as [string, string, string],
	rooms: ['комната', 'комнаты', 'комнат'] as [string, string, string],
	nights: ['ночь', 'ночи', 'ночей'] as [string, string, string],
	floors: ['этаж', 'этажа', 'этажей'] as [string, string, string],
	sqm: ['м²', 'м²', 'м²'] as [string, string, string] // не склоняется
};
