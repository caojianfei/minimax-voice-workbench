<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Search, Star, Library, User, Heart, X, Check } from 'lucide-vue-next'
import { useFavorites } from '../composables/useFavorites'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  voices: {
    type: Array,
    default: () => []
  },
  height: {
    type: String,
    default: '400px'
  }
})

const emit = defineEmits(['update:modelValue', 'select', 'favorite'])
const { t } = useI18n()
const { toggleFavorite, isFavorite } = useFavorites()

// State
const activeTab = ref('all')
const searchQuery = ref('')
const debouncedQuery = ref('')
const searchInput = ref(null)
const scrollContainer = ref(null)

// Virtual Scroll State
const itemHeight = ref(80) // px
const buffer = 5
const scrollTop = ref(0)
const containerHeight = ref(400) // Default, will be measured

// Tabs
const tabs = [
  { key: 'all', label: '音色库', icon: Library },
  { key: 'custom', label: '我的音色', icon: User },
  { key: 'favorites', label: '收藏音色', icon: Heart }
]

// Debounce Search
let debounceTimer = null
watch(searchQuery, (newVal) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    debouncedQuery.value = newVal
  }, 300)
})

// Filter Voices
const filteredVoices = computed(() => {
  let result = props.voices

  if (activeTab.value === 'favorites') {
    result = result.filter(v => isFavorite(v.voice_id))
  } else if (activeTab.value === 'custom') {
    result = result.filter(v => v.type === 'cloned' || v.type === 'generated')
  }

  // 2. Search Filter
  if (debouncedQuery.value) {
    const q = debouncedQuery.value.toLowerCase()
    result = result.filter(v => v.name.toLowerCase().includes(q))
  }

  return result
})

// Virtual Scroll Logic
const visibleItems = computed(() => {
  const start = Math.floor(scrollTop.value / itemHeight.value)
  const visibleCount = Math.ceil(containerHeight.value / itemHeight.value)
  const total = filteredVoices.value.length
  
  const startIndex = Math.max(0, start - buffer)
  const endIndex = Math.min(total, start + visibleCount + buffer)
  
  return filteredVoices.value.slice(startIndex, endIndex).map((voice, index) => ({
    ...voice,
    virtualIndex: startIndex + index,
    top: (startIndex + index) * itemHeight.value
  }))
})

const totalHeight = computed(() => filteredVoices.value.length * itemHeight.value)

const onScroll = (e) => {
  scrollTop.value = e.target.scrollTop
}

// Interactions
const handleSelect = (voice) => {
  emit('update:modelValue', voice.voice_id)
  emit('select', voice)
}

const handleFavorite = (e, voice) => {
  e.stopPropagation()
  toggleFavorite(voice.voice_id)
  emit('favorite', voice)
}

const typeLabel = (voice) => {
  if (voice.type === 'system') return '系统音色'
  if (voice.type === 'cloned') return '复刻音色'
  if (voice.type === 'generated') return '设计音色'
  return (voice.type || '').toString()
}

const descriptionText = (voice) => {
  const candidates = [voice.description, voice.prompt, voice.voice_id].filter(Boolean)
  return candidates[0] || ''
}

const avatarText = (voice) => {
  const name = (voice.name || '').toString().trim()
  return name ? name.charAt(0).toUpperCase() : '?'
}

// Update container height on mount/resize
let mql = null
let onResize = null
let onMqlChange = null

const syncLayout = () => {
  const mobile = mql ? mql.matches : window.innerWidth <= 640
  itemHeight.value = mobile ? 56 : 80
  if (scrollContainer.value) {
    containerHeight.value = scrollContainer.value.clientHeight
  }
}

onMounted(() => {
  mql = window.matchMedia('(max-width: 640px)')
  onMqlChange = () => syncLayout()
  if (mql.addEventListener) mql.addEventListener('change', onMqlChange)
  else mql.addListener(onMqlChange)

  onResize = () => {
    if (scrollContainer.value) {
      containerHeight.value = scrollContainer.value.clientHeight
    }
  }
  window.addEventListener('resize', onResize)

  syncLayout()
})

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
  if (onResize) window.removeEventListener('resize', onResize)
  if (mql && onMqlChange) {
    if (mql.removeEventListener) mql.removeEventListener('change', onMqlChange)
    else mql.removeListener(onMqlChange)
  }
})

// Highlight matching text
const escapeRegExp = (string) => {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

const highlightText = (text) => {
  if (!debouncedQuery.value) return text
  const safeQuery = escapeRegExp(debouncedQuery.value)
  const regex = new RegExp(`(${safeQuery})`, 'gi')
  return text.replace(regex, '<span class="highlight">$1</span>')
}

// Expose internal methods if needed
defineExpose({
  clearSearch: () => {
    searchQuery.value = ''
    debouncedQuery.value = ''
  }
})
</script>

<template>
  <div class="voice-selector" :style="{ height }">
    <!-- Tabs -->
    <div class="tabs">
      <button 
        v-for="tab in tabs" 
        :key="tab.key"
        @click="activeTab = tab.key"
        class="tab-btn"
        :class="{ 'is-active': activeTab === tab.key }"
      >
        <component :is="tab.icon" class="tab-icon" />
        <span class="tab-label">{{ tab.label }}</span>
        <div v-if="activeTab === tab.key" class="tab-indicator"></div>
      </button>
    </div>

    <!-- Search -->
    <div class="search">
      <div class="search-row">
        <div class="search-wrap">
          <Search class="search-icon" />
          <input 
            ref="searchInput"
            v-model="searchQuery"
            type="text"
            :placeholder="t('voices.searchPlaceholder')"
            class="search-input"
          />
          <button 
            v-if="searchQuery"
            @click="searchQuery = ''"
            class="clear-btn"
            type="button"
          >
            <X class="clear-icon" />
          </button>
        </div>
      </div>
    </div>

    <!-- List -->
    <div 
      ref="scrollContainer"
      class="list"
      @scroll="onScroll"
    >
      <div v-if="filteredVoices.length > 0" :style="{ height: totalHeight + 'px' }" class="list-inner">
        <div 
          v-for="voice in visibleItems" 
          :key="voice.voice_id"
          class="voice-row"
          :class="{ 'is-selected': modelValue === voice.voice_id }"
          :style="{ height: itemHeight + 'px', top: voice.top + 'px' }"
          @click="handleSelect(voice)"
        >
          <div class="avatar" :class="`avatar-${voice.type}`">
            {{ avatarText(voice) }}
          </div>

          <div class="content">
            <div class="name" v-html="highlightText(voice.name)"></div>
            <div class="desc">{{ descriptionText(voice) }}</div>
            <div class="meta">
              <span class="tag">{{ typeLabel(voice) }}</span>
            </div>
          </div>

          <div class="actions">
            <button
              class="select-btn"
              type="button"
              :class="{ selected: modelValue === voice.voice_id }"
              @click="(e) => { e.stopPropagation(); handleSelect(voice) }"
            >
              <Check v-if="modelValue === voice.voice_id" class="select-icon" />
              {{ modelValue === voice.voice_id ? '已选择' : '选择' }}
            </button>

            <button 
              class="icon-btn"
              type="button"
              @click="(e) => handleFavorite(e, voice)"
              :aria-label="isFavorite(voice.voice_id) ? 'Unfavorite' : 'Favorite'"
            >
              <Star class="fav-star" :class="{ active: isFavorite(voice.voice_id) }" />
            </button>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="empty">
        <div class="empty-icon-wrap">
          <Search v-if="searchQuery" class="empty-icon" />
          <component :is="tabs.find(t => t.key === activeTab).icon" v-else class="empty-icon" />
        </div>
        <p class="empty-title">
          {{ searchQuery ? 'No voices found' : 'No voices in this category' }}
        </p>
        <p v-if="activeTab === 'favorites' && !searchQuery" class="empty-sub">
          Mark voices as favorite to see them here
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.voice-selector {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background: var(--bg-primary);
}

.tabs {
  display: flex;
  gap: var(--space-4);
  padding: 0 var(--space-2);
  border-bottom: 1px solid var(--border-color);
  background: transparent;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-4) var(--space-2);
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-secondary);
  background: transparent;
  position: relative;
  transition: color var(--transition-fast);
}

.tab-btn:hover {
  color: var(--text-primary);
}

.tab-btn.is-active {
  color: var(--primary);
}

.tab-icon {
  width: 16px;
  height: 16px;
}

.tab-label {
  display: inline;
}

.tab-indicator {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  background: var(--primary);
}

.search {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-primary);
}

.search-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.search-wrap {
  position: relative;
  flex: 1;
}

.search-icon {
  position: absolute;
  left: var(--space-3);
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: var(--text-tertiary);
  pointer-events: none;
}

.search-wrap input.search-input {
  padding: var(--space-2) calc(var(--space-3) + 16px + var(--space-2)) var(--space-2) calc(var(--space-3) + 16px + var(--space-2));
  padding-right: calc(var(--space-3) + 16px + var(--space-3));
  background: var(--bg-secondary);
}

.clear-btn {
  position: absolute;
  right: 6px;
  top: 50%;
  transform: translateY(-50%);
  width: 28px;
  height: 28px;
  border-radius: var(--radius-full);
  background: transparent;
  color: var(--text-tertiary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.clear-btn:hover {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.clear-icon {
  width: 16px;
  height: 16px;
}

.list {
  flex: 1;
  overflow-y: auto;
  position: relative;
  background: var(--bg-primary);
}

.list-inner {
  position: relative;
  width: 100%;
}

.voice-row {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-2) var(--space-4);
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  user-select: none;
  background: transparent;
  transition: background var(--transition-fast);
}

.voice-row:hover {
  background: var(--bg-secondary);
}

.voice-row.is-selected {
  background: var(--primary-bg);
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  color: var(--text-primary);
  background: var(--bg-tertiary);
  flex: 0 0 auto;
  border: 1px solid var(--border-color);
}

.avatar-system {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.25);
}

.avatar-cloned {
  background: rgba(6, 182, 212, 0.12);
  border-color: rgba(6, 182, 212, 0.25);
}

.avatar-generated {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.25);
}

.content {
  flex: 1;
  min-width: 0;
}

.name {
  font-weight: 700;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.25;
}

.desc {
  margin-top: 2px;
  font-size: 0.8rem;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.25;
}

.meta {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.tag {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 10px;
  border-radius: var(--radius-full);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 0.75rem;
  font-weight: 800;
  flex: 0 0 auto;
}

.actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex: 0 0 auto;
}

.select-btn {
  height: 32px;
  padding: 0 14px;
  border-radius: var(--radius-full);
  background: #0f172a;
  color: #ffffff;
  font-weight: 800;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: transform var(--transition-fast), background var(--transition-fast), opacity var(--transition-fast);
}

[data-theme='dark'] .select-btn {
  background: #0b1220;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.select-btn:hover {
  transform: translateY(-1px);
}

.select-btn.selected {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.select-icon {
  width: 14px;
  height: 14px;
}

.icon-btn {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background: transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast);
  border: 1px solid transparent;
  color: var(--text-tertiary);
}

.icon-btn:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-color);
}

.fav-star {
  width: 16px;
  height: 16px;
}

.fav-star {
  fill: transparent;
  transition: transform var(--transition-fast), color var(--transition-fast), fill var(--transition-fast);
}

.fav-star.active {
  color: var(--warning);
  fill: var(--warning);
  transform: scale(1.1);
}

.empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  text-align: center;
}

.empty-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  background: var(--bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--space-3);
  color: var(--text-tertiary);
}

.empty-icon {
  width: 24px;
  height: 24px;
}

.empty-title {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-secondary);
}

.empty-sub {
  margin: var(--space-1) 0 0 0;
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

:deep(.highlight) {
  background-color: rgba(253, 224, 71, 0.4);
  border-radius: 2px;
  padding: 0 1px;
}

@media (max-width: 640px) {
  .tabs {
    gap: var(--space-2);
    padding: 0 var(--space-1);
  }

  .tab-btn {
    padding: var(--space-3) var(--space-2);
  }

  .tab-label {
    display: none;
  }

  .search {
    padding: var(--space-3);
  }

  .voice-row {
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
  }

  .avatar,
  .desc,
  .meta,
  .select-btn {
    display: none;
  }
}
</style>
