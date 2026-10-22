// component ที่ยกมาจาก Day 1 (ClaimForm, ClaimItem, ClaimList, SearchBox) ต้องยังทำงานเหมือนเดิมใน Day 2
import { mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClaimForm from '../src/components/ClaimForm.vue'
import ClaimItem from '../src/components/ClaimItem.vue'
import ClaimList from '../src/components/ClaimList.vue'
import SearchBox from '../src/components/SearchBox.vue'
import type { Claim } from '../src/types/claim'

const claim: Claim = { id: 7, patientRef: 'P-1', hospital: 'รพ.ทดสอบ', amount: 1234.5, status: 'PENDING' }
const global = { stubs: { RouterLink: RouterLinkStub } }

describe('ClaimForm (จาก Day 1)', () => {
  it('ไม่โชว์ error ตอนเปิดหน้า / กดส่งว่าง -> error 3 ข้อ ไม่ emit', async () => {
    const w = mount(ClaimForm)
    expect(w.find('.errors').exists()).toBe(false)
    await w.find('form').trigger('submit')
    expect(w.findAll('.errors li')).toHaveLength(3)
    expect(w.emitted('submit')).toBeUndefined()
  })

  it('กรอกครบ -> emit (trim) แล้วล้างฟอร์ม', async () => {
    const w = mount(ClaimForm)
    const [a, b, c] = w.findAll('input')
    await a.setValue('  P-2001 ')
    await b.setValue('รพ.ทดสอบ')
    await c.setValue('1500')
    await w.find('form').trigger('submit')
    expect(w.emitted('submit')![0][0]).toEqual({ patientRef: 'P-2001', hospital: 'รพ.ทดสอบ', amount: 1500 })
    expect((a.element as HTMLInputElement).value).toBe('')
  })
})

describe('ClaimItem', () => {
  it('PENDING มีปุ่มอนุมัติ/ปฏิเสธ/ลบ และ emit', async () => {
    const w = mount(ClaimItem, { props: { claim }, global })
    const btns = w.findAll('button')
    expect(btns.map((b) => b.text())).toEqual(['อนุมัติ', 'ปฏิเสธ', 'ลบ'])
    await btns[0].trigger('click')
    await btns[2].trigger('click')
    expect(w.emitted('change-status')![0]).toEqual([7, 'APPROVED'])
    expect(w.emitted('remove')![0]).toEqual([7])
  })

  it('สถานะอื่นมีปุ่มย้อนกลับแทน', () => {
    const w = mount(ClaimItem, { props: { claim: { ...claim, status: 'APPROVED' } }, global })
    expect(w.text()).toContain('ย้อนเป็นรอตรวจสอบ')
    expect(w.text()).not.toContain('ปฏิเสธ')
  })

  it('รหัสผู้ป่วยเป็นลิงก์ไปหน้า detail', () => {
    const w = mount(ClaimItem, { props: { claim }, global })
    expect(w.findComponent(RouterLinkStub).props('to')).toEqual({ name: 'claim-detail', params: { id: 7 } })
  })
})

describe('ClaimList', () => {
  it('ว่าง -> แสดง named slot empty', () => {
    const w = mount(ClaimList, { props: { claims: [] }, slots: { empty: 'ว่างเปล่า' }, global })
    expect(w.text()).toBe('ว่างเปล่า')
  })

  it('re-emit change-status / remove จาก item', async () => {
    const w = mount(ClaimList, { props: { claims: [claim] }, global })
    await w.findAll('button')[1].trigger('click')
    await w.findAll('button')[2].trigger('click')
    expect(w.emitted('change-status')![0]).toEqual([7, 'REJECTED'])
    expect(w.emitted('remove')![0]).toEqual([7])
  })
})

describe('SearchBox (defineModel)', () => {
  it('v-model สองทาง', async () => {
    const w = mount(SearchBox, { props: { modelValue: 'a', 'onUpdate:modelValue': (v: string) => w.setProps({ modelValue: v }) } })
    expect((w.find('input').element as HTMLInputElement).value).toBe('a')
    await w.find('input').setValue('xyz')
    expect(w.props('modelValue')).toBe('xyz')
  })
})
