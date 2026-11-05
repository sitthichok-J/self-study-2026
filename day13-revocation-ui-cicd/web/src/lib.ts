// กฎตรวจข้อมูลฝั่ง client ต้องตรงกับ backend (Day 8) แต่ backend ยังเป็นตัวตัดสินสุดท้าย
export const PATIENT_REF_RE = /^FAKE-[A-Z0-9]{3,12}$/
export const MAX_SATANG = 1_000_000_000

export interface FormErrors {
  patient_ref?: string
  amount?: string
}

// แปลงบาท (สตริง) เป็นสตางค์ด้วยการแยกสตริง ไม่ใช้ float เพื่อไม่ให้ 0.29 กลายเป็น 28.999...
export function bahtToSatang(input: string): number | null {
  const s = input.trim()
  const m = /^(\d{1,10})(?:\.(\d{1,2}))?$/.exec(s)
  if (!m) return null
  const baht = Number(m[1])
  const frac = Number((m[2] ?? '').padEnd(2, '0'))
  const satang = baht * 100 + frac
  return Number.isSafeInteger(satang) ? satang : null
}

export function validateClaimForm(patientRef: string, amountBaht: string): { errors: FormErrors; satang: number | null } {
  const errors: FormErrors = {}
  if (!PATIENT_REF_RE.test(patientRef.trim())) {
    errors.patient_ref = 'ต้องอยู่ในรูป FAKE- ตามด้วยตัวอักษรพิมพ์ใหญ่หรือตัวเลข 3-12 ตัว'
  }
  const satang = bahtToSatang(amountBaht)
  if (satang === null) {
    errors.amount = 'กรอกเป็นบาท ทศนิยมไม่เกิน 2 ตำแหน่ง'
  } else if (satang <= 0) {
    errors.amount = 'จำนวนเงินต้องมากกว่า 0'
  } else if (satang > MAX_SATANG) {
    errors.amount = 'จำนวนเงินเกินที่ระบบรับได้'
  }
  return { errors, satang: Object.keys(errors).length ? null : satang }
}

export function formatBaht(satang: number): string {
  const whole = Math.floor(satang / 100)
  const frac = String(satang % 100).padStart(2, '0')
  return `${String(whole).replace(/\B(?=(\d{3})+(?!\d))/g, ',')}.${frac}`
}

// crypto.randomUUID ใช้ได้เฉพาะ secure context (https หรือ localhost)
// เปิดผ่าน http://<ip ในวง LAN> จะไม่มี จึงต้องมีทางสำรอง
export function newIdempotencyKey(): string {
  const c = globalThis.crypto
  if (c && typeof c.randomUUID === 'function') return c.randomUUID()
  const b = new Uint8Array(16)
  if (c && typeof c.getRandomValues === 'function') c.getRandomValues(b)
  else for (let i = 0; i < 16; i++) b[i] = Math.floor(Math.random() * 256)
  b[6] = (b[6]! & 0x0f) | 0x40
  b[8] = (b[8]! & 0x3f) | 0x80
  const h = [...b].map((x) => x.toString(16).padStart(2, '0')).join('')
  return `${h.slice(0, 8)}-${h.slice(8, 12)}-${h.slice(12, 16)}-${h.slice(16, 20)}-${h.slice(20)}`
}

// ปลายทางหลัง login ต้องเป็น path ภายในแอปเท่านั้น (กัน open redirect เช่น //evil.com หรือ https://evil.com)
export function safeRedirect(target: unknown): string {
  if (typeof target !== 'string' || !target.startsWith('/')) return '/'
  if (target.startsWith('//') || target.includes('\\') || /[\u0000-\u001f]/.test(target)) return '/'
  if (target === '/login' || target.startsWith('/login?')) return '/'
  return target
}

// 429 จาก Kong = ลอง login ถี่เกินเพดานต่อ IP; Retry-After (วินาที) อ่านได้เพราะ Kong เปิดให้ผ่าน CORS
export function loginLimitMessage(retryAfter?: number): string {
  if (retryAfter && retryAfter > 0) return `ลองเข้าสู่ระบบถี่เกินไป รออีก ${Math.ceil(retryAfter)} วินาทีแล้วลองใหม่`
  return 'ลองเข้าสู่ระบบถี่เกินไป รอสักครู่แล้วลองใหม่'
}
