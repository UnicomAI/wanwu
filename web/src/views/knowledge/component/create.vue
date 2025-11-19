<template>
  <div>
    <el-dialog
      top="10vh"
      :title="
        this.isEdit
          ? $t('knowledgeManage.editInfo')
          : $t('knowledgeManage.createKnowledge')
      "
      :close-on-click-modal="false"
      :visible.sync="dialogVisible"
      width="70%"
      :before-close="handleClose"
      class="knowledge-create-dialog"
    >
      <el-form
        :model="ruleForm"
        ref="ruleForm"
        label-width="140px"
        class="demo-ruleForm"
        :rules="rules"
        @submit.native.prevent
      >
        <el-form-item
          :label="$t('knowledgeManage.knowledgeName') + '：'"
          prop="name"
        >
          <el-input
            v-model="ruleForm.name"
            :placeholder="$t('knowledgeManage.categoryNameRules')"
            maxlength="50"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('knowledgeManage.desc') + ':'"
          prop="description"
        >
          <el-input
            v-model="ruleForm.description"
            :placeholder="$t('common.input.inputDesc')"
          ></el-input>
        </el-form-item>
        <el-form-item label="Embedding" prop="embeddingModelInfo.modelId">
          <el-select
            v-model="ruleForm.embeddingModelInfo.modelId"
            :placeholder="$t('common.select.placeholder')"
            value-key="modelId"
            :disabled="isEdit"
          >
            <el-option
              v-for="item in EmbeddingOptions"
              :key="item.modelId"
              :label="item.displayName"
              :value="item.modelId"
            >
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item prop="knowledgeGraph.switch">
          <template #label>
            <span>{{ $t("knowledgeManage.create.knowledgeGraph") }}:</span>
            <el-tooltip
              class="item"
              effect="dark"
              placement="top-start"
              popper-class="knowledge-graph-tooltip"
            >
              <span class="el-icon-question question-icon"></span>
              <template #content>
                <p
                  v-for="(item, i) in knowledgeGraphTips"
                  :key="i"
                  class="tooltip-item"
                >
                  <span class="tooltip-title">{{ item.title }}</span>
                  <span class="tooltip-content">{{ item.content }}</span>
                </p>
              </template>
            </el-tooltip>
          </template>
          <el-switch v-model="ruleForm.knowledgeGraph.switch"></el-switch>
        </el-form-item>
        <el-form-item
          :label="$t('knowledgeManage.create.modelSelect') + ':'"
          prop="knowledgeGraph.llmModelId"
          v-if="ruleForm.knowledgeGraph.switch"
        >
          <el-select
            v-model="ruleForm.knowledgeGraph.llmModelId"
            :placeholder="$t('knowledgeManage.create.modelSearchPlaceholder')"
            @visible-change="visibleChange"
            :loading-text="$t('knowledgeManage.create.modelLoading')"
            class="cover-input-icon model-select"
            :disabled="isPublish"
            :loading="modelLoading"
            filterable
            value-key="modelId"
          >
            <el-option
              class="model-option-item"
              v-for="item in knowledgeGraphModelOptions"
              :key="item.modelId"
              :value="item.modelId"
              :label="item.displayName"
            >
              <div class="model-option-content">
                <span class="model-name">{{ item.displayName }}</span>
                <div
                  class="model-select-tags"
                  v-if="item.tags && item.tags.length > 0"
                >
                  <span
                    v-for="(tag, tagIdx) in item.tags"
                    :key="tagIdx"
                    class="model-select-tag"
                    >{{ tag.text }}</span
                  >
                </div>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('knowledgeManage.create.uploadSchema') + ':'"
          v-if="ruleForm.knowledgeGraph.switch"
        >
          <el-upload
            action=""
            :auto-upload="false"
            :show-file-list="false"
            :on-change="uploadOnChange"
            :file-list="fileList"
            :limit="1"
            drag
            accept=".xlsx,.xls"
            class="upload-box"
          >
            <div>
              <div>
                <img
                  :src="require('@/assets/imgs/uploadImg.png')"
                  class="upload-img"
                />
                <p class="click-text">
                  {{ $t("knowledgeManage.create.dragUpload")
                  }}<span class="clickUpload">{{
                    $t("knowledgeManage.create.clickUpload")
                  }}</span>
                </p>
              </div>
              <div class="tips">
                <p>
                  <span class="red">*</span
                  >{{ $t("knowledgeManage.create.schemaTip1") }}
                  <a
                    class="template_downLoad"
                    href="#"
                    @click.prevent.stop="downloadTemplate"
                    >{{ $t("knowledgeManage.create.templateDownload") }}</a
                  >
                </p>
                <p>
                  <span class="red">*</span
                  >{{ $t("knowledgeManage.create.schemaTip2") }}
                </p>
              </div>
            </div>
          </el-upload>
          <!-- UploadFile of List -->
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
        <el-button @click="handleClose()">{{
          $t("common.confirm.cancel")
        }}</el-button>
        <el-button type="primary" @click="submitForm('ruleForm')">{{
          $t("common.confirm.confirm")
        }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import { mapActions, mapGetters } from "vuex";
import { createKnowledgeItem, editKnowledgeItem } from "@/api/knowledge";
import { selectModelList } from "@/api/modelAccess";
import { KNOWLEDGE_GRAPH_TIPS } from "../config";
import uploadChunk from "@/mixins/uploadChunk";
import { delfile } from "@/api/chunkFile";
export default {
  mixins: [uploadChunk],
  data() {
    let checkName = (rule, value, callback) => {
      const reg = /^[\u4E00-\u9FA5a-z0-9_-]+$/;
      if (!reg.test(value)) {
        callback(new Error(this.$t("knowledgeManage.inputErrorTips")));
      } else {
        return callback();
      }
    };
    return {
      dialogVisible: false,
      ruleForm: {
        name: "",
        description: "",
        embeddingModelInfo: {
          modelId: "",
        },
        knowledgeGraph: {
          llmModelId: "",
          schemaUrl: "",
          switch: false,
        },
      },
      EmbeddingOptions: [],
      knowledgeGraphModelOptions: [],
      modelLoading: false,
      knowledgeGraphTips: KNOWLEDGE_GRAPH_TIPS,
      maxSizeBytes: 0, // Setting为0，所HasFile都走切片Upload
      rules: {
        name: [
          {
            required: true,
            message: this.$t("knowledgeManage.knowledgeNameRules"),
            trigger: "blur",
          },
          { validator: checkName, trigger: "blur" },
        ],
        description: [
          {
            required: true,
            message: this.$t("knowledgeManage.inputDesc"),
            trigger: "blur",
          },
        ],
        "embeddingModelInfo.modelId": [
          {
            required: true,
            message: this.$t("common.select.placeholder"),
            trigger: "blur",
          },
        ],
        "knowledgeGraph.llmModelId": [
          {
            required: true,
            message: this.$t("knowledgeManage.create.selectModel"),
            trigger: "change",
          },
        ],
      },
      isEdit: false,
      knowledgeId: "",
      isPublish: false,
    };
  },
  watch: {
    embeddingList: {
      handler(val) {
        if (val) {
          this.EmbeddingOptions = val;
        }
      },
    },
  },
  computed: {
    ...mapGetters("app", ["embeddingList"]),
  },
  created() {
    this.getEmbeddingList();
    this.getModelData(); //GetModelList
  },
  methods: {
    ...mapActions("app", ["getEmbeddingList"]),
    visibleChange(val) {
      //下拉框Show of  when 候RequestModelList
      if (val) {
        this.getModelData();
      }
    },
    async downloadTemplate() {
      const url = "/user/api/v1/static/docs/graph_schema.xlsx";
      const fileName = "graph_schema.xlsx";
      try {
        const response = await fetch(url);
        if (!response.ok)
          throw new Error(this.$t("knowledgeManage.create.fileNotExist"));

        const blob = await response.blob();
        const blobUrl = URL.createObjectURL(blob);

        const a = document.createElement("a");
        a.href = blobUrl;
        a.download = fileName;
        a.click();

        URL.revokeObjectURL(blobUrl); // 释放Memory
      } catch (error) {
        this.$message.error(this.$t("knowledgeManage.create.downloadFailed"));
      }
    },
    async getModelData() {
      this.modelLoading = true;
      const res = await selectModelList();
      if (res.code === 0) {
        this.knowledgeGraphModelOptions = (res.data.list || []).filter(
          (item) => !item.config || item.config.visionSupport !== "support"
        );
        this.modelLoading = false;
      }
      this.modelLoading = false;
    },
    handleClose() {
      this.dialogVisible = false;
      this.clearform();
    },
    clearform() {
      (this.isEdit = false), (this.knowledgeId = "");
      this.$refs.ruleForm.resetFields();
      this.$refs.ruleForm.clearValidate();
      this.fileList = [];
      this.cancelAllRequests();
      this.file = null;
      this.fileIndex = 0;
      this.fileUuid = "";
    },
    uploadOnChange(file, fileList) {
      if (!fileList.length) return;
      this.fileList = fileList;
      if (
        this.verifyEmpty(file) !== false &&
        this.verifyFormat(file) !== false &&
        this.verifyRepeat(file) !== false
      ) {
        setTimeout(() => {
          this.fileList.map((file, index) => {
            if (file.progressStatus && file.progressStatus !== "success") {
              this.$set(file, "progressStatus", "exception");
              this.$set(file, "showRetry", "false");
              this.$set(file, "showResume", "false");
              this.$set(file, "showRemerge", "false");
              if (file.size > this.maxSizeBytes) {
                this.$set(file, "fileType", "maxFile");
              } else {
                this.$set(file, "fileType", "minFile");
              }
            }
          });
        }, 10);
        //开始切片Upload(IfNoFile正在Upload)
        if (this.file === null) {
          this.startUpload();
        } else {
          //IfUpload当中Has新 of File加入
          if (this.file.progressStatus === "success") {
            this.startUpload(this.fileIndex);
          }
        }
      }
    },
    //  VerifyFileIs Empty
    verifyEmpty(file) {
      if (file.size <= 0) {
        setTimeout(() => {
          this.$message.warning(
            file.name + this.$t("knowledgeManage.filterFile")
          );
          this.fileList = this.fileList.filter(
            (files) => files.name !== file.name
          );
        }, 50);
        return false;
      }
      return true;
    },
    //  VerifyFileFormat
    verifyFormat(file) {
      const nameType = ["xlsx", "xls"];
      const fileName = file.name;
      const isSupportedFormat = nameType.some((ext) =>
        fileName.endsWith(`.${ext}`)
      );
      if (!isSupportedFormat) {
        setTimeout(() => {
          this.$message.warning(
            file.name + this.$t("knowledgeManage.fileTypeError")
          );
          this.fileList = this.fileList.filter(
            (files) => files.name !== file.name
          );
        }, 50);
        return false;
      } else {
        const fileType = file.name.split(".").pop();
        const limit20 = ["xlsx", "xls"];
        let isLimit20 = file.size / 1024 / 1024 < 20;
        let num = 0;
        if (limit20.includes(fileType)) {
          num = 20;
          if (!isLimit20) {
            setTimeout(() => {
              this.$message.error(
                this.$t("knowledgeManage.limitSize") + `${num}MB!`
              );
              this.fileList = this.fileList.filter(
                (files) => files.name !== file.name
              );
            }, 50);
            return false;
          }
          return true;
        }
        return true;
      }
    },
    //  VerifyFile重复
    verifyRepeat(file) {
      let res = true;
      setTimeout(() => {
        this.fileList = this.fileList.reduce((accumulator, current) => {
          const length = accumulator.filter(
            (obj) => obj.name === current.name
          ).length;
          if (length === 0) {
            accumulator.push(current);
          } else {
            this.$message.warning(
              current.name + this.$t("knowledgeManage.fileExist")
            );
            res = false;
          }
          return accumulator;
        }, []);
        return res;
      }, 50);
    },
    filterSize(size) {
      if (!size) return "";
      let num = 1024.0; //byte
      if (size < num) return size + "B";
      if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + "KB"; //kb
      if (size < Math.pow(num, 3))
        return (size / Math.pow(num, 2)).toFixed(2) + "MB"; //M
      if (size < Math.pow(num, 4))
        return (size / Math.pow(num, 3)).toFixed(2) + "G"; //G
      return (size / Math.pow(num, 4)).toFixed(2) + "T"; //T
    },
    handleRemove(item, index) {
      if (item.percentage < 100) {
        this.fileList.splice(index, 1);
        this.cancelAllRequests();
        return;
      }
      // IfFile已UploadSuccess，NeedDeleteService器上 of File
      if (this.resList && this.resList[index] && this.resList[index]["name"]) {
        this.delfile({
          fileList: [this.resList[index]["name"]],
          isExpired: true,
        });
        this.resList.splice(index, 1);
      }
      this.fileList = this.fileList.filter((files) => files.name !== item.name);
      if (this.fileList.length === 0) {
        this.file = null;
        this.ruleForm.knowledgeGraph.schemaUrl = "";
      } else {
        this.fileIndex--;
      }
    },
    delfile(data) {
      delfile(data).then((res) => {
        if (res.code === 0) {
          this.$message.success(
            this.$t("knowledgeManage.create.deleteSuccess")
          );
        }
      });
    },
    uploadFile(fileName, oldName, filePath) {
      this.ruleForm.knowledgeGraph.schemaUrl = filePath || fileName;
      this.fileIndex++;
      if (this.fileIndex < this.fileList.length) {
        this.startUpload(this.fileIndex);
      }
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (valid) {
          if (this.isEdit) {
            this.editKnowledge();
          } else {
            this.createKnowledge();
          }
          this.$parent.clearIptValue();
        } else {
          return false;
        }
      });
    },
    createKnowledge() {
      createKnowledgeItem(this.ruleForm)
        .then((res) => {
          if (res.code === 0) {
            this.$message.success(
              this.$t("knowledgeManage.create.createSuccess")
            );
            this.$emit("reloadData");
            this.dialogVisible = false;
          }
        })
        .catch((error) => {
          this.$message.error(error);
        });
    },
    editKnowledge() {
      const data = {
        ...this.ruleForm,
        knowledgeId: this.knowledgeId,
      };
      editKnowledgeItem(data)
        .then((res) => {
          if (res.code === 0) {
            this.$message.success(
              this.$t("knowledgeManage.create.editSuccess")
            );
            this.$emit("reloadData");
            this.clearform();
            this.dialogVisible = false;
          }
        })
        .catch((error) => {
          this.$message.error(error);
        });
    },
    showDialog(row) {
      this.dialogVisible = true;
      this.isEdit = Boolean(row);
      if (row) {
        this.knowledgeId = row.knowledgeId;
        this.ruleForm = {
          name: row.name,
          description: row.description,
          embeddingModelInfo: {
            modelId: row.embeddingModelInfo.modelId,
          },
          knowledgeGraph: row.knowledgeGraph,
        };
      } else {
        this.ruleForm = {
          name: "",
          description: "",
          embeddingModelInfo: {
            modelId: "",
          },
          knowledgeGraph: {
            llmModelId: "",
            schemaUrl: "",
            switch: false,
          },
        };
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.knowledge-create-dialog {
  /deep/ .el-dialog__body {
    max-height: 60vh;
    overflow-y: auto;
    padding: 20px;
  }
  /deep/ .el-form-item {
    .el-select {
      width: 100%;
    }
  }
}

.question-icon {
  cursor: pointer;
  color: #909399;
}

.upload-box {
  height: auto;
  min-height: 190px;
  width: 100% !important;

  .upload-img {
    width: 56px;
    height: 56px;
    margin-top: 20px;
  }

  .click-text {
    .clickUpload {
      color: $color;
      font-weight: bold;
    }
  }

  .tips {
    padding: 0 20px;
    p {
      line-height: 1.6;
      color: #666666 !important;
      .red {
        color: #f56c6c;
      }
      .template_downLoad {
        margin-left: 5px;
        color: $color;
        cursor: pointer;
      }
    }
  }
}

.file-list {
  padding: 20px 0;
  .document_lise {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .document_lise_item {
    cursor: pointer;
    padding: 5px 10px;
    list-style: none;
    background: #fff;
    border-radius: 4px;
    box-shadow: 1px 2px 2px #ddd;
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
        align-items: center;
        .progress {
          width: 400px;
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
      .handleBtn {
        display: flex;
        align-items: center;
        .check.success {
          color: #67c23a;
        }
        .close.fail {
          color: #f56c6c;
        }
        .error {
          color: #f56c6c;
          cursor: pointer;
          font-size: 18px;
        }
      }
    }
  }
  .document_lise_item:hover {
    background: #eceefe;
  }
}
</style>

<style lang="scss">
.knowledge-graph-tooltip {
  max-width: 400px !important;

  .tooltip-item {
    margin: 0;
    padding: 4px 0;

    .tooltip-title {
      font-weight: bold;
      margin-right: 8px;
    }

    .tooltip-content {
      display: inline-block;
    }
  }
}
</style>