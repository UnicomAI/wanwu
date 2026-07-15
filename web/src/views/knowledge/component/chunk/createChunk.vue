<template>
  <el-dialog
    :title="$t('knowledgeManage.create.createChunk')"
    :visible.sync="dialogVisible"
    width="40%"
    :before-close="handleClose"
  >
    <el-form
      :model="ruleForm"
      ref="ruleForm"
      label-width="100px"
      class="demo-ruleForm"
    >
      <el-form-item class="itemCenter" v-if="!isChildChunk">
        <el-radio-group v-model="createType" @input="typeChange($event)">
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
        v-if="createType === 'file' && !isChildChunk"
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
      <template v-if="createType === 'single' || isChildChunk">
        <el-form-item
          :label="$t('knowledgeManage.create.chunkContent')"
          prop="content"
          :rules="[
            {
              required: true,
              message: $t('knowledgeManage.create.chunkContentPlaceholder'),
              trigger: 'blur',
            },
          ]"
        >
          <uploadImgMd
            v-if="isMultiModal"
            :placeholder="$t('knowledgeManage.create.chunkContentPlaceholder')"
            v-model="ruleForm.content"
            :knowledgeId="this.ruleForm.knowledgeId"
          ></uploadImgMd>
          <el-input
            v-else
            :placeholder="$t('knowledgeManage.create.chunkContentPlaceholder')"
            v-model="ruleForm.content"
            type="textarea"
            :rows="4"
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('knowledgeManage.create.chunkKeywords')"
          prop="labels"
          v-if="!isChildChunk"
        >
          <el-tag
            :key="tag"
            v-for="(tag, index) in ruleForm.labels"
            closable
            :disable-transitions="false"
            @close="handleTagClose(index)"
          >
            {{ tag }}
          </el-tag>
          <el-input
            class="input-new-tag"
            v-if="inputVisible"
            v-model="inputValue"
            ref="saveTagInput"
            size="small"
            @keyup.enter.native="handleInputConfirm"
            @blur="handleInputConfirm"
          ></el-input>
          <el-button
            v-else
            class="button-new-tag"
            size="small"
            @click="showInput"
          >
            + {{ $t('knowledgeManage.create.chunkKeywords') }}
          </el-button>
        </el-form-item>
        <el-form-item :label="$t('knowledgeManage.create.typeTitle')">
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
import fileUpload from '@/components/fileUpload';
import uploadImgMd from '@/components/uploadImgMd.vue';
import { USER_API } from '@/utils/requestConstants';
import {
  createSegment,
  createBatchSegment,
  createSegmentChild,
} from '@/api/knowledge';
import { MULTIMODAL } from '@/views/knowledge/constants';

export default {
  components: { fileUpload, uploadImgMd },
  props: {
    parentId: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      btnLoading: false,
      accept: '.csv',
      checkType: [],
      inputVisible: false,
      inputValue: '',
      createType: 'single',
      ruleForm: {
        content: '',
        docId: '',
        knowledgeId: '',
        labels: [],
        fileUploadId: '',
      },
      dialogVisible: false,
      templateUrl: `${USER_API}/files/docs/segment.csv`,
      isChildChunk: false,
      category: null,
    };
  },
  computed: {
    isMultiModal() {
      return Number(this.category) === MULTIMODAL;
    },
  },
  methods: {
    typeChange(val) {
      if (this.isChildChunk) {
        this.createType = 'single';
        return;
      }
      if (val === 'single') {
        this.ruleForm.fileUploadId = '';
        this.$refs.fileUpload.clearFileList();
      } else {
        this.clearForm();
        this.$refs.ruleForm.clearValidate();
      }
    },
    uploadFile(fileUploadId) {
      this.ruleForm.fileUploadId = fileUploadId;
    },
    handleClose() {
      this.dialogVisible = false;
    },
    showDialog(docId, knowledgeId, isChildChunk = false, category = null) {
      this.dialogVisible = true;
      this.isChildChunk = isChildChunk;
      this.category = category;
      this.ruleForm.docId = docId;
      this.ruleForm.knowledgeId = knowledgeId;
      this.clearForm();
    },
    showInput() {
      this.inputVisible = true;
      this.$nextTick(_ => {
        this.$refs.saveTagInput.$refs.input.focus();
      });
    },
    handleTagClose(index) {
      this.ruleForm.labels.splice(index, 1);
    },
    handleInputConfirm() {
      if (this.inputValue) {
        this.ruleForm.labels.push(this.inputValue);
        this.inputVisible = false;
        this.inputValue = '';
      } else {
        this.$message.warning(
          this.$t('knowledgeManage.create.chunkKeywordsPlaceholder'),
        );
      }
    },
    submit(formName) {
      if (this.createType === 'single' || this.isChildChunk) {
        this.handleSingle(formName);
      } else {
        this.handleFile();
      }
    },
    handleSingle(formName) {
      this.$refs[formName].validate(valid => {
        if (valid) {
          this.btnLoading = true;
          if (this.isChildChunk) {
            this.createChildChunk();
          } else {
            this.createParentChunk();
          }
        } else {
          return false;
        }
      });
    },
    createParentChunk() {
      const data = this.isChildChunk
        ? { content: this.ruleForm.content, docId: this.ruleForm.docId }
        : {
            content: this.ruleForm.content,
            docId: this.ruleForm.docId,
            labels: this.ruleForm.labels,
          };
      createSegment(data)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(
              this.$t('knowledgeManage.create.createSuccess'),
            );
            if (!this.checkType.length) {
              this.dialogVisible = false;
              this.$emit('updateDataBatch');
            } else {
              this.clearForm();
              this.$emit('updateData');
            }
            this.btnLoading = false;
          }
        })
        .catch(() => {
          this.btnLoading = false;
        });
    },
    createChildChunk() {
      const data = {
        content: [this.ruleForm.content],
        docId: this.ruleForm.docId,
        parentId: this.parentId,
      };
      createSegmentChild(data)
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
            this.$emit('updateChildData');
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
        docId: this.ruleForm.docId,
      };
      createBatchSegment(data)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(
              this.$t('knowledgeManage.create.createSuccess'),
            );
            this.dialogVisible = false;
            this.btnLoading = false;
            this.$emit('updateDataBatch');
          }
        })
        .catch(() => {
          this.btnLoading = false;
        });
    },
    clearForm() {
      this.ruleForm.content = '';
      if (!this.isChildChunk) {
        this.ruleForm.labels = [];
      }
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
