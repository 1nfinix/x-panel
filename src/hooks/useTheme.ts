import type { GlobalThemeOverrides } from 'naive-ui'
import { darkTheme, useOsTheme } from 'naive-ui'
import { computed, watch } from 'vue'
import { useAppStore } from '@/store'

// Keep these values in sync with the CSS custom properties in global.less.
const palette = {
  light: {
    canvas: '#f5f3ee',
    surface: '#ffffff',
    surfaceSoft: '#eeeae3',
    surfaceCard: '#eae7e0',
    hairline: '#e2ded7',
    ink: '#141413',
    body: '#3d3d3a',
    muted: '#6c6a64',
    primary: '#cc785c',
    primaryActive: '#a9583e',
    focusRing: 'rgba(204, 120, 92, .15)',
  },
  dark: {
    canvas: '#151412',
    surface: '#1e1d1a',
    surfaceSoft: '#24231f',
    surfaceCard: '#292824',
    hairline: '#2e2c28',
    ink: '#faf9f5',
    body: '#e3dfd7',
    muted: '#a09d96',
    primary: '#cc785c',
    primaryActive: '#d98d73',
    focusRing: 'rgba(204, 120, 92, .28)',
  },
}

const fontBody = 'Inter, -apple-system, BlinkMacSystemFont, "PingFang SC", "Hiragino Sans GB", "Segoe UI", "Microsoft YaHei", sans-serif'

export function useTheme() {
  const appStore = useAppStore()
  const osTheme = useOsTheme()
  const isDark = computed(() => appStore.theme === 'auto' ? osTheme.value === 'dark' : appStore.theme === 'dark')
  const theme = computed(() => isDark.value ? darkTheme : undefined)

  const themeOverrides = computed<GlobalThemeOverrides>(() => {
    const colors = isDark.value ? palette.dark : palette.light
    return {
      common: {
        primaryColor: colors.primary,
        primaryColorHover: colors.primaryActive,
        primaryColorPressed: colors.primaryActive,
        primaryColorSuppl: colors.primary,
        bodyColor: colors.canvas,
        cardColor: colors.surface,
        modalColor: colors.surface,
        popoverColor: colors.surface,
        tableColor: colors.surface,
        inputColor: colors.surface,
        textColorBase: colors.ink,
        textColor1: colors.ink,
        textColor2: colors.body,
        textColor3: colors.muted,
        borderColor: colors.hairline,
        dividerColor: colors.hairline,
        borderRadius: '8px',
        borderRadiusSmall: '6px',
        fontFamily: fontBody,
        fontSize: '14px',
      },
      Card: { borderRadius: '12px' },
      Input: {
        border: `1px solid ${colors.hairline}`,
        borderFocus: `1px solid ${colors.primary}`,
        borderHover: `1px solid ${colors.primary}`,
        boxShadowFocus: `0 0 0 3px ${colors.focusRing}`,
      },
      Select: {
        peers: {
          InternalSelection: {
            border: `1px solid ${colors.hairline}`,
            borderHover: `1px solid ${colors.primary}`,
            borderFocus: `1px solid ${colors.primary}`,
            boxShadowFocus: `0 0 0 3px ${colors.focusRing}`,
          },
        },
      },
      Alert: {
        colorInfo: colors.surfaceCard,
        borderInfo: `1px solid ${colors.hairline}`,
        iconColorInfo: colors.ink,
        titleTextColorInfo: colors.ink,
        contentTextColorInfo: colors.body,
        colorWarning: colors.surfaceCard,
        borderWarning: `1px solid ${colors.hairline}`,
        iconColorWarning: colors.primary,
        contentTextColorWarning: colors.body,
      },
      Button: {
        fontWeightStrong: '500',
        textColor: colors.ink,
        color: colors.surface,
        border: `1px solid ${colors.hairline}`,
        colorHover: colors.surface,
        textColorHover: colors.ink,
        borderHover: `1px solid ${colors.primary}`,
        colorSecondary: colors.surface,
        colorSecondaryHover: colors.surfaceSoft,
        colorSecondaryPressed: colors.surfaceCard,
        colorPrimary: colors.primary,
        colorHoverPrimary: colors.primaryActive,
        textColorPrimary: '#ffffff',
        textColorHoverPrimary: '#ffffff',
        textColorPressedPrimary: '#ffffff',
      },
      Tag: { borderRadius: '9999px' },
      Dialog: { borderRadius: '12px' },
      Modal: { borderRadius: '12px' },
      Menu: { itemColorActive: colors.surfaceCard, itemTextColorActive: colors.ink },
      Dropdown: { optionColorActive: colors.surfaceCard, optionTextColorActive: colors.ink },
    }
  })

  watch(isDark, value => document.documentElement.classList.toggle('dark', value), { immediate: true })

  return { theme, themeOverrides }
}
