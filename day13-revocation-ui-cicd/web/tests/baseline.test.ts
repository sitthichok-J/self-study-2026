import { describe, expect, it } from 'vitest'
import { PAGE_SIZE } from '@/stores/filters'
import { bahtToSatang, formatBaht, MAX_SATANG, PATIENT_REF_RE, validateClaimForm } from '@/lib'
import { REASON_LABEL, STATUS_LABEL, STATUSES, isFinal, isSettled } from '@/types/claim'

// baseline: ภาพถ่ายพฤติกรรมปัจจุบัน ถ้าแก้กฎโดยตั้งใจ ให้รัน  npm run test:update
describe('baseline', () => {
  it('กฎตรวจฟอร์ม', () => {
    const cases: Array<[string, string]> = [
      ['FAKE-AB12', '1500'],
      ['FAKE-AB12', '0.29'],
      ['FAKE-AB12', '10000000'],
      ['FAKE-AB12', '10000000.01'],
      ['FAKE-AB12', '0'],
      ['FAKE-AB12', '1.234'],
      ['FAKE-AB', '10'],
      ['fake-ab12', '10'],
      ['', ''],
    ]
    expect(cases.map(([r, a]) => ({ r, a, ...validateClaimForm(r, a) }))).toMatchSnapshot()
  })

  it('ค่าคงที่และป้ายชื่อ', () => {
    expect({
      pattern: PATIENT_REF_RE.source,
      maxSatang: MAX_SATANG,
      pageSize: PAGE_SIZE,
      statuses: STATUSES,
      statusLabel: STATUS_LABEL,
      reasonLabel: REASON_LABEL,
      final: STATUSES.map((s) => [s, isFinal(s)]),
      settled: STATUSES.flatMap((s) => [undefined, 'ok', 'needs_review'].map((r) => [s, r ?? null, isSettled({ status: s, decision_reason: r })])),
    }).toMatchSnapshot()
  })

  it('การแปลงจำนวนเงิน', () => {
    const inputs = ['0.01', '0.1', '1', '19.99', '1234567.89']
    expect(inputs.map((i) => [i, bahtToSatang(i), formatBaht(bahtToSatang(i)!)])).toMatchSnapshot()
  })
})
