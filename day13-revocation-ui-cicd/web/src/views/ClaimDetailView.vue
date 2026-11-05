<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import { decideClaim, getClaim, type Decision } from '@/api/claims'
import { ApiError, describeError } from '@/api/http'
import { usePolling } from '@/composables/usePolling'
import { formatBaht } from '@/lib'
import { useAuth } from '@/stores/auth'
import { isSettled, REASON_LABEL, STATUS_LABEL } from '@/types/claim'

const props = defineProps<{ id: string }>()

// worker ตัดสินแบบ async จึง poll ทุก 2 วินาทีจนสถานะไม่ใช่ PENDING
const { data: claim, error, loading, running, restart } = usePolling(() => getClaim(props.id), {
  intervalMs: 2000,
  isDone: isSettled,
})

const auth = useAuth()
// ปุ่มโผล่เมื่อเป็น staff และเคลมยัง PENDING เท่านั้น (ซ่อนปุ่มไม่ใช่การป้องกัน: API ตรวจสิทธิ์เอง)
const canDecide = computed(() => auth.isStaff && claim.value?.status === 'PENDING')

// กดครั้งแรกแค่ถามยืนยัน กดครั้งที่สองถึงยิงจริง (การตัดสินแก้กลับไม่ได้)
const confirming = ref<Decision | null>(null)
const deciding = ref(false)
const decideError = ref('')

async function onDecide(action: Decision) {
  if (deciding.value) return
  if (confirming.value !== action) {
    confirming.value = action
    return
  }
  deciding.value = true
  decideError.value = ''
  try {
    claim.value = await decideClaim(props.id, action)
    confirming.value = null
  } catch (e) {
    confirming.value = null
    if (e instanceof ApiError && e.status === 409) {
      decideError.value = 'มีคนตัดสินเคลมนี้ไปแล้ว'
      restart() // โหลดสถานะล่าสุดมาแสดง
    } else if (e instanceof ApiError && e.status === 403) {
      decideError.value = 'ไม่มีสิทธิ์ตัดสินเคลม (เฉพาะเจ้าหน้าที่)'
    } else {
      decideError.value = describeError(e)
    }
  } finally {
    deciding.value = false
  }
}

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
    <div v-if="canDecide" class="decide" data-test="decide">
      <p class="hint">เคลมนี้รอเจ้าหน้าที่ตัดสิน</p>
      <div class="row">
        <BaseButton data-test="approve" :disabled="deciding" @click="onDecide('approve')">
          {{ confirming === 'approve' ? 'ยืนยันอนุมัติ' : 'อนุมัติ' }}
        </BaseButton>
        <BaseButton variant="danger" data-test="reject" :disabled="deciding" @click="onDecide('reject')">
          {{ confirming === 'reject' ? 'ยืนยันไม่อนุมัติ' : 'ไม่อนุมัติ' }}
        </BaseButton>
        <BaseButton v-if="confirming" variant="ghost" data-test="cancel" :disabled="deciding" @click="confirming = null">ยกเลิก</BaseButton>
      </div>
    </div>
    <p v-if="decideError" class="error" role="alert" data-test="decide-error">{{ decideError }}</p>
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
.decide { display: flex; flex-direction: column; gap: 8px; }
.decide .row { display: flex; gap: 10px; flex-wrap: wrap; }
.retry { font: inherit; cursor: pointer; }
.actions a { text-decoration: none; }
</style>
