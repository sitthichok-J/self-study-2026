<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { createClaim } from '@/api/claims'
import BaseButton from '@/components/BaseButton.vue'
import { describeError } from '@/api/http'
import { newIdempotencyKey, validateClaimForm, type FormErrors } from '@/lib'

const router = useRouter()
const form = reactive({ patient_ref: 'FAKE-', amount: '' })
const errors = ref<FormErrors>({})
const submitting = ref(false)
const submitError = ref('')
let attempted = false // หลังกดส่งครั้งแรก ให้ตรวจซ้ำทุกครั้งที่แก้ ข้อความ error จะได้หายเมื่อแก้ถูก

// key เดียวต่อหนึ่งชุดข้อมูล: กดซ้ำหรือ retry หลังเน็ตหลุดจะได้ claim เดิม ไม่เกิดซ้ำ
// เปลี่ยนข้อมูลเมื่อไหร่ถือเป็นคำขอใหม่ จึงออก key ใหม่
let key = newIdempotencyKey()
watch(form, () => {
  key = newIdempotencyKey()
  if (attempted) errors.value = validateClaimForm(form.patient_ref, form.amount).errors
})

async function onSubmit() {
  if (submitting.value) return
  submitError.value = ''
  attempted = true
  const v = validateClaimForm(form.patient_ref, form.amount)
  errors.value = v.errors
  if (v.satang === null) return

  submitting.value = true
  try {
    const { claim } = await createClaim({ patient_ref: form.patient_ref.trim(), amount: v.satang }, key)
    await router.push(`/claims/${claim.id}`)
  } catch (e) {
    submitError.value = describeError(e)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="card">
    <h2>ส่งเคลมใหม่</h2>
    <p class="hint">ใช้ข้อมูลสมมติเท่านั้น รหัสผู้ป่วยต้องขึ้นต้นด้วย FAKE-</p>
    <form novalidate @submit.prevent="onSubmit">
      <div class="fields">
        <label>
          <span>รหัสผู้ป่วย (สมมติ)</span>
          <input v-model="form.patient_ref" data-test="patient" autocomplete="off" :aria-invalid="!!errors.patient_ref" />
        </label>
        <label>
          <span>จำนวนเงิน (บาท)</span>
          <input v-model="form.amount" data-test="amount" inputmode="decimal" autocomplete="off" placeholder="เช่น 1500.50" :aria-invalid="!!errors.amount" />
        </label>
      </div>
      <p v-if="errors.patient_ref" class="error" data-test="patient-error">{{ errors.patient_ref }}</p>
      <p v-if="errors.amount" class="error" data-test="amount-error">{{ errors.amount }}</p>
      <p v-if="submitError" class="error" role="alert" data-test="submit-error">{{ submitError }}</p>
      <BaseButton type="submit" data-test="submit" :disabled="submitting">{{ submitting ? 'กำลังส่ง…' : '+ ส่งเคลม' }}</BaseButton>
    </form>
  </section>
</template>

<style scoped>
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; }
h2 { margin: 0; font-size: 16px; }
form { display: flex; flex-direction: column; gap: 12px; }
.fields { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
label { display: flex; flex-direction: column; gap: 4px; font-size: 13px; color: var(--muted); }
input { padding: 8px 10px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); color: var(--text); font: inherit; }
input:focus { outline: 2px solid var(--accent); outline-offset: -1px; }
input[aria-invalid='true'] { border-color: var(--danger); }
form > :last-child { align-self: flex-start; }
@media (max-width: 640px) { .fields { grid-template-columns: 1fr; } }
</style>
