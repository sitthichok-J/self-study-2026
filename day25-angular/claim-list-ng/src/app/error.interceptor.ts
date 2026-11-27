import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';
import { toApiError } from './api-error';

// แปลง HttpErrorResponse เป็น ApiError ที่เดียว ส่วน error ชนิดอื่นปล่อยผ่าน
export const errorInterceptor: HttpInterceptorFn = (req, next) =>
  next(req).pipe(
    catchError((err: unknown) => throwError(() => (err instanceof HttpErrorResponse ? toApiError(err) : err))),
  );
