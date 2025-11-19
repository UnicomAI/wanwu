import { registerMicroApps } from 'qiankun'
import { store } from "@/store"
import { basePath } from "@/utils/config"

export const useQiankun = () => {
  const apps = [
      {
          name: 'test', // 子AppName
          entry: process.env.NODE_ENV === 'development'?'http://localhost:8081/' : window.location.origin + basePath + '/sub/test/', // 子App入口
          container: '#container', // 子App所在容器
          props: () => ({ user: store.state.user }), // 传参给子App
          activeRule: basePath + '/aibase/portal/test', // 子AppTrigger规Then（Path）
      },
  ]

  registerMicroApps(apps, {
    beforeLoad: [
      app => {
        console.log(`${app.name} beforeLoad phase`)
      }
    ],
    beforeMount: [
      app => {
        console.log(`${app.name} beforeMount phase`)
      }
    ],
    afterMount: [
      app => {
        console.log(`${app.name} afterMount phase`)
      }
    ],
    beforeUnmount: [
      app => {
        console.log(`${app.name} beforeUnmount phase`)
      }
    ],
    afterUnmount: [
      app => {
        console.log(`${app.name} afterUnmount phase`)
      }
    ]
  })
}