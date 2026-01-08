<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { Plus, Trash2, Play, Mic, Cloud, Palette, Monitor, Copy, Wand2, Pause } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const voices = ref([])
const keys = ref([])
const showModal = ref(false)
const modalMode = ref('clone') // 'clone' or 'design'
const loading = ref(false)
const currentTab = ref('system')
const playingAudio = ref(null)

const tabs = computed(() => [
  { key: 'system', label: '系统音色', icon: Monitor },
  { key: 'cloned', label: '复刻音色', icon: Copy },
  { key: 'generated', label: '设计音色', icon: Wand2 },
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

// Categorize voices by type
const categorizedVoices = computed(() => {
  const categories = {
    system: [],
    cloned: [],
    generated: []
  }
  
  voices.value.forEach(voice => {
    if (categories[voice.type]) {
      categories[voice.type].push(voice)
    }
  })
  
  return categories
})

const currentVoices = computed(() => {
  return categorizedVoices.value[currentTab.value] || []
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
              <h3>{{ voice.name }}</h3>
              <div class="voice-meta">
                <span class="voice-id" :title="voice.voice_id">{{ voice.voice_id.substring(0, 12) }}...</span>
              </div>
            </div>
            
            <button 
              v-if="voice.demo_audio || voice.preview"
              class="play-btn" 
              :class="{ playing: playingAudio === voice.voice_id }"
              @click="toggleAudio(voice)"
            >
              <component :is="playingAudio === voice.voice_id ? Pause : Play" size="20" fill="currentColor" />
            </button>
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
          <Mic size="48" />
        </div>
        <h3>No voices found</h3>
        <p>Create a new voice or sync from server</p>
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
                <div class="file-input-wrapper">
                  <input type="file" @change="handleFileChange" accept="audio/*" />
                </div>
                <p class="hint">{{ t('voices.hintSample') }}</p>
              </div>
              
              <div class="form-group">
                <label>{{ t('voices.labelPromptFile') }}</label>
                <div class="file-input-wrapper">
                   <input type="file" @change="handlePromptFileChange" accept="audio/*" />
                </div>
                <p class="hint">{{ t('voices.hintPromptFile') }}</p>
              </div>
              
              <div class="form-group">
                <label>{{ t('voices.labelPromptText') }}</label>
                <input v-model="form.prompt_text" type="text" :placeholder="t('voices.phPromptText')" />
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

              <div class="checkbox-grid">
                <label class="checkbox-label">
                  <input type="checkbox" v-model="form.noise_reduction" />
                  <span>{{ t('voices.labelNoiseReduction') }}</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" v-model="form.volume_normalization" />
                  <span>{{ t('voices.labelVolumeNorm') }}</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" v-model="form.watermark" />
                  <span>{{ t('voices.labelWatermark') }}</span>
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
  width: 550px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  padding: 0;
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

.checkbox-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: var(--space-3);
  margin-top: var(--space-2);
}

@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(99, 102, 241, 0.4); }
  70% { box-shadow: 0 0 0 10px rgba(99, 102, 241, 0); }
  100% { box-shadow: 0 0 0 0 rgba(99, 102, 241, 0); }
}

@media (max-width: 768px) {
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
