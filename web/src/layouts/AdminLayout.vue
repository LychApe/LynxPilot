<script setup lang="ts">
import {
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  PoweroffOutlined,
  ReloadOutlined,
} from '@antdv-next/icons'
import { h, ref } from 'vue'
import { useRouter } from 'vue-router'
import req from '@/utils/req'
import { message } from 'antdv-next'

const router = useRouter()
const collapsed = ref(false)

const powerLoading = ref(false)

async function handlePowerMenu({ key }: { key: string }) {
  powerLoading.value = true
  try {
    if (key === 'reboot') {
      await req.get('/private/server/reboot')
      message.success('重启面板指令已发送')
    } else {
      await req.get('/private/server/shutdown')
      message.success('关闭面板指令已发送')
    }
  } catch {
    message.error(key === 'reboot' ? '重启失败' : '关机失败')
  } finally {
    powerLoading.value = false
  }
}

function handleLogout() {
  localStorage.removeItem('token')
  localStorage.removeItem('expires_at')
  router.push('/login')
}
</script>

<template>
  <a-layout class="admin-layout">
    <a-layout-sider
      v-model:collapsed="collapsed"
      collapsible
      :trigger="null"
      :width="220"
      class="admin-sider"
    >
      <div class="sider-logo">
        <img src="/LynxPilot.svg" alt="LynxPilot" class="logo-icon">
        <span v-if="!collapsed" class="logo-text">LynxPilot</span>
      </div>
      <SideNav />
    </a-layout-sider>

    <a-layout>
      <a-layout-header class="admin-header">
        <span
          class="trigger"
          @click="collapsed = !collapsed"
        >
          <MenuUnfoldOutlined v-if="collapsed" />
          <MenuFoldOutlined v-else />
        </span>

        <div class="header-actions">
          <a-dropdown
            :menu="{
              items: [
                { key: 'reboot', icon: () => h(ReloadOutlined), label: '重启面板' },
                { key: 'shutdown', icon: () => h(PoweroffOutlined), label: '关闭面板', danger: true },
              ],
              onClick: handlePowerMenu,
            }"
            :trigger="['click']"
          >
            <a-button
              type="text"
              class="power-btn"
              :loading="powerLoading"
              @click.prevent
            >
              <template #icon>
                <PoweroffOutlined />
              </template>
            </a-button>
          </a-dropdown>

          <a-button
            type="text"
            class="logout-btn"
            @click="handleLogout"
          >
            <template #icon>
              <LogoutOutlined />
            </template>
            退出
          </a-button>
        </div>
      </a-layout-header>

      <a-layout-content class="admin-content">
        <RouterView />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
}

.admin-sider {
  background: #fff;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.06);
  position: relative;
  z-index: 10;
}

.sider-logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: 1px solid #f0f0f0;
  overflow: hidden;
}

.logo-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  white-space: nowrap;
  color: #1677ff;
}

.admin-header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  position: relative;
  z-index: 9;
}

.trigger {
  font-size: 18px;
  cursor: pointer;
  padding: 0 8px;
  transition: color 0.2s;
}

.trigger:hover {
  color: #1677ff;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.power-btn {
  color: rgba(0, 0, 0, 0.65);
}

.power-btn:hover {
  color: #1677ff;
}

.logout-btn {
  color: rgba(0, 0, 0, 0.65);
}

.logout-btn:hover {
  color: #ff4d4f;
}

.admin-content {
  margin: 24px;
  padding: 24px;
  background: #fff;
  border-radius: 8px;
  min-height: 280px;
}
</style>
