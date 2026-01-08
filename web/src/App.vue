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
        <Activity class="brand-icon" />
        <span>Minimax Voice</span>
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
      <div class="footer-wrapper">
        <Footer />
      </div>
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

/* Mobile Header */
.mobile-header {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: var(--bg-secondary);
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
  color: var(--primary);
}

.icon-btn {
  background: transparent;
  color: var(--text-primary);
  padding: var(--space-2);
}

.sidebar {
  width: 260px;
  background-color: var(--bg-secondary);
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
  z-index: 45;
  opacity: 0;
  transition: opacity var(--transition-normal);
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
  justify-content: space-between;
}

.nav-menu {
  flex: 1;
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-decoration: none;
  transition: all var(--transition-fast);
  font-weight: 500;
}

.nav-item:hover {
  background-color: var(--bg-tertiary);
  color: var(--text-primary);
}

.nav-item.active {
  background-color: var(--primary-bg);
  color: var(--primary);
}

.sidebar-footer {
  padding: var(--space-4);
  border-top: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.theme-btn, .lang-btn {
  width: 100%;
  border: none;
  cursor: pointer;
  background: transparent;
  font-size: 0.95rem;
  font-family: inherit;
}

.main-content {
  flex: 1;
  margin-left: 260px; /* Sidebar width */
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  transition: margin-left var(--transition-normal);
}

.content-wrapper {
  flex: 1;
  padding: var(--space-6);
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.footer-wrapper {
  padding: 0 var(--space-6);
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

/* Responsive Design */
@media (max-width: 768px) {
  .mobile-header {
    display: flex;
  }

  .sidebar {
    transform: translateX(-100%);
  }

  .sidebar.open {
    transform: translateX(0);
  }

  .sidebar-overlay {
    display: block;
    pointer-events: none;
  }

  .sidebar-overlay.show {
    opacity: 1;
    pointer-events: auto;
  }

  .main-content {
    margin-left: 0;
    padding-top: 60px; /* Header height */
  }

  .content-wrapper {
    padding: var(--space-4);
  }
  
  .md\:hidden {
    display: block;
  }
  
  .close-btn {
    background: transparent;
    color: var(--text-secondary);
    padding: var(--space-1);
  }
}

@media (min-width: 769px) {
  .md\:hidden {
    display: none;
  }
}
</style>
