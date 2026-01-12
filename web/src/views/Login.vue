<template>
  <div class="login-container">
    <div class="login-card">
      <div class="brand-section">
        <div class="logo-wrapper">
          <Activity class="brand-logo" :size="32" />
        </div>
        <h1 class="brand-title">Minimax Voice</h1>
        <p class="brand-subtitle">{{ $t('login.title') }}</p>
      </div>

      <form class="login-form" @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="username">{{ $t('login.username') }}</label>
          <div class="input-wrapper">
            <User class="input-icon" :size="18" />
            <input
              id="username"
              type="text"
              v-model="username"
              :placeholder="$t('login.username')"
              style="padding-left: 48px;"
              required
              :disabled="loading"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="password">{{ $t('login.password') }}</label>
          <div class="input-wrapper">
            <Lock class="input-icon" :size="18" />
            <input
              id="password"
              type="password"
              v-model="password"
              :placeholder="$t('login.password')"
              style="padding-left: 48px;"
              required
              :disabled="loading"
            />
          </div>
        </div>

        <div v-if="captchaEnabled" class="form-group">
          <div id="turnstile-container"></div>
        </div>

        <div v-if="error" class="error-message">
          <AlertCircle :size="16" />
          <span>{{ error }}</span>
        </div>

        <button type="submit" class="submit-btn" :disabled="loading || (captchaEnabled && !captchaReady)">
          <span v-if="loading" class="spinner"></span>
          <span v-else>{{ $t('login.submit') }}</span>
        </button>
      </form>
      
      <div class="footer-links">
        <!-- Optional: Add links like Forgot Password or Sign Up if needed later -->
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/client'
import { Activity, User, Lock, AlertCircle } from 'lucide-vue-next'

/* global turnstile */

const route = useRoute()
const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const captchaEnabled = ref(false)
const captchaSiteKey = ref('')
const captchaToken = ref('')
const captchaReady = ref(false)

// loadConfig fetches /api/config to determine whether Turnstile is required.
const loadConfig = async () => {
  try {
    const res = await fetch('/api/config', { headers: { 'Accept': 'application/json' } })
    if (!res.ok) return
    const data = await res.json()
    captchaEnabled.value = !!data.captcha_enabled && !!data.captcha_sitekey
    captchaSiteKey.value = data.captcha_sitekey || ''
  } catch {
    captchaEnabled.value = false
  }
}

// initTurnstile renders Cloudflare Turnstile and stores the generated token.
const initTurnstile = () => {
  if (!captchaEnabled.value || !captchaSiteKey.value) {
    captchaReady.value = true
    return
  }
  if (!window.turnstile) {
    window.addEventListener('load', () => initTurnstile(), { once: true })
    return
  }
  window.turnstile.render('#turnstile-container', {
    sitekey: captchaSiteKey.value,
    callback: (token) => {
      captchaToken.value = token
      captchaReady.value = true
    },
    'error-callback': () => {
      captchaToken.value = ''
      captchaReady.value = false
    },
    'expired-callback': () => {
      captchaToken.value = ''
      captchaReady.value = false
    }
  })
}

loadConfig().then(() => initTurnstile())

const handleLogin = async () => {
  if (captchaEnabled.value && !captchaToken.value) {
    error.value = 'Captcha token required'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const res = await api.post('/login', {
      username: username.value,
      password: password.value,
      turnstile_token: captchaToken.value || undefined
    })
    
    const { access_token, refresh_token } = res.data
    localStorage.setItem('token', access_token)
    localStorage.setItem('refresh_token', refresh_token)
    
    const redirect = typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/') ? route.query.redirect : '/'
    router.replace(redirect)
  } catch (e) {
    if (e.response && e.response.data && e.response.data.error) {
        error.value = e.response.data.error
    } else {
        error.value = 'Login failed'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f3f4f6; /* Light gray background */
  background-image: radial-gradient(#e5e7eb 1px, transparent 1px);
  background-size: 24px 24px;
  padding: 1rem;
}

[data-theme='dark'] .login-container {
  background-color: #111827;
  background-image: radial-gradient(#1f2937 1px, transparent 1px);
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  padding: 2.5rem;
  transition: transform 0.2s;
}

[data-theme='dark'] .login-card {
  background: #1f2937;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.5);
}

.brand-section {
  text-align: center;
  margin-bottom: 2rem;
}

.logo-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #6366f1, #818cf8);
  border-radius: 12px;
  margin-bottom: 1rem;
  color: white;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.brand-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #111827;
  margin-bottom: 0.5rem;
}

[data-theme='dark'] .brand-title {
  color: #f9fafb;
}

.brand-subtitle {
  color: #6b7280;
  font-size: 0.95rem;
}

[data-theme='dark'] .brand-subtitle {
  color: #9ca3af;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

[data-theme='dark'] label {
  color: #d1d5db;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
  pointer-events: none;
  z-index: 1;
}

.login-card .input-wrapper input {
  width: 100%;
  padding: 0.625rem 1rem 0.625rem 48px; /* Space for icon */
  border: 1px solid #d1d5db;
  border-radius: 8px;
  background-color: #fff;
  color: #111827;
  font-size: 0.95rem;
  box-sizing: border-box;
  transition: all 0.2s;
}

[data-theme='dark'] .login-card .input-wrapper input {
  background-color: #374151;
  border-color: #4b5563;
  color: #f9fafb;
}

.login-card .input-wrapper input:focus {
  outline: none;
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.error-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background-color: #fef2f2;
  border: 1px solid #fee2e2;
  border-radius: 8px;
  color: #ef4444;
  font-size: 0.875rem;
}

[data-theme='dark'] .error-message {
  background-color: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.2);
}

.submit-btn {
  margin-top: 0.5rem;
  width: 100%;
  padding: 0.75rem;
  background-color: #6366f1;
  color: white;
  font-weight: 600;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  transition: background-color 0.2s;
  display: flex;
  justify-content: center;
  align-items: center;
}

.submit-btn:hover:not(:disabled) {
  background-color: #4f46e5;
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
