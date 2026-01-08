<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { Mic, Key, Disc, Activity, Languages, Library } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t, locale } = useI18n()

const navItems = [
  { key: 'workbench', path: '/workbench', icon: Mic },
  { key: 'audioManagement', path: '/audio-management', icon: Library },
  { key: 'voices', path: '/voices', icon: Disc },
  { key: 'keys', path: '/keys', icon: Key },
]

const toggleLang = () => {
  locale.value = locale.value === 'zh' ? 'en' : 'zh'
}
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar">
      <div class="brand">
        <Activity class="brand-icon" />
        <span>Minimax</span>
      </div>
      
      <nav class="nav-menu">
        <RouterLink 
          v-for="item in navItems" 
          :key="item.path" 
          :to="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
        >
          <component :is="item.icon" size="20" />
          <span>{{ t('nav.' + item.key) }}</span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <button @click="toggleLang" class="nav-item lang-btn">
          <Languages size="20" />
          <span>{{ locale === 'zh' ? 'English' : '中文' }}</span>
        </button>
      </div>
    </aside>

    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.sidebar {
  width: 240px;
  background-color: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.brand {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 var(--space-6);
  gap: var(--space-3);
  font-weight: 700;
  font-size: 1.25rem;
  border-bottom: 1px solid var(--border-color);
  color: var(--primary);
}

.nav-menu {
  flex: 1;
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.sidebar-footer {
  padding: var(--space-4);
  border-top: 1px solid var(--border-color);
}

.lang-btn {
  width: 100%;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 1rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s;
}

.nav-item:hover {
  background-color: var(--bg-tertiary);
  color: var(--text-primary);
}

.nav-item.router-link-active {
  background-color: var(--primary);
  color: white;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-6);
}
</style>
