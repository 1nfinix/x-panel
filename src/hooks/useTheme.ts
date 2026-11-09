import type { GlobalThemeOverrides } from 'naive-ui'
import { computed } from 'vue'
import { darkTheme } from 'naive-ui'

export function useTheme() {
  const theme = computed(() => darkTheme)

  const themeOverrides = computed<GlobalThemeOverrides>(() => {
    return {
      common: {
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
      },
    }
  })

  document.documentElement.classList.add('dark')

  return { theme, themeOverrides }
}
