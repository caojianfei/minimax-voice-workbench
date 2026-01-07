import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/workbench'
  },
  {
    path: '/workbench',
    name: 'Workbench',
    component: () => import('../views/Workbench.vue')
  },
  {
    path: '/voices',
    name: 'Voices',
    component: () => import('../views/Voices.vue')
  },
  {
    path: '/keys',
    name: 'Keys',
    component: () => import('../views/Keys.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
