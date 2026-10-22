<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import ThemeToggle from './components/ThemeToggle.vue'
import { provideToast } from './composables/useToast'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
const router = useRouter()
const { toasts } = provideToast()

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <main class="page">
    <header class="head">
      <div>
        <h1>Claim Console</h1>
        <p class="sub">Day 2 · Vue 3 ขั้นกลาง — Router, Pinia, composable, API (ข้อมูลสมมติ)</p>
      </div>
      <div class="right">
        <nav v-if="auth.isLoggedIn">
          <RouterLink :to="{ name: 'claims' }">รายการ</RouterLink>
          <RouterLink :to="{ name: 'claim-new' }">เพิ่มเคลม</RouterLink>
          <button class="link" @click="logout">ออกจากระบบ</button>
        </nav>
        <ThemeToggle />
      </div>
    </header>
    <RouterView />
  </main>
  <div class="toasts" aria-live="polite">
    <div v-for="t in toasts" :key="t.id" class="toast">{{ t.text }}</div>
  </div>
</template>

<style scoped>
.page { max-width: 880px; margin: 0 auto; padding: 32px 16px 56px; display: flex; flex-direction: column; gap: 20px; }
.head { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; flex-wrap: wrap; }
h1 { margin: 0; font-size: 28px; }
.sub { margin: 4px 0 0; color: var(--muted); font-size: 14px; }
.right { display: flex; flex-direction: column; align-items: flex-end; gap: 8px; }
nav { display: flex; gap: 14px; align-items: center; font-size: 14px; }
nav a { color: var(--muted); text-decoration: none; }
nav a.router-link-active { color: var(--accent); font-weight: 600; }
.link { background: none; border: 0; color: var(--muted); font: inherit; font-size: 14px; cursor: pointer; padding: 0; }
.toasts { position: fixed; right: 16px; bottom: 16px; display: flex; flex-direction: column; gap: 8px; }
.toast { background: var(--text); color: var(--bg); padding: 8px 14px; border-radius: 8px; font-size: 14px; }
</style>
