import { HttpErrorResponse } from '@angular/common/http';

// error ที่ errorInterceptor แปลงมาจาก HttpErrorResponse เก็บแค่สิ่งที่หน้า UI ต้องใช้
// status: รหัส HTTP (0 = ต่อ backend ไม่ได้) / serverMessage: ข้อความใน body ที่ backend ส่งมา ({"error": "..."})
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly serverMessage: string | null,
  ) {
    super(serverMessage ?? `HTTP ${status}`);
    this.name = 'ApiError';
  }
}

export function toApiError(err: HttpErrorResponse): ApiError {
  const body = err.error as { error?: unknown } | null;
  const serverMessage = body && typeof body.error === 'string' ? body.error : null;
  return new ApiError(err.status, serverMessage);
}

// ให้ component อ่าน error ได้ทั้งสองแบบ: ApiError (ผ่าน interceptor) และ HttpErrorResponse (เทสต์ที่ไม่ได้ลง interceptor)
export function readApiError(err: unknown): { status: number; serverMessage: string | null } | null {
  if (err instanceof ApiError) return { status: err.status, serverMessage: err.serverMessage };
  if (err instanceof HttpErrorResponse) {
    const { status, serverMessage } = toApiError(err);
    return { status, serverMessage };
  }
  return null;
}
