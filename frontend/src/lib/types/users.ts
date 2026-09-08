export interface PublicUser {
	id: string;
	name: string;
	first_name?: string;
	last_name?: string;
	role: string;
	is_verified: boolean;
	listings_count: number;
	created_at: string;
}
