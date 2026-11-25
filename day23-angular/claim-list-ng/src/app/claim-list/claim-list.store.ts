import { Injectable, computed, signal } from '@angular/core';
import { Claim, StatusFilter } from '../claim';
import { ClaimListQuery, claimsResource } from '../claim-resource';

export const PAGE_SIZES = [5, 10, 20] as const;
export const DEFAULT_PAGE_SIZE = 10;

// state ของหน้ารายการอยู่ที่นี่ที่เดียว: ตัวกรอง + หน้า → query → httpResource
// provide ที่ระดับ component (providers: [ClaimListStore]) ให้เกิดใหม่ทุกครั้งที่เปิดหน้า ไม่ค้างข้ามหน้า
@Injectable()
export class ClaimListStore {
  readonly status = signal<StatusFilter>('ALL');
  readonly pageSize = signal<number>(DEFAULT_PAGE_SIZE);
  readonly page = signal(0);

  private readonly query = computed<ClaimListQuery>(() => ({
    status: this.status(),
    pageSize: this.pageSize(),
    page: this.page(),
  }));
  private readonly resource = claimsResource(this.query);

  readonly loading = this.resource.isLoading;
  readonly error = computed<unknown>(() => this.resource.error() ?? null);

  // ตัดแถวที่ขอเกินมาหนึ่งแถว (ไว้ดูว่ามีหน้าถัดไปไหม) ออกก่อนแสดง
  readonly items = computed<Claim[]>(() =>
    this.resource.hasValue() ? this.resource.value().items.slice(0, this.pageSize()) : [],
  );
  readonly hasNext = computed(() => this.resource.hasValue() && this.resource.value().items.length > this.pageSize());
  readonly hasPrev = computed(() => this.page() > 0);

  // เปลี่ยนตัวกรองหรือขนาดหน้าแล้วต้องกลับหน้าแรก ไม่งั้น offset เดิมอาจเลยจำนวนข้อมูลที่เหลือ
  setStatus(status: StatusFilter): void {
    this.status.set(status);
    this.page.set(0);
  }

  setPageSize(size: number): void {
    this.pageSize.set(size);
    this.page.set(0);
  }

  next(): void {
    if (this.hasNext()) this.page.update((p) => p + 1);
  }

  prev(): void {
    if (this.hasPrev()) this.page.update((p) => p - 1);
  }

  reload(): void {
    this.resource.reload();
  }
}
