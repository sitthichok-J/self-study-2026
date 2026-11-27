import { httpResource } from '@angular/common/http';
import { Signal } from '@angular/core';
import { ClaimListResponse, StatusFilter } from './claim';
import { CLAIMS_URL } from './claim.service';

// สถานะ query ของหน้ารายการ: page เริ่มที่ 0
export interface ClaimListQuery {
  status: StatusFilter;
  pageSize: number;
  page: number;
}

// วันที่ 23: ส่ง status / limit / offset ไปให้ backend กรองและแบ่งหน้า (เดิมโหลด 100 รายการแล้วกรองในเครื่อง)
// ขอ pageSize + 1 แถว: ถ้าได้แถวเกิน pageSize แปลว่ายังมีหน้าถัดไป (backend ไม่ส่ง total มาให้)
// backend จำกัด limit สูงสุด 100 (MaxLimit) ดังนั้น pageSize ต้องไม่เกิน 99
// httpResource: signal ใน callback เปลี่ยน → ยิง request ใหม่เอง และยกเลิกตัวเก่าที่ค้างอยู่
// ใช้ใน injection context เท่านั้น
export function claimsResource(query: Signal<ClaimListQuery>) {
  return httpResource<ClaimListResponse>(() => {
    const q = query();
    return {
      url: CLAIMS_URL,
      params: {
        limit: q.pageSize + 1,
        offset: q.page * q.pageSize,
        ...(q.status === 'ALL' ? {} : { status: q.status }),
      },
    };
  });
}
