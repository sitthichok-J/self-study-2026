import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    // วันที่ 3: ต้องลงทะเบียน HttpClient ก่อน ไม่งั้น inject(HttpClient) ใน service จะ error
    provideHttpClient(),
    provideRouter(routes)
  ]
};
