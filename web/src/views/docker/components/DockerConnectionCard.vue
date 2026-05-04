<script setup lang="ts">
import type { DockerConnection } from '@/api/docker'

const model = defineModel<DockerConnection>({ required: true })

const props = defineProps<{
  saving: boolean
  testing: boolean
}>()

const emit = defineEmits<{
  save: []
  test: []
}>()

const hostPresets = [
  { label: '本地默认', host: '', tls: false },
  { label: 'Linux Socket', host: 'unix:///var/run/docker.sock', tls: false },
  { label: 'TCP 2375', host: 'tcp://10.1.0.3:2375', tls: false },
  { label: 'TCP 2376', host: 'tcp://10.1.0.3:2376', tls: true },
]

function applyPreset(host: string, tls: boolean) {
  model.value.host = host
  model.value.tls_verify = tls
  if (!tls) model.value.cert_path = ''
}
</script>

<template>
  <a-card class="docker-settings-card" title="Docker 连接" variant="borderless">
    <p class="docker-settings-card-desc">
      配置本机或远程 Docker API，留空 Host 会使用默认环境连接。
    </p>

    <a-form layout="vertical">
      <a-form-item label="Docker Host">
        <a-input
          v-model:value="model.host"
          placeholder="例如: unix:///var/run/docker.sock 或 tcp://10.1.0.3:2375"
        />
        <div class="docker-settings-help">
          常用格式：unix:///var/run/docker.sock、tcp://IP:2375、tcp://IP:2376
        </div>
      </a-form-item>

      <div class="docker-settings-preset-row">
        <button
          v-for="preset in hostPresets"
          :key="preset.label"
          type="button"
          class="docker-settings-preset"
          @click="applyPreset(preset.host, preset.tls)"
        >
          {{ preset.label }}
        </button>
      </div>

      <a-form-item label="TLS 验证">
        <a-switch v-model:checked="model.tls_verify" />
        <div class="docker-settings-help">
          开启后会在连接测试和 API 请求中使用 TLS 证书。
        </div>
      </a-form-item>

      <a-form-item v-if="model.tls_verify" label="证书目录">
        <a-input
          v-model:value="model.cert_path"
          placeholder="包含 ca.pem、cert.pem、key.pem 的目录路径"
        />
        <div class="docker-settings-help">
          端口 2376 通常需要完整证书目录。
        </div>
      </a-form-item>

      <div class="docker-settings-actions">
        <a-button :loading="props.testing" @click="emit('test')">
          测试连接
        </a-button>
        <a-button type="primary" :loading="props.saving" @click="emit('save')">
          保存连接
        </a-button>
      </div>
    </a-form>
  </a-card>
</template>

<style scoped>
.docker-settings-card { height: 100% }
.docker-settings-card-desc { margin: 0 0 16px; color: rgba(0,0,0,.45); font-size: 13px; line-height: 1.6 }
.docker-settings-help { margin-top: 6px; color: rgba(0,0,0,.35); font-size: 12px; line-height: 1.5 }
.docker-settings-preset-row { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 16px }
.docker-settings-preset { padding: 4px 10px; color: rgba(0,0,0,.65); font-size: 12px; cursor: pointer; background: rgba(0,0,0,.02); border: 1px solid rgba(5,5,5,.08); border-radius: 999px; transition: color .2s, border-color .2s, background .2s }
.docker-settings-preset:hover { color: #1677ff; background: rgba(22,119,255,.04); border-color: rgba(22,119,255,.3) }
.docker-settings-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 4px }
</style>
