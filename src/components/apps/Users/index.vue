<script lang="ts" setup>
import { h, onMounted, reactive, ref } from 'vue'
import { NAlert, NButton, NDataTable, NDropdown, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns, PaginationProps } from 'naive-ui'
import EditUser from './EditUser/index.vue'
import { getPublicVisitUser, setPublicVisitUser, deletes as usersDeletes, getList as usersGetList } from '@/api/panel/users'
import { SvgIcon } from '@/components/common'
import { useAuthStore } from '@/store'
import { t } from '@/locales'
import { AdminAuthRole } from '@/enums/admin'

const message = useMessage()
const authStore = useAuthStore()
const tableIsLoading = ref<boolean>(false)
const editUserDialogShow = ref<boolean>(false)
const keyWord = ref<string>()
const editUserUserInfo = ref<User.Info>()
const dialog = useDialog()
const publicVisitUserId = ref<number | null>(null)
const operationError = ref('')
const listError = ref('')
const publicModeSaving = ref(false)

async function changePublicVisitUser(id: number | null) {
  if (publicModeSaving.value)
    return
  publicModeSaving.value = true
  operationError.value = ''
  try {
    const { code, msg } = await setPublicVisitUser(id)
    if (code !== 0)
      throw new Error(msg || t('common.saveFail'))
    publicVisitUserId.value = id
    message.success(t('common.saveSuccess'))
  }
  catch (error) {
    operationError.value = error instanceof Error ? error.message : t('common.saveFail')
  }
  finally {
    publicModeSaving.value = false
  }
}

const createColumns = ({
  update,
}: {
  update: (row: User.Info) => void
}): DataTableColumns<User.Info> => {
  return [
    {
      title: t('common.username'),
      key: 'username',
      render(row: User.Info) {
        let publicVisitHtml = ''
        if (publicVisitUserId.value && publicVisitUserId.value === row.id)
          publicVisitHtml = `[${t('adminSettingUsers.pblicText')}]-`

        if (row.username === authStore.userInfo?.username)
          return `${publicVisitHtml}${row.username} (${t('adminSettingUsers.currentUseUsername')})`
        return publicVisitHtml + row.username
      },
    },
    {
      title: t('common.nikeName'),
      key: 'name',
    },
    {
      title: t('adminSettingUsers.role'),
      key: 'role',
      render(row) {
        switch (row.role) {
          case AdminAuthRole.admin:
            return h(NTag, { type: 'default' }, t('common.role.admin'))
          case AdminAuthRole.regularUser:
            return h(NTag, t('common.role.regularUser'))
          default:
            return '-'
        }
      },
    },
    {
      title: t('common.action'),
      key: '',
      render(row) {
        const btn = h(
          NButton,
          {
            'strong': true,
            'tertiary': true,
            'size': 'small',
            'aria-label': `${t('common.action')}: ${row.username}`,
          },
          {
            default() {
              return h(
                SvgIcon, {
                  icon: 'mingcute:more-1-fill',
                },
              )
            },
          },
        )

        return h(NDropdown, {
          trigger: 'click',
          onSelect(key: string | number) {
            switch (key) {
              case 'update':
                update(row)
                break
              case 'publicMode':
                changePublicVisitUser(publicVisitUserId.value === row.id ? null : row.id as number)
                break
              case 'delete':
              {
                const deleteError = ref('')
                const confirmDialog = dialog.warning({
                  title: t('common.warning'),
                  content: () => h('div', [h('p', t('adminSettingUsers.deletePromptContent', { name: row.name, username: row.username })), deleteError.value ? h('p', { style: 'color:var(--x-red-fg);margin-top:12px' }, deleteError.value) : null]),
                  positiveText: t('common.confirm'),
                  negativeText: t('common.cancel'),
                  onPositiveClick: async () => {
                    confirmDialog.loading = true
                    deleteError.value = ''
                    try {
                      const { code, msg } = await usersDeletes([row.id as number])
                      if (code !== 0) {
                        deleteError.value = msg || t('common.deleteFail')
                        return false
                      }
                      message.success(t('common.deleteSuccess'))
                      await getList(null)
                    }
                    catch {
                      deleteError.value = t('common.deleteFail')
                      return false
                    }
                    finally {
                      confirmDialog.loading = false
                    }
                  },
                })
                break
              }

              default:
                break
            }
          },
          options: [
            {
              label: t('common.edit'),
              key: 'update',
            },
            {
              label: t('adminSettingUsers.setOrUnsetPublicMode'),
              key: 'publicMode',
            },
            {
              label: t('common.delete'),
              key: 'delete',
            },
          ],
        }, { default: () => btn })
      },
    },
  ]
}

const userList = ref<User.Info[]>()

const columns = createColumns({
  update(row: User.Info) {
    editUserUserInfo.value = row
    editUserDialogShow.value = true
  },
})
const pagination = reactive({
  page: 1,
  showSizePicker: true,
  pageSizes: [10, 30, 50, 100, 200],
  pageSize: 10,
  itemCount: 0,
  onChange: (page: number) => {
    pagination.page = page
    getList(null)
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    getList(null)
  },
  prefix(item: PaginationProps) {
    return t('adminSettingUsers.userCountText', { count: item.itemCount })
  },
})

function handlePageChange(page: number) {
  getList(page)
}

// 添加
function handleAdd() {
  editUserDialogShow.value = true
  editUserUserInfo.value = {}
}

function handelDone() {
  editUserDialogShow.value = false
  message.success(t('common.success'))
  getList(null)
}

async function getList(page: number | null) {
  tableIsLoading.value = true
  const req: AdminUserManage.GetListRequest = {
    page: page || pagination.page,
    limit: pagination.pageSize,
  }
  if (keyWord.value !== '')
    req.keyWord = keyWord.value

  try {
    const { code, data, msg } = await usersGetList<Common.ListResponse<User.Info[]>>(req)
    if (code !== 0)
      throw new Error(msg || t('common.failed'))
    pagination.itemCount = data.count
    userList.value = data.list ?? []
    listError.value = ''
  }
  catch (error) {
    listError.value = error instanceof Error ? error.message : t('common.failed')
  }
  finally {
    tableIsLoading.value = false
  }
}

onMounted(() => {
  getPublicVisitUser<User.Info | undefined>().then(({ code, data }) => {
    if (code === 0 && data)
      publicVisitUserId.value = data.id || null
    else if (code === -1)
      publicVisitUserId.value = null
  }).catch(() => {
    publicVisitUserId.value = null
  })
  getList(null)
})
</script>

<template>
  <div class="users-panel overflow-auto pt-2">
    <p v-if="operationError" class="operation-error" role="alert">
      {{ operationError }}
    </p>
    <p v-if="listError" class="operation-error" role="alert">
      {{ listError }} <NButton size="small" @click="getList(null)">
        {{ t('panelHome.retry') }}
      </NButton>
    </p>
    <NAlert type="info" :bordered="false">
      {{ $t('adminSettingUsers.alertText') }}
    </NAlert>
    <div class="my-[10px]">
      <NButton type="primary" size="small" ghost @click="handleAdd">
        {{ $t('common.add') }}
      </NButton>
    </div>

    <NDataTable
      :columns="columns"
      :data="userList"
      :pagination="pagination"
      :bordered="false"
      :loading="tableIsLoading"
      :remote="true"

      @update:page="handlePageChange"
    />
    <EditUser v-model:visible="editUserDialogShow" :user-info="editUserUserInfo" @done="handelDone" />
  </div>
</template>

<style scoped>
.users-panel { padding: 16px; border-radius: var(--x-r-small); background: var(--x-bg1); }
.operation-error { margin-bottom: 12px; color: var(--x-red-fg); font-size: 13px; }
</style>
