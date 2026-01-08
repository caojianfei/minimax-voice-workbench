<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Trash2, Plus, Key as KeyIcon, Star } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const keys = ref([])
const newKey = ref('')
const newRemark = ref('')
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
  if (newRemark.value.length > 100) return 
  
  loading.value = true
  try {
    await api.post('/keys', { 
      key: newKey.value, 
      platform: 'minimax',
      remark: newRemark.value
    })
    newKey.value = ''
    newRemark.value = ''
    fetchKeys()
  } catch (e) {
    alert(t('keys.alertAddFail'))
  } finally {
    loading.value = false
  }
}

const setDefault = async (id) => {
  try {
    await api.put(`/keys/${id}/default`)
    fetchKeys()
  } catch (e) {
    alert(t('keys.alertSetDefaultFail'))
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
  <div class="page">
    <header class="header">
      <h1>{{ t('keys.title') }}</h1>
      <p class="subtitle">{{ t('keys.subtitle') }}</p>
    </header>

    <div class="card add-key-card">
      <div class="input-stack">
        <input 
          v-model="newKey" 
          type="text" 
          :placeholder="t('keys.placeholder')" 
          class="custom-input"
        />
        <div class="input-row">
          <input 
            v-model="newRemark" 
            type="text" 
            :placeholder="t('keys.placeholderRemark')" 
            class="custom-input flex-2"
            maxlength="100"
          />
          <button @click="addKey" :disabled="loading" class="btn btn-primary">
            <Plus size="18" /> {{ t('keys.add') }}
          </button>
        </div>
      </div>
    </div>

    <div class="keys-list">
      <div v-for="key in keys" :key="key.id" class="key-item card">
        <div class="key-info">
          <KeyIcon class="icon" />
          <div class="details">
            <div class="key-meta">
              <span class="platform">{{ key.platform }}</span>
              <span v-if="key.remark" class="remark-text">{{ key.remark }}</span>
            </div>
            <code class="key-value">{{ key.key.substring(0, 8) }}...{{ key.key.substring(key.key.length - 4) }}</code>
          </div>
        </div>
        
        <div class="actions">
          <span v-if="key.is_default" class="badge-default">
            <Star size="14" fill="currentColor" /> {{ t('keys.default') }}
          </span>
          <button v-else @click="setDefault(key.id)" class="btn-sm btn-outline">
            {{ t('keys.setDefault') }}
          </button>
          
          <button @click="deleteKey(key.id)" class="btn-icon delete">
            <Trash2 size="18" />
          </button>
        </div>
      </div>
      
      <div v-if="keys.length === 0" class="empty-state">
        {{ t('keys.noKeys') }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  max-width: 800px;
  margin: 0 auto;
}

.header {
  margin-bottom: var(--space-6);
}

.subtitle {
  color: var(--text-secondary);
  margin-top: var(--space-2);
}

.add-key-card {
  margin-bottom: var(--space-6);
  padding: var(--space-6);
}

.input-stack {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.input-row {
  display: flex;
  gap: var(--space-4);
}

.custom-input {
  width: 100%;
  padding: var(--space-3) var(--space-4);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  transition: all var(--transition-fast);
}

.custom-input:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.flex-2 {
  flex: 2;
}

.keys-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.key-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-4) var(--space-6);
  transition: all var(--transition-fast);
}

.key-item:hover {
  border-color: var(--primary-light);
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
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
  gap: 4px;
}

.key-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.platform {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
}

.remark-text {
  font-size: 0.875rem;
  color: var(--text-primary);
}

.key-value {
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.875rem;
  color: var(--text-secondary);
  width: fit-content;
}

.actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.badge-default {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  color: var(--primary);
  background: rgba(var(--primary-rgb), 0.1);
  padding: 4px 8px;
  border-radius: 12px;
  font-weight: 500;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 0.75rem;
  border-radius: 4px;
  cursor: pointer;
}

.btn-outline {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}

.btn-outline:hover {
  border-color: var(--text-primary);
  color: var(--text-primary);
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
