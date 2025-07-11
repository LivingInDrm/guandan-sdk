import { test, expect } from '@playwright/test';

test.describe('System Health Tests', () => {
  test('backend health check', async ({ request }) => {
    // Test backend health endpoint
    const response = await request.get('http://localhost:8080/api/health');
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    expect(data).toHaveProperty('status', 'ok');
    expect(data).toHaveProperty('timestamp');
    expect(data).toHaveProperty('rooms');
  });

  test('frontend loads correctly', async ({ page }) => {
    // Navigate to frontend
    await page.goto('/');
    
    // Check page title
    await expect(page).toHaveTitle(/掼蛋游戏/);
    
    // Check main heading
    await expect(page.locator('h1')).toContainText('掼蛋游戏');
    
    // Check seat selection section
    await expect(page.getByRole('heading', { name: '选择座位' })).toBeVisible();
    
    // Check create room section
    await expect(page.getByRole('heading', { name: '创建房间' })).toBeVisible();
    
    // Check join room section
    await expect(page.getByRole('heading', { name: '加入房间' })).toBeVisible();
  });

  test('API endpoints respond correctly', async ({ request }) => {
    // Test room listing (should be empty initially)
    const roomsResponse = await request.get('http://localhost:8080/api/rooms');
    expect(roomsResponse.ok()).toBeTruthy();
    
    const rooms = await roomsResponse.json();
    expect(Array.isArray(rooms)).toBeTruthy();
  });
});