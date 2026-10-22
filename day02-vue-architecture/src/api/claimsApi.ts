import type { Claim, ClaimStatus, NewClaim } from '../types/claim'
import { http } from './http'

export const claimsApi = {
  list: () => http<Claim[]>('/api/claims'),
  get: (id: number) => http<Claim>(`/api/claims/${id}`),
  create: (c: NewClaim) => http<Claim>('/api/claims', { method: 'POST', body: c }),
  updateStatus: (id: number, status: ClaimStatus) =>
    http<Claim>(`/api/claims/${id}`, { method: 'PATCH', body: { status } }),
  remove: (id: number) => http<void>(`/api/claims/${id}`, { method: 'DELETE' }),
}

export const authApi = {
  login: (username: string, password: string) =>
    http<{ token: string }>('/api/login', { method: 'POST', body: { username, password } }),
}
