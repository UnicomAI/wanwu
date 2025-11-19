import axios from 'axios'
import { store } from '@/store/index'
import { Message } from 'element-ui'
import { basePath } from "@/utils/config"
import { ZH } from '@/lang/constants'
import { guid, getXClientId } from "@/utils/util"

// create an axios instance
const service = axios.create({
  baseURL: basePath, // url = base url + request url
  timeout: 600000 // request timeout
})

// request interceptor
service.interceptors.request.use(
  config => {
    const token = store.getters['user/token']
    const user = store.getters['user/userInfo']
    const lang = localStorage.getItem('locale') || ZH // store.getters['user/lang']
    let xClientId = getXClientId()
    if (!xClientId) {
      xClientId = guid()
      localStorage.setItem('xClientId', xClientId)
    }
    config.headers = {
    ...config.headers,
    ...(!config.isOpenUrl && {
      'Authorization': 'Bearer ' + token,
      'x-user-id': user.uid,
      'x-org-id': user.orgId,
      'x-client-id': xClientId,
    }),
    ...(config.hasLang && { 'x-language': lang })
    }
      
    return config
  },
  error => {
    console.log(error)
    return Promise.reject(error)
  }
)

// response interceptor
service.interceptors.response.use(
  response => {
    // ExportFile
    const res = response.data
    if (response.config.responseType === 'blob') {
      return res
    }
    if (response.headers['content-type'] === "text/plain; charset=utf-8") {
      return res
    }
    if (response.headers["new-token"]) {
        store.commit("user/setToken", response.headers["new-token"]);
    }
    if (res.code) res.code = res.code * 1
    if (res.code !== 0) {
      Message.error(res.msg || 'Error', 5 * 1000)
      return Promise.reject(res)
    } else {
      return res
    }
  },
  error => {
    // console.log('err：' + error)
    /**
     * 接口异常Process：
     * 1. 401: No权限Back，会重定向到 login
     * 2. Other异常报错：errMessage 中
    **/
    const errRes = error.response || {}
    const { status, statusText } = errRes
    const { msg } = errRes.data || {}
    const errMessage = {
      400: 'Bad Request',
      403: 'Unauthorized',
      404: 'Not Found',
      500: 'Internal Server Error',
      502: 'Internal Server Error',
      503: 'Internal Server Error',
    }
    if (status === 401) {
      Message.error(msg || 'Unauthorized')
      localStorage.removeItem('access_cert')
      store.state.user.token = ''
      window.location.href = window.location.origin + basePath +'/aibase/login'
    } else if (Object.keys(errMessage).includes(String(status))) {
      Message.error(msg || statusText || errMessage[status])
    } else {
      Message.error(error.message || errMessage[500])
    }
    return Promise.reject(error)
  }
)

export default service
