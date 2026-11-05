import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/http'
import * as api from '@/api/auth'
import { safeRedirect } from '@/lib'
import LoginView from '@/views/LoginView.vue'

vi.mock('@/api/auth')
const m = vi.mocked(api)

const b64url = (o: unknown) => btoa(JSON.stringify(o)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
const tok = { access_token: `${b64url({ alg: 'HS256' })}.${b64url({ sub: 'alice', exp: 9999999999 })}.s`, token_type: 'Bearer', expires_in: 600 }

async function setup(path = '/login') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<p>home</p>' } },
      { path: '/claims/:id', component: { template: '<p>claim</p>' } },
      { path: '/login', component: LoginView },
    ],
  })
  router.push(path)
  await router.isReady()
  const w = mount(LoginView, { global: { plugins: [router] } })
  return { w, router }
}

async function fill(w: Awaited<ReturnType<typeof setup>>['w'], u: string, p: string) {
  await w.get('[data-test=username]').setValue(u)
  await w.get('[data-test=password]').setValue(p)
  await w.get('form').trigger('submit')
  await flushPromises()
}

beforeEach(() => {
  vi.resetAllMocks()
  setActivePinia(createPinia())
})

describe('LoginView', () => {
  it('ช่องว่าง: แสดงข้อความและไม่เรียก API', async () => {
    const { w } = await setup()
    await fill(w, '  ', '')
    expect(w.get('[data-test=error]').text()).toContain('กรอก')
    expect(m.loginRequest).not.toHaveBeenCalled()
  })

  it('สำเร็จ: ตัดช่องว่างชื่อผู้ใช้ แล้วไปหน้าแรก', async () => {
    m.loginRequest.mockResolvedValue(tok)
    const { w, router } = await setup()
    await fill(w, ' alice ', 'pw')
    expect(m.loginRequest).toHaveBeenCalledWith('alice', 'pw')
    expect(router.currentRoute.value.fullPath).toBe('/')
  })

  it('กลับไปหน้าที่ถูกเด้งมาตาม ?redirect=', async () => {
    m.loginRequest.mockResolvedValue(tok)
    const { w, router } = await setup('/login?redirect=/claims/7')
    await fill(w, 'alice', 'pw')
    expect(router.currentRoute.value.fullPath).toBe('/claims/7')
  })

  it('redirect ไปเว็บอื่นไม่ได้', async () => {
    m.loginRequest.mockResolvedValue(tok)
    const { w, router } = await setup('/login?redirect=//evil.example')
    await fill(w, 'alice', 'pw')
    expect(router.currentRoute.value.fullPath).toBe('/')
  })

  it('401: ข้อความเดียวกัน ไม่บอกว่าผิดที่ชื่อหรือรหัส และล้างช่องรหัสผ่าน', async () => {
    m.loginRequest.mockRejectedValue(new ApiError(401, 'unauthorized'))
    const { w } = await setup()
    await fill(w, 'alice', 'bad')
    expect(w.get('[data-test=error]').text()).toBe('ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง')
    expect((w.get('[data-test=password]').element as HTMLInputElement).value).toBe('')
  })

  it('error อื่น: ไม่ใช้ข้อความ 401', async () => {
    m.loginRequest.mockRejectedValue(new ApiError(503, 'down'))
    const { w } = await setup()
    await fill(w, 'alice', 'pw')
    expect(w.get('[data-test=error]').text()).not.toBe('ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง')
  })

  it('ช่องรหัสผ่านเป็น type=password และปุ่มถูกปิดระหว่างส่ง', async () => {
    let resolve!: (v: typeof tok) => void
    m.loginRequest.mockReturnValue(new Promise((r) => (resolve = r)))
    const { w } = await setup()
    expect(w.get('[data-test=password]').attributes('type')).toBe('password')
    await w.get('[data-test=username]').setValue('alice')
    await w.get('[data-test=password]').setValue('pw')
    await w.get('form').trigger('submit')
    expect(w.get('[data-test=submit]').attributes('disabled')).toBeDefined()
    await w.get('form').trigger('submit')
    expect(m.loginRequest).toHaveBeenCalledTimes(1)
    resolve(tok)
    await flushPromises()
  })
})

describe('safeRedirect', () => {
  it.each([
    [undefined, '/'], [null, '/'], [5, '/'], ['', '/'], ['claims', '/'], ['//evil.com', '/'], ['/\\evil.com', '/'],
    ['http://evil.com', '/'], ['/a\nb', '/'], ['/login', '/'], ['/login?redirect=/x', '/'],
    ['/claims/1', '/claims/1'], ['/claims?status=PENDING', '/claims?status=PENDING'],
  ])('%j -> %s', (input, want) => expect(safeRedirect(input)).toBe(want))
})
