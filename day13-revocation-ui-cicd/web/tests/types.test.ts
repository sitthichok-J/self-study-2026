import { describe, expect, it } from 'vitest'
import { isFinal, isSettled, REASON_LABEL } from '@/types/claim'

describe('isSettled (เมื่อไหร่หยุด poll)', () => {
  it.each([
    ['PENDING', undefined, false],
    ['PENDING', 'needs_review', true],
    ['PENDING', 'something_else', false],
    ['APPROVED', 'ok', true],
    ['APPROVED', 'manual_approved', true],
    ['REJECTED', 'over_limit', true],
    ['REJECTED', undefined, true],
  ] as const)('%s / %s -> %s', (status, reason, want) => {
    expect(isSettled({ status, decision_reason: reason })).toBe(want)
  })

  it('isFinal ไม่เปลี่ยนความหมาย: PENDING ที่รอตรวจยังไม่ final', () => {
    expect(isFinal('PENDING')).toBe(false)
    expect(isFinal('APPROVED')).toBe(true)
  })
})

describe('REASON_LABEL ครบทุก reason ที่ backend ส่งได้', () => {
  it.each(['ok', 'over_limit', 'needs_review', 'manual_approved', 'manual_rejected'])('%s มีป้ายภาษาไทย', (k) => {
    expect(REASON_LABEL[k]).toBeTruthy()
  })
})
