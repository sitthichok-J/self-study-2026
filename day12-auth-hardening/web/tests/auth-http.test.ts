import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError, configureAuth, http } from '@/api/http'
import { jsonResponse, stubFetch } from './helpers'

afterEach(() => {
  configureAuth(null)
  vi.unstubAllGlobals()
})

const auth = (over: Partial<Parameters<typeof configureAuth>[0] & object> = {}) => {
  const state = { token: 'tok-1' as string | null }
  const hooks = {
    getToken: vi.fn(() => state.token),
    refresh: vi.fn(async () => {
      state.token = 'tok-2'
      return true
    }),
    onLoggedOut: vi.fn(),
    ...over,
  }
  configureAuth(hooks)
  return { hooks, state }
}

const headersOf = (fn: ReturnType<typeof vi.fn>, i = 0) => (fn.mock.calls[i]![1] as RequestInit).headers as Record<string, string>

describe('http + auth hooks', () => {
  it('แนบ Bearer เมื่อมี token และไม่แนบเมื่อไม่มี', async () => {
    const { state } = auth()
    const f = stubFetch({ body: {} }, { body: {} })
    await http('/claims')
    state.token = null
    await http('/claims')
    expect(headersOf(f, 0).Authorization).toBe('Bearer tok-1')
    expect(headersOf(f, 1)).not.toHaveProperty('Authorization')
  })

  it('ไม่ตั้ง hook = ทำงานเหมือน Day 9 (ไม่มี Authorization, ไม่ลองซ้ำ)', async () => {
    const f = stubFetch({ status: 401, body: { error: 'unauthorized' } })
    await expect(http('/claims')).rejects.toMatchObject({ status: 401 })
    expect(f).toHaveBeenCalledTimes(1)
    expect(headersOf(f)).not.toHaveProperty('Authorization')
  })

  it('401 -> refresh -> ลองซ้ำครั้งเดียวด้วย token ใหม่', async () => {
    const { hooks } = auth()
    const f = stubFetch({ status: 401, body: { error: 'unauthorized' } }, { body: { ok: 1 } })
    const r = await http<{ ok: number }>('/claims')
    expect(r.data.ok).toBe(1)
    expect(hooks.refresh).toHaveBeenCalledTimes(1)
    expect(f).toHaveBeenCalledTimes(2)
    expect(headersOf(f, 0).Authorization).toBe('Bearer tok-1')
    expect(headersOf(f, 1).Authorization).toBe('Bearer tok-2')
    expect(hooks.onLoggedOut).not.toHaveBeenCalled()
  })

  it('ลองซ้ำแล้วยัง 401: หยุด ไม่วน refresh อีก และถือว่า session จบ', async () => {
    const { hooks } = auth()
    const f = stubFetch({ status: 401, body: {} }, { status: 401, body: { error: 'unauthorized' } })
    await expect(http('/claims')).rejects.toMatchObject({ status: 401 })
    expect(f).toHaveBeenCalledTimes(2)
    expect(hooks.refresh).toHaveBeenCalledTimes(1)
  })

  it('refresh ไม่สำเร็จ: onLoggedOut ถูกเรียก และได้ 401 เดิมกลับไป (ไม่ยิงซ้ำ)', async () => {
    const { hooks } = auth({ refresh: vi.fn(async () => false) })
    const f = stubFetch({ status: 401, body: { error: 'unauthorized' } })
    const e = await http('/claims').catch((x) => x)
    expect(e).toBeInstanceOf(ApiError)
    expect(e.status).toBe(401)
    expect(f).toHaveBeenCalledTimes(1)
    expect(hooks.onLoggedOut).toHaveBeenCalledTimes(1)
  })

  it('POST ที่ถูกลองซ้ำ ใช้ body และ Idempotency-Key เดิม', async () => {
    auth()
    const f = stubFetch({ status: 401, body: {} }, { status: 201, body: { id: 'x' } })
    await http('/claims', { method: 'POST', headers: { 'Idempotency-Key': 'k-1', 'Content-Type': 'application/json' }, body: '{"a":1}' })
    for (const i of [0, 1]) {
      expect(headersOf(f, i)['Idempotency-Key']).toBe('k-1')
      expect((f.mock.calls[i]![1] as RequestInit).body).toBe('{"a":1}')
    }
  })

  it('error อื่นที่ไม่ใช่ 401 ไม่แตะ refresh', async () => {
    const { hooks } = auth()
    stubFetch({ status: 404, body: { error: 'claim not found' } }, { status: 500, body: {} }, { status: 429, body: {} })
    for (const s of [404, 500, 429]) await expect(http('/claims')).rejects.toMatchObject({ status: s })
    expect(hooks.refresh).not.toHaveBeenCalled()
    expect(hooks.onLoggedOut).not.toHaveBeenCalled()
  })

  it('เส้นทาง /auth/* ไม่เข้าวงจร refresh (กันวนไม่รู้จบ)', async () => {
    const { hooks } = auth()
    const f = stubFetch({ status: 401, body: { error: 'unauthorized' } })
    await expect(http('/auth/refresh')).rejects.toMatchObject({ status: 401 })
    expect(hooks.refresh).not.toHaveBeenCalled()
    expect(f).toHaveBeenCalledTimes(1)
  })

  it('เมื่อ refresh เจอ network error ยังได้ 401 เดิมและเรียก onLoggedOut', async () => {
    const { hooks } = auth({ refresh: vi.fn(async () => false) })
    stubFetch({ status: 401, body: {} })
    await expect(http('/claims')).rejects.toBeInstanceOf(ApiError)
    expect(hooks.onLoggedOut).toHaveBeenCalled()
  })
})

describe('http + auth: network', () => {
  beforeEach(() => auth())
  it('fetch ล้มเหลว = ApiError status 0 ไม่ใช่การ refresh', async () => {
    stubFetch(new TypeError('Failed to fetch'))
    await expect(http('/claims')).rejects.toMatchObject({ status: 0 })
  })
  it('response ไม่ใช่ 401 ไม่ถูกอ่านซ้ำ', async () => {
    const f = vi.fn().mockResolvedValue(jsonResponse({ body: { a: 1 } }))
    vi.stubGlobal('fetch', f)
    expect((await http<{ a: number }>('/x')).data.a).toBe(1)
    expect(f).toHaveBeenCalledTimes(1)
  })
})
