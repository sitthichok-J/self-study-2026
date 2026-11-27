import { test, expect, cardIds } from './fixtures';

test.describe('หน้ารายการ (SSR)', () => {
  test('HTML ดิบจาก server มีการ์ด 10 ใบ ก่อนรัน JavaScript', async ({ request, api }) => {
    void api;
    const html = await (await request.get('/claims')).text();
    expect(html.match(/<app-claim-card/g)).toHaveLength(10);
    expect(html).toContain('CLM-023');
    expect(html).toContain('id="ng-state"');
  });

  test('เปิดหน้าแล้วเห็น CLM-023 ถึง CLM-014 และหน้า 1', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    const ids = await cardIds(page);
    expect(ids[0]).toBe('CLM-023');
    expect(ids.at(-1)).toBe('CLM-014');
    await expect(page.getByText('หน้า 1')).toBeVisible();
    await expect(page.getByRole('button', { name: /ก่อนหน้า/ })).toBeDisabled();
  });

  test('ถัดไปสองครั้งถึงหน้าสุดท้าย (3 รายการ) ปุ่มถัดไปกดไม่ได้ แล้วย้อนกลับได้', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    const next = page.getByRole('button', { name: /ถัดไป/ });
    await next.click();
    await expect(page.getByText('หน้า 2')).toBeVisible();
    await next.click();
    await expect(page.getByText('หน้า 3')).toBeVisible();
    await expect(page.locator('.card-id')).toHaveCount(3);
    await expect(next).toBeDisabled();
    await page.getByRole('button', { name: /ก่อนหน้า/ }).click();
    await expect(page.getByText('หน้า 2')).toBeVisible();
  });

  test('กรอง PENDING จากหน้า 3 แล้วกลับหน้า 1 และทุกใบเป็น PENDING', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    const next = page.getByRole('button', { name: /ถัดไป/ });
    await next.click();
    await expect(page.getByText('หน้า 2')).toBeVisible();
    await page.getByLabel('สถานะ').selectOption('PENDING');
    await expect(page.getByText('หน้า 1')).toBeVisible();
    // ระหว่างโหลดรายการหายไป (ดู README Day 23) จึงต้องรอจนมีการ์ด และทุกใบเป็น PENDING ไม่ใช่เช็คครั้งเดียว
    await expect
      .poll(async () => {
        const badges = (await page.locator('.status').allTextContents()).map((b) => b.trim());
        return badges.length > 0 && badges.every((b) => b === 'PENDING');
      })
      .toBe(true);
  });

  test('เลือกขนาดหน้า 5 แล้วเหลือการ์ด 5 ใบ', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await page.getByLabel('ต่อหน้า').selectOption('5');
    await expect(page.locator('.card-id')).toHaveCount(5);
  });

  test('backend ตอบ 503 สองครั้ง: retry ของ interceptor กู้ได้ ไม่เห็น error', async ({ page, api }) => {
    await page.goto('/claims');
    await expect(page.locator('.card-id').first()).toBeVisible();
    await api.fail(2);
    await page.getByRole('button', { name: /ถัดไป/ }).click();
    await expect(page.getByText('หน้า 2')).toBeVisible();
    await expect(page.locator('.card-id').first()).toHaveText('CLM-013', { timeout: 15_000 });
    await expect(page.locator('.error')).toHaveCount(0);
  });

  test('backend ตอบ 503 สามครั้ง: ขึ้น error แล้วกดลองใหม่หาย', async ({ page, api }) => {
    await page.goto('/claims');
    await expect(page.locator('.card-id').first()).toBeVisible();
    await api.fail(3);
    await page.getByRole('button', { name: /ถัดไป/ }).click();
    await expect(page.locator('.error')).toContainText('โหลดรายการไม่สำเร็จ', { timeout: 15_000 });
    await page.getByRole('button', { name: 'ลองใหม่' }).click();
    await expect(page.locator('.card-id').first()).toBeVisible();
    await expect(page.locator('.error')).toHaveCount(0);
  });
});

test.describe('หน้ารายละเอียด', () => {
  test('เปิด /claims/CLM-003 ตรง ๆ: server render มีรายละเอียดใน HTML', async ({ request, page, api }) => {
    void api;
    const html = await (await request.get('/claims/CLM-003')).text();
    expect(html).toContain('FAKE-EF56');
    await page.goto('/claims/CLM-003');
    await expect(page.locator('.d-status')).toHaveText('REJECTED');
    await expect(page.locator('.d-reason')).toHaveText('amount_over_limit');
  });

  test('กดรายละเอียดจากรายการแล้วไปหน้า detail และกลับได้', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await page.locator('app-claim-card', { hasText: 'CLM-022' }).getByRole('link', { name: 'รายละเอียด' }).click();
    await expect(page).toHaveURL(/\/claims\/CLM-022$/);
    await expect(page.locator('.d-id')).toHaveText('CLM-022');
    await page.getByRole('link', { name: /กลับไปรายการ/ }).click();
    await expect(page).toHaveURL(/\/claims$/);
  });

  test('id ที่ไม่มี: ขึ้นข้อความโหลดเคลมไม่สำเร็จ', async ({ page, api }) => {
    void api;
    await page.goto('/claims/CLM-999');
    await expect(page.locator('.error')).toContainText('โหลดเคลมไม่สำเร็จ');
  });
});

test.describe('สร้างเคลม (guard + ฟอร์ม)', () => {
  test('ยังไม่ล็อกอิน: เข้า /claims/new แล้วถูกส่งกลับ /claims', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await page.getByRole('link', { name: 'สร้างเคลม', exact: true }).click();
    // guard ส่งกลับพร้อม query ?login=required (ดู auth.guard.ts)
    await expect(page).toHaveURL(/\/claims\?login=required$/);
    await expect(page.locator('.card-id').first()).toBeVisible();
  });

  test('ล็อกอิน (จำลอง) แล้วสร้างเคลม: กลับหน้ารายการและเคลมใหม่อยู่ใบแรก', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await page.getByRole('button', { name: /ล็อกอิน/ }).click();
    await page.getByRole('link', { name: 'สร้างเคลม', exact: true }).click();
    await expect(page).toHaveURL(/\/claims\/new$/);
    await page.getByLabel('patient_ref').fill('FAKE-E2E1');
    await page.getByLabel('ยอดเงิน (บาท)').fill('1234.5');
    await page.getByRole('button', { name: 'ส่งเคลม' }).click();
    await expect(page).toHaveURL(/\/claims$/);
    await expect(page.locator('.card-id').first()).toHaveText('CLM-024');
    await expect(page.locator('app-claim-card').first()).toContainText('FAKE-E2E1');
  });

  test('patient_ref ผิดรูปแบบ: ขึ้น error ที่ช่อง และไม่ยิง POST', async ({ page, api }) => {
    void api;
    const posts: string[] = [];
    page.on('request', (r) => r.method() === 'POST' && posts.push(r.url()));
    await page.goto('/claims');
    await page.getByRole('button', { name: /ล็อกอิน/ }).click();
    await page.getByRole('link', { name: 'สร้างเคลม', exact: true }).click();
    await page.getByLabel('patient_ref').fill('abc');
    await page.getByLabel('ยอดเงิน (บาท)').fill('10');
    await page.getByRole('button', { name: 'ส่งเคลม' }).click();
    await expect(page.locator('[data-field="patient_ref"]')).toBeVisible();
    expect(posts).toHaveLength(0);
  });

  test('backend ปฏิเสธ (400): error ขึ้นบนฟอร์ม ยังอยู่หน้าเดิม', async ({ page, api }) => {
    await api.reject(1);
    await page.goto('/claims');
    await page.getByRole('button', { name: /ล็อกอิน/ }).click();
    await page.getByRole('link', { name: 'สร้างเคลม', exact: true }).click();
    await page.getByLabel('patient_ref').fill('FAKE-E2E2');
    await page.getByLabel('ยอดเงิน (บาท)').fill('50');
    await page.getByRole('button', { name: 'ส่งเคลม' }).click();
    await expect(page).toHaveURL(/\/claims\/new$/);
    await expect(page.locator('.server-error, .field-error').first()).toBeVisible();
  });
});
