import { afterEach, describe, expect, it, vi } from 'vitest'
import { createClaim, getClaim, listClaims } from '@/api/claims'
import { ApiError, describeError } from '@/api/http'
import { sampleClaim, stubFetch } from './helpers'

afterEach(() => vi.unstubAllGlobals())

const urlOf = (fn: ReturnType<typeof vi.fn>, i = 0) => fn.mock.calls[i]![0] as string
const initOf = (fn: ReturnType<typeof vi.fn>, i = 0) => fn.mock.calls[i]![1] as RequestInit

describe('listClaims', () => {
  it('ไม่ส่ง query ถ้าไม่มี param', async () => {
    const f = stubFetch({ body: { items: [], limit: 20, offset: 0 } })
    await listClaims()
    expect(urlOf(f)).toBe('/api/claims')
  })

  it('ใส่ status limit offset ใน query และไม่ใส่ status ว่าง', async () => {
    const f = stubFetch({ body: { items: [], limit: 11, offset: 20 } }, { body: { items: [], limit: 11, offset: 0 } })
    await listClaims({ status: 'APPROVED', limit: 11, offset: 20 })
    await listClaims({ status: '', limit: 11, offset: 0 })
    expect(urlOf(f, 0)).toBe('/api/claims?status=APPROVED&limit=11&offset=20')
    expect(urlOf(f, 1)).toBe('/api/claims?limit=11&offset=0')
  })
})

describe('getClaim', () => {
  it('encode id ใน path', async () => {
    const f = stubFetch({ body: sampleClaim() })
    await getClaim('a/b')
    expect(urlOf(f)).toBe('/api/claims/a%2Fb')
  })

  it('404 เป็น ApiError status 404 พร้อมข้อความจาก backend', async () => {
    stubFetch({ status: 404, body: { error: 'claim not found' } })
    await expect(getClaim('x')).rejects.toMatchObject({ status: 404, message: 'claim not found' })
  })
})

describe('createClaim', () => {
  it('ส่ง Idempotency-Key, Content-Type และ body เป็นสตางค์', async () => {
    const f = stubFetch({ status: 201, body: sampleClaim() })
    const r = await createClaim({ patient_ref: 'FAKE-AB12', amount: 150000 }, 'key-1')
    const init = initOf(f)
    expect(init.method).toBe('POST')
    expect(init.headers).toMatchObject({ 'Idempotency-Key': 'key-1', 'Content-Type': 'application/json' })
    expect(JSON.parse(init.body as string)).toEqual({ patient_ref: 'FAKE-AB12', amount: 150000 })
    expect(r.replayed).toBe(false)
  })

  it('อ่าน Idempotent-Replayed: true เป็น replayed', async () => {
    stubFetch({ status: 200, body: sampleClaim(), headers: { 'Idempotent-Replayed': 'true' } })
    expect((await createClaim({ patient_ref: 'FAKE-AB12', amount: 1 }, 'k')).replayed).toBe(true)
  })

  it('409 key ซ้ำแต่ body ต่าง', async () => {
    stubFetch({ status: 409, body: { error: 'idempotency key reused with a different request' } })
    const e = await createClaim({ patient_ref: 'FAKE-AB12', amount: 1 }, 'k').catch((x) => x)
    expect(e).toBeInstanceOf(ApiError)
    expect(describeError(e)).toContain('ไม่ตรงกัน')
  })
})

describe('http errors', () => {
  it('fetch ล้ม (เน็ตหลุด/gateway ล่ม) เป็น status 0', async () => {
    stubFetch(new TypeError('Failed to fetch'))
    const e = await getClaim('x').catch((x) => x)
    expect(e).toMatchObject({ status: 0 })
    expect(describeError(e)).toContain('ติดต่อเซิร์ฟเวอร์ไม่ได้')
  })

  it('429 จาก gateway (Kong ใช้ field message) พร้อม Retry-After', async () => {
    stubFetch({ status: 429, body: { message: 'API rate limit exceeded' }, headers: { 'Retry-After': '12' } })
    const e = await getClaim('x').catch((x) => x)
    expect(e).toMatchObject({ status: 429, retryAfter: 12, message: 'API rate limit exceeded' })
    expect(describeError(e)).toContain('12 วินาที')
  })

  it.each([['0'], ['-3'], ['abc'], ['']])('Retry-After %j ใช้ไม่ได้ -> ไม่มี retryAfter', async (h) => {
    stubFetch({ status: 429, body: { message: 'x' }, headers: { 'Retry-After': h } })
    const e = await getClaim('x').catch((x) => x)
    expect(e).toMatchObject({ status: 429 })
    expect(e.retryAfter).toBeUndefined()
  })

  it('429 ที่ไม่มี Retry-After ก็ยังมีข้อความ', async () => {
    stubFetch({ status: 429, body: {} })
    expect(describeError(await getClaim('x').catch((x) => x))).toContain('ถี่เกินไป')
  })

  it('ตัวตอบที่ไม่ใช่ JSON (เช่นหน้า error ของ proxy) ไม่ทำให้พัง', async () => {
    stubFetch({ status: 502, rawText: '<html>Bad Gateway</html>' })
    const e = await getClaim('x').catch((x) => x)
    expect(e).toMatchObject({ status: 502, message: 'HTTP 502' })
    expect(describeError(e)).toContain('ระบบขัดข้อง')
  })

  it('error ที่ไม่ใช่ ApiError ไม่รั่วรายละเอียด', () => {
    expect(describeError(new Error('secret stack'))).not.toContain('secret')
  })
})
