import router from './index'
import { store } from '@/store/index'
import { fetchPermFirPath } from '@/utils/util'
import { PERMS as menuPerms } from "./constants"
import { basePath } from "@/utils/config"

const white_list = [basePath + '/aibase', '/oauth', '/login', '/webChat', '/register', '/reset', '/templateSquare']
export const PERMS = menuPerms

export const checkPerm = (perm) => {
    // 不传Permission点，Table示不NeedPermission控制，Back true
    if (!perm) return true
    // 传Permission点，CheckYesNo在PermissionList，在Back true，NoThen false
    const permission = store.getters['user/permission']
    const orgPermission = permission.orgPermission
    if (orgPermission && orgPermission.length) {
        return Array.isArray(perm)
            ? perm.some(item => orgPermission.includes(item))
            : orgPermission.includes(perm)
    }
    return false
}

export const formatPerms = (perms) => {
    return perms && perms.length ? perms.map(item => item.perm) : []
}

router.beforeEach(async (to, from, next) => {
  let token = ''
  let access_cert = localStorage.getItem("access_cert") && JSON.parse(localStorage.getItem("access_cert"))
  if(access_cert){
      token = access_cert.user.token
  }
  if (token && !access_cert.user.is2FA) {
      if (to.path === '/') {
          const {path} = fetchPermFirPath()
          next({path})
      } else {
          next()
      }
  } else {
      if(white_list.some(item =>{ return to.path.indexOf(item)>-1 })){
          next()
        }else{
          window.location.href = window.location.origin + basePath + '/aibase/login'
        }
  }
})
