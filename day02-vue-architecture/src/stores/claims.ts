import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { claimsApi } from '../api/claimsApi'
import type { Claim, ClaimStatus, NewClaim } from '../types/claim'

// Pinia (setup store): state ร่วมหลายหน้า — list/detail/form ใช้ชุดเดียวกัน
export const useClaimsStore = defineStore('claims', () => {
  const items = ref<Claim[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const keyword = ref('')
  const statusFilter = ref<ClaimStatus | 'ALL'>('ALL')

  const filtered = computed(() => {
    const k = keyword.value.trim().toLowerCase()
    return items.value.filter(
      (c) =>
        (statusFilter.value === 'ALL' || c.status === statusFilter.value) &&
        (!k || c.patientRef.toLowerCase().includes(k) || c.hospital.toLowerCase().includes(k)),
    )
  })
  const counts = computed(() => ({
    ALL: items.value.length,
    PENDING: items.value.filter((c) => c.status === 'PENDING').length,
    APPROVED: items.value.filter((c) => c.status === 'APPROVED').length,
    REJECTED: items.value.filter((c) => c.status === 'REJECTED').length,
  }))
  const byId = (id: number) => items.value.find((c) => c.id === id)

  async function load() {
    loading.value = true
    error.value = null
    try {
      items.value = await claimsApi.list()
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }
  async function add(c: NewClaim) {
    const created = await claimsApi.create(c)
    items.value.unshift(created)
    return created
  }
  async function changeStatus(id: number, status: ClaimStatus) {
    const updated = await claimsApi.updateStatus(id, status)
    const i = items.value.findIndex((c) => c.id === id)
    if (i >= 0) items.value[i] = updated
    return updated
  }
  async function remove(id: number) {
    await claimsApi.remove(id)
    items.value = items.value.filter((c) => c.id !== id)
  }

  return { items, loading, error, keyword, statusFilter, filtered, counts, byId, load, add, changeStatus, remove }
})
