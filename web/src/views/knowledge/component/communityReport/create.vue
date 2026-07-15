<template>
  <el-dialog
    :title="title"
    :visible.sync="dialogVisible"
    width="50%"
    :before-close="handleClose"
  >
    <el-form
      :model="ruleForm"
      ref="ruleForm"
      label-width="120px"
      class="demo-ruleForm"
    >
      <el-form-item class="itemCenter">
        <el-radio-group
          v-model="createType"
          @input="typeChange($event)"
          v-if="type === 'add'"
        >
          <el-radio-button :label="'single'">
            {{ $t('knowledgeManage.create.single') }}
          </el-radio-button>
          <el-radio-button :label="'file'">
            {{ $t('knowledgeManage.create.file') }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item
        :label="$t('knowledgeManage.create.file')"
        v-if="createType === 'file'"
        prop="fileUploadId"
        :rules="[
          {
            required: true,
            message: $t('common.input.placeholder'),
            trigger: 'blur',
          },
        ]"
      >
        <fileUpload
          ref="fileUpload"
          :templateUrl="templateUrl"
          @uploadFile="uploadFile"
          :accept="accept"
        />
      </el-form-item>
      <template v-if="createType === 'single'">
        <el-form-item
          :label="$t('knowledgeManage.create.title')"
          prop="title"
          :rules="[
            {
              required: true,
              message: $t('knowledgeManage.create.titlePlaceholder'),
              trigger: 'blur',
            },
          ]"
        >
          <el-input
            :placeholder="$t('knowledgeManage.create.titlePlaceholder')"
            v-model="ruleForm.title"
            maxlength="100"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('knowledgeManage.create.content')"
          prop="content"
          :rules="[
            {
              required: true,
              message: $t('knowledgeManage.create.contentPlaceholder'),
              trigger: 'blur',
            },
          ]"
        >
          <el-input
            :placeholder="$t('knowledgeManage.create.contentPlaceholder')"
            v-model="ruleForm.content"
            type="textarea"
            :rows="6"
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('knowledgeManage.create.typeTitle')"
          v-if="type === 'add'"
        >
          <el-checkbox-group v-model="checkType">
            <el-checkbox label="more" name="type">
              {{ $t('knowledgeManage.create.continue') }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </template>
    </el-form>
    <span slot="footer" class="dialog-footer">
      <el-button @click="dialogVisible = false">
        {{ $t('common.confirm.cancel') }}
      </el-button>
      <el-button
        type="primary"
        @click="submit('ruleForm')"
        :loading="btnLoading"
      >
        {{ $t('common.confirm.confirm') }}
      </el-button>
    </span>
  </el-dialog>
</template>
<script>
import {
  createBatchCommunityReport,
  createCommunityReport,
  editCommunityReportList,
} from '@/api/knowledge';
import fileUpload from '@/components/fileUpload';
import {
  createSegment,
  createBatchSegment,
  createSegmentChild,
} from '@/api/knowledge';

export default {
  components: { fileUpload },
  data() {
    return {
      btnLoading: false,
      accept: '.csv',
      checkType: [],
      createType: 'single',
      ruleForm: {
        content: '',
        knowledgeId: '',
        title: '',
        fileUploadId: '',
        contentId: '',
      },
      type: 'add',
      dialogVisible: false,
      templateUrl: '/user/api/v1/files/docs/report.csv',
      title: '',
    };
  },
  methods: {
    typeChange(val) {
      if (val === 'single') {
        this.ruleForm.fileUploadId = '';
        this.$refs.fileUpload && this.$refs.fileUpload.clearFileList();
      } else {
        this.clearForm();
        this.$refs.ruleForm && this.$refs.ruleForm.clearValidate();
      }
    },
    uploadFile(fileUploadId) {
      this.ruleForm.fileUploadId = fileUploadId;
    },
    handleClose() {
      this.dialogVisible = false;
    },
    showDialog(knowledgeId, type = '', item = null) {
      this.type = type;
      this.createType = 'single';
      if (type === 'edit') {
        this.title = this.$t('knowledgeManage.communityReport.viewDetail');
        this.dialogVisible = true;
        this.ruleForm.knowledgeId = knowledgeId;
        if (item) {
          this.ruleForm.content = item.content || '';
          this.ruleForm.title = item.title || '';
          this.ruleForm.contentId = item.contentId || '';
        } else {
          this.clearForm();
        }
      } else {
        this.title = this.$t(
          'knowledgeManage.communityReport.addCommunityReport',
        );
        this.dialogVisible = true;
        this.ruleForm.knowledgeId = knowledgeId;
        this.clearForm();
      }
    },
    submit(formName) {
      if (this.createType === 'single') {
        this.handleSingle(formName);
      } else {
        this.handleFile();
      }
    },
    handleSingle(formName) {
      this.$refs[formName].validate(valid => {
        if (valid) {
          this.btnLoading = true;
          if (this.type === 'add') {
            this.createCommunityReport();
          } else {
            this.editCommunityReportList();
          }
        } else {
          return false;
        }
      });
    },
    editCommunityReportList() {
      const data = {
        content: this.ruleForm.content,
        knowledgeId: this.ruleForm.knowledgeId,
        title: this.ruleForm.title,
        contentId: this.ruleForm.contentId,
      };
      editCommunityReportList(data)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(
              this.$t('knowledgeManage.create.editSuccess'),
            );
            this.clearForm();
            this.$emit('refreshData');
            this.dialogVisible = false;
            this.btnLoading = false;
          }
        })
        .catch(() => {
          this.btnLoading = false;
        });
    },
    createCommunityReport() {
      const data = {
        content: this.ruleForm.content,
        knowledgeId: this.ruleForm.knowledgeId,
        title: this.ruleForm.title,
      };
      createCommunityReport(data)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(
              this.$t('knowledgeManage.create.createSuccess'),
            );
            if (!this.checkType.length) {
              this.dialogVisible = false;
            } else {
              this.clearForm();
            }
            this.$emit('refreshData');
            this.btnLoading = false;
          }
        })
        .catch(() => {
          this.btnLoading = false;
        });
    },
    handleFile() {
      this.btnLoading = true;
      const data = {
        fileUploadId: this.ruleForm.fileUploadId,
        knowledgeId: this.ruleForm.knowledgeId,
      };
      createBatchCommunityReport(data)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(
              this.$t('knowledgeManage.create.createSuccess'),
            );
            this.dialogVisible = false;
            this.btnLoading = false;
            this.$emit('refreshData');
          }
        })
        .catch(() => {
          this.btnLoading = false;
        });
    },
    clearForm() {
      this.ruleForm.content = '';
      this.ruleForm.title = '';
      this.ruleForm.fileUploadId = '';
      this.ruleForm.contentId = '';
      this.$refs.fileUpload && this.$refs.fileUpload.clearFileList();
      this.checkType = [];
    },
  },
};
</script>
<style lang="scss" scoped>
.itemCenter {
  display: flex;
  justify-content: center;

  ::v-deep .el-form-item__content {
    margin-left: 0 !important;
  }
}

.el-tag {
  margin-right: 5px;
  color: #3848f7;
  border-color: #3848f7;
  background: $color_opacity;
}

::v-deep {
  .el-tag .el-tag__close {
    color: #3848f7 !important;
  }

  .el-tag .el-tag__close:hover {
    color: #fff !important;
    background: #3848f7;
  }

  .el-checkbox__input.is-checked + .el-checkbox__label {
    color: #3848f7;
  }
}
</style>
