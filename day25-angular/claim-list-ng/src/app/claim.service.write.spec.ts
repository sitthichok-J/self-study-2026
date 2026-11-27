import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ClaimService, CLAIMS_URL } from './claim.service';
import { bahtToSatang } from './claim';
import { TEST_CLAIMS } from './claim.fixtures';

describe('ClaimService (Day 4: get / create)', () => {
  let service: ClaimService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting()] });
    service = TestBed.inject(ClaimService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('get() ยิง GET /api/claims/{id} และ encode id', () => {
    let got: unknown;
    service.get('CLM 1/x').subscribe((c) => (got = c));
    const req = http.expectOne(`${CLAIMS_URL}/CLM%201%2Fx`);
    expect(req.request.method).toBe('GET');
    req.flush(TEST_CLAIMS[0]);
    expect(got).toEqual(TEST_CLAIMS[0]);
  });

  it('create() ยิง POST พร้อม body {patient_ref, amount} และ header Idempotency-Key', () => {
    service.create({ patient_ref: 'FAKE-AB12', amount: 150000 }, 'key-1').subscribe();
    const req = http.expectOne(CLAIMS_URL);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ patient_ref: 'FAKE-AB12', amount: 150000 });
    expect(req.request.headers.get('Idempotency-Key')).toBe('key-1');
    req.flush(TEST_CLAIMS[0], { status: 201, statusText: 'Created' });
  });
});

describe('bahtToSatang', () => {
  it('แปลงบาทเป็นสตางค์จำนวนเต็ม และกัน floating point เพี้ยน', () => {
    expect(bahtToSatang(1500)).toBe(150000);
    expect(bahtToSatang(0.01)).toBe(1);
    expect(bahtToSatang(19.99)).toBe(1999); // 19.99 * 100 = 1998.9999999999998
    expect(bahtToSatang(0.07)).toBe(7); // 0.07 * 100 = 7.000000000000001
  });
});
