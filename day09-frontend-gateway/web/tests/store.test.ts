import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { PAGE_SIZE, useFilters } from '@/stores/filters'

describe('filters store', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('ค่าเริ่มต้น', () => {
    const f = useFilters()
    expect(f.status).toBe('')
    expect(f.page).toBe(0)
    expect(f.offset).toBe(0)
    expect(f.limit).toBe(PAGE_SIZE)
  })

  it('offset คิดจากหน้า', () => {
    const f = useFilters()
    f.next()
    f.next()
    expect(f.offset).toBe(2 * PAGE_SIZE)
  })

  it('prev ไม่ติดลบ', () => {
    const f = useFilters()
    f.prev()
    expect(f.page).toBe(0)
  })

  it('เปลี่ยน filter แล้วกลับหน้าแรก', () => {
    const f = useFilters()
    f.next()
    f.next()
    f.setStatus('APPROVED')
    expect(f.page).toBe(0)
    expect(f.status).toBe('APPROVED')
  })
})
