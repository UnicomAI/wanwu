<template>
  <div v-if="langOptions && langOptions.length">
    <el-dropdown trigger="click" style="margin-left: 22px; color: #fff; cursor: pointer" @command="handleChangeLang">
      <span>
        {{langValue}} <i class="el-icon-arrow-down"></i>
      </span>
      <el-dropdown-menu slot="dropdown">
        <el-dropdown-item
          v-for="(item, index) in langOptions"
          :key="item.code + index"
          :command="item"
        >
          {{item.name}}
        </el-dropdown-item>
      </el-dropdown-menu>
    </el-dropdown>
  </div>
</template>

<script>
import { getLangList, changeLang } from "@/api/user"
import { ZH } from "@/lang/constants"

export default {
  props: {
    isLogin: false
  },
  data() {
    return {
      langValue: '',
      langOptions: [],
    }
  },
  mounted() {
    this.getLangList()
  },
  methods: {
    getLangList() {
      getLangList().then(res => {
        const {languages, defaultLanguage} = res.data || {}
        this.langOptions = languages || []
        const langCode = localStorage.getItem("locale")

        // If no default language is cached locally, store the backend default language and reload
        if (!langCode) {
          window.localStorage.setItem('locale', defaultLanguage.code || ZH)
          window.location.reload()
          return
        }

        const lang = languages.find(item => item.code === langCode)
        this.setLangValue(lang || {})
      })
    },
    setLangValue(item) {
      window.localStorage.setItem('locale', item.code)
      this.langValue = item.name
      this.$store.state.user.lang = item.code
      this.$i18n.locale = item.code
    },
    async handleChangeLang(item) {
      if (!this.isLogin) {
        await changeLang({language: item.code})
        window.location.reload()
      }
      this.setLangValue(item)
    }
  }
}
</script>

<style lang="scss" scoped>

</style>
