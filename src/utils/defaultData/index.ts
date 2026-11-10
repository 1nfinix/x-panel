import type { Language } from '@/store/modules/app/helper'

import { groupTints } from '@/utils/groupColors'

export const defautSwatchesBackground = ['#00000000', '#000000', '#ffffff', ...groupTints]

export const languageOptions: { label: string; key: Language; value: Language }[] = [
  { label: 'English', key: 'en-US', value: 'en-US' },
  { label: '简体中文', key: 'zh-CN', value: 'zh-CN' },
]
