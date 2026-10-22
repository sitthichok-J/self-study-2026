<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { onMounted, toRef } from 'vue'
import { RouterLink } from 'vue-router'
import BaseButton from '../components/BaseButton.vue'
import ClaimList from '../components/ClaimList.vue'
import SearchBox from '../components/SearchBox.vue'
import { usePagination } from '../composables/usePagination'
import { useToast } from '../composables/useToast'
import { useClaimsStore } from '../stores/claims'
import { STATUS_LABEL, type ClaimStatus } from '../types/claim'

// layout เดียวกับ Day 1 แต่ state ย้ายจาก ref ใน App.vue ไปอยู่ใน Pinia store
const store = useClaimsStore()
const toast = useToast()
const { filtered, loading, error, keyword, statusFilter, counts } = storeToRefs(store)
const { page, pageCount, paged, next, prev } = usePagination(toRef(() => filtered.value), 5)

onMounted(() => { if (!store.items.length) store.load() })

const totalAmount = () => filtered.value.reduce((s, c) => s + c.amount, 0)

async function run(action: () => Promise<unknown>, okText: string) {
  try {
    await action()
    toast.show(okText)
  } catch (e) {
    toast.show(e instanceof Error ? e.message : String(e))
  }
}
const changeStatus = (id: number, s: ClaimStatus) =>
  run(() => store.changeStatus(id, s), `เปลี่ยนสถานะเป็น "${STATUS_LABEL[s]}" แล้ว`)
const remove = (id: number) => run(() => store.remove(id), 'ลบแล้ว')

const tabs: Array<{ key: ClaimStatus | 'ALL'; label: string }> = [
  { key: 'ALL', label: 'ทั้งหมด' },
  ...(Object.keys(STATUS_LABEL) as ClaimStatus[]).map((k) => ({ key: k, label: STATUS_LABEL[k] })),
]
</script>

<template>
  <section class="stats">
    <div class="stat"><span>ทั้งหมด</span><b>{{ counts.ALL }}</b></div>
    <div class="stat"><span>รอตรวจสอบ</span><b>{{ counts.PENDING }}</b></div>
    <div class="stat"><span>อนุมัติ</span><b>{{ counts.APPROVED }}</b></div>
    <div class="stat"><span>ปฏิเสธ</span><b>{{ counts.REJECTED }}</b></div>
  </section>

  <section class="card">
    <div class="toolbar">
      <div class="row">
        <SearchBox v-model="keyword" />
        <RouterLink :to="{ name: 'claim-new' }" class="add"><BaseButton>+ เพิ่มเคลม</BaseButton></RouterLink>
      </div>
      <div class="tabs">
        <button v-for="t in tabs" :key="t.key" :class="{ active: statusFilter === t.key }" @click="statusFilter = t.key">
          {{ t.label }}
        </button>
      </div>
    </div>

    <p v-if="loading" class="state">กำลังโหลด…</p>
    <p v-else-if="error" class="state error" role="alert">
      {{ error }} <button class="retry" @click="store.load()">ลองใหม่</button>
    </p>
    <ClaimList v-else :claims="paged" @change-status="changeStatus" @remove="remove">
      <template #empty>ไม่พบรายการที่ตรงกับเงื่อนไข</template>
    </ClaimList>

    <div v-if="pageCount > 1" class="pager">
      <button :disabled="page === 1" @click="prev">ก่อนหน้า</button>
      <span>หน้า {{ page }} / {{ pageCount }}</span>
      <button :disabled="page === pageCount" @click="next">ถัดไป</button>
    </div>

    <p class="total">
      แสดง {{ filtered.length }} รายการ · รวม
      <b>฿ {{ totalAmount().toLocaleString('th-TH', { minimumFractionDigits: 2 }) }}</b>
    </p>
  </section>
</template>

<style scoped>
.stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
.stat { background: var(--surface); border: 1px solid var(--border); border-radius: 10px; padding: 12px 14px; display: flex; flex-direction: column; }
.stat span { font-size: 13px; color: var(--muted); }
.stat b { font-size: 24px; }
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; }
.toolbar { display: flex; flex-direction: column; gap: 10px; }
.row { display: flex; gap: 10px; align-items: center; }
.row > :first-child { flex: 1; }
.add { text-decoration: none; }
.tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.tabs button { border: 1px solid var(--border); background: transparent; color: var(--muted); border-radius: 999px; padding: 5px 14px; font: inherit; font-size: 13px; cursor: pointer; }
.tabs button.active { background: var(--accent); border-color: var(--accent); color: #fff; }
.state { margin: 0; color: var(--muted); }
.error { color: var(--danger); }
.retry { font: inherit; cursor: pointer; }
.pager { display: flex; justify-content: center; align-items: center; gap: 12px; color: var(--muted); font-size: 14px; }
.total { margin: 0; text-align: right; color: var(--muted); font-size: 14px; }
.total b { color: var(--text); }
@media (max-width: 640px) { .stats { grid-template-columns: repeat(2, 1fr); } }
</style>
