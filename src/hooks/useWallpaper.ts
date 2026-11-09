import { ref } from 'vue'

export function useWallpaper() {
  const source = ref('')
  const loaded = ref(false)
  let requestId = 0

  async function show(fallback?: string, automatic = true) {
    const currentRequest = ++requestId
    loaded.value = false
    let chosen = fallback || ''
    if (automatic) {
      try {
        const response = await fetch('/api/openness/wallpapers')
        const result: { code: number; data?: string[] } = await response.json()
        if (result.code === 0 && result.data?.length)
          chosen = result.data[Math.floor(Math.random() * result.data.length)]
      }
      catch {
        // The user's uploaded wallpaper remains the fallback.
      }
    }
    if (currentRequest !== requestId)
      return
    if (!chosen) {
      source.value = ''
      return
    }
    const image = new Image()
    image.onload = () => {
      if (currentRequest === requestId) {
        source.value = chosen
        loaded.value = true
      }
    }
    image.onerror = () => {
      if (currentRequest === requestId && chosen !== fallback && fallback)
        show(fallback, false)
    }
    image.src = chosen
  }

  return { source, loaded, show }
}
