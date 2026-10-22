import { onMounted, ref, type Ref } from 'vue'

// composable: state ถูกสร้างใหม่ทุกครั้งที่เรียก (แยกต่อ instance)
export function useFetch<T>(fetcher: () => Promise<T>, immediate = true) {
  const data = ref<T | null>(null) as Ref<T | null>
  const error = ref<string | null>(null)
  const loading = ref(false)

  async function run() {
    loading.value = true
    error.value = null
    try {
      data.value = await fetcher()
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  if (immediate) onMounted(run)
  return { data, error, loading, refresh: run }
}
