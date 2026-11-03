<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import { getClaim } from '@/api/claims'
import { usePolling } from '@/composables/usePolling'
import { formatBaht } from '@/lib'
import { isFinal, REASON_LABEL, STATUS_LABEL } from '@/types/claim'

const props = defineProps<{ id: string }>()

// worker ตัดสินแบบ async จึง poll ทุก 2 วินาทีจนสถานะไม่ใช่ PENDING
const { data: claim, error, loading, running, restart } = usePolling(() => getClaim(props.id), {
  intervalMs: 2000,
  isDone: (c) => isFinal(c.status),
})

const reason = computed(() => {
  const r = claim.value?.decision_reason
  return r ? (REASON_LABEL[r] ?? r) : ''
})
</script>

<template>
  <section class="card">
    <p v-if="loading" class="hint" data-test="loading">กำลังโหลด…</p>
    <p v-if="error" class="error" role="alert" data-test="error">
      {{ error }} <button type="button" class="retry" data-test="retry" @click="restart">ลองใหม่</button>
    </p>

    <template v-if="claim">
      <h2>{{ claim.patient_ref }}</h2>
      <dl>
        <dt>รหัส</dt><dd data-test="id">{{ claim.id }}</dd>
        <dt>จำนวน (บาท)</dt><dd>฿ {{ formatBaht(claim.amount) }}</dd>
        <dt>สถานะ</dt>
        <dd>
          <span data-test="status" :class="['badge', `badge--${claim.status}`]">{{ STATUS_LABEL[claim.status] }}</span>
          <span v-if="running" data-test="polling" class="hint"> กำลังรอผลการประมวลผล…</span>
        </dd>
        <template v-if="reason"><dt>เหตุผล</dt><dd data-test="reason">{{ reason }}</dd></template>
      </dl>
    </template>
    <div class="actions"><RouterLink to="/"><BaseButton variant="ghost">← กลับไปรายการ</BaseButton></RouterLink></div>
  </section>
</template>

<style scoped>
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; }
h2 { margin: 0; font-size: 20px; }
dl { display: grid; grid-template-columns: 120px 1fr; gap: 6px 12px; margin: 0; }
dt { color: var(--muted); }
dd { margin: 0; word-break: break-all; }
.badge { padding: 3px 10px; border-radius: 999px; font-size: 12px; }
.badge--PENDING { background: var(--warn-bg); color: var(--warn); }
.badge--APPROVED { background: var(--ok-bg); color: var(--ok); }
.badge--REJECTED { background: var(--bad-bg); color: var(--danger); }
.retry { font: inherit; cursor: pointer; }
.actions a { text-decoration: none; }
</style>
