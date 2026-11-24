import { Component, DestroyRef, computed, effect, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ClaimCard } from '../claim-card/claim-card';
import { readApiError } from '../api-error';
import { Claim, ClaimStatus, satangToBaht } from '../claim';
import { ClaimService } from '../claim.service';

export type StatusFilter = 'ALL' | ClaimStatus;

@Component({
  selector: 'app-claim-list',
  imports: [ClaimCard],
  styleUrl: './claim-list.css',
  templateUrl: './claim-list.html',
})
export class ClaimList {
  private readonly service = inject(ClaimService);
  private readonly destroyRef = inject(DestroyRef);

  // วันที่ 3: ข้อมูลมาจาก backend จึงมีสามสถานะ loading / error / ข้อมูล
  protected readonly claims = signal<Claim[]>([]);
  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly selectedId = signal<string | null>(null);

  protected readonly statusFilter = signal<StatusFilter>('ALL');
  protected readonly searchText = signal('');

  protected readonly filtered = computed(() => {
    const status = this.statusFilter();
    const q = this.searchText().trim().toLowerCase();
    return this.claims().filter(
      (c) =>
        (status === 'ALL' || c.status === status) &&
        (q === '' || c.id.toLowerCase().includes(q) || c.patient_ref.toLowerCase().includes(q)),
    );
  });

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
      console.log(`[claim-list] filter เปลี่ยน: status=${this.statusFilter()} search="${this.searchText()}"`);
    });
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.service
      .list()
      // takeUntilDestroyed: ถ้า component ถูกทำลายก่อน response มา จะยกเลิก subscription ให้เอง
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (res) => {
          this.claims.set(res.items);
          this.loading.set(false);
        },
        error: (err: unknown) => {
          this.error.set(describeError(err));
          this.loading.set(false);
        },
      });
  }

  onSelect(id: string): void {
    this.selectedId.set(id);
  }

  onStatusChange(value: string): void {
    this.statusFilter.set(value as StatusFilter);
  }

  onSearch(value: string): void {
    this.searchText.set(value);
  }
}

function describeError(err: unknown): string {
  const api = readApiError(err);
  if (api) {
    return api.status === 0 ? 'เชื่อมต่อ backend ไม่ได้' : `backend ตอบ HTTP ${api.status}`;
  }
  return 'เกิดข้อผิดพลาดที่ไม่ทราบสาเหตุ';
}
