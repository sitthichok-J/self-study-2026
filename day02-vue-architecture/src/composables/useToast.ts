import { inject, provide, ref, type InjectionKey, type Ref } from 'vue'

export interface Toast { id: number; text: string }
export interface ToastApi { toasts: Ref<Toast[]>; show: (text: string, ms?: number) => void }

const ToastKey: InjectionKey<ToastApi> = Symbol('toast')

// provide/inject: ส่ง service ลง tree ลึกๆ โดยไม่ต้องส่ง props ทีละชั้น
export function provideToast(): ToastApi {
  const toasts = ref<Toast[]>([])
  let n = 0
  const show = (text: string, ms = 2500) => {
    const id = ++n
    toasts.value.push({ id, text })
    if (ms > 0) setTimeout(() => { toasts.value = toasts.value.filter((t) => t.id !== id) }, ms)
  }
  const api = { toasts, show }
  provide(ToastKey, api)
  return api
}

export function useToast(): ToastApi {
  const api = inject(ToastKey)
  if (!api) throw new Error('useToast() ต้องอยู่ใต้คอมโพเนนต์ที่เรียก provideToast()')
  return api
}
