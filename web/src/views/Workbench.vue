<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import axios from 'axios'
import { Play, Download, Trash2, Cpu, ChevronDown, ChevronUp, Info } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const tasks = ref([])
const voices = ref([])
const keys = ref([])
const loading = ref(false)
const showAdvanced = ref(false)
const inputType = ref('text') // 'text' or 'file'

const form = ref({
  // Common
  model: 'speech-2.6-hd',
  text: '',
  text_file_id: '',
  voice_id: '',
  key_id: '',
  speed: 1.0,
  vol: 1.0,
  
  // Advanced
  pitch: 0,
  emotion: '',
  language_boost: 'auto',
  
  // Audio Settings
  sample_rate: 32000,
  bitrate: 128000,
  format: 'mp3',
  channel: 1,
  
  // Voice Modify
  voice_modify: {
    pitch: 0,
    intensity: 0,
    timbre: 0
  },
  sound_effects: '',
  
  watermark: false,
  
  // Extra
  pronunciation_dict_str: ''
})

const modelOptions = [
  { value: 'speech-2.6-hd', label: 'Speech 2.6 HD' },
  { value: 'speech-2.6-turbo', label: 'Speech 2.6 Turbo' },
  { value: 'speech-02-hd', label: 'Speech 02 HD' },
  { value: 'speech-02-turbo', label: 'Speech 02 Turbo' },
  { value: 'speech-01-hd', label: 'Speech 01 HD' },
  { value: 'speech-01-turbo', label: 'Speech 01 Turbo' },
]

const emotionOptions = computed(() => [
  { value: '', label: t('workbench.options.auto') },
  { value: 'happy', label: t('workbench.options.happy') },
  { value: 'sad', label: t('workbench.options.sad') },
  { value: 'angry', label: t('workbench.options.angry') },
  { value: 'fearful', label: t('workbench.options.fearful') },
  { value: 'disgusted', label: t('workbench.options.disgusted') },
  { value: 'surprised', label: t('workbench.options.surprised') },
  { value: 'calm', label: t('workbench.options.calm') },
  { value: 'fluent', label: t('workbench.options.fluent') },
  { value: 'whisper', label: t('workbench.options.whisper') },
])

const soundEffectOptions = computed(() => [
  { value: '', label: t('workbench.options.none') },
  { value: 'spacious_echo', label: t('workbench.options.spacious_echo') },
  { value: 'auditorium_echo', label: t('workbench.options.auditorium_echo') },
  { value: 'lofi_telephone', label: t('workbench.options.lofi_telephone') },
  { value: 'robotic', label: t('workbench.options.robotic') },
])

const api = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'
})

const isUploading = ref(false)
const fileInput = ref(null)

const handleFileUpload = async (event) => {
    const file = event.target.files[0]
    if (!file) return

    // Basic validation
    if (file.name.endsWith('.txt') || file.name.endsWith('.zip')) {
        // Proceed
    } else {
        alert('Only .txt and .zip files are allowed')
        return
    }

    if (!form.value.key_id) {
        alert(t('workbench.alertComplete'))
        return
    }

    const formData = new FormData()
    formData.append('file', file)
    formData.append('key_id', form.value.key_id)

    isUploading.value = true
    try {
        const res = await api.post('/synthesis/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        })
        form.value.text_file_id = res.data.data.file_id
    } catch (e) {
        alert('Upload failed: ' + (e.response?.data?.message || e.message))
    } finally {
        isUploading.value = false
        // Reset input so same file can be selected again if needed
        event.target.value = ''
    }
}

let pollInterval = null

const init = async () => {
  try {
    const [tRes, vRes, kRes] = await Promise.all([
      api.get('/synthesis'),
      api.get('/voices'),
      api.get('/keys')
    ])
    tasks.value = tRes.data.data
    voices.value = vRes.data.data
    keys.value = kRes.data.data

    if (voices.value.length > 0) form.value.voice_id = voices.value[0].voice_id
    if (keys.value.length > 0) form.value.key_id = keys.value[0].id
    
    startPolling()
  } catch (e) {
    console.error(e)
  }
}

const generate = async () => {
  // Clear mutually exclusive field based on input type
  if (inputType.value === 'text') {
    form.value.text_file_id = ''
    if (!form.value.text) {
        alert(t('workbench.alertComplete'))
        return
    }
  } else {
    form.value.text = ''
    if (!form.value.text_file_id) {
        alert(t('workbench.alertComplete'))
        return
    }
  }

  if (!form.value.voice_id || !form.value.key_id) {
    alert(t('workbench.alertComplete'))
    return
  }
  
  // Parse Pronunciation Dict
  let pronunciation_dict = undefined
  if (form.value.pronunciation_dict_str) {
      // Allow comma-separated strings like: "word/pron", "word2/pron2"
      // or plain: word/pron, word2/pron2
      // We will wrap them into { tone: [...] }
      const rawStr = form.value.pronunciation_dict_str
      const items = []
      
      // Simple split by comma, then trim and remove quotes if present
      const parts = rawStr.split(',')
      for (const p of parts) {
          let s = p.trim()
          // Remove surrounding quotes if both exist
          if ((s.startsWith('"') && s.endsWith('"')) || (s.startsWith("'") && s.endsWith("'"))) {
              s = s.substring(1, s.length - 1)
          }
          if (s) items.push(s)
      }
      
      if (items.length > 0) {
          pronunciation_dict = { tone: items }
      }
  }

  // Prepare payload
  const payload = {
    ...form.value,
    text_file_id: form.value.text_file_id ? parseInt(form.value.text_file_id) : 0,
    pronunciation_dict
  }

  loading.value = true
  try {
    const res = await api.post('/synthesis', payload)
    tasks.value.unshift(res.data.data) // Add to top
  } catch (e) {
    alert(t('workbench.alertGenFail') + ': ' + (e.response?.data?.message || e.message))
  } finally {
    loading.value = false
  }
}

const deleteTask = async (id) => {
  try {
    await api.delete(`/synthesis/${id}`)
    tasks.value = tasks.value.filter(t => t.id !== id)
  } catch (e) {
    console.error(e)
  }
}

// Polling for async tasks
const startPolling = () => {
  if (pollInterval) clearInterval(pollInterval)
  pollInterval = setInterval(async () => {
    // Check pending tasks
    const pendingTasks = tasks.value.filter(t => t.status !== 'success' && t.status !== 'failed')
    for (const task of pendingTasks) {
       if (task.task_id) {
           try {
             const res = await api.get(`/synthesis/${task.id}/status?key_id=${form.value.key_id}`) 
             const updated = res.data.data
             const idx = tasks.value.findIndex(t => t.id === updated.id)
             if (idx !== -1) tasks.value[idx] = updated
           } catch(e) {}
       }
    }
  }, 5000)
}

onMounted(init)

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
})
</script>

<template>
  <div class="page page-workbench">
    <div class="split-view">
      <!-- Left Panel: Configuration -->
      <div class="config-panel card">
        <header class="panel-header">
          <h2>{{ t('workbench.title') }}</h2>
        </header>

        <div class="form-container">
          <!-- API Key & Model (Common) -->
          <div class="form-row">
            <div class="form-group flex-1">
              <label>{{ t('workbench.labelKey') }}</label>
              <select v-model="form.key_id">
                <option v-for="k in keys" :key="k.id" :value="k.id">
                  {{ k.platform }} - {{ k.key.substring(0,8) }}...
                </option>
              </select>
            </div>
            <div class="form-group flex-1">
              <label class="label-with-tip">
                {{ t('workbench.labelModel') }}
                <div class="tooltip" :data-tip="t('workbench.tips.model')"><Info size="14"/></div>
              </label>
              <select v-model="form.model">
                <option v-for="opt in modelOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </div>
          </div>

          <!-- Input Type Selector -->
          <div class="form-group">
            <div class="radio-group">
              <label :class="{ active: inputType === 'text' }">
                <input type="radio" v-model="inputType" value="text"> {{ t('workbench.inputType.text') }}
              </label>
              <label :class="{ active: inputType === 'file' }">
                <input type="radio" v-model="inputType" value="file"> {{ t('workbench.inputType.file') }}
              </label>
            </div>
          </div>

          <!-- Content Input -->
          <div class="form-group" v-if="inputType === 'text'">
             <label class="label-with-tip">
                {{ t('workbench.labelText') }}
                <div class="tooltip" :data-tip="t('workbench.tips.text')"><Info size="14"/></div>
             </label>
            <textarea 
              v-model="form.text" 
              :placeholder="t('workbench.phText')" 
              rows="5"
              maxlength="50000"
            ></textarea>
            <div class="char-count">{{ form.text.length }} / 50000</div>
          </div>

          <div class="form-group" v-else>
             <label class="label-with-tip">
                {{ t('workbench.labelTextFileId') }}
                <div class="tooltip" :data-tip="t('workbench.tips.textFileId')"><Info size="14"/></div>
             </label>
            <div class="file-upload-row">
                <input 
                  type="file" 
                  ref="fileInput"
                  accept=".txt,.zip"
                  @change="handleFileUpload"
                  style="display: none"
                />
                <button class="btn secondary" @click="$refs.fileInput.click()">
                   {{ isUploading ? t('workbench.btnUploading') : t('workbench.btnUploadFile') }}
                </button>
                <input 
                  v-model="form.text_file_id" 
                  type="number" 
                  placeholder="File ID" 
                  readonly
                  class="file-id-input"
                />
            </div>
          </div>

          <!-- Voice Selection -->
          <div class="form-group">
             <label class="label-with-tip">
                {{ t('workbench.labelVoice') }}
                <div class="tooltip" :data-tip="t('workbench.tips.voiceId')"><Info size="14"/></div>
             </label>
            <select v-model="form.voice_id">
              <option v-for="v in voices" :key="v.voice_id" :value="v.voice_id">
                {{ v.name }} ({{ v.type }})
              </option>
            </select>
          </div>

          <!-- Speed & Volume -->
          <div class="form-row">
            <div class="form-group flex-1">
               <label class="label-with-tip">
                  {{ t('workbench.labelSpeed') }}: {{ form.speed }}
                  <div class="tooltip" :data-tip="t('workbench.tips.speed')"><Info size="14"/></div>
               </label>
              <input type="range" v-model.number="form.speed" min="0.5" max="2.0" step="0.1" />
            </div>
            <div class="form-group flex-1">
               <label class="label-with-tip">
                  {{ t('workbench.labelVol') }}: {{ form.vol }}
                  <div class="tooltip" :data-tip="t('workbench.tips.vol')"><Info size="14"/></div>
               </label>
              <input type="range" v-model.number="form.vol" min="0.1" max="10.0" step="0.1" />
            </div>
          </div>

          <!-- Advanced Toggle -->
          <div class="advanced-toggle" @click="showAdvanced = !showAdvanced">
            <span>{{ t('workbench.sectionAdvanced') }}</span>
            <component :is="showAdvanced ? ChevronUp : ChevronDown" size="16" />
          </div>

          <!-- Advanced Settings Panel -->
          <div v-show="showAdvanced" class="advanced-panel">
            <!-- Pitch & Emotion -->
            <div class="form-row">
              <div class="form-group flex-1">
                 <label class="label-with-tip">
                    {{ t('workbench.labelPitch') }}: {{ form.pitch }}
                    <div class="tooltip" :data-tip="t('workbench.tips.pitch')"><Info size="14"/></div>
                 </label>
                <input type="range" v-model.number="form.pitch" min="-12" max="12" step="1" />
              </div>
              <div class="form-group flex-1">
                 <label class="label-with-tip">
                    {{ t('workbench.labelEmotion') }}
                    <div class="tooltip" :data-tip="t('workbench.tips.emotion')"><Info size="14"/></div>
                 </label>
                <select v-model="form.emotion">
                  <option v-for="opt in emotionOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                </select>
              </div>
            </div>

            <!-- Language Boost & Watermark -->
            <div class="form-row">
               <div class="form-group flex-1">
                 <label class="label-with-tip">
                    {{ t('workbench.labelLangBoost') }}
                    <div class="tooltip" :data-tip="t('workbench.tips.langBoost')"><Info size="14"/></div>
                 </label>
                 <select v-model="form.language_boost">
                   <option value="auto">{{ t('workbench.options.auto') }}</option>
                   <option value="Chinese">{{ t('workbench.options.chinese') }}</option>
                   <option value="English">{{ t('workbench.options.english') }}</option>
                   <option value="Japanese">{{ t('workbench.options.japanese') }}</option>
                 </select>
               </div>
               <div class="form-group flex-1 checkbox-group">
                 <label class="label-with-tip">
                    {{ t('workbench.labelWatermark') }}
                    <div class="tooltip" :data-tip="t('workbench.tips.watermark')"><Info size="14"/></div>
                 </label>
                 <input type="checkbox" v-model="form.watermark" />
               </div>
            </div>

            <!-- Audio Settings -->
            <fieldset class="fieldset">
               <legend>{{ t('workbench.labelAudioSetting') }}</legend>
               <div class="form-row">
                 <div class="form-group flex-1">
                   <label>{{ t('workbench.labelSampleRate') }}</label>
                   <select v-model.number="form.sample_rate">
                     <option :value="32000">32000Hz</option>
                     <option :value="44100">44100Hz</option>
                     <option :value="24000">24000Hz</option>
                   </select>
                 </div>
                 <div class="form-group flex-1">
                   <label>{{ t('workbench.labelBitrate') }}</label>
                   <select v-model.number="form.bitrate">
                     <option :value="32000">32kbps</option>
                     <option :value="64000">64kbps</option>
                     <option :value="128000">128kbps</option>
                     <option :value="256000">256kbps</option>
                   </select>
                 </div>
               </div>
               <div class="form-row">
                 <div class="form-group flex-1">
                   <label>{{ t('workbench.labelFormat') }}</label>
                   <select v-model="form.format">
                     <option value="mp3">MP3</option>
                     <option value="wav">WAV</option>
                     <option value="pcm">PCM</option>
                     <option value="flac">FLAC</option>
                   </select>
                 </div>
                 <div class="form-group flex-1">
                   <label>{{ t('workbench.labelChannel') }}</label>
                   <select v-model.number="form.channel">
                     <option :value="1">{{ t('workbench.options.mono') }}</option>
                     <option :value="2">{{ t('workbench.options.stereo') }}</option>
                   </select>
                 </div>
               </div>
            </fieldset>

            <!-- Voice Modification -->
            <fieldset class="fieldset">
               <legend>{{ t('workbench.labelVoiceModify') }}</legend>
               <div class="form-group">
                 <label class="label-with-tip">
                    {{ t('workbench.labelPronunciationDict') }}
                    <div class="tooltip" :data-tip="t('workbench.tips.pronunciationDict')"><Info size="14"/></div>
                 </label>
                 <textarea 
                    v-model="form.pronunciation_dict_str" 
                    :placeholder="t('workbench.phPronunciationDict')"
                    rows="3"
                 ></textarea>
               </div>
               <div class="form-group">
                 <label class="label-with-tip">
                    {{ t('workbench.labelVmPitch') }}: {{ form.voice_modify.pitch }}
                    <div class="tooltip" :data-tip="t('workbench.tips.vmPitch')"><Info size="14"/></div>
                 </label>
                 <input type="range" v-model.number="form.voice_modify.pitch" min="-100" max="100" />
               </div>
               <div class="form-group">
                 <label class="label-with-tip">
                    {{ t('workbench.labelVmIntensity') }}: {{ form.voice_modify.intensity }}
                    <div class="tooltip" :data-tip="t('workbench.tips.vmIntensity')"><Info size="14"/></div>
                 </label>
                 <input type="range" v-model.number="form.voice_modify.intensity" min="-100" max="100" />
               </div>
               <div class="form-group">
                 <label class="label-with-tip">
                    {{ t('workbench.labelVmTimbre') }}: {{ form.voice_modify.timbre }}
                    <div class="tooltip" :data-tip="t('workbench.tips.vmTimbre')"><Info size="14"/></div>
                 </label>
                 <input type="range" v-model.number="form.voice_modify.timbre" min="-100" max="100" />
               </div>
               <div class="form-group">
                 <label>{{ t('workbench.labelVmEffects') }}</label>
                 <select v-model="form.sound_effects">
                   <option v-for="opt in soundEffectOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                 </select>
               </div>
            </fieldset>
          </div>

          <button @click="generate" :disabled="loading" class="btn btn-primary btn-block btn-lg">
            <component :is="loading ? Cpu : Play" size="20" />
            {{ loading ? t('workbench.btnGenerating') : t('workbench.btnGenerate') }}
          </button>
        </div>
      </div>

      <!-- Right Panel: History -->
      <div class="history-panel card">
        <header class="panel-header">
          <h2>{{ t('workbench.historyTitle') }}</h2>
        </header>
        <div class="task-list">
          <div v-for="task in tasks" :key="task.id" class="task-card">
            <div class="task-header">
              <span class="task-id">#{{ task.id }}</span>
              <span class="status-badge" :class="task.status">
                 {{ t('workbench.status' + (task.status.charAt(0).toUpperCase() + task.status.slice(1))) || task.status }}
              </span>
            </div>
            <div class="task-body">
              <p class="task-text">{{ task.text.substring(0, 100) }}{{ task.text.length > 100 ? '...' : '' }}</p>
              <div class="task-meta">
                <span>{{ task.voice_id }}</span>
                <span>{{ new Date(task.created_at).toLocaleString() }}</span>
              </div>
              <div v-if="task.error" class="task-error">{{ task.error }}</div>
            </div>
            <div class="task-actions" v-if="task.status === 'success'">
              <audio controls :src="task.output" class="audio-player"></audio>
              <a :href="task.output" download class="btn-icon">
                <Download size="18" />
              </a>
            </div>
             <div class="task-actions" v-if="task.status !== 'processing'">
                <button @click="deleteTask(task.id)" class="btn-icon delete">
                   <Trash2 size="18" />
                </button>
             </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-workbench {
  height: calc(100vh - 80px); /* Adjust based on navbar height */
  overflow: hidden;
}

.split-view {
  display: flex;
  gap: var(--space-6);
  height: 100%;
}

.config-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  min-width: 400px;
  max-width: 500px;
}

.history-panel {
  flex: 2;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.form-container {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-2);
}

.form-row {
  display: flex;
  gap: var(--space-4);
}

.flex-1 {
  flex: 1;
}

.advanced-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  user-select: none;
  font-weight: 500;
  color: var(--text-secondary);
  transition: background 0.2s;
}

.advanced-toggle:hover {
  background: var(--border-color);
  color: var(--text-primary);
}

.advanced-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-2);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  margin-top: calc(var(--space-2) * -1);
  border-top: none;
  padding-top: var(--space-4);
}

.fieldset {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: var(--space-3);
  margin: 0;
}

.fieldset legend {
  padding: 0 var(--space-2);
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.radio-group {
  display: flex;
  gap: var(--space-4);
}

.radio-group label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
}

.radio-group label.active {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.label-with-tip {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.tooltip {
  position: relative;
  display: inline-flex;
  color: var(--text-tertiary);
  cursor: help;
}

.tooltip:hover::after {
  content: attr(data-tip);
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: var(--bg-secondary);
  color: var(--text-primary);
  padding: var(--space-2);
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  white-space: nowrap;
  z-index: 10;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  border: 1px solid var(--border-color);
  pointer-events: none;
  margin-bottom: var(--space-2);
}

.checkbox-group {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.checkbox-group input {
    width: auto;
}

.char-count {
  text-align: right;
  font-size: 0.8rem;
  color: var(--text-tertiary);
  margin-top: var(--space-1);
}

.btn-lg {
  padding: var(--space-3);
  font-size: 1.1rem;
}

/* History Styles (Reused/Adapted) */
.task-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.task-card {
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: var(--space-4);
}

.task-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.task-id {
  font-family: monospace;
  color: var(--text-secondary);
}

.status-badge {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.8rem;
  background: var(--bg-primary);
}

.status-badge.success { color: var(--success); }
.status-badge.failed { color: var(--error); }
.status-badge.processing { color: var(--warning); }

.task-body {
  margin-bottom: var(--space-3);
}

.task-text {
  color: var(--text-primary);
  margin-bottom: var(--space-2);
  word-break: break-all;
}

.task-meta {
  display: flex;
  gap: var(--space-4);
  font-size: 0.85rem;
  color: var(--text-tertiary);
}

.task-error {
  color: var(--error);
  font-size: 0.9rem;
  margin-top: var(--space-2);
}

.task-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  justify-content: flex-end;
}

.audio-player {
  height: 32px;
  flex: 1;
}

@media (max-width: 768px) {
  .split-view {
    flex-direction: column;
    overflow-y: auto;
  }
  
  .config-panel, .history-panel {
    max-width: none;
    overflow: visible;
  }
}
</style>
