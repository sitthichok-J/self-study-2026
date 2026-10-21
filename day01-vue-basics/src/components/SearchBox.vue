<script setup lang="ts">
// Day 1: defineModel = v-model บน component (two-way binding)
const model = defineModel<string>({ default: '' })
</script>

<template>
  <input
    v-model="model"
    class="search"
    type="search"
    placeholder="ค้นหา รหัสผู้ป่วย / โรงพยาบาล…"
  />
</template>

<style scoped>
.search {
  width: 100%;
  padding: 9px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  font: inherit;
}
.search:focus { outline: 2px solid var(--accent); outline-offset: -1px; }
</style>
