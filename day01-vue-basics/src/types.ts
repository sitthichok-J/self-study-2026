export type ClaimStatus = 'PENDING' | 'APPROVED' | 'REJECTED'

export interface Claim {
  id: number
  patientRef: string
  hospital: string
  amount: number
  status: ClaimStatus
}

export const STATUS_LABEL: Record<ClaimStatus, string> = {
  PENDING: 'รอตรวจสอบ',
  APPROVED: 'อนุมัติ',
  REJECTED: 'ปฏิเสธ',
}
