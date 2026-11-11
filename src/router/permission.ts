import type { Router } from 'vue-router'
import { useUserStore } from '@/store/modules/user'
import { useAuthStore } from '@/store/modules/auth'
import { getAuthInfo } from '@/api/system/user'

export function setupPageGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    if (to.path === '/' && !useAuthStore().token) {
      try {
        const response = await getAuthInfo(true)
        if (response.code !== 0) {
          next('/login')
          return
        }
      }
      catch {
        next('/login')
        return
      }
    }
    // const authStore = useAuthStoreWithout()
    const userStore = useUserStore()
    // 非管理员路由拦截
    if (userStore.userInfo.role !== 1 && to.path.includes('admin'))
      next({ name: '404' })

    else
      next()
  })
}
