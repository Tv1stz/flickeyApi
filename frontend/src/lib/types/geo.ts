export interface GeoSuggestItem {
	country: string;
	city: string;
	street: string;
	house_number: string;
	raw_address: string;
	latitude: number;
	longitude: number;
}

export interface GeoSuggestResponse {
	results: GeoSuggestItem[];
}

export interface GeoReverseResponse {
	item: GeoSuggestItem | null;
}
