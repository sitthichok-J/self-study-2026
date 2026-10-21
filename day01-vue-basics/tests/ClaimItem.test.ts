import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClaimItem from '../src/components/ClaimItem.vue'

const claim = { id: 7, patientRef: 'P-1', hospital: 'รพ.ทดสอบ', amount: 1234.5, status: 'PENDING' as const }

describe('ClaimItem', () => {
  it('PENDING มีปุ่มอนุมัติ/ปฏิเสธ และ emit change-status', async () => {
    const w = mount(ClaimItem, { props: { claim } })
    const btns = w.findAll('button')
    expect(btns.map((b) => b.text())).toEqual(['อนุมัติ', 'ปฏิเสธ', 'ลบ'])
    await btns[0].trigger('click')
    expect(w.emitted('change-status')![0]).toEqual([7, 'APPROVED'])
  })

  it('สถานะอื่นมีปุ่มย้อนกลับแทน', () => {
    const w = mount(ClaimItem, { props: { claim: { ...claim, status: 'APPROVED' } } })
    expect(w.text()).toContain('ย้อนเป็นรอตรวจสอบ')
    expect(w.text()).not.toContain('ปฏิเสธ')
  })

  it('ปุ่มลบ emit remove พร้อม id', async () => {
    const w = mount(ClaimItem, { props: { claim } })
    await w.findAll('button').at(-1)!.trigger('click')
    expect(w.emitted('remove')![0]).toEqual([7])
  })
})
