import { HttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { appConfig } from './app.config';
import { ClaimService } from './claim.service';

// baseline snapshot: request ที่ออกไปหา backend จริง (method, url, header) หลังผ่าน interceptor ทั้งสามตัว
// ถ้ามีใครเปลี่ยน header/ลำดับ/เส้นทาง snapshot นี้จะต่างทันที (ตั้งใจเปลี่ยนก็อัปเดต snapshot)
describe('baseline: request ที่ออกไป backend', () => {
  it('list, get, create', () => {
    vi.spyOn(crypto, 'randomUUID').mockReturnValue('00000000-0000-4000-8000-000000000001');
    TestBed.configureTestingModule({ providers: [...appConfig.providers, provideHttpClientTesting()] });
    const service = TestBed.inject(ClaimService);
    const ctl = TestBed.inject(HttpTestingController);
    void TestBed.inject(HttpClient);

    service.list({ status: 'APPROVED', limit: 10, offset: 0 }).subscribe();
    service.get('0b5f6c1e-aaaa-bbbb-cccc-111122223333').subscribe();
    service.create({ patient_ref: 'FAKE-AB12', amount: 150000 }, 'form-key-1').subscribe();

    const seen = [
      ...ctl.match(() => true).map((r) => ({
        method: r.request.method,
        url: r.request.urlWithParams,
        idempotencyKey: r.request.headers.get('Idempotency-Key'),
        body: r.request.body,
      })),
    ];
    expect(seen).toMatchSnapshot();
    vi.restoreAllMocks();
  });
});
