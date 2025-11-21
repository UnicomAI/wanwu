<template>
  <div class="">
    <el-upload
      v-if="fileList.length === 0"
      :class="['upload-box']"
      drag
      action=""
      :show-file-list="false"
      :auto-upload="false"
      :limit="1"
      :accept="fileType"
      :file-list="fileList"
      :on-change="uploadOnChange"
    >
      <div>
        <i class="el-icon-upload"></i>
        <p>Drag file here, orclick to upload</p>
        <div class="tips">
          <p>
            Supported
            {{ fileType }}
            format
          </p>
        </div>
      </div>
    </el-upload>
    <div class="fileList_box" v-if="fileList.length > 0">
      <p :class="{ fail: fail, resSuccess: !loading && !fail }">
        <svg-icon :icon-class="fileTypeName" />
        <el-tooltip
          class="item"
          effect="dark"
          :content="fileList[0].name"
          placement="top-start"
        >
          <span>{{ fileList[0].name }}</span>
        </el-tooltip>
        <el-tooltip
          class="item"
          effect="dark"
          content="Delete"
          placement="top"
          v-show="!loading"
        >
          <i class="el-icon-delete" @click="handleDelete"></i>
        </el-tooltip>

        <i class="el-icon-loading" v-if="loading"></i>
        <i class="el-icon-success" v-if="!loading && !fail"></i>
        <el-tooltip
          v-if="fail"
          class="item"
          effect="dark"
          content="Re-upload"
          placement="top"
          v-show="!loading"
        >
          <i class="el-icon-refresh" @click="handleAgin"></i>
        </el-tooltip>
      </p>
    </div>
  </div>
</template>

<script>
import {externalUpload} from "@/api/workflow";
import axios from "axios";
export default {
  props: ["index"],
  data() {
    return {
      fileList: [],
      fileType: ".jpg,.jpeg,.png,.webp,.txt,.pdf,.docx,.xlsx,.csv,.pptx",
      //UploadFile弹框
      loading: false,
      fail: false,
    };
  },
  watch: {
    loading(val) {
      this.$emit("handleDisabled", val);
      if (val) {
        this.$emit("handleCancel", {
          index: this.index,
          source: this.source,
        });
      }
    },
  },
  created() {},
  methods: {
    handleDelete() {
      this.loading = false;
      this.source && this.source.cancel();
      this.$emit("handleUploadSuccess", { index: this.index });

      setTimeout(() => {
        this.fail = false;
        this.fileList = [];
      }, 10);
    },
    handleAgin() {
      this.loading = true;
      this.fail = false;
      this.preUpload();
    },
    uploadOnChange(file, fileList) {
      if (this.handleVerifySize(file) && this.handleVerifyFormat(file)) {
        this.loading = true;
        this.fileList = fileList;
        this.preUpload();
      } else {
      }
    },
    // CheckFileYesNoIs Empty
    handleVerifySize(file) {
      if (file.size <= 0) {
        this.fileList = [];
        this.$message.error("FileCannot be empty");
        return false;
      }
      return true;
    },
    // CheckFileFormat
    handleVerifyFormat(file) {
      let fileSplitArr = file.name.split(".");
      let fileType = fileSplitArr[fileSplitArr.length - 1];
      let res = this.fileType.split(",").includes("." + fileType);
      if (!res) {
        this.fileList = [];
        this.$message.error("Fileformat不正确");
        return false;
      }
      return res;
    },
    preUpload() {
      let file = this.fileList[0] || {}
      let filenames = file.name.split(".")
      if (filenames.length > 1) {
        filenames.splice(filenames.length - 1, 1)
      }
      const fname = filenames.join(".")
      this.source = axios.CancelToken.source();

      const config = { headers: { "Content-Type": "multipart/form-data" } }
      let formData = new FormData()
      formData.append("file", file.raw)
      formData.append("file_name", fname)

      externalUpload(formData, config).then((res)=>{
        const { download_link: link } = res.data || {}
        this.loading = false;
        const params = {
          index: this.index,
          url: link,
          fileId: link,
          file: file,
        };
        this.$emit("handleUploadSuccess", params);
        this.$message.success("FileUploadSuccess");
      }).catch(()=>{
        this.loading = false;
        this.fail = true;
        this.$message.error("File" + this.fileList[0].name + "UploadFailed");
      })
    },
  },
  computed: {
    fileTypeName() {
      if (this.fileList.length > 0) {
        let fileSplitArr = this.fileList[0].name.split(".");
        let fileType = fileSplitArr[fileSplitArr.length - 1];
        return fileType;
      }
      return "";
    },
  },
};
</script>

<style lang="scss" scoped>
/deep/.el-upload {
  width: 100%;
  .el-upload-dragger {
    width: 100%;
    border-color: $color;
    p {
      padding: 0 20px;
      color: $color;
      font-size: 12px;
    }
    .el-icon-upload {
      color: $color;
    }
  }
}
.svg-icon {
  width: 20px;
  height: 20px;
}
.fileList_box {
  p {
    display: flex;
    align-items: center;
    padding: 10px;
    background: rgba(220, 221, 230, 0.6);

    border-radius: 5px;
    cursor: pointer;
    &:hover {
      background: rgba(220, 221, 230, 1);

      .el-icon-delete {
        display: block !important;
      }
      .el-icon-loading {
        display: none !important;
      }
      .el-icon-success {
        display: none !important;
      }
    }
    span {
      margin: 0 6% 0 3%;
      max-width: 80%;
      font-size: 14px;
      white-space: nowrap; /* 禁止换row */
      overflow: hidden; /* hide溢出Content */
      text-overflow: ellipsis; /* Exceed用省略号Table示 */
    }
    .el-icon-delete,
    .el-icon-refresh {
      display: none;

      &:hover {
        font-weight: bold;
      }
    }
    .el-icon-delete,
    .el-icon-loading,
    .el-icon-success,
    .el-icon-refresh {
      color: $color;
    }

    &.fail {
      span {
        color: $color !important;
      }
      .el-icon-refresh {
        display: block;
      }
      &:hover {
        .el-icon-refresh {
          margin-left: 5px;
        }
      }
      background: $color_opacity !important;

      box-shadow: 0px 0px 0px 0px rgba(211, 58, 58, 0);
      animation: scale 0.6s ease-in-out;
    }
    @keyframes scale {
      0% {
        box-shadow: 0px 0px 0px 0px rgba(211, 58, 58, 0);
      }
      25% {
        box-shadow: 0px 0px 5px 2px rgba(211, 58, 58, 0.4);
      }
      50% {
        box-shadow: 0px 0px 0px 0px rgba(211, 58, 58, 0);
      }
      75% {
        box-shadow: 0px 0px 5px 2px rgba(211, 58, 58, 0.4);
      }
      100% {
        box-shadow: 0px 0px 0px 0px rgba(211, 58, 58, 0);
      }
    }
  }
}
</style>
