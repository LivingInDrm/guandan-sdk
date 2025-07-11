import { test, expect } from '@playwright/test';

test.describe('OAuth Debug Tests', () => {
  test('debug OAuth authentication error', async ({ page }) => {
    const consoleLogs: string[] = [];
    const consoleErrors: string[] = [];
    const networkErrors: string[] = [];

    // Capture console logs
    page.on('console', (msg) => {
      const text = `[${msg.type()}] ${msg.text()}`;
      consoleLogs.push(text);
      console.log('Console:', text);
    });

    // Capture console errors
    page.on('pageerror', (error) => {
      const text = `Page Error: ${error.message}`;
      consoleErrors.push(text);
      console.log('Page Error:', text);
    });

    // Capture network requests and responses
    page.on('request', (request) => {
      console.log('Request:', request.method(), request.url());
    });

    page.on('response', async (response) => {
      const url = response.url();
      const status = response.status();
      const statusText = response.statusText();
      
      console.log('Response:', response.request().method(), url, status, statusText);
      
      // Check for OAuth-related errors
      if (status === 401 && url.includes('/api/')) {
        try {
          const responseBody = await response.text();
          console.log('401 Response Body:', responseBody);
          networkErrors.push(`401 ${url}: ${responseBody}`);
        } catch (e) {
          console.log('Could not read response body');
        }
      }
    });

    // Navigate to the application
    console.log('Navigating to application...');
    await page.goto('/', { waitUntil: 'networkidle' });

    // Wait for the page to load
    await page.waitForTimeout(2000);

    // Try to create a room
    console.log('Testing room creation...');
    await page.locator('button').filter({ hasText: '东 (East)' }).click();
    
    const roomName = `OAuth Debug Test ${Date.now()}`;
    await page.fill('input[placeholder="输入房间名称"]', roomName);
    await page.click('button:has-text("创建房间")');

    // Wait for navigation or error
    await page.waitForTimeout(5000);

    // Check current URL
    const currentUrl = page.url();
    console.log('Current URL:', currentUrl);

    // Try to join a room if creation succeeded
    if (currentUrl.includes('/room/')) {
      console.log('Room creation succeeded, testing WebSocket...');
      
      // Wait for WebSocket connection
      await page.waitForTimeout(5000);
      
      // Check if game state is loaded
      const gameStateVisible = await page.locator('text=/\\d+\\/4 人/').isVisible();
      console.log('Game state visible:', gameStateVisible);
      
    } else {
      console.log('Room creation failed or redirected');
    }

    // Log all captured information
    console.log('\n=== CONSOLE LOGS ===');
    consoleLogs.forEach(log => console.log(log));

    console.log('\n=== CONSOLE ERRORS ===');
    consoleErrors.forEach(error => console.log(error));

    console.log('\n=== NETWORK ERRORS ===');
    networkErrors.forEach(error => console.log(error));

    // Check if there are any OAuth-related errors
    const hasOAuthError = networkErrors.some(error => 
      error.includes('oauth') || 
      error.includes('OAuth') || 
      error.includes('authentication_error') ||
      error.includes('Invalid OAuth token')
    );

    if (hasOAuthError) {
      console.log('\n=== OAUTH ERROR DETECTED ===');
      networkErrors.forEach(error => {
        if (error.includes('oauth') || error.includes('OAuth') || error.includes('authentication')) {
          console.log('OAuth Error:', error);
        }
      });
    }

    // Test should pass regardless of errors for debugging purposes
    expect(true).toBe(true);
  });
});