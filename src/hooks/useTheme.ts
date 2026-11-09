import type { GlobalThemeOverrides } from 'naive-ui'
import { darkTheme, useOsTheme } from 'naive-ui'
import { computed, watch } from 'vue'
import { useAppStore } from '@/store'

export function useTheme() {
  const appStore = useAppStore()
  const osTheme = useOsTheme()
  const isDark = computed(() => appStore.theme === 'auto' ? osTheme.value === 'dark' : appStore.theme === 'dark')
  const theme = computed(() => isDark.value ? darkTheme : undefined)

  const themeOverrides = computed<GlobalThemeOverrides>(() => {
    return {
      common: isDark.value
        ? {
            primaryColor: '#76d99b',
            primaryColorHover: '#9aebba',
            primaryColorPressed: '#50b879',
            primaryColorSuppl: '#76d99b',
            bodyColor: '#111827',
            cardColor: '#1b2734',
            modalColor: '#1b2734',
            popoverColor: '#1b2734',
            textColorBase: '#f2f7f4',
            borderRadius: '8px',
          }
        : {
            primaryColor: '#237a49',
            primaryColorHover: '#1d693e',
            primaryColorPressed: '#155632',
            primaryColorSuppl: '#237a49',
            bodyColor: '#F7F5F0',
            cardColor: '#FFFFFF',
            modalColor: '#FFFFFF',
            popoverColor: '#FFFFFF',
            textColorBase: '#111827',
            borderColor: '#D9DEE7',
            borderRadius: '8px',
          },
    }
  })

  watch(isDark, value => document.documentElement.classList.toggle('dark', value), { immediate: true })

  return { theme, themeOverrides }
}
