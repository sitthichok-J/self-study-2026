<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import ClaimForm from '../components/ClaimForm.vue'
import { useToast } from '../composables/useToast'
import { useClaimsStore } from '../stores/claims'
import type { NewClaim } from '../types/claim'

// ฟอร์มเป็น component เดียวกับ Day 1 (ClaimForm) หน้านี้แค่ต่อกับ store + router
const store = useClaimsStore()
const router = useRouter()
const toast = useToast()
const serverError = ref('')

async function onSubmit(payload: NewClaim) {
  serverError.value = ''
  try {
    const created = await store.add(payload)
    toast.show(`เพิ่มเคลม ${created.patientRef} แล้ว`)
    await router.push({ name: 'claim-detail', params: { id: created.id } })
  } catch (e) {
    serverError.value = e instanceof Error ? e.message : String(e)
  }
}
</script>

<template>
  <section class="card">
    <h2>เพิ่มเคลมใหม่</h2>
    <ClaimForm @submit="onSubmit" />
    <p v-if="serverError" class="error" role="alert">{{ serverError }}</p>
  </section>
</template>

<style scoped>
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; }
h2 { margin: 0; font-size: 16px; }
.error { margin: 0; color: var(--danger); font-size: 13px; }
</style>
