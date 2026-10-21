<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { STATUS_LABEL, type Claim, type ClaimStatus } from './types'
import ClaimForm from './components/ClaimForm.vue'
import ClaimList from './components/ClaimList.vue'
import SearchBox from './components/SearchBox.vue'

// ข้อมูลสมมติทั้งหมด (ไม่ใช่ข้อมูลจริง)
const claims = ref<Claim[]>([
  { id: 1, patientRef: 'P-1001', hospital: 'รพ.ตัวอย่าง A', amount: 12500, status: 'PENDING' },
  { id: 2, patientRef: 'P-1002', hospital: 'รพ.ตัวอย่าง B', amount: 8200.5, status: 'APPROVED' },
  { id: 3, patientRef: 'P-1003', hospital: 'รพ.ตัวอย่าง A', amount: 31000, status: 'REJECTED' },
  { id: 4, patientRef: 'P-1004', hospital: 'รพ.ตัวอย่าง C', amount: 4300, status: 'PENDING' },
])
const keyword = ref('')
const statusFilter = ref<ClaimStatus | 'ALL'>('ALL')
let nextId = 5

// computed: กรอง + สรุป (cache และ track dependency อัตโนมัติ)
const filtered = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  return claims.value.filter(
    (c) =>
      (statusFilter.value === 'ALL' || c.status === statusFilter.value) &&
      (!k || c.patientRef.toLowerCase().includes(k) || c.hospital.toLowerCase().includes(k)),
  )
})
const totalAmount = computed(() => filtered.value.reduce((s, c) => s + c.amount, 0))
const count = (s: ClaimStatus) => claims.value.filter((c) => c.status === s).length

// watch: log เมื่อค่าค้นหาเปลี่ยน / lifecycle
watch(keyword, (now, before) => console.log(`keyword: "${before}" -> "${now}"`))
onMounted(() => console.log('App mounted, claims =', claims.value.length))

function addClaim(p: { patientRef: string; hospital: string; amount: number }) {
  claims.value.unshift({ id: nextId++, status: 'PENDING', ...p })
}
function changeStatus(id: number, status: ClaimStatus) {
  const c = claims.value.find((x) => x.id === id)
  if (c) c.status = status
}
function remove(id: number) {
  claims.value = claims.value.filter((c) => c.id !== id)
}

const tabs: Array<{ key: ClaimStatus | 'ALL'; label: string }> = [
  { key: 'ALL', label: 'ทั้งหมด' },
  ...(Object.keys(STATUS_LABEL) as ClaimStatus[]).map((k) => ({ key: k, label: STATUS_LABEL[k] })),
]
</script>

<template>
  <main class="page">
    <header>
      <h1>Claim List</h1>
      <p class="sub">Day 1 · Vue 3 พื้นฐาน — ref, computed, watch, props/emits, slots (ข้อมูลสมมติ)</p>
    </header>

    <section class="stats">
      <div class="stat"><span>ทั้งหมด</span><b>{{ claims.length }}</b></div>
      <div class="stat"><span>รอตรวจสอบ</span><b>{{ count('PENDING') }}</b></div>
      <div class="stat"><span>อนุมัติ</span><b>{{ count('APPROVED') }}</b></div>
      <div class="stat"><span>ปฏิเสธ</span><b>{{ count('REJECTED') }}</b></div>
    </section>

    <section class="card">
      <h2>เพิ่มเคลมใหม่</h2>
      <ClaimForm @submit="addClaim" />
    </section>

    <section class="card">
      <div class="toolbar">
        <SearchBox v-model="keyword" />
        <div class="tabs">
          <button
            v-for="t in tabs"
            :key="t.key"
            :class="{ active: statusFilter === t.key }"
            @click="statusFilter = t.key"
          >
            {{ t.label }}
          </button>
        </div>
      </div>

      <ClaimList :claims="filtered" @change-status="changeStatus" @remove="remove">
        <template #empty>ไม่พบรายการที่ตรงกับเงื่อนไข</template>
      </ClaimList>

      <p class="total">
        แสดง {{ filtered.length }} รายการ · รวม
        <b>฿ {{ totalAmount.toLocaleString('th-TH', { minimumFractionDigits: 2 }) }}</b>
      </p>
    </section>
  </main>
</template>

<style scoped>
.page { max-width: 880px; margin: 0 auto; padding: 32px 16px 56px; display: flex; flex-direction: column; gap: 20px; }
h1 { margin: 0; font-size: 28px; }
h2 { margin: 0 0 12px; font-size: 16px; }
.sub { margin: 4px 0 0; color: var(--muted); font-size: 14px; }
.stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
.stat { background: var(--surface); border: 1px solid var(--border); border-radius: 10px; padding: 12px 14px; display: flex; flex-direction: column; }
.stat span { font-size: 13px; color: var(--muted); }
.stat b { font-size: 24px; }
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; }
.card h2 { margin-bottom: 0; }
.toolbar { display: flex; flex-direction: column; gap: 10px; }
.tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.tabs button { border: 1px solid var(--border); background: transparent; color: var(--muted); border-radius: 999px; padding: 5px 14px; font: inherit; font-size: 13px; cursor: pointer; }
.tabs button.active { background: var(--accent); border-color: var(--accent); color: #fff; }
.total { margin: 0; text-align: right; color: var(--muted); font-size: 14px; }
.total b { color: var(--text); }
@media (max-width: 640px) { .stats { grid-template-columns: repeat(2, 1fr); } }
</style>
