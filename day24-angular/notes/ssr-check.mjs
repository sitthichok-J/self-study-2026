import { createRequire } from 'node:module'; const require = createRequire(import.meta.url); const { chromium } = require('playwright'); import fs from 'node:fs';
const out = process.argv[2]; fs.mkdirSync(out, { recursive: true });
const B = 'http://localhost:4300'; const log = []; const L = (s) => { console.log(s); log.push(s); };
const b = await chromium.launch();
// A. HTML ดิบจาก server (ไม่รัน JS)
for (const p of ['/claims', '/claims/CLM-003', '/claims/new']) {
  const html = await (await fetch(B + p)).text();
  L(`A ${p}: HTML ${html.length} bytes | การ์ด ${(html.match(/<app-claim-card/g) || []).length} | pager ${html.includes('class="pager"')} | detail ${html.includes('class="detail"')} | มี ng-state ${html.includes('id="ng-state"')} | ng-server-context ${(html.match(/ng-server-context="(\w+)"/) || [])[1] || '-'}`);
}
// B. ปิด JS แล้วเปิดหน้า
{
  const c = await b.newContext({ javaScriptEnabled: false, viewport: { width: 1000, height: 800 }, deviceScaleFactor: 2 }); const pg = await c.newPage();
  await pg.goto(B + '/claims'); L('B ปิด JavaScript: เห็นการ์ด ' + (await pg.locator('app-claim-card').count()) + ' ใบ, ปุ่มถัดไป disabled=' + (await pg.locator('.next').isDisabled()));
  await pg.screenshot({ path: `${out}/s1-nojs.png` }); await c.close();
}
// C. เปิดด้วย JS: DOM ถูกใช้ซ้ำไหม, มี request /api ตอนโหลดไหม, console error
{
  const c = await b.newContext({ viewport: { width: 1000, height: 900 }, deviceScaleFactor: 2 }); const pg = await c.newPage();
  const reqs = [], files = [], errs = [];
  pg.on('request', (r) => { const u = new URL(r.url()); if (u.pathname.startsWith('/api/')) reqs.push(r.method() + ' ' + u.pathname + u.search); else if (/\.(js|css)$/.test(u.pathname)) files.push(u.pathname.slice(1)); });
  pg.on('console', (m) => { if (['error', 'warning'].includes(m.type())) errs.push(m.type() + ': ' + m.text()); }); pg.on('pageerror', (e) => errs.push('pageerror: ' + e.message));
  await pg.addInitScript(() => { document.addEventListener('DOMContentLoaded', () => { window.__first = document.querySelector('.card-id'); window.__ssrCards = document.querySelectorAll('app-claim-card').length; }); });
  await pg.goto(B + '/claims'); await pg.waitForLoadState('networkidle'); await pg.waitForTimeout(500);
  const same = await pg.evaluate(() => ({ same: window.__first === document.querySelector('.card-id'), connected: window.__first?.isConnected, ssr: window.__ssrCards, now: document.querySelectorAll('app-claim-card').length }));
  L(`C โหลด /claims ด้วย JS: การ์ดใน HTML ตอน DOMContentLoaded ${same.ssr} → ตอนนี้ ${same.now} | node เดิมถูกใช้ซ้ำ=${same.same} (isConnected=${same.connected})`);
  L('C request ไป /api ตอนโหลดหน้า (browser): ' + (reqs.length ? reqs.join(', ') : 'ไม่มีเลย (ใช้ข้อมูลที่ฝังมากับ HTML)'));
  L('C ไฟล์ js/css ที่โหลดตอนเปิด /claims: ' + files.join(', '));
  L('C console error/warning: ' + (errs.length ? errs.join(' | ') : 'ไม่มี'));
  await pg.screenshot({ path: `${out}/s2-hydrated.png` });
  // ไปหน้า detail ด้วย client navigation → โหลด chunk lazy
  const before = files.length; await pg.click('app-claim-card >> nth=2'); await pg.waitForTimeout(300);
  L('C คลิกการ์ด (ยังอยู่หน้า list) ไฟล์ใหม่: ' + (files.slice(before).join(', ') || 'ไม่มี'));
  await pg.goto(B + '/claims'); await pg.waitForLoadState('networkidle'); const b2 = files.length;
  await pg.click('button.auth'); await pg.evaluate(() => document.querySelector('a[href="/claims/new"]').click()); await pg.waitForTimeout(800);
  L('C กดลิงก์ "สร้างเคลม" (client navigation) ไฟล์ใหม่: ' + (files.slice(b2).join(', ') || 'ไม่มี') + ' | url=' + new URL(pg.url()).pathname);
  const b3 = files.length; await pg.goto(B + '/claims/CLM-003'); await pg.waitForLoadState('networkidle');
  L('C เปิด /claims/CLM-003 ตรง ๆ (server render) ไฟล์ที่โหลด: ' + files.slice(b3).join(', ') + ' | detail=' + (await pg.locator('.d-id').innerText()) + ' | request /api: ' + (reqs.slice(1).join(', ') || 'ไม่มี'));
  await pg.screenshot({ path: `${out}/s5-detail.png` });
  await c.close();
}
// D. event replay: หน่วง main js 2 วินาที กดถัดไปก่อน hydrate เสร็จ
{
  const c = await b.newContext({ viewport: { width: 1000, height: 900 }, deviceScaleFactor: 2 }); const pg = await c.newPage();
  const reqs = []; pg.on('request', (r) => { const u = new URL(r.url()); if (u.pathname.startsWith('/api/')) reqs.push(u.pathname + u.search); });
  await pg.route(/main-.*\.js$/, async (r) => { await new Promise((x) => setTimeout(x, 2500)); r.continue(); });
  await pg.goto(B + '/claims', { waitUntil: 'domcontentloaded' });
  await pg.waitForSelector('app-claim-card');
  const t0 = Date.now(); await pg.screenshot({ path: `${out}/s3-before-hydrate.png` });
  L('D ก่อน JS มาถึง: เห็นการ์ด ' + (await pg.locator('app-claim-card').count()) + ' ใบ, หน้า "' + (await pg.locator('.page-info').innerText()) + '" ปุ่มถัดไป disabled=' + (await pg.locator('.next').isDisabled()));
  await pg.click('.next', { noWaitAfter: true }); L('D คลิก "ถัดไป" ตอนที่ยังไม่ hydrate (+' + (Date.now() - t0) + ' ms)');
  await pg.waitForFunction(() => document.querySelector('.page-info')?.textContent.includes('2'), null, { timeout: 15000 });
  await pg.waitForFunction(() => document.querySelectorAll('app-claim-card').length > 0, null, { timeout: 15000 });
  L('D หลัง hydrate: หน้า "' + (await pg.locator('.page-info').innerText()) + '" การ์ด ' + (await pg.locator('app-claim-card').count()) + ' ใบ | request /api: ' + reqs.join(', '));
  await pg.screenshot({ path: `${out}/s4-replayed.png` }); await c.close();
}
await b.close(); fs.writeFileSync(`${out}/../cap24-log.txt`, log.join('\n') + '\n');
