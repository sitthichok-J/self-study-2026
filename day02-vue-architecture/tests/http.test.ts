import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { claimsApi } from '../src/api/claimsApi'
import { ApiError, http, TOKEN_KEY } from '../src/api/http'

const ok = (body: unknown, status = 200) => new Response(body === null ? null : JSON.stringify(body), { status })

describe('http / claimsApi', () => {
  const fetchMock = vi.fn()
  beforeEach(() => { localStorage.clear(); fetchMock.mockReset(); vi.stubGlobal('fetch', fetchMock) })
  afterEach(() => vi.unstubAllGlobals())

  it('แนบ Authorization เมื่อมี token', async () => {
    localStorage.setItem(TOKEN_KEY, 'abc')
    fetchMock.mockResolvedValue(ok([]))
    await http('/api/claims')
    expect(fetchMock.mock.calls[0][1].headers.Authorization).toBe('Bearer abc')
  })

  it('ไม่มี token -> ไม่แนบ header', async () => {
    fetchMock.mockResolvedValue(ok([]))
    await http('/api/claims')
    expect(fetchMock.mock.calls[0][1].headers.Authorization).toBeUndefined()
  })

  it('error -> ApiError พร้อม message จาก server', async () => {
    fetchMock.mockResolvedValue(ok({ message: 'ไม่พบเคลม' }, 404))
    await expect(http('/x')).rejects.toMatchObject({ name: 'ApiError', status: 404, message: 'ไม่พบเคลม' })
    fetchMock.mockResolvedValue(new Response('boom', { status: 500 }))
    await expect(http('/x')).rejects.toBeInstanceOf(ApiError)
  })

  it('204 -> undefined', async () => {
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }))
    expect(await http('/x', { method: 'DELETE' })).toBeUndefined()
  })

  it('claimsApi เรียก method/URL/body ถูกต้อง', async () => {
    fetchMock.mockResolvedValue(ok({}))
    await claimsApi.updateStatus(3, 'APPROVED')
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/claims/3')
    expect(init.method).toBe('PATCH')
    expect(init.body).toBe(JSON.stringify({ status: 'APPROVED' }))
  })
})
