<script setup lang="ts">
import { LockOutlined, UserOutlined } from '@antdv-next/icons'
import { message } from 'antdv-next'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '@/api/user'

const router = useRouter()
const loading = ref(false)

const model = reactive({
  username: '',
  password: '',
})

async function handleFinish() {
  loading.value = true
  try {
    const res: any = await login(model)
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('expires_at', res.data.expires_at)
    message.success('登录成功')
    router.push('/')
  } catch {
    // 错误已由 req 拦截器统一处理
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-flex class="login-container" justify="center" align="center">
    <a-card class="login-card" variant="borderless">
      <a-flex vertical align="center" :gap="4">
        <h1 class="login-title">
          LynxPilot
        </h1>
        <p class="login-subtitle">
          登录到管理面板
        </p>
      </a-flex>

      <a-divider />

      <a-form
        name="login"
        :model="model"
        layout="vertical"
        @finish="handleFinish"
      >
        <a-form-item
          name="username"
          :rules="[{ required: true, message: '请输入用户名' }]"
        >
          <a-input
            v-model:value="model.username"
            placeholder="请输入用户名…"
            size="large"
            autocomplete="username"
            :spellcheck="false"
          >
            <template #prefix>
              <UserOutlined aria-hidden="true" />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[{ required: true, message: '请输入密码' }]"
        >
          <a-input-password
            v-model:value="model.password"
            placeholder="请输入密码…"
            size="large"
            autocomplete="current-password"
          >
            <template #prefix>
              <LockOutlined aria-hidden="true" />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button
            block
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
          >
            登录
          </a-button>
        </a-form-item>
      </a-form>
    </a-card>
  </a-flex>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  padding: 24px;
  background: #f0f2f5;
}

.login-card {
  width: 400px;
}

.login-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.login-subtitle {
  margin: 0;
  color: #8c8c8c;
  font-size: 14px;
}
</style>
