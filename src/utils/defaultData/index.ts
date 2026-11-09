import type { Language } from '@/store/modules/app/helper'

export const defautSwatchesBackground = [
  '#00000000',
  '#000000',
  '#ffffff',
  '#5db872',
  '#5db8a6',
  '#e8a55a',
  '#c64545',
  '#cc785c',
]

export const languageOptions: { label: string; key: Language; value: Language }[] = [
  { label: 'English', key: 'en-US', value: 'en-US' },
  { label: '简体中文', key: 'zh-CN', value: 'zh-CN' },
]
