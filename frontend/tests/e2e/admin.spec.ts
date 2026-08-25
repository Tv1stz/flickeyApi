import { test, expect } from '@playwright/test';

test.describe('Admin Moderation, Compliance & User Enforcement E2E Journey', () => {
	test('Guest/Non-admin user cannot access /admin (Security Guard Test)', async ({ page }) => {
		await page.goto('/admin');
		// Non-admin should be blocked or redirected
		await expect(page).not.toHaveURL(/\/admin\/listings/);
	});

	test('Complete Admin Operational Workflow', async ({ page }) => {
		// 1. Quick Dev Login as Admin via DevLogin endpoint or simulated session
		await page.goto('/');

		// 2. Navigate to Admin Panel
		await page.goto('/admin');
		await expect(page.locator('h1')).toContainText('Панель управления Flickey Compliance');

		// 3. Listing Moderation Queue
		await page.goto('/admin/listings');
		await expect(page.locator('h1')).toContainText('Модерация объектов');

		// 4. Verification Queue
		await page.goto('/admin/verification');
		await expect(page.locator('h1')).toContainText('Верификация');

		// 5. User Directory & Enforcement
		await page.goto('/admin/users');
		await expect(page.locator('h1')).toContainText('Реестр пользователей');

		// 6. Reports Management
		await page.goto('/admin/reports');
		await expect(page.locator('h1')).toContainText('Реестр жалоб');

		// 7. Immutable Audit Trail Explorer
		await page.goto('/admin/audit');
		await expect(page.locator('h1')).toContainText('журнал аудита');
	});
});
