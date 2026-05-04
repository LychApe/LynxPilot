import req from '@/utils/req'

interface PublicServerStatus {
  installed: boolean
}

export function getStatus<T = PublicServerStatus>() {
  return req.get<unknown, { data: T }>('/public/server/status')
}

export function getPrivateStatus() {
  return req.get('/private/server/status')
}
