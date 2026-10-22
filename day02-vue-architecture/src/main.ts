import { createPinia } from 'pinia'
import { createApp } from 'vue'
import './styles.css'
import App from './App.vue'
import { createAppRouter } from './router'
import { installUnauthorizedRedirect } from './router/session'

const router = createAppRouter()
createApp(App).use(createPinia()).use(router).mount('#app')
installUnauthorizedRedirect(router)
