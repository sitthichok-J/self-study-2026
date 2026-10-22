import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { createMemoryHistory } from 'vue-router'
import { TOKEN_KEY } from '../src/api/http'
import { createAppRouter, routes } from '../src/router'

function setup(loggedIn: boolean) {
  localStorage.clear()
  if (loggedIn) localStorage.setItem(TOKEN_KEY, 't')
  setActivePinia(createPinia())
  return createAppRouter(createMemoryHistory())
}

describe('router + guard', () => {
  beforeEach(() => localStorage.clear())

  it('ยังไม่ login -> redirect ไป /login พร้อม redirect query', async () => {
    const r = setup(false)
    await r.push('/claims/2')
    expect(r.currentRoute.value.name).toBe('login')
    expect(r.currentRoute.value.query.redirect).toBe('/claims/2')
  })

  it('login แล้วเข้าหน้าได้ และ / ไป claims', async () => {
    const r = setup(true)
    await r.push('/')
    expect(r.currentRoute.value.name).toBe('claims')
  })

  it('login แล้วเข้า /login -> เด้งไป claims', async () => {
    const r = setup(true)
    await r.push('/login')
    expect(r.currentRoute.value.name).toBe('claims')
  })

  it('dynamic route: id เป็นตัวเลข, ไม่ใช่ตัวเลข -> not-found', async () => {
    const r = setup(true)
    await r.push('/claims/12')
    expect(r.currentRoute.value.name).toBe('claim-detail')
    expect(r.currentRoute.value.params.id).toBe('12')
    await r.push('/claims/abc')
    expect(r.currentRoute.value.name).toBe('not-found')
  })

  it('ทุกหน้าที่ไม่ใช่ public ต้องมี title และ component เป็น lazy', () => {
    for (const rt of routes.filter((x) => x.name && x.name !== 'not-found')) {
      expect(typeof rt.component).toBe('function')
      expect(rt.meta?.title).toBeTruthy()
    }
  })

  it('afterEach ตั้ง document.title', async () => {
    const r = setup(true)
    await r.push('/claims/new')
    expect(document.title).toBe('เพิ่มเคลม · Day 2')
  })
})
