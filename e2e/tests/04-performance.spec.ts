import { test, expect } from '@playwright/test';

test.describe('Performance Tests', () => {
  test('page load performance', async ({ page }) => {
    const startTime = Date.now();
    
    await page.goto('/');
    
    // Wait for page to be fully loaded
    await page.waitForLoadState('networkidle');
    
    const loadTime = Date.now() - startTime;
    
    // Should load within 5 seconds
    expect(loadTime).toBeLessThan(5000);
    
    // Check for critical content
    await expect(page.locator('h1')).toBeVisible();
  });

  test('WebSocket connection latency', async ({ page, request }) => {
    // Create room via API
    const createResponse = await request.post('http://localhost:8080/api/room', {
      data: { roomName: 'Latency Test Room' }
    });
    const createData = await createResponse.json();
    const roomId = createData.roomId;
    
    await page.goto('/');
    
    // Navigate to room and measure connection time
    const startTime = Date.now();
    
    await page.goto(`/room/${roomId}?seat=0`);
    
    // Wait for WebSocket connection
    await expect(page.locator('text=已连接')).toBeVisible({ timeout: 10000 });
    
    const connectionTime = Date.now() - startTime;
    
    // Connection should be established within 3 seconds
    expect(connectionTime).toBeLessThan(3000);
  });

  test('action response time', async ({ page, request }) => {
    // Create and join room
    const createResponse = await request.post('http://localhost:8080/api/room', {
      data: { roomName: 'Response Time Test' }
    });
    const createData = await createResponse.json();
    const roomId = createData.roomId;
    
    await page.goto(`/room/${roomId}?seat=0`);
    await expect(page.locator('text=已连接')).toBeVisible();
    
    // Test button click response time
    const startTime = Date.now();
    
    // Try to pass (if enabled)
    const passButton = page.locator('button:has-text("过牌")');
    if (await passButton.isEnabled()) {
      await passButton.click();
      
      // Wait for some UI feedback
      await page.waitForTimeout(100);
      
      const responseTime = Date.now() - startTime;
      
      // Should respond within 200ms as per requirements
      expect(responseTime).toBeLessThan(200);
    }
  });

  test('memory usage stability', async ({ page }) => {
    await page.goto('/');
    
    // Get initial memory
    const initialMemory = await page.evaluate(() => {
      return (performance as any).memory?.usedJSHeapSize || 0;
    });
    
    // Perform multiple actions
    for (let i = 0; i < 10; i++) {
      await page.locator('button').filter({ hasText: '东 (East)' }).click();
      await page.locator('button').filter({ hasText: '南 (South)' }).click();
      await page.waitForTimeout(100);
    }
    
    // Check memory after actions
    const finalMemory = await page.evaluate(() => {
      return (performance as any).memory?.usedJSHeapSize || 0;
    });
    
    // Memory shouldn't grow significantly (within 50MB)
    if (initialMemory > 0 && finalMemory > 0) {
      const memoryGrowth = finalMemory - initialMemory;
      expect(memoryGrowth).toBeLessThan(50 * 1024 * 1024); // 50MB
    }
  });

  test('network resilience', async ({ page, context }) => {
    await page.goto('/');
    
    // Simulate slow network
    await context.route('**/*', async (route) => {
      await new Promise(resolve => setTimeout(resolve, 100)); // 100ms delay
      await route.continue();
    });
    
    // Actions should still work with network delay
    await page.locator('button').filter({ hasText: '东 (East)' }).click();
    await expect(page.locator('button').filter({ hasText: '东 (East)' })).toHaveClass(/border-blue-500/);
  });

  test('concurrent connections', async ({ browser }) => {
    const contexts = [];
    const pages = [];
    
    // Create 10 concurrent connections
    for (let i = 0; i < 10; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();
      contexts.push(context);
      pages.push(page);
    }
    
    try {
      // Navigate all pages simultaneously
      const startTime = Date.now();
      
      await Promise.all(
        pages.map(page => page.goto('/'))
      );
      
      // Wait for all pages to load
      await Promise.all(
        pages.map(page => expect(page.locator('h1')).toBeVisible())
      );
      
      const totalTime = Date.now() - startTime;
      
      // Should handle concurrent load within 10 seconds
      expect(totalTime).toBeLessThan(10000);
      
    } finally {
      // Clean up
      await Promise.all(contexts.map(context => context.close()));
    }
  });
});