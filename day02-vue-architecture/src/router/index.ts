import {
  createRouter,
  createWebHistory,
  type NavigationGuard,
  type RouteRecordRaw,
  type RouterHistory,
} from 'vue-router'
import { useAuthStore } from '../stores/auth'

declare module 'vue-router' {
  interface RouteMeta { public?: boolean; title?: string }
}

// navigation guard: ต้อง login ก่อน (ยกเว้นหน้าที่ meta.public)
export const authGuard: NavigationGuard = (to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    return auth.isLoggedIn && to.name === 'login' ? { name: 'claims' } : true
  }
  if (!auth.isLoggedIn) return { name: 'login', query: { redirect: to.fullPath } }
  return true
}

export const routes: RouteRecordRaw[] = [
  { path: '/', redirect: { name: 'claims' } },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'), // lazy loading: แยก chunk
    meta: { public: true, title: 'เข้าสู่ระบบ' },
  },
  {
    path: '/claims',
    name: 'claims',
    component: () => import('../views/ClaimListView.vue'),
    meta: { title: 'รายการเคลม' },
  },
  {
    path: '/claims/new',
    name: 'claim-new',
    component: () => import('../views/ClaimFormView.vue'),
    meta: { title: 'เพิ่มเคลม' },
  },
  {
    path: '/claims/:id(\\d+)', // dynamic route (เฉพาะตัวเลข)
    name: 'claim-detail',
    component: () => import('../views/ClaimDetailView.vue'),
    props: (route) => ({ id: Number(route.params.id) }),
    meta: { title: 'รายละเอียดเคลม' },
  },
  { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('../views/NotFoundView.vue'), meta: { public: true } },
]

export function createAppRouter(history: RouterHistory = createWebHistory()) {
  const router = createRouter({ history, routes })
  router.beforeEach(authGuard)
  router.afterEach((to) => {
    document.title = `${to.meta.title ?? 'Claim Console'} · Day 2`
  })
  return router
}
