export type ClaimStatus = 'PENDING' | 'APPROVED' | 'REJECTED';
export type StatusFilter = 'ALL' | ClaimStatus;

// รูปร่างตรงกับ response ของ backend Day 8 (day08-backend/internal/domain/claim.go)
// amount เป็นสตางค์ (จำนวนเต็ม) ไม่ใช่บาท
export interface Claim {
  id: string;
  patient_ref: string;
  amount: number;
  status: ClaimStatus;
  decision_reason?: string;
  created_at: string;
  updated_at: string;
}

// GET /claims?status=&limit=&offset= ตอบ {items, limit, offset}
export interface ClaimListResponse {
  items: Claim[];
  limit: number;
  offset: number;
}

export const satangToBaht = (satang: number): number => satang / 100;

// POST /claims รับแค่ patient_ref กับ amount (สตางค์) backend ใช้ DisallowUnknownFields
// จึงส่งฟิลด์อื่น เช่น "ประเภท" ไม่ได้ (ถ้าส่งจะได้ 400)
export interface CreateClaimInput {
  patient_ref: string;
  amount: number;
}

// ผู้ใช้กรอกเป็นบาท backend ต้องการสตางค์จำนวนเต็ม: ปัดเพื่อกัน floating point (เช่น 19.99 * 100 = 1998.9999...)
export const bahtToSatang = (baht: number): number => Math.round(baht * 100);

// ตรงกับ backend: patient_ref ต้องเป็น FAKE-XXXX, ยอดสูงสุด 1_000_000_000 สตางค์ (= 10,000,000 บาท)
export const PATIENT_REF_PATTERN = /^FAKE-[A-Z0-9]{3,12}$/;
export const MAX_AMOUNT_BAHT = 10_000_000;
