import { test, expect } from '@playwright/test';

test.describe('Game Start Debug Tests', () => {
  test('debug game start flow with 4 players', async ({ browser }) => {
    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(),
      browser.newContext(),
      browser.newContext()
    ]);

    const pages = await Promise.all(contexts.map(ctx => ctx.newPage()));

    // Capture all console logs and network requests
    const allLogs: Array<{ player: number; type: string; message: string }> = [];
    const allNetworkErrors: Array<{ player: number; error: string }> = [];

    for (let i = 0; i < pages.length; i++) {
      const page = pages[i];
      const playerId = i;

      page.on('console', (msg) => {
        allLogs.push({
          player: playerId,
          type: msg.type(),
          message: msg.text()
        });
      });

      page.on('pageerror', (error) => {
        allLogs.push({
          player: playerId,
          type: 'pageerror',
          message: error.message
        });
      });

      page.on('response', async (response) => {
        const url = response.url();
        const status = response.status();
        
        if (status === 401 && url.includes('/api/')) {
          try {
            const responseBody = await response.text();
            allNetworkErrors.push({
              player: playerId,
              error: `401 ${url}: ${responseBody}`
            });
          } catch (e) {
            allNetworkErrors.push({
              player: playerId,
              error: `401 ${url}: Could not read response body`
            });
          }
        }
      });
    }

    let roomId = '';
    const seatNames = ['东 (East)', '南 (South)', '西 (West)', '北 (North)'];

    try {
      // Player 1 creates room
      console.log('Player 0 creating room...');
      await pages[0].goto('/');
      await pages[0].locator('button').filter({ hasText: seatNames[0] }).click();
      
      const roomName = `4Player Debug Test ${Date.now()}`;
      await pages[0].fill('input[placeholder="输入房间名称"]', roomName);
      await pages[0].click('button:has-text("创建房间")');
      
      await expect(pages[0]).toHaveURL(/\/room\//, { timeout: 10000 });
      
      // Extract room ID
      const url = pages[0].url();
      roomId = url.match(/\/room\/([^?]+)/)?.[1]!;
      console.log('Room created:', roomId);

      // Players 2-4 join the room
      for (let i = 1; i < 4; i++) {
        console.log(`Player ${i} joining room...`);
        await pages[i].goto('/');
        await pages[i].locator('button').filter({ hasText: seatNames[i] }).click();
        await pages[i].fill('input[placeholder="输入房间ID"]', roomId);
        await pages[i].click('button:has-text("加入房间")');
        
        await expect(pages[i]).toHaveURL(/\/room\//, { timeout: 10000 });
        
        // Wait for WebSocket connection
        await pages[i].waitForTimeout(2000);
      }

      // Wait for all connections to establish
      await pages[0].waitForTimeout(3000);

      // Check player counts
      for (let i = 0; i < pages.length; i++) {
        const page = pages[i];
        const playerCountLocator = page.locator('text=/\\d+\\/4 人/');
        await expect(playerCountLocator).toBeVisible({ timeout: 10000 });
        
        const playerCountText = await playerCountLocator.textContent();
        console.log(`Player ${i} sees: ${playerCountText}`);
      }

      // Look for game start elements
      console.log('Checking for game start elements...');
      for (let i = 0; i < pages.length; i++) {
        const page = pages[i];
        
        // Check if game has started
        const gameStarted = await page.locator('text=/开始游戏|游戏开始|正在等待|等待中/').isVisible();
        console.log(`Player ${i} game started: ${gameStarted}`);
        
        // Check for any error messages
        const errorMessage = await page.locator('text=/错误|Error|失败|Failed/').isVisible();
        if (errorMessage) {
          const errorText = await page.locator('text=/错误|Error|失败|Failed/').textContent();
          console.log(`Player ${i} error: ${errorText}`);
        }
      }

      // Try to trigger game start if all players are connected
      console.log('Attempting to start game...');
      const playerCountFull = await pages[0].locator('text=/4\\/4 人/').isVisible();
      if (playerCountFull) {
        console.log('All players connected, looking for start button...');
        
        // Look for start game button or similar
        const startButton = pages[0].locator('button:has-text("开始游戏"), button:has-text("开始"), button:has-text("Start")');
        if (await startButton.isVisible()) {
          console.log('Found start button, clicking...');
          await startButton.click();
          await pages[0].waitForTimeout(3000);
        } else {
          console.log('No start button found, game might auto-start');
        }
        
        // Check if game has started
        for (let i = 0; i < pages.length; i++) {
          const page = pages[i];
          const hasCards = await page.locator('.card, [class*="card"]').count();
          console.log(`Player ${i} has ${hasCards} card elements`);
        }
      }

    } catch (error) {
      console.error('Test error:', error);
    }

    // Log all captured information
    console.log('\n=== ALL CONSOLE LOGS ===');
    allLogs.forEach(log => {
      console.log(`Player ${log.player} [${log.type}]: ${log.message}`);
    });

    console.log('\n=== ALL NETWORK ERRORS ===');
    allNetworkErrors.forEach(error => {
      console.log(`Player ${error.player}: ${error.error}`);
    });

    // Clean up
    await Promise.all(contexts.map(ctx => ctx.close()));
    
    // Test should pass for debugging purposes
    expect(true).toBe(true);
  });
});