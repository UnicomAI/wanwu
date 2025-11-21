<template>
  <div class="main-content docs-page-content">
    <div class="mark__content">
      <div class="markdown-body" v-html="mdContent"></div>
    </div>
  </div>
</template>
<script>

const highlight = require('highlight.js')
import 'highlight.js/styles/github.css'
import { marked } from 'marked'
import { getMarkdown } from "@/api/docs"
import { DOC_FIRST_KEY } from "../constants"
import { Loading } from 'element-ui'

marked.setOptions({
  renderer: new marked.Renderer(),
  gfm: true,
  tables: true,
  breaks: false,
  pedantic: false,
  sanitize: false,
  smartLists: true,
  smartypants: false,
  highlight: function (code) {
    return highlight.highlightAuto(code).value
  }
})

export default {
  data(){
    return {
      mdContent: '',
    }
  },
  watch: {
    $route: {
      handler (val, oldValue) {
        if (val !== oldValue) {
          this.getMarkdown(val.params.id)
          const docsPageContent = document.querySelector('.el-main')
          if (docsPageContent) docsPageContent.scrollTo(0, 0)
        }
      },
      // 深度ObserveListen
      deep: true
    }
  },
  created() {
    this.getMarkdown(this.$route.params.id)
  },
  methods: {
    docContentScrollTop() {
      const docPageMain = document.querySelector('.doc-page-main')
      if (docPageMain) docPageMain.scrollTop = 0
    },
    getMarkdown(path) {
      if (path === DOC_FIRST_KEY) return

      const loadingInstance = Loading.service({
        target: document.querySelector('.docs-page-content')
      })
      getMarkdown({path}).then(res => {
        const mdContent = marked(res.data || this.$t('common.noData'))
        this.mdContent =  mdContent
        loadingInstance.close()
        this.docContentScrollTop()
      }).catch(() => {
        this.mdContent = this.$t('common.noData')
        loadingInstance.close()
      })
    }
  }
}
</script>

<style lang="scss" scoped>
@import "@/assets/showDocs/showdoc.scss";
.docs-page-content {
  margin: 0 0 50px;
}
</style>
