<script setup lang='ts'>
import { ref } from 'vue'
import { t } from '@/locales'
defineProps<{
  src: string
}>()

const emit = defineEmits<{
  (event: 'click'): void
  (event: 'refresh'): void
}>()

const randCode = ref<string>('0')

function handleClick() {
  randCode.value = String(rand(100, 99999))
  emit('click')
}

function rand(min: number, max: number) {
  return Math.floor(Math.random() * (max - min)) + min
}

defineExpose({
  // 刷新验证码
  refresh() {
    handleClick()
  },
})
</script>

<template>
  <!-- <div> -->
  <button type="button" :aria-label="t('login.refreshCaptcha')" @keydown.enter.stop @click="handleClick">
    <img :src="`${src}?${randCode}`" :alt="t('login.captchaPlaceholder')">
  </button>
  <!-- </div> -->
</template>

<style scoped>
button { display: block; width: 100%; height: 100%; padding: 0; border: 0; background: transparent; cursor: pointer; }
button:focus-visible { outline: 3px solid var(--x-focus); outline-offset: -3px; }
</style>
