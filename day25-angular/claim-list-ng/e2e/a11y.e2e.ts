import AxeBuilder from '@axe-core/playwright';
import { test, expect } from './fixtures';

// Day 25: axe-core ตรวจกฎ WCAG 2.x A/AA ที่ตรวจด้วยเครื่องได้ (ครอบคลุมแค่ส่วนหนึ่งของปัญหาจริง ไม่ได้แทนการลองใช้ screen reader / keyboard)
const TAGS = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

async function scan(page: import('@playwright/test').Page) {
  const r = await new AxeBuilder({ page }).withTags(TAGS).analyze();
  return r.violations.map((v) => `${v.id} (${v.impact}) x${v.nodes.length}: ${v.nodes[0].target.join(' ')}`);
}

test.describe('accessibility (axe)', () => {
  test('/claims', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await expect(page.locator('.card-id').first()).toBeVisible();
    expect(await scan(page)).toEqual([]);
  });

  test('/claims/CLM-003', async ({ page, api }) => {
    void api;
    await page.goto('/claims/CLM-003');
    await expect(page.locator('.d-id')).toBeVisible();
    expect(await scan(page)).toEqual([]);
  });

  test('/claims/new (reactive forms) หลังล็อกอิน', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await page.getByRole('button', { name: /ล็อกอิน/ }).click();
    await page.getByRole('link', { name: 'สร้างเคลม', exact: true }).click();
    await expect(page.getByLabel('patient_ref')).toBeVisible();
    expect(await scan(page)).toEqual([]);
  });

  test('/claims/new มี error ขึ้นที่ช่อง', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await page.getByRole('button', { name: /ล็อกอิน/ }).click();
    await page.getByRole('link', { name: 'สร้างเคลม', exact: true }).click();
    await page.getByRole('button', { name: 'ส่งเคลม' }).click();
    await expect(page.locator('.field-error').first()).toBeVisible();
    expect(await scan(page)).toEqual([]);
  });

  test('/claims/new-signal (Signal Forms) หลังล็อกอิน', async ({ page, api }) => {
    void api;
    await page.goto('/claims');
    await page.getByRole('button', { name: /ล็อกอิน/ }).click();
    await page.getByRole('link', { name: /Signal Forms/ }).click();
    await expect(page).toHaveURL(/new-signal$/);
    expect(await scan(page)).toEqual([]);
  });
});
