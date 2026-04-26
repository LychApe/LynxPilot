import req from '@/utils/req'

export function getStatus() {
  return req.get('/public/server/status')
}
