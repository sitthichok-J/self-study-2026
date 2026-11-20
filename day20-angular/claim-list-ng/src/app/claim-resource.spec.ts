import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ApplicationRef, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { claimsResource } from './claim-resource';
import { CLAIMS_URL } from './claim.service';
import { StatusFilter } from './claim-list/claim-list';
import { TEST_RESPONSE } from './claim.fixtures';

// resource ตั้งค่าผลหลัง response เข้ามาแบบ async จึงต้องรอให้แอป stable ก่อนอ่าน value()/error()
async function settle() {
  TestBed.tick();
  await TestBed.inject(ApplicationRef).whenStable();
}

describe('claimsResource (httpResource)', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting()] });
  });

  it('โหลดข้อมูลครั้งแรกโดยไม่ส่ง status และเปลี่ยนเป็น loading → มีค่า', async () => {
    const http = TestBed.inject(HttpTestingController);
    const status = signal<StatusFilter>('ALL');
    const res = TestBed.runInInjectionContext(() => claimsResource(status));
    TestBed.tick();
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    expect(req.request.params.has('status')).toBe(false);
    expect(res.isLoading()).toBe(true);
    req.flush(TEST_RESPONSE);
    await settle();
    expect(res.isLoading()).toBe(false);
    expect(res.value()?.items.length).toBe(5);
    http.verify();
  });

  it('เปลี่ยน signal ตัวกรองแล้วยิง request ใหม่พร้อม status', async () => {
    const http = TestBed.inject(HttpTestingController);
    const status = signal<StatusFilter>('ALL');
    const res = TestBed.runInInjectionContext(() => claimsResource(status));
    TestBed.tick();
    http.expectOne((r) => r.url === CLAIMS_URL).flush(TEST_RESPONSE);
    await settle();
    status.set('REJECTED');
    // ห้าม await whenStable ตรงนี้: request ยังค้างอยู่ (นับเป็น pending task) จะรอไม่จบ → แค่ tick ให้ resource ยิง request
    TestBed.tick();
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    expect(req.request.params.get('status')).toBe('REJECTED');
    req.flush({ items: [TEST_RESPONSE.items[2]], limit: 100, offset: 0 });
    await settle();
    expect(res.value()?.items.map((c) => c.id)).toEqual(['CLM-003']);
    http.verify();
  });

  it('error ของ backend ขึ้นที่ error() แทนที่จะโยนออกมา', async () => {
    const http = TestBed.inject(HttpTestingController);
    const res = TestBed.runInInjectionContext(() => claimsResource(signal<StatusFilter>('ALL')));
    TestBed.tick();
    http.expectOne((r) => r.url === CLAIMS_URL).flush('x', { status: 503, statusText: 'Unavailable' });
    await settle();
    expect(res.error()).toBeTruthy();
    // ห้ามอ่าน value() ตอน error: resource โยน error ออกมา ให้เช็คด้วย hasValue() แทน
    expect(res.hasValue()).toBe(false);
    http.verify();
  });
});
