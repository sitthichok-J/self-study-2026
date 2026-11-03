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

export async function http<T>(path: string, init: RequestInit = {}): Promise<HttpResult<T>> {
  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}`, {
      ...init,
      headers: { Accept: 'application/json', ...(init.headers as Record<string, string> | undefined) },
    })
  } catch {
    throw new ApiError(0, 'ติดต่อเซิร์ฟเวอร์ไม่ได้')
  }

  if (!res.ok) {
    let msg = ''
    try {
      const body = (await res.json()) as { error?: string; message?: string }
      msg = body.error ?? body.message ?? ''
    } catch {
      /* ตัวตอบไม่ใช่ JSON */
    }
    const ra = Number(res.headers.get('Retry-After'))
    throw new ApiError(res.status, msg || `HTTP ${res.status}`, Number.isFinite(ra) && ra > 0 ? ra : undefined)
  }

  const data = (await res.json()) as T
  return { data, status: res.status, headers: res.headers }
}

export function describeError(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.status === 0) return 'ติดต่อเซิร์ฟเวอร์ไม่ได้ ลองใหม่อีกครั้ง'
    if (e.status === 429) return `ส่งคำขอถี่เกินไป${e.retryAfter ? ` รอ ${e.retryAfter} วินาทีแล้วลองใหม่` : ' รอสักครู่แล้วลองใหม่'}`
    if (e.status === 404) return 'ไม่พบข้อมูล'
    if (e.status === 409) return 'คำขอซ้ำกับรายการเดิมแต่ข้อมูลไม่ตรงกัน'
    if (e.status >= 500) return 'ระบบขัดข้อง ลองใหม่อีกครั้ง'
    return e.message
  }
  return 'เกิดข้อผิดพลาดที่ไม่คาดคิด'
}
