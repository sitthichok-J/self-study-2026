import { ref, shallowRef } from 'vue'
import { describeError } from '@/api/http'

// โหลดข้อมูลแบบสม่ำเสมอ: loading / error / data ในที่เดียว
// ถ้ามี run() ซ้อนกัน ใช้ผลของครั้งล่าสุดเท่านั้น (กันคำตอบเก่ามาทับ)
export function useLoad<T>(fn: () => Promise<T>) {
  const data = shallowRef<T | null>(null)
  const loading = ref(false)
  const error = ref('')
  let ticket = 0

  async function run(): Promise<void> {
    const mine = ++ticket
    loading.value = true
    error.value = ''
    try {
      const r = await fn()
      if (mine !== ticket) return
      data.value = r
    } catch (e) {
      if (mine !== ticket) return
      error.value = describeError(e)
    } finally {
      if (mine === ticket) loading.value = false
    }
  }

  return { data, loading, error, run }
}
