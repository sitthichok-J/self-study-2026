<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError, describeError } from '@/api/http'
import BaseButton from '@/components/BaseButton.vue'
import { loginLimitMessage, safeRedirect } from '@/lib'
import { useAuth } from '@/stores/auth'

const auth = useAuth()
const route = useRoute()
const router = useRouter()

const username = ref('')
const password = ref('')
const submitting = ref(false)
const error = ref('')

async function onSubmit() {
  if (submitting.value) return
  error.value = ''
  if (!username.value.trim() || !password.value) {
    error.value = 'กรอกชื่อผู้ใช้และรหัสผ่าน'
    return
  }
  submitting.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    await router.push(safeRedirect(route.query.redirect))
  } catch (e) {
    // 401 จาก /auth/login = ชื่อหรือรหัสผิด (ฝั่ง API ตอบเหมือนกันทั้งสองกรณี ไม่บอกว่าผิดที่ตัวไหน)
    if (e instanceof ApiError && e.status === 401) error.value = 'ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง'
    else if (e instanceof ApiError && e.status === 429) error.value = loginLimitMessage(e.retryAfter)
    else error.value = describeError(e)
    password.value = ''
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="card">
    <h2>เข้าสู่ระบบ</h2>
    <form novalidate @submit.prevent="onSubmit">
      <label>
        ชื่อผู้ใช้
        <input v-model="username" data-test="username" name="username" type="text" autocomplete="username" autocapitalize="none" spellcheck="false" maxlength="64" />
      </label>
      <label>
        รหัสผ่าน
        <input v-model="password" data-test="password" name="password" type="password" autocomplete="current-password" maxlength="256" />
      </label>
      <p v-if="error" class="error" role="alert" data-test="error">{{ error }}</p>
      <BaseButton type="submit" data-test="submit" :disabled="submitting">{{ submitting ? 'กำลังเข้าสู่ระบบ…' : 'เข้าสู่ระบบ' }}</BaseButton>
    </form>
  </section>
</template>

<style scoped>
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; max-width: 380px; }
h2 { margin: 0; font-size: 20px; }
form { display: flex; flex-direction: column; gap: 12px; }
label { display: flex; flex-direction: column; gap: 4px; font-size: 14px; color: var(--muted); }
input { font: inherit; padding: 8px 10px; border-radius: 8px; border: 1px solid var(--border); background: var(--surface); color: var(--text); }
.error { margin: 0; color: var(--danger); font-size: 14px; }
</style>
