import { ref, watch } from 'vue'

const STORAGE_KEY = 'minimax_voice_favorites'

const favorites = ref(new Set())
let initialized = false

export function useFavorites() {
  if (!initialized) {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      try {
        const parsed = JSON.parse(stored)
        if (Array.isArray(parsed)) {
          favorites.value = new Set(parsed)
        }
      } catch (e) {
        console.error('Failed to parse favorites', e)
      }
    }
    initialized = true

    // Sync across tabs/windows
    window.addEventListener('storage', (e) => {
      if (e.key === STORAGE_KEY && e.newValue) {
        try {
          const parsed = JSON.parse(e.newValue)
          favorites.value = new Set(parsed)
        } catch (e) {}
      }
    })
  }

  // Watch for changes and save
  watch(favorites, (newVal) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(Array.from(newVal)))
  }, { deep: true })

  const toggleFavorite = (voiceId) => {
    if (favorites.value.has(voiceId)) {
      favorites.value.delete(voiceId)
    } else {
      favorites.value.add(voiceId)
    }
    // Trigger reactivity for Set
    favorites.value = new Set(favorites.value)
  }

  const isFavorite = (voiceId) => {
    return favorites.value.has(voiceId)
  }

  return {
    favorites,
    toggleFavorite,
    isFavorite
  }
}
