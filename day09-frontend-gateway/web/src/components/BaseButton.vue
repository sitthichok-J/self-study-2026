<script setup lang="ts">
// Day 1: props พร้อม type + default slot
withDefaults(
  defineProps<{
    variant?: 'primary' | 'danger' | 'ghost'
    type?: 'button' | 'submit'
  }>(),
  { variant: 'primary', type: 'button' },
)
</script>

<template>
  <button :type="type" :class="['btn', `btn--${variant}`]">
    <slot />
  </button>
</template>

<style scoped>
.btn {
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 6px 14px;
  font: inherit;
  font-size: 14px;
  cursor: pointer;
}
.btn--primary { background: var(--accent); color: #fff; }
.btn--primary:hover { filter: brightness(1.1); }
.btn--danger { background: transparent; color: var(--danger); border-color: var(--danger); }
.btn--danger:hover { background: var(--danger); color: #fff; }
.btn--ghost { background: transparent; color: var(--muted); border-color: var(--border); }
.btn--ghost:hover { color: var(--text); border-color: var(--text); }
</style>
