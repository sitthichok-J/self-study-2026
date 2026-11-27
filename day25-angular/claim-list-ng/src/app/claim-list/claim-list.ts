import { Component, computed, effect, inject, signal } from '@angular/core';
import { ClaimCard } from '../claim-card/claim-card';
import { readApiError } from '../api-error';
import { ClaimStatus, StatusFilter, satangToBaht } from '../claim';
import { ClaimListStore, PAGE_SIZES } from './claim-list.store';

export type { StatusFilter } from '../claim';

@Component({
  selector: 'app-claim-list',
  imports: [ClaimCard],
  providers: [ClaimListStore],
  styleUrl: './claim-list.css',
  templateUrl: './claim-list.html',
})
export class ClaimList {
  protected readonly store = inject(ClaimListStore);
  protected readonly pageSizes = PAGE_SIZES;

  // วันที่ 23: status กับหน้าอยู่ใน store (ส่งไป backend) ที่เหลือใน component คือค้นหาในหน้าที่โหลดมา กับเลือกการ์ด
  protected readonly claims = this.store.items;
  protected readonly loading = this.store.loading;
  protected readonly error = computed(() => (this.store.error() ? describeError(this.store.error()) : null));
  protected readonly selectedId = signal<string | null>(null);

  protected readonly searchText = signal('');

  // backend ไม่มี search: ค้นหาเฉพาะรายการในหน้าที่โหลดมาแล้ว
  protected readonly filtered = computed(() => {
    const q = this.searchText().trim().toLowerCase();
    return this.claims().filter((c) => q === '' || c.id.toLowerCase().includes(q) || c.patient_ref.toLowerCase().includes(q));
  });

  // ยอดรวม/จำนวนต่อสถานะ นับเฉพาะรายการในหน้านี้ (backend ไม่ส่ง total มาให้)
  protected readonly totalAmount = computed(() =>
    satangToBaht(this.claims().reduce((sum, c) => sum + c.amount, 0)),
  );
  protected readonly countByStatus = computed(() => {
    const counts: Record<ClaimStatus, number> = { PENDING: 0, APPROVED: 0, REJECTED: 0 };
    for (const c of this.claims()) counts[c.status]++;
    return counts;
  });
  protected readonly filteredAmount = computed(() =>
    satangToBaht(this.filtered().reduce((sum, c) => sum + c.amount, 0)),
  );

  constructor() {
    effect(() => {
      console.log(`[claim-list] filter เปลี่ยน: status=${this.store.status()} search="${this.searchText()}"`);
    });
  }

  load(): void {
    this.store.reload();
  }

  onSelect(id: string): void {
    this.selectedId.set(id);
  }

  onStatusChange(value: string): void {
    this.store.setStatus(value as StatusFilter);
  }

  onSearch(value: string): void {
    this.searchText.set(value);
  }

  onPageSize(value: string): void {
    this.store.setPageSize(Number(value));
  }
}

function describeError(err: unknown): string {
  const api = readApiError(err);
  if (api) {
    return api.status === 0 ? 'เชื่อมต่อ backend ไม่ได้' : `backend ตอบ HTTP ${api.status}`;
  }
  return 'เกิดข้อผิดพลาดที่ไม่ทราบสาเหตุ';
}
