export type ClaimStatus = 'PENDING' | 'APPROVED' | 'REJECTED';

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
