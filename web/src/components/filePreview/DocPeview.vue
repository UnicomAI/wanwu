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
          responseType: 'blob', // Set response file format
          url: this.fileUrl
        })

        const blob = response.data;
        const arrayBuffer = await blob.arrayBuffer();
        // Use mammoth to convert .docx into HTML
        // Render to page
        // RendertoPage
        this.$nextTick(() => {
          this.$refs.wordRef.innerHTML = html; 
          this.loading = false;
        });
      }catch(error){
         this.$nextTick(() => {
            this.$refs.wordRef.innerHTML = `<p style="color: red;">Preview failed: ${error.message}</p>`;
            this.loading = false;
        });
      }
    },
    renderedHandler() {
      console.log('Rendering complete')
    },
    errorHandler() {
      console.log('Rendering failed')
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
