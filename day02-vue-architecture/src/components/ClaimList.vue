<script setup lang="ts">
import type { Claim, ClaimStatus } from '../types/claim'
import ClaimItem from './ClaimItem.vue'

// Day 1: component รายการ + named slot "empty"
defineProps<{ claims: Claim[] }>()
const emit = defineEmits<{
  'change-status': [id: number, status: ClaimStatus]
  remove: [id: number]
}>()
</script>

<template>
  <ul v-if="claims.length" class="list">
    <ClaimItem
      v-for="c in claims"
      :key="c.id"
      :claim="c"
      @change-status="(id, s) => emit('change-status', id, s)"
      @remove="(id) => emit('remove', id)"
    />
  </ul>
  <div v-else class="empty">
    <slot name="empty">ไม่มีรายการ</slot>
  </div>
</template>

<style scoped>
.list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 8px; }
.empty {
  padding: 32px;
  text-align: center;
  color: var(--muted);
  border: 1px dashed var(--border);
  border-radius: 10px;
}
</style>
