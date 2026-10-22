import vue from '@vitejs/plugin-vue'
import type { Connect, Plugin } from 'vite'
import { defineConfig } from 'vite'
import { createMockApi } from './mock-api/mockApi.ts'

// ใช้ mock API ใน dev/preview ถ้าจะต่อ API จริง (เช่น Go) ให้ตั้ง VITE_API_TARGET แล้วใช้ proxy แทน
function mockApiPlugin(): Plugin {
  const api = createMockApi()
  const middleware: Connect.NextHandleFunction = (req, res, next) => {
    const url = req.originalUrl ?? req.url ?? ''
    if (!url.startsWith('/api')) return next()
    let raw = ''
    req.on('data', (chunk) => (raw += chunk))
    req.on('end', () => {
      let body: unknown
      try {
        body = raw ? JSON.parse(raw) : undefined
      } catch {
        body = undefined
      }
      const r = api.handle(req.method ?? 'GET', url, req.headers.authorization, body)
      res.statusCode = r.status
      res.setHeader('Content-Type', 'application/json')
      res.end(r.body === null ? '' : JSON.stringify(r.body))
    })
  }
  return {
    name: 'mock-api',
    configureServer: (s) => void s.middlewares.use(middleware),
    configurePreviewServer: (s) => void s.middlewares.use(middleware),
  }
}

const target = process.env.VITE_API_TARGET

export default defineConfig({
  plugins: [vue(), ...(target ? [] : [mockApiPlugin()])],
  server: target ? { proxy: { '/api': { target, changeOrigin: true } } } : {},
})
