import { describe, it, expect, vi } from 'vitest';
import { adminOverviewApi } from '$lib/api/admin/overview';
import { adminListingsApi } from '$lib/api/admin/listings';
import { adminUsersApi } from '$lib/api/admin/users';
import { adminVerificationApi } from '$lib/api/admin/verification';
import { adminReportsApi } from '$lib/api/admin/reports';
import { adminAuditApi } from '$lib/api/admin/audit';

describe('Admin API Modules Unit Tests', () => {
	it('should request overview metrics', async () => {
		const mockOverview = {
			total_listings: 10,
			pending_listings: 2,
			approved_listings: 7,
			rejected_listings: 1,
			total_users: 25,
			active_users: 23,
			blocked_users: 2,
			pending_verifications: 3,
			open_reports: 1
		};
		globalThis.fetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			headers: new Headers({ 'content-type': 'application/json' }),
			json: async () => mockOverview
		});

		const res = await adminOverviewApi.getOverview();
		expect(res.total_listings).toBe(10);
		expect(res.pending_listings).toBe(2);
	});

	it('should support moderation actions with reason taxonomy', async () => {
		const mockListing = { id: 'listing-1', status: 'published', name: 'Minsk Flat' };
		globalThis.fetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			headers: new Headers({ 'content-type': 'application/json' }),
			json: async () => mockListing
		});

		const res = await adminListingsApi.moderateListing('listing-1', {
			action: 'approve',
			reason: 'Passed verification',
			note: 'Looks good'
		});
		expect((res as any).status).toBe('published');
	});

	it('should support user enforcement helper functions', async () => {
		const mockUser = { id: 'user-1', phone: '+375291112233', role: 'guest', status: 'suspended' };
		globalThis.fetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			headers: new Headers({ 'content-type': 'application/json' }),
			json: async () => mockUser
		});

		const res = await adminUsersApi.restrictUser('user-1', 'Rules violation', 'Note test', 7);
		expect(res.status).toBe('suspended');
	});
});
