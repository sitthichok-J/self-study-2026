import { beforeEach, describe, expect, it } from 'vitest'
import { createMockApi, DEMO_TOKEN, SEED } from '../mock-api/mockApi'

const auth = `Bearer ${DEMO_TOKEN}`

describe('mock API', () => {
  let api: ReturnType<typeof createMockApi>
  beforeEach(() => { api = createMockApi() })

  it('login สำเร็จ/ล้มเหลว', () => {
    expect(api.handle('POST', '/api/login', undefined, { username: 'demo', password: 'demo1234' }).body).toEqual({ token: DEMO_TOKEN })
    expect(api.handle('POST', '/api/login', undefined, { username: 'demo', password: 'x' }).status).toBe(401)
  })

  it('ไม่มี token -> 401', () => {
    expect(api.handle('GET', '/api/claims', undefined, undefined).status).toBe(401)
  })

  it('GET list / get by id / 404', () => {
    expect(api.handle('GET', '/api/claims', auth, undefined).body).toHaveLength(SEED.length)
    expect(api.handle('GET', '/api/claims/2', auth, undefined).status).toBe(200)
    expect(api.handle('GET', '/api/claims/999', auth, undefined).status).toBe(404)
  })

  it('POST สร้างเคลม PENDING และ validate', () => {
    const ok = api.handle('POST', '/api/claims', auth, { patientRef: 'P-9', hospital: 'H', amount: 10 })
    expect(ok.status).toBe(201)
    expect(ok.body).toMatchObject({ id: 5, status: 'PENDING' })
    expect(api.handle('POST', '/api/claims', auth, { patientRef: '', hospital: 'H', amount: 10 }).status).toBe(422)
  })

  it('PATCH สถานะ / DELETE', () => {
    expect(api.handle('PATCH', '/api/claims/1', auth, { status: 'APPROVED' }).body).toMatchObject({ status: 'APPROVED' })
    expect(api.handle('PATCH', '/api/claims/1', auth, { status: 'XX' }).status).toBe(422)
    expect(api.handle('DELETE', '/api/claims/1', auth, undefined).status).toBe(204)
    expect(api.handle('GET', '/api/claims/1', auth, undefined).status).toBe(404)
  })

  it('แต่ละ instance แยกข้อมูลกัน', () => {
    api.handle('DELETE', '/api/claims/1', auth, undefined)
    expect(createMockApi().handle('GET', '/api/claims/1', auth, undefined).status).toBe(200)
  })
})
