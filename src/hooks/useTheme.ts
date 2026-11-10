import type { GlobalThemeOverrides } from 'naive-ui'
import { darkTheme, useOsTheme } from 'naive-ui'
import { computed, watch } from 'vue'
import { useAppStore } from '@/store'

// Keep these literal colors in sync with the matching --x-* tokens in src/styles/global.less.
// Naive UI derives colors with seemly and cannot parse CSS custom properties.
export const themeColors = {
  light: {
    bg0: '#F1F3F6',
    modal: '#F9FAFB',
    bg1: '#F9FAFB',
    bg2: '#FFFFFF',
    hover: '#EBEEF2',
    ph: '#E3E7EC',
    text1: '#1D1D1F',
    text2: '#636368',
    sep: '#E3E7EC',
    primary: '#0071E3',
    primaryHover: '#0062C4',
    primaryActive: '#00539F',
    onPrimary: '#FFFFFF',
    focus: 'rgba(0, 113, 227, .35)',
    inverseBg: '#1D1D1F',
    inverseFg: '#FFFFFF',
    blueTint: '#E6F1FC',
    blueFg: '#0062C4',
    greenSys: '#34C759',
    redSys: '#FF3B30',
  },
  dark: {
    bg0: '#1B1C1F',
    modal: '#242529', // dark overlays sit one step above the page
    bg1: '#242529',
    bg2: '#2F3035',
    hover: '#3A3B41',
    ph: '#3A3B41',
    text1: '#F5F5F7',
    text2: '#ADADB2',
    sep: '#3A3B41',
    primary: '#0071E3',
    primaryHover: '#0062C4',
    primaryActive: '#00539F',
    onPrimary: '#FFFFFF',
    focus: 'rgba(10, 132, 255, .55)',
    inverseBg: '#F5F5F7',
    inverseFg: '#1D1D1F',
    blueTint: '#18314B',
    blueFg: '#6CB4FF',
    greenSys: '#30D158',
    redSys: '#FF453A',
  },
} as const

export function useThemeColors() {
  const appStore = useAppStore()
  const osTheme = useOsTheme()
  const isDark = computed(() => appStore.theme === 'auto' ? osTheme.value === 'dark' : appStore.theme === 'dark')
  const colors = computed(() => isDark.value ? themeColors.dark : themeColors.light)
  return { isDark, colors }
}

export function resolveNaiveColor(value: string | undefined, colors: typeof themeColors.light | typeof themeColors.dark, fallback: string) {
  if (!value)
    return fallback
  const legacyTokens: Record<string, string> = {
    bg0: colors.bg0,
    bg1: colors.bg1,
    bg2: colors.bg2,
    text1: colors.text1,
    text2: colors.text2,
    sep: colors.sep,
    primary: colors.primary,
  }
  const match = /^var\(--x-([a-z0-9-]+)\)$/.exec(value)
  if (match)
    return legacyTokens[match[1]] || fallback
  for (const token of ['bg2', 'text1', 'sep'] as const) {
    if (value.toLowerCase() === themeColors.light[token].toLowerCase() || value.toLowerCase() === themeColors.dark[token].toLowerCase())
      return colors[token]
  }
  return /^(#(?:[0-9a-f]{3}|[0-9a-f]{4}|[0-9a-f]{6}|[0-9a-f]{8})$|rgba?\(|hsla?\()/i.test(value) ? value : fallback
}

export function useTheme() {
  const { isDark, colors } = useThemeColors()
  const theme = computed(() => isDark.value ? darkTheme : undefined)
  const themeOverrides = computed<GlobalThemeOverrides>(() => {
    const c = colors.value
    const focusShadow = `0 0 0 2px ${c.primary}, 0 0 0 5px ${c.focus}`
    return {
      common: {
        primaryColor: c.primary,
        primaryColorHover: c.primaryHover,
        primaryColorPressed: c.primaryActive,
        primaryColorSuppl: c.primaryHover,
        bodyColor: c.bg0,
        cardColor: c.bg1,
        modalColor: c.modal,
        popoverColor: c.bg1,
        inputColor: c.bg1,
        textColorBase: c.text1,
        textColor1: c.text1,
        textColor2: c.text2,
        textColor3: c.text2,
        borderColor: c.sep,
        dividerColor: c.sep,
        borderRadius: '12px',
        heightMedium: '44px',
        fontFamily: 'Onest, -apple-system, "PingFang SC", "Helvetica Neue", "Microsoft YaHei", sans-serif',
        fontSize: '15px',
      },
      Button: {
        heightMedium: '44px',
        borderRadiusTiny: '999px',
        borderRadiusSmall: '999px',
        borderRadiusMedium: '999px',
        borderRadiusLarge: '999px',
        border: 'none',
        borderHover: 'none',
        borderPressed: 'none',
        borderFocus: 'none',
        borderDisabled: 'none',
        borderPrimary: 'none',
        borderHoverPrimary: 'none',
        borderPressedPrimary: 'none',
        borderFocusPrimary: 'none',
        color: c.bg1,
        colorHover: c.hover,
        colorPressed: c.ph,
        colorSecondary: c.bg1,
        colorSecondaryHover: c.hover,
        textColor: c.text1,
        colorPrimary: c.primary,
        colorHoverPrimary: c.primaryHover,
        colorPressedPrimary: c.primaryActive,
        textColorPrimary: c.onPrimary,
        textColorHoverPrimary: c.onPrimary,
        textColorPressedPrimary: c.onPrimary,
        // Text-style primary buttons are links: use the link tone so they pass 4.5:1 on dark cards.
        textColorTextPrimary: c.blueFg,
        textColorTextHoverPrimary: c.blueFg,
        textColorTextPressedPrimary: c.blueFg,
        textColorTextFocusPrimary: c.blueFg,
      },
      Input: {
        heightMedium: '44px',
        borderRadius: '12px',
        color: c.bg1,
        colorFocus: c.bg1,
        border: 'none',
        borderHover: 'none',
        borderFocus: 'none',
        boxShadowFocus: focusShadow,
      },
      Select: {
        peers: {
          // The menu is a panel, not a capsule: keep it on the small-card radius.
          InternalSelectMenu: { borderRadius: '12px', color: c.bg1, boxShadow: 'var(--x-shadow-overlay)', optionColorPending: c.hover, optionTextColorActive: c.blueFg, optionCheckColor: c.blueFg },
          InternalSelection: {
            heightMedium: '44px',
            borderRadius: '12px',
            color: c.bg1,
            border: 'none',
            borderHover: 'none',
            borderFocus: 'none',
            boxShadowFocus: focusShadow,
          },
        },
      },
      Card: { borderRadius: '20px', borderColor: c.bg1, boxShadow: 'none', color: c.bg1 },
      Modal: { color: c.modal, boxShadow: 'var(--x-shadow-overlay)' },
      Dialog: { borderRadius: '20px', border: 'none', color: c.modal, boxShadow: 'var(--x-shadow-overlay)' },
      Popover: { borderRadius: '12px', color: c.bg1, boxShadow: 'var(--x-shadow-overlay)' },
      Dropdown: { borderRadius: '12px', color: c.bg1, boxShadow: 'var(--x-shadow-overlay)', optionColorHover: c.bg2, optionColorActive: c.blueTint, optionTextColorActive: c.blueFg },
      Tooltip: { borderRadius: '999px', color: c.inverseBg, textColor: c.inverseFg, boxShadow: 'var(--x-shadow-overlay)' },
      Tag: { borderRadius: '999px' },
      // Panels and multi-line surfaces must not inherit the 999px capsule radius from common.
      ColorPicker: { borderRadius: '12px' },
      Alert: { borderRadius: '12px' },
      DataTable: { borderRadius: '12px' },
      Upload: { borderRadius: '12px' },
      Message: {
        color: c.inverseBg,
        colorInfo: c.inverseBg,
        colorSuccess: c.inverseBg,
        colorError: c.inverseBg,
        colorWarning: c.inverseBg,
        colorLoading: c.inverseBg,
        textColor: c.inverseFg,
        textColorInfo: c.inverseFg,
        textColorSuccess: c.inverseFg,
        textColorError: c.inverseFg,
        textColorWarning: c.inverseFg,
        textColorLoading: c.inverseFg,
        iconColorSuccess: c.greenSys,
        iconColorError: c.redSys,
        borderRadius: '999px',
        boxShadow: 'var(--x-shadow-overlay)',
      },
      Notification: {
        color: c.inverseBg,
        textColor: c.inverseFg,
        iconColorSuccess: c.greenSys,
        iconColorError: c.redSys,
        boxShadow: 'var(--x-shadow-overlay)',
        borderRadius: '999px',
      },
    }
  })
  watch(isDark, value => document.documentElement.classList.toggle('dark', value), { immediate: true })
  return { theme, themeOverrides }
}
