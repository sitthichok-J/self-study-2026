<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import BaseButton from './components/BaseButton.vue'
import ThemeToggle from './components/ThemeToggle.vue'
import { useAuth } from './stores/auth'

const auth = useAuth()
const router = useRouter()

async function onLogout() {
  await auth.logout()
  await router.push({ name: 'login' })
}
</script>

<template>
  <main class="page">
    <header class="head">
      <div>
        <h1>Claim Tracker Lite</h1>
        <p class="sub">Day 11 · Login + Gateway (ข้อมูลสมมติ)</p>
      </div>
      <div class="right">
        <nav v-if="auth.isAuthenticated">
          <RouterLink to="/" exact-active-class="router-link-active">รายการ</RouterLink>
          <RouterLink to="/new" exact-active-class="router-link-active">ส่งเคลม</RouterLink>
          <span class="who" data-test="who">{{ auth.username }}</span>
          <BaseButton variant="ghost" data-test="logout" @click="onLogout">ออกจากระบบ</BaseButton>
        </nav>
        <ThemeToggle />
      </div>
    </header>
    <RouterView />
  </main>
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
.who { color: var(--muted); }
</style>
