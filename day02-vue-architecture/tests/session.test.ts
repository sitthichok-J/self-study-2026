import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory } from 'vue-router'
import { http, setUnauthorizedHandler, TOKEN_KEY } from '../src/api/http'
import { createAppRouter } from '../src/router'
import { installUnauthorizedRedirect } from '../src/router/session'
import { useAuthStore } from '../src/stores/auth'

const res = (status: number) => new Response(JSON.stringify({ message: 'x' }), { status })

describe('401 handling', () => {
  const fetchMock = vi.fn()
  beforeEach(() => {
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
    localStorage.clear()
    setActivePinia(createPinia())
  })
  afterEach(() => {
    vi.unstubAllGlobals()
    setUnauthorizedHandler(null)
  })

  it('http: 401 ที่ไม่ใช่ /api/login เรียก handler', async () => {
    const h = vi.fn()
    setUnauthorizedHandler(h)
    fetchMock.mockResolvedValue(res(401))
    await expect(http('/api/claims')).rejects.toMatchObject({ status: 401 })
    expect(h).toHaveBeenCalledTimes(1)
  })

  it('http: 401 ของ /api/login (รหัสผิด) ไม่เรียก handler', async () => {
    const h = vi.fn()
    setUnauthorizedHandler(h)
    fetchMock.mockResolvedValue(res(401))
    await expect(http('/api/login', { method: 'POST', body: {} })).rejects.toBeDefined()
    expect(h).not.toHaveBeenCalled()
  })

  it('http: error อื่น (500) ไม่เรียก handler', async () => {
    const h = vi.fn()
    setUnauthorizedHandler(h)
    fetchMock.mockResolvedValue(res(500))
    await expect(http('/api/claims')).rejects.toBeDefined()
    expect(h).not.toHaveBeenCalled()
  })

  it('token หมดอายุกลางทาง -> logout และไป /login พร้อม redirect', async () => {
    localStorage.setItem(TOKEN_KEY, 'old')
    setActivePinia(createPinia())
    const router = createAppRouter(createMemoryHistory())
    installUnauthorizedRedirect(router)
    await router.push('/claims/3')
    expect(router.currentRoute.value.name).toBe('claim-detail')

    fetchMock.mockResolvedValue(res(401))
    await expect(http('/api/claims/3')).rejects.toBeDefined()

    await vi.waitFor(() => expect(router.currentRoute.value.name).toBe('login'), { timeout: 5000 })
    expect(router.currentRoute.value.query.redirect).toBe('/claims/3')
    expect(useAuthStore().isLoggedIn).toBe(false)
    expect(localStorage.getItem(TOKEN_KEY)).toBeNull()
  })

  it('ไม่ได้ login อยู่แล้ว -> 401 ไม่ทำอะไรซ้ำ', async () => {
    const router = createAppRouter(createMemoryHistory())
    installUnauthorizedRedirect(router)
    await router.push('/login')
    fetchMock.mockResolvedValue(res(401))
    await expect(http('/api/claims')).rejects.toBeDefined()
    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBeUndefined()
  })
})
