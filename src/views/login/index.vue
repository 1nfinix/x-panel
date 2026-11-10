<script setup lang="ts">
import { NButton, NCard, NForm, NFormItem, NInput, NSelect, useMessage } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { login } from '@/api'
import { useAppStore, useAuthStore, usePanelState } from '@/store'
import { SvgIcon, XBrand } from '@/components/common'
import { router } from '@/router'
import { t } from '@/locales'
import { languageOptions } from '@/utils/defaultData'
import type { Language } from '@/store/modules/app/helper'
import Captcha from '@/components/common/Captcha/index.vue'
import { getLoginConfig } from '@/api/openness'
import { useWallpaper } from '@/hooks/useWallpaper'

// const userStore = useUserStore()
const authStore = useAuthStore()
const appStore = useAppStore()
const ms = useMessage()
const loading = ref(false)
const fieldError = ref<{ username?: string; password?: string; vcode?: string }>({})
const languageValue = ref<Language>(appStore.language)

const isShowCaptcha = ref(true)
const captchaRef = ref<InstanceType<typeof Captcha> | null>(null)
const panelState = usePanelState()
const { source: wallpaperSource, loaded: wallpaperLoaded, show: showWallpaper } = useWallpaper()
// const isShowRegister = ref<boolean>(false)

const form = ref<Login.LoginReqest>({
  username: '',
  password: '',
  vcode: '',
})

onMounted(async () => {
  if (panelState.panelConfig.showWallpaper)
    showWallpaper(panelState.panelConfig.backgroundImageSrc, panelState.panelConfig.autoWallpaper !== false)
  try {
    const response = await getLoginConfig<{ loginCaptcha: boolean }>()
    if (response.code === 0)
      isShowCaptcha.value = response.data.loginCaptcha
  }
  catch {
    isShowCaptcha.value = true
  }
})

const loginPost = async () => {
  if (loading.value)
    return
  fieldError.value = {}
  if (!form.value.username) {
    fieldError.value.username = t('form.required')
    return
  }
  if (!form.value.password) {
    fieldError.value.password = t('form.required')
    return
  }
  loading.value = true
  try {
    const res = await login<Login.LoginResponse>(form.value)
    if (res.code === 0) {
      authStore.setToken(res.data.token)
      authStore.setUserInfo(res.data)

      setTimeout(() => {
        ms.success(`Hi ${res.data.name},${t('login.welcomeMessage')}`)
        loading.value = false
        router.push({ path: '/' })
      }, 500)
    }
    else {
      loading.value = false
      const message = res.code === 1101 ? t('login.captchaRequired') : res.code === 1102 ? t('login.captchaInvalid') : res.code === 1003 ? t('login.credentialsInvalid') : res.msg
      if (res.code === 1101 || res.code === 1102)
        fieldError.value.vcode = message
      else
        fieldError.value.password = message
      form.value.vcode = ''
      captchaRef.value?.refresh()
    }
  }
  catch (error) {
    loading.value = false
    fieldError.value.password = t('login.networkError')
    form.value.vcode = ''
    captchaRef.value?.refresh()
  }
}

function handleSubmit() {
  // 点击登录按钮触发
  loginPost()
}

function handleChangeLanuage(value: Language) {
  languageValue.value = value
  appStore.setLanguage(value)
}
</script>

<template>
  <div class="login-container">
    <div v-if="panelState.panelConfig.showWallpaper" class="login-wallpaper" :class="{ loaded: wallpaperLoaded }" :style="{ backgroundImage: wallpaperLoaded ? `url(${wallpaperSource})` : 'none' }" /><div v-if="panelState.panelConfig.showWallpaper" class="wallpaper-mask" />
    <NCard class="login-card">
      <div class="mb-5 flex items-center justify-end">
        <div class="mr-2">
          <SvgIcon icon="ion-language" style="width: 20px;height: 20px;" />
        </div>
        <div class="min-w-[100px]">
          <NSelect v-model:value="languageValue" size="small" :options="languageOptions" :aria-label="t('common.language')" @update-value="handleChangeLanuage" />
        </div>
      </div>

      <div class="login-title">
        <span class="login-brand"><XBrand /></span>
        <strong class="login-heading">
          {{ panelState.panelConfig.logoText || $t('common.appName') }}
        </strong>
      </div>
      <NForm class="login-form" :model="form" label-width="100px" @keydown.enter="handleSubmit">
        <div class="login-fields">
          <NFormItem :show-feedback="!!fieldError.username" :feedback="fieldError.username" :validation-status="fieldError.username ? 'error' : undefined">
            <NInput v-model:value="form.username" :placeholder="$t('login.usernamePlaceholder')" :input-props="{ 'aria-label': $t('login.usernamePlaceholder') }">
              <template #prefix>
                <SvgIcon icon="ph:user-bold" />
              </template>
            </NInput>
          </NFormItem>

          <NFormItem :show-feedback="!!fieldError.password" :feedback="fieldError.password" :validation-status="fieldError.password ? 'error' : undefined">
            <NInput v-model:value="form.password" type="password" :placeholder="$t('login.passwordPlaceholder')" :input-props="{ 'aria-label': $t('login.passwordPlaceholder') }">
              <template #prefix>
                <SvgIcon icon="mdi:password-outline" />
              </template>
            </NInput>
          </NFormItem>

          <NFormItem v-if="isShowCaptcha" :show-feedback="!!fieldError.vcode" :feedback="fieldError.vcode" :validation-status="fieldError.vcode ? 'error' : undefined">
            <div class="captcha-image" :title="t('login.refreshCaptcha')">
              <Captcha ref="captchaRef" src="/api/captcha/getImage" />
            </div>
            <NInput v-model:value="form.vcode" type="text" maxlength="4" :placeholder="t('login.captchaPlaceholder')" :input-props="{ 'aria-label': t('login.captchaPlaceholder') }" />
          </NFormItem>
        </div>
        <div class="login-submit">
          <NButton type="primary" block :loading="loading" @click="handleSubmit">
            {{ $t('login.loginButton') }}
          </NButton>
        </div>

        <!-- <div class="flex justify-end">
          <NButton v-if="isShowRegister" text type="primary" class="flex" @click="$router.push({ path: '/register' })">
            注册
          </NButton>
          <NButton text type="primary" class="flex" @click="$router.push({ path: '/resetPassword' })">
            忘记密码?
          </NButton>
        </div> -->

        <div class="login-footer flex justify-center text-text2 text-[13px]">
          Powered By <a href="https://github.com/1nfinix/x-panel" target="_blank" class="ml-1 text-text2">X-Panel</a>
        </div>
      </NForm>
    </NCard>
  </div>
</template>

<style scoped>
.login-container { position: relative; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 24px; background-color: var(--x-bg0); }
.login-wallpaper { position: absolute; inset: 0; background-position: center; background-size: cover; opacity: 0; transition: opacity 600ms ease; }
.login-wallpaper.loaded { opacity: 1; }
.wallpaper-mask { position: absolute; inset: 0; background: rgba(18,18,20,.55); }
.login-card { position: relative; width: min(100%, 430px); border: 0; border-radius: var(--x-r-card); background: var(--x-bg1); color: var(--x-text1); box-shadow: none; }
.login-card :deep(.n-card__content) { padding: 40px; }
.login-card :deep(.n-input) { min-height: 44px; border-radius: var(--x-r-small); background: var(--x-bg2); }
.login-card :deep(.n-button--primary-type) { height: 44px; border-radius: var(--x-r-pill); }
.login-title { display: flex; align-items: center; justify-content: center; gap: 10px; margin: 16px 0 32px; }
.login-brand { display: grid; place-items: center; width: 44px; height: 44px; border-radius: var(--x-r-icon); background: var(--x-bg2); }.login-brand :deep(svg) { width: 28px; height: 28px; }
.login-fields { display: grid; gap: 16px; }
.login-fields :deep(.n-form-item) { margin: 0; }
.login-submit { margin-top: 24px; }
.login-submit .n-button { width: 100%; }
.login-footer { margin-top: 16px; }
.login-heading { font-family: var(--x-font); font-size: 22px; color: var(--x-text1); font-weight: 600; letter-spacing: -.02em; }
.captcha-image { flex: none; width: 160px; height: 48px; margin-right: 12px; cursor: pointer; border-radius: var(--x-r-icon); overflow: hidden; }
.captcha-image :deep(img) { width: 160px; height: 48px; }
@media (max-width: 390px) { .captcha-image { width: 138px; } .captcha-image :deep(img) { width: 138px; } }
@media (prefers-reduced-motion: reduce) { .login-wallpaper { transition: none; } }
</style>
