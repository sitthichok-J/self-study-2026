<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import BaseButton from './BaseButton.vue'

// Day 1: emits พร้อม type, reactive object, computed validation
const emit = defineEmits<{
  submit: [payload: { patientRef: string; hospital: string; amount: number }]
}>()

const form = reactive({ patientRef: '', hospital: '', amount: 0 })
// แสดง error หลังกดส่งครั้งแรกเท่านั้น ไม่ให้ผู้ใช้เห็นข้อความแดงตั้งแต่เปิดหน้า
const submitted = ref(false)

const errors = computed(() => {
  const e: string[] = []
  if (!form.patientRef.trim()) e.push('กรุณากรอกรหัสผู้ป่วย (ข้อมูลสมมติ)')
  if (!form.hospital.trim()) e.push('กรุณากรอกชื่อโรงพยาบาล')
  if (!(form.amount > 0)) e.push('จำนวนเงินต้องมากกว่า 0')
  return e
})

function onSubmit() {
  submitted.value = true
  if (errors.value.length) return
  emit('submit', {
    patientRef: form.patientRef.trim(),
    hospital: form.hospital.trim(),
    amount: Number(form.amount),
  })
  form.patientRef = ''
  form.hospital = ''
  form.amount = 0
  submitted.value = false
}
</script>

<template>
  <form class="form" @submit.prevent="onSubmit">
    <div class="fields">
      <label>
        <span>รหัสผู้ป่วย</span>
        <input v-model.trim="form.patientRef" placeholder="เช่น P-1004" />
      </label>
      <label>
        <span>โรงพยาบาล</span>
        <input v-model.trim="form.hospital" placeholder="เช่น รพ.ตัวอย่าง" />
      </label>
      <label>
        <span>จำนวนเงิน (บาท)</span>
        <input v-model.number="form.amount" type="number" min="0" />
      </label>
    </div>
    <ul v-if="submitted && errors.length" class="errors">
      <li v-for="msg in errors" :key="msg">{{ msg }}</li>
    </ul>
    <BaseButton type="submit">+ เพิ่มเคลม</BaseButton>
  </form>
</template>

<style scoped>
.form { display: flex; flex-direction: column; gap: 12px; }
.fields { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
label { display: flex; flex-direction: column; gap: 4px; font-size: 13px; color: var(--muted); }
input {
  padding: 8px 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  font: inherit;
}
input:focus { outline: 2px solid var(--accent); outline-offset: -1px; }
.errors { margin: 0; padding-left: 18px; color: var(--danger); font-size: 13px; }
form > :last-child { align-self: flex-start; }
@media (max-width: 640px) { .fields { grid-template-columns: 1fr; } }
</style>
