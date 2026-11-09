import { ref } from 'vue'
import { fillFavicons } from '@/api/panel/itemIcon'
import { t } from '@/locales'

interface FillResult { id: number; success: boolean; reason?: string; iconUrl?: string }
interface FillFailure { id: number; reason: string }

export function useFaviconFill() {
  const running = ref(false)
  const done = ref(0)
  const total = ref(0)
  const success = ref(0)
  const failures = ref<FillFailure[]>([])

  function reset() {
    done.value = 0
    total.value = 0
    success.value = 0
    failures.value = []
  }

  async function run(ids: number[], onSuccess?: (result: FillResult) => void) {
    if (running.value)
      return
    reset()
    running.value = true
    total.value = ids.length
    try {
      for (const id of ids) {
        try {
          const response = await fillFavicons<FillResult[]>([id])
          const result = response.data?.[0]
          if (response.code === 0 && result?.success) {
            success.value++
            onSuccess?.(result)
          }
          else {
            failures.value.push({ id, reason: result?.reason || response.msg })
          }
        }
        catch {
          failures.value.push({ id, reason: t('apps.baseSettings.fillFailed') })
        }
        done.value++
      }
    }
    finally {
      running.value = false
    }
  }

  return { running, done, total, success, failures, reset, run }
}
