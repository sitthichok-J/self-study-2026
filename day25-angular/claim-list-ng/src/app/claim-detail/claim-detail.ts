import { HttpErrorResponse, httpResource } from '@angular/common/http';
import { Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Claim, satangToBaht } from '../claim';
import { CLAIMS_URL } from '../claim.service';

@Component({
  selector: 'app-claim-detail',
  imports: [RouterLink],
  styleUrl: './claim-detail.css',
  templateUrl: './claim-detail.html',
})
export class ClaimDetail {
  // :id จาก URL เข้ามาทาง input() เพราะเปิด withComponentInputBinding() ใน app.config
  readonly id = input.required<string>();

  // httpResource อ่าน id() เป็น signal: URL เปลี่ยนแล้วยิง request ใหม่เอง (ไม่ต้องจัดการ subscribe)
  protected readonly claim = httpResource<Claim>(() => `${CLAIMS_URL}/${encodeURIComponent(this.id())}`);

  protected readonly amountBaht = computed(() => {
    // ห้ามอ่าน value() ตอน resource อยู่ในสถานะ error (มันโยน error) จึงเช็ค hasValue() ก่อน
    return this.claim.hasValue() ? satangToBaht(this.claim.value().amount) : null;
  });

  protected readonly errorMessage = computed(() => {
    const err = this.claim.error();
    if (!err) return null;
    const status = err instanceof HttpErrorResponse ? err.status : (err as { status?: number }).status;
    if (status === 404) return 'ไม่พบเคลมนี้';
    if (status === 0) return 'เชื่อมต่อ backend ไม่ได้';
    return `backend ตอบ HTTP ${status ?? 'ไม่ทราบ'}`;
  });
}
