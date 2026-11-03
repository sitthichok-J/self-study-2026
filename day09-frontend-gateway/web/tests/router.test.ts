import { createMemoryHistory } from 'vue-router'
import { describe, expect, it } from 'vitest'
import { makeRouter, routes } from '@/router'

describe('router', () => {
  const resolve = async (path: string) => {
    const r = makeRouter(createMemoryHistory())
    await r.push(path)
    return r.currentRoute.value
  }

  it('เส้นทางหลัก', async () => {
    expect((await resolve('/')).name).toBe('list')
    expect((await resolve('/new')).name).toBe('new')
    const d = await resolve('/claims/abc')
    expect(d.name).toBe('detail')
    expect(d.params.id).toBe('abc')
  })

  it('path ที่ไม่รู้จักไปหน้า not-found (nginx ส่ง index.html ทุก path มาให้ SPA)', async () => {
    expect((await resolve('/no/such/page')).name).toBe('not-found')
  })

  it('รายชื่อ route ไม่เปลี่ยนโดยไม่ตั้งใจ', () => {
    expect(routes.map((r) => r.name)).toEqual(['list', 'new', 'detail', 'not-found'])
  })
})
