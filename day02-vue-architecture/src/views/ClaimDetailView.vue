<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { claimsApi } from '../api/claimsApi'
import BaseButton from '../components/BaseButton.vue'
import { useToast } from '../composables/useToast'
import { useClaimsStore } from '../stores/claims'
import { STATUS_LABEL, type Claim } from '../types/claim'

// route พร้อม props: id มาจาก path (ดู router/index.ts)
const props = defineProps<{ id: number }>()
const store = useClaimsStore()
const router = useRouter()
const toast = useToast()

// ถ้า store มีข้อมูลแล้วใช้ของใน store ไม่งั้นดึงจาก API (เช่นเปิดลิงก์ตรง)
const fetched = ref<Claim | null>(null)
const error = ref('')
const loading = ref(false)
const claim = computed(() => store.byId(props.id) ?? fetched.value)

onMounted(async () => {
  if (store.byId(props.id)) return
  loading.value = true
  try {
    fetched.value = await claimsApi.get(props.id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
})

async function setStatus(status: Claim['status']) {
  try {
    const updated = await claimsApi.updateStatus(props.id, status)
    if (store.byId(props.id)) await store.changeStatus(props.id, status)
    else fetched.value = updated
    toast.show(`เปลี่ยนสถานะเป็น "${STATUS_LABEL[status]}" แล้ว`)
  } catch (e) {
    toast.show(e instanceof Error ? e.message : String(e))
  }
}
async function remove() {
  try {
    await store.remove(props.id)
    toast.show('ลบแล้ว')
    router.push({ name: 'claims' })
  } catch (e) {
    toast.show(e instanceof Error ? e.message : String(e))
  }
}
</script>

<template>
  <section class="card">
    <p v-if="loading">กำลังโหลด…</p>
    <p v-else-if="error" class="error" role="alert">{{ error }}</p>
    <template v-else-if="claim">
      <h2>{{ claim.patientRef }}</h2>
      <dl>
        <dt>โรงพยาบาล</dt><dd>{{ claim.hospital }}</dd>
        <dt>จำนวนเงิน</dt><dd>฿ {{ claim.amount.toLocaleString('th-TH', { minimumFractionDigits: 2 }) }}</dd>
        <dt>สถานะ</dt><dd>{{ STATUS_LABEL[claim.status] }}</dd>
      </dl>
      <div class="actions">
        <BaseButton v-if="claim.status !== 'APPROVED'" @click="setStatus('APPROVED')">อนุมัติ</BaseButton>
        <BaseButton v-if="claim.status !== 'REJECTED'" variant="ghost" @click="setStatus('REJECTED')">ปฏิเสธ</BaseButton>
        <BaseButton variant="danger" @click="remove">ลบ</BaseButton>
        <BaseButton variant="ghost" @click="router.push({ name: 'claims' })">กลับ</BaseButton>
      </div>
    </template>
  </section>
</template>

<style scoped>
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; }
h2 { margin: 0; font-size: 20px; }
dl { display: grid; grid-template-columns: 120px 1fr; gap: 6px 12px; margin: 0 0 16px; }
dt { color: var(--muted); }
dd { margin: 0; }
.actions { display: flex; gap: 8px; flex-wrap: wrap; }
.error { color: var(--danger); }
</style>
