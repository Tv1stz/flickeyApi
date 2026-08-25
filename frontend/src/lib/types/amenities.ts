export interface AmenityItem {
	id: string;
	name: string;
}

export interface AmenityCategory {
	id: string;
	name: string;
	amenities: AmenityItem[];
}

export interface AmenitiesResponse {
	housing_type: string | null;
	categories: AmenityCategory[];
}
