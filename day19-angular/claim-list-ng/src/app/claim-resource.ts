import { httpResource } from '@angular/common/http';
import { Signal } from '@angular/core';
import { ClaimListResponse } from './claim';
import { CLAIMS_URL } from './claim.service';
import { StatusFilter } from './claim-list/claim-list';

// แบบฝึกข้อ 4: ดึงรายการด้วย httpResource (API แบบ signal) เทียบกับ ClaimService + Observable
// ตัวกรองสถานะเป็น signal พอค่าเปลี่ยน resource จะยิง request ใหม่เอง ไม่ต้อง subscribe เอง
// ใช้ใน injection context เท่านั้น (เช่น field ของ component หรือ constructor)
export function claimsResource(status: Signal<StatusFilter>) {
  return httpResource<ClaimListResponse>(() => {
    const s = status();
    return {
      url: CLAIMS_URL,
      params: { limit: 100, offset: 0, ...(s === 'ALL' ? {} : { status: s }) },
    };
  });
}
