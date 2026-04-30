<script setup lang="ts">
import type { FileEntry, ListResult } from '@/api/file'
import {
  createDir,
  createFile,
  deleteFile,
  getBasePath,
  listFiles,
  readFile,
  renameFile,
  saveFile,
  setBasePath,
  uploadFile,
} from '@/api/file'
import {
  DeleteOutlined,
  EditOutlined,
  FileOutlined,
  FolderOpenOutlined,
  FolderOutlined,
  HomeOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
  UploadOutlined,
} from '@antdv-next/icons'
import { Modal, message } from 'antdv-next'
import { computed, h, onMounted, ref, shallowRef } from 'vue'

const loading = shallowRef(false)
const currentPath = ref('')
const parentPath = ref('')
const files = ref<FileEntry[]>([])
const searchText = ref('')

const settingVisible = shallowRef(false)
const settingLoading = shallowRef(false)
const basePathValue = ref('/')

const mkdirVisible = shallowRef(false)
const mkdirLoading = shallowRef(false)
const mkdirName = ref('')

const touchVisible = shallowRef(false)
const touchLoading = shallowRef(false)
const touchName = ref('')

const renameVisible = shallowRef(false)
const renameLoading = shallowRef(false)
const renameOldName = ref('')
const renamePath = ref('')
const renameNewName = ref('')

const editorVisible = shallowRef(false)
const editorLoading = shallowRef(false)
const editorSaving = shallowRef(false)
const editorPath = ref('')
const editorContent = ref('')

const uploading = shallowRef(false)

const filteredFiles = computed(() => {
  if (!searchText.value) return files.value
  const kw = searchText.value.toLowerCase()
  return files.value.filter((f) => f.name.toLowerCase().includes(kw))
})

const dirCount = computed(() => files.value.filter((f) => f.is_dir).length)
const fileCount = computed(() => files.value.filter((f) => !f.is_dir).length)

function formatSize(size: number) {
  if (size === 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let val = size
  while (val >= 1024 && i < units.length - 1) {
    val /= 1024
    i++
  }
  return `${val.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString()
}

function getFileIcon(file: FileEntry) {
  return file.is_dir ? FolderOutlined : FileOutlined
}

async function fetchFiles(path = '') {
  loading.value = true
  try {
    const res = await listFiles(path)
    const data = res.data as ListResult
    currentPath.value = data.path ?? ''
    parentPath.value = data.parent ?? ''
    files.value = data.entries ?? []
  } catch {
    message.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

function navigateTo(path: string) {
  fetchFiles(path)
}

function goHome() {
  fetchFiles('')
}

function goBack() {
  if (parentPath.value !== undefined) {
    fetchFiles(parentPath.value)
  }
}

async function handleOpen(file: FileEntry) {
  if (file.is_dir) {
    navigateTo(file.path)
  } else {
    await openEditor(file.path)
  }
}

async function openEditor(path: string) {
  editorVisible.value = true
  editorLoading.value = true
  editorPath.value = path
  editorContent.value = ''
  try {
    const res = await readFile(path)
    editorContent.value = res.data?.content ?? ''
  } catch {
    message.error('读取文件失败')
  } finally {
    editorLoading.value = false
  }
}

async function handleSaveFile() {
  editorSaving.value = true
  try {
    await saveFile(editorPath.value, editorContent.value)
    message.success('文件已保存')
  } catch {
    message.error('保存失败')
  } finally {
    editorSaving.value = false
  }
}

function handleMkdir() {
  mkdirName.value = ''
  mkdirVisible.value = true
}

async function confirmMkdir() {
  if (!mkdirName.value.trim()) {
    message.warning('请输入目录名称')
    return
  }
  mkdirLoading.value = true
  try {
    const fullPath = currentPath.value ? `${currentPath.value}/${mkdirName.value.trim()}` : mkdirName.value.trim()
    await createDir(fullPath)
    message.success('目录已创建')
    mkdirVisible.value = false
    await fetchFiles(currentPath.value)
  } catch {
    message.error('创建目录失败')
  } finally {
    mkdirLoading.value = false
  }
}

function handleTouch() {
  touchName.value = ''
  touchVisible.value = true
}

async function confirmTouch() {
  if (!touchName.value.trim()) {
    message.warning('请输入文件名称')
    return
  }
  touchLoading.value = true
  try {
    const fullPath = currentPath.value ? `${currentPath.value}/${touchName.value.trim()}` : touchName.value.trim()
    await createFile(fullPath)
    message.success('文件已创建')
    touchVisible.value = false
    await fetchFiles(currentPath.value)
  } catch {
    message.error('创建文件失败')
  } finally {
    touchLoading.value = false
  }
}

function handleRename(file: FileEntry) {
  renamePath.value = file.path
  renameOldName.value = file.name
  renameNewName.value = file.name
  renameVisible.value = true
}

async function confirmRename() {
  if (!renameNewName.value.trim()) {
    message.warning('请输入新名称')
    return
  }
  renameLoading.value = true
  try {
    await renameFile(renamePath.value, renameNewName.value.trim())
    message.success('重命名成功')
    renameVisible.value = false
    await fetchFiles(currentPath.value)
  } catch {
    message.error('重命名失败')
  } finally {
    renameLoading.value = false
  }
}

function handleDelete(file: FileEntry) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除 "${file.name}" 吗？${file.is_dir ? '目录下所有内容将被删除。' : ''}`,
    icon: () => h(DeleteOutlined),
    okType: 'danger',
    okText: '删除',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteFile(file.path)
        message.success('已删除')
        await fetchFiles(currentPath.value)
      } catch {
        message.error('删除失败')
      }
    },
  })
}

async function handleUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true
  input.onchange = async () => {
    if (!input.files || input.files.length === 0) return
    uploading.value = true
    try {
      for (const file of Array.from(input.files)) {
        await uploadFile(currentPath.value, file)
      }
      message.success('上传成功')
      await fetchFiles(currentPath.value)
    } catch {
      message.error('上传失败')
    } finally {
      uploading.value = false
    }
  }
  input.click()
}

async function openSettings() {
  settingVisible.value = true
  settingLoading.value = true
  try {
    const res = await getBasePath()
    basePathValue.value = res.data?.base_path ?? '/'
  } catch {
    basePathValue.value = '/'
  } finally {
    settingLoading.value = false
  }
}

async function saveSettings() {
  settingLoading.value = true
  try {
    await setBasePath(basePathValue.value)
    message.success('设置已保存')
    settingVisible.value = false
    await fetchFiles('')
  } catch {
    message.error('保存设置失败')
  } finally {
    settingLoading.value = false
  }
}

const breadcrumbs = computed(() => {
  if (!currentPath.value) return []
  return currentPath.value.split('/').filter(Boolean)
})

function breadcrumbPath(index: number) {
  return breadcrumbs.value.slice(0, index + 1).join('/')
}

onMounted(() => fetchFiles())
</script>

<template>
  <div class="file-page">
    <div class="file-heading">
      <div>
        <span class="file-eyebrow">Files</span>
        <h2 class="file-title">文件管理</h2>
      </div>
      <div class="file-heading-actions">
        <a-button @click="openSettings"><template #icon><SettingOutlined /></template>根目录设置</a-button>
        <a-button @click="fetchFiles(currentPath)"><template #icon><ReloadOutlined :spin="loading" /></template>刷新</a-button>
      </div>
    </div>

    <div class="stats-grid">
      <div class="stat-card"><div class="stat-value">{{ files.length }}</div><div class="stat-label">全部</div></div>
      <div class="stat-card stat-dir"><div class="stat-value">{{ dirCount }}</div><div class="stat-label">目录</div></div>
      <div class="stat-card stat-file"><div class="stat-value">{{ fileCount }}</div><div class="stat-label">文件</div></div>
    </div>

    <div class="file-toolbar">
      <a-input-search v-model:value="searchText" placeholder="搜索文件名..." class="search-input" allow-clear>
        <template #prefix><SearchOutlined /></template>
      </a-input-search>
      <div class="toolbar-actions">
        <a-button @click="goHome"><template #icon><HomeOutlined /></template></a-button>
        <a-button :disabled="!currentPath" @click="goBack"><template #icon><FolderOpenOutlined /></template>上级</a-button>
        <a-button @click="handleMkdir"><template #icon><PlusOutlined /></template>新建目录</a-button>
        <a-button @click="handleTouch"><template #icon><FileOutlined /></template>新建文件</a-button>
        <a-button :loading="uploading" @click="handleUpload"><template #icon><UploadOutlined /></template>上传</a-button>
      </div>
    </div>

    <div class="breadcrumb-bar">
      <span class="breadcrumb-item breadcrumb-root" @click="goHome">/</span>
      <template v-for="(seg, idx) in breadcrumbs" :key="idx">
        <span class="breadcrumb-sep">/</span>
        <span class="breadcrumb-item" @click="navigateTo(breadcrumbPath(idx))">{{ seg }}</span>
      </template>
    </div>

    <a-spin :spinning="loading">
      <div class="file-table-wrapper">
        <table class="file-table">
          <thead>
            <tr>
              <th class="col-name">名称</th>
              <th class="col-size">大小</th>
              <th class="col-time">修改时间</th>
              <th class="col-action">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredFiles.length === 0 && !loading">
              <td colspan="4" class="empty-cell">暂无文件</td>
            </tr>
            <tr
              v-for="file in filteredFiles"
              :key="file.path"
              class="file-row"
              @dblclick="handleOpen(file)"
            >
              <td class="col-name">
                <span class="file-name-cell">
                  <component :is="getFileIcon(file)" class="file-icon" :class="{ 'file-icon-dir': file.is_dir }" />
                  <span>{{ file.name }}</span>
                </span>
              </td>
              <td class="col-size">{{ file.is_dir ? '-' : formatSize(file.size) }}</td>
              <td class="col-time">{{ formatTime(file.mod_time) }}</td>
              <td class="col-action">
                <div class="action-btns">
                  <a-button v-if="!file.is_dir" type="link" size="small" @click="openEditor(file.path)">
                    <template #icon><EditOutlined /></template>编辑
                  </a-button>
                  <a-button type="link" size="small" @click="handleRename(file)">
                    <template #icon><EditOutlined /></template>重命名
                  </a-button>
                  <a-button type="link" size="small" danger @click="handleDelete(file)">
                    <template #icon><DeleteOutlined /></template>删除
                  </a-button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </a-spin>

    <a-modal v-model:open="settingVisible" title="根目录设置" :confirm-loading="settingLoading" ok-text="保存" cancel-text="取消" @ok="saveSettings">
      <a-spin :spinning="settingLoading">
        <div class="config-form">
          <div class="config-field">
            <label class="config-label">文件管理根路径</label>
            <a-input v-model:value="basePathValue" placeholder="例如: /home/user/data" />
            <div class="config-help">设置文件管理的根目录，所有文件操作将限制在此目录内。</div>
          </div>
        </div>
      </a-spin>
    </a-modal>

    <a-modal v-model:open="mkdirVisible" title="新建目录" :confirm-loading="mkdirLoading" ok-text="创建" cancel-text="取消" @ok="confirmMkdir">
      <div class="config-form">
        <div class="config-field">
          <label class="config-label">目录名称</label>
          <a-input v-model:value="mkdirName" placeholder="请输入目录名称" @press-enter="confirmMkdir" />
        </div>
      </div>
    </a-modal>

    <a-modal v-model:open="touchVisible" title="新建文件" :confirm-loading="touchLoading" ok-text="创建" cancel-text="取消" @ok="confirmTouch">
      <div class="config-form">
        <div class="config-field">
          <label class="config-label">文件名称</label>
          <a-input v-model:value="touchName" placeholder="请输入文件名称" @press-enter="confirmTouch" />
        </div>
      </div>
    </a-modal>

    <a-modal v-model:open="renameVisible" title="重命名" :confirm-loading="renameLoading" ok-text="确认" cancel-text="取消" @ok="confirmRename">
      <div class="config-form">
        <div class="config-field">
          <label class="config-label">当前名称: {{ renameOldName }}</label>
          <a-input v-model:value="renameNewName" placeholder="请输入新名称" @press-enter="confirmRename" />
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="editorVisible"
      :title="`编辑文件 - ${editorPath}`"
      width="800px"
      :footer="null"
      :destroy-on-close="true"
    >
      <a-spin :spinning="editorLoading">
        <div class="editor-wrapper">
          <a-textarea
            v-model:value="editorContent"
            :rows="20"
            class="editor-textarea"
            style="font-family: monospace"
          />
          <div class="editor-actions">
            <a-button type="primary" :loading="editorSaving" @click="handleSaveFile">保存</a-button>
            <a-button @click="editorVisible = false">关闭</a-button>
          </div>
        </div>
      </a-spin>
    </a-modal>
  </div>
</template>

<style scoped>
.file-page { min-width: 0 }
.file-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 16px; margin-bottom: 20px }
.file-eyebrow { display: block; margin-bottom: 4px; color: #1677ff; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase }
.file-title { margin: 0; color: rgba(0,0,0,.88); font-size: 24px; font-weight: 700; line-height: 1.25 }
.file-heading-actions { display: flex; align-items: center; gap: 12px }
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 20px }
.stat-card { padding: 16px 20px; background: #fff; border: 1px solid rgba(5,5,5,.06); border-radius: 10px; text-align: center }
.stat-value { color: rgba(0,0,0,.88); font-size: 28px; font-weight: 700 }
.stat-dir .stat-value { color: #1677ff }
.stat-file .stat-value { color: #52c41a }
.stat-label { margin-top: 4px; color: rgba(0,0,0,.45); font-size: 13px }
.file-toolbar { display: flex; align-items: center; gap: 16px; margin-bottom: 12px }
.search-input { max-width: 320px }
.toolbar-actions { display: flex; align-items: center; gap: 8px; margin-left: auto }
.breadcrumb-bar { display: flex; align-items: center; gap: 0; margin-bottom: 12px; padding: 8px 12px; background: rgba(0,0,0,.02); border: 1px solid rgba(5,5,5,.06); border-radius: 8px; font-size: 13px; overflow-x: auto }
.breadcrumb-item { color: #1677ff; cursor: pointer; white-space: nowrap; padding: 2px 4px; border-radius: 4px; transition: background .15s }
.breadcrumb-item:hover { background: rgba(22,119,255,.08) }
.breadcrumb-root { font-weight: 600 }
.breadcrumb-sep { color: rgba(0,0,0,.25); margin: 0 2px }
.file-table-wrapper { border: 1px solid rgba(5,5,5,.06); border-radius: 10px; overflow: hidden }
.file-table { width: 100%; border-collapse: collapse }
.file-table th { padding: 10px 16px; color: rgba(0,0,0,.45); font-size: 13px; font-weight: 500; text-align: left; background: rgba(0,0,0,.02); border-bottom: 1px solid rgba(5,5,5,.06) }
.file-table td { padding: 10px 16px; border-bottom: 1px solid rgba(5,5,5,.04); font-size: 13px }
.file-row { cursor: pointer; transition: background .15s }
.file-row:hover { background: rgba(22,119,255,.04) }
.col-name { min-width: 200px }
.col-size { width: 100px; color: rgba(0,0,0,.45) }
.col-time { width: 180px; color: rgba(0,0,0,.45) }
.col-action { width: 200px }
.file-name-cell { display: inline-flex; align-items: center; gap: 8px; color: rgba(0,0,0,.88); font-weight: 500 }
.file-icon { font-size: 16px; color: rgba(0,0,0,.35) }
.file-icon-dir { color: #1677ff }
.action-btns { display: inline-flex; align-items: center; gap: 0 }
.empty-cell { padding: 40px 16px; color: rgba(0,0,0,.25); text-align: center; font-size: 14px }
.config-form { display: flex; flex-direction: column; gap: 16px; padding: 8px 0 }
.config-field { display: flex; flex-direction: column; gap: 6px }
.config-label { color: rgba(0,0,0,.88); font-size: 13px; font-weight: 500 }
.config-help { color: rgba(0,0,0,.35); font-size: 12px }
.editor-wrapper { display: flex; flex-direction: column; gap: 12px }
.editor-textarea { font-size: 13px }
.editor-actions { display: flex; gap: 8px; justify-content: flex-end }

@media (max-width: 768px) {
  .file-heading { flex-direction: column; align-items: flex-start; gap: 8px }
  .file-heading-actions { flex-wrap: wrap }
  .stats-grid { grid-template-columns: repeat(3, 1fr) }
  .file-toolbar { flex-direction: column; align-items: stretch }
  .search-input { max-width: none }
  .toolbar-actions { flex-wrap: wrap; margin-left: 0 }
}
</style>
