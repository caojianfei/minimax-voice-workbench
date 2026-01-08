<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { Plus, Trash2, Play, Mic, Cloud, Palette } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const voices = ref([])
const keys = ref([])
const showModal = ref(false)
const modalMode = ref('clone') // 'clone' or 'design'
const loading = ref(false)

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
    
    if (keys.value.length > 0 && !form.value.key_id) {
      // form.value.key_id = keys.value[0].id
    }
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

onMounted(fetchData)
</script>

<template>
  <div class="page page-voices">
    <header class="header">
      <div>
        <h1>{{ t('voices.title') }}</h1>
        <p class="subtitle">{{ t('voices.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <div v-if="defaultKey" class="current-key-display">
             <span class="text-muted">{{ t('voices.currentKey') }}:</span>
             <code class="key-val">{{ defaultKey.remark || (defaultKey.key.substring(0,8) + '...') }}</code>
        </div>
        <div v-else class="text-warning">
            {{ t('voices.noDefaultKey') }}
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
    </header>

    <!-- Categorized Voice Lists -->
    <div class="voice-categories">
      <!-- System Voices -->
      <div v-if="categorizedVoices.system.length > 0" class="category-section">
        <h2 class="category-title">系统音色 (System Voices)</h2>
        <div class="voices-grid">
          <div v-for="voice in categorizedVoices.system" :key="voice.id" class="card voice-card">
            <div class="voice-icon">
              <Mic size="24" />
            </div>
            <div class="voice-info">
              <h3>{{ voice.name }}</h3>
              <span class="voice-id">{{ voice.voice_id }}</span>
              <span class="badge badge-system">{{ voice.type }}</span>
            </div>
            <div class="actions">
              <button @click="deleteVoice(voice)" class="btn-icon delete">
                <Trash2 size="18" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Cloned Voices -->
      <div v-if="categorizedVoices.cloned.length > 0" class="category-section">
        <h2 class="category-title">复刻音色 (Cloned Voices)</h2>
        <div class="voices-grid">
          <div v-for="voice in categorizedVoices.cloned" :key="voice.id" class="card voice-card">
            <div class="voice-icon">
              <Mic size="24" />
            </div>
            <div class="voice-info">
              <h3>{{ voice.name }}</h3>
              <span class="voice-id">{{ voice.voice_id }}</span>
              <span class="badge badge-cloned">{{ voice.type }}</span>
            </div>
            <div class="actions">
              <audio v-if="voice.demo_audio" :src="voice.demo_audio" controls class="mini-player"></audio>
              <button @click="deleteVoice(voice)" class="btn-icon delete">
                <Trash2 size="18" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Generated Voices -->
      <div v-if="categorizedVoices.generated.length > 0" class="category-section">
        <h2 class="category-title">设计音色 (Generated Voices)</h2>
        <div class="voices-grid">
          <div v-for="voice in categorizedVoices.generated" :key="voice.id" class="card voice-card">
            <div class="voice-icon">
              <Mic size="24" />
            </div>
            <div class="voice-info">
              <h3>{{ voice.name }}</h3>
              <span class="voice-id">{{ voice.voice_id }}</span>
              <span class="badge badge-generated">{{ voice.type }}</span>
            </div>
            <div class="actions">
              <audio v-if="voice.preview" :src="voice.preview" controls class="mini-player"></audio>
              <button @click="deleteVoice(voice)" class="btn-icon delete">
                <Trash2 size="18" />
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <div v-if="voices.length === 0" class="empty-state">
        {{ t('voices.noVoices') }}
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay">
      <div class="modal card">
        <h2>{{ modalMode === 'clone' ? t('voices.modalTitle') : t('voices.modalDesign') }}</h2>
        
        <!-- Common Fields -->
        <div class="form-group">
          <label>{{ t('voices.labelName') }}</label>
          <input v-model="form.name" type="text" :placeholder="t('voices.phName')" />
        </div>

        <!-- Clone Fields -->
        <template v-if="modalMode === 'clone'">
            <div class="form-group">
              <label>{{ t('voices.labelSample') }}</label>
              <input type="file" @change="handleFileChange" accept="audio/*" />
              <p class="hint">{{ t('voices.hintSample') }}</p>
            </div>
            
            <div class="form-group">
              <label>{{ t('voices.labelPromptFile') }}</label>
              <input type="file" @change="handlePromptFileChange" accept="audio/*" />
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

            <div class="form-row">
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
              <textarea v-model="form.prompt" :placeholder="t('voices.phPrompt')"></textarea>
            </div>
            <div class="form-group">
              <label>{{ t('voices.labelPreview') }}</label>
              <input v-model="form.preview_text" type="text" :placeholder="t('voices.phPreview')" />
            </div>
        </template>

        <div class="modal-actions">
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
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-6);
}

.header-actions {
    display: flex;
    gap: var(--space-2);
    align-items: center;
}

.current-key-display {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-right: 16px;
    font-size: 0.875rem;
}

.text-muted {
    color: var(--text-secondary);
}

.key-val {
    background: var(--bg-tertiary);
    padding: 2px 6px;
    border-radius: 4px;
    font-family: monospace;
    color: var(--text-primary);
}

.text-warning {
    color: var(--error);
    margin-right: 16px;
    font-size: 0.875rem;
}

.voice-categories {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

.category-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.category-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  padding-bottom: var(--space-2);
  border-bottom: 2px solid var(--border-color);
}

.voices-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--space-4);
}

.voice-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4);
  transition: transform 0.2s;
}

.voice-card:hover {
  transform: translateY(-2px);
  border-color: var(--primary);
}

.voice-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  background: var(--bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-light);
}

.voice-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.voice-info h3 {
  margin: 0;
  font-size: 1rem;
}

.voice-id {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-family: monospace;
}

.badge {
  display: inline-block;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 4px;
  width: fit-content;
  margin-top: 4px;
}

.badge-system {
  background: #3b82f6;
  color: white;
}

.badge-cloned {
  background: #10b981;
  color: white;
}

.badge-generated {
  background: #8b5cf6;
  color: white;
}

.mini-player {
    height: 30px;
    width: 120px;
}

.actions {
  display: flex;
  gap: var(--space-2);
  align-items: center;
}

/* Modal */
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
}

.modal {
  width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-row {
  display: flex;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.hint {
  font-size: 0.8rem;
  color: var(--text-tertiary);
  margin: 0;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-2);
}

textarea {
  min-height: 80px;
  resize: vertical;
}
</style>
