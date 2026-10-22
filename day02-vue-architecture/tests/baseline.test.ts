// Baseline ของ Day 2: บันทึกโครงสร้าง route, state เริ่มต้นของ store, และหน้าตาของแต่ละ view
// แดง = พฤติกรรม/หน้าตาเปลี่ยน ถ้าตั้งใจ ให้รัน `npm run test:update` แล้วดู diff ก่อน commit
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { claimsApi } from '../src/api/claimsApi'
import { TOKEN_KEY } from '../src/api/http'
import { routes } from '../src/router'
import { useAuthStore } from '../src/stores/auth'
import { useClaimsStore } from '../src/stores/claims'
import { STATUS_LABEL, type Claim } from '../src/types/claim'
import { mountAt } from './helpers'

vi.mock('../src/api/claimsApi', () => ({
  claimsApi: { list: vi.fn(), get: vi.fn(), create: vi.fn(), updateStatus: vi.fn(), remove: vi.fn() },
  authApi: { login: vi.fn() },
}))

const claims: Claim[] = [
  { id: 1, patientRef: 'P-1001', hospital: 'รพ.ตัวอย่าง A', amount: 12500, status: 'PENDING' },
  { id: 2, patientRef: 'P-1002', hospital: 'รพ.ตัวอย่าง B', amount: 8200.5, status: 'APPROVED' },
]

beforeEach(() => {
  vi.resetAllMocks()
  localStorage.clear()
  localStorage.setItem(TOKEN_KEY, 't')
  document.body.innerHTML = ''
})

describe('baseline: ค่าคงที่และโครงสร้าง', () => {
  it('STATUS_LABEL', () => {
    expect(STATUS_LABEL).toEqual({ PENDING: 'รอตรวจสอบ', APPROVED: 'อนุมัติ', REJECTED: 'ปฏิเสธ' })
  })

  it('ตาราง route (path, name, public)', () => {
    expect(routes.map((r) => [r.path, r.name ?? '(redirect)', !!r.meta?.public])).toMatchInlineSnapshot(`
      [
        [
          "/",
          "(redirect)",
          false,
        ],
        [
          "/login",
          "login",
          true,
        ],
        [
          "/claims",
          "claims",
          false,
        ],
        [
          "/claims/new",
          "claim-new",
          false,
        ],
        [
          "/claims/:id(\\d+)",
          "claim-detail",
          false,
        ],
        [
          "/:pathMatch(.*)*",
          "not-found",
          true,
        ],
      ]
    `)
  })

  it('state เริ่มต้นของ store', () => {
    setActivePinia(createPinia())
    const c = useClaimsStore()
    expect({ items: c.items, loading: c.loading, error: c.error, keyword: c.keyword, statusFilter: c.statusFilter }).toEqual({
      items: [], loading: false, error: null, keyword: '', statusFilter: 'ALL',
    })
    localStorage.clear()
    setActivePinia(createPinia())
    expect(useAuthStore().isLoggedIn).toBe(false)
  })
})

describe('baseline: snapshot ของแต่ละหน้า', () => {
  it('login', async () => {
    localStorage.clear()
    expect((await mountAt('/login')).w.html()).toMatchSnapshot()
  })
  it('list', async () => {
    vi.mocked(claimsApi.list).mockResolvedValue(claims)
    expect((await mountAt('/claims')).w.html()).toMatchSnapshot()
  })
  it('detail', async () => {
    vi.mocked(claimsApi.get).mockResolvedValue(claims[0])
    expect((await mountAt('/claims/1')).w.html()).toMatchSnapshot()
  })
  it('form', async () => {
    expect((await mountAt('/claims/new')).w.html()).toMatchSnapshot()
  })
  it('not found', async () => {
    expect((await mountAt('/nope')).w.html()).toMatchSnapshot()
  })
})
