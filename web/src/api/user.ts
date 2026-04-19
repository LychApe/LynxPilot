import req from '@/utils/req'

export function register(data: { username: string; password: string; email: string }) {
  return req.post('/public/user/register', data)
}

export function login(data: { username: string; password: string }) {
  return req.post<{ token: string; expires_at: string }>('/public/user/login', data)
}
