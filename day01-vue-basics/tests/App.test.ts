import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import App from '../src/App.vue'

const rows = (w: ReturnType<typeof mount>) => w.findAll('li.item')

describe('App (integration)', () => {
  it('เริ่มต้นแสดง 4 รายการ', () => {
    const w = mount(App)
    expect(rows(w)).toHaveLength(4)
  })

  it('ค้นหาตามโรงพยาบาล', async () => {
    const w = mount(App)
    await w.find('.toolbar input').setValue('รพ.ตัวอย่าง A')
    expect(rows(w)).toHaveLength(2)
  })

  it('กรองตามสถานะ รอตรวจสอบ', async () => {
    const w = mount(App)
    const tab = w.findAll('.tabs button').find((b) => b.text() === 'รอตรวจสอบ')!
    await tab.trigger('click')
    expect(rows(w)).toHaveLength(2)
  })

  it('อนุมัติรายการ -> จำนวนรอตรวจสอบลดลง', async () => {
    const w = mount(App)
    const pendingStat = () => w.findAll('.stat').find((s) => s.text().includes('รอตรวจสอบ'))!.find('b').text()
    expect(pendingStat()).toBe('2')
    await rows(w)[0].findAll('button').find((b) => b.text() === 'อนุมัติ')!.trigger('click')
    expect(pendingStat()).toBe('1')
  })

  it('ไม่พบรายการ -> แสดงข้อความ empty slot', async () => {
    const w = mount(App)
    await w.find('.toolbar input').setValue('ไม่มีอยู่จริง')
    expect(rows(w)).toHaveLength(0)
    expect(w.text()).toContain('ไม่พบรายการที่ตรงกับเงื่อนไข')
  })

  it('เพิ่มเคลมใหม่ผ่านฟอร์ม', async () => {
    const w = mount(App)
    const inputs = w.find('form').findAll('input')
    await inputs[0].setValue('P-9999')
    await inputs[1].setValue('รพ.ใหม่')
    await inputs[2].setValue('500')
    await w.find('form').trigger('submit')
    expect(rows(w)).toHaveLength(5)
    expect(rows(w)[0].text()).toContain('P-9999')
  })
})
