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
        language: 'Simplified Chinese',
        ...zh,
        ...zhLocale
    },
    en: {
        language: 'English',
        ...en,
        ...enLocale
    },
}
// Force English as default - initialize localStorage if not set or if set to Chinese
const savedLocale = localStorage.getItem('locale')
if (!savedLocale || savedLocale === 'zh') {
    localStorage.setItem('locale', 'en')
}

const i18n = new VueI18n({
    locale: 'en', // Always default to English
    messages
})

// Export messages for language switching
export { i18n, messages }
