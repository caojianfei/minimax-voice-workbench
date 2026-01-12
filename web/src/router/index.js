import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        redirect: '/workbench'
      },
      {
        path: 'workbench',
        name: 'Workbench',
        component: () => import('../views/Workbench.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'audio-management',
        name: 'AudioManagement',
        component: () => import('../views/AudioManagement.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'voices',
        name: 'Voices',
        component: () => import('../views/Voices.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'keys',
        name: 'Keys',
        component: () => import('../views/Keys.vue'),
        meta: { requiresAuth: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

let authEnabledCache = { value: null, ts: 0 }

// normalizeRedirect 将 redirect 参数规范为站内路径
function normalizeRedirect(redirect) {
  if (typeof redirect !== 'string') return '/'
  if (!redirect.startsWith('/')) return '/'
  return redirect
}

async function checkAuthStatus() {
  const now = Date.now()
  if (authEnabledCache.value !== null && now - authEnabledCache.ts < 2000) {
    return authEnabledCache.value
  }

  try {
    const res = await fetch('/api/config', { headers: { 'Accept': 'application/json' } })
    if (!res.ok) throw new Error(`config status ${res.status}`)
    const data = await res.json()
    authEnabledCache = { value: Boolean(data?.auth_enabled), ts: now, captcha: !!data?.captcha_enabled, siteKey: data?.captcha_sitekey || '' }
    return authEnabledCache.value
  } catch {
    return true
  }
}

router.beforeEach(async (to, from, next) => {
  // Check if authentication is enabled globally
  const authEnabled = await checkAuthStatus();
  
  // Login page rules:
  // - If auth disabled, skip login and enter system
  // - If already logged in, skip login and enter system
  if (to.name === 'Login') {
    if (!authEnabled) {
      next({ path: '/' })
      return
    }

    const token = localStorage.getItem('token')
    if (token) {
      const redirect = normalizeRedirect(to.query.redirect)
      next({ path: redirect })
      return
    }

    next()
    return
  }

  if (!authEnabled) {
    next();
    return;
  }

  const token = localStorage.getItem('token');
  
  if (to.meta.requiresAuth && !token) {
    next({ name: 'Login', query: { redirect: to.fullPath } });
  } else {
    next();
  }
});

export default router
