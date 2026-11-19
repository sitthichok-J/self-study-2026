import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { RouterTestingHarness } from '@angular/router/testing';
import { App } from './app';
import { AuthService } from './auth.service';
import { routes } from './app.routes';
import { CLAIMS_URL } from './claim.service';
import { TEST_RESPONSE } from './claim.fixtures';

describe('App', () => {
  beforeEach(async () => {
    TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [App],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter(routes)],
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(App);
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('แสดงหัวข้อ เมนู และ router-outlet', async () => {
    const fixture = TestBed.createComponent(App);
    await fixture.whenStable();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('h1')?.textContent).toContain('Claim Tracker');
    expect(Array.from(el.querySelectorAll('nav a')).map((a) => a.getAttribute('href'))).toEqual(['/claims', '/claims/new']);
    expect(el.querySelector('router-outlet')).toBeTruthy();
  });

  it('เข้า / แล้วเด้งไป /claims และโหลดรายการ', async () => {
    const harness = await RouterTestingHarness.create();
    await harness.navigateByUrl('/');
    TestBed.inject(HttpTestingController).expectOne((r) => r.url === CLAIMS_URL).flush(TEST_RESPONSE);
    await harness.fixture.whenStable();
    expect(TestBed.inject(AuthService).loggedIn()).toBe(false);
    expect(harness.routeNativeElement?.querySelectorAll('app-claim-card').length).toBe(5);
  });

  it('ปุ่มล็อกอิน (จำลอง) สลับสถานะและข้อความปุ่ม', async () => {
    const fixture = TestBed.createComponent(App);
    await fixture.whenStable();
    const btn = () => (fixture.nativeElement as HTMLElement).querySelector('button.auth') as HTMLButtonElement;
    expect(btn().textContent).toContain('ล็อกอิน');
    btn().click();
    await fixture.whenStable();
    expect(TestBed.inject(AuthService).loggedIn()).toBe(true);
    expect(btn().textContent).toContain('ออกจากระบบ');
    btn().click();
    await fixture.whenStable();
    expect(TestBed.inject(AuthService).loggedIn()).toBe(false);
  });
});
