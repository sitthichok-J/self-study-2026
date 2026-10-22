<template>
  <section>
    <h2>ไม่พบหน้านี้</h2>
    <RouterLink :to="{ name: 'claims' }">กลับไปรายการเคลม</RouterLink>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>
