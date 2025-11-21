<template>
  <div class="urlBatch">
    <el-form ref="dynamicValidateForm" size="mini" class="urlAnalysisForm">
      <el-form-item>
        <el-upload
          class="upload-demo"
          ref="batUrl"
          drag
          multiple
          action=""
          :show-file-list="false"
          :auto-upload="false"
          :file-list="fileList"
          :limit="1"
          accept=".xlsx"
          :disabled="!categoryId"
          :on-change="handleChange"
          :on-remove="handleRemove"
        >
          <div @click="handleDisClick">
            <i
              class="el-icon-upload"
              v-if="fileList.length <= 0 || fileList[0].back === 'error'"
            ></i>
            <i
              class="el-icon-refresh"
              :class="{
                rotate:
                  fileList[0].loading === true && fileList[0].back === 'true',
              }"
              v-if="fileList.length > 0 && fileList[0].back === 'true'"
            ></i>
            <div class="el-upload__text" v-if="fileList.length <= 0">
              {{$t('knowledgeManage.dragFileTips')}}<em>&nbsp;{{$t('knowledgeManage.clickUpload')}}</em>
            </div>
          </div>

          <div class="el-upload__tip" slot="tip" style="color: red">
            {{$t('knowledgeManage.clickUploadTips')}}&nbsp;&nbsp;
            <a :href="templateUrl">{{$t('knowledgeManage.downTemplate')}}</a>
            <br />
            {{$t('knowledgeManage.notReshContent')}}
          </div>
        </el-upload>
        <div
          class="start"
          v-if="fileList.length > 0 && fileList[0].back === 'true'"
        >
          <el-button
            type="primary"
            :loading="loading.start"
            v-if="fileList[0].agin !== true"
            @click="handleClick('start')"
            >{{$t('knowledgeManage.startAnalysis')}}</el-button
          >
          <el-button
            type="primary"
            :loading="loading.agin"
            v-if="fileList[0].agin === true"
            @click="handleClick('agin')"
            >{{$t('knowledgeManage.refreshAnalysis')}}</el-button
          >
          <div>
            {{$t('knowledgeManage.total')}}<span>{{
              Number(batchRes.fail_count) + Number(batchRes.success_count) || 0
            }}</span
            >{{$t('knowledgeManage.Piece')}}，<span>{{ batchRes.success_count || 0 }}</span>
           {{$t('knowledgeManage.analysisFinish')}}，<span>{{ batchRes.fail_count || 0 }}</span>
            {{$t('knowledgeManage.analysisFail')}}&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
            <!-- v-if="fileList.length > 0 && fileList[0].agin === true" -->

            <a
              href="#"
              v-if="fileList.length > 0 && fileList[0].agin === true"
              @click.prevent="drawer = true"
              >{{$t('knowledgeManage.viewDetail')}}</a
            >
          </div>
          <div>{{$t('knowledgeManage.analysisFailTips')}}</div>
        </div>
        <div
          class="aginUp"
          v-if="fileList.length > 0 && fileList[0].back === 'error'"
        >
          <el-button type="primary" @click="submitVisible">{{$t('knowledgeManage.reUpload')}}</el-button>
        </div>
        <transition name="el-zoom-in-top">
          <ul class="document_lises">
            <li
              v-for="(item, index) in fileList"
              :key="index"
              :class="{
                document_error: item.back === 'error',
                document_loading: item.loading === true,
              }"
            >
              <span>{{ item.name }}</span>
              <div>
                <span class="size">{{ filterSize(item.size) }}</span>
                &nbsp;&nbsp;&nbsp;
                <span class="result_icon">
                  <i class="el-icon-error" @click="handleRemove(item)"></i>
                  <i class="el-icon-loading" v-if="item.loading === true"></i>
                  <i class="el-icon-success" v-if="item.back === 'true'"></i>
                  <i class="el-icon-warning" v-if="item.back === 'error'"></i>
                </span>
              </div>
            </li></ul
        ></transition>
      </el-form-item>
    </el-form>

    <el-drawer
      :title="$t('knowledgeManage.analysisDetail')"
      size="80%"
      :custom-class="'urlBatch_drawer'"
      :visible.sync="drawer"
      :append-to-body="true"
      :direction="'rtl'"
      :before-close="handleClose"
    >
      <div class="content">
        <div class="item">
          <h4>{{$t('knowledgeManage.analysisSuccess')}}</h4>
          <el-table
            border
            :data="batchRes.success"
            :max-height="maxHeight"
            style="width: 100%"
            :header-cell-style="{
              background: '#666',
              color: '#fff',
            }"
          >
            <el-table-column
              type="index"
              :label="$t('knowledgeManage.ID')"
              align="center"
              width="50"
            >
            </el-table-column>
            <el-table-column prop="url" label="URL" align="center">
            </el-table-column>
          </el-table>
        </div>
        <div class="item">
          <h4>{{$t('knowledgeManage.fail')}}</h4>
          <el-table
            border
            :data="batchRes.error"
            :max-height="maxHeight"
            style="width: 100%"
            :header-cell-style="{
              background: '#CC071B',
              color: '#fff',
            }"
          >
            <el-table-column
              type="index"
              :label="$t('knowledgeManage.ID')"
              width="50"
              align="center"
            >
            </el-table-column>
            <el-table-column
              prop="url"
              label="URL"
              show-overflow-tooltip
              align="center"
            >
            </el-table-column>
          </el-table>
        </div>
      </div>
      <div class="demo-drawer__footer">
        <el-button @click="handleClose">{{$t('knowledgeManage.backAnalysis')}}</el-button>
        <el-button type="primary" @click="handleSave" :loading="loading.btn"
          >{{$t('knowledgeManage.saveIntoData')}}</el-button
        >
      </div>
    </el-drawer>
  </div>
</template>
<script>
import { batchurl, batchUrlTaskStatus, importBatchUrl } from "@/api/knowledge";
import axios from "axios";

export default {
  props: {
    categoryId: {
      type: String,
      default: "",
    },
    validate: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      templateUrl: "",
      source: [],
      backId: "", // UploadDocumentBack of id
      batchRes: {}, // ParseResult
      maxHeight: "",
      drawer: false,
      fileList: [],
      urlConut: 1,
      backResult: [],
      loading: {
        url: false,
        start: false,
        agin: false,
        btn: false,
      },
      oldList: [{ value: "" }], //Save上一次urlResult
    };
  },
  created() {
    this.maxHeight = window.innerHeight - 210.5;
    const fileName = 'url_demo.xlsx';
    this.templateUrl = window.location.origin + '/img/downFile/' + fileName;
  },
  mounted() {
    this.$emit("handleSetDisabled", true);
  },
  methods: {
    handleDisClick() {
      if (!this.categoryId) {
        this.$emit("handleValidate");
      }
    },
    handleClose() {
      this.drawer = false;
    },
    handleClick(type) {
      if (type === "start") {
        this.loading.start = true;
        this.$set(this.fileList[0], "loading", true);
        this.startPolling();
      } else {
        this.loading.start = false;
        this.loading.agin = true;
        this.$set(this.fileList[0], "loading", true);
        this.aginPolling();
      }
    },
    startPolling() {
      batchUrlTaskStatus({
        taskId: this.backId,
      }).then((res) => {
          if (res.code === 0) {
            if (res.data.completed !== true) {
              setTimeout(() => {
                this.startPolling();
              }, 5000);
              return;
            } else {
              this.batchRes = res.data;
              this.batchRes.success = this.batchRes.url_list.filter((item) => {
                return item.status === 10;
              });
              this.batchRes.error = this.batchRes.url_list.filter((item) => {
                return item.status === 57;
              });
              console.log(this.batchRes);
              this.$emit("handleSetBatchDisabled", false);
            }
          } else {
            this.$emit("handleSetBatchDisabled", true);
          }
          this.loading.start = false;
          this.$set(this.fileList[0], "agin", true);
          this.$set(this.fileList[0], "loading", false);
        }).catch((err) => {
          this.loading.start = false;
          this.$set(this.fileList[0], "agin", true);
          this.$set(this.fileList[0], "loading", false);
          this.$emit("handleSetBatchDisabled", true);
        });
    },
    aginPolling() {
      batchUrlTaskStatus({
        taskId: this.backId,
      })
        .then((res) => {
          if (res.code === 0) {
            if (res.data.completed !== true) {
              setTimeout(() => {
                this.aginPolling();
              }, 5000);
              return;
            } else {
              this.batchRes = res.data;
              this.batchRes.success = this.batchRes.url_list.filter((item) => {
                return item.status === 10;
              });
              this.batchRes.error = this.batchRes.url_list.filter((item) => {
                return item.status === 57;
              });
            }
          } else {
            this.$emit("handleSetBatchDisabled", true);
          }
          this.$emit("handleSetBatchDisabled", false);
          this.loading.agin = false;
          this.$set(this.fileList[0], "loading", false);
        })
        .catch((err) => {
          this.loading.agin = false;
          this.$set(this.fileList[0], "loading", false);
          this.$emit("handleSetBatchDisabled", true);
        });
    },
    async handleChange(file, fileList) {
      await this.$emit("handleValidate");
      if (this.validate) {
        if (!this.verifyFormat(file) || !this.verifyEmpty(file)) return false;
        this.fileList = fileList;
        this.submitVisible();
      } else {
        this.fileList = [];
      }
    },
    handleRemove() {
      this.fileList = [];
      this.batchRes = {};
      this.$emit("handleSetBatchDisabled", true);
      this.$set(this.fileList[0], "loading", false);
      this.loading.start = false;
      this.loading.agin = false;
    },
    //  VerifyFileIs Empty
    verifyEmpty(file) {
      if (file.size <= 0) {
        this.$message.warning(file.name + this.$t('knowledgeManage.filterFile'));
        this.fileList = this.fileList.filter(
          (files) => files.name !== file.name
        );
        return false;
      }
      return true;
    },
    submitVisible() {
      //Upload
      // this.$refs["batUrl"].validate((valid) => {
      if (!this.fileList.length) {
        this.$message.warning(this.$t('knowledgeManage.selectFile'));
        return;
      }
      if (this.fileList[0].status === "ready") {
        // const type = this.fileList[0].name.split(".")[1];
        let formData = new FormData();
        formData.append("file", this.fileList[0].raw);
        // formData.append("fileType", type);
        formData.append("categoryId", this.categoryId);
        this.importDoc(formData, 0);
      }
      // }
      // });
    },
    importDoc(data, index) {
      // this.uploading = true;
      this.$set(this.fileList[index], "loading", true);
      this.$set(this.fileList[index], "back", null);
      const cancel = axios.CancelToken.source();
      this.source.push(cancel);
      batchurl(data, cancel.token)
        .then((res) => {
          if (res.code === 0) {
            this.backId = res.data;
            // this.$set(this.fileList[index], "id", res.data);
            this.$set(this.fileList[index], "back", "true");
            this.$set(this.fileList[index], "loading", false);
            // this.resultDisabled = false;
            this.$message.success(this.$t('knowledgeManage.uploadSuccess'));
            // this.uploading = false;
          } else {
            // this.uploading = false;
            this.$set(this.fileList[index], "back", "error");
            this.$set(this.fileList[index], "loading", false);
            // this.$message.error(res.message);
          }
        })
        .catch((err) => {
          this.$set(this.fileList[index], "back", "error");
          this.$set(this.fileList[index], "loading", false);
        });
    },
    //  VerifyFileFormatsize
    verifyFormat(file) {
      var testmsg = file.name.substring(file.name.lastIndexOf(".") + 1);

      const extension = testmsg === "xlsx";

      if (!extension) {
        this.$message.warning(file.name + this.$t('knowledgeManage.fileTypeError'));
        this.fileList = this.fileList.filter(
          (files) => files.name !== file.name
        );
        return false;
      } else {
        const isLt15 = file.size / 1024 / 1024 < 15;
        if (!isLt15) {
          this.$message.error(this.$t('knowledgeManage.fileSizeTips'));
          this.fileList = this.fileList.filter(
            (files) => files.name !== file.name
          );
          return false;
        }
      }
      return true;
    },
    filterSize(size) {
      if (!size) return "";
      var num = 1024.0; //byte
      if (size < num) return size + "B";
      if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + "KB"; //kb
      if (size < Math.pow(num, 3))
        return (size / Math.pow(num, 2)).toFixed(2) + "MB"; //M
      if (size < Math.pow(num, 4))
        return (size / Math.pow(num, 3)).toFixed(2) + "G"; //G
      return (size / Math.pow(num, 4)).toFixed(2) + "T"; //T
    },
    handleSave(type) {
      this.loading.btn = true;
      this.$emit("handleSetBatchStatus", { type: "loading", value: true });
      importBatchUrl({ taskId: this.backId })
        .then((res) => {
          this.drawer = false;
          this.loading.btn = false;
          this.$emit("handleSetBatchStatus", {
            type: "result",
            value: "success",
          });
        })
        .catch((err) => {
          this.loading.btn = false;
          this.$emit("handleSetBatchStatus", {
            type: "result",
            value: "error",
          });
        });
    },
    handlebtnSave() {
      this.loading.btn = true;
      this.$emit("handleSetBatchStatus", { type: "loading", value: true });
      importBatchUrl({ taskId: this.backId })
        .then((res) => {
          this.drawer = false;
          this.loading.btn = false;
          setTimeout(() => {
            this.$emit("handleSetBatchStatus", {
              type: "result",
              value: "success",
            });
          }, 500);
        })
        .catch((err) => {
          this.loading.btn = false;
          setTimeout(() => {
            this.$emit("handleSetBatchStatus", {
              type: "result",
              value: "success",
            });
          }, 500);
        });
    },
    // async download(url, name) {
    //   // const res = await BatchUrlDemo();
    //   // console.log(res);
    //   const blob = new Blob(["https://122.13.25.19:17776/cubm/demo.xlsx"], {
    //     type: "application/octet-stream",
    //   });
    //   const blobUrl = window.URL.createObjectURL(blob); // 将blobObject转is一URL
    //   const link = document.createElement("a");
    //   link.href = blobUrl;
    //   link.download = "template.xlsx";
    //   link.click(); // StartDownload
    //   window.URL.revokeObjectURL(link.href); // Download完毕DeleteaTag
    // },
  },
  computed: {},
};
</script>
<style lang="scss">
.urlBatch_drawer {
  .demo-drawer__footer {
    position: absolute;
    bottom: 0;
    left: 10px;
    right: 10px;
    height: 50px;
    text-align: center;
  }
  .content {
    display: flex;
    height: calc(100% - 50px);
    padding: 20px;

    .item {
      width: 50%;
      height: 100%;
      padding: 5px;
      overflow-y: auto;
      border-right: 1px dashed #bbb;
      border-bottom: 1px dashed #bbb;

      &:last-child {
        border-right: 0;
      }
      h4 {
        padding-bottom: 10px;
        font-size: 16px;
        text-align: center;
      }
    }
  }
}
.urlBatch {
  .el-upload {
    width: 100%;
    em {
      color: #f72324;
      font-weight: bold;
    }
  }
  .el-upload-dragger {
    background-color: #fff !important;
    border: 1px dashed #d9d9d9 !important;
    border-radius: 6px !important;
    box-sizing: border-box !important;
    width: 100% !important;
    height: 160px !important;
    margin: 0 auto;
  }
  .aginUp {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 160px;
    text-align: center;
    z-index: 1;
    padding-top: 110px;
  }
  .start {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 160px;
    text-align: center;
    z-index: 1;
    padding-top: 70px;

    .el-button {
      margin-bottom: 5px;
      &.el-button--mini {
        height: auto;
      }
      span {
        font-size: 12px;
      }
    }
    div {
      margin: 0;
      padding: 0;
      line-height: 1.5;
      font-size: 12px;
      color: #333;

      &:last-child {
        color: #999;
      }

      span {
        font-size: 18px;
        font-weight: bold;
      }
      a {
        color: #f72324;
        font-size: 12px;
      }
    }
  }
  .el-upload-dragger .el-icon-refresh {
    font-size: 50px;
    color: #f72324 !important;
    margin: 10px 0 16px;
    line-height: 50px;

    &.rotate {
      animation: infinite-rotation 1.3s linear infinite;
    }
  }
  @keyframes infinite-rotation {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(-360deg);
    }
  }
  .el-upload-dragger .el-upload__text {
    color: #afafaf !important;
  }
  .document_lises {
    list-style: none;

    li {
      display: flex;
      justify-content: space-between;
      font-size: 12px;
      padding: 7px;
      border-radius: 3px;
      line-height: 1;

      .el-icon-success {
        display: block;
      }
      .el-icon-error {
        display: none;
      }
      &:hover {
        cursor: pointer;
        background: #eee;

        .el-icon-success {
          display: none;
        }
        .el-icon-warning {
          display: none;
        }
        .el-icon-error {
          display: block;
        }
      }
      &.document_loading {
        .el-icon-error {
          display: none;
        }
        .el-icon-success {
          display: none;
        }
        .el-icon-warning {
          display: none;
        }
        &:hover {
          cursor: pointer;
          background: #eee;
        }
      }
      .el-icon-success {
        color: #67c23a;
      }

      .result_icon {
        float: right;
      }
      .size {
        font-weight: bold;
      }
    }
    .document_error {
      color: #e60001;
    }
  }
}
</style>
