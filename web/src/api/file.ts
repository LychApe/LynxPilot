import req from '@/utils/req'

export interface FileEntry {
  name: string
  path: string
  is_dir: boolean
  size: number
  mod_time: string
  mode: string
}

export interface ListResult {
  path: string
  parent: string
  entries: FileEntry[]
}

export interface FileDetail {
  name: string
  path: string
  is_dir: boolean
  size: number
  mod_time: string
  mode: string
}

export function listFiles(path = '') {
  return req.get<unknown, { data: ListResult }>('/private/file/list', {
    params: { path },
  })
}

export function getFileInfo(path: string) {
  return req.get<unknown, { data: FileDetail }>('/private/file/info', {
    params: { path },
  })
}

export function readFile(path: string) {
  return req.get<unknown, { data: { content: string; path: string } }>('/private/file/read', {
    params: { path },
  })
}

export function saveFile(path: string, content: string) {
  return req.post('/private/file/save', { path, content })
}

export function createDir(path: string) {
  return req.post('/private/file/mkdir', { path })
}

export function createFile(path: string) {
  return req.post('/private/file/touch', { path })
}

export function deleteFile(path: string) {
  return req.post('/private/file/delete', { path })
}

export function renameFile(path: string, new_name: string) {
  return req.post('/private/file/rename', { path, new_name })
}

export function uploadFile(path: string, file: File) {
  const form = new FormData()
  form.append('file', file)
  form.append('path', path)
  return req.post('/private/file/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000,
  })
}

export function downloadFile(path: string) {
  return req.get('/private/file/download', {
    params: { path },
    responseType: 'blob',
  })
}

export function getBasePath() {
  return req.get<unknown, { data: { base_path: string } }>('/private/file/base-path')
}

export function setBasePath(base_path: string) {
  return req.put('/private/file/base-path', { base_path })
}
