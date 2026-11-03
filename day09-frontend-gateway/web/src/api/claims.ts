import type { Claim, ClaimPage, ClaimStatus, NewClaim } from '@/types/claim'
import { http } from './http'

export interface ListParams {
  status?: ClaimStatus | ''
  limit?: number
  offset?: number
}

export function listClaims(p: ListParams = {}): Promise<ClaimPage> {
  const q = new URLSearchParams()
  if (p.status) q.set('status', p.status)
  if (p.limit !== undefined) q.set('limit', String(p.limit))
  if (p.offset !== undefined) q.set('offset', String(p.offset))
  const qs = q.toString()
  return http<ClaimPage>(`/claims${qs ? `?${qs}` : ''}`).then((r) => r.data)
}

export function getClaim(id: string): Promise<Claim> {
  return http<Claim>(`/claims/${encodeURIComponent(id)}`).then((r) => r.data)
}

export interface CreateResult {
  claim: Claim
  replayed: boolean // true = backend ตอบของเดิมเพราะ Idempotency-Key ซ้ำ
}

export async function createClaim(body: NewClaim, idempotencyKey: string): Promise<CreateResult> {
  const r = await http<Claim>('/claims', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Idempotency-Key': idempotencyKey },
    body: JSON.stringify(body),
  })
  return { claim: r.data, replayed: r.headers.get('Idempotent-Replayed') === 'true' }
}
