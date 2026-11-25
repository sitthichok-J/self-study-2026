import { HttpClient, HttpInterceptorFn, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { tap } from 'rxjs';

// ทดลองเอง: ลำดับใน withInterceptors([...]) ทำงานอย่างไร (ผลนี้คือที่มาของคอมเมนต์ใน app.config.ts)
describe('ลำดับ interceptor ใน withInterceptors', () => {
  it('ตัวแรกใน array เห็น request ก่อน และเห็น response หลังสุด', () => {
    const log: string[] = [];
    const named = (name: string): HttpInterceptorFn => (req, next) => {
      log.push(`${name}:req`);
      return next(req).pipe(tap({ next: (e) => e.type === 4 && log.push(`${name}:res`) }));
    };
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([named('A'), named('B')])), provideHttpClientTesting()],
    });
    TestBed.inject(HttpClient).get('/x').subscribe();
    TestBed.inject(HttpTestingController).expectOne('/x').flush({});
    expect(log).toEqual(['A:req', 'B:req', 'B:res', 'A:res']);
  });
});
