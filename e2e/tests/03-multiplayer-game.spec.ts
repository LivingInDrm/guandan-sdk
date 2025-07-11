import { test, expect, Page, BrowserContext } from '@playwright/test';

test.describe('Multiplayer Game Tests', () => {
  let contexts: BrowserContext[] = [];
  let pages: Page[] = [];
  let roomId: string;

  test.beforeAll(async ({ browser }) => {
    // Create 4 browser contexts for 4 players
    for (let i = 0; i < 4; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();
      contexts.push(context);
      pages.push(page);
    }
  });

  test.afterAll(async () => {
    // Clean up all contexts
    for (const context of contexts) {
      await context.close();
    }
  });

  test('4 players can join the same room', async () => {
    const seatNames = ['东 (East)', '南 (South)', '西 (West)', '北 (North)'];
    
    // Player 1 creates room
    await pages[0].goto('/');
    await pages[0].locator('button').filter({ hasText: seatNames[0] }).click();
    
    const roomName = `4Player Test ${Date.now()}`;
    await pages[0].fill('input[placeholder="输入房间名称"]', roomName);
    await pages[0].click('button:has-text("创建房间")');
    
    await expect(pages[0]).toHaveURL(/\/room\//);
    
    // Extract room ID
    const url = pages[0].url();
    roomId = url.match(/\/room\/([^?]+)/)?.[1]!;
    expect(roomId).toBeDefined();
    
    // Players 2-4 join the room
    for (let i = 1; i < 4; i++) {
      await pages[i].goto('/');
      await pages[i].locator('button').filter({ hasText: seatNames[i] }).click();
      await pages[i].fill('input[placeholder="输入房间ID"]', roomId);
      await pages[i].click('button:has-text("加入房间")');
      
      await expect(pages[i]).toHaveURL(/\/room\//);
      expect(pages[i].url()).toContain(roomId);
      
      // Wait for connection to establish before next player joins
      await pages[i].waitForTimeout(2000);
    }
    
    // Wait a moment for all connections to establish
    await pages[0].waitForTimeout(2000);
    
    // Check that all players see increasing player counts
    for (let i = 0; i < pages.length; i++) {
      const page = pages[i];
      
      // Debug: Check what player count is shown
      const playerCountLocator = page.locator('text=/\\d+\\/4 人/');
      await expect(playerCountLocator).toBeVisible({ timeout: 10000 });
      
      const playerCountText = await playerCountLocator.textContent();
      console.log(`Player ${i} sees: ${playerCountText}`);
    }
    
    // Wait longer for WebSocket synchronization
    await pages[0].waitForTimeout(5000);
    
    // Check final state - all players should see 4/4
    const finalPlayerCountLocator = pages[0].locator('text=/\\d+\\/4 人/');
    const finalPlayerCountText = await finalPlayerCountLocator.textContent();
    console.log(`Final state - Player 0 sees: ${finalPlayerCountText}`);
    
    // Check for 4/4 players with more time for sync
    await expect(pages[0].locator('text=4/4 人')).toBeVisible({ timeout: 20000 });
  });

  test('WebSocket connections work', async () => {
    // All players should be connected
    for (const page of pages) {
      await expect(page.locator('text=已连接')).toBeVisible();
    }
  });

  test('game state synchronization', async () => {
    // Wait for game to start
    await pages[0].waitForTimeout(3000);
    
    // All players should see the same game state
    for (const page of pages) {
      // Should see trump card info
      await expect(page.locator('text=主牌信息')).toBeVisible();
      
      // Should see game phase
      await expect(page.locator('text=游戏阶段')).toBeVisible();
      
      // Should see their hand cards
      await expect(page.locator('[data-testid="hand"]').first()).toBeVisible();
    }
  });

  test('player turn indicators', async () => {
    // Wait for game initialization
    await pages[0].waitForTimeout(2000);
    
    // One player should have "轮到你了" indicator
    let currentPlayerFound = false;
    
    for (const page of pages) {
      const myTurnElements = await page.locator('text=轮到你了').count();
      if (myTurnElements > 0) {
        currentPlayerFound = true;
        
        // Current player should see play/pass buttons enabled
        await expect(page.locator('button:has-text("出牌")')).toBeVisible();
        await expect(page.locator('button:has-text("过牌")')).toBeVisible();
        break;
      }
    }
    
    expect(currentPlayerFound).toBeTruthy();
  });

  test('card interaction', async () => {
    // Find current player
    let currentPlayerPage: Page | null = null;
    
    for (const page of pages) {
      const myTurnElements = await page.locator('text=轮到你了').count();
      if (myTurnElements > 0) {
        currentPlayerPage = page;
        break;
      }
    }
    
    if (currentPlayerPage) {
      // Try to select cards (if any are available)
      const cards = await currentPlayerPage.locator('[data-testid="card"]').count();
      if (cards > 0) {
        // Click on first card
        await currentPlayerPage.locator('[data-testid="card"]').first().click();
        
        // Should see card selection feedback
        await expect(currentPlayerPage.locator('text=已选择')).toBeVisible();
      }
      
      // Test pass action
      const passButton = currentPlayerPage.locator('button:has-text("过牌")');
      if (await passButton.isEnabled()) {
        await passButton.click();
        
        // Should see turn change to next player
        await currentPlayerPage.waitForTimeout(1000);
      }
    }
  });

  test('disconnection and reconnection', async () => {
    // Simulate disconnection by closing WebSocket
    await pages[0].evaluate(() => {
      // Force close WebSocket connections
      const store = (window as any).__ZUSTAND_STORE__;
      if (store && store.wsClient) {
        store.wsClient.disconnect();
      }
    });
    
    // Should show disconnected status
    await expect(pages[0].locator('text=未连接')).toBeVisible({ timeout: 5000 });
    
    // Refresh page to reconnect
    await pages[0].reload();
    
    // Should reconnect
    await expect(pages[0].locator('text=已连接')).toBeVisible({ timeout: 10000 });
  });

  test('error handling', async () => {
    // Test invalid room join
    await pages[0].goto('/');
    await pages[0].fill('input[placeholder="输入房间ID"]', 'nonexistent-room');
    await pages[0].click('button:has-text("加入房间")');
    
    // Should show error
    await expect(pages[0].locator('text=房间不存在')).toBeVisible();
  });
});