// Baseline: บันทึกพฤติกรรม/โครงสร้างปัจจุบันของแอปไว้เป็นจุดอ้างอิง
// ถ้าแก้โค้ดแล้วเทสต์นี้แดง = พฤติกรรมเปลี่ยน ต้องตัดสินใจว่าตั้งใจหรือไม่
// ตั้งใจ -> อัปเดต snapshot ด้วย `npm run test:update` และเพิ่ม/แก้ unit test ของส่วนที่เปลี่ยน
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import App from '../src/App.vue'
import BaseButton from '../src/components/BaseButton.vue'
import ClaimForm from '../src/components/ClaimForm.vue'
import ClaimItem from '../src/components/ClaimItem.vue'
import ClaimList from '../src/components/ClaimList.vue'
import SearchBox from '../src/components/SearchBox.vue'
import ThemeToggle from '../src/components/ThemeToggle.vue'
import { STATUS_LABEL } from '../src/types'

beforeEach(() => {
  localStorage.clear()
  document.documentElement.removeAttribute('data-theme')
})

describe('baseline: ข้อมูลและค่าคงที่', () => {
  it('STATUS_LABEL', () => {
    expect(STATUS_LABEL).toMatchInlineSnapshot(`
      {
        "APPROVED": "อนุมัติ",
        "PENDING": "รอตรวจสอบ",
        "REJECTED": "ปฏิเสธ",
      }
    `)
  })

  it('สถานะเริ่มต้นของ App: 4 รายการ, สรุปจำนวนและยอดรวม', () => {
    const w = mount(App)
    const stats = w.findAll('.stat').map((s) => s.text())
    expect(stats).toEqual(['ทั้งหมด4', 'รอตรวจสอบ2', 'อนุมัติ1', 'ปฏิเสธ1'])
    expect(w.find('.total').text()).toContain('4 รายการ')
    expect(w.find('.total').text()).toContain('56,000.50')
  })
})

describe('baseline: snapshot ของแต่ละคอมโพเนนต์', () => {
  const claim = { id: 1, patientRef: 'P-1001', hospital: 'รพ.ตัวอย่าง A', amount: 12500, status: 'PENDING' as const }

  it('BaseButton', () => {
    expect(mount(BaseButton, { slots: { default: 'ปุ่ม' } }).html()).toMatchSnapshot()
  })
  it('SearchBox', () => {
    expect(mount(SearchBox, { props: { modelValue: '' } }).html()).toMatchSnapshot()
  })
  it('ClaimForm (ก่อนกดส่ง)', () => {
    expect(mount(ClaimForm).html()).toMatchSnapshot()
  })
  it('ClaimItem (PENDING)', () => {
    expect(mount(ClaimItem, { props: { claim } }).html()).toMatchSnapshot()
  })
  it('ClaimList (มีรายการ / ว่าง)', () => {
    expect(mount(ClaimList, { props: { claims: [claim] } }).html()).toMatchSnapshot()
    expect(
      mount(ClaimList, { props: { claims: [] }, slots: { empty: 'ว่าง' } }).html(),
    ).toMatchSnapshot()
  })
  it('ThemeToggle', () => {
    expect(mount(ThemeToggle).html()).toMatchSnapshot()
  })
  it('App ทั้งหน้า', () => {
    expect(mount(App).html()).toMatchSnapshot()
  })
})
