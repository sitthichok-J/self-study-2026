import { HttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { RouterTestingHarness } from '@angular/router/testing';
import { ApiError } from './api-error';
import { appConfig } from './app.config';
import { ClaimDetail } from './claim-detail/claim-detail';
import { CLAIMS_URL } from './claim.service';
import { TEST_CLAIMS } from './claim.fixtures';
import { RETRY_DELAY_MS } from './retry.interceptor';

// Day 20: อุดช่องโหว่ที่จดไว้ใน Day 19 เทสต์ routes สร้าง config ของตัวเอง จึงไม่จับกรณี app.config.ts ลืม withComponentInputBinding()
// ไฟล์นี้ใช้ providers ของ appConfig จริง (เปลี่ยนเฉพาะ backend ของ HttpClient เป็นตัวทดสอบ)
function setup() {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({
    providers: [...appConfig.providers, provideHttpClientTesting(), { provide: RETRY_DELAY_MS, useValue: 0 }],
  });
  return TestBed.inject(HttpTestingController);
}

describe('appConfig (providers ตัวจริง)', () => {
  it('เปิด withComponentInputBinding: :id ใน URL เข้า input id ของ ClaimDetail', async () => {
    const http = setup();
    const harness = await RouterTestingHarness.create();
    const detail = await harness.navigateByUrl('/claims/CLM-002', ClaimDetail);
    TestBed.tick();
    expect(detail.id()).toBe('CLM-002');
    http.expectOne(`${CLAIMS_URL}/CLM-002`).flush(TEST_CLAIMS[1]);
  });

  // provideHttpClientTesting ให้ HttpClient อยู่แล้ว เทสต์ "inject ClaimService ได้" จึงจับการลืม provideHttpClient() ไม่ได้ ใช้การนับ provider แทน
  it('baseline: จำนวน provider ใน appConfig (ถ้าเพิ่ม/ลดต้องตั้งใจ)', () => {
    expect(appConfig.providers.length).toBe(3);
  });

  // วันที่ 6: interceptor ต้องถูกลงทะเบียนใน appConfig จริง (จำนวน provider ไม่เปลี่ยนเมื่อมี withInterceptors จึงต้องมีเทสต์พฤติกรรม)
  describe('interceptor ที่ลงใน appConfig', () => {
    const tick = () => new Promise((r) => setTimeout(r, 5));

    it('POST ที่ไม่มี key ถูกเติม Idempotency-Key', () => {
      const http = setup();
      TestBed.inject(HttpClient).post('/x', {}).subscribe();
      const req = http.expectOne('/x');
      expect(req.request.headers.get('Idempotency-Key')).toMatch(/^[0-9a-f-]{36}$/);
      req.flush({});
    });

    it('GET ได้ 503 สามครั้ง: ยิง 3 request แล้วได้ ApiError (retry อยู่ในสุด เห็น error ดิบก่อนถูกแปลง)', async () => {
      const http = setup();
      let err: unknown;
      TestBed.inject(HttpClient).get('/x').subscribe({ error: (e) => (err = e) });
      for (let i = 0; i < 3; i++) {
        http.expectOne('/x').flush({ error: 'down' }, { status: 503, statusText: 'Service Unavailable' });
        await tick();
      }
      http.expectNone('/x');
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(503);
      expect((err as ApiError).serverMessage).toBe('down');
    });

    it('POST ได้ 503 ไม่ retry และได้ ApiError', async () => {
      const http = setup();
      let err: unknown;
      TestBed.inject(HttpClient).post('/x', {}).subscribe({ error: (e) => (err = e) });
      http.expectOne('/x').flush('x', { status: 503, statusText: 'Service Unavailable' });
      await tick();
      http.expectNone('/x');
      expect((err as ApiError).status).toBe(503);
    });
  });
});
