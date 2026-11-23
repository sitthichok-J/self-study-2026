import { HttpErrorResponse, provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ClaimService, CLAIMS_URL } from './claim.service';
import { TEST_RESPONSE } from './claim.fixtures';

describe('ClaimService', () => {
  let service: ClaimService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting()] });
    service = TestBed.inject(ClaimService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('เรียก GET /api/claims พร้อม limit=100 offset=0 เป็นค่าเริ่มต้น', () => {
    let result: unknown;
    service.list().subscribe((r) => (result = r));
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    expect(req.request.method).toBe('GET');
    expect(req.request.params.get('limit')).toBe('100');
    expect(req.request.params.get('offset')).toBe('0');
    expect(req.request.params.has('status')).toBe(false);
    req.flush(TEST_RESPONSE);
    expect(result).toEqual(TEST_RESPONSE);
  });

  it('ใส่ status, limit, offset ตามที่ส่งมา', () => {
    service.list({ status: 'APPROVED', limit: 10, offset: 20 }).subscribe();
    const req = http.expectOne((r) => r.url === CLAIMS_URL);
    expect(req.request.params.get('status')).toBe('APPROVED');
    expect(req.request.params.get('limit')).toBe('10');
    expect(req.request.params.get('offset')).toBe('20');
    req.flush(TEST_RESPONSE);
  });

  it('ส่งต่อ error ของ HTTP ให้ผู้เรียก', () => {
    let error: HttpErrorResponse | undefined;
    service.list().subscribe({ error: (e) => (error = e) });
    http.expectOne((r) => r.url === CLAIMS_URL).flush('boom', { status: 500, statusText: 'Server Error' });
    expect(error?.status).toBe(500);
  });
});
