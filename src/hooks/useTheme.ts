import type { GlobalThemeOverrides } from 'naive-ui'
import { darkTheme, useOsTheme } from 'naive-ui'
import { computed, watch } from 'vue'
import { useAppStore } from '@/store'

// Naive consumes the same CSS tokens as the app; html.dark switches both themes.
const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#0071E3',
    primaryColorHover: '#0062C4',
    primaryColorPressed: '#00539F',
    primaryColorSuppl: '#0062C4',
    bodyColor: 'var(--x-bg0)',
    cardColor: 'var(--x-bg1)',
    modalColor: 'var(--x-bg0)',
    popoverColor: 'var(--x-bg1)',
    inputColor: 'var(--x-bg1)',
    textColorBase: 'var(--x-text1)',
    textColor1: 'var(--x-text1)',
    textColor2: 'var(--x-text2)',
    textColor3: 'var(--x-text2)',
    borderColor: 'var(--x-sep)',
    dividerColor: 'var(--x-sep)',
    borderRadius: '999px',
    heightMedium: '44px',
    fontFamily: 'Onest, -apple-system, "PingFang SC", "Helvetica Neue", "Microsoft YaHei", sans-serif',
    fontSize: '15px',
  },
  Button: {
    heightMedium: '44px',
    borderRadiusMedium: '999px',
    border: 'none',
    borderHover: 'none',
    borderPressed: 'none',
    borderFocus: 'none',
    borderDisabled: 'none',
    borderPrimary: 'none',
    borderHoverPrimary: 'none',
    borderPressedPrimary: 'none',
    borderFocusPrimary: 'none',
    color: 'var(--x-bg1)',
    colorHover: 'var(--x-hover)',
    colorPressed: 'var(--x-ph)',
    colorSecondary: 'var(--x-bg1)',
    colorSecondaryHover: 'var(--x-hover)',
    textColor: 'var(--x-text1)',
    colorPrimary: 'var(--x-primary)',
    colorHoverPrimary: 'var(--x-primary-hover)',
    colorPressedPrimary: 'var(--x-primary-active)',
    textColorPrimary: 'var(--x-on-primary)',
    textColorHoverPrimary: 'var(--x-on-primary)',
    textColorPressedPrimary: 'var(--x-on-primary)',
  },
  Input: {
    heightMedium: '44px',
    borderRadius: '999px',
    color: 'var(--x-bg1)',
    colorFocus: 'var(--x-bg1)',
    border: 'none',
    borderHover: 'none',
    borderFocus: 'none',
    boxShadowFocus: '0 0 0 2px var(--x-primary), 0 0 0 5px var(--x-focus)',
  },
  Select: {
    peers: {
      InternalSelection: {
        heightMedium: '44px',
        borderRadius: '999px',
        color: 'var(--x-bg1)',
        border: 'none',
        borderHover: 'none',
        borderFocus: 'none',
        boxShadowFocus: '0 0 0 2px var(--x-primary), 0 0 0 5px var(--x-focus)',
      },
    },
  },
  Card: { borderRadius: '32px', borderColor: 'transparent', boxShadow: 'none', color: 'var(--x-bg1)' },
  Modal: { color: 'var(--x-bg0)', boxShadow: 'none' },
  Dialog: { borderRadius: '24px', border: 'none', color: 'var(--x-bg0)' },
  Popover: { borderRadius: '24px', color: 'var(--x-bg1)', boxShadow: 'none' },
  Dropdown: { borderRadius: '24px', color: 'var(--x-bg1)', optionColorHover: 'var(--x-bg2)', optionColorActive: 'var(--x-blue-tint)', optionTextColorActive: 'var(--x-blue-fg)' },
  Tooltip: { borderRadius: '999px', color: 'var(--x-inverse-bg)', textColor: 'var(--x-inverse-fg)', boxShadow: 'none' },
  Tag: { borderRadius: '999px' },
  Message: {
    color: 'var(--x-inverse-bg)',
    colorInfo: 'var(--x-inverse-bg)',
    colorSuccess: 'var(--x-inverse-bg)',
    colorError: 'var(--x-inverse-bg)',
    colorWarning: 'var(--x-inverse-bg)',
    colorLoading: 'var(--x-inverse-bg)',
    textColor: 'var(--x-inverse-fg)',
    textColorInfo: 'var(--x-inverse-fg)',
    textColorSuccess: 'var(--x-inverse-fg)',
    textColorError: 'var(--x-inverse-fg)',
    textColorWarning: 'var(--x-inverse-fg)',
    textColorLoading: 'var(--x-inverse-fg)',
    iconColorSuccess: 'var(--x-green-sys)',
    iconColorError: 'var(--x-red-sys)',
    borderRadius: '999px',
    boxShadow: 'none',
  },
  Notification: { color: 'var(--x-inverse-bg)', textColor: 'var(--x-inverse-fg)', iconColorSuccess: 'var(--x-green-sys)', iconColorError: 'var(--x-red-sys)', boxShadow: 'none', borderRadius: '999px' },
}

export function useTheme() {
  const appStore = useAppStore()
  const osTheme = useOsTheme()
  const isDark = computed(() => appStore.theme === 'auto' ? osTheme.value === 'dark' : appStore.theme === 'dark')
  const theme = computed(() => isDark.value ? darkTheme : undefined)
  watch(isDark, value => document.documentElement.classList.toggle('dark', value), { immediate: true })
  return { theme, themeOverrides }
}
