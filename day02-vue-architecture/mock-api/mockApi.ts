// Mock API สำหรับ dev (แทน json-server): เก็บข้อมูลใน memory ข้อมูลสมมติทั้งหมด
import type { Claim, ClaimStatus, NewClaim } from '../src/types/claim.ts'

export interface MockResult {
  status: number
  body: unknown
}

export const SEED: Claim[] = [
  { id: 1, patientRef: 'P-1001', hospital: 'รพ.ตัวอย่าง A', amount: 12500, status: 'PENDING' },
  { id: 2, patientRef: 'P-1002', hospital: 'รพ.ตัวอย่าง B', amount: 8200.5, status: 'APPROVED' },
  { id: 3, patientRef: 'P-1003', hospital: 'รพ.ตัวอย่าง A', amount: 31000, status: 'REJECTED' },
  { id: 4, patientRef: 'P-1004', hospital: 'รพ.ตัวอย่าง C', amount: 4300, status: 'PENDING' },
]

export const DEMO_USER = { username: 'demo', password: 'demo1234' }
export const DEMO_TOKEN = 'demo-token-not-a-real-secret'

const STATUSES: ClaimStatus[] = ['PENDING', 'APPROVED', 'REJECTED']

export function createMockApi(seed: Claim[] = SEED) {
  let claims: Claim[] = seed.map((c) => ({ ...c }))
  let nextId = Math.max(0, ...claims.map((c) => c.id)) + 1

  function handle(method: string, url: string, authHeader: string | undefined, body: unknown): MockResult {
    const path = url.split('?')[0].replace(/\/+$/, '')

    if (method === 'POST' && path === '/api/login') {
      const b = (body ?? {}) as { username?: string; password?: string }
      if (b.username === DEMO_USER.username && b.password === DEMO_USER.password) {
        return { status: 200, body: { token: DEMO_TOKEN } }
      }
      return { status: 401, body: { message: 'ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง' } }
    }

    if (authHeader !== `Bearer ${DEMO_TOKEN}`) {
      return { status: 401, body: { message: 'ต้องเข้าสู่ระบบ' } }
    }

    if (path === '/api/claims') {
      if (method === 'GET') return { status: 200, body: claims }
      if (method === 'POST') {
        const b = (body ?? {}) as Partial<NewClaim>
        if (!b.patientRef || !b.hospital || !(Number(b.amount) > 0)) {
          return { status: 422, body: { message: 'ข้อมูลไม่ครบ' } }
        }
        const claim: Claim = {
          id: nextId++,
          patientRef: b.patientRef,
          hospital: b.hospital,
          amount: Number(b.amount),
          status: 'PENDING',
        }
        claims = [claim, ...claims]
        return { status: 201, body: claim }
      }
    }

    const m = path.match(/^\/api\/claims\/(\d+)$/)
    if (m) {
      const id = Number(m[1])
      const claim = claims.find((c) => c.id === id)
      if (!claim) return { status: 404, body: { message: 'ไม่พบเคลม' } }
      if (method === 'GET') return { status: 200, body: claim }
      if (method === 'PATCH') {
        const status = (body as { status?: ClaimStatus } | undefined)?.status
        if (!status || !STATUSES.includes(status)) return { status: 422, body: { message: 'สถานะไม่ถูกต้อง' } }
        claim.status = status
        return { status: 200, body: claim }
      }
      if (method === 'DELETE') {
        claims = claims.filter((c) => c.id !== id)
        return { status: 204, body: null }
      }
    }

    return { status: 404, body: { message: 'ไม่พบ endpoint' } }
  }

  return { handle }
}
