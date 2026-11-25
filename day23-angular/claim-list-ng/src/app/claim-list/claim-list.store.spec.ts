import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, TestRequest, provideHttpClientTesting } from '@angular/common/http/testing';
import { ApplicationRef } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Claim } from '../claim';
import { CLAIMS_URL } from '../claim.service';
import { ClaimListStore } from './claim-list.store';

const base = { created_at: '2026-11-25T01:00:00Z', updated_at: '2026-11-25T01:00:00Z' };
// แถวปลอม n แถว เริ่มที่ลำดับ start (id CLM-001, CLM-002, ...)
const rows = (n: number, start = 1): Claim[] =>
  Array.from({ length: n }, (_, i) => {
    const k = String(start + i).padStart(3, '0');
    return { id: `CLM-${k}`, patient_ref: `FAKE-T${k}`, amount: 1000, status: 'PENDING' as const, ...base };
  });

function setup() {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting(), ClaimListStore] });
  const http = TestBed.inject(HttpTestingController);
  const store = TestBed.inject(ClaimListStore);
  // รอ request ที่ resource ยิงหลัง tick แล้วตอบด้วย items ที่กำหนด
  const serve = async (items: Claim[]): Promise<TestRequest> => {
    TestBed.tick();
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    req.flush({ items, limit: Number(req.request.params.get('limit')), offset: Number(req.request.params.get('offset')) });
    await TestBed.inject(ApplicationRef).whenStable();
    return req;
  };
  return { http, store, serve };
}

const p = (r: TestRequest, k: string) => r.request.params.get(k);

describe('ClaimListStore (วันที่ 23)', () => {
  it('request แรก: หน้า 1 ขนาด 10 → limit 11 offset 0 ไม่มี status', async () => {
    const { store, serve } = setup();
    const r = await serve(rows(3));
    expect([p(r, 'limit'), p(r, 'offset'), r.request.params.has('status')]).toEqual(['11', '0', false]);
    expect(store.items().length).toBe(3);
    expect(store.hasNext()).toBe(false);
  });

  it('ได้ 11 แถว (เกิน pageSize) → แสดงแค่ 10 และมีหน้าถัดไป', async () => {
    const { store, serve } = setup();
    await serve(rows(11));
    expect(store.items().length).toBe(10);
    expect(store.items().at(-1)?.id).toBe('CLM-010');
    expect(store.hasNext()).toBe(true);
  });

  it('ได้พอดี 10 แถว → ไม่มีหน้าถัดไป (ไม่มีหน้าว่างต่อท้าย)', async () => {
    const { store, serve } = setup();
    await serve(rows(10));
    expect(store.items().length).toBe(10);
    expect(store.hasNext()).toBe(false);
  });

  it('next() ไปหน้า 2 ด้วย offset 10, prev() กลับหน้า 1 ด้วย offset 0', async () => {
    const { store, serve } = setup();
    await serve(rows(11));
    store.next();
    const r2 = await serve(rows(4, 11));
    expect(p(r2, 'offset')).toBe('10');
    expect(store.page()).toBe(1);
    expect(store.items().map((c) => c.id)).toEqual(['CLM-011', 'CLM-012', 'CLM-013', 'CLM-014']);
    expect(store.hasPrev()).toBe(true);
    store.prev();
    const r1 = await serve(rows(11));
    expect(p(r1, 'offset')).toBe('0');
    expect(store.page()).toBe(0);
  });

  it('next() ตอนไม่มีหน้าถัดไป และ prev() ตอนอยู่หน้าแรก ไม่ทำอะไรและไม่ยิง request', async () => {
    const { http, store, serve } = setup();
    await serve(rows(3));
    store.next();
    store.prev();
    TestBed.tick();
    expect(store.page()).toBe(0);
    http.expectNone((r) => r.url === CLAIMS_URL);
  });

  it('เปลี่ยน status ตอนอยู่หน้า 3 → กลับหน้าแรก และยิง request เดียว (offset 0 พร้อม status)', async () => {
    const { http, store, serve } = setup();
    await serve(rows(11));
    store.next();
    await serve(rows(11, 11));
    store.next();
    await serve(rows(2, 21));
    expect(store.page()).toBe(2);
    store.setStatus('APPROVED');
    TestBed.tick();
    const reqs = http.match((r) => r.url === CLAIMS_URL);
    expect(reqs.length).toBe(1);
    expect([p(reqs[0], 'status'), p(reqs[0], 'offset')]).toEqual(['APPROVED', '0']);
    reqs[0].flush({ items: [], limit: 11, offset: 0 });
    expect(store.page()).toBe(0);
  });

  it('เปลี่ยนขนาดหน้ากลับหน้าแรกและขอ limit = ขนาดใหม่ + 1', async () => {
    const { http, store, serve } = setup();
    await serve(rows(11));
    store.next();
    await serve(rows(5, 11));
    store.setPageSize(5);
    TestBed.tick();
    const reqs = http.match((r) => r.url === CLAIMS_URL);
    expect(reqs.length).toBe(1);
    expect([p(reqs[0], 'limit'), p(reqs[0], 'offset')]).toEqual(['6', '0']);
    reqs[0].flush({ items: rows(6), limit: 6, offset: 0 });
    await TestBed.inject(ApplicationRef).whenStable();
    expect(store.items().length).toBe(5);
    expect(store.hasNext()).toBe(true);
  });

  it('ตั้งสถานะเดิมซ้ำ ไม่ยิง request ใหม่', async () => {
    const { http, store, serve } = setup();
    await serve(rows(3));
    store.setStatus('ALL');
    TestBed.tick();
    http.expectNone((r) => r.url === CLAIMS_URL);
  });

  it('reload() ยิง request ซ้ำด้วย query เดิม', async () => {
    const { store, serve } = setup();
    await serve(rows(11));
    store.next();
    await serve(rows(2, 11));
    store.reload();
    const r = await serve(rows(2, 11));
    expect(p(r, 'offset')).toBe('10');
  });

  it('backend ตอบ error → error() มีค่า items ว่าง แล้ว reload สำเร็จ items กลับมา', async () => {
    const { http, store } = setup();
    TestBed.tick();
    http.expectOne((r) => r.url === CLAIMS_URL).flush('x', { status: 503, statusText: 'Unavailable' });
    await TestBed.inject(ApplicationRef).whenStable();
    expect(store.error()).toBeTruthy();
    expect(store.items()).toEqual([]);
    expect(store.hasNext()).toBe(false);
    store.reload();
    TestBed.tick();
    http.expectOne((r) => r.url === CLAIMS_URL).flush({ items: rows(2), limit: 11, offset: 0 });
    await TestBed.inject(ApplicationRef).whenStable();
    expect(store.error()).toBeNull();
    expect(store.items().length).toBe(2);
  });

  // baseline วันที่ 23: ลำดับ query ที่ออกไป backend เมื่อผู้ใช้ทำชุดการกระทำนี้ (ถ้าตั้งใจเปลี่ยน ให้รัน ng test -u)
  it('baseline: query ที่ส่งไป backend ตามลำดับการกระทำของผู้ใช้', async () => {
    const { store, serve } = setup();
    const log: string[] = [];
    const rec = (label: string, r: TestRequest) =>
      log.push(`${label}: ${r.request.urlWithParams.replace(CLAIMS_URL, '')}`);
    rec('เปิดหน้า', await serve(rows(11)));
    store.next();
    rec('ถัดไป', await serve(rows(11, 11)));
    store.setStatus('PENDING');
    rec('กรอง PENDING (กลับหน้าแรก)', await serve(rows(3)));
    store.setPageSize(5);
    rec('ขนาดหน้า 5', await serve(rows(6)));
    store.next();
    rec('ถัดไป', await serve(rows(1, 6)));
    store.setStatus('ALL');
    rec('กลับ ALL (กลับหน้าแรก)', await serve(rows(6)));
    expect(log).toMatchSnapshot();
  });
});
