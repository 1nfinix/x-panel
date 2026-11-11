<script setup lang='ts'>
import { computed, defineAsyncComponent, defineComponent, h, useAttrs } from 'vue'

interface Props {
  icon?: string
}

defineProps<Props>()

const Icon = defineAsyncComponent({
  loader: () => import('@iconify/vue').then(module => module.Icon),
  delay: 0,
  loadingComponent: defineComponent({
    inheritAttrs: false,
    setup(_, { attrs }) {
      return () => h('span', {
        ...attrs,
        'aria-hidden': true,
        'style': [attrs.style, { display: 'inline-block', width: '1em', height: '1em', verticalAlign: '-0.125em' }],
      })
    },
  }),
})

const attrs = useAttrs()

const bindAttrs = computed<{ class: string; style: string }>(() => ({
  class: (attrs.class as string) || '',
  style: (attrs.style as string) || '',
}))
</script>

<template>
  <Icon :icon="icon ?? ''" v-bind="bindAttrs" />
</template>
