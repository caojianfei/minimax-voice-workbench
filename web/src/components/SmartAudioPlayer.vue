<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  src: {
    type: String,
    required: true
  },
  format: {
    type: String,
    default: 'mp3'
  },
  sampleRate: {
    type: Number,
    default: 32000
  },
  channels: {
    type: Number,
    default: 1
  }
})

const audioUrl = ref('')
const error = ref('')
const loading = ref(false)
const flacRepairTried = ref(false)

const setObjectUrl = (blob) => {
  if (audioUrl.value && audioUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(audioUrl.value)
  }
  audioUrl.value = URL.createObjectURL(blob)
}

const findSubarrayIndex = (haystack, needle) => {
  if (!haystack || !needle || needle.length === 0) return -1
  for (let i = 0; i <= haystack.length - needle.length; i++) {
    let ok = true
    for (let j = 0; j < needle.length; j++) {
      if (haystack[i + j] !== needle[j]) {
        ok = false
        break
      }
    }
    if (ok) return i
  }
  return -1
}

const writeString = (view, offset, string) => {
  for (let i = 0; i < string.length; i++) {
    view.setUint8(offset + i, string.charCodeAt(i))
  }
}

const addWavHeader = (samples, sampleRate, numChannels, bitDepth) => {
  const buffer = new ArrayBuffer(44 + samples.byteLength)
  const view = new DataView(buffer)

  /* RIFF identifier */
  writeString(view, 0, 'RIFF')
  /* RIFF chunk length */
  view.setUint32(4, 36 + samples.byteLength, true)
  /* RIFF type */
  writeString(view, 8, 'WAVE')
  /* format chunk identifier */
  writeString(view, 12, 'fmt ')
  /* format chunk length */
  view.setUint32(16, 16, true)
  /* sample format (raw) */
  view.setUint16(20, 1, true)
  /* channel count */
  view.setUint16(22, numChannels, true)
  /* sample rate */
  view.setUint32(24, sampleRate, true)
  /* byte rate (sample rate * block align) */
  view.setUint32(28, sampleRate * numChannels * (bitDepth / 8), true)
  /* block align (channel count * bytes per sample) */
  view.setUint16(32, numChannels * (bitDepth / 8), true)
  /* bits per sample */
  view.setUint16(34, bitDepth, true)
  /* data chunk identifier */
  writeString(view, 36, 'data')
  /* data chunk length */
  view.setUint32(40, samples.byteLength, true)

  // Copy PCM data
  new Uint8Array(buffer, 44).set(new Uint8Array(samples))

  return buffer
}

const processAudio = async () => {
  if (!props.src) return
  
  error.value = ''
  flacRepairTried.value = false
  
  // If not PCM, use direct URL
  // Note: Modern browsers support FLAC. If FLAC fails, we might need a decoder, 
  // but usually it works if the server sends correct MIME or if the browser sniffs it.
  if (props.format !== 'pcm') {
    audioUrl.value = props.src
    return
  }

  // Handle PCM
  loading.value = true
  try {
    const response = await fetch(props.src)
    if (!response.ok) throw new Error('Failed to fetch audio')
    
    const arrayBuffer = await response.arrayBuffer()
    
    // Assume 16-bit PCM (Minimax usually outputs 16-bit)
    const wavBuffer = addWavHeader(
      arrayBuffer, 
      props.sampleRate || 32000, 
      props.channels || 1, 
      16
    )
    
    const blob = new Blob([wavBuffer], { type: 'audio/wav' })
    setObjectUrl(blob)
  } catch (e) {
    console.error('PCM conversion failed', e)
    error.value = 'Failed to load audio'
  } finally {
    loading.value = false
  }
}

const tryRepairFlac = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await fetch(props.src)
    if (!response.ok) throw new Error('Failed to fetch audio')

    const arrayBuffer = await response.arrayBuffer()
    const u8 = new Uint8Array(arrayBuffer)
    const flacMagic = new Uint8Array([0x66, 0x4c, 0x61, 0x43]) // fLaC
    const idx = findSubarrayIndex(u8, flacMagic)

    if (idx < 0) {
      throw new Error('FLAC header not found')
    }

    const fixed = arrayBuffer.slice(idx)
    setObjectUrl(new Blob([fixed], { type: 'audio/flac' }))
  } catch (e) {
    console.error('FLAC repair failed', e)
    error.value = 'Failed to load audio'
  } finally {
    loading.value = false
  }
}

const onAudioError = async () => {
  if (props.format !== 'flac') return
  if (flacRepairTried.value) return
  flacRepairTried.value = true
  await tryRepairFlac()
}

watch(() => props.src, processAudio)

onMounted(processAudio)

onUnmounted(() => {
  if (audioUrl.value && audioUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(audioUrl.value)
  }
})
</script>

<template>
  <div class="smart-player">
    <div v-if="loading" class="player-loading">Loading audio...</div>
    <div v-else-if="error" class="player-error">{{ error }}</div>
    <audio v-else controls :src="audioUrl" class="audio-element" @error="onAudioError"></audio>
  </div>
</template>

<style scoped>
.smart-player {
  width: 100%;
  display: flex;
  align-items: center;
}

.audio-element {
  width: 100%;
  height: 32px;
}

.player-loading {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.player-error {
  font-size: 0.8rem;
  color: var(--error);
}
</style>
