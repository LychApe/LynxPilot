<script setup lang="ts">
import type { ContainerUIPrefs } from '@/api/docker'

const model = defineModel<ContainerUIPrefs>({ required: true })

const props = defineProps<{
  saving: boolean
}>()

const emit = defineEmits<{
  save: []
}>()
</script>

<template>
  <a-card class="docker-settings-card" title="界面偏好" variant="borderless">
    <p class="docker-settings-card-desc">
      控制容器管理页的默认显示和自动刷新行为。
    </p>

    <a-form layout="vertical">
      <a-row :gutter="12">
        <a-col :xs="24" :md="12">
          <a-form-item label="自动刷新间隔（秒）">
            <a-input-number v-model:value="model.auto_refresh_interval" :min="0" :step="1" style="width: 100%" />
            <div class="docker-settings-help">
              设为 0 可关闭自动刷新。
            </div>
          </a-form-item>
        </a-col>
        <a-col :xs="24" :md="12">
          <a-form-item label="默认显示已停止容器">
            <a-switch v-model:checked="model.show_stopped_default" />
            <div class="docker-settings-help">
              关闭后进入容器列表默认只显示运行中的容器。
            </div>
          </a-form-item>
        </a-col>
      </a-row>

      <div class="docker-settings-actions">
        <a-button type="primary" :loading="props.saving" @click="emit('save')">
          保存偏好
        </a-button>
      </div>
    </a-form>
  </a-card>
</template>

<style scoped>
.docker-settings-card { height: 100% }
.docker-settings-card-desc { margin: 0 0 16px; color: rgba(0,0,0,.45); font-size: 13px; line-height: 1.6 }
.docker-settings-help { margin-top: 6px; color: rgba(0,0,0,.35); font-size: 12px; line-height: 1.5 }
.docker-settings-actions { display: flex; justify-content: flex-end; margin-top: 4px }
</style>
