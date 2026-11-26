import { Claim, ClaimListResponse } from './claim';

// ข้อมูลสมมติสำหรับเทสต์ (รูปร่างเดียวกับ backend Day 8, amount เป็นสตางค์)
const base = { created_at: '2026-11-18T01:00:00Z', updated_at: '2026-11-18T01:00:00Z' };
export const TEST_CLAIMS: Claim[] = [
  { id: 'CLM-001', patient_ref: 'FAKE-AB12', amount: 150000, status: 'PENDING', ...base },
  { id: 'CLM-002', patient_ref: 'FAKE-CD34', amount: 3200000, status: 'APPROVED', ...base },
  { id: 'CLM-003', patient_ref: 'FAKE-EF56', amount: 480000, status: 'REJECTED', decision_reason: 'amount_over_limit', ...base },
  { id: 'CLM-004', patient_ref: 'FAKE-GH78', amount: 76000, status: 'PENDING', ...base },
  { id: 'CLM-005', patient_ref: 'FAKE-JK90', amount: 1250000, status: 'APPROVED', ...base },
];
export const TEST_RESPONSE: ClaimListResponse = { items: TEST_CLAIMS, limit: 100, offset: 0 };
