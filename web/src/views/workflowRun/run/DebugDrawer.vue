<template>
  <div class="debug">
    <div class="config-header">
      <!--<i class="el-icon-close close-icon" @click="preClose"></i>-->
      <div style="color: #888; cursor: pointer" @click="$router.go(-1)">Back</div>
      <div class="header-name">{{startNode.workflowName || '--'}}</div>
    </div>
    <div class="form">
      <div class="form-item" v-for="(n, i) in startNode.data.outputs" :key="i">
        <div class="form-item--label">
          <span class="name">{{ n.name }}</span>
          <span class="required" v-if="n.required">*</span>
          <el-tooltip
            class="item"
            effect="dark"
            :content="n.desc || 'NoDescription'"
            placement="top-start"
          >
            <i class="el-icon-question desc-icon"></i>
          </el-tooltip>
          <span class="type">{{ n.type }}</span>
          <el-select
            size="mini"
            v-model="n.upLoadType"
            v-if="n.type === 'fileUrl'"
            placeholder="Please select"
            @change="handleUrlChange(i)"
          >
            <el-option :label="'UploadFile'" :value="'0'"> </el-option>
            <el-option :label="'EnterUrl'" :value="'1'"> </el-option>
          </el-select>
        </div>
        <div v-if="n.type !== 'fileUrl'">
          <el-input
            v-if="
              n.type != 'array' ||
              (n.type === 'fileUrl' && n.upLoadType === '1')
            "
            class="form-item--input"
            v-model="n.value.content"
            placeholder="Please fill inEnterParameterValue"
          ></el-input>
          <ArrayEditor
            v-else
            style="height: 200px; overflow: auto"
            :value="n.value.content"
            :language="'json'"
            :n="i"
            @handleChange="handleChange"
            :theme="'vs'"
          />
        </div>
        <div v-else style="padding-top: 10px">
          <el-input
            v-if="n.upLoadType === '1'"
            class="form-item--input"
            v-model="n.value.content"
            placeholder="Please fill inEnterParameterValue"
          ></el-input>
          <Upload
            v-if="n.upLoadType === '0'"
            :ref="'upload' + i"
            :index="i"
            @handleUploadSuccess="handleUploadSuccess"
            @handleDisabled="handleDisabled"
            @handleCancel="handleCancel"
          ></Upload>
        </div>
      </div>

      <div
        class="params-content"
        v-if="startNode.data.settings.staticAuthToken"
      >
        <div class="params-content-item">
          <span>token：</span>
          <span>{{ startNode.data.settings.staticAuthToken.slice(0, 6) + "******" }}</span>
        </div>
      </div>
    </div>
    <div class="btns">
      <el-button
        size="mini"
        type="primary"
        @click="preRun"
        :disabled="runDisabled"
      >
        <i class="el-icon-caret-right"></i>&nbsp;&nbsp;运row
      </el-button>
      <el-button
        v-if="isStream && sessionStatus !== -1"
        size="mini"
        @click="stopSse"
        type="danger"
      >
        <i class="el-icon-video-pause"></i>&nbsp;&nbsp;终止输出
      </el-button>
    </div>

    <div class="result answer-content" v-if="isStream">
      <p class="result-title" v-show="runResponse">Run Result:</p>
      <div
        v-if="showDSBtn(runResponse)"
        class="deepseek"
        @click.prevent.stop="toggle($event)"
      >
        <img src="@/assets/imgs/sikao.png" alt="" />
        {{ thinkText }}
        <i
          v-bind:class="{
            'el-icon-arrow-down': !isOpen,
            'el-icon-arrow-up': isOpen,
          }"
        ></i>
      </div>
      <div v-bind:class="{ 'ds-res': showDSBtn(runResponse) }">
        <div class="answer_content_box" v-html="replaceHTML(runResponse)"></div>
      </div>
    </div>
  </div>
</template>

<script>
import sseMethod from "@/mixins/sseMethod.js";
import ArrayEditor from "@/views/arrayEditor";
import Upload from "./upload.vue";
import { getQueryString } from "@/utils/util.js";
import { getWorkFlowParams } from "@/api/workflow";
import { mapGetters } from "vuex"

export default {
  components: { ArrayEditor, Upload },
  mixins: [sseMethod],
  data() {
    return {
      source: [], // 储存目前Upload of Interface
      isStream: getQueryString("isStream"),
      startNode: {
        data: {
          outputs: [],
          settings: {}
        },
        workflowName: ''
      },
      runDisabled: false,
      thinkText: "",
      isOpen: true,
      workflowId: ''
    };
  },
  computed: {
    ...mapGetters('app', ['sessionStatus'])
  },
  created() {
    this.workflowId = this.$route.query.id
    this.getWorkFlowParams()
  },
  methods: {
    getWorkFlowParams() {
      getWorkFlowParams({
        workflowID: this.workflowId
      }).then((res) => {
        this.startNode = res.data || {}
      })
    },
    // Upload功能Add代码
    handleUploadSuccess(obj) {
      /** obj
       * obj.index 下标 第几Parameter
       * obj.url Back of  downloadUrl
       * obj.fileId Back of  fileId
       * obj.file 选in of FileInformation */
      this.startNode.data.outputs[obj.index].value.content = obj.fileId;
    },
    // UploadinTriggerEvent,Back目前YesNo在UploadFile
    handleDisabled(val) {
      // val: true 代Tablein progressUpload
      this.runDisabled = val;
    },
    // 储存目前in progressUpload of InterfaceInformation
    handleCancel(obj) {
      this.source.push(obj.source);
    },
    handleUrlChange(i) {
      this.startNode.data.outputs[i].value.content = "";
    },
    // ******************* end  **********************
    setArray(data) {
      try {
        let arr = JSON.parse(data);
        return arr;
      } catch (error) {
        return [data];
      }
    },
    handleChange(value, i) {
      this.startNode.data.outputs[i].value.content = value;
    },
    preRun() {
      this.runDisabled = true;

      let params = {};
      this.startNode.data.outputs.forEach((n) => {
        params[n.name] =
          n.type === "array" && n.value.content
            ? this.setArray(n.value.content)
            : n.value.content;
      });
      this.$emit("doDebug", params);
    },
    stopSse() {
      if (this.sessionStatus !== -1) {
        this.setStoreSessionStatus(-1);
        this.sseOnCloseCallBack();
        this._print && this._print.stop();
      }
    },
    toggle(event) {
      // console.log(event.target.className);
      const name = event.target.className;
      if (
        name === "deepseek" ||
        name === "el-icon-arrow-up" ||
        name === "el-icon-arrow-down"
      ) {
        // this.showDs = !this.showDs;
        this.isOpen = !this.isOpen;

        let elm = null;
        if (name === "el-icon-arrow-up" || name === "el-icon-arrow-down") {
          elm = event.target.parentNode.parentNode
            .getElementsByClassName("answer_content_box")[0]
            .getElementsByTagName("section")[0];
        } else {
          elm = event.target.parentNode
            .getElementsByClassName("answer_content_box")[0]
            .getElementsByTagName("section")[0];
        }
        if (!this.isOpen) {
          elm.className = "hideDs";
        } else {
          elm.className = "";
        }
      }
    },
    showDSBtn(data) {
      const pattern = /<\/?think>/;
      const matches = data.match(pattern);
      if (!matches) {
        return false;
      }
      return true;
    },
    replaceHTML(data) {
      let _data = data;
      let a = new RegExp("<think>");
      let b = new RegExp("</think>");
      if (b.test(data)) {
        this.thinkText = "already深度思考";
      } else {
        if (this.sessionStatus === -1) {
          if (a.test(data) && !b.test(data)) {
            this.thinkText = "思考alreadyStop";
          }
        } else {
          this.thinkText = "Thinking...";
        }
      }

      // IfNoBack前缀，Then补上
      if (b.test(data) && !a.test(data)) {
        _data = "<think>" + data;
      }
      _data = _data.replace(/<think>\s*\n/g, "<think>");

      return _data.replace(/think>/g, "section>");
    },
  },
};
</script>

<style lang="scss" scoped>
.debug {
  position: absolute;
  display: flex;
  flex-direction: column;
  width: 420px;
  height: calc(100% - 20px);
  top: 10px;
  bottom: 10px;
  left: 10px;
  padding: 20px;
  border-left: 1px solid #ddd;
  background-color: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 4px 0 rgba(0,0,0,0.15);
}
.config-header {
  position: relative;
  .close-icon {
    position: absolute;
    right: 0;
    top: 0;
    font-size: 16px;
    padding: 2px;
    cursor: pointer;
  }
  .header-icon {
    vertical-align: middle;
    width: 26px;
    margin-right: 10px;
    object-fit: contain;
  }
  .header-name {
    font-size: 15px;
    border-bottom: 1px solid #dedede;
    padding:  20px 0 8px;
  }
  .desc {
    padding: 10px 0;
    border-bottom: 1px solid #e8e9eb;
  }
}
.form {
  margin-top: 10px;
  .form-item {
    margin: 20px 0;
    .form-item--label {
      margin-bottom: 10px;
      .name {
        font-size: 12px;
      }
      .required {
        color: #e60001;
        margin-left: 3px;
      }
      .desc-icon {
        margin: 0 5px;
        font-size: 16px;
      }
      .type {
        padding: 0 4px;
        text-align: center;
        color: #5c5f66;
        border-radius: 4px;
        background: #e8e9eb;
        font-size: 12px;
        font-weight: 400;
        line-height: 20px;
      }
    }
    .form-item--input {
    }
  }
}
/deep/.el-select {
  float: right;
  width: 99px;

  .el-input .el-input__inner {
    height: 28px;
    font-size: 12px;
  }
}
.run-bt {
  margin-top: 10px;
  width: 100%;
  i {
    font-size: 20px;
  }
}
.btns {
  /*display: flex;
  flex-direction: column;*/
  text-align: right;
  margin-top: 20px;
  .el-button {
    margin-left: 10px;
  }
}
.cancel-bt {
  margin-top: 10px;
  width: 100%;
  margin-left: 0;
  i {
    font-size: 20px;
  }
}
.result {
  height: calc(100% - 320px);
  overflow: auto;
  padding: 10px 0;
  .result-title {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 10px;
  }
  .result-content {
    margin-top: 30px;
    .result-item {
      margin: 10px 0;
      div {
      }
    }
  }
  .green {
    color: #67c23a;
  }
  .red {
    color: #e60001;
  }
}

.params-content {
  margin-top: 4px;
  .params-content-item {
    padding: 4px 0;
    display: flex;
    background-color: #e8e9eb;
    border-radius: 4px;
    color: #5c5f66;
    span {
      margin-left: 5px;
      font-size: 12px;
      display: inline-block;
      vertical-align: middle;
    }
  }
}

.answer-content {
  // width: calc(100% - 30px);
  position: relative;
  margin-left: 14px;
  color: #333;
  // bottom: 40px;
  white-space: normal !important;
  li {
    display: revert !important;
  }
  img {
    max-width: 100%;
  }
  .answer_content_box {
    // line-height: 3;
    // white-space: pre-wrap;

    /deep/ p {
      margin-bottom: 20px;
    }
    /deep/li {
      margin: 0 0 10px 0;
      p {
        margin-bottom: 0;
      }
      // line-height: 1;
    }
    & /deep/ :last-child {
      margin-bottom: 0 !important;
    }
  }
}
.ds-res {
  /deep/ section {
    color: #8b8b8b;
    position: relative;
    font-size: 13px;
    padding-left: 15px;
    margin-bottom: 10px;
    font-weight: 700;
    // white-space: pre-wrap;
    * {
      font-size: 13px;
      // white-space: pre-wrap;
    }
  }
  /deep/ section::before {
    content: "";
    position: absolute;
    height: 100%;
    width: 2px;
    background: #ddd;
    left: 1px;
  }
  /deep/.hideDs {
    display: none;
  }
}

.deepseek {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  color: rgb(1, 3, 56);
  font-weight: bold;
  background: #ebeff8;
  cursor: pointer;
  padding: 0 10px;
  padding-left: 15px;
  border-radius: 5px;
  margin-bottom: 5px;
  font-size: 12px;
  width: 150px;
  img {
    width: 12px;
    height: 12px;
  }
}
</style>
