<script setup lang="ts">
import type { FormInstance } from 'antdv-next'
import { LockOutlined, MailOutlined, UserOutlined } from '@antdv-next/icons'
import { message } from 'antdv-next'
import { reactive, ref, shallowRef, watch } from 'vue'
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '@/api/user'
import { getStatus } from '@/api/server'

const router = useRouter()
const formRef = shallowRef<FormInstance>()
const loading = ref(false)

onMounted(async () => {
  try {
    const res: any = await getStatus()
    if (res.data.installed) {
      router.replace('/login')
    }
  } catch {
    // 接口失败则允许继续访问安装页
  }
})

const model = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})

// 密码变更时重新校验确认密码
watch(
  () => model.password,
  () => {
    if (model.confirmPassword) {
      formRef.value?.validateFields?.(['confirmPassword'])
    }
  },
)

const confirmPasswordRules = [
  { required: true, whitespace: true, message: '请再次输入密码' },
  {
    validator: async (_rule: any, value: string) => {
      if (value && value !== model.password) {
        return Promise.reject(new Error('两次输入的密码不一致'))
      }
    },
  },
]

async function handleFinish() {
  loading.value = true
  try {
    const { confirmPassword, ...data } = model
    await register(data)
    message.success('安装完成，正在跳转...')
    router.push('/login')
  } catch {
    // 错误已由 req 拦截器统一处理
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="setup-container">
    <a-card class="setup-card" title="LynxPilot 初始设置" variant="borderless">
      <p class="setup-desc">创建管理员账号以开始使用面板</p>
      <a-form
        ref="formRef"
        name="setup"
        :model="model"
        layout="vertical"
        style="max-width: 400px"
        @finish="handleFinish"
      >
        <a-form-item name="username" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model:value="model.username" placeholder="用户名" size="large">
            <template #prefix>
              <UserOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          name="email"
          :rules="[
            { required: true, message: '请输入邮箱' },
            { type: 'email', message: '邮箱格式不正确' },
          ]"
        >
          <a-input v-model:value="model.email" placeholder="邮箱" size="large">
            <template #prefix>
              <MailOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[{ required: true, whitespace: true, message: '请输入密码' }]"
          has-feedback
          validate-trigger="blur"
        >
          <a-input-password v-model:value="model.password" placeholder="密码" size="large">
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item
          name="confirmPassword"
          :rules="confirmPasswordRules"
          has-feedback
          validate-trigger="blur"
        >
          <a-input-password
            v-model:value="model.confirmPassword"
            placeholder="确认密码"
            size="large"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button block type="primary" html-type="submit" size="large" :loading="loading">
            完成安装
          </a-button>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<style scoped>
.setup-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #f0f2f5;
}

.setup-card {
  width: 480px;
}

.setup-desc {
  color: #666;
  margin-bottom: 24px;
}
</style>
