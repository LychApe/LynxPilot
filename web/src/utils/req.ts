import axios from 'axios'
import { message } from 'antdv-next'
import router from '@/router'

const req = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// 请求拦截：自动附带 JWT
req.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：统一错误处理
req.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const status = err.response?.status
    const msg = err.response?.data?.message || '请求失败'

    if (status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }

    message.error(msg)
    return Promise.reject(err)
  },
)

export default req
