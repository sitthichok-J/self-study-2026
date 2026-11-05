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
  needs_review: 'รอเจ้าหน้าที่ตรวจ', // Day 13: ยอดช่วง 20,000-50,000 บาท worker ไม่ตัดสินเอง
  manual_approved: 'เจ้าหน้าที่อนุมัติ',
  manual_rejected: 'เจ้าหน้าที่ไม่อนุมัติ',
}

export const isFinal = (s: ClaimStatus): boolean => s !== 'PENDING'

// หยุด poll ได้เมื่อ "ไม่มีอะไรจะเปลี่ยนเอง" แล้ว: ตัดสินแล้ว หรือรอคนตรวจ (worker ทำเสร็จแล้ว)
export const isSettled = (c: { status: ClaimStatus; decision_reason?: string }): boolean =>
  isFinal(c.status) || c.decision_reason === 'needs_review'
