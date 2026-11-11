<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { t } from '@/locales'
import { useThemeColors } from '@/hooks/useTheme'

const props = defineProps<{
  src: string
}>()

const emit = defineEmits<{
  (event: 'click'): void
  (event: 'refresh'): void
}>()

const { isDark } = useThemeColors()
const refreshCount = ref(0)
const imageSrc = computed(() => `${props.src}${props.src.includes('?') ? '&' : '?'}theme=${isDark.value ? 'dark' : 'light'}&r=${refreshCount.value}`)

function handleClick() {
  refreshCount.value++
  emit('click')
}

watch(isDark, handleClick)

defineExpose({
  // 刷新验证码
  refresh() {
    handleClick()
  },
})
</script>

<template>
  <button type="button" :aria-label="t('login.refreshCaptcha')" @keydown.enter.stop @click="handleClick">
    <img :src="imageSrc" :alt="t('login.captchaLabel')">
  </button>
</template>

<style scoped>
button { display: block; width: 100%; height: 100%; padding: 0; border: 0; background: transparent; cursor: pointer; }
button:focus-visible { outline: 3px solid var(--x-focus); outline-offset: -3px; }
</style>
