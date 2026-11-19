import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router, RouterStateSnapshot, UrlTree, provideRouter } from '@angular/router';
import { AuthService } from './auth.service';
import { authGuard } from './auth.guard';

const run = () =>
  TestBed.runInInjectionContext(() => authGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot));

describe('authGuard', () => {
  beforeEach(() => {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({ providers: [provideRouter([])] });
  });

  it('ล็อกอินแล้วผ่าน (true)', () => {
    TestBed.inject(AuthService).login();
    expect(run()).toBe(true);
  });

  it('ยังไม่ล็อกอินได้ UrlTree ไป /claims?login=required (ไม่ได้ navigate เอง)', () => {
    const result = run();
    expect(result).toBeInstanceOf(UrlTree);
    expect(TestBed.inject(Router).serializeUrl(result as UrlTree)).toBe('/claims?login=required');
  });

  it('ล็อกเอาท์หลังล็อกอินแล้วถูกกันอีกครั้ง', () => {
    const auth = TestBed.inject(AuthService);
    auth.login();
    auth.logout();
    expect(run()).toBeInstanceOf(UrlTree);
  });
});
