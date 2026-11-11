import { h, ref } from 'vue'
import type { useDialog } from 'naive-ui'
import { t } from '@/locales'

type Dialog = ReturnType<typeof useDialog>

export function confirmDelete(dialog: Dialog, options: {
  object: string
  name: string
  detail?: string
  onDelete: () => Promise<string | void>
}) {
  const error = ref('')
  let deleting = false
  const confirmDialog = dialog.error({
    class: 'delete-confirm',
    title: t('common.deleteObject', { object: options.object }),
    content: () => h('div', [
      h('p', t('common.deleteNamed', { name: options.name })),
      options.detail ? h('p', { class: 'delete-confirm-detail' }, options.detail) : null,
      error.value ? h('p', { class: 'delete-confirm-error', role: 'alert' }, error.value) : null,
    ]),
    positiveText: t('common.delete'),
    negativeText: t('common.cancel'),
    positiveButtonProps: { type: 'error' },
    negativeButtonProps: { secondary: true, ghost: false },
    onPositiveClick: async () => {
      if (deleting)
        return false
      deleting = true
      error.value = ''
      confirmDialog.loading = true
      try {
        const failure = await options.onDelete()
        if (failure) {
          error.value = failure
          return false
        }
      }
      catch (cause) {
        error.value = cause instanceof Error ? cause.message : t('common.deleteFail')
        return false
      }
      finally {
        deleting = false
        confirmDialog.loading = false
      }
    },
  })
}
