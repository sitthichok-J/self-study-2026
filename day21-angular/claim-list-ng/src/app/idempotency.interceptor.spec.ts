import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { IDEMPOTENCY_HEADER, idempotencyInterceptor } from './idempotency.interceptor';

describe('idempotencyInterceptor', () => {
  let client: HttpClient;
  let ctl: HttpTestingController;

  beforeEach(() => {
    vi.spyOn(crypto, 'randomUUID').mockReturnValue('00000000-0000-4000-8000-000000000001');
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([idempotencyInterceptor])), provideHttpClientTesting()],
    });
    client = TestBed.inject(HttpClient);
    ctl = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    ctl.verify();
    vi.restoreAllMocks();
  });

  it('POST ที่ไม่มี key ถูกเติม Idempotency-Key', () => {
    client.post('/x', {}).subscribe();
    const req = ctl.expectOne('/x');
    expect(req.request.headers.get(IDEMPOTENCY_HEADER)).toBe('00000000-0000-4000-8000-000000000001');
    req.flush({});
  });

  it('POST ที่ผู้เรียกใส่ key มาแล้ว ไม่ถูกเขียนทับ', () => {
    client.post('/x', {}, { headers: { [IDEMPOTENCY_HEADER]: 'mine' } }).subscribe();
    const req = ctl.expectOne('/x');
    expect(req.request.headers.get(IDEMPOTENCY_HEADER)).toBe('mine');
    req.flush({});
  });

  it('GET ไม่ถูกเติม key', () => {
    client.get('/x').subscribe();
    const req = ctl.expectOne('/x');
    expect(req.request.headers.has(IDEMPOTENCY_HEADER)).toBe(false);
    req.flush({});
  });
});
