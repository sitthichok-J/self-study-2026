import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory } from 'vue-router'
import { ApiError } from '@/api/http'
import * as api from '@/api/auth'
import { makeRouter } from '@/router'
import { installAuthGuard } from '@/router/guard'
import { useAuth } from '@/stores/auth'

vi.mock('@/api/auth')
const m = vi.mocked(api)
const tokens = { access_token: 'x.eyJzdWIiOiJhbGljZSJ9.y', token_type: 'Bearer', expires_in: 600 }

function setup() {
  setActivePinia(createPinia())
  const auth = useAuth()
  const router = makeRouter(createMemoryHistory())
  installAuthGuard(router, auth)
  return { auth, router }
}

beforeEach(() => vi.resetAllMocks())

describe('auth guard', () => {
  it('ยังไม่ login: ทุกหน้าที่ต้อง login ไป /login พร้อมจำปลายทาง', async () => {
    m.refreshRequest.mockRejectedValue(new ApiError(401, 'unauthorized'))
    const { router } = setup()
    for (const path of ['/', '/new', '/claims/abc?x=1']) {
      await router.push(path)
      expect(router.currentRoute.value.name).toBe('login')
      expect(router.currentRoute.value.query.redirect).toBe(path)
    }
  })

  it('มี refresh cookie: restore สำเร็จแล้วเข้าหน้าที่ขอได้เลย ไม่เห็นหน้า login', async () => {
    m.refreshRequest.mockResolvedValue(tokens)
    const { router, auth } = setup()
    await router.push('/claims/abc')
    expect(router.currentRoute.value.name).toBe('detail')
    expect(auth.isAuthenticated).toBe(true)
  })

  it('login แล้วเปิด /login จะถูกพากลับหน้ารายการ', async () => {
    m.refreshRequest.mockResolvedValue(tokens)
    const { router } = setup()
    await router.push('/login')
    expect(router.currentRoute.value.name).toBe('list')
  })

  it('/login เข้าได้โดยไม่ login และไม่วนซ้ำ', async () => {
    m.refreshRequest.mockRejectedValue(new ApiError(401, 'unauthorized'))
    const { router } = setup()
    await router.push('/login')
    expect(router.currentRoute.value.name).toBe('login')
  })

  it('หน้า not-found เปิดได้โดยไม่ login', async () => {
    m.refreshRequest.mockRejectedValue(new ApiError(401, 'unauthorized'))
    const { router } = setup()
    await router.push('/no/such/page')
    expect(router.currentRoute.value.name).toBe('not-found')
  })

  it('restore ถูกเรียกครั้งเดียวแม้เปลี่ยนหน้าหลายครั้ง', async () => {
    m.refreshRequest.mockResolvedValue(tokens)
    const { router } = setup()
    await router.push('/')
    await router.push('/new')
    await router.push('/claims/z')
    expect(m.refreshRequest).toHaveBeenCalledTimes(1)
  })

  it('logout แล้วเข้าหน้าที่ต้อง login จะถูกเด้งไป login', async () => {
    m.refreshRequest.mockResolvedValue(tokens)
    m.logoutRequest.mockResolvedValue()
    const { router, auth } = setup()
    await router.push('/')
    await auth.logout()
    await router.push('/new')
    expect(router.currentRoute.value.name).toBe('login')
  })
})
