import { defineConfig } from '@playwright/test';

// Day 25: e2e รันกับ production build แบบ SSR (node dist/.../server.mjs) + mock API ไม่ใช่ ng serve
// ต้อง `npm run build` ก่อน (npm run ci:e2e ทำให้) พอร์ตแยกจากที่ใช้เล่นเอง จะได้ไม่ชนกับ mock/ng serve ที่เปิดค้าง
const API_PORT = 8090;
const APP_PORT = 4310;

export default defineConfig({
  testDir: './e2e',
  testMatch: '**/*.e2e.ts',
  fullyParallel: false, // mock API มี state เดียว (reset ก่อนทุกเทสต์) จึงรันทีละตัว
  workers: 1,
  retries: process.env['CI'] ? 1 : 0,
  reporter: process.env['CI'] ? [['github'], ['list']] : 'list',
  timeout: 30_000,
  use: {
    baseURL: `http://localhost:${APP_PORT}`,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [{ name: 'chromium', use: { browserName: 'chromium' } }],
  webServer: [
    {
      command: 'node ../mock-api/server.mjs',
      env: { PORT: String(API_PORT), DELAY_MS: '50' },
      url: `http://localhost:${API_PORT}/claims?limit=1`,
      reuseExistingServer: !process.env['CI'],
    },
    {
      command: 'node dist/claim-list-ng/server/server.mjs',
      env: { PORT: String(APP_PORT), BACKEND_URL: `http://localhost:${API_PORT}` },
      url: `http://localhost:${APP_PORT}/claims`,
      reuseExistingServer: !process.env['CI'],
    },
  ],
});
