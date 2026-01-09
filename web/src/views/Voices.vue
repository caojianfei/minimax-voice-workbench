<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import axios from 'axios'
import { Plus, Trash2, Play, Mic, Cloud, Palette, Monitor, Copy, Wand2, Pause, Heart, Star, Search, X } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { useFavorites } from '../composables/useFavorites'

const { t } = useI18n()
const { toggleFavorite, isFavorite } = useFavorites()

const voices = ref([])
const keys = ref([])
const showModal = ref(false)
const modalMode = ref('clone') // 'clone' or 'design'
const loading = ref(false)
const currentTab = ref('system')
const playingAudio = ref(null)
const sampleFileInput = ref(null)
const promptFileInput = ref(null)
const searchQuery = ref('')
const debouncedQuery = ref('')

// Debounce Search
let debounceTimer = null
watch(searchQuery, (newVal) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    debouncedQuery.value = newVal
  }, 300)
})

const tabs = computed(() => [
  { key: 'system', label: '系统音色', icon: Monitor },
  { key: 'cloned', label: '复刻音色', icon: Copy },
  { key: 'generated', label: '设计音色', icon: Wand2 },
  { key: 'favorites', label: '我的收藏', icon: Heart },
])

// Speech model options
const modelOptions = [
  { value: 'speech-2.6-hd', label: 'Speech 2.6 HD' },
  { value: 'speech-2.6-turbo', label: 'Speech 2.6 Turbo' },
  { value: 'speech-02-hd', label: 'Speech 02 HD' },
  { value: 'speech-02-turbo', label: 'Speech 02 Turbo' },
  { value: 'speech-01-hd', label: 'Speech 01 HD' },
  { value: 'speech-01-turbo', label: 'Speech 01 Turbo' },
]

// Form Data
const form = ref({
  name: '',
  file: null,
  prompt_file: null,
  prompt_text: '',
  demo_text: '',
  model: 'speech-2.6-hd',
  noise_reduction: false,
  volume_normalization: false,
  watermark: false,
  // For design
  prompt: '',
  preview_text: 'Hello, this is a test voice.',
})

const api = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'
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

// Categorize voices by type
const categorizedVoices = computed(() => {
  const categories = {
    system: [],
    cloned: [],
    generated: [],
    favorites: []
  }
  
  voices.value.forEach(voice => {
    if (categories[voice.type]) {
      categories[voice.type].push(voice)
    }
    if (isFavorite(voice.voice_id)) {
      categories.favorites.push(voice)
    }
  })
  
  return categories
})

const currentVoices = computed(() => {
  let list = categorizedVoices.value[currentTab.value] || []
  
  if (debouncedQuery.value) {
    const q = debouncedQuery.value.toLowerCase()
    list = list.filter(v => v.name.toLowerCase().includes(q))
  }
  
  return list
})

const defaultKey = computed(() => {
  return keys.value.find(k => k.is_default) || keys.value[0]
})

const fetchData = async () => {
  try {
    const [vRes, kRes] = await Promise.all([
      api.get('/voices'),
      api.get('/keys')
    ])
    voices.value = vRes.data.data
    keys.value = kRes.data.data
  } catch (e) {
    console.error(e)
  }
}

const openModal = (mode) => {
  modalMode.value = mode
  showModal.value = true
}

const handleFileChange = (e) => {
  form.value.file = e.target.files[0]
}

const handlePromptFileChange = (e) => {
  form.value.prompt_file = e.target.files[0]
}

const submitForm = async () => {
  if (modalMode.value === 'clone') await cloneVoice()
  else await designVoice()
}

const cloneVoice = async () => {
  if (!form.value.name || !form.value.file) {
    alert(t('voices.alertFill'))
    return
  }

  loading.value = true
  const formData = new FormData()
  formData.append('name', form.value.name)
  formData.append('file', form.value.file)
  
  if (form.value.prompt_file) {
    formData.append('prompt_file', form.value.prompt_file)
  }
  if (form.value.prompt_text) {
    formData.append('prompt_text', form.value.prompt_text)
  }
  if (form.value.demo_text) {
    formData.append('demo_text', form.value.demo_text)
    formData.append('model', form.value.model)
  }
  formData.append('noise_reduction', form.value.noise_reduction)
  formData.append('volume_normalization', form.value.volume_normalization)
  formData.append('watermark', form.value.watermark)

  try {
    await api.post('/voices/clone', formData)
    cleanupModal()
  } catch (e) {
    alert(t('voices.alertCloneFail') + ': ' + (e.response?.data?.message || e.message))
    loading.value = false
  }
}

const designVoice = async () => {
  if (!form.value.prompt || !form.value.preview_text) {
    alert(t('voices.alertFill'))
    return
  }

  loading.value = true
  try {
    const payload = {
      name: form.value.name,
      prompt: form.value.prompt,
      preview_text: form.value.preview_text,
      watermark: form.value.watermark
    }
    await api.post('/voices/design', payload)
    cleanupModal()
  } catch (e) {
    alert(t('voices.alertDesignFail') + ': ' + (e.response?.data?.message || e.message))
    loading.value = false
  }
}

const cleanupModal = () => {
  showModal.value = false
  loading.value = false
  form.value.name = ''
  form.value.file = null
  form.value.prompt_file = null
  form.value.prompt_text = ''
  form.value.demo_text = ''
  form.value.prompt = ''
  form.value.noise_reduction = false
  form.value.volume_normalization = false
  form.value.watermark = false
  if (sampleFileInput.value) sampleFileInput.value.value = ''
  if (promptFileInput.value) promptFileInput.value.value = ''
  fetchData()
}

const deleteVoice = async (voice) => {
  if (!confirm(t('voices.confirmDelete'))) return
  
  try {
    await api.delete(`/voices/${voice.id}`)
    fetchData()
  } catch (e) {
    alert(t('voices.alertDeleteFail'))
  }
}

const syncVoices = async () => {
  loading.value = true
  try {
    const res = await api.post(`/voices/sync`)
    alert(t('voices.syncSuccess', { count: res.data.data.added }))
    fetchData()
  } catch (e) {
    alert(t('voices.alertSyncFail') + ': ' + (e.response?.data?.message || e.message))
  } finally {
    loading.value = false
  }
}

const toggleAudio = (voice) => {
  const audioId = 'audio-' + voice.voice_id
  const audioEl = document.getElementById(audioId)
  
  if (!audioEl) return

  if (playingAudio.value === voice.voice_id) {
    audioEl.pause()
    playingAudio.value = null
  } else {
    // Stop currently playing
    if (playingAudio.value) {
      const current = document.getElementById('audio-' + playingAudio.value)
      if (current) current.pause()
    }
    
    audioEl.currentTime = 0
    audioEl.play()
    playingAudio.value = voice.voice_id
    
    audioEl.onended = () => {
      playingAudio.value = null
    }
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <div class="header-content">
        <div>
          <h1>{{ t('voices.title') }}</h1>
          <p class="subtitle">{{ t('voices.subtitle') }}</p>
        </div>
        
        <div class="header-actions">
          <div v-if="defaultKey" class="key-badge">
             <span class="text-muted">{{ t('voices.currentKey') }}:</span>
             <code class="key-val">{{ defaultKey.remark || (defaultKey.key.substring(0,8) + '...') }}</code>
          </div>
          
          <button @click="syncVoices" :disabled="loading" class="btn btn-secondary">
            <Cloud size="18" /> {{ t('voices.sync') }}
          </button>
          
          <button @click="openModal('design')" class="btn btn-secondary">
            <Palette size="18" /> {{ t('voices.designNew') }}
          </button>
          
          <button @click="openModal('clone')" class="btn btn-primary">
            <Plus size="18" /> {{ t('voices.cloneNew') }}
          </button>
        </div>
      </div>
      
      <!-- Tabs Navigation -->
      <div class="tabs-wrapper">
        <nav class="tabs-nav">
          <button 
            v-for="tab in tabs" 
            :key="tab.key"
            class="tab-item"
            :class="{ active: currentTab === tab.key }"
            @click="currentTab = tab.key"
          >
            <component :is="tab.icon" size="18" />
            <span>{{ tab.label }}</span>
            <span class="badge badge-neutral ml-2">{{ categorizedVoices[tab.key]?.length || 0 }}</span>
          </button>
        </nav>
        
        <!-- Search Bar -->
        <div class="search-bar">
          <Search class="search-icon" size="16" />
          <input 
            v-model="searchQuery" 
            type="text" 
            :placeholder="t('voices.searchPlaceholder') || 'Search voices...'"
            class="search-input"
          />
          <button v-if="searchQuery" @click="searchQuery = ''" class="clear-btn">
            <X size="16" />
          </button>
        </div>
      </div>
    </header>

    <!-- Voice Grid -->
    <div class="voice-content">
      <div v-if="currentVoices.length > 0" class="voices-grid">
        <div v-for="voice in currentVoices" :key="voice.voice_id" class="voice-card card">
          <div class="voice-header">
            <div class="voice-avatar" :class="`avatar-${voice.type}`">
              {{ voice.name.charAt(0).toUpperCase() }}
            </div>
            <div class="voice-info">
              <h3 v-html="highlightText(voice.name)"></h3>
              <div class="voice-meta">
                <span class="voice-id" :title="voice.voice_id">{{ voice.voice_id.substring(0, 12) }}...</span>
              </div>
            </div>
            
            <div class="voice-actions-top">
              <button 
                class="favorite-btn" 
                :class="{ active: isFavorite(voice.voice_id) }"
                @click.stop="toggleFavorite(voice.voice_id)"
              >
                <Star size="18" :fill="isFavorite(voice.voice_id) ? 'currentColor' : 'none'" />
              </button>

              <button 
                v-if="voice.demo_audio || voice.preview"
                class="play-btn" 
                :class="{ playing: playingAudio === voice.voice_id }"
                @click="toggleAudio(voice)"
              >
                <component :is="playingAudio === voice.voice_id ? Pause : Play" size="20" fill="currentColor" />
              </button>
            </div>
            <audio 
              :id="'audio-' + voice.voice_id" 
              :src="voice.demo_audio || voice.preview" 
              style="display: none"
            ></audio>
          </div>
          
          <div class="voice-footer">
            <span class="badge" :class="`badge-${voice.type}`">{{ voice.type }}</span>
            <button 
              v-if="voice.type !== 'system'" 
              @click="deleteVoice(voice)" 
              class="btn-icon delete"
              title="Delete Voice"
            >
              <Trash2 size="16" />
            </button>
          </div>
        </div>
      </div>
      
      <div v-else class="empty-state">
        <div class="empty-icon">
          <Search v-if="searchQuery" size="48" />
          <Mic v-else size="48" />
        </div>
        <h3>{{ searchQuery ? 'No voices found' : 'No voices in this category' }}</h3>
        <p>{{ searchQuery ? 'Try a different search term' : 'Create a new voice or sync from server' }}</p>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay">
      <div class="modal card">
        <header class="modal-header">
          <h2>{{ modalMode === 'clone' ? t('voices.modalTitle') : t('voices.modalDesign') }}</h2>
          <button class="close-btn" @click="showModal = false">×</button>
        </header>
        
        <div class="modal-body">
          <!-- Common Fields -->
          <div class="form-group">
            <label>{{ t('voices.labelName') }}</label>
            <input v-model="form.name" type="text" :placeholder="t('voices.phName')" />
          </div>

          <!-- Clone Fields -->
          <template v-if="modalMode === 'clone'">
              <div class="form-group">
                <label>{{ t('voices.labelSample') }}</label>
                <div class="file-picker">
                  <input
                    id="clone-sample-file"
                    ref="sampleFileInput"
                    class="file-picker-input"
                    type="file"
                    @change="handleFileChange"
                    accept="audio/*"
                  />
                  <label class="file-picker-btn" for="clone-sample-file">{{ t('voices.selectFile') }}</label>
                  <span class="file-picker-name">{{ form.file?.name || t('voices.noFileSelected') }}</span>
                </div>
                <p class="hint">{{ t('voices.hintSample') }}</p>
              </div>

              <div class="pair-group">
                <div class="pair-header">
                  <div class="pair-title">{{ t('voices.pairPromptTitle') }}</div>
                  <div class="pair-subtitle">{{ t('voices.pairPromptSubtitle') }}</div>
                </div>

                <div class="pair-grid">
                  <div class="form-group">
                    <label>{{ t('voices.labelPromptFile') }}</label>
                    <div class="file-picker">
                      <input
                        id="clone-prompt-file"
                        ref="promptFileInput"
                        class="file-picker-input"
                        type="file"
                        @change="handlePromptFileChange"
                        accept="audio/*"
                      />
                      <label class="file-picker-btn" for="clone-prompt-file">{{ t('voices.selectFile') }}</label>
                      <span class="file-picker-name">{{ form.prompt_file?.name || t('voices.noFileSelected') }}</span>
                    </div>
                    <p class="hint">{{ t('voices.hintPromptFile') }}</p>
                  </div>

                  <div class="form-group">
                    <label>{{ t('voices.labelPromptText') }}</label>
                    <textarea
                      v-model="form.prompt_text"
                      class="pair-textarea"
                      :placeholder="t('voices.phPromptText')"
                      rows="2"
                    ></textarea>
                  </div>
                </div>
              </div>

              <div class="form-group">
                <label>{{ t('voices.labelDemoText') }}</label>
                <textarea v-model="form.demo_text" :placeholder="t('voices.phDemoText')" maxlength="1000"></textarea>
                <p class="hint">{{ t('voices.hintDemoText') }}</p>
              </div>

              <div class="form-group" v-if="form.demo_text">
                <label>{{ t('voices.labelModel') }}</label>
                <select v-model="form.model">
                  <option v-for="opt in modelOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <div class="options-grid">
                <label class="option-toggle">
                  <input type="checkbox" v-model="form.noise_reduction" />
                  <span class="option-text">{{ t('voices.labelNoiseReduction') }}</span>
                </label>
                <label class="option-toggle">
                  <input type="checkbox" v-model="form.volume_normalization" />
                  <span class="option-text">{{ t('voices.labelVolumeNorm') }}</span>
                </label>
                <label class="option-toggle">
                  <input type="checkbox" v-model="form.watermark" />
                  <span class="option-text">{{ t('voices.labelWatermark') }}</span>
                </label>
              </div>
          </template>

          <!-- Design Fields -->
          <template v-else>
              <div class="form-group">
                <label>{{ t('voices.labelPrompt') }}</label>
                <textarea v-model="form.prompt" :placeholder="t('voices.phPrompt')" rows="4"></textarea>
              </div>
              <div class="form-group">
                <label>{{ t('voices.labelPreview') }}</label>
                <input v-model="form.preview_text" type="text" :placeholder="t('voices.phPreview')" />
              </div>
          </template>
        </div>

        <div class="modal-footer">
          <button @click="showModal = false" class="btn btn-secondary">{{ t('voices.cancel') }}</button>
          <button @click="submitForm" :disabled="loading" class="btn btn-primary">
            {{ loading ? (modalMode==='clone'? t('voices.cloning'):t('voices.designing')) : (modalMode==='clone'? t('voices.startCloning'):t('voices.startDesigning')) }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.page-header {
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  padding: var(--space-6) var(--space-6) 0;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--space-6);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.key-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-tertiary);
  padding: 4px 12px;
  border-radius: var(--radius-full);
  font-size: 0.875rem;
  margin-right: var(--space-2);
}

.key-val {
  font-family: monospace;
  font-weight: 600;
  color: var(--primary);
}

.tabs-wrapper {
  margin-top: var(--space-2);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-bar {
  display: flex;
  align-items: center;
  position: relative;
  width: 240px;
  margin-right: 4px;
}

.search-bar input.search-input {
  width: 100%;
  padding: 8px 30px 8px 32px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  font-size: 0.875rem;
  transition: all 0.2s;
  height: 36px;
}

.search-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 2px var(--primary-light);
  outline: none;
}

.search-icon {
  position: absolute;
  left: 10px;
  color: var(--text-tertiary);
  pointer-events: none;
}

.clear-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  border-radius: 50%;
}

.clear-btn:hover {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.voice-actions-top {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.favorite-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: transparent;
  color: var(--text-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  border: none;
  cursor: pointer;
}

.favorite-btn:hover {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.favorite-btn.active {
  color: #fbbf24; /* Amber-400 */
}

/* Ensure highlight class works */
:deep(.highlight) {
  background-color: rgba(253, 224, 71, 0.4); /* Yellow-200 with opacity */
  color: inherit;
  border-radius: 2px;
  padding: 0 1px;
}

.tabs-nav {
  display: flex;
  gap: var(--space-6);
}

.tab-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) 0;
  background: transparent;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-weight: 500;
  transition: all var(--transition-fast);
}

.tab-item:hover {
  color: var(--text-primary);
}

.tab-item.active {
  color: var(--primary);
  border-bottom-color: var(--primary);
}

.voice-content {
  flex: 1;
  padding: var(--space-6);
  overflow-y: auto;
}

.voices-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--space-6);
}

.voice-card {
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  border: 1px solid var(--border-color);
  transition: all var(--transition-fast);
}

.voice-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  border-color: var(--primary-light);
}

.voice-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.voice-avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  font-weight: 700;
  color: white;
  flex-shrink: 0;
}

.avatar-system { background: linear-gradient(135deg, #3b82f6, #2563eb); }
.avatar-cloned { background: linear-gradient(135deg, #10b981, #059669); }
.avatar-generated { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }

.voice-info {
  flex: 1;
  min-width: 0;
}

.voice-info h3 {
  margin: 0;
  font-size: 1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.voice-meta {
  display: flex;
  gap: var(--space-2);
  margin-top: 2px;
}

.voice-id {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  font-family: monospace;
}

.play-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.play-btn:hover {
  background: var(--primary);
  color: white;
}

.play-btn.playing {
  background: var(--primary);
  color: white;
  animation: pulse 2s infinite;
}

.voice-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: var(--space-3);
  border-top: 1px solid var(--border-color);
}

.badge-system { background: var(--info-bg); color: var(--info); }
.badge-cloned { background: var(--success-bg); color: var(--success); }
.badge-generated { background: var(--primary-bg); color: var(--primary); }

.btn-icon.delete:hover {
  color: var(--error);
  background: var(--error-bg);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-10);
  color: var(--text-tertiary);
  text-align: center;
}

.empty-icon {
  margin-bottom: var(--space-4);
  color: var(--border-color);
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  backdrop-filter: blur(4px);
}

.modal {
  width: min(550px, calc(100vw - 2 * var(--space-4)));
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  padding: 0;
  box-shadow: var(--shadow-2xl);
  border: 1px solid var(--border-color);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-group label {
  margin-bottom: 0;
}

.hint {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-tertiary);
  line-height: 1.4;
}

.file-picker {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: 0.75rem 0.75rem;
  min-height: 56px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-primary);
}

.file-picker-input {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.file-picker-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-weight: 600;
  margin: 0;
  cursor: pointer;
  user-select: none;
  transition: all var(--transition-fast);
}

.file-picker-btn:hover {
  border-color: var(--text-tertiary);
  background: var(--bg-quaternary);
}

.file-picker-name {
  min-width: 0;
  flex: 1;
  color: var(--text-secondary);
  font-size: 0.9rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pair-group {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
  background: var(--bg-secondary);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.pair-header {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pair-title {
  font-weight: 700;
  color: var(--text-primary);
}

.pair-subtitle {
  font-size: 0.85rem;
  color: var(--text-tertiary);
  line-height: 1.4;
}

.pair-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
  min-width: 0;
}

.pair-grid .file-picker {
  height: 56px;
  padding: 0.5rem 0.75rem;
}

.pair-textarea {
  min-height: 56px;
  height: 56px;
  resize: none;
}

.modal-header {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
}

.close-btn {
  font-size: 1.5rem;
  color: var(--text-tertiary);
  background: transparent;
}

.modal-body {
  padding: var(--space-6);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.modal-footer {
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  background: var(--bg-secondary);
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: var(--space-3);
  margin-top: var(--space-2);
}

.option-toggle {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: 0.75rem 0.875rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  cursor: pointer;
  user-select: none;
  transition: all var(--transition-fast);
}

.option-toggle:hover {
  border-color: var(--text-tertiary);
  background: var(--bg-secondary);
}

.option-toggle input[type='checkbox'] {
  margin: 0;
  flex-shrink: 0;
}

.option-text {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.9rem;
}

@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(99, 102, 241, 0.4); }
  70% { box-shadow: 0 0 0 10px rgba(99, 102, 241, 0); }
  100% { box-shadow: 0 0 0 0 rgba(99, 102, 241, 0); }
}

@media (max-width: 768px) {
  .modal-footer {
    flex-direction: column-reverse;
    align-items: stretch;
  }

  .modal-footer .btn {
    width: 100%;
  }

  .pair-grid {
    grid-template-columns: 1fr;
  }

  .header-content {
    flex-direction: column;
    gap: var(--space-4);
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-start;
  }
  
  .tabs-nav {
    overflow-x: auto;
    padding-bottom: 0;
  }
  
  .tab-item {
    white-space: nowrap;
  }
}
</style>
