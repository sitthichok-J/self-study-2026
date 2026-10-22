<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { STATUS_LABEL, type Claim, type ClaimStatus } from '../types/claim'
import BaseButton from './BaseButton.vue'

// ต่อยอดจาก Day 1: เพิ่มลิงก์รหัสผู้ป่วยไปหน้า detail (RouterLink)
const props = defineProps<{ claim: Claim }>()
const emit = defineEmits<{
  'change-status': [id: number, status: ClaimStatus]
  remove: [id: number]
}>()

const amountText = computed(() =>
  props.claim.amount.toLocaleString('th-TH', { minimumFractionDigits: 2 }),
)
</script>

<template>
  <li class="item">
    <div class="main">
      <RouterLink class="ref" :to="{ name: 'claim-detail', params: { id: claim.id } }">{{ claim.patientRef }}</RouterLink>
      <span class="hospital">{{ claim.hospital }}</span>
    </div>
    <div class="amount">฿ {{ amountText }}</div>
    <span :class="['badge', `badge--${claim.status}`]">
      {{ STATUS_LABEL[claim.status] }}
    </span>
    <div class="actions">
      <!-- v-if / v-else-if: ปุ่มเปลี่ยนตามสถานะ -->
      <template v-if="claim.status === 'PENDING'">
        <BaseButton @click="emit('change-status', claim.id, 'APPROVED')">อนุมัติ</BaseButton>
        <BaseButton variant="ghost" @click="emit('change-status', claim.id, 'REJECTED')">
          ปฏิเสธ
        </BaseButton>
      </template>
      <BaseButton v-else variant="ghost" @click="emit('change-status', claim.id, 'PENDING')">
        ย้อนเป็นรอตรวจสอบ
      </BaseButton>
      <BaseButton variant="danger" @click="emit('remove', claim.id)">ลบ</BaseButton>
    </div>
  </li>
</template>

<style scoped>
.item {
  display: grid;
  grid-template-columns: 1fr 140px 120px auto;
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
.hospital { font-size: 13px; color: var(--muted); }
.amount { font-variant-numeric: tabular-nums; text-align: right; }
.badge { padding: 3px 10px; border-radius: 999px; font-size: 12px; text-align: center; }
.badge--PENDING { background: var(--warn-bg); color: var(--warn); }
.badge--APPROVED { background: var(--ok-bg); color: var(--ok); }
.badge--REJECTED { background: var(--bad-bg); color: var(--danger); }
.actions { display: flex; gap: 6px; justify-content: flex-end; }
@media (max-width: 760px) {
  .item { grid-template-columns: 1fr 1fr; }
  .actions { grid-column: 1 / -1; justify-content: flex-start; }
}
</style>
