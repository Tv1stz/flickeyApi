import { describe, it, expect } from 'vitest';
import { formatCurrency, maskPhone, translateHousingType, translateListingStatus } from '../../src/lib/utils';
import { ApiError } from '../../src/lib/types/api';

describe('Frontend Utilities', () => {
	it('formats BYN currency correctly', () => {
		expect(formatCurrency(150, 'BYN')).toContain('150');
		expect(formatCurrency(150, 'BYN')).toContain('BYN');
	});

	it('masks phone numbers for safe display', () => {
		const masked = maskPhone('+375291234567');
		expect(masked).toBe('+375 •• ••• •• 67');
	});

	it('translates housing types correctly', () => {
		expect(translateHousingType('apartment')).toBe('Квартира');
		expect(translateHousingType('house')).toBe('Дом / Коттедж');
		expect(translateHousingType('manor')).toBe('Усадьба');
	});

	it('translates listing statuses with badges', () => {
		const published = translateListingStatus('published');
		expect(published.label).toBe('Опубликовано');
		expect(published.color).toContain('emerald');

		const pending = translateListingStatus('pending_review');
		expect(pending.label).toBe('На модерации');
	});

	it('instantiates ApiError with status and code', () => {
		const err = new ApiError('INVALID_OTP', 'The provided OTP is invalid', 400, { attempts_left: 2 });
		expect(err.code).toBe('INVALID_OTP');
		expect(err.message).toBe('The provided OTP is invalid');
		expect(err.status).toBe(400);
		expect(err.details).toEqual({ attempts_left: 2 });
	});
});
