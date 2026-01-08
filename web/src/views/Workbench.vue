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
    <div class="workbench-layout">
      <!-- Left: Form Section -->
      <section class="config-section">
        <div class="card main-card">
          <header class="section-header">
            <div class="header-title">
              <Mic class="icon-primary" />
              <h2>{{ t('workbench.title') }}</h2>
            </div>
            <div v-if="defaultKey" class="key-badge">
              <Key size="14" />
              <span>{{ defaultKey.remark || (defaultKey.key.substring(0,8) + '...') }}</span>
            </div>
          </header>

          <div class="form-scroll-area">
            <!-- Basic Configuration Group -->
            <div class="form-group-section">
              <h3 class="group-title">{{ t('workbench.sectionBasic') || '基础配置' }}</h3>
              <div class="form-grid">
                <div class="form-item">
                  <label class="label-with-tip">
                    {{ t('workbench.labelModel') }}
                    <div class="tooltip" :data-tip="t('workbench.tips.model')" :title="t('workbench.tips.model')"><Info size="14"/></div>
                  </label>
                  <select v-model="form.model" class="custom-select">
                    <option v-for="opt in modelOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                  </select>
                </div>

                <div class="form-item">
                  <label class="label-with-tip">
                    {{ t('workbench.labelVoice') }}
                    <div class="tooltip" :data-tip="t('workbench.tips.voiceId')" :title="t('workbench.tips.voiceId')"><Info size="14"/></div>
                  </label>
                  <select v-model="form.voice_id" class="custom-select">
                    <option v-for="v in voices" :key="v.voice_id" :value="v.voice_id">
                      {{ v.name }}
                    </option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Input Section -->
            <div class="form-group-section">
              <div class="flex justify-between items-center mb-4">
                <h3 class="group-title mb-0">{{ t('workbench.sectionInput') || '文本输入' }}</h3>
                <div class="segmented-control">
                  <button 
                    :class="{ active: inputType === 'text' }" 
                    @click="inputType = 'text'"
                  >{{ t('workbench.inputType.text') }}</button>
                  <button 
                    :class="{ active: inputType === 'file' }" 
                    @click="inputType = 'file'"
                  >{{ t('workbench.inputType.file') }}</button>
                </div>
              </div>

              <div v-if="inputType === 'text'" class="form-item">
                <textarea 
                  v-model="form.text" 
                  :placeholder="t('workbench.phText')" 
                  rows="6"
                  maxlength="50000"
                  class="modern-textarea"
                ></textarea>
                <div class="textarea-footer">
                  <span class="char-count" :class="{ 'text-error': form.text.length > 45000 }">
                    {{ form.text.length.toLocaleString() }} / 50,000
                  </span>
                </div>
              </div>

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
                    <component :is="isUploading ? Cpu : Library" :class="{ 'animate-spin': isUploading }" size="32" />
                    <p>{{ isUploading ? t('workbench.btnUploading') : (form.text_file_id ? '文件已就绪' : t('workbench.btnUploadFile')) }}</p>
                    <span v-if="form.text_file_id" class="file-id-tag">ID: {{ form.text_file_id }}</span>
                  </div>
                </label>
              </div>
            </div>

            <!-- Audio Parameters -->
            <div class="form-group-section">
              <h3 class="group-title">{{ t('workbench.sectionParameters') || '参数调节' }}</h3>
              <div class="slider-grid">
                <div class="slider-item">
                  <div class="slider-header">
                    <label>{{ t('workbench.labelSpeed') }}</label>
                    <span class="value-badge">{{ form.speed.toFixed(1) }}x</span>
                  </div>
                  <input type="range" v-model.number="form.speed" min="0.5" max="2.0" step="0.1" class="modern-range" />
                </div>
                <div class="slider-item">
                  <div class="slider-header">
                    <label>{{ t('workbench.labelVol') }}</label>
                    <span class="value-badge">{{ form.vol.toFixed(1) }}x</span>
                  </div>
                  <input type="range" v-model.number="form.vol" min="0.1" max="10.0" step="0.1" class="modern-range" />
                </div>
              </div>
            </div>

            <!-- Advanced Settings -->
            <div class="advanced-collapsible" :class="{ expanded: showAdvanced }">
              <button class="collapsible-trigger" @click="showAdvanced = !showAdvanced">
                <div class="flex items-center gap-2">
                  <div class="icon-wrapper">
                    <ChevronDown class="chevron-icon" />
                  </div>
                  <span>{{ t('workbench.sectionAdvanced') }}</span>
                </div>
              </button>
              
              <div class="collapsible-content">
                <div class="form-grid mt-4">
                  <div class="form-item">
                    <label class="label-with-tip">
                       <span>{{ t('workbench.labelPitch') }}: {{ form.pitch }}</span>
                       <div class="tooltip" :data-tip="t('workbench.tips.pitch')" :title="t('workbench.tips.pitch')"><Info size="14"/></div>
                    </label>
                    <input type="range" v-model.number="form.pitch" min="-12" max="12" step="1" class="modern-range" />
                  </div>
                  <div class="form-item">
                    <label class="label-with-tip">
                       <span>{{ t('workbench.labelEmotion') }}</span>
                       <div class="tooltip" :data-tip="t('workbench.tips.emotion')"><Info size="14"/></div>
                    </label>
                    <select v-model="form.emotion" class="custom-select">
                      <option v-for="opt in emotionOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                    </select>
                  </div>
                  <div class="form-item">
                    <label class="label-with-tip">
                       <span>{{ t('workbench.labelLangBoost') }}</span>
                       <div class="tooltip" :data-tip="t('workbench.tips.langBoost')"><Info size="14"/></div>
                    </label>
                    <select v-model="form.language_boost" class="custom-select">
                      <option value="auto">{{ t('workbench.options.auto') }}</option>
                      <option value="Chinese">{{ t('workbench.options.chinese') }}</option>
                      <option value="English">{{ t('workbench.options.english') }}</option>
                      <option value="Japanese">{{ t('workbench.options.japanese') }}</option>
                    </select>
                  </div>
                </div>

                <div class="form-item mt-4">
                  <div class="checkbox-card" @click="form.watermark = !form.watermark">
                    <div class="checkbox-info">
                      <span class="checkbox-label">{{ t('workbench.labelWatermark') }}</span>
                      <p class="checkbox-desc">{{ t('workbench.tips.watermark') }}</p>
                    </div>
                    <div class="switch" :class="{ active: form.watermark }"></div>
                  </div>
                </div>

                <!-- Audio Settings Fieldset -->
                <div class="form-fieldset mt-6">
                  <h4 class="fieldset-title">{{ t('workbench.labelAudioSetting') }}</h4>
                  <div class="form-grid">
                    <div class="form-item">
                      <label>{{ t('workbench.labelSampleRate') }}</label>
                      <select v-model.number="form.sample_rate" class="custom-select">
                        <option :value="32000">32000Hz</option>
                        <option :value="44100">44100Hz</option>
                        <option :value="24000">24000Hz</option>
                      </select>
                    </div>
                    <div class="form-item">
                      <label>{{ t('workbench.labelBitrate') }}</label>
                      <select v-model.number="form.bitrate" class="custom-select">
                        <option :value="32000">32kbps</option>
                        <option :value="64000">64kbps</option>
                        <option :value="128000">128kbps</option>
                        <option :value="256000">256kbps</option>
                      </select>
                    </div>
                    <div class="form-item">
                      <label>{{ t('workbench.labelFormat') }}</label>
                      <select v-model="form.format" class="custom-select">
                        <option value="mp3">MP3</option>
                        <option value="wav">WAV</option>
                        <option value="pcm">PCM</option>
                        <option value="flac">FLAC</option>
                      </select>
                    </div>
                    <div class="form-item">
                      <label>{{ t('workbench.labelChannel') }}</label>
                      <select v-model.number="form.channel" class="custom-select">
                        <option :value="1">{{ t('workbench.options.mono') }}</option>
                        <option :value="2">{{ t('workbench.options.stereo') }}</option>
                      </select>
                    </div>
                  </div>
                </div>

                <!-- Voice Modify Fieldset -->
                <div class="form-fieldset mt-6">
                  <h4 class="fieldset-title">{{ t('workbench.labelVoiceModify') }}</h4>
                  <div class="form-item mb-4">
                    <label class="label-with-tip">
                       <span>{{ t('workbench.labelPronunciationDict') }}</span>
                       <div class="tooltip" :data-tip="t('workbench.tips.pronunciationDict')"><Info size="14"/></div>
                    </label>
                    <textarea 
                       v-model="form.pronunciation_dict_str" 
                       :placeholder="t('workbench.phPronunciationDict')"
                       rows="3"
                       class="modern-textarea short"
                    ></textarea>
                  </div>
                  
                  <div class="slider-grid">
                    <div class="slider-item">
                      <div class="slider-header">
                        <label class="label-with-tip">
                           <span>{{ t('workbench.labelVmPitch') }}</span>
                           <div class="tooltip" :data-tip="t('workbench.tips.vmPitch')"><Info size="14"/></div>
                        </label>
                        <span class="value-badge">{{ form.voice_modify.pitch }}</span>
                      </div>
                      <input type="range" v-model.number="form.voice_modify.pitch" min="-100" max="100" class="modern-range" />
                    </div>
                    <div class="slider-item">
                      <div class="slider-header">
                        <label class="label-with-tip">
                           <span>{{ t('workbench.labelVmIntensity') }}</span>
                           <div class="tooltip" :data-tip="t('workbench.tips.vmIntensity')"><Info size="14"/></div>
                        </label>
                        <span class="value-badge">{{ form.voice_modify.intensity }}</span>
                      </div>
                      <input type="range" v-model.number="form.voice_modify.intensity" min="-100" max="100" class="modern-range" />
                    </div>
                    <div class="slider-item">
                       <div class="slider-header">
                        <label class="label-with-tip">
                           <span>{{ t('workbench.labelVmTimbre') }}</span>
                           <div class="tooltip" :data-tip="t('workbench.tips.vmTimbre')"><Info size="14"/></div>
                        </label>
                        <span class="value-badge">{{ form.voice_modify.timbre }}</span>
                      </div>
                      <input type="range" v-model.number="form.voice_modify.timbre" min="-100" max="100" class="modern-range" />
                    </div>
                  </div>
                  
                  <div class="form-item mt-4">
                    <label>{{ t('workbench.labelVmEffects') }}</label>
                    <select v-model="form.sound_effects" class="custom-select">
                      <option v-for="opt in soundEffectOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <footer class="form-footer">
            <button 
              @click="generate" 
              :disabled="loading" 
              class="btn btn-primary btn-xl w-full"
            >
              <component :is="loading ? Cpu : Play" :class="{ 'animate-spin': loading }" size="20" />
              <span>{{ loading ? t('workbench.btnGenerating') : t('workbench.btnGenerate') }}</span>
            </button>
          </footer>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.workbench-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.workbench-layout {
  display: flex;
  gap: var(--space-6);
  height: 100%;
}

.config-section {
  flex: 1;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
}

.main-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
  border-radius: var(--radius-xl);
}

.section-header {
  padding: var(--space-6);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-secondary);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.header-title h2 {
  margin: 0;
  font-size: 1.25rem;
}

.icon-primary {
  color: var(--primary);
}

.key-badge {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.form-scroll-area {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

.group-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--space-4);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--space-4);
}

.label-with-tip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  width: auto;
}

.tooltip {
  position: relative;
  display: inline-flex;
  color: var(--text-tertiary);
  cursor: help;
  line-height: 1;
}

.tooltip::after {
  content: attr(data-tip);
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  padding: 6px 10px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: 0.75rem;
  white-space: pre-wrap;
  width: max-content;
  max-width: 200px;
  opacity: 0;
  visibility: hidden;
  transition: all var(--transition-fast);
  pointer-events: none;
  z-index: 10;
  margin-bottom: 8px;
  box-shadow: var(--shadow-md);
  font-weight: normal;
  text-align: center;
}

.tooltip:hover::after {
  opacity: 1;
  visibility: visible;
  transform: translateX(-50%) translateY(-4px);
}

.form-fieldset {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  padding: var(--space-6); /* Increased padding */
  border: 1px solid var(--border-color);
  margin-top: var(--space-6); /* Ensure consistent spacing */
}

.fieldset-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-secondary);
  margin: 0 0 var(--space-6) 0; /* Increased bottom margin */
  text-transform: uppercase;
  letter-spacing: 0.05em;
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.fieldset-title::before {
  content: '';
  display: block;
  width: 4px;
  height: 16px;
  background: var(--primary);
  border-radius: var(--radius-full);
}

.modern-textarea.short {
  min-height: 80px;
}

.custom-select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2364748b'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right var(--space-3) center;
  background-size: 1rem;
  padding-right: var(--space-10);
}

.segmented-control {
  display: flex;
  background: var(--bg-tertiary);
  padding: 2px;
  border-radius: var(--radius-md);
}

.segmented-control button {
  padding: var(--space-1) var(--space-4);
  border-radius: var(--radius-sm);
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.segmented-control button.active {
  background: var(--bg-primary);
  color: var(--primary);
  box-shadow: var(--shadow-sm);
}

.modern-textarea {
  resize: vertical;
  min-height: 160px;
  line-height: 1.6;
}

.textarea-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-2);
}

.char-count {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.file-upload-zone {
  position: relative;
}

.upload-label {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-10);
  border: 2px dashed var(--border-color);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
  background: var(--bg-primary);
}

.upload-label:hover {
  border-color: var(--primary);
  background: var(--primary-bg);
}

.upload-label.uploading {
  cursor: wait;
  opacity: 0.7;
}

.upload-content {
  text-align: center;
  color: var(--text-secondary);
}

.file-id-tag {
  display: inline-block;
  margin-top: var(--space-2);
  padding: 2px 8px;
  background: var(--success-bg);
  color: var(--success);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  font-weight: 600;
}

.slider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--space-6);
}

.slider-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-2);
}

.value-badge {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--primary);
  background: var(--primary-bg);
  padding: 2px 8px;
  border-radius: var(--radius-full);
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

.advanced-collapsible {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.collapsible-trigger {
  width: 100%;
  padding: var(--space-4);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-weight: 600;
  transition: background var(--transition-fast);
}

.collapsible-trigger:hover {
  background: var(--bg-tertiary);
}

.chevron-icon {
  transition: transform var(--transition-normal);
}

.expanded .chevron-icon {
  transform: rotate(180deg);
}

.collapsible-content {
  max-height: 0;
  overflow: hidden;
  transition: all var(--transition-normal);
  padding: 0 var(--space-4);
}

.expanded .collapsible-content {
  max-height: none;
  overflow: visible;
  padding: var(--space-4);
  border-top: 1px solid var(--border-color);
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
