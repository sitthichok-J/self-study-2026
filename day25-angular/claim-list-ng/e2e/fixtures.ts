import { test as base, expect } from '@playwright/test';

const API = 'http://localhost:8090';

// ทุกเทสต์เริ่มจากข้อมูลตั้งต้นของ mock (23 รายการ เรียงใหม่สุดก่อน) และไม่มี 503 ค้างจากเทสต์ก่อนหน้า
export const test = base.extend<{ api: { fail(n: number): Promise<void>; reject(n: number): Promise<void> } }>({
  api: async ({ request }, use) => {
    await request.post(`${API}/__reset`);
    await use({
      fail: async (n) => void (await request.post(`${API}/__fail?n=${n}`)),
      reject: async (n) => void (await request.post(`${API}/__reject?n=${n}`)),
    });
  },
});

export { expect };

export const cardIds = (page: import('@playwright/test').Page) => page.locator('.card-id').allTextContents();
