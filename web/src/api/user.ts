import { md5 } from 'js-md5'
import req from '@/utils/req'

interface LoginResponse {
  data: {
    token: string
    expires_at: string
  }
}

// 注册接口
export function register(data: { username: string; password: string; email: string }) {
  return req.post('/public/user/register', { ...data, password: md5(data.password) })
}

// 登录接口
export function login(data: { username: string; password: string }) {
  return req.post<unknown, LoginResponse>('/public/user/login', { ...data, password: md5(data.password) })
}
