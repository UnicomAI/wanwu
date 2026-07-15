<template>
  <div>
    <el-dialog
      top="10vh"
      :title="$t('safety.createWord.title')"
      :close-on-click-modal="false"
      :visible.sync="dialogVisible"
      width="50%"
      :before-close="handleClose"
    >
      <el-form
        :model="ruleForm"
        ref="ruleForm"
        label-width="140px"
        class="demo-ruleForm"
        :rules="rules"
        @submit.native.prevent
      >
        <el-form-item class="itemCenter">
          <el-radio-group v-model="ruleForm.importType">
            <el-radio-button :label="'single'">
              {{ $t('safety.createWord.single') }}
            </el-radio-button>
            <el-radio-button :label="'file'">
              {{ $t('safety.createWord.file') }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('safety.createWord.word')"
          prop="word"
          v-if="ruleForm.importType === 'single'"
        >
          <el-input
            v-model="ruleForm.word"
            :placeholder="$t('safety.createWord.wordPlaceholder')"
          />
        </el-form-item>
        <el-form-item
          :label="$t('safety.createWord.type')"
          prop="sensitiveType"
          v-if="ruleForm.importType === 'single'"
        >
          <el-select
            v-model="ruleForm.sensitiveType"
            :placeholder="$t('common.select.placeholder')"
            style="width: 100%"
          >
            <el-option
              v-for="item in sensitiveTypeOptions"
              :key="item.value"
              :label="item.name"
              :value="item.value"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('safety.createWord.file')"
          prop="fileName"
          v-if="ruleForm.importType === 'file'"
        >
          <el-upload
            class="upload-box"
            drag
            action=""
            :show-file-list="false"
            :auto-upload="false"
            accept=".xlsx"
            :file-list="fileList"
            :on-change="uploadOnChange"
          >
            <div>
              <div>
                <img
                  :src="require('@/assets/imgs/uploadImg.png')"
                  class="upload-img"
                />
                <p class="click-text">
                  {{ $t('common.fileUpload.uploadText') }}
                  <span class="clickUpload">
                    {{ $t('common.fileUpload.uploadClick') }}
                  </span>
                  <a
                    class="clickUpload template"
                    @click.stop="handleDownloadTemplate"
                  >
                    {{ $t('common.fileUpload.templateClick') }}
                  </a>
                </p>
              </div>
            </div>
          </el-upload>
          <!-- 上传文件的列表 -->
          <div class="file-list" v-if="fileList.length > 0">
            <transition name="el-zoom-in-top">
              <ul class="document_lise">
                <li
                  v-for="(file, index) in fileList"
                  :key="index"
                  class="document_lise_item"
                >
                  <div style="padding: 8px 0" class="lise_item_box">
                    <span class="size">
                      <img :src="require('@/assets/imgs/fileicon.png')" />
                      {{ file.name }}
                      <span class="file-size">
                        {{ filterSize(file.size) }}
                      </span>
                      <el-progress
                        :percentage="file.percentage"
                        v-if="file.percentage !== 100"
                        :status="file.progressStatus"
                        max="100"
                        class="progress"
                      ></el-progress>
                    </span>
                    <span class="handleBtn">
                      <span>
                        <span v-if="file.percentage === 100">
                          <i
                            class="el-icon-check check success"
                            v-if="file.progressStatus === 'success'"
                          ></i>
                          <i class="el-icon-close close fail" v-else></i>
                        </span>
                        <i
                          class="el-icon-loading"
                          v-else-if="
                            file.percentage !== 100 && index === fileIndex
                          "
                        ></i>
                      </span>
                      <span style="margin-left: 30px">
                        <i
                          class="el-icon-error error"
                          @click="handleRemove(file, index)"
                        ></i>
                      </span>
                    </span>
                  </div>
                </li>
              </ul>
            </transition>
          </div>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="handleClose()">
          {{ $t('common.confirm.cancel') }}
        </el-button>
        <el-button type="primary" @click="submitForm('ruleForm')">
          {{ $t('common.confirm.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import uploadChunk from '@/mixins/uploadChunk';
import { delfile } from '@/api/chunkFile';
import { uploadSensitiveWord } from '@/api/safety';
import { filterSize, authDownload } from '@/utils/util';

export default {
  mixins: [uploadChunk],
  data() {
    return {
      basePath: this.$basePath,
      sensitiveTypeOptions: [
        {
          value: 'Political',
          name: this.$t('common.safetyType.political'),
        },
        {
          value: 'Revile',
          name: this.$t('common.safetyType.revile'),
        },
        {
          value: 'Pornography',
          name: this.$t('common.safetyType.pornography'),
        },
        {
          value: 'ViolentTerror',
          name: this.$t('common.safetyType.violentTerror'),
        },
        {
          value: 'Illegal',
          name: this.$t('common.safetyType.illegal'),
        },
        {
          value: 'InformationSecurity',
          name: this.$t('common.safetyType.informationSecurity'),
        },
        {
          value: 'Other',
          name: this.$t('common.safetyType.other'),
        },
      ],
      dialogVisible: false,
      ruleForm: {
        importType: 'single',
        word: '',
        sensitiveType: '',
        fileName: '',
        tableId: '',
      },
      fileList: [],
      rules: {
        word: [
          {
            required: true,
            message: this.$t('safety.createWord.wordMsg'),
            trigger: 'blur',
          },
        ],
        sensitiveType: [
          {
            required: true,
            message: this.$t('safety.createWord.typeMsg'),
            trigger: 'blur',
          },
        ],
        fileName: [
          {
            required: true,
            message: this.$t('safety.createWord.fileMsg'),
            trigger: 'blur',
          },
        ],
      },
    };
  },
  methods: {
    async handleDownloadTemplate() {
      try {
        await authDownload(
          this.basePath + '/user/api/v1/files/docs/sensitive.xlsx',
          'sensitive.xlsx',
        );
      } catch (error) {
        this.$message.error(this.$t('common.fileUpload.downloadError'));
      }
    },
    uploadOnChange(file, fileList) {
      if (!fileList.length) return;
      const newFileList = fileList[fileList.length - 1];
      this.fileList = [newFileList];
      if (this.fileList.length > 0) {
        this.maxSizeBytes = 0;
        this.isExpire = true;
        this.startUpload();
      }
    },
    filterSize,
    handleRemove(item, index) {
      const data = { fileList: [this.resList[index]['name']], isExpired: true };
      delfile(data).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.delete'));
        }
      });
      this.fileList = this.fileList.filter(files => files.name !== item.name);
      if (this.fileList.length === 0) {
        this.file = null;
      } else {
        this.fileIndex--;
      }
    },
    uploadFile(chunkFileName) {
      this.ruleForm.fileName = chunkFileName;
    },
    handleClose() {
      this.dialogVisible = false;
      this.ruleForm.tableId = '';
      this.clearform();
    },
    clearform() {
      this.fileList = [];
      this.$refs.ruleForm.resetFields();
      this.$refs.ruleForm.clearValidate();
    },
    submitForm(formName) {
      this.$refs[formName].validate(valid => {
        if (valid) {
          uploadSensitiveWord(this.ruleForm)
            .then(res => {
              if (res.code == 0) {
                this.$message.success(this.$t('common.message.success'));
                this.$emit('reload');
                this.dialogVisible = false;
              }
            })
            .catch(err => {});
        } else {
          return false;
        }
      });
    },
    showDialog(tableId) {
      this.dialogVisible = true;
      this.ruleForm.tableId = tableId;
      this.clearform();
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

.upload-box {
  .upload-img {
    width: 56px;
    height: 56px;
    margin-top: 10px;
  }

  .clickUpload,
  .template {
    color: $color;
    font-weight: bold;
  }

  .template {
    margin-left: 10px;
  }
}

.file-list {
  padding: 20px 0;

  .document_lise_item {
    cursor: pointer;
    padding: 0 10px;
    list-style: none;
    border-radius: 4px;
    border: 1px solid #7684fd;
    display: flex;
    align-items: center;
    margin-bottom: 10px;

    .lise_item_box {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .size {
        display: flex;
        flex: 1;
        align-items: center;

        .progress {
          width: 200px;
          margin-left: 30px;
        }

        img {
          width: 18px;
          height: 18px;
          margin-bottom: -3px;
        }

        .file-size {
          margin-left: 10px;
        }
      }
    }
  }

  .document_lise_item:hover {
    background: #eceefe;
  }
}
</style>
