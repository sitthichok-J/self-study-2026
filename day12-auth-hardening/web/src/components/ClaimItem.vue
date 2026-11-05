<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { formatBaht } from '@/lib'
import { STATUS_LABEL, type Claim } from '@/types/claim'

// หน้าตาเดียวกับ ClaimItem ของ Day 2 แต่ข้อมูลมาจาก API จริง (snake_case, จำนวนเงินเป็นสตางค์)
// ไม่มีปุ่มอนุมัติ/ลบ เพราะ backend ตัดสินเองผ่าน worker
const props = defineProps<{ claim: Claim }>()
const shortId = computed(() => props.claim.id.slice(0, 8))
</script>

<template>
  <li class="item" data-test="row">
    <div class="main">
      <RouterLink class="ref" :to="`/claims/${claim.id}`">{{ claim.patient_ref }}</RouterLink>
      <span class="sub">รหัส {{ shortId }}</span>
    </div>
    <div class="amount">฿ {{ formatBaht(claim.amount) }}</div>
    <span :class="['badge', `badge--${claim.status}`]">{{ STATUS_LABEL[claim.status] }}</span>
  </li>
</template>

<style scoped>
.item {
  display: grid;
  grid-template-columns: 1fr 160px 120px;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
}
.main { display: flex; flex-direction: column; }
.ref { font-weight: 700; color: var(--text); text-decoration: none; }
.ref:hover { color: var(--accent); text-decoration: underline; }
.sub { font-size: 13px; color: var(--muted); }
.amount { font-variant-numeric: tabular-nums; text-align: right; }
.badge { padding: 3px 10px; border-radius: 999px; font-size: 12px; text-align: center; }
.badge--PENDING { background: var(--warn-bg); color: var(--warn); }
.badge--APPROVED { background: var(--ok-bg); color: var(--ok); }
.badge--REJECTED { background: var(--bad-bg); color: var(--danger); }
@media (max-width: 760px) { .item { grid-template-columns: 1fr 1fr; } }
</style>
