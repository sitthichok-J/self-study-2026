import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideRouter, withComponentInputBinding } from '@angular/router';
import { routes } from './app.routes';
import { errorInterceptor } from './error.interceptor';
import { idempotencyInterceptor } from './idempotency.interceptor';
import { retryInterceptor } from './retry.interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    // วันที่ 3: ต้องลงทะเบียน HttpClient ก่อน ไม่งั้น inject(HttpClient) ใน service จะ error
    // วันที่ 6: ลำดับใน array = ลำดับตอนส่ง request (ตัวแรกอยู่นอกสุด) retry อยู่ในสุดจึงเห็น HttpErrorResponse ดิบ
    // error อยู่นอกกว่าจึงแปลงเป็น ApiError หลัง retry หมดแล้ว (ยืนยันด้วยเทสต์ interceptor-order กับ app.config)
    provideHttpClient(withInterceptors([idempotencyInterceptor, errorInterceptor, retryInterceptor])),
    // withComponentInputBinding: ค่า :id ใน URL ถูกส่งเข้า input() ชื่อ id ของ component โดยตรง
    provideRouter(routes, withComponentInputBinding())
  ]
};
