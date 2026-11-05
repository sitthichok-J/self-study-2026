<script setup lang="ts">
import type { Claim } from '@/types/claim'
import ClaimItem from './ClaimItem.vue'

// เหมือน ClaimList ของ Day 2: รายการ + named slot "empty"
defineProps<{ claims: Claim[] }>()
</script>

<template>
  <ul v-if="claims.length" class="list">
    <ClaimItem v-for="c in claims" :key="c.id" :claim="c" />
  </ul>
  <div v-else class="empty" data-test="empty">
    <slot name="empty">ไม่มีรายการ</slot>
  </div>
</template>

<style scoped>
.list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 8px; }
.empty { padding: 32px; text-align: center; color: var(--muted); border: 1px dashed var(--border); border-radius: 10px; }
</style>
