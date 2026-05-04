<script setup lang="ts">
import type { ContainerDefaults } from '@/api/docker'

const model = defineModel<ContainerDefaults>({ required: true })

const props = defineProps<{
  saving: boolean
}>()

const emit = defineEmits<{
  save: []
}>()

const restartPolicyOptions = [
  { label: 'no', value: 'no' },
  { label: 'always', value: 'always' },
  { label: 'unless-stopped', value: 'unless-stopped' },
  { label: 'on-failure', value: 'on-failure' },
]
</script>

<template>
  <a-card class="docker-settings-card" title="容器默认值" variant="borderless">
    <p class="docker-settings-card-desc">
      这里的默认值会用于可视化 Compose 配置生成。
    </p>

    <a-form layout="vertical">
      <a-row :gutter="12">
        <a-col :xs="24" :md="12">
          <a-form-item label="重启策略">
            <a-select v-model:value="model.restart_policy" :options="restartPolicyOptions" />
          </a-form-item>
        </a-col>
        <a-col :xs="24" :md="12">
          <a-form-item label="日志驱动">
            <a-input v-model:value="model.log_driver" placeholder="例如: json-file" />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="12">
        <a-col :xs="24" :md="12">
          <a-form-item label="日志最大大小">
            <a-input v-model:value="model.log_max_size" placeholder="例如: 10m" />
          </a-form-item>
        </a-col>
        <a-col :xs="24" :md="12">
          <a-form-item label="日志保留文件数">
            <a-input-number v-model:value="model.log_max_file" :min="1" :step="1" style="width: 100%" />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="12">
        <a-col :xs="24" :md="12">
          <a-form-item label="CPU 限制">
            <a-input v-model:value="model.cpu_limit" placeholder="例如: 0.5 或 500m" />
          </a-form-item>
        </a-col>
        <a-col :xs="24" :md="12">
          <a-form-item label="内存限制">
            <a-input v-model:value="model.memory_limit" placeholder="例如: 512m" />
          </a-form-item>
        </a-col>
      </a-row>

      <div class="docker-settings-actions">
        <a-button type="primary" :loading="props.saving" @click="emit('save')">
          保存默认值
        </a-button>
      </div>
    </a-form>
  </a-card>
</template>

<style scoped>
.docker-settings-card { height: 100% }
.docker-settings-card-desc { margin: 0 0 16px; color: rgba(0,0,0,.45); font-size: 13px; line-height: 1.6 }
.docker-settings-actions { display: flex; justify-content: flex-end; margin-top: 4px }
</style>
