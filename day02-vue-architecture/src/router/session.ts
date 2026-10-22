import type { Router } from 'vue-router'
import { setUnauthorizedHandler } from '../api/http'
import { useAuthStore } from '../stores/auth'

// เจอ 401 กลางทาง: ล้าง token แล้วไปหน้า login พร้อมจำหน้าที่อยู่ไว้ใน redirect
export function installUnauthorizedRedirect(router: Router) {
  setUnauthorizedHandler(() => {
    const auth = useAuthStore()
    if (!auth.isLoggedIn) return
    auth.logout()
    const here = router.currentRoute.value
    if (here.name === 'login') return
    router.replace({ name: 'login', query: { redirect: here.fullPath } })
  })
}
