import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { InjectionToken, inject } from '@angular/core';
import { retry, throwError, timer } from 'rxjs';

// retry เฉพาะ GET ที่ได้ 503 สูงสุด 2 ครั้ง (รวมเป็น 3 request) POST ไม่ retry
export const MAX_GET_RETRIES = 2;
// เวลารอก่อน retry = delay x ครั้งที่ retry (300ms, 600ms) เทสต์ override เป็น 0
export const RETRY_DELAY_MS = new InjectionToken<number>('RETRY_DELAY_MS', { providedIn: 'root', factory: () => 300 });

export const retryInterceptor: HttpInterceptorFn = (req, next) => {
  if (req.method !== 'GET') return next(req);
  const delayMs = inject(RETRY_DELAY_MS);
  return next(req).pipe(
    retry({
      count: MAX_GET_RETRIES,
      delay: (err: unknown, attempt) =>
        err instanceof HttpErrorResponse && err.status === 503 ? timer(delayMs * attempt) : throwError(() => err),
    }),
  );
};
