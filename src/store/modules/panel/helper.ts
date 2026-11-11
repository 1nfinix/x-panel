import { ss } from '@/utils/storage'
import { PanelPanelConfigStyleEnum, PanelStateNetworkModeEnum } from '@/enums'
import defaultBackground from '@/assets/defaultBackground.webp'
const LOCAL_NAME = 'panelStorage'

const defaultFooterHtml = '<div class="flex justify-center text-muted-soft" style="margin-top:24px">Powered By <a href="https://github.com/1nfinix/x-panel" target="_blank" class="ml-1">X-Panel</a></div>'

export function defaultStatePanelConfig(): Panel.panelConfig {
  return {
    backgroundImageSrc: defaultBackground,
    backgroundBlur: 0,
    backgroundMaskNumber: 0.55,
    autoWallpaper: true,
    showWallpaper: false,
    iconStyle: PanelPanelConfigStyleEnum.icon,
    iconTextColor: '#ffffff',
    iconTextInfoHideDescription: false,
    iconTextIconHideTitle: false,
    logoText: 'X-Panel',
    logoImageSrc: '',
    clockShowSecond: false,
    clockShow: false,
    maxWidth: 1200,
    maxWidthUnit: 'px',
    marginX: 5,
    footerHtml: defaultFooterHtml,
    systemMonitorShow: false,
    systemMonitorShowTitle: true,
    systemMonitorPublicVisitModeShow: false,
    netModeChangeButtonShow: true,

  }
}

export function defaultState(): Panel.State {
  return {
    rightSiderCollapsed: false,
    leftSiderCollapsed: false,
    networkMode: PanelStateNetworkModeEnum.wan,
    panelConfig: { ...defaultStatePanelConfig() },
  }
}

export function getLocalState(): Panel.State {
  const localState = ss.get(LOCAL_NAME)
  const state = { ...defaultState(), ...localState }
  state.panelConfig = withoutLegacyPageMargins(state.panelConfig)
  return state
}

export function withoutLegacyPageMargins(config: Panel.panelConfig): Panel.panelConfig {
  const current = { ...config } as Panel.panelConfig & { marginTop?: number; marginBottom?: number }
  delete current.marginTop
  delete current.marginBottom
  return current
}

export function setLocalState(state: Panel.State) {
  ss.set(LOCAL_NAME, state)
}

export function removeLocalState() {
  ss.remove(LOCAL_NAME)
}
