<template>
  <el-dialog
    :title="$t('docDownload.title')"
    :visible.sync="dialogVisible"
    width="400px"
    append-to-body
    :close-on-click-modal="false"
    :before-close="handleClose"
  >
    <div class="link-text" @click="handleDownload">{{$t('docDownload.name')}}</div>
    <span slot="footer" class="dialog-footer">
      <el-button type="primary" @click="handleDownload">{{$t('common.button.confirm')}}</el-button>
    </span>
  </el-dialog>
</template>

<script>
import { docDownload } from "@/api/user"

export default {
  data() {
    return {
      dialogVisible: false,
    }
  },
  mounted() {
  },
  methods: {
    openDialog() {
      this.dialogVisible = true
    },
    handleClose() {
      this.dialogVisible = false
    },
    handleDownload() {
      docDownload().then(res => {
        if (!(res.data && res.data.docCenterPath)) {
          this.$message.error(res.msg || '暂无下载路径')
          return
        }
        const url = window.location.origin + this.$basePath + '/user/api' + res.data.docCenterPath
        window.open(url)
        this.handleClose()
      })
    }
  }
}
</script>

<style lang="scss" scoped>
.link-text {
  color: $color;
  text-decoration: underline;
  font-size: 14px;
  cursor: pointer;
}
</style>
