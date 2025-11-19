<template>
  <div ref="wordRef" id="fileShow" class="container" :loading="loading" style="width: 100%; height: 100%;overflow-y:auto;padding:10px;"></div>
</template>
<script>
import axios from 'axios'
import mammoth from 'mammoth';
export default {
  name: 'DocPeview',
  components: {},
  data() {
    return {
      // fileUrl: "http://192.168.0.190:9000/xt0097/cd3379be4f474b47ac9301115e908c491739172844961.docx"
      fileUrl: "",
      loading: true
    }
  },
  watch: {
    $route: {
      deep: true,
      handler: function (to) {
        if (to.query) {
          let fileUrl = to.query.fileUrl
          this.fileUrl = fileUrl
          this.getDocPreview();
        }
      },
      immediate: true
    }
  },
  created() { },
  mounted() {
    this.$nextTick(() => {
      this.getDocPreview()
    })
  },
  methods: {
    async getDocPreview() {
      this.loading = true
      try{
        const response  = await axios({
          method: 'get',
          responseType: 'blob', // SettingResponseFileFormat
          url: this.fileUrl
        })

        const blob = response.data;
        const arrayBuffer = await blob.arrayBuffer();
        // Use mammoth 将 .docx 转为 HTML
        const { value: html } = await mammoth.convertToHtml({ arrayBuffer });
        // Render到Page
        this.$nextTick(() => {
          this.$refs.wordRef.innerHTML = html; 
          this.loading = false;
        });
      }catch(error){
         this.$nextTick(() => {
            this.$refs.wordRef.innerHTML = `<p style="color: red;">预览Failed: ${error.message}</p>`;
            this.loading = false;
        });
      }
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
<style lang="scss">
ocx-preview p {
  margin: 1em 0;
}
.docx-preview h1, .docx-preview h2, .docx-preview h3 {
  color: #333;
}
.docx-preview table {
  border-collapse: collapse;
  width: 100%;
}
.docx-preview td, .docx-preview th {
  border: 1px solid #ddd;
  padding: 8px;
}
.docx-preview img {
  max-width: 100%;
  height: auto;
}
</style>
