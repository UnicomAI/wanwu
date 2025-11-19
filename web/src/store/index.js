import Vue from 'vue'
import Vuex from 'vuex'
import VuexPersistence from 'vuex-persist'
import { login } from './module/login'
import { user } from './module/user'
import { app } from './module/app'
import { workflow } from './module/workflow'


Vue.use(Vuex)
// UserInformation持久化
const vuexLocal = new VuexPersistence({
    key:'access_cert',
    storage: window.localStorage,
    modules: ['user']
})
//Knowledge BaseSelect AllPermission持久化
const permissionLocal = new VuexPersistence({
    key:'permission_data',
    storage: window.localStorage,
    modules: ['app'],
    reducer:(state) => {
        return {
            app:{
                permissionType: state.app.permissionType
            }
        }
    },
    filter:(mutation) => {
        return mutation.type === 'app/SET_PERMISSION_TYPE' || 
               mutation.type === 'app/CLEAR_PERMISSION_TYPE'
    },
    restoreState:(key,storage) => {
        const userData = localStorage.getItem('access_cert')
        if (!userData) {
            return {}
        }
        const savedData = storage.getItem(key)
        if (!savedData) {
            return {}
        }
        try {
            const parsed = JSON.parse(savedData)
            return parsed.app || {}
        } catch(e) {
            return {}
        }
    }
})

export const store = new Vuex.Store({
    modules: {
        login,
        user,
        app,
        workflow,
    },
    plugins: [vuexLocal.plugin, permissionLocal.plugin]
})
