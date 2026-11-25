import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ApplicationRef } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { ClaimDetail } from './claim-detail';
import { CLAIMS_URL } from '../claim.service';
import { TEST_CLAIMS } from '../claim.fixtures';

// httpResource ตั้งผลหลัง response แบบ async: tick ให้ resource ยิง request ก่อน แล้วค่อยรอ stable หลัง flush
async function create(id: string) {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({ imports: [ClaimDetail], providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])] });
  const fixture = TestBed.createComponent(ClaimDetail);
  fixture.componentRef.setInput('id', id);
  fixture.detectChanges();
  TestBed.tick();
  const http = TestBed.inject(HttpTestingController);
  const settle = async () => {
    TestBed.tick();
    await TestBed.inject(ApplicationRef).whenStable();
    fixture.detectChanges();
  };
  return { fixture, http, settle, el: fixture.nativeElement as HTMLElement };
}

describe('ClaimDetail (httpResource + id จาก input)', () => {
  it('ยิง GET ด้วย id ที่ได้จาก URL และแสดง loading ระหว่างรอ', async () => {
    const { http, el } = await create('CLM-002');
    const req = http.expectOne(`${CLAIMS_URL}/CLM-002`);
    expect(req.request.method).toBe('GET');
    expect(el.querySelector('.loading')?.textContent).toContain('กำลังโหลด');
    req.flush(TEST_CLAIMS[1]);
  });

  it('โหลดสำเร็จแสดงรายละเอียด ยอดเป็นบาท', async () => {
    const { http, settle, el } = await create('CLM-002');
    http.expectOne(`${CLAIMS_URL}/CLM-002`).flush(TEST_CLAIMS[1]);
    await settle();
    expect(el.querySelector('.d-id')?.textContent).toBe('CLM-002');
    expect(el.querySelector('.d-ref')?.textContent).toBe('FAKE-CD34');
    expect(el.querySelector('.d-amount')?.textContent).toBe('32000 บาท');
    expect(el.querySelector('.d-status')?.textContent).toBe('APPROVED');
    expect(el.querySelector('.d-reason')).toBeNull();
    expect(el.querySelector('.loading')).toBeNull();
  });

  it('REJECTED แสดงเหตุผล', async () => {
    const { http, settle, el } = await create('CLM-003');
    http.expectOne(`${CLAIMS_URL}/CLM-003`).flush(TEST_CLAIMS[2]);
    await settle();
    expect(el.querySelector('.d-reason')?.textContent).toBe('amount_over_limit');
  });

  it('404 แสดง "ไม่พบเคลมนี้" และไม่โยน error จากการอ่าน value()', async () => {
    const { http, settle, el } = await create('nope');
    http.expectOne(`${CLAIMS_URL}/nope`).flush({ error: 'claim not found' }, { status: 404, statusText: 'Not Found' });
    await settle();
    expect(el.querySelector('.error')?.textContent).toContain('ไม่พบเคลมนี้');
    expect(el.querySelector('.detail')).toBeNull();
  });

  it('503 แสดงรหัส HTTP', async () => {
    const { http, settle, el } = await create('x');
    http.expectOne(`${CLAIMS_URL}/x`).flush('x', { status: 503, statusText: 'Unavailable' });
    await settle();
    expect(el.querySelector('.error')?.textContent).toContain('backend ตอบ HTTP 503');
  });

  it('เปลี่ยน id แล้วยิง request ใหม่', async () => {
    const { fixture, http, settle } = await create('CLM-001');
    http.expectOne(`${CLAIMS_URL}/CLM-001`).flush(TEST_CLAIMS[0]);
    await settle();
    fixture.componentRef.setInput('id', 'CLM-004');
    fixture.detectChanges();
    TestBed.tick();
    http.expectOne(`${CLAIMS_URL}/CLM-004`).flush(TEST_CLAIMS[3]);
    await settle();
    expect((fixture.nativeElement as HTMLElement).querySelector('.d-id')?.textContent).toBe('CLM-004');
  });
});
