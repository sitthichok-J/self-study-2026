export class ApiError extends Error {
  constructor(
    public status: number, // 0 = ติดต่อเซิร์ฟเวอร์ไม่ได้
    message: string,
    public retryAfter?: number, // วินาที ใช้กับ 429
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export const API_BASE: string = (import.meta.env?.VITE_API_BASE as string | undefined) ?? '/api'

export interface HttpResult<T> {
  data: T
  status: number
  headers: Headers
}

// Day 11: http() ไม่รู้จัก store โดยตรง (กัน import วนกัน) ให้ main.ts ผูกมือเข้ามาตอนเริ่มแอป
export interface AuthHooks {
  getToken: () => string | null // access token ที่อยู่ในหน่วยความจำ (ไม่เก็บใน localStorage)
  refresh: () => Promise<boolean> // ขอ token ใหม่ด้วย refresh cookie; true = สำเร็จ
  onLoggedOut: () => void // refresh ไม่สำเร็จ: ถือว่า session จบแล้ว
}

let hooks: AuthHooks | null = null
export function configureAuth(h: AuthHooks | null): void {
  hooks = h
}

// เส้นทาง /auth/* ใช้ cookie ไม่ใช่ Bearer และห้ามวนลองใหม่เองเมื่อได้ 401
export const AUTH_PREFIX = '/auth/'
// ต้องตรงกับ CSRFHeader/CSRFValue ฝั่ง Go (internal/adapter/http/authhandler.go)
export const CSRF_HEADER = 'X-Requested-With'
export const CSRF_VALUE = 'claims-web'

async function toApiError(res: Response): Promise<ApiError> {
  let msg = ''
  try {
    const body = (await res.json()) as { error?: string; message?: string }
    msg = body.error ?? body.message ?? ''
  } catch {
    /* ตัวตอบไม่ใช่ JSON */
  }
  const ra = Number(res.headers.get('Retry-After'))
  return new ApiError(res.status, msg || `HTTP ${res.status}`, Number.isFinite(ra) && ra > 0 ? ra : undefined)
}

async function send(path: string, init: RequestInit): Promise<Response> {
  try {
    return await fetch(`${API_BASE}${path}`, init)
  } catch {
    throw new ApiError(0, 'ติดต่อเซิร์ฟเวอร์ไม่ได้')
  }
}

export async function http<T>(path: string, init: RequestInit = {}, retried = false): Promise<HttpResult<T>> {
  const token = hooks?.getToken() ?? null
  const res = await send(path, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init.headers as Record<string, string> | undefined),
    },
  })

  // access token หมดอายุ: ขอใหม่แล้วลองซ้ำ "ครั้งเดียว" (POST ซ้ำได้ปลอดภัยเพราะ 401 มาก่อนทำงานจริง
  // และ Idempotency-Key ใน headers ยังเป็นค่าเดิม)
  if (res.status === 401 && hooks && !retried && !path.startsWith(AUTH_PREFIX)) {
    if (await hooks.refresh()) return http<T>(path, init, true)
    hooks.onLoggedOut()
  }

  if (!res.ok) throw await toApiError(res)

  const data = (await res.json()) as T
  return { data, status: res.status, headers: res.headers }
}

// เรียก /auth/*: ส่ง cookie ไปด้วย (credentials) + header กัน CSRF, ไม่แนบ Bearer
// bearer (ถ้ามี) ใช้กับ logout เท่านั้น: server จะขึ้นบัญชีเพิกถอน access token นั้นทันที
export async function authRequest<T>(path: string, body?: unknown, bearer?: string | null): Promise<T | null> {
  const res = await send(path, {
    method: 'POST',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      [CSRF_HEADER]: CSRF_VALUE,
      ...(bearer ? { Authorization: `Bearer ${bearer}` } : {}),
      ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (!res.ok) throw await toApiError(res)
  if (res.status === 204) return null
  return (await res.json()) as T
}

export function describeError(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.status === 0) return 'ติดต่อเซิร์ฟเวอร์ไม่ได้ ลองใหม่อีกครั้ง'
    if (e.status === 401) return 'เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่'
    if (e.status === 429) return `ส่งคำขอถี่เกินไป${e.retryAfter ? ` รอ ${e.retryAfter} วินาทีแล้วลองใหม่` : ' รอสักครู่แล้วลองใหม่'}`
    if (e.status === 404) return 'ไม่พบข้อมูล'
    if (e.status === 409) return 'คำขอซ้ำกับรายการเดิมแต่ข้อมูลไม่ตรงกัน'
    if (e.status >= 500) return 'ระบบขัดข้อง ลองใหม่อีกครั้ง'
    return e.message
  }
  return 'เกิดข้อผิดพลาดที่ไม่คาดคิด'
}
