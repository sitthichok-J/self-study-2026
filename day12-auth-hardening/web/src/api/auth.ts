import { authRequest } from './http'

export interface TokenResponse {
  access_token: string
  token_type: string
  expires_in: number // วินาที
}

// refresh token อยู่ใน cookie HttpOnly ที่ JS อ่านไม่ได้ มีแต่ access token เท่านั้นที่กลับมาใน body
export async function loginRequest(username: string, password: string): Promise<TokenResponse> {
  return (await authRequest<TokenResponse>('/auth/login', { username, password }))!
}

export async function refreshRequest(): Promise<TokenResponse> {
  return (await authRequest<TokenResponse>('/auth/refresh'))!
}

export async function logoutRequest(): Promise<void> {
  await authRequest('/auth/logout')
}
