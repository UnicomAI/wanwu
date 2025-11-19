import Vue from 'vue'
import VueI18n from 'vue-i18n'
import enLocale from 'element-ui/lib/locale/lang/en'
import zhLocale from 'element-ui/lib/locale/lang/zh-CN'
import en from './en'
import zh from './zh'
import {ZH} from "@/lang/constants"

Vue.use(VueI18n)

const messages = {
    zh: {
        language: '简体中文',
        ...zh,
        ...zhLocale
    },
    en: {
        language: 'English',
        ...en,
        ...enLocale
    },
}
const i18n = new VueI18n({
    locale: localStorage.getItem('locale') || 'en', // Default to English, language type stored in localStorage
    messages
})

// 导出 messages 给切换语言的时候用
export { i18n, messages }
