import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from './auth.service';

// functional guard: ยังไม่ล็อกอินให้เด้งกลับ /claims (คืน UrlTree แทนการสั่ง navigate เอง)
export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  return auth.loggedIn() ? true : router.createUrlTree(['/claims'], { queryParams: { login: 'required' } });
};
