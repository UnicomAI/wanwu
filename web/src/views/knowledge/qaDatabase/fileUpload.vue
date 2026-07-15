<template>
  <el-dialog
    :title="$t('knowledgeManage.fileUpload')"
    :visible.sync="dialogVisible"
    :close-on-click-modal="false"
    width="60%"
    :before-close="handleClose"
    class="qa-upload-dialog"
  >
    <el-form
      :model="uploadForm"
      ref="uploadForm"
      label-width="140px"
      class="upload-form"
    >
      <el-form-item>
        <fileUpload
          ref="fileUpload"
          :templateUrl="templateUrl"
          :accept="'.csv'"
          :maxSize="10"
          :maxFileCount="1"
          @uploadFile="handleUploadFile"
        >
          <template #upload-tips>
            <div class="qa-upload-extra">
              <p>
                {{ $t('knowledgeManage.qaDatabase.uploadTips') }}
              </p>
              <p>
                {{ $t('knowledgeManage.qaDatabase.uploadTips1') }}
              </p>
            </div>
          </template>
        </fileUpload>
      </el-form-item>
    </el-form>
    <span slot="footer" class="dialog-footer">
      <el-button @click="handleClose">
        {{ $t('common.confirm.cancel') }}
      </el-button>
      <el-button
        type="primary"
        @click="handleConfirm"
        :loading="loading"
        :disabled="!uploadedFileId"
      >
        {{ $t('knowledgeManage.confirmImport') }}
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import fileUpload from '@/components/fileUpload.vue';
import { qaDocImport } from '@/api/qaDatabase';
import { USER_API } from '@/utils/requestConstants';
import { authDownload } from '@/utils/util';

export default {
  name: 'QaFileUpload',
  components: {
    fileUpload,
  },
  props: {
    knowledgeId: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      dialogVisible: false,
      loading: false,
      uploadedFileId: null,
      uploadForm: {},
      templateUrl: `${USER_API}/files/docs/qa_pair_template.csv`,
    };
  },
  methods: {
    showDialog() {
      this.dialogVisible = true;
      this.uploadedFileId = null;
      // 清除文件列表
      this.$nextTick(() => {
        if (this.$refs.fileUpload) {
          this.$refs.fileUpload.clearFileList();
        }
      });
    },
    handleClose() {
      this.dialogVisible = false;
      this.uploadedFileId = null;
      this.$nextTick(() => {
        if (this.$refs.fileUpload) {
          this.$refs.fileUpload.clearFileList();
        }
      });
    },
    handleUploadFile(...args) {
      const [, , filePath] = args;
      this.uploadedFileId = filePath;
    },
    handleConfirm() {
      if (!this.uploadedFileId) {
        this.$message.warning(this.$t('knowledgeManage.selectFile'));
        return;
      }

      this.loading = true;
      const data = {
        knowledgeId: this.knowledgeId,
        docInfoList: [
          {
            docUrl: this.uploadedFileId,
          },
        ],
      };

      qaDocImport(data)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('app.qaUplodFileTips'));
            this.handleClose();
            this.$emit('updateData', 'fileUpload');
          }
        })
        .catch(() => {})
        .finally(() => {
          this.loading = false;
        });
    },
    async downloadTemplate() {
      try {
        await authDownload(
          `${USER_API}/files/docs/qa_import_template.xlsx`,
          'qa_import_template.xlsx',
        );
      } catch (error) {
        this.$message.error(this.$t('knowledgeManage.create.downloadFailed'));
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.qa-upload-dialog {
  ::v-deep .el-dialog__header {
    padding: 20px 20px 10px;
    border-bottom: 1px solid #f0f0f0;
  }

  ::v-deep .el-dialog__body {
    padding: 20px;
  }

  ::v-deep .el-dialog__footer {
    padding: 10px 20px 20px;
    border-top: 1px solid #f0f0f0;
  }

  ::v-deep .el-form-item__content {
    margin-left: 0 !important;
  }
}

.upload-form {
  ::v-deep .el-form-item__label {
    font-weight: 500;
    color: #333;
  }
}

.upload-tips {
  margin-top: 15px;
  padding: 15px;
  background: #f7f8fa;
  border-radius: 4px;

  p {
    margin: 8px 0;
    font-size: 12px;
    color: #666;
    line-height: 1.5;

    .red {
      color: #f56c6c;
      margin-right: 4px;
    }

    .template-download {
      color: $color;
      font-weight: bold;
      margin-left: 4px;
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;

  .el-button {
    border-radius: 4px;
    padding: 8px 16px;
  }

  .el-button--primary {
    background: $btn_bg;
    border-color: $btn_bg;

    &:hover {
      background: #2a3cc7;
      border-color: #2a3cc7;
    }

    &:disabled {
      background: #c0c4cc;
      border-color: #c0c4cc;
    }
  }
}
</style>
