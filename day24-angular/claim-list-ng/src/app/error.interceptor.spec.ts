import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ApiError } from './api-error';
import { errorInterceptor } from './error.interceptor';

describe('errorInterceptor', () => {
  let client: HttpClient;
  let ctl: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([errorInterceptor])), provideHttpClientTesting()],
    });
    client = TestBed.inject(HttpClient);
    ctl = TestBed.inject(HttpTestingController);
  });

  afterEach(() => ctl.verify());

  it('แปลง HTTP error เป็น ApiError พร้อม status และข้อความจาก body', () => {
    let err: unknown;
    client.get('/x').subscribe({ error: (e) => (err = e) });
    ctl.expectOne('/x').flush({ error: 'claim not found' }, { status: 404, statusText: 'Not Found' });
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(404);
    expect((err as ApiError).serverMessage).toBe('claim not found');
  });

  it('ต่อ backend ไม่ได้ (status 0) ได้ ApiError status 0', () => {
    let err: unknown;
    client.get('/x').subscribe({ error: (e) => (err = e) });
    ctl.expectOne('/x').error(new ProgressEvent('error'));
    expect((err as ApiError).status).toBe(0);
  });

  it('response สำเร็จผ่านไปโดยไม่ถูกแตะ', () => {
    let body: unknown;
    client.get('/x').subscribe((b) => (body = b));
    ctl.expectOne('/x').flush({ ok: true });
    expect(body).toEqual({ ok: true });
  });
});
