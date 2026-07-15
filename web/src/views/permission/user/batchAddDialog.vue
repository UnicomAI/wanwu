<template>
  <div>
    <el-dialog
      :title="$t('user.button.batchAdd')"
      :visible.sync="dialogVisible"
      append-to-body
      :close-on-click-modal="false"
      width="500px"
    >
      <el-form
        label-width="100px"
        :model="uploadForm"
        :rules="rules"
        ref="uploadForm"
      >
        <el-form-item :label="$t('user.upload.upload')" prop="file">
          <el-upload
            class="avatar-uploader"
            action=""
            name="file"
            :show-file-list="false"
            :http-request="handleUpload"
            :on-error="handleAvatarError"
            accept=".xlsx"
          >
            <i
              style="font-size: 26px; margin-right: 16px; margin-top: 5px"
              class="el-icon-upload"
            />
            <span>{{ $t('user.upload.hint') }}</span>
            <div style="text-align: left; line-height: normal">
              {{ uploadForm.file ? uploadForm.file.name : '' }}
            </div>
          </el-upload>
          <el-button type="text" @click="downloadTemp">
            {{ $t('user.upload.downloadTemp') }}
          </el-button>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button
          size="small"
          :loading="uploading"
          type="primary"
          @click="handleSubmit"
        >
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
    <el-dialog
      :title="$t('common.confirm.title')"
      :visible.sync="hintVisible"
      append-to-body
      :close-on-click-modal="false"
      width="600px"
    >
      <div style="margin-top: -20px">
        <p>
          <span>{{ $t('user.dialog.total') }}{{ errorData?.total || 0 }}</span>
          <span style="margin-left: 20px">
            {{ $t('user.dialog.failed') }}{{ errorData?.failed || 0 }}
          </span>
          <span style="margin-left: 20px">
            {{ $t('user.dialog.success') }}{{ errorData?.success || 0 }}
          </span>
        </p>
        <el-table :data="errorData?.errors || []">
          <el-table-column
            prop="row"
            :label="$t('user.table.row')"
          ></el-table-column>
          <el-table-column
            prop="username"
            :label="$t('user.table.username')"
          ></el-table-column>
          <el-table-column
            prop="reason"
            :label="$t('user.table.reason')"
          ></el-table-column>
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import { batchCreateUser } from '@/api/permission/user';
import { authDownload } from '@/utils/util';
export default {
  components: { Pagination },
  props: {
    orgId: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      dialogVisible: false,
      uploadForm: {
        file: '',
      },
      rules: {
        file: [
          {
            required: true,
            message: this.$t('uploadDialog.noUpload'),
            trigger: 'change',
          },
        ],
      },
      uploading: false,
      errorData: {},
      hintVisible: false,
    };
  },
  methods: {
    openDialog() {
      this.dialogVisible = true;
    },
    async downloadTemp() {
      try {
        await authDownload(
          this.$basePath + '/user/api/v1/files/docs/users.xlsx',
          'users.xlsx',
        );
      } catch (error) {
        this.$message.error(this.$t('uploadDialog.uploadError'));
      }
    },
    handleUpload(res) {
      if (res.file) {
        this.uploadForm.file = res.file;
        this.$refs.uploadForm.clearValidate('file');
      }
    },
    handleAvatarError() {
      this.$message.error(this.$t('uploadDialog.uploadError'));
    },
    handleClose() {
      this.dialogVisible = false;
      for (let key in this.uploadForm) {
        this.uploadForm[key] = '';
      }
      this.$refs.uploadForm.resetFields();
    },
    handleSubmit() {
      this.$refs.uploadForm.validate(valid => {
        if (valid) {
          const formData = new FormData();
          const config = { headers: { 'Content-Type': 'multipart/form-data' } };
          for (let key in this.uploadForm) {
            formData.append(key, this.uploadForm[key]);
          }
          formData.append('orgId', this.orgId);
          this.uploading = true;
          batchCreateUser(formData, config)
            .then(res => {
              this.uploading = false;
              const { errors } = res.data || {};
              if (errors && errors.length > 0) {
                this.errorData = res.data || {};
                this.hintVisible = true;
                return;
              }
              this.$message.success(this.$t('common.message.success'));
              this.handleClose();
              this.$emit('reloadData');
            })
            .catch(() => (this.uploading = false));
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.avatar-uploader {
  margin-bottom: -12px;
  ::v-deep .el-upload:focus {
    border-color: #606266 !important;
    color: #606266 !important;
  }
}
</style>
