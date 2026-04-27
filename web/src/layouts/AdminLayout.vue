<script setup lang="ts">
import { LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from '@antdv-next/icons'
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const collapsed = ref(false)

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
