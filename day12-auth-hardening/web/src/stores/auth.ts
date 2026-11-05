import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { loginRequest, logoutRequest, refreshRequest, type TokenResponse } from '@/api/auth'

// อ่านชื่อผู้ใช้จาก payload ของ JWT "เพื่อแสดงผลเท่านั้น" ห้ามใช้ตัดสินสิทธิ์
// (ฝั่ง UI ตรวจลายเซ็นไม่ได้ สิทธิ์จริงถูกตรวจที่ Kong และ API)
export function subjectOf(jwt: string): string {
  try {
    const part = jwt.split('.')[1] ?? ''
    const json = atob(part.replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(part.length / 4) * 4, '='))
    const sub = (JSON.parse(json) as { sub?: unknown }).sub
    return typeof sub === 'string' ? sub : ''
  } catch {
    return ''
  }
}

export const useAuth = defineStore('auth', () => {
  // access token อยู่ในหน่วยความจำเท่านั้น: ไม่เขียน localStorage/sessionStorage
  // (XSS อ่านได้ทั้งคู่) รีโหลดหน้าแล้วหาย ต้องขอใหม่จาก refresh cookie (restore)
  const token = ref<string | null>(null)
  const username = ref('')
  const restored = ref(false)
  const isAuthenticated = computed(() => token.value !== null)

  function setSession(r: TokenResponse) {
    token.value = r.access_token
    username.value = subjectOf(r.access_token)
  }

  function clear() {
    token.value = null
    username.value = ''
  }

  // refresh token หมุนทุกครั้ง: ถ้ามี 401 หลายตัวพร้อมกันแล้วต่างคนต่างยิง refresh ด้วย token เดียวกัน
  // ตัวที่สองจะถูกมองเป็นการใช้ซ้ำและ session ทั้งสายจะถูกเพิกถอน จึงต้องรวมเป็นคำขอเดียว (single-flight)
  let inflight: Promise<boolean> | null = null
  function refresh(): Promise<boolean> {
    if (!inflight) {
      inflight = refreshRequest()
        .then((r) => {
          setSession(r)
          return true
        })
        .catch(() => {
          clear()
          return false
        })
        .finally(() => {
          inflight = null
        })
    }
    return inflight
  }

  // เปิดหน้าใหม่/รีโหลด: ลองขอ access token จาก cookie หนึ่งครั้ง
  async function restore(): Promise<void> {
    if (restored.value) return
    if (!token.value) await refresh()
    restored.value = true
  }

  async function login(user: string, password: string): Promise<void> {
    setSession(await loginRequest(user, password))
    restored.value = true
  }

  async function logout(): Promise<void> {
    try {
      await logoutRequest()
    } catch {
      /* เครือข่ายล่ม: อย่างน้อยล้างฝั่งเครื่องนี้ (cookie จะหมดอายุเอง) */
    }
    clear()
  }

  return { token, username, restored, isAuthenticated, refresh, restore, login, logout, clear }
})
