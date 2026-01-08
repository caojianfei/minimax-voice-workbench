<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import axios from 'axios'
import { Play, Download, Trash2, Cpu, ChevronDown, ChevronUp, Info, Key, Library } from 'lucide-vue-next'
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
  // key_id: '',
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

    if (!file.name.endsWith('.txt') && !file.name.endsWith('.zip')) {
        alert('Only .txt and .zip files are allowed')
        return
    }

    const formData = new FormData()
    formData.append('file', file)

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
    // if (keys.value.length > 0) form.value.key_id = keys.value[0].id
    
    startPolling()
  } catch (e) {
    console.error(e)
  }
}

const defaultKey = computed(() => {
  return keys.value.find(k => k.is_default) || keys.value[0]
})

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

  if (!form.value.voice_id) {
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
  // Construct payload strictly according to MiniMax API documentation
  const payload = {
    model: form.value.model,
    text: inputType.value === 'text' ? form.value.text : undefined,
    text_file_id: inputType.value === 'file' && form.value.text_file_id ? parseInt(form.value.text_file_id) : undefined,
    language_boost: form.value.language_boost,
    voice_setting: {
      voice_id: form.value.voice_id,
      speed: form.value.speed,
      vol: form.value.vol,
      pitch: form.value.pitch
    },
    audio_setting: {
      audio_sample_rate: form.value.sample_rate,
      bitrate: form.value.bitrate,
      format: form.value.format,
      channel: form.value.channel
    },
    voice_modify: {
      pitch: form.value.voice_modify.pitch,
      intensity: form.value.voice_modify.intensity,
      timbre: form.value.voice_modify.timbre,
      sound_effects: form.value.sound_effects || undefined
    },
    pronunciation_dict,
    aigc_watermark: form.value.watermark
  }

  loading.value = true
  try {
    const res = await api.post('/synthesis', payload)
    alert(t('workbench.statusSuccess') + ' - Task ID: ' + res.data.data.id)
  } catch (e) {
    alert(t('workbench.alertGenFail') + ': ' + (e.response?.data?.message || e.message))
  } finally {
    loading.value = false
  }
}

onMounted(init)
</script>

<template>
  <div class="workbench-container">
    <div class="workbench-header">
      <div class="header-left">
        <h1>{{ t('workbench.title') }}</h1>
        <p class="subtitle">{{ t('workbench.subtitle') || 'Create high-quality speech with Minimax models' }}</p>
      </div>
      <div class="header-right">
        <div v-if="defaultKey" class="key-badge">
          <Key size="14" />
          <span>{{ defaultKey.remark || (defaultKey.key.substring(0,8) + '...') }}</span>
        </div>
      </div>
    </div>

    <div class="workbench-grid">
      <!-- Left Column: Configuration -->
      <aside class="config-panel">
        <div class="scroll-container">
          <!-- Model & Voice Card -->
          <div class="card config-card">
            <h3 class="card-title">{{ t('workbench.sectionBasic') || 'Basic Setup' }}</h3>
            
            <div class="form-group">
              <label class="label-with-tip">
                {{ t('workbench.labelModel') }}
                <div class="tooltip" :data-tip="t('workbench.tips.model')"><Info size="14"/></div>
              </label>
              <div class="select-wrapper">
                <select v-model="form.model" class="custom-select">
                  <option v-for="opt in modelOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label class="label-with-tip">
                {{ t('workbench.labelVoice') }}
                <div class="tooltip" :data-tip="t('workbench.tips.voiceId')"><Info size="14"/></div>
              </label>
              <div class="select-wrapper">
                <select v-model="form.voice_id" class="custom-select">
                  <option v-for="v in voices" :key="v.voice_id" :value="v.voice_id">
                    {{ v.name }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <!-- Parameters Card -->
          <div class="card config-card">
            <h3 class="card-title">{{ t('workbench.sectionParameters') || 'Audio Parameters' }}</h3>
            
            <div class="slider-group">
              <div class="slider-header">
                <label>{{ t('workbench.labelSpeed') }}</label>
                <span class="value-badge">{{ form.speed.toFixed(1) }}x</span>
              </div>
              <input type="range" v-model.number="form.speed" min="0.5" max="2.0" step="0.1" class="modern-range" />
            </div>

            <div class="slider-group">
              <div class="slider-header">
                <label>{{ t('workbench.labelVol') }}</label>
                <span class="value-badge">{{ form.vol.toFixed(1) }}x</span>
              </div>
              <input type="range" v-model.number="form.vol" min="0.1" max="10.0" step="0.1" class="modern-range" />
            </div>
            
            <div class="slider-group">
              <div class="slider-header">
                <label class="label-with-tip">
                   <span>{{ t('workbench.labelPitch') }}</span>
                </label>
                <span class="value-badge">{{ form.pitch }}</span>
              </div>
              <input type="range" v-model.number="form.pitch" min="-12" max="12" step="1" class="modern-range" />
            </div>
          </div>

          <!-- Advanced Settings -->
          <div class="card config-card">
            <button class="collapsible-header" @click="showAdvanced = !showAdvanced">
              <span>{{ t('workbench.sectionAdvanced') }}</span>
              <component :is="showAdvanced ? ChevronUp : ChevronDown" size="16" />
            </button>
            
            <Transition name="expand">
              <div v-show="showAdvanced" class="collapsible-content">
                <div class="form-group">
                  <label>{{ t('workbench.labelEmotion') }}</label>
                  <select v-model="form.emotion" class="custom-select">
                    <option v-for="opt in emotionOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                  </select>
                </div>

                <div class="form-group">
                  <label>{{ t('workbench.labelLangBoost') }}</label>
                  <select v-model="form.language_boost" class="custom-select">
                    <option value="auto">{{ t('workbench.options.auto') }}</option>
                    <option value="Chinese">{{ t('workbench.options.chinese') }}</option>
                    <option value="English">{{ t('workbench.options.english') }}</option>
                    <option value="Japanese">{{ t('workbench.options.japanese') }}</option>
                  </select>
                </div>

                <div class="form-group">
                  <div
                    class="checkbox-card compact"
                    role="switch"
                    tabindex="0"
                    :aria-checked="form.watermark"
                    @click="form.watermark = !form.watermark"
                    @keydown.space.prevent="form.watermark = !form.watermark"
                    @keydown.enter.prevent="form.watermark = !form.watermark"
                  >
                    <div class="checkbox-info">
                      <span class="checkbox-label">{{ t('workbench.labelWatermark') }}</span>
                      <span class="checkbox-desc">{{ t('workbench.tips.watermark') }}</span>
                    </div>
                    <div class="switch" :class="{ active: form.watermark }"></div>
                  </div>
                </div>

                <div class="separator"></div>
                
                <h4 class="sub-title">{{ t('workbench.labelAudioSetting') }}</h4>
                <div class="grid-2">
                  <div class="form-group">
                    <label class="small-label">{{ t('workbench.labelSampleRate') }}</label>
                    <select v-model.number="form.sample_rate" class="custom-select sm">
                      <option :value="32000">32k</option>
                      <option :value="44100">44.1k</option>
                      <option :value="24000">24k</option>
                    </select>
                  </div>
                  <div class="form-group">
                    <label class="small-label">{{ t('workbench.labelBitrate') }}</label>
                    <select v-model.number="form.bitrate" class="custom-select sm">
                      <option :value="32000">32k</option>
                      <option :value="64000">64k</option>
                      <option :value="128000">128k</option>
                    </select>
                  </div>
                  <div class="form-group">
                    <label class="small-label">{{ t('workbench.labelFormat') }}</label>
                    <select v-model="form.format" class="custom-select sm">
                      <option value="mp3">MP3</option>
                      <option value="wav">WAV</option>
                      <option value="pcm">PCM</option>
                      <option value="flac">FLAC</option>
                    </select>
                  </div>
                </div>

                <div class="separator"></div>

                <h4 class="sub-title">{{ t('workbench.labelVoiceModify') }}</h4>
                
                <div class="form-group">
                  <label class="label-with-tip">
                     <span>{{ t('workbench.labelPronunciationDict') }}</span>
                     <div class="tooltip" :data-tip="t('workbench.tips.pronunciationDict')"><Info size="14"/></div>
                  </label>
                  <textarea 
                     v-model="form.pronunciation_dict_str" 
                     :placeholder="t('workbench.phPronunciationDict')"
                     rows="2"
                     class="custom-textarea"
                  ></textarea>
                </div>

                <div class="slider-group">
                   <div class="slider-header">
                    <label class="label-with-tip">
                       <span>{{ t('workbench.labelVmPitch') }}</span>
                    </label>
                    <span class="value-badge">{{ form.voice_modify.pitch }}</span>
                  </div>
                  <input type="range" v-model.number="form.voice_modify.pitch" min="-100" max="100" class="modern-range" />
                </div>

                <div class="slider-group">
                   <div class="slider-header">
                    <label class="label-with-tip">
                       <span>{{ t('workbench.labelVmIntensity') }}</span>
                    </label>
                    <span class="value-badge">{{ form.voice_modify.intensity }}</span>
                  </div>
                  <input type="range" v-model.number="form.voice_modify.intensity" min="-100" max="100" class="modern-range" />
                </div>

                <div class="slider-group">
                   <div class="slider-header">
                    <label class="label-with-tip">
                       <span>{{ t('workbench.labelVmTimbre') }}</span>
                    </label>
                    <span class="value-badge">{{ form.voice_modify.timbre }}</span>
                  </div>
                  <input type="range" v-model.number="form.voice_modify.timbre" min="-100" max="100" class="modern-range" />
                </div>

                <div class="form-group">
                  <label>{{ t('workbench.labelVmEffects') }}</label>
                  <select v-model="form.sound_effects" class="custom-select">
                    <option v-for="opt in soundEffectOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                  </select>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </aside>

      <!-- Right Column: Workspace -->
      <main class="workspace-panel">
        <div class="card workspace-card">
          <div class="workspace-header">
             <div class="segmented-control">
              <button 
                :class="{ active: inputType === 'text' }" 
                @click="inputType = 'text'"
              >
                <span class="tab-label">{{ t('workbench.inputType.text') }}</span>
              </button>
              <button 
                :class="{ active: inputType === 'file' }" 
                @click="inputType = 'file'"
              >
                <span class="tab-label">{{ t('workbench.inputType.file') }}</span>
              </button>
            </div>
            
            <div class="char-count" v-if="inputType === 'text'">
              <span :class="{ 'text-error': form.text.length > 45000 }">
                {{ form.text.length.toLocaleString() }}
              </span>
              <span class="text-muted"> / 50,000</span>
            </div>
          </div>

          <div class="workspace-content">
            <textarea 
              v-if="inputType === 'text'"
              v-model="form.text" 
              :placeholder="t('workbench.phText')" 
              class="main-textarea"
              spellcheck="false"
            ></textarea>

            <div v-else class="file-upload-zone">
              <input 
                type="file" 
                ref="fileInput"
                accept=".txt,.zip"
                @change="handleFileUpload"
                class="hidden-input"
                id="file-upload"
              />
              <label for="file-upload" class="upload-label" :class="{ uploading: isUploading }">
                <div class="upload-content">
                  <component :is="isUploading ? Cpu : Library" :class="{ 'animate-spin': isUploading }" size="48" />
                  <h3>{{ isUploading ? t('workbench.btnUploading') : (form.text_file_id ? 'File Ready' : 'Upload Text/Zip') }}</h3>
                  <p v-if="!isUploading && !form.text_file_id" class="text-muted">Drag & drop or click to upload</p>
                  <span v-if="form.text_file_id" class="file-id-tag">File ID: {{ form.text_file_id }}</span>
                </div>
              </label>
            </div>
          </div>

          <div class="workspace-footer">
            <div class="pronunciation-bar">
               <!-- Optional: Pronunciation dict input could go here or be hidden in advanced -->
            </div>
            <button 
              @click="generate" 
              :disabled="loading" 
              class="btn btn-primary btn-xl generate-btn"
            >
              <component :is="loading ? Cpu : Play" :class="{ 'animate-spin': loading }" size="20" />
              <span>{{ loading ? t('workbench.btnGenerating') : t('workbench.btnGenerate') }}</span>
            </button>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.workbench-container {
  height: calc(100vh - var(--header-height) - 40px); /* Approx height adjustment */
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.workbench-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.subtitle {
  margin: 0;
  font-size: 0.95rem;
}

.workbench-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 500px 1fr;
  gap: var(--space-6);
  min-height: 0; /* Important for nested scrolling */
}

/* Left Panel */
.config-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.scroll-container {
  overflow-y: auto;
  padding-right: var(--space-2);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.config-card {
  padding: var(--space-5);
  background: var(--bg-primary); /* Stand out from bg-secondary background of layout */
}

.card-title {
  font-size: 0.875rem;
  text-transform: uppercase;
  color: var(--text-tertiary);
  margin-bottom: var(--space-4);
  letter-spacing: 0.05em;
}

.form-group {
  margin-bottom: var(--space-4);
}

.label-with-tip {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.tooltip {
  color: var(--text-tertiary);
  cursor: help;
}

.slider-group {
  margin-bottom: var(--space-5);
}

.slider-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-2);
  font-size: 0.875rem;
  font-weight: 500;
}

.value-badge {
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 0.75rem;
}

.collapsible-header {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: transparent;
  color: var(--text-primary);
  font-weight: 600;
  padding: 0;
}

.collapsible-content {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-color);
  overflow: hidden;
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease-in-out;
  max-height: 1200px;
  opacity: 1;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  margin-top: 0;
  border-top-color: transparent;
}

.checkbox-flex {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
}

.grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-3);
}

.small-label {
  font-size: 0.75rem;
  margin-bottom: var(--space-1);
}

.custom-select.sm {
  padding: var(--space-1) var(--space-2);
  font-size: 0.875rem;
}

/* Right Panel */
.workspace-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.workspace-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
  background: var(--bg-primary);
}

.workspace-header {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-secondary);
}

.segmented-control {
  background: var(--bg-tertiary);
  padding: 4px;
  border-radius: var(--radius-md);
  display: flex;
  gap: 4px;
}

.segmented-control button {
  padding: 4px 16px;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
  background: transparent;
}

.segmented-control button.active {
  background: var(--bg-primary);
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
}

.workspace-content {
  flex: 1;
  position: relative;
  display: flex;
  flex-direction: column;
}

.main-textarea {
  flex: 1;
  width: 100%;
  border: none;
  padding: var(--space-6);
  font-size: 1rem;
  line-height: 1.6;
  resize: none;
  outline: none;
  background: transparent;
}

.main-textarea:focus {
  box-shadow: none;
}

.file-upload-zone {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-8);
  background: var(--bg-secondary);
}

.upload-label {
  text-align: center;
  padding: var(--space-10);
  border: 2px dashed var(--border-color);
  border-radius: var(--radius-xl);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.upload-label:hover {
  border-color: var(--primary);
  background: var(--primary-bg);
}

.workspace-footer {
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--border-color);
  background: var(--bg-secondary);
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.generate-btn {
  min-width: 160px;
}

/* Responsive */
@media (max-width: 1024px) {
  .workbench-grid {
    grid-template-columns: 1fr;
    height: auto;
  }
  
  .workbench-container {
    height: auto;
  }

  .main-textarea {
    min-height: 300px;
  }
}

.modern-range {
  height: 6px;
  -webkit-appearance: none;
  background: var(--bg-primary); /* Changed from bg-tertiary to bg-primary for contrast */
  border-radius: var(--radius-full);
  outline: none;
  width: 100%; /* Ensure full width */
  cursor: pointer;
}

.modern-range::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid white;
  box-shadow: var(--shadow-sm);
  transition: transform var(--transition-fast);
}

.modern-range::-webkit-slider-thumb:hover {
  transform: scale(1.2);
}

.checkbox-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4);
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.checkbox-card.compact {
  padding: var(--space-3);
  user-select: none;
}

.checkbox-card:hover {
  border-color: var(--primary-light);
  background: var(--bg-secondary);
}

.checkbox-info {
  display: flex;
  flex-direction: column;
}

.checkbox-label {
  font-weight: 600;
  color: var(--text-primary);
}

.checkbox-desc {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin: 0;
}

.switch {
  width: 40px;
  height: 20px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-full);
  position: relative;
  transition: all var(--transition-fast);
}

.switch::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  background: white;
  border-radius: 50%;
  transition: all var(--transition-fast);
}

.switch.active {
  background: var(--primary);
}

.switch.active::after {
  left: 22px;
}

.form-footer {
  padding: var(--space-6);
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
}

.btn-xl {
  padding: var(--space-4);
  font-size: 1.125rem;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.hidden-input {
  display: none;
}

@media (max-width: 1024px) {
  .config-section {
    max-width: 100%;
  }
}
</style>
