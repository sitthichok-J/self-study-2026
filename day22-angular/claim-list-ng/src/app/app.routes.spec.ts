import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { Router, provideRouter, withComponentInputBinding } from '@angular/router';
import { RouterTestingHarness } from '@angular/router/testing';
import { AuthService } from './auth.service';
import { routes } from './app.routes';
import { CLAIMS_URL } from './claim.service';
import { TEST_CLAIMS, TEST_RESPONSE } from './claim.fixtures';

function setup() {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({
    providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter(routes, withComponentInputBinding())],
  });
  return { http: TestBed.inject(HttpTestingController), router: TestBed.inject(Router), auth: TestBed.inject(AuthService) };
}

describe('routes', () => {
  it('baseline: ตาราง route ปัจจุบัน (path, redirect, lazy, guard)', () => {
    const table = routes.map((r) => ({
      path: r.path,
      redirectTo: r.redirectTo ?? null,
      eager: !!r.component,
      lazy: !!r.loadComponent,
      guards: (r.canActivate ?? []).length,
    }));
    expect(table).toMatchSnapshot();
  });

  it('/claims/new อยู่ก่อน /claims/:id (ไม่งั้น "new" ถูกอ่านเป็น id)', () => {
    const paths = routes.map((r) => r.path);
    expect(paths.indexOf('claims/new')).toBeLessThan(paths.indexOf('claims/:id'));
  });

  it('หน้ารายละเอียดและหน้าสร้างเป็น lazy ส่วนหน้ารายการ eager', () => {
    const by = (p: string) => routes.find((r) => r.path === p)!;
    expect(by('claims/new').loadComponent).toBeTypeOf('function');
    expect(by('claims/:id').loadComponent).toBeTypeOf('function');
    expect(by('claims').component).toBeDefined();
  });

  it('/ เด้งไป /claims', async () => {
    const { http, router } = setup();
    const harness = await RouterTestingHarness.create();
    await harness.navigateByUrl('/');
    http.expectOne((r) => r.url === CLAIMS_URL).flush(TEST_RESPONSE);
    await harness.fixture.whenStable();
    expect(router.url).toBe('/claims');
  });

  it('/claims/new ตอนยังไม่ล็อกอิน ถูก guard เด้งกลับ /claims?login=required', async () => {
    const { http, router } = setup();
    const harness = await RouterTestingHarness.create();
    await harness.navigateByUrl('/claims/new');
    http.expectOne((r) => r.url === CLAIMS_URL).flush(TEST_RESPONSE);
    await harness.fixture.whenStable();
    expect(router.url).toBe('/claims?login=required');
  });

  it('/claims/new ตอนล็อกอินแล้วเข้าได้ และโหลดฟอร์มแบบ lazy', async () => {
    const { auth, router } = setup();
    auth.login();
    const harness = await RouterTestingHarness.create();
    await harness.navigateByUrl('/claims/new');
    expect(router.url).toBe('/claims/new');
    expect(harness.routeNativeElement?.tagName.toLowerCase()).toBe('app-claim-form');
  });

  it('/claims/:id ส่ง id จาก URL เข้า input() ของ component และยิง GET ด้วย id นั้น', async () => {
    const { http } = setup();
    const harness = await RouterTestingHarness.create();
    await harness.navigateByUrl('/claims/CLM-005');
    harness.fixture.detectChanges();
    TestBed.tick();
    http.expectOne(`${CLAIMS_URL}/CLM-005`).flush(TEST_CLAIMS[4]);
    expect(harness.routeNativeElement?.tagName.toLowerCase()).toBe('app-claim-detail');
    expect(harness.routeNativeElement?.querySelector('h2')?.textContent).toContain('CLM-005');
  });
});
