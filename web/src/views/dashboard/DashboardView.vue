<script setup lang="ts">
import type { MenuItemType } from 'antdv-next'
import {
  CloudServerOutlined,
  DashboardOutlined,
  DownOutlined,
  HddOutlined,
  MonitorOutlined,
} from '@antdv-next/icons'
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { getPrivateStatus, getStatus } from '@/api/server'

interface PanelStatus {
  version: string
  memory: string
  uptime: string
}

interface PrivateServerStatus {
  ip_addresses: string[]
  uptime: string
  uptime_seconds: number
  load: LoadStatus
  cpu: CPUStatus
  memory: MemoryStatus
  storage: StorageStatus
  kernel: KernelStatus
  distribution: DistributionStatus
}

interface LoadStatus {
  load1: number
  load5: number
  load15: number
  per_core_load1: number
}

interface CPUStatus {
  usage_percent: number
  logical_cores: number
  physical_cores?: number
  model_name?: string
  architecture: string
}

interface MemoryStatus {
  total_text: string
  used_text: string
  available_text: string
  used_percent: number
}

interface StorageStatus {
  filesystems: FilesystemStatus[]
}

interface FilesystemStatus {
  path: string
  filesystem?: string
  total_text: string
  used_text: string
  available_text: string
  used_percent: number
}

interface KernelStatus {
  release?: string
  goos: string
  goarch: string
}

interface DistributionStatus {
  pretty_name?: string
  name?: string
  version?: string
}

interface MetricItem {
  key: string
  label: string
  value: string
  detail: string
  percent: number
  icon: typeof MonitorOutlined
}

const loading = shallowRef(true)
const autoRefreshEnabled = shallowRef(true)
const refreshIntervalSeconds = shallowRef(5)
const refreshIntervalOptions = [
  { label: '5 秒', value: 5 },
  { label: '10 秒', value: 10 },
  { label: '30 秒', value: 30 },
  { label: '1 分钟', value: 60 },
  { label: '5 分钟', value: 300 },
]

const intervalMenuItems: MenuItemType[] = refreshIntervalOptions.map((opt) => ({
  key: String(opt.value),
  label: opt.label,
}))

const currentIntervalLabel = computed(() => {
  const match = refreshIntervalOptions.find(
    (opt) => opt.value === refreshIntervalSeconds.value,
  )
  return match?.label ?? `${refreshIntervalSeconds.value} 秒`
})

function onIntervalMenuClick({ key }: { key: string }) {
  refreshIntervalSeconds.value = Number(key)
}
const panelStatus = ref<PanelStatus | null>(null)
const serverStatus = ref<PrivateServerStatus | null>(null)
let timer: ReturnType<typeof setInterval> | null = null

const refreshHint = computed(() => {
  return autoRefreshEnabled.value
    ? `每 ${refreshIntervalSeconds.value} 秒自动刷新`
    : '自动刷新已关闭'
})

async function fetchStatus() {
  try {
    const [panelRes, serverRes] = await Promise.allSettled([
      getStatus(),
      getPrivateStatus(),
    ])

    if (panelRes.status === 'fulfilled') {
      panelStatus.value = panelRes.value.data as PanelStatus
    }

    if (serverRes.status === 'fulfilled') {
      serverStatus.value = serverRes.value.data as PrivateServerStatus
    }
  } finally {
    loading.value = false
  }
}

const primaryFilesystem = computed(() => {
  const filesystems = serverStatus.value?.storage.filesystems ?? []
  return filesystems.find((item) => item.path === '/') ?? filesystems[0] ?? null
})

const serverName = computed(() => {
  const distribution = serverStatus.value?.distribution
  if (distribution?.pretty_name) return distribution.pretty_name
  if (distribution?.name && distribution.version) {
    return `${distribution.name} ${distribution.version}`
  }
  return distribution?.name ?? '未知发行版'
})

const kernelInfo = computed(() => {
  const kernel = serverStatus.value?.kernel
  if (!kernel) return '-'
  return kernel.release ?? `${kernel.goos}/${kernel.goarch}`
})

const loadInfo = computed(() => {
  const server = serverStatus.value
  if (!server) return '-'
  if (server.kernel.goos === 'windows') return '不适用'
  return `${server.load.load1} / ${server.load.load5} / ${server.load.load15}`
})

const cpuDetail = computed(() => {
  const cpu = serverStatus.value?.cpu
  if (!cpu) return '-'

  const cores = cpu.physical_cores
    ? `${cpu.physical_cores} 核 / ${cpu.logical_cores} 线程`
    : `${cpu.logical_cores} 线程`

  return cpu.model_name ? `${cores} · ${cpu.model_name}` : cores
})

const serverMetrics = computed<MetricItem[]>(() => {
  const server = serverStatus.value
  const storage = primaryFilesystem.value

  return [
    {
      key: 'cpu',
      label: 'CPU',
      value: formatPercent(server?.cpu.usage_percent),
      detail: cpuDetail.value,
      percent: server?.cpu.usage_percent ?? 0,
      icon: MonitorOutlined,
    },
    {
      key: 'memory',
      label: '内存',
      value: formatPercent(server?.memory.used_percent),
      detail: server
        ? `${server.memory.used_text} / ${server.memory.total_text}，可用 ${server.memory.available_text}`
        : '-',
      percent: server?.memory.used_percent ?? 0,
      icon: CloudServerOutlined,
    },
    {
      key: 'storage',
      label: '存储',
      value: formatPercent(storage?.used_percent),
      detail: storage
        ? `${storage.path} ${storage.used_text} / ${storage.total_text}，可用 ${storage.available_text}`
        : '-',
      percent: storage?.used_percent ?? 0,
      icon: HddOutlined,
    },
  ]
})

const panelItems = computed(() => [
  { label: '版本', value: panelStatus.value?.version ?? '-' },
  { label: '运行', value: panelStatus.value?.uptime ?? '-' },
  { label: '占用', value: panelStatus.value?.memory ?? '-' },
])

const hostIPs = computed(() => {
  const ips = serverStatus.value?.ip_addresses
  if (!ips || ips.length === 0) return '-'
  return ips.join('、')
})

const hostUptime = computed(() => {
  return serverStatus.value?.uptime ?? '-'
})

const systemItems = computed(() => [
  { label: '系统', value: serverName.value },
  { label: '内核', value: kernelInfo.value },
  { label: '负载', value: loadInfo.value },
  { label: '地址', value: hostIPs.value },
  { label: '运行', value: hostUptime.value },
])

function formatPercent(value?: number) {
  if (typeof value !== 'number') return '-'
  return `${value.toFixed(2)}%`
}

function normalizePercent(value: number) {
  return Math.min(Math.max(value, 0), 100)
}

const allFilesystems = computed(() => {
  return serverStatus.value?.storage.filesystems ?? []
})

function stopRefreshTimer() {
  if (!timer) return
  clearInterval(timer)
  timer = null
}

function startRefreshTimer() {
  stopRefreshTimer()
  if (!autoRefreshEnabled.value) return
  timer = setInterval(fetchStatus, refreshIntervalSeconds.value * 1000)
}

watch([autoRefreshEnabled, refreshIntervalSeconds], startRefreshTimer)

onMounted(() => {
  fetchStatus()
  startRefreshTimer()
})

onBeforeUnmount(() => {
  stopRefreshTimer()
})
</script>

<template>
  <div class="dashboard-page">
    <a-spin :spinning="loading">
      <div class="dashboard-heading">
        <div>
          <span class="dashboard-eyebrow">Overview</span>
          <h2 class="dashboard-title">运行概览</h2>
        </div>
        <div
          class="refresh-controls"
          :class="{ 'refresh-controls-disabled': !autoRefreshEnabled }"
        >
          <div class="refresh-status">
            <span class="refresh-dot" />
            <span class="refresh-hint">{{ refreshHint }}</span>
          </div>
          <div class="refresh-divider" />
          <label class="refresh-switch-row">
            <span class="refresh-label">自动</span>
            <a-switch
              v-model:checked="autoRefreshEnabled"
              size="small"
            />
          </label>
          <label class="refresh-interval-control">
            <span class="refresh-label">间隔</span>
            <a-dropdown
              :menu="{ items: intervalMenuItems, onClick: onIntervalMenuClick, selectedKeys: [String(refreshIntervalSeconds)] }"
              :disabled="!autoRefreshEnabled"
              :trigger="['click']"
            >
              <span
                class="interval-trigger"
                :class="{ 'interval-trigger-disabled': !autoRefreshEnabled }"
                @click.prevent
              >
                {{ currentIntervalLabel }}
                <DownOutlined class="interval-arrow" />
              </span>
            </a-dropdown>
          </label>
        </div>
      </div>

      <div class="overview-grid">
        <div class="overview-card panel-card">
          <div class="overview-header">
            <DashboardOutlined class="overview-icon" />
            <div>
              <span class="overview-title">面板状态</span>
              <span class="overview-subtitle">LynxPilot 服务进程</span>
            </div>
          </div>
          <div class="info-grid panel-info-grid">
            <div
              v-for="item in panelItems"
              :key="item.label"
              class="info-item"
            >
              <span class="overview-label">{{ item.label }}</span>
              <span class="overview-value">{{ item.value }}</span>
            </div>
          </div>
        </div>

        <div class="overview-card server-card">
          <div class="overview-header">
            <CloudServerOutlined class="overview-icon server-icon" />
            <div>
              <span class="overview-title">主机状态</span>
              <span class="overview-subtitle">{{ serverName }}</span>
            </div>
          </div>

          <div class="info-grid system-info-grid">
            <div
              v-for="item in systemItems"
              :key="item.label"
              class="info-item"
            >
              <span class="overview-label">{{ item.label }}</span>
              <span class="overview-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="metric-grid">
        <div
          v-for="metric in serverMetrics"
          :key="metric.key"
          class="metric-card"
          :class="{ 'metric-card--storage': metric.key === 'storage' }"
        >
          <div class="metric-heading">
            <span class="metric-name">
              <component :is="metric.icon" />
              {{ metric.label }}
            </span>
            <span class="metric-value">{{ metric.value }}</span>
          </div>
          <div class="metric-bar">
            <span
              class="metric-bar-inner"
              :style="{ width: `${normalizePercent(metric.percent)}%` }"
            />
          </div>
          <div class="metric-detail">
            {{ metric.detail }}
          </div>
          <template v-if="metric.key === 'storage' && allFilesystems.length > 1">
            <div class="storage-hover-trigger" />
            <div class="storage-popup">
              <div class="storage-popup-title">全部存储</div>
              <div v-for="fs in allFilesystems" :key="fs.path" class="storage-popup-item">
                <div class="storage-popup-row">
                  <span class="storage-popup-path">{{ fs.path }}</span>
                  <span class="storage-popup-pct">{{ formatPercent(fs.used_percent) }}</span>
                </div>
                <div class="storage-popup-bar">
                  <span class="storage-popup-bar-inner" :style="{ width: `${normalizePercent(fs.used_percent)}%` }" />
                </div>
                <div class="storage-popup-info">
                  {{ fs.used_text }} / {{ fs.total_text }}，可用 {{ fs.available_text }}
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
.dashboard-page {
  min-width: 0;
}

.dashboard-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.dashboard-eyebrow {
  display: block;
  margin-bottom: 4px;
  color: #1677ff;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.dashboard-title {
  margin: 0;
  color: rgba(0, 0, 0, 0.88);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.25;
}

.refresh-controls {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  min-width: 0;
  padding: 8px 10px 8px 12px;
  background: rgba(22, 119, 255, 0.04);
  border: 1px solid rgba(22, 119, 255, 0.12);
  border-radius: 999px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6);
  transition:
    background 0.2s ease,
    border-color 0.2s ease;
}

.refresh-controls-disabled {
  background: rgba(0, 0, 0, 0.025);
  border-color: rgba(5, 5, 5, 0.08);
}

.refresh-status,
.refresh-switch-row,
.refresh-interval-control {
  display: inline-flex;
  align-items: center;
}

.refresh-status {
  gap: 7px;
  min-width: 0;
}

.refresh-dot {
  width: 7px;
  height: 7px;
  background: #52c41a;
  border-radius: 50%;
  box-shadow: 0 0 0 3px rgba(82, 196, 26, 0.16);
}

.refresh-controls-disabled .refresh-dot {
  background: #bfbfbf;
  box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.06);
}

.refresh-divider {
  width: 1px;
  height: 18px;
  background: rgba(5, 5, 5, 0.08);
}

.refresh-switch-row {
  gap: 8px;
}

.refresh-hint,
.refresh-label {
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
  white-space: nowrap;
}

.refresh-interval-control {
  gap: 6px;
}

.interval-trigger {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 10px;
  color: rgba(0, 0, 0, 0.65);
  font-size: 12px;
  white-space: nowrap;
  cursor: pointer;
  background: #fff;
  border: 1px solid rgba(5, 5, 5, 0.1);
  border-radius: 999px;
  transition:
    color 0.2s,
    border-color 0.2s;
}

.interval-trigger:hover {
  color: #1677ff;
  border-color: #1677ff;
}

.interval-trigger-disabled {
  color: rgba(0, 0, 0, 0.25);
  cursor: not-allowed;
  background: rgba(0, 0, 0, 0.02);
}

.interval-trigger-disabled:hover {
  color: rgba(0, 0, 0, 0.25);
  border-color: rgba(5, 5, 5, 0.1);
}

.interval-arrow {
  font-size: 10px;
}

.overview-grid {
  display: grid;
  grid-template-columns: minmax(260px, 0.85fr) minmax(420px, 1.35fr);
  gap: 16px;
  align-items: stretch;
  margin-bottom: 16px;
}

.overview-card {
  position: relative;
  overflow: hidden;
  background: #fff;
  border: 1px solid rgba(5, 5, 5, 0.06);
  border-radius: 14px;
  padding: 22px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
}

.server-card {
  background:
    linear-gradient(135deg, rgba(22, 119, 255, 0.06), transparent 42%),
    #fff;
}

.overview-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.overview-icon {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  color: #1677ff;
  font-size: 20px;
  background: #e6f4ff;
  border-radius: 12px;
}

.server-icon {
  color: #389e0d;
  background: #f6ffed;
}

.overview-title {
  display: block;
  color: rgba(0, 0, 0, 0.88);
  font-size: 16px;
  font-weight: 600;
}

.overview-subtitle {
  display: block;
  max-width: 100%;
  margin-top: 2px;
  overflow: hidden;
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-grid {
  display: grid;
  gap: 12px;
}

.panel-info-grid {
  grid-template-columns: 1fr;
}

.system-info-grid {
  grid-template-columns: repeat(6, minmax(0, 1fr));
}

.system-info-grid .info-item:nth-child(-n+3) {
  grid-column: span 2;
}

.system-info-grid .info-item:nth-child(n+4) {
  grid-column: span 3;
}

.info-item {
  min-width: 0;
  padding: 12px;
  background: rgba(0, 0, 0, 0.018);
  border: 1px solid rgba(5, 5, 5, 0.04);
  border-radius: 10px;
}

.overview-label {
  display: block;
  margin-bottom: 6px;
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
}

.overview-value {
  display: block;
  min-width: 0;
  overflow: hidden;
  color: rgba(0, 0, 0, 0.88);
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.metric-card {
  position: relative;
  min-width: 0;
  padding: 18px;
  background: #fff;
  border: 1px solid rgba(5, 5, 5, 0.06);
  border-radius: 14px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
}

.metric-card--storage {
  cursor: default;
}

.storage-hover-trigger {
  position: absolute;
  inset: 0;
  z-index: 1;
}

.storage-popup {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 50;
  min-width: 320px;
  padding: 14px;
  pointer-events: none;
  opacity: 0;
  background: #fff;
  border: 1px solid rgba(5, 5, 5, 0.1);
  border-radius: 14px;
  box-shadow: 0 12px 40px rgba(15, 23, 42, 0.14);
  transform: translateY(4px);
  transition: opacity .2s, transform .2s;
}

.metric-card--storage:hover .storage-popup {
  pointer-events: auto;
  opacity: 1;
  transform: translateY(0);
}

.storage-popup-title {
  margin-bottom: 10px;
  color: rgba(0, 0, 0, 0.88);
  font-size: 13px;
  font-weight: 600;
}

.storage-popup-item {
  padding: 8px 0;
}

.storage-popup-item + .storage-popup-item {
  border-top: 1px solid rgba(5, 5, 5, 0.06);
}

.storage-popup-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.storage-popup-path {
  color: rgba(0, 0, 0, 0.88);
  font-size: 12px;
  font-weight: 600;
  font-family: 'SFMono-Regular', Consolas, monospace;
}

.storage-popup-pct {
  color: rgba(0, 0, 0, 0.65);
  font-size: 12px;
  font-weight: 600;
}

.storage-popup-bar {
  height: 4px;
  margin-top: 6px;
  overflow: hidden;
  background: rgba(22, 119, 255, 0.09);
  border-radius: 999px;
}

.storage-popup-bar-inner {
  display: block;
  height: 100%;
  background: #1677ff;
  border-radius: inherit;
}

.storage-popup-info {
  margin-top: 4px;
  color: rgba(0, 0, 0, 0.45);
  font-size: 11px;
}

.metric-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.metric-name {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: rgba(0, 0, 0, 0.65);
  font-size: 13px;
}

.metric-value {
  color: rgba(0, 0, 0, 0.88);
  font-size: 18px;
  font-weight: 700;
}

.metric-bar {
  height: 8px;
  overflow: hidden;
  background: rgba(22, 119, 255, 0.09);
  border-radius: 999px;
}

.metric-bar-inner {
  display: block;
  height: 100%;
  background: #1677ff;
  border-radius: inherit;
  transition: width 0.2s ease;
}

.metric-detail {
  min-height: 36px;
  margin-top: 10px;
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
  line-height: 1.5;
  word-break: break-word;
}

@media (max-width: 900px) {
  .overview-grid {
    grid-template-columns: 1fr;
  }

  .system-info-grid,
  .metric-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .dashboard-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }

  .refresh-controls {
    justify-content: flex-start;
    flex-wrap: wrap;
    border-radius: 14px;
  }

  .refresh-divider {
    display: none;
  }

  .overview-card,
  .metric-card {
    padding: 16px;
  }

  .dashboard-title {
    font-size: 22px;
  }
}
</style>
