import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { authApi, claimsApi } from '../src/api/claimsApi'
import { TOKEN_KEY } from '../src/api/http'
import { useAuthStore } from '../src/stores/auth'
import { useClaimsStore } from '../src/stores/claims'
import type { Claim } from '../src/types/claim'

vi.mock('../src/api/claimsApi', () => ({
  claimsApi: { list: vi.fn(), get: vi.fn(), create: vi.fn(), updateStatus: vi.fn(), remove: vi.fn() },
  authApi: { login: vi.fn() },
}))

const data: Claim[] = [
  { id: 1, patientRef: 'P-1', hospital: 'รพ.A', amount: 100, status: 'PENDING' },
  { id: 2, patientRef: 'P-2', hospital: 'รพ.B', amount: 200, status: 'APPROVED' },
  { id: 3, patientRef: 'P-3', hospital: 'รพ.A', amount: 300, status: 'REJECTED' },
]

beforeEach(() => {
  setActivePinia(createPinia())
  localStorage.clear()
  vi.resetAllMocks()
})

describe('claims store', () => {
  it('load สำเร็จ/ล้มเหลว', async () => {
    const s = useClaimsStore()
    vi.mocked(claimsApi.list).mockResolvedValue(data)
    await s.load()
    expect(s.items).toHaveLength(3)
    expect(s.loading).toBe(false)

    vi.mocked(claimsApi.list).mockRejectedValue(new Error('เน็ตหลุด'))
    await s.load()
    expect(s.error).toBe('เน็ตหลุด')
  })

  it('getters: filtered + counts', async () => {
    const s = useClaimsStore()
    vi.mocked(claimsApi.list).mockResolvedValue(data)
    await s.load()
    s.keyword = 'รพ.a'
    expect(s.filtered.map((c) => c.id)).toEqual([1, 3])
    s.statusFilter = 'REJECTED'
    expect(s.filtered.map((c) => c.id)).toEqual([3])
    expect(s.counts).toEqual({ ALL: 3, PENDING: 1, APPROVED: 1, REJECTED: 1 })
  })

  it('add / changeStatus / remove อัปเดต state', async () => {
    const s = useClaimsStore()
    s.items = [...data]
    const created: Claim = { id: 4, patientRef: 'P-4', hospital: 'X', amount: 5, status: 'PENDING' }
    vi.mocked(claimsApi.create).mockResolvedValue(created)
    await s.add({ patientRef: 'P-4', hospital: 'X', amount: 5 })
    expect(s.items[0].id).toBe(4)

    vi.mocked(claimsApi.updateStatus).mockResolvedValue({ ...data[0], status: 'APPROVED' })
    await s.changeStatus(1, 'APPROVED')
    expect(s.byId(1)?.status).toBe('APPROVED')

    vi.mocked(claimsApi.remove).mockResolvedValue(undefined)
    await s.remove(2)
    expect(s.byId(2)).toBeUndefined()
  })

  it('add ล้มเหลว -> throw และ state ไม่เปลี่ยน', async () => {
    const s = useClaimsStore()
    vi.mocked(claimsApi.create).mockRejectedValue(new Error('422'))
    await expect(s.add({ patientRef: 'a', hospital: 'b', amount: 1 })).rejects.toThrow('422')
    expect(s.items).toHaveLength(0)
  })
})

describe('auth store', () => {
  it('login เก็บ token ลง state + localStorage, logout ล้าง', async () => {
    const a = useAuthStore()
    expect(a.isLoggedIn).toBe(false)
    vi.mocked(authApi.login).mockResolvedValue({ token: 't1' })
    await a.login('u', 'p')
    expect(a.isLoggedIn).toBe(true)
    expect(localStorage.getItem(TOKEN_KEY)).toBe('t1')
    a.logout()
    expect(a.isLoggedIn).toBe(false)
    expect(localStorage.getItem(TOKEN_KEY)).toBeNull()
  })

  it('login ล้มเหลว -> ไม่มี token', async () => {
    const a = useAuthStore()
    vi.mocked(authApi.login).mockRejectedValue(new Error('401'))
    await expect(a.login('u', 'bad')).rejects.toThrow()
    expect(a.isLoggedIn).toBe(false)
  })

  it('อ่าน token เดิมจาก localStorage ตอนสร้าง store', () => {
    localStorage.setItem(TOKEN_KEY, 'old')
    setActivePinia(createPinia())
    expect(useAuthStore().isLoggedIn).toBe(true)
  })
})
