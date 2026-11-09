<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { get } from '@/api/system/about'
import srcSvglogo from '@/assets/logo.svg'

interface Version {
  versionName: string
  versionCode: number
}

const versionName = ref('')

onMounted(() => {
  get<Version>().then((res) => {
    if (res.code === 0)
      versionName.value = res.data.versionName
  })
})
</script>

<template>
  <div class="about-panel">
    <img :src="srcSvglogo" width="100" height="100" alt="X-Panel">
    <h2>{{ $t('common.appName') }}</h2>
    <p v-if="versionName" class="version">
      v{{ versionName }}
    </p>
    <p>{{ $t('apps.about.maintainer') }}<a href="https://github.com/1nfinix" target="_blank" rel="noopener noreferrer">1nfinix</a></p>
    <p>{{ $t('apps.about.repository') }}<a href="https://github.com/1nfinix/x-panel" target="_blank" rel="noopener noreferrer">1nfinix/x-panel</a></p>
    <p><a href="https://github.com/hslr-s/sun-panel" target="_blank" rel="noopener noreferrer">{{ $t('apps.about.basedOn') }}</a></p>
  </div>
</template>

<style scoped>
.about-panel { display: flex; flex-direction: column; align-items: center; padding: 28px 16px; color: #f2f7f4; text-align: center; }
.about-panel h2 { margin: 12px 0 0; font-size: 28px; font-weight: 700; }
.about-panel p { margin: 8px 0; }
.about-panel .version { color: #9fb1a8; }
.about-panel a { color: #76d99b; text-underline-offset: 3px; }
.about-panel a:focus-visible { outline: 2px solid #76d99b; outline-offset: 3px; }
</style>
