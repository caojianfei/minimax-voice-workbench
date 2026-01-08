<script setup>
import { ref, watch, onMounted } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { Mic, Key, Disc, Activity, Languages, Library, Sun, Moon, Menu, X } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import Footer from './components/Footer.vue'

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

// Theme Logic
const theme = ref(localStorage.getItem('theme') || 'light')
const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
}

watch(theme, (newTheme) => {
  document.documentElement.setAttribute('data-theme', newTheme)
  localStorage.setItem('theme', newTheme)
}, { immediate: true })

// Mobile Sidebar Logic
const isSidebarOpen = ref(false)
const toggleSidebar = () => isSidebarOpen.value = !isSidebarOpen.value
const closeSidebar = () => isSidebarOpen.value = false

watch(() => route.path, () => {
  closeSidebar()
})
</script>

<template>
  <div class="app-layout">
    <!-- Mobile Header -->
    <header class="mobile-header">
      <button @click="toggleSidebar" class="icon-btn">
        <Menu />
      </button>
      <div class="brand-mobile">
        <Activity class="brand-icon" />
        <span>Minimax</span>
      </div>
      <div class="placeholder"></div>
    </header>

    <!-- Sidebar Overlay -->
    <div 
      class="sidebar-overlay" 
      :class="{ show: isSidebarOpen }"
      @click="closeSidebar"
    ></div>

    <aside class="sidebar" :class="{ open: isSidebarOpen }">
      <div class="brand">
        <div class="brand-logo">
          <Activity class="brand-icon" />
        </div>
        <span class="brand-text">Minimax Voice</span>
        <button @click="closeSidebar" class="close-btn md:hidden">
          <X size="20" />
        </button>
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
          <div v-if="route.path === item.path" class="active-indicator"></div>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <button @click="toggleTheme" class="nav-item theme-btn">
          <component :is="theme === 'light' ? Moon : Sun" size="20" />
          <span>{{ theme === 'light' ? 'Dark Mode' : 'Light Mode' }}</span>
        </button>
        <button @click="toggleLang" class="nav-item lang-btn">
          <Languages size="20" />
          <span>{{ locale === 'zh' ? 'English' : '中文' }}</span>
        </button>
      </div>
    </aside>

    <main class="main-content">
      <div class="content-wrapper">
        <RouterView />
      </div>
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background-color: var(--bg-secondary); /* Main bg is slightly darker/lighter than sidebar depending on theme */
  color: var(--text-primary);
}

/* Mobile Header */
.mobile-header {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: var(--header-height);
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  padding: 0 var(--space-4);
  align-items: center;
  justify-content: space-between;
  z-index: 40;
}

.brand-mobile {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-weight: 700;
  color: var(--text-primary);
}

.icon-btn {
  background: transparent;
  color: var(--text-secondary);
  padding: var(--space-2);
  border-radius: var(--radius-md);
}

.icon-btn:hover {
  background: var(--bg-tertiary);
}

.sidebar {
  width: var(--sidebar-width);
  background-color: var(--bg-primary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  z-index: 50;
  transition: transform var(--transition-normal);
}

.sidebar-overlay {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(2px);
  z-index: 45;
  opacity: 0;
  transition: opacity var(--transition-normal);
  pointer-events: none;
}

.sidebar-overlay.show {
  opacity: 1;
  pointer-events: auto;
}

.brand {
  height: var(--header-height);
  display: flex;
  align-items: center;
  padding: 0 var(--space-6);
  gap: var(--space-3);
  border-bottom: 1px solid var(--border-color);
}

.brand-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
  border-radius: var(--radius-md);
  color: white;
}

.brand-text {
  font-weight: 700;
  font-size: 1.125rem;
  color: var(--text-primary);
  letter-spacing: -0.025em;
}

.nav-menu {
  flex: 1;
  padding: var(--space-6) var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  overflow-y: auto;
}

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-decoration: none;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.nav-item:hover {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.nav-item.active {
  background-color: var(--primary-bg);
  color: var(--primary);
  font-weight: 600;
}

.active-indicator {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 16px;
  background-color: var(--primary);
  border-radius: 0 var(--radius-full) var(--radius-full) 0;
}

.sidebar-footer {
  padding: var(--space-6) var(--space-4);
  border-top: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  background-color: var(--bg-primary);
}

.theme-btn, .lang-btn {
  width: 100%;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  background: transparent;
  font-size: 0.9rem;
  font-family: inherit;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.theme-btn:hover, .lang-btn:hover {
  background-color: var(--bg-tertiary);
  color: var(--text-primary);
  border-color: var(--text-tertiary);
}

.main-content {
  flex: 1;
  margin-left: var(--sidebar-width);
  min-height: 100vh;
  padding: var(--space-6);
  transition: margin-left var(--transition-normal);
  max-width: 100%;
  overflow-x: hidden;
}

.content-wrapper {
  max-width: 1600px;
  margin: 0 auto;
  height: 100%;
}

/* Responsive */
@media (max-width: 1024px) {
  .sidebar {
    transform: translateX(-100%);
    box-shadow: var(--shadow-xl);
  }
  
  .sidebar.open {
    transform: translateX(0);
  }
  
  .sidebar-overlay {
    display: block;
  }
  
  .main-content {
    margin-left: 0;
    padding-top: calc(var(--header-height) + var(--space-6));
  }
  
  .mobile-header {
    display: flex;
  }

  .md\:hidden {
    display: block;
  }
}

.md\:hidden {
  display: none;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: var(--space-1);
  cursor: pointer;
  margin-left: auto;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background-color: var(--bg-tertiary);
  color: var(--text-primary);
}
</style>
