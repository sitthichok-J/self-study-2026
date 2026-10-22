import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { authApi } from '../api/claimsApi'
import { TOKEN_KEY } from '../api/http'

function readToken(): string | null {
  try { return localStorage.getItem(TOKEN_KEY) } catch { return null }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(readToken())
  const isLoggedIn = computed(() => !!token.value)

  async function login(username: string, password: string) {
    const r = await authApi.login(username, password)
    token.value = r.token
    try { localStorage.setItem(TOKEN_KEY, r.token) } catch { /* ignore */ }
  }
  function logout() {
    token.value = null
    try { localStorage.removeItem(TOKEN_KEY) } catch { /* ignore */ }
  }
  return { token, isLoggedIn, login, logout }
})
