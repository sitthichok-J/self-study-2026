import { defineStore } from 'pinia'
import type { ClaimStatus } from '@/types/claim'

export const PAGE_SIZE = 10

export const useFilters = defineStore('filters', {
  state: () => ({ status: '' as ClaimStatus | '', page: 0 }),
  getters: {
    limit: () => PAGE_SIZE,
    offset: (s) => s.page * PAGE_SIZE,
  },
  actions: {
    setStatus(s: ClaimStatus | '') {
      this.status = s
      this.page = 0 // เปลี่ยน filter แล้วต้องกลับหน้าแรกเสมอ
    },
    next() {
      this.page++
    },
    prev() {
      if (this.page > 0) this.page--
    },
  },
})
