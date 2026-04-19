import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/install/setup',
    },
    {
      path: '/install/setup',
      name: 'setup',
      component: () => import('@/views/install/SetupView.vue'),
    },
  ],
})

export default router
