import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import { configureAuth } from './api/http'
import { makeRouter } from './router'
import { installAuthGuard } from './router/guard'
import { useAuth } from './stores/auth'
import './styles.css'

const app = createApp(App)
const pinia = createPinia()
const router = makeRouter()
app.use(pinia).use(router)

// ผูก http() กับ store: แนบ Bearer, ขอ token ใหม่เมื่อเจอ 401 และพาไปหน้า login ถ้า session จบ
const auth = useAuth(pinia)
installAuthGuard(router, auth)
configureAuth({
  getToken: () => auth.token,
  refresh: auth.refresh,
  onLoggedOut: () => {
    auth.clear()
    void router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
  },
})

app.mount('#app')
