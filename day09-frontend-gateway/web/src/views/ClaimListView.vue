<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import ClaimList from '@/components/ClaimList.vue'
import { listClaims } from '@/api/claims'
import { useLoad } from '@/composables/useLoad'
import { useFilters } from '@/stores/filters'
import { STATUSES, STATUS_LABEL, type ClaimStatus } from '@/types/claim'

const f = useFilters()

// ขอ limit+1 แถว เพื่อรู้ว่ามีหน้าถัดไปจริงหรือไม่ (API ไม่ส่งจำนวนรวมมา)
const { data, loading, error, run } = useLoad(() =>
  listClaims({ status: f.status, limit: f.limit + 1, offset: f.offset }),
)

const rows = computed(() => (data.value?.items ?? []).slice(0, f.limit))
const hasNext = computed(() => (data.value?.items.length ?? 0) > f.limit)

onMounted(run)
watch(() => [f.status, f.page], run)

function onStatus(e: Event) {
  f.setStatus((e.target as HTMLSelectElement).value as ClaimStatus | '')
}
</script>

<template>
  <section class="card">
    <div class="toolbar">
      <label>
        สถานะ
        <select data-test="status" :value="f.status" @change="onStatus">
          <option value="">ทั้งหมด</option>
          <option v-for="s in STATUSES" :key="s" :value="s">{{ STATUS_LABEL[s] }}</option>
        </select>
      </label>
      <RouterLink to="/new" class="add"><BaseButton>+ ส่งเคลม</BaseButton></RouterLink>
      <BaseButton variant="ghost" data-test="refresh" :disabled="loading" @click="run">รีเฟรช</BaseButton>
    </div>

    <p v-if="error" class="error" role="alert" data-test="error">
      {{ error }} <button type="button" class="retry" data-test="retry" @click="run">ลองใหม่</button>
    </p>
    <p v-else-if="loading && !data" class="state" data-test="loading">กำลังโหลด…</p>
    <ClaimList v-else :claims="rows">
      <template #empty>ยังไม่มีรายการ</template>
    </ClaimList>

    <div class="pager">
      <BaseButton variant="ghost" data-test="prev" :disabled="f.page === 0 || loading" @click="f.prev()">ก่อนหน้า</BaseButton>
      <span data-test="page">หน้า {{ f.page + 1 }}</span>
      <BaseButton variant="ghost" data-test="next" :disabled="!hasNext || loading" @click="f.next()">ถัดไป</BaseButton>
    </div>
  </section>
</template>

<style scoped>
.card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 18px; display: flex; flex-direction: column; gap: 14px; }
.toolbar { display: flex; gap: 10px; align-items: flex-end; flex-wrap: wrap; }
.toolbar label { display: flex; flex-direction: column; gap: 4px; font-size: 13px; color: var(--muted); margin-right: auto; }
select { padding: 8px 10px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); color: var(--text); font: inherit; }
.add { text-decoration: none; }
.state { margin: 0; color: var(--muted); }
.retry { font: inherit; cursor: pointer; }
.pager { display: flex; justify-content: center; align-items: center; gap: 12px; color: var(--muted); font-size: 14px; }
</style>
