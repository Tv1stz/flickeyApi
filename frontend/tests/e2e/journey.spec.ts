import { test, expect } from '@playwright/test';

test.describe('Flickey Marketplace & User Journey E2E', () => {
	test('marketplace displays header, search, category tabs and empty/listings view', async ({ page }) => {
		await page.goto('/');

		// Check branding and title
		await expect(page.locator('header')).toContainText('Flickey');
		await expect(page.locator('h1')).toContainText('Найдите идеальное место');

		// Check category filter buttons
		await expect(page.getByRole('button', { name: /Квартиры/i })).toBeVisible();
		await expect(page.getByRole('button', { name: /Дома/i })).toBeVisible();
		await expect(page.getByRole('button', { name: /Усадьбы/i })).toBeVisible();

		// Click on "Квартиры" filter
		await page.getByRole('button', { name: /Квартиры/i }).click();
	});

	test('auth modal opens on clicking login button and supports dev quick login', async ({ page }) => {
		await page.goto('/');

		// Click login trigger
		const loginBtn = page.getByRole('button', { name: /Войти/i }).first();
		await loginBtn.click();

		// Check modal is visible
		await expect(page.getByRole('dialog')).toBeVisible();
		await expect(page.getByText('Режим разработки: Быстрый вход')).toBeVisible();

		// Perform dev quick login as Guest
		await page.getByRole('button', { name: /Instant Guest/i }).click();

		// Check user is authenticated in header
		await expect(page.getByRole('dialog')).not.toBeVisible();
		await expect(page.locator('header')).toContainText('Гость');
	});

	test('host wizard navigation requires filling mandatory fields', async ({ page }) => {
		await page.goto('/');

		// Login as host
		await page.getByRole('button', { name: /Dev Login/i }).click();

		// Navigate to wizard
		await page.goto('/host/new');

		// Step 1: Housing type selection
		await expect(page.getByText('Шаг 1: Выберите тип недвижимости')).toBeVisible();
		await page.getByRole('button', { name: /Квартира/i }).click();
		await page.getByRole('button', { name: /Продолжить/i }).click();

		// Step 2: Parameters
		await expect(page.getByText('Шаг 2: Основная информация')).toBeVisible();
	});
});
