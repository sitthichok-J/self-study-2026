import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/http'
import * as api from '@/api/auth'
import { roleOf, subjectOf, useAuth } from '@/stores/auth'

vi.mock('@/api/auth')
const m = vi.mocked(api)

const b64url = (o: unknown) => btoa(JSON.stringify(o)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
const jwt = (sub: string, role?: string) => `${b64url({ alg: 'HS256', typ: 'JWT' })}.${b64url({ sub, role, exp: 9999999999 })}.sig`
const tokens = (sub = 'alice', role?: string) => ({ access_token: jwt(sub, role), token_type: 'Bearer', expires_in: 600 })

beforeEach(() => {
  vi.resetAllMocks()
  setActivePinia(createPinia())
  localStorage.clear()
  sessionStorage.clear()
})
afterEach(() => vi.useRealTimers())

describe('subjectOf', () => {
  it('อ่าน sub จาก payload', () => expect(subjectOf(jwt('alice'))).toBe('alice'))
  it('ไม่ใช่ JWT / payload พัง / sub ไม่ใช่ string = ค่าว่าง ไม่ throw', () => {
    for (const bad of ['', 'abc', 'a.b.c', 'a.!!!.c', `x.${b64url({ sub: 5 })}.y`, `x.${b64url({})}.y`]) expect(subjectOf(bad)).toBe('')
  })
})

describe('subjectOf base64url', () => {
  it('payload ที่ base64 มี - และ _ ก็อ่านได้', () => {
    let found = ''
    for (const c of ['???', '????', '?????', '>>>>', '~~~~', 'ÿÿÿ']) {
      const enc = b64url({ sub: c })
      if (enc.includes('_') || enc.includes('-')) { found = c; break }
    }
    expect(found).not.toBe('')
    expect(subjectOf(`x.${b64url({ sub: found })}.y`)).toBe(found)
  })
})

describe('auth store', () => {
  it('เริ่มต้นยังไม่ login', () => {
    const a = useAuth()
    expect(a.isAuthenticated).toBe(false)
    expect(a.token).toBeNull()
  })

  it('login ตั้ง token และชื่อ แต่ไม่เขียนลง localStorage/sessionStorage', async () => {
    m.loginRequest.mockResolvedValue(tokens('alice'))
    const a = useAuth()
    await a.login('alice', 'pw')
    expect(a.isAuthenticated).toBe(true)
    expect(a.username).toBe('alice')
    expect(m.loginRequest).toHaveBeenCalledWith('alice', 'pw')
    expect(localStorage.length).toBe(0)
    expect(sessionStorage.length).toBe(0)
  })

  it('login ผิด: ไม่มี session และ error ส่งต่อให้หน้า login', async () => {
    m.loginRequest.mockRejectedValue(new ApiError(401, 'invalid username or password'))
    const a = useAuth()
    await expect(a.login('alice', 'bad')).rejects.toMatchObject({ status: 401 })
    expect(a.isAuthenticated).toBe(false)
  })

  it('restore: มี cookie = ได้ token กลับมา', async () => {
    m.refreshRequest.mockResolvedValue(tokens('bob'))
    const a = useAuth()
    await a.restore()
    expect(a.isAuthenticated).toBe(true)
    expect(a.username).toBe('bob')
    expect(a.restored).toBe(true)
  })

  it('login สำเร็จ: restored เป็น true (ไม่ต้อง restore ซ้ำ)', async () => {
    m.loginRequest.mockResolvedValue(tokens('alice'))
    const a = useAuth()
    await a.login('alice', 'pw')
    expect(a.restored).toBe(true)
    await a.restore()
    expect(m.refreshRequest).not.toHaveBeenCalled()
  })

  it('restore: ไม่มี cookie/หมดอายุ = ยังไม่ login แต่ restored และไม่ throw', async () => {
    m.refreshRequest.mockRejectedValue(new ApiError(401, 'unauthorized'))
    const a = useAuth()
    await expect(a.restore()).resolves.toBeUndefined()
    expect(a.isAuthenticated).toBe(false)
    expect(a.restored).toBe(true)
  })

  it('restore ทำครั้งเดียวต่อการโหลดหน้า', async () => {
    m.refreshRequest.mockRejectedValue(new ApiError(401, 'unauthorized'))
    const a = useAuth()
    await a.restore()
    await a.restore()
    expect(m.refreshRequest).toHaveBeenCalledTimes(1)
  })

  it('login แล้ว restore ไม่ยิง refresh ซ้ำ', async () => {
    m.loginRequest.mockResolvedValue(tokens())
    const a = useAuth()
    await a.login('alice', 'pw')
    await a.restore()
    expect(m.refreshRequest).not.toHaveBeenCalled()
  })

  it('refresh หลายตัวพร้อมกันรวมเป็นคำขอเดียว (กัน refresh token ถูกใช้ซ้ำ)', async () => {
    let release!: () => void
    m.refreshRequest.mockImplementation(() => new Promise((res) => (release = () => res(tokens('alice')))))
    const a = useAuth()
    const all = Promise.all([a.refresh(), a.refresh(), a.refresh()])
    release()
    expect(await all).toEqual([true, true, true])
    expect(m.refreshRequest).toHaveBeenCalledTimes(1)
    // ผ่านไปแล้วขอใหม่ได้
    m.refreshRequest.mockResolvedValue(tokens('alice'))
    await a.refresh()
    expect(m.refreshRequest).toHaveBeenCalledTimes(2)
  })

  it('refresh ล้มเหลว: ล้าง token และคืน false', async () => {
    m.loginRequest.mockResolvedValue(tokens())
    m.refreshRequest.mockRejectedValue(new ApiError(401, 'unauthorized'))
    const a = useAuth()
    await a.login('alice', 'pw')
    expect(await a.refresh()).toBe(false)
    expect(a.isAuthenticated).toBe(false)
    expect(a.username).toBe('')
  })

  it('refresh ที่ล้มเหลวไม่ทำให้ค้างอยู่ในสถานะ inflight', async () => {
    m.refreshRequest.mockRejectedValueOnce(new ApiError(0, 'down')).mockResolvedValueOnce(tokens('alice'))
    const a = useAuth()
    expect(await a.refresh()).toBe(false)
    expect(await a.refresh()).toBe(true)
  })

  it('logout ล้าง session และเรียก API', async () => {
    m.loginRequest.mockResolvedValue(tokens())
    m.logoutRequest.mockResolvedValue()
    const a = useAuth()
    await a.login('alice', 'pw')
    await a.logout()
    expect(m.logoutRequest).toHaveBeenCalledTimes(1)
    expect(a.isAuthenticated).toBe(false)
    expect(a.username).toBe('')
  })

  it('logout ตอนเครือข่ายล่ม: ยังล้างฝั่งเครื่องนี้และไม่ throw', async () => {
    m.loginRequest.mockResolvedValue(tokens())
    m.logoutRequest.mockRejectedValue(new ApiError(0, 'down'))
    const a = useAuth()
    await a.login('alice', 'pw')
    await expect(a.logout()).resolves.toBeUndefined()
    expect(a.isAuthenticated).toBe(false)
  })
})

describe('role (แสดงผลเท่านั้น)', () => {
  it('roleOf อ่าน role จาก payload; ไม่มี/ไม่ใช่ string/พัง = ค่าว่าง', () => {
    expect(roleOf(jwt('ops', 'staff'))).toBe('staff')
    expect(roleOf(jwt('alice'))).toBe('')
    for (const bad of ['', 'abc', 'a.!!!.c', `x.${b64url({ role: 7 })}.y`]) expect(roleOf(bad)).toBe('')
  })

  it('login เป็น staff -> isStaff; เป็น user -> ไม่ใช่; clear ล้างค่า', async () => {
    const a = useAuth()
    m.loginRequest.mockResolvedValueOnce(tokens('ops', 'staff'))
    await a.login('ops', 'x')
    expect(a.isStaff).toBe(true)
    expect(a.role).toBe('staff')
    a.clear()
    expect(a.isStaff).toBe(false)
    expect(a.role).toBe('')
    m.loginRequest.mockResolvedValueOnce(tokens('alice', 'user'))
    await a.login('alice', 'x')
    expect(a.isStaff).toBe(false)
  })

  it('role ใน payload ที่มีค่าแปลก (STAFF, "staff ") ไม่นับเป็น staff', async () => {
    const a = useAuth()
    for (const r of ['STAFF', 'staff ', 'admin']) {
      m.loginRequest.mockResolvedValueOnce(tokens('x', r))
      await a.login('x', 'y')
      expect(a.isStaff).toBe(false)
    }
  })

  it('refresh ที่ได้ token ใหม่ อัปเดต role ตามด้วย', async () => {
    const a = useAuth()
    m.refreshRequest.mockResolvedValueOnce(tokens('ops', 'staff'))
    await a.refresh()
    expect(a.isStaff).toBe(true)
  })

  it('logout ส่ง access token ปัจจุบันไปให้ API เพิกถอน แล้วล้าง session', async () => {
    const a = useAuth()
    const t = tokens('ops', 'staff')
    m.loginRequest.mockResolvedValueOnce(t)
    m.logoutRequest.mockResolvedValue()
    await a.login('ops', 'x')
    await a.logout()
    expect(m.logoutRequest).toHaveBeenCalledWith(t.access_token)
    expect(a.token).toBeNull()
    expect(a.isStaff).toBe(false)
  })
})
