<template>
  <div class="txt-box">
    <div class="txt-header">
      <span>txt</span>
      <span class="copy" @click="()=>{copy && copycb()}">
        <span class="el-icon-copy-document"></span>
        Copy
      </span>
    </div>
    <div class="txt" :loading="loading" style="width: 100%; height: 100%;white-space: pre-wrap;">{{ textContent }}</div>
  </div>
</template>
<script>
import axios from 'axios'
export default {
  name: 'DocPeview',
  components: {},
  data() {
    return {
      fileUrl: "",
      loading: true,
      textContent:""
    }
  },
  watch: {
    $route: {
      deep: true,
      handler: function (to) {
        if (to.query) {
          let fileUrl = to.query.fileUrl
          this.fileUrl = fileUrl
        }
      },
      immediate: true
    }
  },
  created() { },
  mounted() {
    this.$nextTick(() => {
      this.getPreview()
    })
  },
  methods: {
    copy(){
      text = this.textContent.replaceAll('<br/>','\n')
      var textareaEl = document.createElement('textarea');
      textareaEl.setAttribute('readonly', 'readonly');
      textareaEl.value = text;
      document.body.appendChild(textareaEl);
      textareaEl.select();
      var res = document.execCommand('copy');
      document.body.removeChild(textareaEl);
      return res;
    },
    copycb(){
      this.$message.success(this.$t('agent.copyTips'))
    },
    async getPreview() {
      this.loading = true
      axios({
        method: 'get',
        responseType: "text",
        url: this.fileUrl,
        headers: {
          'Accept': 'text/plain; charset=utf-8'
       }
      }).then(({ data }) => {
        this.$nextTick(() => {
          this.textContent = data
          this.loading = false
        })
      })
    },
    renderedHandler() {
      console.log('渲染完成')
    },
    errorHandler() {
      console.log('渲染Failed')
    }
  }
}
</script>
<style lang="scss" scoped>
.txt-box {
  padding:50px 20px 0 20px;
  border-radius:6px;
  height: 100%;
  overflow-y: auto;
  background: #f5f5f5;
  .txt-header{
    width:100%;
    height:40px;
    position: fixed;
    padding:0 20px;
    top:0;
    left: 0;
    background: #e5e5e5;
    border-top-left-radius: 6px;
    border-top-right-radius:6px;
    display: flex;
    justify-content:space-between;
    align-items:center;
    .copy{
      cursor: pointer;
    }
  }
}
</style>
