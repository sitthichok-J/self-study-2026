import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h, type Component } from 'vue'
import { createMemoryHistory, RouterView } from 'vue-router'
import { provideToast } from '../src/composables/useToast'
import { createAppRouter } from '../src/router'

// ครอบ component ด้วย provideToast เหมือน App.vue จริง
export function withToast(inner: Component, props: Record<string, unknown> = {}) {
  return defineComponent({
    setup() {
      provideToast()
      return () => h(inner, props)
    },
  })
}

export async function mountAt(path: string, comp?: Component): Promise<{ w: VueWrapper; router: ReturnType<typeof createAppRouter> }> {
  const pinia = createPinia()
  setActivePinia(pinia)
  const router = createAppRouter(createMemoryHistory())
  router.push(path)
  await router.isReady()
  const w = mount(comp ?? withToast(RouterView), {
    global: { plugins: [pinia, router] },
  })
  await flushPromises()
  return { w, router }
}
