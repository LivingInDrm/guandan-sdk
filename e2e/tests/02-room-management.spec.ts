import { test, expect } from '@playwright/test';

test.describe('Room Management Tests', () => {
  test('create and join room flow', async ({ page, request }) => {
    await page.goto('/');
    
    // Select seat (East - seat 0)
    await page.locator('button').filter({ hasText: '东 (East)' }).click();
    
    // Create room
    const roomName = `Test Room ${Date.now()}`;
    await page.fill('input[placeholder="输入房间名称"]', roomName);
    await page.click('button:has-text("创建房间")');
    
    // Should navigate to room page
    await expect(page).toHaveURL(/\/room\//);
    
    // Extract room ID from URL
    const url = page.url();
    const roomId = url.match(/\/room\/([^?]+)/)?.[1];
    expect(roomId).toBeDefined();
    
    // Check room page elements
    await expect(page.locator('text=房间')).toBeVisible();
    await expect(page.locator('text=等待中')).toBeVisible();
    await expect(page.locator('text=1/4 人')).toBeVisible();
    
    // Verify room was created via API
    const roomsResponse = await request.get('http://localhost:8080/api/rooms');
    const rooms = await roomsResponse.json();
    const createdRoom = rooms.find((room: any) => room.roomId === roomId);
    expect(createdRoom).toBeDefined();
    expect(createdRoom.playerCount).toBe(1);
    expect(createdRoom.maxPlayers).toBe(4);
  });

  test('join existing room', async ({ page, request }) => {
    // First create a room via API
    const createResponse = await request.post('http://localhost:8080/api/room', {
      data: { roomName: 'API Test Room' }
    });
    const createData = await createResponse.json();
    const roomId = createData.roomId;
    
    await page.goto('/');
    
    // Select seat (South - seat 1)
    await page.locator('button').filter({ hasText: '南 (South)' }).click();
    
    // Join room
    await page.fill('input[placeholder="输入房间ID"]', roomId);
    await page.click('button:has-text("加入房间")');
    
    // Should navigate to room page
    await expect(page).toHaveURL(/\/room\//);
    
    // Check we're in the correct room
    expect(page.url()).toContain(roomId);
  });

  test('room listing and refresh', async ({ page, request }) => {
    // Create a room via API
    const createResponse = await request.post('http://localhost:8080/api/room', {
      data: { roomName: 'Listing Test Room' }
    });
    
    await page.goto('/');
    
    // Check room list section
    await expect(page.locator('text=房间列表')).toBeVisible();
    
    // Click refresh button
    await page.click('button:has-text("刷新")');
    
    // Should see at least one room in the list (check for room ID format)
    await expect(page.locator('text=/房间 room_\\d+/').first()).toBeVisible();
    
    // Check room list contains our created room
    const roomsResponse = await request.get('http://localhost:8080/api/rooms');
    const rooms = await roomsResponse.json();
    expect(rooms.length).toBeGreaterThan(0);
  });

  test('input validation', async ({ page }) => {
    await page.goto('/');
    
    // Try to create room with empty name
    await page.click('button:has-text("创建房间")');
    
    // Should show error
    await expect(page.locator('text=请输入房间名称')).toBeVisible();
    
    // Try to join room with empty ID
    await page.click('button:has-text("加入房间")');
    
    // Should show error
    await expect(page.locator('text=请输入房间ID')).toBeVisible();
  });
});