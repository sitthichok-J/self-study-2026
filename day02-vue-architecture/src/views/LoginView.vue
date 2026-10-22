<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BaseButton from '../components/BaseButton.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const username = ref('')
const password = ref('')
const error = ref('')
const busy = ref(false)

async function submit() {
  error.value = ''
  busy.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    await router.replace(redirect.startsWith('/') ? redirect : { name: 'claims' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <form class="card" @submit.prevent="submit">
    <h2>เข้าสู่ระบบ</h2>
    <p class="hint">บัญชีสาธิต: demo / demo1234 (mock API)</p>
    <label><span>ชื่อผู้ใช้</span><input v-model="username" autocomplete="username" /></label>
    <label><span>รหัสผ่าน</span><input v-model="password" type="password" autocomplete="current-password" /></label>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <BaseButton type="submit">{{ busy ? 'กำลังเข้าสู่ระบบ…' : 'เข้าสู่ระบบ' }}</BaseButton>
  </form>
</template>

<style scoped>
.card { max-width: 360px; margin: 24px auto; background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 24px; display: flex; flex-direction: column; gap: 12px; }
h2 { margin: 0; font-size: 18px; }
.hint { margin: 0; color: var(--muted); font-size: 13px; }
label { display: flex; flex-direction: column; gap: 4px; font-size: 13px; color: var(--muted); }
input { padding: 8px 10px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); color: var(--text); font: inherit; }
.error { margin: 0; color: var(--danger); font-size: 13px; }
</style>
