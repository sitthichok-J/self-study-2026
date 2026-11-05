import { effectScope, nextTick } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/http'
import { useLoad } from '@/composables/useLoad'
import { usePolling } from '@/composables/usePolling'

describe('useLoad', () => {
  it('ตั้ง loading แล้วเก็บ data', async () => {
    const { data, loading, error, run } = useLoad(async () => 42)
    const p = run()
    expect(loading.value).toBe(true)
    await p
    expect(loading.value).toBe(false)
    expect(data.value).toBe(42)
    expect(error.value).toBe('')
  })

  it('error เป็นข้อความไทย และล้าง error เมื่อ run ใหม่สำเร็จ', async () => {
    let fail = true
    const { error, run } = useLoad(async () => {
      if (fail) throw new ApiError(500, 'x')
      return 1
    })
    await run()
    expect(error.value).toContain('ระบบขัดข้อง')
    fail = false
    await run()
    expect(error.value).toBe('')
  })

  it('คำตอบเก่าที่มาช้ากว่าไม่ทับคำตอบใหม่', async () => {
    const resolvers: Array<(v: string) => void> = []
    const { data, loading, run } = useLoad(() => new Promise<string>((r) => resolvers.push(r)))
    const first = run()
    const second = run()
    resolvers[1]!('ใหม่')
    await second
    resolvers[0]!('เก่า')
    await first
    expect(data.value).toBe('ใหม่')
    expect(loading.value).toBe(false)
  })
})

describe('usePolling', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  const mk = <T>(fn: () => Promise<T>, isDone: (v: T) => boolean, extra = {}) => {
    const scope = effectScope()
    const r = scope.run(() => usePolling(fn, { intervalMs: 1000, isDone, ...extra }))!
    return { scope, ...r }
  }

  it('poll ต่อจน isDone แล้วหยุด', async () => {
    const seq = ['PENDING', 'PENDING', 'APPROVED']
    const fn = vi.fn(async () => seq.shift()!)
    const { data, running } = mk(fn, (v) => v !== 'PENDING')
    await vi.advanceTimersByTimeAsync(0)
    expect(data.value).toBe('PENDING')
    expect(running.value).toBe(true)
    await vi.advanceTimersByTimeAsync(2000)
    expect(data.value).toBe('APPROVED')
    expect(running.value).toBe(false)
    await vi.advanceTimersByTimeAsync(10_000)
    expect(fn).toHaveBeenCalledTimes(3)
  })

  it('หยุด poll เมื่อ scope ถูกทำลาย (ออกจากหน้า)', async () => {
    const fn = vi.fn(async () => 'PENDING')
    const { scope } = mk(fn, () => false)
    await vi.advanceTimersByTimeAsync(2500)
    const calls = fn.mock.calls.length
    scope.stop()
    await vi.advanceTimersByTimeAsync(10_000)
    expect(fn).toHaveBeenCalledTimes(calls)
  })

  it('คำตอบที่มาหลัง stop ไม่ตั้ง timer ใหม่', async () => {
    let release!: (v: string) => void
    const fn = vi.fn(() => new Promise<string>((r) => (release = r)))
    const { scope, data } = mk(fn, () => false)
    await vi.advanceTimersByTimeAsync(0)
    scope.stop()
    release('late')
    await vi.advanceTimersByTimeAsync(5000)
    expect(data.value).toBeNull()
    expect(fn).toHaveBeenCalledTimes(1)
  })

  it('พลาดแล้วเว้นช่วงนานขึ้น และกลับมาปกติเมื่อสำเร็จ', async () => {
    let n = 0
    const fn = vi.fn(async () => {
      n++
      if (n <= 2) throw new ApiError(0, 'down')
      return 'OK'
    })
    const { error, data } = mk(fn, () => false)
    await vi.advanceTimersByTimeAsync(0)
    expect(fn).toHaveBeenCalledTimes(1)
    expect(error.value).toContain('ติดต่อเซิร์ฟเวอร์ไม่ได้')
    await vi.advanceTimersByTimeAsync(1999) // ครั้งที่ 2 รอ 2 วินาที
    expect(fn).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1)
    expect(fn).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(4000) // ครั้งที่ 3 รอ 4 วินาที
    expect(fn).toHaveBeenCalledTimes(3)
    expect(data.value).toBe('OK')
    expect(error.value).toBe('')
  })

  it('พลาดติดกันเกิน maxFailures แล้วหยุด', async () => {
    const fn = vi.fn(async () => {
      throw new ApiError(500, 'x')
    })
    const { running } = mk(fn, () => false, { maxFailures: 3 })
    await vi.advanceTimersByTimeAsync(60_000)
    expect(fn).toHaveBeenCalledTimes(3)
    expect(running.value).toBe(false)
  })

  it('404 หยุดทันที ไม่ retry', async () => {
    const fn = vi.fn(async () => {
      throw new ApiError(404, 'claim not found')
    })
    const { running, error } = mk(fn, () => false)
    await vi.advanceTimersByTimeAsync(60_000)
    expect(fn).toHaveBeenCalledTimes(1)
    expect(running.value).toBe(false)
    expect(error.value).toBe('ไม่พบข้อมูล')
  })

  it('restart เริ่ม poll ใหม่หลังหยุด', async () => {
    let fail = true
    const fn = vi.fn(async () => {
      if (fail) throw new ApiError(404, 'x')
      return 'DONE'
    })
    const { restart, data, running } = mk(fn, (v) => v === 'DONE')
    await vi.advanceTimersByTimeAsync(0)
    expect(running.value).toBe(false)
    fail = false
    restart()
    await vi.advanceTimersByTimeAsync(0)
    await nextTick()
    expect(data.value).toBe('DONE')
  })
})
