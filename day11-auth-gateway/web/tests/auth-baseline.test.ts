import { describe, expect, it } from 'vitest'
import { AUTH_PREFIX, CSRF_HEADER, CSRF_VALUE } from '@/api/http'
import { safeRedirect } from '@/lib'

// baseline: สัญญาระหว่าง UI กับ API ที่เปลี่ยนแล้วต้องรู้ตัว (ต้องตรงกับ Go: CSRFHeader/CSRFValue)
describe('auth contract baseline', () => {
  it('ค่าคงที่และการตัดสิน redirect', () => {
    expect({
      csrf: [CSRF_HEADER, CSRF_VALUE],
      prefix: AUTH_PREFIX,
      redirects: ['/claims/1', '//x', '/login', 'http://x', '/a?b=1'].map(safeRedirect),
    }).toMatchSnapshot()
  })
})
