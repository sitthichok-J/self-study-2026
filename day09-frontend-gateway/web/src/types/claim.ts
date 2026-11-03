export const STATUSES = ['PENDING', 'APPROVED', 'REJECTED'] as const
export type ClaimStatus = (typeof STATUSES)[number]

export interface Claim {
  id: string
  patient_ref: string
  amount: number // satang (หน่วยย่อย) เป็นจำนวนเต็มเสมอ
  status: ClaimStatus
  decision_reason?: string
  created_at: string
  updated_at: string
}

export interface ClaimPage {
  items: Claim[]
  limit: number
  offset: number
}

export interface NewClaim {
  patient_ref: string
  amount: number // satang
}

export const STATUS_LABEL: Record<ClaimStatus, string> = {
  PENDING: 'รอประมวลผล',
  APPROVED: 'อนุมัติ',
  REJECTED: 'ไม่อนุมัติ',
}

export const REASON_LABEL: Record<string, string> = {
  ok: 'ผ่านเกณฑ์',
  over_limit: 'เกินวงเงิน',
}

export const isFinal = (s: ClaimStatus): boolean => s !== 'PENDING'
