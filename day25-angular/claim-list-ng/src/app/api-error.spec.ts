import { HttpErrorResponse } from '@angular/common/http';
import { ApiError, readApiError, toApiError } from './api-error';

const http = (status: number, error: unknown) => new HttpErrorResponse({ status, error });

describe('api-error', () => {
  it('toApiError เก็บ status กับข้อความ error จาก body', () => {
    const e = toApiError(http(409, { error: 'key reused' }));
    expect(e).toBeInstanceOf(ApiError);
    expect(e.status).toBe(409);
    expect(e.serverMessage).toBe('key reused');
  });

  it('body ที่ไม่ใช่รูป {error: string} ได้ serverMessage เป็น null', () => {
    expect(toApiError(http(500, 'plain text')).serverMessage).toBeNull();
    expect(toApiError(http(500, { error: 42 })).serverMessage).toBeNull();
    expect(toApiError(http(500, null)).serverMessage).toBeNull();
  });

  it('status 0 (ต่อ backend ไม่ได้) เก็บเป็น 0', () => {
    expect(toApiError(http(0, null)).status).toBe(0);
  });

  it('readApiError อ่านได้ทั้ง ApiError และ HttpErrorResponse แต่ error อื่นได้ null', () => {
    expect(readApiError(new ApiError(404, 'nf'))).toEqual({ status: 404, serverMessage: 'nf' });
    expect(readApiError(http(400, { error: 'bad' }))).toEqual({ status: 400, serverMessage: 'bad' });
    expect(readApiError(new Error('x'))).toBeNull();
    expect(readApiError('x')).toBeNull();
  });
});
