// ชั้น HTTP ชั้นเดียว: ใส่ token, แปลง error เป็น ApiError ให้ component ไม่ต้องรู้เรื่อง fetch
export const TOKEN_KEY = 'day02-token'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// ให้ชั้นบนกำหนดว่าเจอ 401 แล้วทำอะไร (ไม่ import store/router ที่นี่ เพื่อไม่ให้ชั้น API ผูกกับ UI)
let unauthorizedHandler: (() => void) | null = null
export function setUnauthorizedHandler(fn: (() => void) | null) {
  unauthorizedHandler = fn
}

export async function http<T>(url: string, init: { method?: string; body?: unknown } = {}): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  let token: string | null = null
  try {
    token = localStorage.getItem(TOKEN_KEY)
  } catch { /* ignore */ }
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(url, {
    method: init.method ?? 'GET',
    headers,
    body: init.body === undefined ? undefined : JSON.stringify(init.body),
  })
  if (!res.ok) {
    // 401 นอกหน้า login = token หมดอายุ/ไม่ถูกต้อง (401 ของ /api/login คือรหัสผิด ให้หน้า login แสดง error เอง)
    if (res.status === 401 && url !== '/api/login') unauthorizedHandler?.()
    let message = `HTTP ${res.status}`
    try {
      const data = (await res.json()) as { message?: string }
      if (data.message) message = data.message
    } catch { /* body ว่าง/ไม่ใช่ JSON */ }
    throw new ApiError(res.status, message)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}
