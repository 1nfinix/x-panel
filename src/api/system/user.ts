import { post } from '@/utils/request'

// export function getInfo<T>() {
//   return post<T>({
//     url: '/user/getInfo',
//   })
// }

export function getAuthInfo<T>(skipAuthRedirect = false) {
  return post<T>({
    url: '/user/getAuthInfo',
    skipAuthRedirect,
  })
}

export function getReferralCode<T>() {
  return post<T>({
    url: '/user/getReferralCode',
  })
}

export function updateInfo<T>(name: string) {
  return post<T>({
    url: '/user/updateInfo',
    data: { name },
  })
}

export function updatePassword<T>(oldPassword: string, newPassword: string) {
  return post<T>({
    url: '/user/updatePassword',
    data: { newPassword, oldPassword },
  })
}
