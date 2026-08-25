import { apiRequest } from './client';
import type { AmenitiesResponse } from '$lib/types/amenities';
import type { HousingType } from '$lib/types/listings';

export const amenitiesApi = {
	/**
	 * Fetch grouped amenities catalog with optional housing type filtering.
	 */
	getAmenities(housingType?: HousingType): Promise<AmenitiesResponse> {
		const query = housingType ? `?housing_type=${encodeURIComponent(housingType)}` : '';
		return apiRequest<AmenitiesResponse>(`/amenities${query}`, {
			method: 'GET',
			skipAuth: true
		});
	}
};
