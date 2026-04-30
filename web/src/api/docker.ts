import req from '@/utils/req'

export interface ContainerPort {
  ip: string
  private_port: number
  public_port: number
  type: string
}

export interface ContainerInfo {
  id: string
  names: string[]
  image: string
  state: string
  status: string
  created: number
  ports: ContainerPort[]
  command: string
}

export interface ContainerDetail {
  id: string
  name: string
  image: string
  state: string
  status: string
  created: number
  command: string
  env: string[]
  ports: ContainerPort[]
  network_mode: string
  ip_address: string
  gateway: string
  mac_address: string
  restart_count: number
  started_at: string
  finished_at: string
}

export interface ContainerStats {
  cpu_percent: number
  memory_usage: number
  memory_limit: number
  memory_percent: number
  network_rx: number
  network_tx: number
  block_read: number
  block_write: number
  pids: number
  memory_usage_text: string
  memory_limit_text: string
}

export function pingDocker() {
  return req.get<unknown, { data: { available: boolean } }>('/private/docker/ping')
}

export function listContainers(all = false) {
  return req.get<unknown, { data: ContainerInfo[] }>('/private/docker/containers', {
    params: { all },
  })
}

export function searchContainers(name: string) {
  return req.get<unknown, { data: ContainerInfo[] }>('/private/docker/containers/search', {
    params: { name },
  })
}

export function getContainerDetail(id: string) {
  return req.get<unknown, { data: ContainerDetail }>(`/private/docker/containers/${id}`)
}

export function getContainerStats(id: string) {
  return req.get<unknown, { data: ContainerStats }>(`/private/docker/containers/${id}/stats`)
}

export function startContainer(id: string) {
  return req.post(`/private/docker/containers/${id}/start`)
}

export function stopContainer(id: string) {
  return req.post(`/private/docker/containers/${id}/stop`)
}

export function restartContainer(id: string) {
  return req.post(`/private/docker/containers/${id}/restart`)
}

export function removeContainer(id: string, force = false) {
  return req.delete(`/private/docker/containers/${id}`, {
    params: { force },
  })
}

export function getContainerLogs(id: string, tail = '100') {
  return req.get<unknown, { data: { logs: string } }>(`/private/docker/containers/${id}/logs`, {
    params: { tail },
  })
}

export interface ImageInfo {
  id: string
  short_id: string
  repo_tags: string[]
  repo_digests: string[]
  size: number
  size_text: string
  created: number
  containers: number
  labels: Record<string, string>
}

export interface RegistryConfig {
  name: string
  server_address: string
  username: string
  password?: string
}

export interface VolumeInfo {
  name: string
  driver: string
  mountpoint: string
  created_at: string
  scope: string
  labels: Record<string, string>
  options: Record<string, string>
  size: number
  size_text: string
}

export interface PruneResult {
  deleted: number
  space_reclaimed: number
  space_text: string
}

export interface CreateVolumeRequest {
  name: string
  driver: string
  labels?: Record<string, string>
  options?: Record<string, string>
}

export function listImages() {
  return req.get<unknown, { data: ImageInfo[] }>('/private/docker/images')
}

export function pullImage(image: string, registry?: string) {
  return req.post('/private/docker/images/pull', { image, registry }, { timeout: 120000 })
}

export function removeImage(id: string, force = false) {
  return req.delete(`/private/docker/images/${encodeURIComponent(id)}`, { params: { force } })
}

export function tagImage(source: string, target: string) {
  return req.post('/private/docker/images/tag', { source, target })
}

export function pruneImages() {
  return req.post<unknown, { data: PruneResult }>('/private/docker/images/prune')
}

export function listRegistries() {
  return req.get<unknown, { data: RegistryConfig[] }>('/private/docker/registries')
}

export function saveRegistries(data: RegistryConfig[]) {
  return req.put('/private/docker/registries', data)
}

export function testRegistry(data: RegistryConfig) {
  return req.post('/private/docker/registries/test', data)
}

export function listVolumes() {
  return req.get<unknown, { data: VolumeInfo[] }>('/private/docker/volumes')
}

export function createVolume(data: CreateVolumeRequest) {
  return req.post('/private/docker/volumes', data)
}

export function removeVolume(name: string, force = false) {
  return req.delete(`/private/docker/volumes/${encodeURIComponent(name)}`, { params: { force } })
}

export function pruneVolumes() {
  return req.post<unknown, { data: PruneResult }>('/private/docker/volumes/prune')
}

export interface DockerConnection {
  host: string
  tls_verify: boolean
  cert_path: string
}

export interface DockerPingResult {
  available: boolean
  custom_connection: boolean
  host: string
}

export function getDockerConnection() {
  return req.get<unknown, { data: DockerConnection }>('/private/setting/docker/connection')
}

export function saveDockerConnection(conn: DockerConnection) {
  return req.put('/private/setting/docker/connection', conn)
}

export function testDockerConnection(conn: DockerConnection) {
  return req.post('/private/setting/docker/connection/test', conn)
}

export interface NetworkInfo {
  id: string
  name: string
  driver: string
  scope: string
  internal: boolean
  attachable: boolean
  ipam_driver: string
  subnets: string[]
  labels: Record<string, string>
  containers: NetworkContainer[]
}

export interface NetworkContainer {
  id: string
  name: string
  ipv4: string
  ipv6: string
}

export interface CreateNetworkRequest {
  name: string
  driver: string
  subnet: string
  gateway: string
  internal: boolean
  attachable: boolean
  labels?: Record<string, string>
}

export function listNetworks() {
  return req.get<unknown, { data: NetworkInfo[] }>('/private/docker/networks')
}

export function createNetwork(data: CreateNetworkRequest) {
  return req.post('/private/docker/networks', data)
}

export function removeNetwork(id: string) {
  return req.delete(`/private/docker/networks/${id}`)
}

export function inspectNetwork(id: string) {
  return req.get<unknown, { data: NetworkInfo }>(`/private/docker/networks/${id}`)
}

export function connectContainerToNetwork(networkId: string, containerId: string) {
  return req.post(`/private/docker/networks/${networkId}/connect`, { container_id: containerId })
}

export function disconnectContainerFromNetwork(networkId: string, containerId: string) {
  return req.post(`/private/docker/networks/${networkId}/disconnect`, { container_id: containerId })
}

export interface ComposeProject {
  name: string
  status: string
  running: number
  stopped: number
  services: string[]
  networks: string[]
}

export function checkComposeAvailable() {
  return req.get<unknown, { data: { available: boolean } }>('/private/docker/compose/available')
}

export function listComposeProjects() {
  return req.get<unknown, { data: ComposeProject[] }>('/private/docker/compose/projects')
}

export function composeUp(content: string, projectName?: string) {
  return req.post('/private/docker/compose/up', { content, project_name: projectName })
}

export function composeDown(name: string, removeVolumes = false) {
  return req.post(`/private/docker/compose/${name}/down`, { remove_volumes: removeVolumes })
}

export function composeRestart(name: string) {
  return req.post(`/private/docker/compose/${name}/restart`)
}

export function composeStop(name: string) {
  return req.post(`/private/docker/compose/${name}/stop`)
}

export function composeStart(name: string) {
  return req.post(`/private/docker/compose/${name}/start`)
}

export function composeLogs(name: string, tail = '100') {
  return req.get<unknown, { data: { logs: string } }>(`/private/docker/compose/${name}/logs`, {
    params: { tail },
  })
}

export function getComposeConfig(name: string) {
  return req.get<unknown, { data: { content: string } }>(`/private/docker/compose/${name}/config`)
}
