import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ApplicationRef, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { ClaimListQuery, claimsResource } from './claim-resource';
import { CLAIMS_URL } from './claim.service';
import { TEST_RESPONSE } from './claim.fixtures';

// resource ตั้งค่าผลหลัง response เข้ามาแบบ async จึงต้องรอให้แอป stable ก่อนอ่าน value()/error()
async function settle() {
  TestBed.tick();
  await TestBed.inject(ApplicationRef).whenStable();
}

const q = (over: Partial<ClaimListQuery> = {}): ClaimListQuery => ({ status: 'ALL', pageSize: 10, page: 0, ...over });

describe('claimsResource (httpResource) วันที่ 23: status + หน้า ส่งไป backend', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting()] });
  });

  it('โหลดครั้งแรก: ไม่ส่ง status, ขอ pageSize+1 แถว, offset 0 และเปลี่ยนเป็น loading → มีค่า', async () => {
    const http = TestBed.inject(HttpTestingController);
    const res = TestBed.runInInjectionContext(() => claimsResource(signal(q())));
    TestBed.tick();
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    expect(req.request.params.has('status')).toBe(false);
    expect(req.request.params.get('limit')).toBe('11');
    expect(req.request.params.get('offset')).toBe('0');
    expect(res.isLoading()).toBe(true);
    req.flush(TEST_RESPONSE);
    await settle();
    expect(res.isLoading()).toBe(false);
    expect(res.value()?.items.length).toBe(5);
    http.verify();
  });

  it('เปลี่ยน status แล้วยิง request ใหม่พร้อม status', async () => {
    const http = TestBed.inject(HttpTestingController);
    const query = signal(q());
    const res = TestBed.runInInjectionContext(() => claimsResource(query));
    TestBed.tick();
    http.expectOne((r) => r.url === CLAIMS_URL).flush(TEST_RESPONSE);
    await settle();
    query.set(q({ status: 'REJECTED' }));
    // ห้าม await whenStable ตรงนี้: request ยังค้างอยู่ (นับเป็น pending task) จะรอไม่จบ → แค่ tick ให้ resource ยิง request
    TestBed.tick();
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    expect(req.request.params.get('status')).toBe('REJECTED');
    req.flush({ items: [TEST_RESPONSE.items[2]], limit: 11, offset: 0 });
    await settle();
    expect(res.value()?.items.map((c) => c.id)).toEqual(['CLM-003']);
    http.verify();
  });

  it('offset = page * pageSize และ limit = pageSize + 1 (หน้า 3 ขนาด 5 → offset 10 limit 6)', () => {
    const http = TestBed.inject(HttpTestingController);
    TestBed.runInInjectionContext(() => claimsResource(signal(q({ page: 2, pageSize: 5 }))));
    TestBed.tick();
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    expect(req.request.params.get('offset')).toBe('10');
    expect(req.request.params.get('limit')).toBe('6');
    req.flush(TEST_RESPONSE);
  });

  it('เปลี่ยนหน้าระหว่างที่ request เก่ายังค้าง: request เก่าถูกยกเลิก และใช้ผลของอันใหม่', async () => {
    const http = TestBed.inject(HttpTestingController);
    const query = signal(q());
    const res = TestBed.runInInjectionContext(() => claimsResource(query));
    TestBed.tick();
    const first = http.expectOne((r) => r.url === CLAIMS_URL);
    query.set(q({ page: 1 }));
    TestBed.tick();
    expect(first.cancelled).toBe(true);
    const second = http.expectOne((r) => r.url === CLAIMS_URL && r.params.get('offset') === '10');
    second.flush({ items: [TEST_RESPONSE.items[4]], limit: 11, offset: 10 });
    await settle();
    expect(res.value()?.items.map((c) => c.id)).toEqual(['CLM-005']);
    http.verify();
  });

  it('error ของ backend ขึ้นที่ error() แทนที่จะโยนออกมา', async () => {
    const http = TestBed.inject(HttpTestingController);
    const res = TestBed.runInInjectionContext(() => claimsResource(signal(q())));
    TestBed.tick();
    http.expectOne((r) => r.url === CLAIMS_URL).flush('x', { status: 503, statusText: 'Unavailable' });
    await settle();
    expect(res.error()).toBeTruthy();
    // ห้ามอ่าน value() ตอน error: resource โยน error ออกมา ให้เช็คด้วย hasValue() แทน
    expect(res.hasValue()).toBe(false);
    http.verify();
  });
});
