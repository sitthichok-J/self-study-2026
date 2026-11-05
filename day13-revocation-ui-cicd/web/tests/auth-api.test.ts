import { afterEach, describe, expect, it, vi } from 'vitest'
import { loginRequest, logoutRequest, refreshRequest } from '@/api/auth'
import { CSRF_HEADER, CSRF_VALUE, configureAuth } from '@/api/http'
import { stubFetch } from './helpers'

afterEach(() => {
  configureAuth(null)
  vi.unstubAllGlobals()
})

const call = (f: ReturnType<typeof vi.fn>, i = 0) => ({ url: f.mock.calls[i]![0] as string, init: f.mock.calls[i]![1] as RequestInit })
const tokens = { access_token: 'a.b.c', token_type: 'Bearer', expires_in: 600 }

describe('auth api', () => {
  it('login: POST JSON พร้อม credentials และ header กัน CSRF โดยไม่แนบ Bearer', async () => {
    configureAuth({ getToken: () => 'should-not-be-sent', refresh: async () => true, onLoggedOut: () => {} })
    const f = stubFetch({ body: tokens })
    const r = await loginRequest('alice', 'pw-pw-pw-pw-1')
    const { url, init } = call(f)
    expect(url).toBe('/api/auth/login')
    expect(init.method).toBe('POST')
    expect(init.credentials).toBe('include')
    expect(init.headers).toMatchObject({ 'Content-Type': 'application/json', [CSRF_HEADER]: CSRF_VALUE })
    expect(init.headers).not.toHaveProperty('Authorization')
    expect(JSON.parse(init.body as string)).toEqual({ username: 'alice', password: 'pw-pw-pw-pw-1' })
    expect(r).toEqual(tokens)
  })

  it('refresh: POST ไม่มี body พร้อม cookie และ header กัน CSRF', async () => {
    const f = stubFetch({ body: tokens })
    await refreshRequest()
    const { url, init } = call(f)
    expect(url).toBe('/api/auth/refresh')
    expect(init.credentials).toBe('include')
    expect(init.body).toBeUndefined()
    expect(init.headers).toMatchObject({ [CSRF_HEADER]: CSRF_VALUE })
    expect(init.headers).not.toHaveProperty('Content-Type')
  })

  it('logout: POST และรับ 204 ได้ (ไม่มี body)', async () => {
    const f = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', f)
    await expect(logoutRequest()).resolves.toBeUndefined()
    expect(call(f).url).toBe('/api/auth/logout')
    expect(call(f).init.credentials).toBe('include')
  })

  it('logout ส่ง access token เป็น Bearer เมื่อมี (ให้ server เพิกถอน) และไม่ส่งเมื่อไม่มี', async () => {
    const f = vi.fn().mockImplementation(() => Promise.resolve(new Response(null, { status: 204 })))
    vi.stubGlobal('fetch', f)
    await logoutRequest('tok-123')
    expect(call(f).init.headers).toMatchObject({ Authorization: 'Bearer tok-123', [CSRF_HEADER]: CSRF_VALUE })
    await logoutRequest(null)
    await logoutRequest()
    expect(f.mock.calls[1][1].headers).not.toHaveProperty('Authorization')
    expect(f.mock.calls[2][1].headers).not.toHaveProperty('Authorization')
  })

  it('login และ refresh ไม่แนบ Bearer (ใช้ cookie เท่านั้น)', async () => {
    const f = stubFetch({ body: tokens }, { body: tokens })
    await loginRequest('a', 'b')
    await refreshRequest()
    expect(f.mock.calls[0][1].headers).not.toHaveProperty('Authorization')
    expect(f.mock.calls[1][1].headers).not.toHaveProperty('Authorization')
  })

  it('401 และ 429 กลายเป็น ApiError พร้อม status และ Retry-After', async () => {
    stubFetch({ status: 401, body: { error: 'invalid username or password' } }, { status: 429, body: { message: 'API rate limit exceeded' }, headers: { 'Retry-After': '30' } })
    await expect(loginRequest('a', 'b')).rejects.toMatchObject({ status: 401, message: 'invalid username or password' })
    await expect(loginRequest('a', 'b')).rejects.toMatchObject({ status: 429, retryAfter: 30 })
  })

  it('network error เป็น status 0', async () => {
    stubFetch(new TypeError('Failed to fetch'))
    await expect(refreshRequest()).rejects.toMatchObject({ status: 0 })
  })

  it('header CSRF ตรงกับฝั่ง Go', () => {
    expect(CSRF_HEADER).toBe('X-Requested-With')
    expect(CSRF_VALUE).toBe('claims-web')
  })
})
