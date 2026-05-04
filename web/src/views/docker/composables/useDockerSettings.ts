import type { ContainerDefaults, ContainerUIPrefs, DockerConnection } from '@/api/docker'
import {
  getAllSettings,
  saveContainerDefaults,
  saveDockerConnection,
  saveUIPrefs,
  testDockerConnection,
} from '@/api/docker'
import { message } from 'antdv-next'
import { ref, shallowRef } from 'vue'

export function useDockerSettings() {
  const loading = shallowRef(false)
  const savingConnection = shallowRef(false)
  const savingDefaults = shallowRef(false)
  const savingPrefs = shallowRef(false)
  const connection = ref<DockerConnection>({ host: '', tls_verify: false, cert_path: '' })
  const containerDefaults = ref<ContainerDefaults>({
    restart_policy: '',
    log_driver: '',
    log_max_size: '',
    log_max_file: 3,
    cpu_limit: '',
    memory_limit: '',
  })
  const uiPrefs = ref<ContainerUIPrefs>({
    auto_refresh_interval: 10,
    show_stopped_default: true,
  })
  const testLoading = shallowRef(false)

  async function loadAll() {
    loading.value = true
    try {
      const res = await getAllSettings()
      const data = res.data
      if (data) {
        connection.value = {
          host: data.connection.host ?? '',
          tls_verify: data.connection.tls_verify ?? false,
          cert_path: data.connection.cert_path ?? '',
        }
        containerDefaults.value = {
          restart_policy: data.container_defaults.restart_policy ?? '',
          log_driver: data.container_defaults.log_driver ?? '',
          log_max_size: data.container_defaults.log_max_size ?? '',
          log_max_file: data.container_defaults.log_max_file ?? 3,
          cpu_limit: data.container_defaults.cpu_limit ?? '',
          memory_limit: data.container_defaults.memory_limit ?? '',
        }
        uiPrefs.value = {
          auto_refresh_interval: data.ui_prefs.auto_refresh_interval ?? 10,
          show_stopped_default: data.ui_prefs.show_stopped_default !== false,
        }
      }
      return true
    } catch {
      message.error('加载设置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  async function saveConnection() {
    savingConnection.value = true
    try {
      await saveDockerConnection(connection.value)
      message.success('Docker 连接配置已保存')
      return true
    } catch {
      message.error('保存 Docker 连接配置失败')
      return false
    } finally {
      savingConnection.value = false
    }
  }

  async function testConnection() {
    testLoading.value = true
    try {
      await testDockerConnection(connection.value)
      message.success('连接测试成功')
      return true
    } catch (err: unknown) {
      const serverMsg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      message.error(serverMsg ?? '连接测试失败')
      return false
    } finally {
      testLoading.value = false
    }
  }

  async function saveDefaults() {
    savingDefaults.value = true
    try {
      await saveContainerDefaults(containerDefaults.value)
      message.success('容器默认配置已保存')
      return true
    } catch {
      message.error('保存容器默认配置失败')
      return false
    } finally {
      savingDefaults.value = false
    }
  }

  async function savePrefs() {
    savingPrefs.value = true
    try {
      await saveUIPrefs(uiPrefs.value)
      message.success('界面偏好已保存')
      return true
    } catch {
      message.error('保存界面偏好失败')
      return false
    } finally {
      savingPrefs.value = false
    }
  }

  return {
    loading,
    savingConnection,
    savingDefaults,
    savingPrefs,
    connection,
    containerDefaults,
    uiPrefs,
    testLoading,
    loadAll,
    saveConnection,
    testConnection,
    saveDefaults,
    savePrefs,
  }
}
