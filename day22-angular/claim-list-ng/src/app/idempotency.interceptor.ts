import { HttpInterceptorFn } from '@angular/common/http';

export const IDEMPOTENCY_HEADER = 'Idempotency-Key';

// POST ที่ยังไม่มี Idempotency-Key จะถูกเติมให้ (backend ตอบ 400 ถ้าไม่มี)
// ถ้าผู้เรียกใส่มาแล้ว (เช่นฟอร์มที่ถือ key เดียวต่อการส่ง) จะไม่แตะ ไม่งั้นการ retry จะได้ key ใหม่และสร้างซ้ำ
export const idempotencyInterceptor: HttpInterceptorFn = (req, next) => {
  if (req.method !== 'POST' || req.headers.has(IDEMPOTENCY_HEADER)) return next(req);
  return next(req.clone({ setHeaders: { [IDEMPOTENCY_HEADER]: crypto.randomUUID() } }));
};
