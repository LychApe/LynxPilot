<script setup lang="ts">
import {
  DashboardOutlined,
  CloudServerOutlined,
  SettingOutlined,
} from '@antdv-next/icons'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const selectedKeys = computed(() => [route.path])

const menuItems = [
  { key: '/dashboard', icon: DashboardOutlined, label: '概览' },
  { key: '/docker', icon: CloudServerOutlined, label: '容器管理' },
  { key: '/settings', icon: SettingOutlined, label: '系统设置' },
]

function handleMenuClick({ key }: { key: string }) {
  router.push(key)
}
</script>

<template>
  <a-menu
    mode="inline"
    :selected-keys="selectedKeys"
    :items="menuItems"
    @click="handleMenuClick"
  />
</template>
