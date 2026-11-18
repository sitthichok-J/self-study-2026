export type ClaimStatus = 'PENDING' | 'APPROVED' | 'REJECTED';

export interface Claim {
  id: string;
  ownerName: string;
  amount: number;
  status: ClaimStatus;
}

// ข้อมูลสมมติทั้งหมด (วันที่ 3 จะเปลี่ยนไปดึงจาก backend จริง)
export const MOCK_CLAIMS: Claim[] = [
  { id: 'CLM-001', ownerName: 'สมชาย ตัวอย่าง', amount: 1500, status: 'PENDING' },
  { id: 'CLM-002', ownerName: 'สมหญิง ทดสอบ', amount: 32000, status: 'APPROVED' },
  { id: 'CLM-003', ownerName: 'วิชัย สมมติ', amount: 4800, status: 'REJECTED' },
  { id: 'CLM-004', ownerName: 'มานี ตัวอย่าง', amount: 760, status: 'PENDING' },
  { id: 'CLM-005', ownerName: 'ปิติ ทดสอบ', amount: 12500, status: 'APPROVED' },
];
