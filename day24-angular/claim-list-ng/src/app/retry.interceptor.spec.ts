import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, TestRequest, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { MAX_GET_RETRIES, RETRY_DELAY_MS, retryInterceptor } from './retry.interceptor';

const tick = () => new Promise((r) => setTimeout(r, 5)); // delay = 0 ใน timer ยังเป็น macrotask

describe('retryInterceptor', () => {
  let client: HttpClient;
  let ctl: HttpTestingController;
  const fail = (req: TestRequest, status = 503) => req.flush('x', { status, statusText: 'err' });

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([retryInterceptor])),
        provideHttpClientTesting(),
        { provide: RETRY_DELAY_MS, useValue: 0 },
      ],
    });
    client = TestBed.inject(HttpClient);
    ctl = TestBed.inject(HttpTestingController);
  });

  afterEach(() => ctl.verify());

  it('จำนวน retry สูงสุดเป็น 2', () => {
    expect(MAX_GET_RETRIES).toBe(2);
  });

  it('GET ได้ 503 สองครั้งแล้วสำเร็จครั้งที่สาม ผู้เรียกได้ข้อมูลปกติ', async () => {
    let body: unknown;
    client.get('/x').subscribe((b) => (body = b));
    fail(ctl.expectOne('/x'));
    await tick();
    fail(ctl.expectOne('/x'));
    await tick();
    ctl.expectOne('/x').flush({ ok: 1 });
    expect(body).toEqual({ ok: 1 });
  });

  it('GET ได้ 503 สามครั้ง หยุดที่ 3 request แล้วส่ง error ออกไป', async () => {
    let status: number | undefined;
    client.get('/x').subscribe({ error: (e) => (status = e.status) });
    fail(ctl.expectOne('/x'));
    await tick();
    fail(ctl.expectOne('/x'));
    await tick();
    fail(ctl.expectOne('/x'));
    await tick();
    ctl.expectNone('/x'); // ไม่มี request ที่สี่
    expect(status).toBe(503);
  });

  it('GET ได้ 500 หรือ 404 ไม่ retry', async () => {
    const statuses: number[] = [];
    client.get('/a').subscribe({ error: (e) => statuses.push(e.status) });
    fail(ctl.expectOne('/a'), 500);
    client.get('/b').subscribe({ error: (e) => statuses.push(e.status) });
    fail(ctl.expectOne('/b'), 404);
    await tick();
    ctl.expectNone('/a');
    ctl.expectNone('/b');
    expect(statuses).toEqual([500, 404]);
  });

  it('POST ได้ 503 ไม่ retry (กันสร้างเคลมซ้ำ)', async () => {
    let status: number | undefined;
    client.post('/x', {}).subscribe({ error: (e) => (status = e.status) });
    fail(ctl.expectOne('/x'));
    await tick();
    ctl.expectNone('/x');
    expect(status).toBe(503);
  });
});
