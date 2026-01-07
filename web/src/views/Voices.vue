<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Plus, Trash2, Play, Mic, Cloud, Palette } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const voices = ref([])
const keys = ref([])
const showModal = ref(false)
const modalMode = ref('clone') // 'clone' or 'design'
const loading = ref(false)

// Form Data - Shared for Clone/Design
const form = ref({
  name: '',
  key_id: '',
  file: null,    // for clone
  prompt: '',    // for design
  preview_text: 'Hello, this is a test voice.', // for design
  watermark: false
})

const api = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'
})

const fetchData = async () => {
  try {
    const [vRes, kRes] = await Promise.all([
      api.get('/voices'),
      api.get('/keys')
    ])
    voices.value = vRes.data.data
    keys.value = kRes.data.data
    
    // Auto specific key if only one
    if (keys.value.length > 0 && !form.value.key_id) {
      form.value.key_id = keys.value[0].id
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

const submitForm = async () => {
  if (modalMode.value === 'clone') await cloneVoice()
  else await designVoice()
}

const cloneVoice = async () => {
  if (!form.value.name || !form.value.file || !form.value.key_id) {
    alert(t('voices.alertFill'))
    return
  }

  loading.value = true
  const formData = new FormData()
  formData.append('name', form.value.name)
  formData.append('key_id', form.value.key_id)
  formData.append('file', form.value.file)

  try {
    await api.post('/voices/clone', formData)
    cleanupModal()
  } catch (e) {
    alert(t('voices.alertCloneFail') + ': ' + (e.response?.data?.message || e.message))
    loading.value = false
  }
}

const designVoice = async () => {
  if (!form.value.prompt || !form.value.preview_text || !form.value.key_id) {
    alert(t('voices.alertFill'))
    return
  }

  loading.value = true
  try {
    const payload = {
      name: form.value.name,
      key_id: form.value.key_id,
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
  form.value.prompt = ''
  fetchData()
}

const deleteVoice = async (voice) => {
  if (!confirm(t('voices.confirmDelete'))) return
  
  // Need key for remote delete. Try form key or first key?
  // Ideally, prompt user for key or use a selected "active key".
  // For now: use first key if available, or try without.
  const keyId = form.value.key_id || (keys.value[0] ? keys.value[0].id : '')
  
  try {
    await api.delete(`/voices/${voice.id}?key_id=${keyId}`)
    fetchData()
  } catch (e) {
    alert(t('voices.alertDeleteFail'))
  }
}

const syncVoices = async () => {
  const keyId = form.value.key_id || (keys.value[0] ? keys.value[0].id : '')
  if (!keyId) {
    alert('Please add an API Key first')
    return
  }
  
  loading.value = true
  try {
    const res = await api.post(`/voices/sync?key_id=${keyId}`)
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
        <div class="key-selector">
          <!-- Optional Key Selector for Global Acts like Sync -->
          <select v-model="form.key_id" placeholder="Select Key">
             <option v-for="k in keys" :key="k.id" :value="k.id">
              {{ k.platform }} - {{ k.key.substring(0,8) }}...
            </option>
          </select>
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

    <div class="voices-grid">
      <div v-for="voice in voices" :key="voice.id" class="card voice-card">
        <div class="voice-icon">
          <Mic size="24" />
        </div>
        <div class="voice-info">
          <h3>{{ voice.name }}</h3>
          <span class="voice-id">{{ voice.voice_id }}</span>
          <span class="badge">{{ voice.type }}</span>
        </div>
        <div class="actions">
           <!-- Preview? If design -->
           <audio v-if="voice.preview" :src="voice.preview" controls class="mini-player"></audio>
           <button @click="deleteVoice(voice)" class="btn-icon delete">
            <Trash2 size="18" />
          </button>
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
        
        <!-- Common -->
        <div class="form-group">
          <label>{{ t('voices.labelKey') }}</label>
          <select v-model="form.key_id">
            <option v-for="k in keys" :key="k.id" :value="k.id">
              {{ k.platform }} - {{ k.key.substring(0,8) }}...
            </option>
          </select>
        </div>
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

.key-selector select {
    padding: var(--space-2);
    width: 150px;
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
  background: var(--bg-primary);
  border-radius: 4px;
  color: var(--text-tertiary);
  width: fit-content;
  margin-top: 4px;
}

.mini-player {
    height: 30px;
    width: 100px;
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
  width: 450px;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
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
</style>
