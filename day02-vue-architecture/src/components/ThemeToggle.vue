<script setup lang="ts">
import { ref, watchEffect } from 'vue'

type Theme = 'system' | 'light' | 'dark'
const KEY = 'claim-list-theme'

function load(): Theme {
  try {
    const v = localStorage.getItem(KEY)
    if (v === 'light' || v === 'dark') return v
  } catch { /* localStorage ใช้ไม่ได้ ก็ใช้ system */ }
  return 'system'
}

// state ของธีมอยู่ในคอมโพเนนต์นี้เอง (อ่านค่าเริ่มต้นจาก localStorage)
const theme = ref<Theme>(load())

// watchEffect: ตั้ง data-theme บน <html> ทุกครั้งที่ theme เปลี่ยน แล้วจำค่าไว้
watchEffect(() => {
  const root = document.documentElement
  if (theme.value === 'system') root.removeAttribute('data-theme')
  else root.setAttribute('data-theme', theme.value)
  try {
    if (theme.value === 'system') localStorage.removeItem(KEY)
    else localStorage.setItem(KEY, theme.value)
  } catch { /* ignore */ }
})

const options: Array<{ key: Theme; label: string }> = [
  { key: 'system', label: 'ตามระบบ' },
  { key: 'light', label: 'สว่าง' },
  { key: 'dark', label: 'มืด' },
]
</script>

<template>
  <div class="theme" role="group" aria-label="โหมดสี">
    <button
      v-for="o in options"
      :key="o.key"
      :class="{ active: theme === o.key }"
      :aria-pressed="theme === o.key"
      @click="theme = o.key"
    >
      {{ o.label }}
    </button>
  </div>
</template>

<style scoped>
.theme { display: inline-flex; border: 1px solid var(--border); border-radius: 999px; overflow: hidden; }
button { border: 0; background: transparent; color: var(--muted); padding: 5px 12px; font: inherit; font-size: 13px; cursor: pointer; }
button.active { background: var(--accent); color: #fff; }
</style>
