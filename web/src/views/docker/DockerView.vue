<script setup lang="ts">
import type { ComposeProject, ContainerInfo, DockerConnection, DockerPingResult, ImageInfo, NetworkInfo, RegistryConfig, VolumeInfo } from '@/api/docker'
import {
  checkComposeAvailable,
  composeDown,
  composeRestart,
  composeStart,
  composeStop,
  composeUp,
  createNetwork,
  createVolume,
  getComposeConfig,
  getDockerConnection,
  listImages,
  listComposeProjects,
  listContainers,
  listNetworks,
  listRegistries,
  listVolumes,
  pingDocker,
  pruneImages,
  pruneVolumes,
  removeContainer,
  removeImage,
  removeNetwork,
  removeVolume,
  restartContainer,
  saveDockerConnection,
  saveRegistries,
  startContainer,
  stopContainer,
  tagImage,
  testRegistry,
  testDockerConnection,
  pullImage,
} from '@/api/docker'
import {
  AppstoreOutlined,
  CaretRightOutlined,
  CloudOutlined,
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  FormOutlined,
  LinkOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
} from '@antdv-next/icons'
import { Modal, message } from 'antdv-next'
import { computed, h, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'

const loading = shallowRef(false)
const dockerAvailable = shallowRef(true)
const activeTab = ref('containers')
const refreshing = shallowRef(false)

const containers = ref<ContainerInfo[]>([])
const networks = ref<NetworkInfo[]>([])
const composeProjects = ref<ComposeProject[]>([])
const images = ref<ImageInfo[]>([])
const registries = ref<RegistryConfig[]>([])
const volumes = ref<VolumeInfo[]>([])
const composeAvailable = shallowRef(false)
const searchText = ref('')
const showAll = ref(true)

const configVisible = shallowRef(false)
const configLoading = shallowRef(false)
const testLoading = shallowRef(false)
const configForm = ref<DockerConnection>({ host: '', tls_verify: false, cert_path: '' })

const createNetworkVisible = shallowRef(false)
const createNetworkLoading = shallowRef(false)
const newNetwork = ref({ name: '', driver: 'bridge', subnet: '', gateway: '', internal: false, attachable: false })

const pullImageVisible = shallowRef(false)
const pullImageLoading = shallowRef(false)
const pullImageForm = ref({ image: '', registry: '' })
const tagImageVisible = shallowRef(false)
const tagImageLoading = shallowRef(false)
const tagImageForm = ref({ source: '', target: '' })
const registryVisible = shallowRef(false)
const registryLoading = shallowRef(false)
const registryForm = ref<RegistryConfig[]>([])
const createVolumeVisible = shallowRef(false)
const createVolumeLoading = shallowRef(false)
const newVolume = ref({ name: '', driver: 'local', labels: '', options: '' })

const composeUpVisible = shallowRef(false)
const composeUpLoading = shallowRef(false)
const composeContent = ref('')
const composeProjectName = ref('')

const composeLines = computed(() => {
  const count = (composeContent.value || '').split('\n').length
  return Array.from({ length: Math.max(count, 12) })
})
const composeEditorRef = ref<HTMLElement | null>(null)

function onComposeEditorScroll(e: Event) {
  const textarea = e.target as HTMLTextAreaElement
  const gutter = composeEditorRef.value?.querySelector('.compose-editor-gutter')
  if (gutter) gutter.scrollTop = textarea.scrollTop
}

interface ComposePreset {
  key: string
  name: string
  desc: string
  content: string
}

const composePresets: ComposePreset[] = [
  {
    key: 'nginx',
    name: 'Nginx',
    desc: 'Web 服务器 / 反向代理',
    content: `services:
  nginx:
    image: nginx:latest
    container_name: nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./nginx/html:/usr/share/nginx/html
      - ./nginx/logs:/var/log/nginx
`,
  },
  {
    key: 'mysql',
    name: 'MySQL',
    desc: '关系型数据库',
    content: `services:
  mysql:
    image: mysql:8.0
    container_name: mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root123456
      MYSQL_DATABASE: app
      MYSQL_USER: app
      MYSQL_PASSWORD: app123456
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password

volumes:
  mysql_data:
`,
  },
  {
    key: 'redis',
    name: 'Redis',
    desc: '内存键值数据库',
    content: `services:
  redis:
    image: redis:7-alpine
    container_name: redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes --requirepass redis123456

volumes:
  redis_data:
`,
  },
  {
    key: 'postgres',
    name: 'PostgreSQL',
    desc: '关系型数据库',
    content: `services:
  postgres:
    image: postgres:16-alpine
    container_name: postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres123456
      POSTGRES_DB: app
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
`,
  },
  {
    key: 'mongo',
    name: 'MongoDB',
    desc: 'NoSQL 文档数据库',
    content: `services:
  mongo:
    image: mongo:7
    container_name: mongo
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: mongo123456
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:
`,
  },
  {
    key: 'portainer',
    name: 'Portainer',
    desc: 'Docker 可视化管理面板',
    content: `services:
  portainer:
    image: portainer/portainer-ce:latest
    container_name: portainer
    restart: unless-stopped
    ports:
      - "9443:9443"
      - "9000:9000"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - portainer_data:/data

volumes:
  portainer_data:
`,
  },
  {
    key: 'wordpress',
    name: 'WordPress',
    desc: '博客 / CMS 系统',
    content: `services:
  wordpress:
    image: wordpress:latest
    container_name: wordpress
    restart: unless-stopped
    ports:
      - "8080:80"
    environment:
      WORDPRESS_DB_HOST: wordpress-db:3306
      WORDPRESS_DB_USER: wordpress
      WORDPRESS_DB_PASSWORD: wp123456
      WORDPRESS_DB_NAME: wordpress
    volumes:
      - wp_html:/var/www/html
    depends_on:
      - wordpress-db

  wordpress-db:
    image: mysql:8.0
    container_name: wordpress-db
    restart: unless-stopped
    environment:
      MYSQL_DATABASE: wordpress
      MYSQL_USER: wordpress
      MYSQL_PASSWORD: wp123456
      MYSQL_ROOT_PASSWORD: root123456
    volumes:
      - wp_db_data:/var/lib/mysql

volumes:
  wp_html:
  wp_db_data:
`,
  },
  {
    key: 'lnmp',
    name: 'LNMP',
    desc: 'Nginx + MySQL + PHP',
    content: `services:
  nginx:
    image: nginx:alpine
    container_name: lnmp-nginx
    restart: unless-stopped
    ports:
      - "80:80"
    volumes:
      - ./www:/usr/share/nginx/html
      - ./nginx/conf.d:/etc/nginx/conf.d
    depends_on:
      - php

  php:
    image: php:8.2-fpm-alpine
    container_name: lnmp-php
    restart: unless-stopped
    volumes:
      - ./www:/usr/share/nginx/html

  mysql:
    image: mysql:8.0
    container_name: lnmp-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root123456
      MYSQL_DATABASE: app
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
`,
  },
]

const presetVisible = shallowRef(false)

function applyPreset(preset: ComposePreset) {
  if (composeContent.value && !composeContent.value.startsWith('#')) {
    composeContent.value = preset.content + composeContent.value
  } else {
    composeContent.value = preset.content
  }
  presetVisible.value = false
  message.success(`已载入 ${preset.name} 模板`)
}

interface VisualService {
  name: string
  image: string
  container_name: string
  restart: string
  ports: string
  environment: string
  volumes: string
  command: string
  depends_on: string
}

const visualVisible = shallowRef(false)
const visualServices = ref<VisualService[]>([])

function openVisualConfig() {
  if (visualServices.value.length === 0) {
    addVisualService()
  }
  visualVisible.value = true
}

function addVisualService() {
  visualServices.value.push({
    name: '',
    image: '',
    container_name: '',
    restart: 'no',
    ports: '',
    environment: '',
    volumes: '',
    command: '',
    depends_on: '',
  })
}

function removeVisualService(index: number) {
  visualServices.value.splice(index, 1)
}

function generateComposeFromVisual(): string {
  const lines = ['services:']
  const volumeNames: string[] = []

  for (const svc of visualServices.value) {
    const name = svc.name.trim()
    if (!name) continue

    lines.push(`  ${name}:`)
    if (svc.image.trim()) lines.push(`    image: ${svc.image.trim()}`)
    if (svc.container_name.trim()) lines.push(`    container_name: ${svc.container_name.trim()}`)
    if (svc.restart && svc.restart !== 'no') lines.push(`    restart: ${svc.restart}`)

    if (svc.ports.trim()) {
      lines.push('    ports:')
      for (const p of svc.ports.split('\n')) {
        const v = p.trim()
        if (v) lines.push(`      - "${v}"`)
      }
    }

    if (svc.environment.trim()) {
      lines.push('    environment:')
      for (const e of svc.environment.split('\n')) {
        const v = e.trim()
        if (v) lines.push(`      ${v.includes('=') ? v : v + '='}`)
      }
    }

    if (svc.volumes.trim()) {
      lines.push('    volumes:')
      for (const v of svc.volumes.split('\n')) {
        const val = v.trim()
        if (!val) continue
        lines.push(`      - ${val}`)
        if (!val.startsWith('./') && !val.startsWith('/') && !val.includes(':')) {
          volumeNames.push(val)
        }
      }
    }

    if (svc.command.trim()) lines.push(`    command: ${svc.command.trim()}`)

    if (svc.depends_on.trim()) {
      lines.push('    depends_on:')
      for (const d of svc.depends_on.split(',')) {
        const v = d.trim()
        if (v) lines.push(`      - ${v}`)
      }
    }
  }

  if (volumeNames.length > 0) {
    lines.push('')
    lines.push('volumes:')
    for (const vn of new Set(volumeNames)) {
      lines.push(`  ${vn}:`)
    }
  }

  return lines.join('\n') + '\n'
}

function applyVisualConfig() {
  const hasValid = visualServices.value.some((s) => s.name.trim())
  if (!hasValid) {
    message.warning('请至少配置一个服务')
    return
  }
  composeContent.value = generateComposeFromVisual()
  visualVisible.value = false
  message.success('已生成 Compose 配置')
}

const composeEditLoading = shallowRef(false)

async function editComposeProject(name: string) {
  composeEditLoading.value = true
  composeProjectName.value = name
  try {
    const res = await getComposeConfig(name)
    composeContent.value = res.data?.content ?? ''
    composeUpVisible.value = true
  } catch {
    message.error('获取配置失败，可能部署时未保留配置文件')
  } finally {
    composeEditLoading.value = false
  }
}

async function visualEditComposeProject(name: string) {
  composeEditLoading.value = true
  try {
    const res = await getComposeConfig(name)
    const content = res.data?.content ?? ''
    parseYamlToVisualServices(content)
    visualVisible.value = true
  } catch {
    message.error('获取配置失败，可能部署时未保留配置文件')
  } finally {
    composeEditLoading.value = false
  }
}

function parseYamlToVisualServices(yaml: string) {
  const services: VisualService[] = []
  const lines = yaml.split('\n')
  let inServices = false

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i] ?? ''

    if (line.trim().startsWith('services:')) {
      inServices = true
      continue
    }

    if (!inServices) continue

    const svcMatch = line.match(/^(\s{2})(\S+):\s*$/)
    if (svcMatch) {
      const indent = svcMatch[1]!.length
      if (indent === 2) {
        const t = line.trim()
        if (t.startsWith('volumes:') || t.startsWith('networks:') || t.startsWith('configs:') || t.startsWith('secrets:')) {
          inServices = false
          continue
        }
        services.push({
          name: svcMatch[2] ?? '',
          image: '',
          container_name: '',
          restart: 'no',
          ports: '',
          environment: '',
          volumes: '',
          command: '',
          depends_on: '',
        })
        continue
      }
    }

    if (services.length === 0) continue
    const svc = services[services.length - 1]!
    const trimmed = line.trim()

    if (trimmed.startsWith('image:')) {
      svc.image = trimmed.replace(/^image:\s*/, '')
    } else if (trimmed.startsWith('container_name:')) {
      svc.container_name = trimmed.replace(/^container_name:\s*/, '')
    } else if (trimmed.startsWith('restart:')) {
      svc.restart = trimmed.replace(/^restart:\s*/, '')
    } else if (trimmed.startsWith('command:')) {
      svc.command = trimmed.replace(/^command:\s*/, '')
    } else if (trimmed.startsWith('ports:')) {
      const ports: string[] = []
      for (let j = i + 1; j < lines.length; j++) {
        const pLine = lines[j]
        if (!pLine) break
        if (/^\s{4,6}- /.test(pLine)) {
          ports.push(pLine.trim().replace(/^- ["']?/, '').replace(/["']$/, ''))
        } else {
          break
        }
      }
      svc.ports = ports.join('\n')
    } else if (trimmed.startsWith('environment:')) {
      const envs: string[] = []
      for (let j = i + 1; j < lines.length; j++) {
        const eLine = lines[j]
        if (!eLine || !/^\s{4,6}[A-Z_]/.test(eLine)) break
        envs.push(eLine.trim())
      }
      svc.environment = envs.join('\n')
    } else if (trimmed.startsWith('volumes:')) {
      const vols: string[] = []
      for (let j = i + 1; j < lines.length; j++) {
        const vLine = lines[j]
        if (!vLine || !/^\s{4,6}- /.test(vLine)) break
        vols.push(vLine.trim().replace(/^- /, '').replace(/^["']|["']$/g, ''))
      }
      svc.volumes = vols.join('\n')
    } else if (trimmed.startsWith('depends_on:')) {
      const deps: string[] = []
      for (let j = i + 1; j < lines.length; j++) {
        const dLine = lines[j]
        if (!dLine || !/^\s{4,6}- /.test(dLine)) break
        deps.push(dLine.trim().replace(/^- /, ''))
      }
      svc.depends_on = deps.join(', ')
    }
  }

  visualServices.value = services.length > 0 ? services : [{
    name: '', image: '', container_name: '', restart: 'no', ports: '', environment: '', volumes: '', command: '', depends_on: '',
  }]
}

let refreshTimer: ReturnType<typeof setInterval> | null = null

const filteredContainers = computed(() => {
  if (!searchText.value) return containers.value
  const keyword = searchText.value.toLowerCase()
  return containers.value.filter(
    (c) =>
      c.names.some((n) => n.toLowerCase().includes(keyword))
      || c.image.toLowerCase().includes(keyword)
      || c.id.toLowerCase().includes(keyword),
  )
})

const runningCount = computed(() => containers.value.filter((c) => c.state === 'running').length)
const stoppedCount = computed(() => containers.value.filter((c) => c.state !== 'running').length)
const imageCount = computed(() => images.value.length)
const volumeCount = computed(() => volumes.value.length)
const registryOptions = computed(() => registries.value.map((r) => ({ label: r.name, value: r.name })))

function getDisplayName(names?: string[]) {
  if (!names || names.length === 0) return '-'
  return names[0]?.startsWith('/') === true ? names[0].slice(1) : (names[0] ?? '-')
}

function getStateColor(state: string) {
  switch (state) {
    case 'running': return '#52c41a'
    case 'paused': return '#faad14'
    case 'exited':
    case 'dead': return '#ff4d4f'
    default: return '#d9d9d9'
  }
}

function getStateLabel(state: string) {
  switch (state) {
    case 'running': return '运行中'
    case 'paused': return '已暂停'
    case 'exited': return '已停止'
    case 'dead': return '已死亡'
    case 'created': return '已创建'
    case 'restarting': return '重启中'
    default: return state
  }
}

function getComposeStatusLabel(status: string) {
  switch (status) {
    case 'running': return '全部运行'
    case 'stopped': return '全部停止'
    case 'partial': return '部分运行'
    default: return status
  }
}

function getComposeStatusColor(status: string) {
  switch (status) {
    case 'running': return '#52c41a'
    case 'stopped': return '#ff4d4f'
    case 'partial': return '#faad14'
    default: return '#d9d9d9'
  }
}

function formatPorts(ports: ContainerInfo['ports']) {
  if (!ports || ports.length === 0) return '-'
  return ports.filter((p) => p.public_port > 0).map((p) => `${p.ip || '0.0.0.0'}:${p.public_port}->${p.private_port}/${p.type}`).join(', ') || '-'
}

function formatTimeFromSeconds(seconds?: number) {
  if (!seconds) return '-'
  return new Date(seconds * 1000).toLocaleString()
}

function parseKeyValueLines(text: string) {
  const result: Record<string, string> = {}
  for (const line of text.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed) continue
    const idx = trimmed.indexOf('=')
    if (idx <= 0) continue
    result[trimmed.slice(0, idx).trim()] = trimmed.slice(idx + 1).trim()
  }
  return result
}

async function fetchContainers() {
  try {
    const res = await listContainers(showAll.value)
    containers.value = res.data ?? []
  } catch { containers.value = [] }
}

async function fetchNetworks() {
  try {
    const res = await listNetworks()
    networks.value = res.data ?? []
  } catch { networks.value = [] }
}

async function fetchComposeProjects() {
  try {
    const res = await listComposeProjects()
    composeProjects.value = res.data ?? []
  } catch { composeProjects.value = [] }
}

async function fetchImages() {
  try {
    const res = await listImages()
    images.value = res.data ?? []
  } catch { images.value = [] }
}

async function fetchRegistries() {
  try {
    const res = await listRegistries()
    registries.value = res.data ?? []
  } catch { registries.value = [] }
}

async function fetchVolumes() {
  try {
    const res = await listVolumes()
    volumes.value = res.data ?? []
  } catch { volumes.value = [] }
}

async function fetchDockerStatus() {
  try {
    const res = await pingDocker()
    const data = res.data as DockerPingResult
    dockerAvailable.value = data.available
  } catch { dockerAvailable.value = false }
}

async function loadData() {
  loading.value = true
  try {
    await fetchDockerStatus()
    if (dockerAvailable.value) {
      const [, , , , composeRes] = await Promise.allSettled([
        fetchContainers(),
        fetchImages(),
        fetchVolumes(),
        fetchRegistries(),
        checkComposeAvailable(),
      ])
      if (composeRes.status === 'fulfilled') {
        composeAvailable.value = composeRes.value.data?.available ?? false
      }
    }
  } finally { loading.value = false }
}

async function refreshData() {
  if (refreshing.value) return
  refreshing.value = true
  try {
    if (activeTab.value === 'containers') await fetchContainers()
    else if (activeTab.value === 'networks') await fetchNetworks()
    else if (activeTab.value === 'compose') await fetchComposeProjects()
    else if (activeTab.value === 'images') await fetchImages()
    else if (activeTab.value === 'registries') await fetchRegistries()
    else if (activeTab.value === 'volumes') await fetchVolumes()
  } finally { refreshing.value = false }
}

async function handleStart(id: string) {
  try { await startContainer(id); message.success('容器启动成功'); await fetchContainers() }
  catch { message.error('容器启动失败') }
}
async function handleStop(id: string) {
  try { await stopContainer(id); message.success('容器停止成功'); await fetchContainers() }
  catch { message.error('容器停止失败') }
}
async function handleRestart(id: string) {
  try { await restartContainer(id); message.success('容器重启成功'); await fetchContainers() }
  catch { message.error('容器重启失败') }
}
function handleRemove(id: string, name: string) {
  Modal.confirm({
    title: '确认删除容器', content: `确定要删除容器 "${name}" 吗？`, icon: () => h(ExclamationCircleOutlined), okType: 'danger', okText: '删除', cancelText: '取消',
    async onOk() {
      try { await removeContainer(id, true); message.success('容器删除成功'); await fetchContainers() }
      catch { message.error('容器删除失败') }
    },
  })
}

async function handleRemoveNetwork(id: string, name: string) {
  Modal.confirm({
    title: '确认删除网络', content: `确定要删除网络 "${name}" 吗？`, icon: () => h(ExclamationCircleOutlined), okType: 'danger', okText: '删除', cancelText: '取消',
    async onOk() {
      try { await removeNetwork(id); message.success('网络已删除'); await fetchNetworks() }
      catch { message.error('删除网络失败') }
    },
  })
}

async function handleCreateNetwork() {
  if (!newNetwork.value.name) { message.warning('请输入网络名称'); return }
  createNetworkLoading.value = true
  try {
    await createNetwork(newNetwork.value)
    message.success('网络创建成功')
    createNetworkVisible.value = false
    newNetwork.value = { name: '', driver: 'bridge', subnet: '', gateway: '', internal: false, attachable: false }
    await fetchNetworks()
  } catch { message.error('创建网络失败') }
  finally { createNetworkLoading.value = false }
}

async function handlePullImage() {
  if (!pullImageForm.value.image) { message.warning('请输入镜像名称'); return }
  pullImageLoading.value = true
  try {
    await pullImage(pullImageForm.value.image, pullImageForm.value.registry || undefined)
    message.success('镜像拉取成功')
    pullImageVisible.value = false
    pullImageForm.value = { image: '', registry: '' }
    await fetchImages()
  } catch { message.error('镜像拉取失败') }
  finally { pullImageLoading.value = false }
}

function handleRemoveImage(id: string, name: string) {
  Modal.confirm({
    title: '确认删除镜像', content: `确定要删除镜像 "${name}" 吗？`, icon: () => h(ExclamationCircleOutlined), okType: 'danger', okText: '删除', cancelText: '取消',
    async onOk() {
      try { await removeImage(id, true); message.success('镜像已删除'); await fetchImages() }
      catch { message.error('删除镜像失败') }
    },
  })
}

function openTagImage(image: ImageInfo) {
  tagImageForm.value = { source: image.repo_tags?.[0] ?? image.id, target: '' }
  tagImageVisible.value = true
}

async function handleTagImage() {
  if (!tagImageForm.value.source || !tagImageForm.value.target) { message.warning('请输入源镜像和目标标签'); return }
  tagImageLoading.value = true
  try {
    await tagImage(tagImageForm.value.source, tagImageForm.value.target)
    message.success('镜像标签已创建')
    tagImageVisible.value = false
    await fetchImages()
  } catch { message.error('创建标签失败') }
  finally { tagImageLoading.value = false }
}

function handlePruneImages() {
  Modal.confirm({
    title: '确认清理未使用镜像', content: '将删除未被容器使用的悬空镜像和可清理镜像，是否继续？', icon: () => h(ExclamationCircleOutlined), okText: '清理', cancelText: '取消',
    async onOk() {
      try {
        const res = await pruneImages()
        message.success(`已清理 ${res.data?.deleted ?? 0} 个镜像，释放 ${res.data?.space_text ?? '-'}`)
        await fetchImages()
      } catch { message.error('清理镜像失败') }
    },
  })
}

function openRegistryConfig() {
  registryForm.value = registries.value.map((item) => ({ ...item }))
  registryVisible.value = true
}

function addRegistry() {
  registryForm.value.push({ name: '', server_address: '', username: '', password: '' })
}

function removeRegistry(index: number) {
  registryForm.value.splice(index, 1)
}

async function handleSaveRegistries() {
  registryLoading.value = true
  try {
    await saveRegistries(registryForm.value.filter((r) => r.name && r.server_address))
    message.success('仓库配置已保存')
    registryVisible.value = false
    await fetchRegistries()
  } catch { message.error('保存仓库配置失败') }
  finally { registryLoading.value = false }
}

async function handleTestRegistry(registry: RegistryConfig) {
  if (!registry.server_address) { message.warning('仓库地址不能为空'); return }
  try { await testRegistry(registry); message.success('仓库登录成功') }
  catch { message.error('仓库登录失败') }
}

async function handleCreateVolume() {
  if (!newVolume.value.name) { message.warning('请输入存储卷名称'); return }
  createVolumeLoading.value = true
  try {
    await createVolume({
      name: newVolume.value.name,
      driver: newVolume.value.driver || 'local',
      labels: parseKeyValueLines(newVolume.value.labels),
      options: parseKeyValueLines(newVolume.value.options),
    })
    message.success('存储卷已创建')
    createVolumeVisible.value = false
    newVolume.value = { name: '', driver: 'local', labels: '', options: '' }
    await fetchVolumes()
  } catch { message.error('创建存储卷失败') }
  finally { createVolumeLoading.value = false }
}

function handleRemoveVolume(name: string) {
  Modal.confirm({
    title: '确认删除存储卷', content: `确定要删除存储卷 "${name}" 吗？被使用中的卷无法删除。`, icon: () => h(ExclamationCircleOutlined), okType: 'danger', okText: '删除', cancelText: '取消',
    async onOk() {
      try { await removeVolume(name); message.success('存储卷已删除'); await fetchVolumes() }
      catch { message.error('删除存储卷失败') }
    },
  })
}

function handlePruneVolumes() {
  Modal.confirm({
    title: '确认清理未使用存储卷', content: '将删除未被任何容器使用的存储卷，数据不可恢复，是否继续？', icon: () => h(ExclamationCircleOutlined), okType: 'danger', okText: '清理', cancelText: '取消',
    async onOk() {
      try {
        const res = await pruneVolumes()
        message.success(`已清理 ${res.data?.deleted ?? 0} 个存储卷，释放 ${res.data?.space_text ?? '-'}`)
        await fetchVolumes()
      } catch { message.error('清理存储卷失败') }
    },
  })
}

function handleComposeDown(name: string) {
  Modal.confirm({
    title: '确认移除 Compose 项目', content: `确定要停止并移除 "${name}" 吗？`, icon: () => h(ExclamationCircleOutlined), okType: 'danger', okText: '移除', cancelText: '取消',
    async onOk() {
      try { await composeDown(name); message.success('Compose 项目已移除'); await fetchComposeProjects() }
      catch { message.error('移除失败') }
    },
  })
}

async function handleComposeUp() {
  if (!composeContent.value) { message.warning('请输入 Compose 内容'); return }
  composeUpLoading.value = true
  try {
    await composeUp(composeContent.value, composeProjectName.value || undefined)
    message.success('Compose 部署成功')
    composeUpVisible.value = false
    composeContent.value = ''
    composeProjectName.value = ''
    await fetchComposeProjects()
  } catch { message.error('Compose 部署失败') }
  finally { composeUpLoading.value = false }
}

async function openConfig() {
  configVisible.value = true
  configLoading.value = true
  try {
    const res = await getDockerConnection()
    const conn = res.data
    configForm.value = { host: conn?.host ?? '', tls_verify: conn?.tls_verify ?? false, cert_path: conn?.cert_path ?? '' }
  } catch { configForm.value = { host: '', tls_verify: false, cert_path: '' } }
  finally { configLoading.value = false }
}

async function handleTestConnection() {
  testLoading.value = true
  try { await testDockerConnection(configForm.value); message.success('连接测试成功') }
  catch (err: unknown) {
    const serverMsg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
    message.error(serverMsg ?? '连接测试失败')
  }
  finally { testLoading.value = false }
}

async function handleSaveConnection() {
  configLoading.value = true
  try { await saveDockerConnection(configForm.value); message.success('配置已保存'); configVisible.value = false; await loadData() }
  catch { message.error('保存配置失败') }
  finally { configLoading.value = false }
}

function startRefreshTimer() { stopRefreshTimer(); refreshTimer = setInterval(refreshData, 10000) }
function stopRefreshTimer() { if (refreshTimer) { clearInterval(refreshTimer); refreshTimer = null } }

const containerColumns = [
  { title: '容器名称', dataIndex: 'names', key: 'name', width: 180 },
  { title: '镜像', dataIndex: 'image', key: 'image', width: 200 },
  { title: '状态', dataIndex: 'state', key: 'state', width: 100 },
  { title: '详情', dataIndex: 'status', key: 'status', width: 140 },
  { title: '端口映射', dataIndex: 'ports', key: 'ports', width: 220 },
  { title: '容器ID', dataIndex: 'id', key: 'id', width: 120 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
]

const networkColumns = [
  { title: '网络名称', dataIndex: 'name', key: 'name', width: 180 },
  { title: '驱动', dataIndex: 'driver', key: 'driver', width: 100 },
  { title: '范围', dataIndex: 'scope', key: 'scope', width: 80 },
  { title: '子网', dataIndex: 'subnets', key: 'subnets', width: 220 },
  { title: '连接容器', dataIndex: 'containers', key: 'containers', width: 180 },
  { title: 'ID', dataIndex: 'id', key: 'id', width: 120 },
  { title: '操作', key: 'action', width: 80, fixed: 'right' as const },
]

const composeColumns = [
  { title: '项目名称', dataIndex: 'name', key: 'name', width: 200 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 120 },
  { title: '服务', dataIndex: 'services', key: 'services', width: 280 },
  { title: '运行/停止', key: 'count', width: 120 },
  { title: '操作', key: 'action', width: 360, fixed: 'right' as const },
]

const imageColumns = [
  { title: '镜像', dataIndex: 'repo_tags', key: 'name', width: 260 },
  { title: '镜像ID', dataIndex: 'short_id', key: 'id', width: 140 },
  { title: '大小', dataIndex: 'size_text', key: 'size', width: 100 },
  { title: '容器引用', dataIndex: 'containers', key: 'containers', width: 100 },
  { title: '创建时间', dataIndex: 'created', key: 'created', width: 180 },
  { title: '操作', key: 'action', width: 160, fixed: 'right' as const },
]

const registryColumns = [
  { title: '名称', dataIndex: 'name', key: 'name', width: 160 },
  { title: '仓库地址', dataIndex: 'server_address', key: 'server_address', width: 260 },
  { title: '用户名', dataIndex: 'username', key: 'username', width: 160 },
]

const volumeColumns = [
  { title: '名称', dataIndex: 'name', key: 'name', width: 200 },
  { title: '驱动', dataIndex: 'driver', key: 'driver', width: 100 },
  { title: '挂载点', dataIndex: 'mountpoint', key: 'mountpoint', width: 360 },
  { title: '大小', dataIndex: 'size_text', key: 'size', width: 100 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 100, fixed: 'right' as const },
]

watch(showAll, () => fetchContainers())
watch(activeTab, () => refreshData())

onMounted(() => { loadData(); startRefreshTimer() })
onBeforeUnmount(() => stopRefreshTimer())
</script>

<template>
  <div class="docker-page">
    <a-spin :spinning="loading">
      <div class="docker-heading">
        <div>
          <span class="docker-eyebrow">Docker</span>
          <h2 class="docker-title">容器管理</h2>
        </div>
        <div class="docker-heading-actions">
          <div class="docker-status" :class="{ 'docker-status-offline': !dockerAvailable }">
            <span class="status-dot" />
            <span class="status-text">{{ dockerAvailable ? 'Docker 已连接' : 'Docker 不可用' }}</span>
          </div>
          <a-button @click="openConfig"><template #icon><SettingOutlined /></template>连接设置</a-button>
          <a-button @click="refreshData"><template #icon><ReloadOutlined :spin="refreshing" /></template>刷新</a-button>
        </div>
      </div>

      <template v-if="dockerAvailable">
        <div class="stats-grid stats-grid-extended">
          <div class="stat-card"><div class="stat-value">{{ containers.length }}</div><div class="stat-label">全部容器</div></div>
          <div class="stat-card stat-running"><div class="stat-value">{{ runningCount }}</div><div class="stat-label">运行中</div></div>
          <div class="stat-card stat-stopped"><div class="stat-value">{{ stoppedCount }}</div><div class="stat-label">已停止</div></div>
          <div class="stat-card"><div class="stat-value">{{ networks.length }}</div><div class="stat-label">网络</div></div>
          <div class="stat-card"><div class="stat-value">{{ imageCount }}</div><div class="stat-label">镜像</div></div>
          <div class="stat-card"><div class="stat-value">{{ volumeCount }}</div><div class="stat-label">存储卷</div></div>
        </div>

        <a-tabs v-model:activeKey="activeTab" class="docker-tabs">
          <a-tab-pane key="containers" tab="容器">
            <div class="docker-toolbar">
              <a-input-search v-model:value="searchText" placeholder="搜索容器名称、镜像..." class="search-input" allow-clear>
                <template #prefix><SearchOutlined /></template>
              </a-input-search>
              <label class="toggle-row"><span class="toggle-label">显示已停止</span><a-switch v-model:checked="showAll" size="small" /></label>
            </div>

            <a-table :columns="containerColumns" :data-source="filteredContainers" :pagination="{ pageSize: 20, showTotal: (t: number) => `共 ${t} 个容器` }" :scroll="{ x: 1060 }" row-key="id" size="middle">
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'"><span class="container-name">{{ getDisplayName(record.names) }}</span></template>
                <template v-else-if="column.key === 'image'"><span class="container-image">{{ record.image }}</span></template>
                <template v-else-if="column.key === 'state'">
                  <span class="state-badge" :style="{ color: getStateColor(record.state) }">
                    <span class="state-indicator" :style="{ background: getStateColor(record.state) }" />{{ getStateLabel(record.state) }}
                  </span>
                </template>
                <template v-else-if="column.key === 'ports'"><span class="port-text">{{ formatPorts(record.ports) }}</span></template>
                <template v-else-if="column.key === 'id'"><span class="container-id">{{ record.id }}</span></template>
                <template v-else-if="column.key === 'action'">
                  <div class="action-btns">
                    <a-button v-if="record.state !== 'running'" type="link" size="small" @click="handleStart(record.id)"><template #icon><CaretRightOutlined /></template>启动</a-button>
                    <a-button v-if="record.state === 'running'" type="link" size="small" @click="handleStop(record.id)">停止</a-button>
                    <a-button v-if="record.state === 'running'" type="link" size="small" @click="handleRestart(record.id)"><template #icon><ReloadOutlined /></template>重启</a-button>
                    <a-button type="link" size="small" danger @click="handleRemove(record.id, getDisplayName(record.names))"><template #icon><DeleteOutlined /></template>删除</a-button>
                  </div>
                </template>
              </template>
            </a-table>
          </a-tab-pane>

          <a-tab-pane key="networks" tab="网络">
            <div class="docker-toolbar">
              <a-button type="primary" @click="createNetworkVisible = true"><template #icon><PlusOutlined /></template>创建网络</a-button>
            </div>
            <a-table :columns="networkColumns" :data-source="networks" :pagination="{ pageSize: 20, showTotal: (t: number) => `共 ${t} 个网络` }" :scroll="{ x: 980 }" row-key="id" size="middle">
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'">
                  <span class="container-name">{{ record.name }}</span>
                  <span v-if="record.internal" class="badge badge-internal">内部</span>
                </template>
                <template v-else-if="column.key === 'subnets'"><span class="port-text">{{ record.subnets?.join(', ') || '-' }}</span></template>
                <template v-else-if="column.key === 'containers'"><span class="port-text">{{ record.containers?.length ?? 0 }} 个容器</span></template>
                <template v-else-if="column.key === 'id'"><span class="container-id">{{ record.id }}</span></template>
                <template v-else-if="column.key === 'action'">
                  <a-button v-if="record.name !== 'bridge' && record.name !== 'host' && record.name !== 'none'" type="link" size="small" danger @click="handleRemoveNetwork(record.id, record.name)"><template #icon><DeleteOutlined /></template>删除</a-button>
                  <span v-else class="port-text">系统保留</span>
                </template>
              </template>
            </a-table>
          </a-tab-pane>

          <a-tab-pane key="images" tab="镜像">
            <div class="docker-toolbar">
              <a-button type="primary" @click="pullImageVisible = true"><template #icon><PlusOutlined /></template>拉取镜像</a-button>
              <a-button @click="handlePruneImages"><template #icon><DeleteOutlined /></template>清理未使用镜像</a-button>
            </div>
            <a-table :columns="imageColumns" :data-source="images" :pagination="{ pageSize: 20, showTotal: (t: number) => `共 ${t} 个镜像` }" :scroll="{ x: 940 }" row-key="id" size="middle">
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'">
                  <div class="tag-list">
                    <span v-for="tag in record.repo_tags" :key="tag" class="container-id">{{ tag }}</span>
                    <span v-if="!record.repo_tags || record.repo_tags.length === 0" class="container-id">&lt;none&gt;</span>
                  </div>
                </template>
                <template v-else-if="column.key === 'id'"><span class="container-id">{{ record.short_id }}</span></template>
                <template v-else-if="column.key === 'created'"><span class="port-text">{{ formatTimeFromSeconds(record.created) }}</span></template>
                <template v-else-if="column.key === 'action'">
                  <div class="action-btns">
                    <a-button type="link" size="small" @click="openTagImage(record)"><template #icon><EditOutlined /></template>标签</a-button>
                    <a-button type="link" size="small" danger @click="handleRemoveImage(record.id, record.repo_tags?.[0] || record.short_id)"><template #icon><DeleteOutlined /></template>删除</a-button>
                  </div>
                </template>
              </template>
            </a-table>
          </a-tab-pane>

          <a-tab-pane key="volumes" tab="存储卷">
            <div class="docker-toolbar">
              <a-button type="primary" @click="createVolumeVisible = true"><template #icon><PlusOutlined /></template>创建存储卷</a-button>
              <a-button danger @click="handlePruneVolumes"><template #icon><DeleteOutlined /></template>清理未使用卷</a-button>
            </div>
            <a-table :columns="volumeColumns" :data-source="volumes" :pagination="{ pageSize: 20, showTotal: (t: number) => `共 ${t} 个存储卷` }" :scroll="{ x: 1040 }" row-key="name" size="middle">
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'"><span class="container-name">{{ record.name }}</span></template>
                <template v-else-if="column.key === 'mountpoint'"><span class="port-text">{{ record.mountpoint }}</span></template>
                <template v-else-if="column.key === 'created_at'"><span class="port-text">{{ record.created_at || '-' }}</span></template>
                <template v-else-if="column.key === 'action'">
                  <a-button type="link" size="small" danger @click="handleRemoveVolume(record.name)"><template #icon><DeleteOutlined /></template>删除</a-button>
                </template>
              </template>
            </a-table>
          </a-tab-pane>

          <a-tab-pane key="registries" tab="仓库">
            <div class="docker-toolbar">
              <a-button type="primary" @click="openRegistryConfig"><template #icon><SettingOutlined /></template>仓库配置</a-button>
              <a-button @click="fetchRegistries"><template #icon><ReloadOutlined /></template>刷新</a-button>
            </div>
            <a-table :columns="registryColumns" :data-source="registries" :pagination="false" row-key="name" size="middle">
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'"><span class="container-name">{{ record.name }}</span></template>
                <template v-else-if="column.key === 'server_address'"><span class="port-text">{{ record.server_address }}</span></template>
              </template>
            </a-table>
          </a-tab-pane>

          <a-tab-pane key="compose" tab="Compose">
            <template v-if="!composeAvailable">
              <div class="compose-unavailable">
                <p>未检测到 Docker Compose 插件，Compose 功能不可用。</p>
                <p style="color: rgba(0,0,0,.35); font-size: 12px">请确认环境中已安装 <code>docker compose</code>（Docker Desktop 内置，Linux 需手动安装插件）。</p>
              </div>
            </template>
            <template v-else>
              <div class="docker-toolbar">
                <a-button type="primary" @click="composeUpVisible = true"><template #icon><PlusOutlined /></template>部署项目</a-button>
                <a-button @click="presetVisible = true"><template #icon><AppstoreOutlined /></template>预配置模板</a-button>
                <a-button @click="openVisualConfig"><template #icon><FormOutlined /></template>可视化配置</a-button>
              </div>
              <a-table :columns="composeColumns" :data-source="composeProjects" :pagination="{ pageSize: 20, showTotal: (t: number) => `共 ${t} 个项目` }" :scroll="{ x: 1080 }" row-key="name" size="middle">
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'name'"><span class="container-name">{{ record.name }}</span></template>
                  <template v-else-if="column.key === 'status'">
                    <span class="state-badge" :style="{ color: getComposeStatusColor(record.status) }">
                      <span class="state-indicator" :style="{ background: getComposeStatusColor(record.status) }" />{{ getComposeStatusLabel(record.status) }}
                    </span>
                  </template>
                  <template v-else-if="column.key === 'services'"><span class="port-text">{{ record.services?.join(', ') || '-' }}</span></template>
                  <template v-else-if="column.key === 'count'">
                    <span style="color:#52c41a">{{ record.running }} 运行</span> / <span style="color:#ff4d4f">{{ record.stopped }} 停止</span>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <div class="action-btns">
                      <a-button type="link" size="small" :loading="composeEditLoading" @click="editComposeProject(record.name)"><template #icon><EditOutlined /></template>编辑</a-button>
                      <a-button type="link" size="small" :loading="composeEditLoading" @click="visualEditComposeProject(record.name)"><template #icon><FormOutlined /></template>可视化</a-button>
                      <a-button v-if="record.status === 'stopped' || record.status === 'partial'" type="link" size="small" @click="composeStart(record.name).then(() => { message.success('已启动'); fetchComposeProjects() }).catch(() => message.error('启动失败'))">启动</a-button>
                      <a-button v-if="record.status === 'running' || record.status === 'partial'" type="link" size="small" @click="composeStop(record.name).then(() => { message.success('已停止'); fetchComposeProjects() }).catch(() => message.error('停止失败'))">停止</a-button>
                      <a-button v-if="record.status === 'running'" type="link" size="small" @click="composeRestart(record.name).then(() => { message.success('已重启'); fetchComposeProjects() }).catch(() => message.error('重启失败'))">重启</a-button>
                      <a-button type="link" size="small" danger @click="handleComposeDown(record.name)">移除</a-button>
                    </div>
                  </template>
                </template>
              </a-table>
            </template>
          </a-tab-pane>
        </a-tabs>
      </template>

      <div v-else class="docker-unavailable">
        <CloudOutlined class="unavailable-icon" />
        <p class="unavailable-title">无法连接到 Docker</p>
        <p class="unavailable-desc">请确认 Docker 服务已启动，或者在连接设置中配置远程 Docker 地址。</p>
        <div class="unavailable-actions">
          <a-button @click="openConfig"><template #icon><LinkOutlined /></template>连接设置</a-button>
          <a-button @click="loadData">重新检测</a-button>
        </div>
      </div>
    </a-spin>

    <a-modal v-model:open="configVisible" title="Docker 连接设置" :confirm-loading="configLoading" ok-text="保存" cancel-text="取消" width="720px" @ok="handleSaveConnection">
      <a-spin :spinning="configLoading">
        <div class="connection-panel">
          <div class="connection-hero">
            <div class="connection-hero-icon"><LinkOutlined /></div>
            <div class="connection-hero-body">
              <div class="connection-hero-title">Docker Engine 连接</div>
              <div class="connection-hero-desc">连接到本机或远程 Docker API，留空 Host 会使用默认环境连接。</div>
            </div>
            <span class="connection-mode-badge">{{ configForm.host ? '远程连接' : '本地默认' }}</span>
          </div>

          <div class="connection-port-grid">
            <button type="button" class="connection-port-card" @click="configForm.tls_verify = false">
              <span class="connection-port-title">2375 / 非 TLS</span>
              <span class="connection-port-desc">适合内网或受信网络，例如 tcp://10.1.0.3:2375</span>
            </button>
            <button type="button" class="connection-port-card connection-port-card-secure" @click="configForm.tls_verify = true">
              <span class="connection-port-title">2376 / TLS</span>
              <span class="connection-port-desc">需要 ca.pem、cert.pem、key.pem 证书目录</span>
            </button>
          </div>

          <div class="connection-section">
            <div class="config-field">
              <label class="config-label">Docker Host</label>
              <a-input v-model:value="configForm.host" placeholder="例如: tcp://10.1.0.3:2375" size="large" />
              <div class="config-help">常用格式：unix:///var/run/docker.sock、tcp://IP:2375、tcp://IP:2376</div>
            </div>
            <div class="connection-presets">
              <button type="button" @click="configForm.host = ''">本地默认</button>
              <button type="button" @click="configForm.host = 'unix:///var/run/docker.sock'">Linux Socket</button>
              <button type="button" @click="configForm.host = 'tcp://10.1.0.3:2375'; configForm.tls_verify = false">10.1.0.3:2375</button>
            </div>
          </div>

          <div class="connection-section connection-tls-section" :class="{ 'connection-tls-enabled': configForm.tls_verify }">
            <div class="connection-switch-row">
              <div>
                <div class="config-label">TLS 验证</div>
                <div class="config-help">端口 2376 会自动启用 TLS；2375 通常关闭 TLS。</div>
              </div>
              <a-switch v-model:checked="configForm.tls_verify" />
            </div>
            <div v-if="configForm.tls_verify" class="config-field">
              <label class="config-label">证书目录</label>
              <a-input v-model:value="configForm.cert_path" placeholder="包含 ca.pem、cert.pem、key.pem 的目录路径" />
              <div class="config-help">留空会跳过证书验证，仅建议在测试环境使用。</div>
            </div>
          </div>

          <div class="connection-test-row">
            <div class="connection-test-copy">
              <div class="config-label">连接测试</div>
              <div class="config-help">保存前建议先测试，错误详情会直接显示在页面提示中。</div>
            </div>
            <a-button :loading="testLoading" @click="handleTestConnection">测试连接</a-button>
          </div>
        </div>
      </a-spin>
    </a-modal>

    <a-modal v-model:open="createNetworkVisible" title="创建网络" :confirm-loading="createNetworkLoading" ok-text="创建" cancel-text="取消" @ok="handleCreateNetwork">
      <div class="config-form">
        <div class="config-field"><label class="config-label">网络名称 *</label><a-input v-model:value="newNetwork.name" placeholder="my-network" /></div>
        <div class="config-field"><label class="config-label">驱动</label><a-select v-model:value="newNetwork.driver" :options="[{value:'bridge',label:'bridge'},{value:'overlay',label:'overlay'},{value:'macvlan',label:'macvlan'},{value:'host',label:'host'}]" /></div>
        <div class="config-field"><label class="config-label">子网 (CIDR)</label><a-input v-model:value="newNetwork.subnet" placeholder="172.20.0.0/16" /></div>
        <div class="config-field"><label class="config-label">网关</label><a-input v-model:value="newNetwork.gateway" placeholder="172.20.0.1" /></div>
        <div class="config-field"><label class="config-label">内部网络</label><a-switch v-model:checked="newNetwork.internal" /></div>
      </div>
    </a-modal>

    <a-modal v-model:open="pullImageVisible" title="拉取镜像" :confirm-loading="pullImageLoading" ok-text="拉取" cancel-text="取消" @ok="handlePullImage">
      <div class="config-form">
        <div class="config-field"><label class="config-label">镜像名称 *</label><a-input v-model:value="pullImageForm.image" placeholder="nginx:latest 或 registry.example.com/app:1.0" /></div>
        <div class="config-field"><label class="config-label">使用仓库认证</label><a-select v-model:value="pullImageForm.registry" :options="[{label:'不使用认证',value:''}, ...registryOptions]" /></div>
      </div>
    </a-modal>

    <a-modal v-model:open="tagImageVisible" title="镜像打标签" :confirm-loading="tagImageLoading" ok-text="创建" cancel-text="取消" @ok="handleTagImage">
      <div class="config-form">
        <div class="config-field"><label class="config-label">源镜像 *</label><a-input v-model:value="tagImageForm.source" /></div>
        <div class="config-field"><label class="config-label">目标标签 *</label><a-input v-model:value="tagImageForm.target" placeholder="registry.example.com/app:1.0" /></div>
      </div>
    </a-modal>

    <a-modal v-model:open="registryVisible" title="仓库管理" :confirm-loading="registryLoading" ok-text="保存" cancel-text="取消" width="760px" @ok="handleSaveRegistries">
      <div class="visual-form">
        <div v-for="(registry, idx) in registryForm" :key="idx" class="visual-service-card">
          <div class="visual-service-header">
            <span class="visual-service-index">仓库 {{ idx + 1 }}</span>
            <div class="action-btns">
              <a-button type="link" size="small" @click="handleTestRegistry(registry)">测试登录</a-button>
              <a-button type="link" size="small" danger @click="removeRegistry(idx)"><template #icon><DeleteOutlined /></template></a-button>
            </div>
          </div>
          <div class="visual-row-2">
            <div class="config-field"><label class="config-label">名称 *</label><a-input v-model:value="registry.name" placeholder="Docker Hub" /></div>
            <div class="config-field"><label class="config-label">仓库地址 *</label><a-input v-model:value="registry.server_address" placeholder="https://index.docker.io/v1/" /></div>
          </div>
          <div class="visual-row-2">
            <div class="config-field"><label class="config-label">用户名</label><a-input v-model:value="registry.username" /></div>
            <div class="config-field"><label class="config-label">密码/Token</label><a-input-password v-model:value="registry.password" /></div>
          </div>
        </div>
        <a-button type="dashed" block @click="addRegistry"><template #icon><PlusOutlined /></template>添加仓库</a-button>
      </div>
    </a-modal>

    <a-modal v-model:open="createVolumeVisible" title="创建存储卷" :confirm-loading="createVolumeLoading" ok-text="创建" cancel-text="取消" @ok="handleCreateVolume">
      <div class="config-form">
        <div class="config-field"><label class="config-label">存储卷名称 *</label><a-input v-model:value="newVolume.name" placeholder="app_data" /></div>
        <div class="config-field"><label class="config-label">驱动</label><a-input v-model:value="newVolume.driver" placeholder="local" /></div>
        <div class="config-field"><label class="config-label">标签（每行 KEY=value）</label><a-textarea v-model:value="newVolume.labels" :rows="3" placeholder="app=demo" style="font-family: monospace" /></div>
        <div class="config-field"><label class="config-label">驱动选项（每行 KEY=value）</label><a-textarea v-model:value="newVolume.options" :rows="3" placeholder="type=nfs" style="font-family: monospace" /></div>
      </div>
    </a-modal>

    <a-modal v-model:open="composeUpVisible" title="部署 Compose 项目" :confirm-loading="composeUpLoading" ok-text="部署" cancel-text="取消" width="680px" @ok="handleComposeUp">
      <div class="config-form">
        <div class="config-field"><label class="config-label">项目名称（可选）</label><a-input v-model:value="composeProjectName" placeholder="留空则使用 compose 文件中的名称" /></div>
        <div class="config-field">
          <label class="config-label">docker-compose.yml 内容 *</label>
          <div class="compose-editor-toolbar">
            <button type="button" class="editor-toolbar-btn" @click="presetVisible = true"><AppstoreOutlined class="editor-toolbar-icon" />预配置模板</button>
            <button type="button" class="editor-toolbar-btn" @click="openVisualConfig"><FormOutlined class="editor-toolbar-icon" />可视化配置</button>
          </div>
          <div ref="composeEditorRef" class="compose-editor">
            <div class="compose-editor-gutter" @scroll.passive>
              <div v-for="(_, i) in composeLines" :key="i" class="line-num">{{ i + 1 }}</div>
            </div>
            <textarea
              v-model="composeContent"
              class="compose-editor-input"
              spellcheck="false"
              placeholder="services:&#10;  web:&#10;    image: nginx&#10;    ports:&#10;      - '80:80'"
              @scroll="onComposeEditorScroll"
            />
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal v-model:open="presetVisible" title="预配置模板" :footer="null" width="720px">
      <div class="preset-grid">
        <button v-for="p in composePresets" :key="p.key" type="button" class="preset-card" @click="applyPreset(p)">
          <div class="preset-name">{{ p.name }}</div>
          <div class="preset-desc">{{ p.desc }}</div>
        </button>
      </div>
    </a-modal>

    <a-modal v-model:open="visualVisible" title="可视化配置 Compose" width="860px" ok-text="生成配置" cancel-text="取消" @ok="applyVisualConfig">
      <div class="visual-form">
        <div v-for="(svc, idx) in visualServices" :key="idx" class="visual-service-card">
          <div class="visual-service-header">
            <span class="visual-service-index">服务 {{ idx + 1 }}</span>
            <a-button v-if="visualServices.length > 1" type="link" size="small" danger @click="removeVisualService(idx)"><template #icon><DeleteOutlined /></template></a-button>
          </div>
          <div class="visual-row-2">
            <div class="config-field">
              <label class="config-label">服务名称 *</label>
              <a-input v-model:value="svc.name" placeholder="如: web" />
            </div>
            <div class="config-field">
              <label class="config-label">镜像 *</label>
              <a-input v-model:value="svc.image" placeholder="如: nginx:latest" />
            </div>
          </div>
          <div class="visual-row-2">
            <div class="config-field">
              <label class="config-label">容器名称</label>
              <a-input v-model:value="svc.container_name" placeholder="如: my-nginx" />
            </div>
            <div class="config-field">
              <label class="config-label">重启策略</label>
              <a-select v-model:value="svc.restart" :options="[{value:'no',label:'no'},{value:'always',label:'always'},{value:'unless-stopped',label:'unless-stopped'},{value:'on-failure',label:'on-failure'}]" />
            </div>
          </div>
          <div class="config-field">
            <label class="config-label">端口映射（每行一个，如 80:80）</label>
            <a-textarea v-model:value="svc.ports" :rows="2" placeholder="80:80&#10;443:443" style="font-family: monospace" />
          </div>
          <div class="config-field">
            <label class="config-label">环境变量（每行一个，如 KEY=value）</label>
            <a-textarea v-model:value="svc.environment" :rows="2" placeholder="MYSQL_ROOT_PASSWORD=root123&#10;MYSQL_DATABASE=app" style="font-family: monospace" />
          </div>
          <div class="config-field">
            <label class="config-label">挂载卷（每行一个，如 ./data:/data 或 data_vol）</label>
            <a-textarea v-model:value="svc.volumes" :rows="2" placeholder="./html:/usr/share/nginx/html&#10;./conf:/etc/nginx/conf.d" style="font-family: monospace" />
          </div>
          <div class="visual-row-2">
            <div class="config-field">
              <label class="config-label">启动命令</label>
              <a-input v-model:value="svc.command" placeholder="如: redis-server --appendonly yes" />
            </div>
            <div class="config-field">
              <label class="config-label">依赖服务（逗号分隔）</label>
              <a-input v-model:value="svc.depends_on" placeholder="如: db, redis" />
            </div>
          </div>
        </div>
        <a-button type="dashed" block @click="addVisualService"><template #icon><PlusOutlined /></template>添加服务</a-button>
      </div>
    </a-modal>
  </div>
</template>

<style scoped>
.docker-page { min-width: 0 }
.docker-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 16px; margin-bottom: 20px }
.docker-eyebrow { display: block; margin-bottom: 4px; color: #1677ff; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase }
.docker-title { margin: 0; color: rgba(0,0,0,.88); font-size: 24px; font-weight: 700; line-height: 1.25 }
.docker-heading-actions { display: flex; align-items: center; gap: 12px }
.docker-status { display: inline-flex; align-items: center; gap: 6px; padding: 4px 12px; background: rgba(82,196,26,.06); border: 1px solid rgba(82,196,26,.2); border-radius: 999px }
.docker-status-offline { background: rgba(255,77,79,.06); border-color: rgba(255,77,79,.2) }
.status-dot { width: 6px; height: 6px; border-radius: 50%; background: #52c41a }
.docker-status-offline .status-dot { background: #ff4d4f }
.status-text { color: rgba(0,0,0,.65); font-size: 12px; white-space: nowrap }
.stats-grid { display: grid; grid-template-columns: repeat(4,1fr); gap: 12px; margin-bottom: 20px }
.stats-grid-extended { grid-template-columns: repeat(6,1fr) }
.stat-card { padding: 16px 20px; background: #fff; border: 1px solid rgba(5,5,5,.06); border-radius: 10px; text-align: center }
.stat-value { color: rgba(0,0,0,.88); font-size: 28px; font-weight: 700 }
.stat-running .stat-value { color: #52c41a }
.stat-stopped .stat-value { color: #ff4d4f }
.stat-label { margin-top: 4px; color: rgba(0,0,0,.45); font-size: 13px }
.docker-toolbar { display: flex; align-items: center; gap: 16px; margin-bottom: 16px }
.search-input { max-width: 320px }
.toggle-row { display: inline-flex; align-items: center; gap: 8px }
.toggle-label { color: rgba(0,0,0,.65); font-size: 13px; white-space: nowrap }
.container-name { color: rgba(0,0,0,.88); font-weight: 600 }
.container-image { color: rgba(0,0,0,.65); font-size: 12px; word-break: break-all }
.container-id { padding: 2px 8px; color: rgba(0,0,0,.45); font-family: monospace; font-size: 12px; background: rgba(0,0,0,.03); border-radius: 4px }
.state-badge { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; font-weight: 500 }
.state-indicator { display: inline-block; width: 6px; height: 6px; border-radius: 50% }
.port-text { color: rgba(0,0,0,.65); font-size: 12px; word-break: break-all }
.action-btns { display: inline-flex; align-items: center; gap: 0 }
.tag-list { display: flex; flex-wrap: wrap; gap: 6px }
.badge { display: inline-block; padding: 1px 6px; margin-left: 6px; font-size: 11px; border-radius: 4px; vertical-align: middle }
.badge-internal { color: #722ed1; background: rgba(114,46,209,.08); border: 1px solid rgba(114,46,209,.2) }
.compose-unavailable { padding: 40px; text-align: center; color: rgba(0,0,0,.45) }
.compose-unavailable code { padding: 2px 6px; background: rgba(0,0,0,.04); border-radius: 4px; font-size: 13px }
.docker-unavailable { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 80px 24px; text-align: center }
.unavailable-icon { margin-bottom: 16px; color: rgba(0,0,0,.15); font-size: 48px }
.unavailable-title { color: rgba(0,0,0,.65); font-size: 16px; font-weight: 500 }
.unavailable-desc { margin: 8px 0 20px; color: rgba(0,0,0,.45); font-size: 14px }
.unavailable-actions { display: flex; gap: 12px }
.connection-panel { display: flex; flex-direction: column; gap: 16px }
.connection-hero { display: flex; align-items: center; gap: 14px; padding: 16px; background: linear-gradient(135deg, rgba(22,119,255,.08), rgba(22,119,255,.02)); border: 1px solid rgba(22,119,255,.12); border-radius: 14px }
.connection-hero-icon { display: inline-flex; align-items: center; justify-content: center; width: 42px; height: 42px; color: #1677ff; font-size: 20px; background: #fff; border: 1px solid rgba(22,119,255,.12); border-radius: 12px; box-shadow: 0 8px 20px rgba(22,119,255,.08) }
.connection-hero-body { flex: 1; min-width: 0 }
.connection-hero-title { color: rgba(0,0,0,.88); font-size: 16px; font-weight: 700 }
.connection-hero-desc { margin-top: 3px; color: rgba(0,0,0,.48); font-size: 13px; line-height: 1.5 }
.connection-mode-badge { flex: none; padding: 4px 10px; color: #1677ff; font-size: 12px; font-weight: 600; background: #fff; border: 1px solid rgba(22,119,255,.18); border-radius: 999px }
.connection-port-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px }
.connection-port-card { display: flex; flex-direction: column; gap: 4px; padding: 13px 14px; text-align: left; cursor: pointer; background: #fff; border: 1px solid rgba(5,5,5,.08); border-radius: 12px; transition: border-color .2s, box-shadow .2s, transform .2s }
.connection-port-card:hover { border-color: rgba(22,119,255,.35); box-shadow: 0 8px 24px rgba(15,23,42,.06); transform: translateY(-1px) }
.connection-port-card-secure:hover { border-color: rgba(82,196,26,.36) }
.connection-port-title { color: rgba(0,0,0,.82); font-size: 13px; font-weight: 700 }
.connection-port-desc { color: rgba(0,0,0,.45); font-size: 12px; line-height: 1.5 }
.connection-section { display: flex; flex-direction: column; gap: 10px; padding: 14px; background: #fff; border: 1px solid rgba(5,5,5,.06); border-radius: 12px }
.connection-presets { display: flex; flex-wrap: wrap; gap: 8px }
.connection-presets button { padding: 4px 10px; color: rgba(0,0,0,.65); font-size: 12px; cursor: pointer; background: rgba(0,0,0,.02); border: 1px solid rgba(5,5,5,.08); border-radius: 999px; transition: color .2s, border-color .2s, background .2s }
.connection-presets button:hover { color: #1677ff; background: rgba(22,119,255,.04); border-color: rgba(22,119,255,.3) }
.connection-tls-section { background: rgba(0,0,0,.012) }
.connection-tls-enabled { background: rgba(82,196,26,.035); border-color: rgba(82,196,26,.16) }
.connection-switch-row { display: flex; align-items: center; justify-content: space-between; gap: 16px }
.connection-test-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 14px; background: rgba(0,0,0,.018); border: 1px dashed rgba(5,5,5,.12); border-radius: 12px }
.connection-test-copy { min-width: 0 }
.config-form { display: flex; flex-direction: column; gap: 16px; padding: 8px 0 }
.config-hint { padding: 10px 12px; color: rgba(0,0,0,.55); font-size: 13px; background: rgba(22,119,255,.03); border: 1px solid rgba(22,119,255,.1); border-radius: 8px }
.config-hint-title { margin-bottom: 6px; color: rgba(0,0,0,.75); font-weight: 500 }
.config-hint-list { margin: 0; padding-left: 18px }
.config-hint-list li { margin-bottom: 2px; line-height: 1.6 }
.config-field { display: flex; flex-direction: column; gap: 6px }
.config-label { color: rgba(0,0,0,.88); font-size: 13px; font-weight: 500 }
.config-help { color: rgba(0,0,0,.35); font-size: 12px }
.config-test { padding-top: 4px; border-top: 1px solid rgba(5,5,5,.06) }
.docker-tabs :deep(.ant-tabs-nav) { margin-bottom: 16px }
.compose-editor { display: flex; border: 1px solid rgba(5,5,5,.15); border-radius: 8px; overflow: hidden; background: #fafafa; height: 320px }
.compose-editor-gutter { flex: none; width: 42px; overflow: hidden; background: rgba(0,0,0,.03); border-right: 1px solid rgba(5,5,5,.08); user-select: none; padding-top: 10px }
.compose-editor-gutter .line-num { height: 22px; padding-right: 12px; color: rgba(0,0,0,.3); font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace; font-size: 12px; line-height: 22px; text-align: right }
.compose-editor-input { flex: 1; resize: none; padding: 10px 12px; border: none; outline: none; background: transparent; color: rgba(0,0,0,.88); font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace; font-size: 13px; line-height: 22px; tab-size: 2 }
.compose-editor-input::placeholder { color: rgba(0,0,0,.25) }
.compose-editor-toolbar { display: flex; gap: 8px; margin-bottom: 8px }
.editor-toolbar-btn { display: inline-flex; align-items: center; gap: 5px; padding: 4px 10px; color: rgba(0,0,0,.65); font-size: 12px; cursor: pointer; background: rgba(0,0,0,.02); border: 1px solid rgba(5,5,5,.08); border-radius: 6px; transition: color .2s, border-color .2s, background .2s }
.editor-toolbar-btn:hover { color: #1677ff; background: rgba(22,119,255,.04); border-color: rgba(22,119,255,.3) }
.editor-toolbar-icon { font-size: 12px }
.preset-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px }
.preset-card { display: flex; flex-direction: column; gap: 4px; padding: 16px 18px; text-align: left; cursor: pointer; background: #fff; border: 1px solid rgba(5,5,5,.08); border-radius: 12px; transition: border-color .2s, box-shadow .2s, transform .2s }
.preset-card:hover { border-color: rgba(22,119,255,.4); box-shadow: 0 6px 20px rgba(15,23,42,.06); transform: translateY(-1px) }
.preset-name { color: rgba(0,0,0,.88); font-size: 15px; font-weight: 600 }
.preset-desc { color: rgba(0,0,0,.45); font-size: 12px }
.visual-form { display: flex; flex-direction: column; gap: 16px; padding: 4px 0 }
.visual-service-card { display: flex; flex-direction: column; gap: 12px; padding: 16px; background: #fff; border: 1px solid rgba(5,5,5,.06); border-radius: 12px }
.visual-service-header { display: flex; align-items: center; justify-content: space-between }
.visual-service-index { color: rgba(0,0,0,.65); font-size: 13px; font-weight: 600 }
.visual-row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px }

@media (max-width: 768px) {
  .docker-heading { flex-direction: column; align-items: flex-start; gap: 8px }
  .docker-heading-actions { flex-wrap: wrap }
  .stats-grid, .stats-grid-extended { grid-template-columns: repeat(2,1fr) }
  .docker-toolbar { flex-direction: column; align-items: stretch }
  .search-input { max-width: none }
  .connection-hero { align-items: flex-start; flex-direction: column }
  .connection-mode-badge { align-self: flex-start }
  .connection-port-grid { grid-template-columns: 1fr }
  .connection-switch-row, .connection-test-row { align-items: flex-start; flex-direction: column }
  .preset-grid { grid-template-columns: 1fr }
  .visual-row-2 { grid-template-columns: 1fr }
}
</style>
