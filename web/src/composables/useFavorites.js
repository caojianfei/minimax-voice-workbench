import { ref } from 'vue'
import axios from 'axios'

const favorites = ref(new Set())
let initialized = false
let initPromise = null

const api = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'
})

const ensureLoaded = () => {
  if (initialized) return initPromise
  initialized = true
  initPromise = api.get('/favorites')
    .then((res) => {
      const ids = res?.data?.data
      if (Array.isArray(ids)) {
        favorites.value = new Set(ids)
      }
    })
    .catch((e) => {
      console.error('Failed to load favorites', e)
    })
  return initPromise
}

export function useFavorites() {
  void ensureLoaded()

  const refreshFavorites = async () => {
    const res = await api.get('/favorites')
    const ids = res?.data?.data
    favorites.value = new Set(Array.isArray(ids) ? ids : [])
  }

  const toggleFavorite = async (voiceId) => {
    if (!voiceId) return

    const wasFavorite = favorites.value.has(voiceId)
    if (wasFavorite) favorites.value.delete(voiceId)
    else favorites.value.add(voiceId)
    favorites.value = new Set(favorites.value)

    try {
      const res = await api.post(`/favorites/${encodeURIComponent(voiceId)}/toggle`)
      const nowFavorite = res?.data?.data?.favorite

      if (typeof nowFavorite === 'boolean') {
        if (nowFavorite) favorites.value.add(voiceId)
        else favorites.value.delete(voiceId)
        favorites.value = new Set(favorites.value)
        return
      }

      await refreshFavorites()
    } catch (e) {
      if (wasFavorite) favorites.value.add(voiceId)
      else favorites.value.delete(voiceId)
      favorites.value = new Set(favorites.value)
      console.error('Failed to toggle favorite', e)
    }
  }

  const isFavorite = (voiceId) => favorites.value.has(voiceId)

  return {
    favorites,
    refreshFavorites,
    toggleFavorite,
    isFavorite
  }
}
