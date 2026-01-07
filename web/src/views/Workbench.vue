<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import { Play, Download, Trash2, Cpu } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const tasks = ref([])
const voices = ref([])
const keys = ref([])
const loading = ref(false)

const form = ref({
  text: '',
  voice_id: '',
  key_id: '',
  speed: 1.0,
  vol: 1.0,
  mode: 'sync' // sync or async
})

const api = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'
})

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
  if (!form.value.text || !form.value.voice_id || !form.value.key_id) {
    alert(t('workbench.alertComplete'))
    return
  }

  loading.value = true
  try {
    const res = await api.post('/synthesis', form.value)
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
    // Filter tasks where status is pending or processing
    const pendingTasks = tasks.value.filter(t => t.status !== 'success' && t.status !== 'failed')
    for (const task of pendingTasks) {
       // Only async tasks have task_id? If task_id > 0
       if (task.task_id) {
           try {
             const res = await api.get(`/synthesis/${task.id}/status?key_id=${form.value.key_id}`) // Use current form key or... ideally task should store key_id?
             // Task in DB doesn't store KeyID. We rely on form key.
             // If user changes key in form, polling might fail if key mismatch?
             // Assuming key provided works.
             const updated = res.data.data
             // Update local
             const idx = tasks.value.findIndex(t => t.id === updated.id)
             if (idx !== -1) tasks.value[idx] = updated
           } catch(e) {}
       }
    }
  }, 5000)
}

onUnmounted(() => {
    if (pollInterval) clearInterval(pollInterval)
})

// Helper for status translation
const getStatusLabel = (status) => {
  const map = {
    'success': t('workbench.statusSuccess'),
    'failed': t('workbench.statusFailed'),
    'pending': t('workbench.statusPending'),
    'processing': t('workbench.statusPending')
  }
  return map[status] || status
}

onMounted(init)
</script>

<template>
  <div class="page page-workbench">
    <div class="workbench-layout">
      <!-- Left Panel: Input -->
      <div class="input-panel card">
        <header>
          <h2><Cpu size="20"/> {{ t('workbench.title') }}</h2>
        </header>
        
        <div class="form-group">
          <label>{{ t('workbench.labelKey') }}</label>
          <select v-model="form.key_id">
             <option v-for="k in keys" :key="k.id" :value="k.id">
              {{ k.platform }} - {{ k.key.substring(0,8) }}...
            </option>
          </select>
        </div>

        <div class="form-group">
          <label>{{ t('workbench.labelVoice') }}</label>
          <select v-model="form.voice_id">
            <option v-for="v in voices" :key="v.id" :value="v.voice_id">
              {{ v.name }} ({{ v.type }})
            </option>
          </select>
        </div>

        <div class="controls-row">
            <div class="form-group checkbox-group">
                <input type="checkbox" id="modeAsync" v-model="form.mode" :true-value="'async'" :false-value="'sync'">
                <label for="modeAsync">{{ t('workbench.modeAsync') }}</label>
            </div>
        </div>

        <div class="controls-row">
          <div class="form-group half">
            <label>{{ t('workbench.labelSpeed') }} ({{ form.speed }}x)</label>
            <input type="range" v-model.number="form.speed" min="0.5" max="2" step="0.1" />
          </div>
          <div class="form-group half">
            <label>{{ t('workbench.labelVol') }} ({{ form.vol }})</label>
            <input type="range" v-model.number="form.vol" min="0.1" max="2" step="0.1" />
          </div>
        </div>

        <div class="form-group flex-1">
          <label>{{ t('workbench.labelText') }}</label>
          <textarea 
            v-model="form.text" 
            :placeholder="t('workbench.phText')" 
            class="text-input"
          ></textarea>
        </div>

        <button @click="generate" :disabled="loading" class="btn btn-primary full-width">
          {{ loading ? t('workbench.btnGenerating') : t('workbench.btnGenerate') }}
        </button>
      </div>

      <!-- Right Panel: History/Output -->
      <div class="history-panel">
        <h3>{{ t('workbench.historyTitle') }}</h3>
        
        <div class="tasks-list">
          <div v-for="task in tasks" :key="task.id" class="card task-item">
             <div class="task-header">
               <span class="task-voice">{{ task.voice_id }}</span>
               <span class="task-time">{{ new Date(task.created_at).toLocaleString() }}</span>
             </div>
             <p class="task-text">{{ task.text }}</p>
             
             <div class="task-status">
                <span v-if="task.status === 'success'" class="success-badge">{{ getStatusLabel('success') }}</span>
                <span v-else-if="task.status === 'failed'" class="error-badge">{{ getStatusLabel('failed') }}: {{ task.error }}</span>
                <span v-else class="pending-badge">{{ getStatusLabel(task.status) }}</span>
             </div>

             <div v-if="task.status === 'success'" class="task-actions">
               <audio controls :src="task.output" class="audio-player"></audio>
               <a :href="task.output" download class="btn-icon">
                 <Download size="16"/>
               </a>
               <button @click="deleteTask(task.id)" class="btn-icon delete">
                 <Trash2 size="16"/>
               </button>
             </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Include existing styles via scoped or ... */
.page-workbench {
  height: 100%;
}

.workbench-layout {
  display: flex;
  gap: var(--space-6);
  height: calc(100vh - 48px);
}

.input-panel {
  flex: 0 0 400px;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  height: 100%;
  overflow-y: auto;
}

.history-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  height: 100%;
  overflow: hidden;
}

h2 {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin: 0;
  font-size: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.flex-1 {
  flex: 1;
}

.text-input {
  flex: 1;
  resize: none;
  min-height: 200px;
}

.controls-row {
  display: flex;
  gap: var(--space-4);
  align-items: center;
}

.checkbox-group {
    flex-direction: row;
    align-items: center;
    gap: var(--space-2);
}

.half {
  flex: 1;
}

.tasks-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding-bottom: var(--space-6);
}

.task-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.task-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  color: var(--text-tertiary);
}

.task-text {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.task-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.audio-player {
  flex: 1;
  height: 32px;
}

.success-badge { color: var(--success); font-size: 0.8rem; }
.error-badge { color: var(--error); font-size: 0.8rem; }
.pending-badge { color: var(--warning); font-size: 0.8rem; }
</style>
