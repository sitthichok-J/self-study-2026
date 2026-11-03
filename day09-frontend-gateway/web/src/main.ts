import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import { makeRouter } from './router'
import './styles.css'

createApp(App).use(createPinia()).use(makeRouter()).mount('#app')
