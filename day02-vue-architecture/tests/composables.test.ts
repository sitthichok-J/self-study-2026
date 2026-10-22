import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { useFetch } from '../src/composables/useFetch'
import { usePagination } from '../src/composables/usePagination'
import { provideToast, useToast } from '../src/composables/useToast'

function host<T>(setup: () => T) {
  let result!: T
  const w = mount(defineComponent({ setup() { result = setup(); return () => h('div') } }))
  return { w, result }
}

describe('useFetch', () => {
  it('โหลดสำเร็จ: loading true -> false และมี data', async () => {
    const { result } = host(() => useFetch(async () => [1, 2]))
    expect(result.loading.value).toBe(true)
    await flushPromises()
    expect(result.loading.value).toBe(false)
    expect(result.data.value).toEqual([1, 2])
    expect(result.error.value).toBeNull()
  })

  it('ล้มเหลว: เก็บ error message', async () => {
    const { result } = host(() => useFetch(async () => { throw new Error('พัง') }))
    await flushPromises()
    expect(result.error.value).toBe('พัง')
    expect(result.data.value).toBeNull()
  })

  it('immediate=false ไม่ยิงเอง, refresh ยิงเมื่อเรียก', async () => {
    let calls = 0
    const { result } = host(() => useFetch(async () => ++calls, false))
    await flushPromises()
    expect(calls).toBe(0)
    await result.refresh()
    expect(result.data.value).toBe(1)
  })

  it('state แยกต่อ instance', async () => {
    const a = host(() => useFetch(async () => 'a')).result
    const b = host(() => useFetch(async () => 'b')).result
    await flushPromises()
    expect([a.data.value, b.data.value]).toEqual(['a', 'b'])
  })
})

describe('usePagination', () => {
  it('แบ่งหน้า next/prev ไม่เกินขอบ', () => {
    const items = ref([1, 2, 3, 4, 5, 6, 7])
    const { result: p } = host(() => usePagination(items, 3))
    expect(p.pageCount.value).toBe(3)
    expect(p.paged.value).toEqual([1, 2, 3])
    p.next(); p.next(); p.next()
    expect(p.page.value).toBe(3)
    expect(p.paged.value).toEqual([7])
    p.prev(); p.prev(); p.prev()
    expect(p.page.value).toBe(1)
  })

  it('รายการสั้นลง -> ดึงหน้ากลับ และว่าง = 1 หน้า', async () => {
    const items = ref([1, 2, 3, 4, 5, 6, 7])
    const { result: p } = host(() => usePagination(items, 3))
    p.next(); p.next()
    items.value = [1]
    await flushPromises()
    expect(p.page.value).toBe(1)
    items.value = []
    expect(p.pageCount.value).toBe(1)
  })
})

describe('useToast (provide/inject)', () => {
  it('ลูกใช้ toast ที่ parent provide ได้', () => {
    let api!: ReturnType<typeof useToast>
    const Child = defineComponent({ setup() { api = useToast(); return () => h('i') } })
    mount(defineComponent({ setup() { provideToast(); return () => h(Child) } }))
    api.show('สวัสดี', 0)
    expect(api.toasts.value.map((t) => t.text)).toEqual(['สวัสดี'])
  })

  it('ไม่มี provider -> throw', () => {
    expect(() => host(() => useToast())).toThrow(/provideToast/)
  })
})
