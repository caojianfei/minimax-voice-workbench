<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import { Download, Trash2, Search, RotateCcw, Filter } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const tasks = ref([])
const voices = ref([])
const loading = ref(false)

const filters = ref({
  text: '',
  status: '',
  voice_id: '',
  start_date: '',
  end_date: ''
})

const api = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'
})

const fetchVoices = async () => {
  try {
    const res = await api.get('/voices')
    voices.value = res.data.data
  } catch (e) {
    console.error(e)
  }
}

const fetchTasks = async () => {
  loading.value = true
  try {
    const params = {}
    if (filters.value.text) params.text = filters.value.text
    if (filters.value.status) params.status = filters.value.status
    if (filters.value.voice_id) params.voice_id = filters.value.voice_id
    if (filters.value.start_date) params.start_date = filters.value.start_date
    if (filters.value.end_date) params.end_date = filters.value.end_date

    const res = await api.get('/synthesis', { params })
    tasks.value = res.data.data
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.value = {
    text: '',
    status: '',
    voice_id: '',
    start_date: '',
    end_date: ''
  }
  fetchTasks()
}

const deleteTask = async (id) => {
  if (!confirm(t('audioManagement.deleteConfirm'))) return
  try {
    await api.delete(`/synthesis/${id}`)
    tasks.value = tasks.value.filter(t => t.id !== id)
  } catch (e) {
    console.error(e)
    alert(t('audioManagement.deleteFail') || 'Delete failed')
  }
}

// Polling
let pollInterval = null
const startPolling = () => {
  if (pollInterval) clearInterval(pollInterval)
  pollInterval = setInterval(async () => {
    const pendingTasks = tasks.value.filter(t => t.status !== 'success' && t.status !== 'failed')
    if (pendingTasks.length === 0) return

    for (const task of pendingTasks) {
       if (task.task_id) {
           try {
             const res = await api.get(`/synthesis/${task.id}/status`) 
             const updated = res.data.data
             const idx = tasks.value.findIndex(t => t.id === updated.id)
             if (idx !== -1) tasks.value[idx] = updated
           } catch(e) {}
       }
    }
  }, 5000)
}

onMounted(() => {
  fetchVoices()
  fetchTasks()
  startPolling()
})

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
})
</script>

<template>
  <div class="page page-audio">
    <header class="page-header">
      <h2>{{ t('audioManagement.title') }}</h2>
    </header>

    <!-- Filters -->
    <div class="filters-card card">
      <div class="filters-row">
        <div class="filter-group">
          <label>{{ t('audioManagement.filters.text') }}</label>
          <input 
            type="text" 
            v-model="filters.text" 
            :placeholder="t('audioManagement.filters.search')" 
            @keyup.enter="fetchTasks"
          />
        </div>
        
        <div class="filter-group">
          <label>{{ t('audioManagement.filters.status') }}</label>
          <select v-model="filters.status" @change="fetchTasks">
            <option value="">{{ t('audioManagement.filters.allStatus') }}</option>
            <option value="success">{{ t('audioManagement.status.success') }}</option>
            <option value="processing">{{ t('audioManagement.status.processing') }}</option>
            <option value="failed">{{ t('audioManagement.status.failed') }}</option>
          </select>
        </div>

        <div class="filter-group">
          <label>{{ t('audioManagement.filters.voice') }}</label>
          <select v-model="filters.voice_id" @change="fetchTasks">
            <option value="">{{ t('audioManagement.filters.allVoices') }}</option>
            <option v-for="v in voices" :key="v.voice_id" :value="v.voice_id">
              {{ v.name }}
            </option>
          </select>
        </div>

        <div class="filter-group">
          <label>{{ t('audioManagement.filters.startDate') }}</label>
          <input type="date" v-model="filters.start_date" />
        </div>

        <div class="filter-group">
          <label>{{ t('audioManagement.filters.endDate') }}</label>
          <input type="date" v-model="filters.end_date" />
        </div>

        <div class="filter-actions">
          <button class="btn btn-primary" @click="fetchTasks">
            <Search size="18" />
            {{ t('audioManagement.filters.search') }}
          </button>
          <button class="btn btn-secondary" @click="resetFilters">
            <RotateCcw size="18" />
            {{ t('audioManagement.filters.reset') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Task List -->
    <div class="task-list-container card">
      <div v-if="loading" class="loading-state">
        Loading...
      </div>
      <div v-else-if="tasks.length === 0" class="empty-state">
        No tasks found.
      </div>
      <div v-else class="task-grid">
        <div v-for="task in tasks" :key="task.id" class="task-card">
          <div class="task-header">
            <span class="task-id">#{{ task.id }}</span>
            <span class="status-badge" :class="task.status">
               {{ t('audioManagement.status.' + task.status) || task.status }}
            </span>
          </div>
          <div class="task-body">
            <p class="task-text" :title="task.text">
              {{ task.text.length > 150 ? task.text.substring(0, 150) + '...' : task.text }}
            </p>
            <div class="task-meta">
              <span class="meta-item">
                <strong>{{ t('audioManagement.filters.voice') }}:</strong> {{ task.voice_id }}
              </span>
              <span class="meta-item">
                <strong>Date:</strong> {{ new Date(task.created_at).toLocaleString() }}
              </span>
            </div>
            <div v-if="task.error" class="task-error">{{ task.error }}</div>
          </div>
          <div class="task-footer">
            <div class="audio-wrapper" v-if="task.status === 'success'">
              <audio controls :src="task.output" class="audio-player"></audio>
            </div>
            <div class="actions">
              <a v-if="task.status === 'success'" :href="task.output" download class="btn-icon" title="Download">
                <Download size="18" />
              </a>
              <button @click="deleteTask(task.id)" class="btn-icon delete" title="Delete">
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
.page-audio {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-4);
  height: calc(100vh - 80px); /* Adjust based on navbar */
  overflow-y: auto;
}

.filters-card {
  padding: var(--space-4);
}

.filters-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
  align-items: flex-end;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.filter-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.filter-group input,
.filter-group select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  min-width: 150px;
}

.filter-actions {
  display: flex;
  gap: var(--space-2);
}

.task-list-container {
  flex: 1;
  padding: var(--space-4);
  overflow-y: auto;
}

.task-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: var(--space-4);
}

.task-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  transition: transform 0.2s;
}

.task-card:hover {
  border-color: var(--primary-color);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-id {
  font-family: monospace;
  color: var(--text-secondary);
  font-size: 0.8rem;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.success {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.status-badge.failed {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.status-badge.processing {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.task-text {
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--text-primary);
  /* Limit lines */
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.task-meta {
  font-size: 0.8rem;
  color: var(--text-secondary);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-error {
  color: var(--error);
  font-size: 0.8rem;
  padding: 4px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 4px;
}

.task-footer {
  margin-top: auto;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  justify-content: space-between;
}

.audio-wrapper {
  flex: 1;
}

.audio-player {
  width: 100%;
  height: 32px;
}

.actions {
  display: flex;
  gap: var(--space-2);
}

.btn-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: var(--bg-primary);
  color: var(--text-primary);
}

.btn-icon.delete:hover {
  color: var(--error);
}
</style>