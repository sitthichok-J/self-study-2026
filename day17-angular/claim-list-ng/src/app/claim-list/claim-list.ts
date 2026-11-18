import { Component, computed, effect, signal } from '@angular/core';
import { ClaimCard } from '../claim-card/claim-card';
import { Claim, ClaimStatus, MOCK_CLAIMS } from '../claim';

export type StatusFilter = 'ALL' | ClaimStatus;

@Component({
  selector: 'app-claim-list',
  // standalone component: component อื่นที่ใช้ใน template ต้องประกาศในอาเรย์ imports ของ component นี้เอง
  imports: [ClaimCard],
  styleUrl: './claim-list.css',
  templateUrl: './claim-list.html',
})
export class ClaimList {
  protected readonly claims = signal<Claim[]>(MOCK_CLAIMS);
  protected readonly selectedId = signal<string | null>(null);

  // วันที่ 2: state ของตัวกรองเก็บเป็น signal ทั้งหมด
  protected readonly statusFilter = signal<StatusFilter>('ALL');
  protected readonly searchText = signal('');

  // รายการที่กรองแล้ว: computed จะคำนวณใหม่เฉพาะตอน signal ที่อ่านอยู่เปลี่ยน (ไม่เรียกฟังก์ชันกรองซ้ำทุกครั้งที่ render)
  protected readonly filtered = computed(() => {
    const status = this.statusFilter();
    const q = this.searchText().trim().toLowerCase();
    return this.claims().filter(
      (c) =>
        (status === 'ALL' || c.status === status) &&
        (q === '' || c.id.toLowerCase().includes(q) || c.ownerName.toLowerCase().includes(q)),
    );
  });

  // ยอดรวมและจำนวนต่อสถานะ คิดจากรายการทั้งหมด (ไม่ขึ้นกับตัวกรอง)
  protected readonly totalAmount = computed(() => this.claims().reduce((sum, c) => sum + c.amount, 0));
  protected readonly countByStatus = computed(() => {
    const counts: Record<ClaimStatus, number> = { PENDING: 0, APPROVED: 0, REJECTED: 0 };
    for (const c of this.claims()) counts[c.status]++;
    return counts;
  });
  // ยอดรวมของรายการที่กรองอยู่ (แสดงข้างผลลัพธ์)
  protected readonly filteredAmount = computed(() => this.filtered().reduce((sum, c) => sum + c.amount, 0));

  constructor() {
    // effect ใช้กับ side effect เท่านั้น (ตรงนี้แค่ log) ส่วนค่าที่คำนวณได้ให้ใช้ computed
    effect(() => {
      console.log(`[claim-list] filter เปลี่ยน: status=${this.statusFilter()} search="${this.searchText()}"`);
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
