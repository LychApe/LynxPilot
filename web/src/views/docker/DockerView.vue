<script setup lang="ts">
import type { ComposeProject, ContainerInfo, DockerConnection, DockerPingResult, NetworkInfo } from '@/api/docker'
import {
  checkComposeAvailable,
  composeDown,
  composeRestart,
  composeStart,
  composeStop,
  composeUp,
  createNetwork,
  getDockerConnection,
  listComposeProjects,
  listContainers,
  listNetworks,
  pingDocker,
  removeContainer,
  removeNetwork,
  restartContainer,
  saveDockerConnection,
  startContainer,
  stopContainer,
  testDockerConnection,
} from '@/api/docker'
import {
  CaretRightOutlined,
  CloudOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
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

const composeUpVisible = shallowRef(false)
const composeUpLoading = shallowRef(false)
const composeContent = ref('')
const composeProjectName = ref('')

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
      const [, composeRes] = await Promise.allSettled([
        fetchContainers(),
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
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
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
        <div class="stats-grid">
          <div class="stat-card"><div class="stat-value">{{ containers.length }}</div><div class="stat-label">全部容器</div></div>
          <div class="stat-card stat-running"><div class="stat-value">{{ runningCount }}</div><div class="stat-label">运行中</div></div>
          <div class="stat-card stat-stopped"><div class="stat-value">{{ stoppedCount }}</div><div class="stat-label">已停止</div></div>
          <div class="stat-card"><div class="stat-value">{{ networks.length }}</div><div class="stat-label">网络</div></div>
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
              </div>
              <a-table :columns="composeColumns" :data-source="composeProjects" :pagination="{ pageSize: 20, showTotal: (t: number) => `共 ${t} 个项目` }" :scroll="{ x: 920 }" row-key="name" size="middle">
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

    <a-modal v-model:open="composeUpVisible" title="部署 Compose 项目" :confirm-loading="composeUpLoading" ok-text="部署" cancel-text="取消" width="680px" @ok="handleComposeUp">
      <div class="config-form">
        <div class="config-field"><label class="config-label">项目名称（可选）</label><a-input v-model:value="composeProjectName" placeholder="留空则使用 compose 文件中的名称" /></div>
        <div class="config-field">
          <label class="config-label">docker-compose.yml 内容 *</label>
          <a-textarea v-model:value="composeContent" :rows="12" placeholder="services:&#10;  web:&#10;    image: nginx&#10;    ports:&#10;      - '80:80'" style="font-family: monospace" />
        </div>
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

@media (max-width: 768px) {
  .docker-heading { flex-direction: column; align-items: flex-start; gap: 8px }
  .docker-heading-actions { flex-wrap: wrap }
  .stats-grid { grid-template-columns: repeat(2,1fr) }
  .docker-toolbar { flex-direction: column; align-items: stretch }
  .search-input { max-width: none }
  .connection-hero { align-items: flex-start; flex-direction: column }
  .connection-mode-badge { align-self: flex-start }
  .connection-port-grid { grid-template-columns: 1fr }
  .connection-switch-row, .connection-test-row { align-items: flex-start; flex-direction: column }
}
</style>
