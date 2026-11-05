import { afterEach, describe, expect, it, vi } from 'vitest'
import { bahtToSatang, formatBaht, MAX_SATANG, newIdempotencyKey, validateClaimForm } from '@/lib'

describe('bahtToSatang', () => {
  it.each([
    ['1', 100],
    ['1500', 150000],
    ['1500.5', 150050],
    ['1500.50', 150050],
    ['0.29', 29], // ถ้าใช้ float จะได้ 28.999...
    ['0.01', 1],
    ['19.99', 1999],
    [' 20 ', 2000],
  ])('%s -> %s', (input, want) => {
    expect(bahtToSatang(input)).toBe(want)
  })

  it.each(['', 'abc', '-1', '1.234', '1,000', '1e3', '.5', '5.', '1 000', '０'])('ปฏิเสธ %j', (input) => {
    expect(bahtToSatang(input)).toBeNull()
  })

  it('ปฏิเสธเลขที่ใหญ่จน integer ไม่ปลอดภัย', () => {
    expect(bahtToSatang('99999999999')).toBeNull() // เกิน 10 หลัก
  })
})

describe('validateClaimForm', () => {
  it('ข้อมูลถูก ได้ satang และไม่มี error', () => {
    expect(validateClaimForm('FAKE-AB12', '1500.50')).toEqual({ errors: {}, satang: 150050 })
  })

  it.each(['', 'FAKE-', 'FAKE-AB', 'fake-ab12', 'FAKE-ab12', 'REAL-AB12', 'FAKE-ABCDEFGHIJKLM', 'FAKE-AB 12'])(
    'patient_ref %j ไม่ผ่าน',
    (ref) => {
      const r = validateClaimForm(ref, '10')
      expect(r.errors.patient_ref).toBeTruthy()
      expect(r.satang).toBeNull()
    },
  )

  it('จำนวนเงิน 0 ไม่ผ่าน', () => {
    expect(validateClaimForm('FAKE-AB12', '0').errors.amount).toContain('มากกว่า 0')
  })

  it('เกินเพดานของ backend ไม่ผ่าน และเท่าเพดานผ่าน', () => {
    expect(validateClaimForm('FAKE-AB12', '10000000.01').errors.amount).toBeTruthy()
    expect(validateClaimForm('FAKE-AB12', '10000000').satang).toBe(MAX_SATANG)
  })

  it('รายงานทั้งสองช่องพร้อมกัน', () => {
    const r = validateClaimForm('x', 'y')
    expect(Object.keys(r.errors).sort()).toEqual(['amount', 'patient_ref'])
  })
})

describe('formatBaht', () => {
  it.each([
    [0, '0.00'],
    [5, '0.05'],
    [100, '1.00'],
    [150050, '1,500.50'],
    [500000000, '5,000,000.00'],
    [1_000_000_000, '10,000,000.00'],
  ])('%s -> %s', (s, want) => {
    expect(formatBaht(s)).toBe(want)
  })
})

describe('newIdempotencyKey', () => {
  const uuidRe = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/
  afterEach(() => vi.unstubAllGlobals())

  it('ใช้ randomUUID เมื่อมี', () => {
    vi.stubGlobal('crypto', { randomUUID: () => 'fixed-uuid' })
    expect(newIdempotencyKey()).toBe('fixed-uuid')
  })

  it('ไม่มี randomUUID (http ที่ไม่ใช่ localhost) ยังได้ UUID v4 ที่ถูกรูป', () => {
    vi.stubGlobal('crypto', { getRandomValues: (a: Uint8Array) => a.fill(0xab) })
    expect(newIdempotencyKey()).toMatch(uuidRe)
  })

  it('ไม่มี crypto เลยก็ยังทำงาน และไม่ซ้ำกัน', () => {
    vi.stubGlobal('crypto', undefined)
    const a = newIdempotencyKey()
    const b = newIdempotencyKey()
    expect(a).toMatch(uuidRe)
    expect(a).not.toBe(b)
  })
})
