import { computed, ref, watch, type Ref } from 'vue'

export function usePagination<T>(items: Ref<T[]>, pageSize = 5) {
  const page = ref(1)
  const pageCount = computed(() => Math.max(1, Math.ceil(items.value.length / pageSize)))
  const paged = computed(() => items.value.slice((page.value - 1) * pageSize, page.value * pageSize))

  // รายการสั้นลง (เช่นกรอง) แล้วหน้าปัจจุบันเกิน ให้ดึงกลับ
  watch(pageCount, (n) => {
    if (page.value > n) page.value = n
  })

  const next = () => { if (page.value < pageCount.value) page.value++ }
  const prev = () => { if (page.value > 1) page.value-- }
  return { page, pageCount, paged, next, prev }
}
