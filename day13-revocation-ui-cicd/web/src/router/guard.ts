import type { Router } from 'vue-router'
import { useAuth } from '@/stores/auth'

// หน้าที่ต้อง login: ไม่มี session ไปหน้า login พร้อมจำปลายทางไว้ใน ?redirect=
export function installAuthGuard(router: Router, auth = useAuth()): void {
  router.beforeEach(async (to) => {
    await auth.restore()
    if (to.meta.public) {
      if (to.name === 'login' && auth.isAuthenticated) return { name: 'list' }
      return true
    }
    if (!auth.isAuthenticated) return { name: 'login', query: { redirect: to.fullPath } }
    return true
  })
}
