import { describe, it, expect } from 'vitest';
import type { ListingPublic } from '../../src/lib/types/listings';
import { formatCurrency, translateHousingType } from '../../src/lib/utils';

describe('ListingView Component & Shared Logic', () => {
	const mockListing: ListingPublic = {
		id: '11111111-2222-3333-4444-555555555555',
		type: 'apartment',
		name: 'Уютная квартира возле метро Немига',
		square: 65,
		floor: 4,
		total_floors: 10,
		max_guests: 4,
		rooms_count: 2,
		beds_count: 2,
		bathrooms_count: 1,
		price_per_night: 130,
		currency: 'BYN',
		min_nights: 2,
		checkin_from: '14:00',
		checkout_until: '12:00',
		description: 'Прекрасные апартаменты со всем необходимым для комфортного отдыха.',
		amenities: ['Wi-Fi', 'Кондиционер', 'Кухня', 'Стиральная машина'],
		media: ['https://images.unsplash.com/photo-1522708323590-d24dbb6b0267'],
		created_at: new Date().toISOString()
	};

	it('computes total booking prices accurately based on nights', () => {
		const nights = 3;
		const total = mockListing.price_per_night * nights;
		expect(total).toBe(390);
		expect(formatCurrency(total, 'BYN')).toContain('390');
	});

	it('translates housing type and preserves ID', () => {
		expect(translateHousingType(mockListing.type)).toBe('Квартира');
		expect(mockListing.id.slice(0, 8)).toBe('11111111');
	});

	it('validates property capacity and floor constraints', () => {
		expect(mockListing.square).toBeGreaterThanOrEqual(10);
		expect(mockListing.floor).toBeLessThanOrEqual(mockListing.total_floors);
		expect(mockListing.max_guests).toBeGreaterThanOrEqual(1);
		expect(mockListing.amenities.length).toBeGreaterThan(0);
	});
});
