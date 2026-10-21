import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClaimForm from '../src/components/ClaimForm.vue'

describe('ClaimForm', () => {
  it('ไม่แสดง error ตอนเปิดหน้า', () => {
    const w = mount(ClaimForm)
    expect(w.find('.errors').exists()).toBe(false)
  })

  it('กดส่งโดยไม่กรอก -> แสดง error 3 ข้อ และไม่ emit', async () => {
    const w = mount(ClaimForm)
    await w.find('form').trigger('submit')
    expect(w.findAll('.errors li')).toHaveLength(3)
    expect(w.emitted('submit')).toBeUndefined()
  })

  it('กรอกครบ -> emit submit (trim ค่า) แล้วล้างฟอร์ม', async () => {
    const w = mount(ClaimForm)
    const [ref, hospital, amount] = w.findAll('input')
    await ref.setValue('  P-2001 ')
    await hospital.setValue('รพ.ทดสอบ')
    await amount.setValue('1500')
    await w.find('form').trigger('submit')

    expect(w.emitted('submit')![0][0]).toEqual({
      patientRef: 'P-2001',
      hospital: 'รพ.ทดสอบ',
      amount: 1500,
    })
    expect((ref.element as HTMLInputElement).value).toBe('')
    expect(w.find('.errors').exists()).toBe(false)
  })

  it('จำนวนเงิน 0 ไม่ผ่าน validation', async () => {
    const w = mount(ClaimForm)
    const [ref, hospital] = w.findAll('input')
    await ref.setValue('P-1')
    await hospital.setValue('รพ.')
    await w.find('form').trigger('submit')
    expect(w.findAll('.errors li')).toHaveLength(1)
    expect(w.emitted('submit')).toBeUndefined()
  })
})
