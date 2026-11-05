import { onScopeDispose, ref, shallowRef } from 'vue'
import { ApiError, describeError } from '@/api/http'

export interface PollOptions<T> {
  intervalMs?: number
  maxFailures?: number // พลาดติดกันกี่ครั้งถึงหยุด
  isDone: (v: T) => boolean
}

// poll จนได้ผลสุดท้าย (isDone) หยุดเมื่อ component ถูกทำลาย
// เมื่อพลาด เว้นช่วงนานขึ้นเป็นสองเท่า (สูงสุด 30 วินาที) และหยุดถ้าพลาดติดกันเกิน maxFailures
// 404 ไม่ต้อง retry เพราะไม่มีทางหายเอง
export function usePolling<T>(fn: () => Promise<T>, opts: PollOptions<T>) {
  const interval = opts.intervalMs ?? 2000
  const maxFailures = opts.maxFailures ?? 5
  const data = shallowRef<T | null>(null)
  const error = ref('')
  const loading = ref(true)
  const running = ref(false)
  let timer: ReturnType<typeof setTimeout> | undefined
  let stopped = false
  let failures = 0

  function stop() {
    stopped = true
    running.value = false
    if (timer) clearTimeout(timer)
  }

  async function tick() {
    if (stopped) return
    try {
      const v = await fn()
      if (stopped) return
      data.value = v
      error.value = ''
      failures = 0
      loading.value = false
      if (opts.isDone(v)) return stop()
      timer = setTimeout(tick, interval)
    } catch (e) {
      if (stopped) return
      loading.value = false
      error.value = describeError(e)
      failures++
      if ((e instanceof ApiError && e.status === 404) || failures >= maxFailures) return stop()
      timer = setTimeout(tick, Math.min(interval * 2 ** failures, 30_000))
    }
  }

  function start() {
    stopped = false
    failures = 0
    running.value = true
    void tick()
  }

  onScopeDispose(stop)
  start()
  return { data, error, loading, running, restart: start, stop }
}
