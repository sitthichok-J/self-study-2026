import { vi } from 'vitest'

export interface FakeResponse {
  status?: number
  body?: unknown
  headers?: Record<string, string>
  rawText?: string
}

export function jsonResponse({ status = 200, body = {}, headers = {}, rawText }: FakeResponse = {}): Response {
  return new Response(rawText ?? JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json', ...headers },
  })
}

export function stubFetch(...responses: Array<FakeResponse | Error>) {
  const fn = vi.fn()
  for (const r of responses) {
    if (r instanceof Error) fn.mockRejectedValueOnce(r)
    else fn.mockResolvedValueOnce(jsonResponse(r))
  }
  vi.stubGlobal('fetch', fn)
  return fn
}

export const sampleClaim = (over: Record<string, unknown> = {}) => ({
  id: '11111111-2222-3333-4444-555555555555',
  patient_ref: 'FAKE-AB12',
  amount: 150000,
  status: 'PENDING',
  created_at: '2026-11-03T01:00:00Z',
  updated_at: '2026-11-03T01:00:00Z',
  ...over,
})

export const flush = () => new Promise<void>((r) => setTimeout(r, 0))
