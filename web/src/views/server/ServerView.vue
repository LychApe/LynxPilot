<script setup lang="ts">
import { PoweroffOutlined, ReloadOutlined } from '@antdv-next/icons'
import { ref } from 'vue'
import req from '@/utils/req'
import { message } from 'antdv-next'

const rebooting = ref(false)
const shuttingDown = ref(false)

async function handleReboot() {
  rebooting.value = true
  try {
    await req.get('/private/server/reboot')
    message.success('重启指令已发送')
  } catch {
    message.error('重启失败')
  } finally {
    rebooting.value = false
  }
}

async function handleShutdown() {
  shuttingDown.value = true
  try {
    await req.get('/private/server/shutdown')
    message.success('关机指令已发送')
  } catch {
    message.error('关机失败')
  } finally {
    shuttingDown.value = false
  }
}
</script>

<template>
  <div>
    <h2 style="margin: 0 0 24px; font-size: 20px; font-weight: 600">
      服务管理
    </h2>

    <a-row :gutter="16">
      <a-col :span="12">
        <a-card
          title="重启服务"
          variant="borderless"
        >
          <p>重启 LynxPilot 服务进程，当前连接将短暂中断。</p>
          <a-button
            type="primary"
            :loading="rebooting"
            @click="handleReboot"
          >
            <template #icon>
              <ReloadOutlined />
            </template>
            重启
          </a-button>
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card
          title="关闭服务"
          variant="borderless"
        >
          <p>完全停止 LynxPilot 服务，之后需要手动重新启动。</p>
          <a-button
            danger
            type="primary"
            :loading="shuttingDown"
            @click="handleShutdown"
          >
            <template #icon>
              <PoweroffOutlined />
            </template>
            关机
          </a-button>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>
