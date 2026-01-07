<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Trash2, Plus, Key as KeyIcon } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const keys = ref([])
const newKey = ref('')
const loading = ref(false)

const api = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'
})

const fetchKeys = async () => {
  try {
    const res = await api.get('/keys')
    keys.value = res.data.data
  } catch (e) {
    console.error(e)
  }
}

const addKey = async () => {
  if (!newKey.value) return
  loading.value = true
  try {
    await api.post('/keys', { key: newKey.value, platform: 'minimax' })
    newKey.value = ''
    fetchKeys()
  } catch (e) {
    alert(t('keys.alertAddFail'))
  } finally {
    loading.value = false
  }
}

const deleteKey = async (id) => {
  if (!confirm(t('keys.confirmDelete'))) return
  try {
    await api.delete(`/keys/${id}`)
    fetchKeys()
  } catch (e) {
    alert(t('keys.alertDeleteFail'))
  }
}

onMounted(fetchKeys)
</script>

<template>
  <div class="page text-white">
    <header class="header">
      <h1>{{ t('keys.title') }}</h1>
      <p class="subtitle">{{ t('keys.subtitle') }}</p>
    </header>

    <div class="card add-key-card">
      <div class="input-group">
        <input 
          v-model="newKey" 
          type="text" 
          :placeholder="t('keys.placeholder')" 
          class="key-input"
        />
        <button @click="addKey" :disabled="loading" class="btn btn-primary">
          <Plus size="18" /> {{ t('keys.add') }}
        </button>
      </div>
    </div>

    <div class="keys-list">
      <div v-for="key in keys" :key="key.id" class="key-item card">
        <div class="key-info">
          <KeyIcon class="icon" />
          <div class="details">
            <span class="platform">{{ key.platform }}</span>
            <code class="key-value">{{ key.key.substring(0, 8) }}...{{ key.key.substring(key.key.length - 4) }}</code>
          </div>
        </div>
        <button @click="deleteKey(key.id)" class="btn-icon delete">
          <Trash2 size="18" />
        </button>
      </div>
      
      <div v-if="keys.length === 0" class="empty-state">
        {{ t('keys.noKeys') }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.header {
  margin-bottom: var(--space-6);
}

.subtitle {
  color: var(--text-secondary);
}

.add-key-card {
  margin-bottom: var(--space-6);
}

.input-group {
  display: flex;
  gap: var(--space-3);
}

.key-input {
  flex: 1;
}

.keys-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.key-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) var(--space-4);
}

.key-info {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.icon {
  color: var(--primary);
}

.details {
  display: flex;
  flex-direction: column;
}

.platform {
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
}

.key-value {
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
}

.btn-icon {
  padding: 8px;
  background: transparent;
  color: var(--text-secondary);
  border-radius: var(--radius-md);
  transition: all 0.2s;
}

.btn-icon:hover {
  background: var(--bg-tertiary);
  color: var(--error);
}

.empty-state {
  text-align: center;
  padding: var(--space-8);
  color: var(--text-secondary);
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--border-color);
}
</style>
