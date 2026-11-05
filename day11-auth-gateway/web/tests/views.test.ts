import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory } from 'vue-router'
import { ApiError } from '@/api/http'
import * as api from '@/api/claims'
import { makeRouter } from '@/router'
import ClaimDetailView from '@/views/ClaimDetailView.vue'
import ClaimFormView from '@/views/ClaimFormView.vue'
import ClaimListView from '@/views/ClaimListView.vue'
import { sampleClaim } from './helpers'

vi.mock('@/api/claims')
const m = vi.mocked(api)

async function mountWith<T>(comp: T, props: Record<string, unknown> = {}, path = '/') {
  const pinia = createPinia()
  setActivePinia(pinia)
  const router = makeRouter(createMemoryHistory())
  router.push(path)
  await router.isReady()
  const w = mount(comp as never, { props: props as never, global: { plugins: [pinia, router] } })
  await flushPromises()
  return { w, router }
}

const claims = (n: number, status = 'APPROVED') =>
  Array.from({ length: n }, (_, i) =>
    sampleClaim({ id: `00000000-0000-0000-0000-${String(i).padStart(12, '0')}`, status }),
  )

beforeEach(() => vi.resetAllMocks())
afterEach(() => vi.useRealTimers())

describe('ClaimListView', () => {
  it('แสดงแถวและขอ limit+1 จาก API', async () => {
    m.listClaims.mockResolvedValue({ items: claims(3), limit: 11, offset: 0 } as never)
    const { w } = await mountWith(ClaimListView)
    expect(w.findAll('[data-test=row]')).toHaveLength(3)
    expect(m.listClaims).toHaveBeenCalledWith({ status: '', limit: 11, offset: 0 })
    expect(w.text()).toContain('1,500.00')
    expect(w.text()).toContain('อนุมัติ')
  })

  it('รายการว่างแสดงข้อความว่าง', async () => {
    m.listClaims.mockResolvedValue({ items: [], limit: 11, offset: 0 })
    const { w } = await mountWith(ClaimListView)
    expect(w.find('[data-test=empty]').exists()).toBe(true)
  })

  it('ได้ 11 แถว แสดงแค่ 10 และเปิดปุ่มถัดไป', async () => {
    m.listClaims.mockResolvedValue({ items: claims(11), limit: 11, offset: 0 } as never)
    const { w } = await mountWith(ClaimListView)
    expect(w.findAll('[data-test=row]')).toHaveLength(10)
    expect(w.find('[data-test=next]').attributes('disabled')).toBeUndefined()
    expect(w.find('[data-test=prev]').attributes('disabled')).toBeDefined()
  })

  it('ได้ 10 แถวพอดี ไม่มีหน้าถัดไป', async () => {
    m.listClaims.mockResolvedValue({ items: claims(10), limit: 11, offset: 0 } as never)
    const { w } = await mountWith(ClaimListView)
    expect(w.find('[data-test=next]').attributes('disabled')).toBeDefined()
  })

  it('กดถัดไปแล้วขอ offset 10', async () => {
    m.listClaims.mockResolvedValue({ items: claims(11), limit: 11, offset: 0 } as never)
    const { w } = await mountWith(ClaimListView)
    await w.find('[data-test=next]').trigger('click')
    await flushPromises()
    expect(m.listClaims).toHaveBeenLastCalledWith({ status: '', limit: 11, offset: 10 })
    expect(w.find('[data-test=page]').text()).toBe('หน้า 2')
  })

  it('เปลี่ยน filter แล้วกลับหน้าแรกและส่ง status', async () => {
    m.listClaims.mockResolvedValue({ items: claims(11), limit: 11, offset: 0 } as never)
    const { w } = await mountWith(ClaimListView)
    await w.find('[data-test=next]').trigger('click')
    await flushPromises()
    await w.find('[data-test=status]').setValue('REJECTED')
    await flushPromises()
    expect(m.listClaims).toHaveBeenLastCalledWith({ status: 'REJECTED', limit: 11, offset: 0 })
    expect(w.find('[data-test=page]').text()).toBe('หน้า 1')
  })

  it('error แสดงข้อความและกดลองใหม่ได้', async () => {
    m.listClaims.mockRejectedValueOnce(new ApiError(429, 'x', 5))
    const { w } = await mountWith(ClaimListView)
    expect(w.find('[data-test=error]').text()).toContain('ถี่เกินไป')
    m.listClaims.mockResolvedValue({ items: claims(1), limit: 11, offset: 0 } as never)
    await w.find('[data-test=retry]').trigger('click')
    await flushPromises()
    expect(w.find('[data-test=error]').exists()).toBe(false)
    expect(w.findAll('[data-test=row]')).toHaveLength(1)
  })
})

describe('ClaimFormView', () => {
  it('ข้อมูลผิด ไม่เรียก API และแสดง error ทั้งสองช่อง', async () => {
    const { w } = await mountWith(ClaimFormView, {}, '/new')
    await w.find('[data-test=patient]').setValue('REAL-1')
    await w.find('[data-test=amount]').setValue('abc')
    await w.find('form').trigger('submit')
    expect(m.createClaim).not.toHaveBeenCalled()
    expect(w.find('[data-test=patient-error]').exists()).toBe(true)
    expect(w.find('[data-test=amount-error]').exists()).toBe(true)
  })

  it('ก่อนกดส่งครั้งแรกยังไม่ขึ้น error แต่หลังกดแล้ว แก้ถูกต้อง error ต้องหายทันที', async () => {
    const { w } = await mountWith(ClaimFormView, {}, '/new')
    await w.find('[data-test=patient]').setValue('x')
    expect(w.find('[data-test=patient-error]').exists()).toBe(false)
    await w.find('form').trigger('submit')
    expect(w.find('[data-test=patient-error]').exists()).toBe(true)
    await w.find('[data-test=patient]').setValue('FAKE-AB12')
    expect(w.find('[data-test=patient-error]').exists()).toBe(false)
    expect(w.find('[data-test=amount-error]').exists()).toBe(true) // ช่องจำนวนเงินยังผิดอยู่
    await w.find('[data-test=amount]').setValue('10')
    expect(w.find('[data-test=amount-error]').exists()).toBe(false)
  })

  it('ส่งเป็นสตางค์ แล้วไปหน้ารายละเอียด', async () => {
    m.createClaim.mockResolvedValue({ claim: sampleClaim() as never, replayed: false })
    const { w, router } = await mountWith(ClaimFormView, {}, '/new')
    await w.find('[data-test=patient]').setValue('FAKE-AB12')
    await w.find('[data-test=amount]').setValue('1500.50')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(m.createClaim).toHaveBeenCalledWith({ patient_ref: 'FAKE-AB12', amount: 150050 }, expect.any(String))
    expect(router.currentRoute.value.path).toBe(`/claims/${sampleClaim().id}`)
  })

  it('ส่งซ้ำด้วยข้อมูลเดิมใช้ Idempotency-Key เดิม แต่แก้ข้อมูลแล้วได้ key ใหม่', async () => {
    m.createClaim.mockRejectedValue(new ApiError(0, 'down'))
    const { w } = await mountWith(ClaimFormView, {}, '/new')
    await w.find('[data-test=patient]').setValue('FAKE-AB12')
    await w.find('[data-test=amount]').setValue('10')
    await w.find('form').trigger('submit')
    await flushPromises()
    await w.find('form').trigger('submit')
    await flushPromises()
    const [k1, k2] = m.createClaim.mock.calls.map((c) => c[1])
    expect(k1).toBe(k2)
    await w.find('[data-test=amount]').setValue('11')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(m.createClaim.mock.calls[2]![1]).not.toBe(k1)
    expect(w.find('[data-test=submit-error]').text()).toContain('ติดต่อเซิร์ฟเวอร์ไม่ได้')
  })

  it('กำลังส่งอยู่ กดซ้ำไม่ยิง API สองครั้ง', async () => {
    let release!: (v: never) => void
    m.createClaim.mockImplementation(() => new Promise((r) => (release = r as never)))
    const { w } = await mountWith(ClaimFormView, {}, '/new')
    await w.find('[data-test=patient]').setValue('FAKE-AB12')
    await w.find('[data-test=amount]').setValue('10')
    await w.find('form').trigger('submit')
    await w.find('form').trigger('submit')
    expect(m.createClaim).toHaveBeenCalledTimes(1)
    expect(w.find('[data-test=submit]').attributes('disabled')).toBeDefined()
    release({ claim: sampleClaim(), replayed: false } as never)
    await flushPromises()
  })

  it('400 จาก backend แสดงข้อความของ backend', async () => {
    m.createClaim.mockRejectedValue(new ApiError(400, 'invalid input: amount must be between 1 and 1000000000'))
    const { w } = await mountWith(ClaimFormView, {}, '/new')
    await w.find('[data-test=patient]').setValue('FAKE-AB12')
    await w.find('[data-test=amount]').setValue('10')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(w.find('[data-test=submit-error]').text()).toContain('invalid input')
  })
})

describe('ClaimDetailView', () => {
  it('claim ที่ผ่านเกณฑ์ (decision_reason ok) แสดงเป็นภาษาไทย ไม่โชว์คำดิบ', async () => {
    vi.useFakeTimers()
    m.getClaim.mockResolvedValue(sampleClaim({ status: 'APPROVED', decision_reason: 'ok' }) as never)
    const { w } = await mountWith(ClaimDetailView, { id: 'x' }, '/claims/x')
    expect(w.find('[data-test=reason]').text()).toBe('ผ่านเกณฑ์')
  })

  it('poll จน worker ตัดสินแล้วหยุด พร้อมแสดงเหตุผล', async () => {
    vi.useFakeTimers()
    m.getClaim
      .mockResolvedValueOnce(sampleClaim() as never)
      .mockResolvedValueOnce(sampleClaim({ status: 'REJECTED', decision_reason: 'over_limit' }) as never)
    const { w } = await mountWith(ClaimDetailView, { id: sampleClaim().id }, `/claims/${sampleClaim().id}`)
    expect(w.find('[data-test=status]').text()).toBe('รอประมวลผล')
    expect(w.find('[data-test=polling]').exists()).toBe(true)
    await vi.advanceTimersByTimeAsync(2000)
    await flushPromises()
    expect(w.find('[data-test=status]').text()).toBe('ไม่อนุมัติ')
    expect(w.find('[data-test=reason]').text()).toBe('เกินวงเงิน')
    expect(w.find('[data-test=polling]').exists()).toBe(false)
    await vi.advanceTimersByTimeAsync(20_000)
    expect(m.getClaim).toHaveBeenCalledTimes(2)
  })

  it('ออกจากหน้าแล้วหยุด poll', async () => {
    vi.useFakeTimers()
    m.getClaim.mockResolvedValue(sampleClaim() as never)
    const { w } = await mountWith(ClaimDetailView, { id: 'x' }, '/claims/x')
    const n = m.getClaim.mock.calls.length
    w.unmount()
    await vi.advanceTimersByTimeAsync(20_000)
    expect(m.getClaim).toHaveBeenCalledTimes(n)
  })

  it('404 แสดงไม่พบ และไม่ poll ต่อ', async () => {
    vi.useFakeTimers()
    m.getClaim.mockRejectedValue(new ApiError(404, 'claim not found'))
    const { w } = await mountWith(ClaimDetailView, { id: 'x' }, '/claims/x')
    expect(w.find('[data-test=error]').text()).toContain('ไม่พบข้อมูล')
    await vi.advanceTimersByTimeAsync(20_000)
    expect(m.getClaim).toHaveBeenCalledTimes(1)
  })
})
