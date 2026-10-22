import { flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { authApi, claimsApi } from '../src/api/claimsApi'
import { ApiError, TOKEN_KEY } from '../src/api/http'
import type { Claim } from '../src/types/claim'
import { mountAt } from './helpers'

vi.mock('../src/api/claimsApi', () => ({
  claimsApi: { list: vi.fn(), get: vi.fn(), create: vi.fn(), updateStatus: vi.fn(), remove: vi.fn() },
  authApi: { login: vi.fn() },
}))

const mk = (id: number, status: Claim['status'] = 'PENDING', hospital = 'รพ.A'): Claim => ({
  id, patientRef: `P-${id}`, hospital, amount: id * 1000, status,
})

beforeEach(() => {
  vi.resetAllMocks()
  localStorage.clear()
  localStorage.setItem(TOKEN_KEY, 't')
  document.body.innerHTML = ''
})

describe('LoginView', () => {
  beforeEach(() => localStorage.clear())

  it('login สำเร็จ -> ไปหน้าที่ redirect ไว้', async () => {
    vi.mocked(authApi.login).mockResolvedValue({ token: 'tok' })
    vi.mocked(claimsApi.get).mockResolvedValue(mk(2))
    const { w, router } = await mountAt('/login?redirect=/claims/2')
    const [u, p] = w.findAll('input')
    await u.setValue('demo')
    await p.setValue('demo1234')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(authApi.login).toHaveBeenCalledWith('demo', 'demo1234')
    // route component เป็น lazy import ต้องรอให้ navigation เสร็จ
    await vi.waitFor(() => expect(router.currentRoute.value.fullPath).toBe('/claims/2'), { timeout: 5000 })
  })

  it('login ผิด -> แสดง error และยังอยู่หน้า login', async () => {
    vi.mocked(authApi.login).mockRejectedValue(new ApiError(401, 'รหัสผ่านไม่ถูกต้อง'))
    const { w, router } = await mountAt('/login')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(w.find('[role="alert"]').text()).toBe('รหัสผ่านไม่ถูกต้อง')
    expect(router.currentRoute.value.name).toBe('login')
  })

  it('ไม่ยอม redirect ไปปลายทางภายนอก', async () => {
    vi.mocked(authApi.login).mockResolvedValue({ token: 'tok' })
    vi.mocked(claimsApi.list).mockResolvedValue([])
    const { w, router } = await mountAt('/login?redirect=https://evil.example')
    await w.find('form').trigger('submit')
    await vi.waitFor(() => expect(router.currentRoute.value.name).toBe('claims'), { timeout: 5000 })
  })
})

describe('ClaimListView', () => {
  const rows = (w: Awaited<ReturnType<typeof mountAt>>['w']) => w.findAll('li.item')

  it('โหลดรายการจาก store/API แล้วแสดง', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue([mk(1), mk(2, 'APPROVED')])
    const { w } = await mountAt('/claims')
    expect(rows(w)).toHaveLength(2)
    expect(w.find('.stats').text()).toContain('ทั้งหมด2')
  })

  it('โหลดไม่สำเร็จ -> แสดง error และปุ่มลองใหม่ทำงาน', async () => {
    vi.mocked(claimsApi.list).mockRejectedValueOnce(new Error('เซิร์ฟเวอร์ล่ม'))
    const { w } = await mountAt('/claims')
    expect(w.find('[role="alert"]').text()).toContain('เซิร์ฟเวอร์ล่ม')
    vi.mocked(claimsApi.list).mockResolvedValue([mk(1)])
    await w.find('.retry').trigger('click')
    await flushPromises()
    expect(rows(w)).toHaveLength(1)
  })

  it('ค้นหา + กรองสถานะ + empty state', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue([mk(1, 'PENDING', 'รพ.A'), mk(2, 'APPROVED', 'รพ.B')])
    const { w } = await mountAt('/claims')
    await w.find('input.search').setValue('รพ.B')
    expect(rows(w)).toHaveLength(1)
    await w.find('input.search').setValue('ไม่มี')
    expect(w.text()).toContain('ไม่พบรายการที่ตรงกับเงื่อนไข')
    await w.find('input.search').setValue('')
    await w.findAll('.tabs button')[1].trigger('click')
    expect(rows(w)).toHaveLength(1)
  })

  it('ปุ่มบนแถว: อนุมัติ/ลบ เรียก API และแสดง toast', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue([mk(1), mk(2)])
    vi.mocked(claimsApi.updateStatus).mockResolvedValue(mk(1, 'APPROVED'))
    vi.mocked(claimsApi.remove).mockResolvedValue(undefined)
    const { w } = await mountAt('/claims')
    await rows(w)[0].findAll('button').find((b) => b.text() === 'อนุมัติ')!.trigger('click')
    await flushPromises()
    expect(claimsApi.updateStatus).toHaveBeenCalledWith(1, 'APPROVED')
    expect(w.find('.stats').text()).toContain('อนุมัติ1')
    await rows(w)[1].findAll('button').find((b) => b.text() === 'ลบ')!.trigger('click')
    await flushPromises()
    expect(claimsApi.remove).toHaveBeenCalledWith(2)
    expect(rows(w)).toHaveLength(1)
  })

  it('action ล้มเหลว -> state ไม่เปลี่ยน', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue([mk(1)])
    vi.mocked(claimsApi.remove).mockRejectedValue(new Error('ลบไม่ได้'))
    const { w } = await mountAt('/claims')
    await rows(w)[0].findAll('button').find((b) => b.text() === 'ลบ')!.trigger('click')
    await flushPromises()
    expect(rows(w)).toHaveLength(1)
  })

  it('สรุปด้านล่าง: จำนวนรายการและยอดรวมตามที่กรอง', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue([mk(1), mk(2, 'APPROVED')])
    const { w } = await mountAt('/claims')
    expect(w.find('.total').text()).toContain('2 รายการ')
    expect(w.find('.total').text()).toContain('3,000.00')
    await w.findAll('.tabs button')[2].trigger('click')
    expect(w.find('.total').text()).toContain('1 รายการ')
    expect(w.find('.total').text()).toContain('2,000.00')
  })

  it('แบ่งหน้า 5 รายการต่อหน้า', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue(Array.from({ length: 7 }, (_, i) => mk(i + 1)))
    const { w } = await mountAt('/claims')
    expect(rows(w)).toHaveLength(5)
    expect(w.find('.pager').text()).toContain('หน้า 1 / 2')
    await w.findAll('.pager button')[1].trigger('click')
    expect(rows(w)).toHaveLength(2)
  })
})

describe('ClaimDetailView', () => {
  it('เปิดลิงก์ตรง (store ว่าง) -> ดึงจาก API', async () => {
    vi.mocked(claimsApi.get).mockResolvedValue(mk(5, 'PENDING', 'รพ.X'))
    const { w } = await mountAt('/claims/5')
    expect(claimsApi.get).toHaveBeenCalledWith(5)
    expect(w.find('h2').text()).toBe('P-5')
    expect(w.text()).toContain('รพ.X')
  })

  it('ไม่พบเคลม -> แสดง error', async () => {
    vi.mocked(claimsApi.get).mockRejectedValue(new ApiError(404, 'ไม่พบเคลม'))
    const { w } = await mountAt('/claims/99')
    expect(w.find('[role="alert"]').text()).toBe('ไม่พบเคลม')
  })

  it('อนุมัติ -> เรียก API, อัปเดตสถานะ, มี toast', async () => {
    vi.mocked(claimsApi.get).mockResolvedValue(mk(5))
    vi.mocked(claimsApi.updateStatus).mockResolvedValue(mk(5, 'APPROVED'))
    const { w } = await mountAt('/claims/5')
    await w.findAll('button').find((b) => b.text() === 'อนุมัติ')!.trigger('click')
    await flushPromises()
    expect(claimsApi.updateStatus).toHaveBeenCalledWith(5, 'APPROVED')
    expect(w.text()).toContain('อนุมัติ')
    expect(w.findAll('button').some((b) => b.text() === 'อนุมัติ')).toBe(false)
  })

  it('ลบ -> กลับหน้ารายการ', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue([mk(5)])
    vi.mocked(claimsApi.remove).mockResolvedValue(undefined)
    const { w, router } = await mountAt('/claims')
    await w.find('li.item a.ref').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.name).toBe('claim-detail')
    await w.findAll('button').find((b) => b.text() === 'ลบ')!.trigger('click')
    await flushPromises()
    expect(claimsApi.remove).toHaveBeenCalledWith(5)
    expect(router.currentRoute.value.name).toBe('claims')
  })
})

describe('ClaimFormView', () => {
  it('ไม่แสดง error ก่อนกดส่ง, กดส่งว่าง -> error 3 ข้อ ไม่เรียก API', async () => {
    const { w } = await mountAt('/claims/new')
    expect(w.find('.errors').exists()).toBe(false)
    await w.find('form').trigger('submit')
    expect(w.findAll('.errors li')).toHaveLength(3)
    expect(claimsApi.create).not.toHaveBeenCalled()
  })

  it('กรอกครบ -> create แล้วไปหน้า detail', async () => {
    vi.mocked(claimsApi.create).mockResolvedValue(mk(9))
    const { w, router } = await mountAt('/claims/new')
    const [a, b, c] = w.findAll('input')
    await a.setValue(' P-9 ')
    await b.setValue('รพ.A')
    await c.setValue('9000')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(claimsApi.create).toHaveBeenCalledWith({ patientRef: 'P-9', hospital: 'รพ.A', amount: 9000 })
    expect(router.currentRoute.value.fullPath).toBe('/claims/9')
  })

  it('server error -> แสดงข้อความ ไม่เปลี่ยนหน้า', async () => {
    vi.mocked(claimsApi.create).mockRejectedValue(new ApiError(422, 'ข้อมูลไม่ครบ'))
    const { w, router } = await mountAt('/claims/new')
    const [a, b, c] = w.findAll('input')
    await a.setValue('P'); await b.setValue('H'); await c.setValue('1')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(w.find('[role="alert"]').text()).toBe('ข้อมูลไม่ครบ')
    expect(router.currentRoute.value.name).toBe('claim-new')
  })
})
