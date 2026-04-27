import req from '@/utils/req'

export function getStatus() {
  return req.get('/public/server/status')
}

export function getPrivateStatus() {
  return req.get('/private/server/status')
}
