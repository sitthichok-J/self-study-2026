// component ที่ยกหน้าตามาจาก Day 2 (BaseButton, ClaimItem, ClaimList) แต่รับข้อมูลรูปแบบ API ของ Day 8
import { mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import BaseButton from '@/components/BaseButton.vue'
import ClaimItem from '@/components/ClaimItem.vue'
import ClaimList from '@/components/ClaimList.vue'
import type { Claim } from '@/types/claim'
import { sampleClaim } from './helpers'

const global = { stubs: { RouterLink: RouterLinkStub } }
const claim = sampleClaim({ amount: 123450, status: 'APPROVED' }) as Claim

describe('BaseButton', () => {
  it('ค่าเริ่มต้น: type=button, variant primary, แสดง slot', () => {
    const w = mount(BaseButton, { slots: { default: 'ส่ง' } })
    expect(w.attributes('type')).toBe('button')
    expect(w.classes()).toContain('btn--primary')
    expect(w.text()).toBe('ส่ง')
  })

  it('ส่ง attribute ต่อ (disabled, data-test) และ click ไปที่ปุ่ม', async () => {
    let clicked = 0
    const w = mount(BaseButton, { attrs: { 'data-test': 'x', onClick: () => clicked++ }, props: { variant: 'ghost' } })
    await w.trigger('click')
    expect(clicked).toBe(1)
    expect(w.attributes('data-test')).toBe('x')
    expect(w.classes()).toContain('btn--ghost')
    await w.setProps({ type: 'submit' })
    expect(w.attributes('type')).toBe('submit')
  })

  it('disabled ถูกส่งต่อ', () => {
    const w = mount(BaseButton, { attrs: { disabled: true } })
    expect(w.attributes('disabled')).toBeDefined()
  })
})

describe('ClaimItem', () => {
  it('แสดงรหัสผู้ป่วย, id สั้น, จำนวนเงินจากสตางค์, ป้ายสถานะ', () => {
    const w = mount(ClaimItem, { props: { claim }, global })
    expect(w.text()).toContain('FAKE-AB12')
    expect(w.text()).toContain('11111111')
    expect(w.text()).not.toContain('11111111-2222')
    expect(w.text()).toContain('฿ 1,234.50')
    expect(w.find('.badge').classes()).toContain('badge--APPROVED')
    expect(w.find('.badge').text()).toBe('อนุมัติ')
  })

  it('ลิงก์ไปหน้ารายละเอียดด้วย id เต็ม', () => {
    const w = mount(ClaimItem, { props: { claim }, global })
    expect(w.findComponent(RouterLinkStub).props('to')).toBe(`/claims/${claim.id}`)
  })

  it.each([
    ['PENDING', 'รอประมวลผล'],
    ['APPROVED', 'อนุมัติ'],
    ['REJECTED', 'ไม่อนุมัติ'],
  ])('สถานะ %s -> ป้าย %s', (status, label) => {
    const w = mount(ClaimItem, { props: { claim: { ...claim, status } as Claim }, global })
    expect(w.find('.badge').classes()).toContain(`badge--${status}`)
    expect(w.find('.badge').text()).toBe(label)
  })
})

describe('ClaimList', () => {
  it('แสดงทุกแถวเป็น data-test=row', () => {
    const w = mount(ClaimList, { props: { claims: [claim, { ...claim, id: 'b'.repeat(36) } as Claim] }, global })
    expect(w.findAll('[data-test=row]')).toHaveLength(2)
    expect(w.find('[data-test=empty]').exists()).toBe(false)
  })

  it('ว่าง -> แสดง slot empty, ไม่มีแถว', () => {
    const w = mount(ClaimList, { props: { claims: [] }, slots: { empty: 'ว่างเปล่า' }, global })
    expect(w.find('[data-test=empty]').text()).toBe('ว่างเปล่า')
    expect(w.findAll('[data-test=row]')).toHaveLength(0)
  })

  it('ว่างและไม่มี slot -> ข้อความเริ่มต้น', () => {
    const w = mount(ClaimList, { props: { claims: [] }, global })
    expect(w.text()).toBe('ไม่มีรายการ')
  })
})
