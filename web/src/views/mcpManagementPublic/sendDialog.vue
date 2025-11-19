<template>
  <div class="add-dialog">
    <el-dialog
      :title="$t('tool.square.send.title')"
      :visible.sync="dialogVisible"
      width="50%"
      :show-close="false"
      :close-on-click-modal="false"
    >
      <div>
        <el-form
          :model="ruleForm"
          :rules="rules"
          ref="ruleForm"
          label-width="130px"
        >
          <el-form-item :label="$t('tool.integrate.name')">
            <div>{{detail.name}}</div>
          </el-form-item>
          <el-form-item :label="$t('tool.integrate.from')">
            <div>{{detail.from}}</div>
          </el-form-item>
          <el-form-item :label="$t('tool.integrate.desc')" class="description-text">
            <div>{{detail.desc}}</div>
          </el-form-item>
          <el-form-item label="MCP ServerURL" prop="serverUrl">
            <el-input v-model="ruleForm.serverUrl"></el-input>
          </el-form-item>
          <el-form-item label=" " style="text-align: right">
            <el-button
              type="primary"
              size="mini"
              @click="handleTools"
              :disabled="isGetMCP"
              :loading="toolsLoading"
            >
              {{$t('tool.integrate.action')}}
            </el-button>
          </el-form-item>
        </el-form>
        <el-divider v-if="mcpList.length > 0"></el-divider>
        <ul class="mcpList" v-if="mcpList.length > 0">
          <li v-for="(item, index) in mcpList" :key="index">{{ item.name }}</li>
        </ul>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button
          type="primary"
          size="mini"
          :disabled="mcpList.length === 0"
          @click="submitForm('ruleForm')"
          :loading="publishLoading"
        >
          {{$t('tool.integrate.publish')}}
        </el-button>
        <el-button @click="handleCancel" size="mini">{{ $t('common.button.cancel') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import { getTools, setCreate } from "@/api/mcp.js";
import { isValidURL } from "@/utils/util";

export default {
  props: ["dialogVisible", "detail"],
  data() {
    const validateUrl = (rule, value, callback) => {
      if (!isValidURL(value)) {
        callback(new Error(this.$t('tool.integrate.serverUrlErr')));
      } else {
        callback();
      }
    };
    return {
      mcpList: [],
      ruleForm: {
        serverUrl: "",
      },
      rules: {
        serverUrl: [
          {
            required: true,
            message: this.$t('tool.integrate.serverUrlMsg'),
            trigger: "blur",
          },
          { validator: validateUrl, trigger: "blur" },
        ],
      },
      toolsLoading: false,
      publishLoading: false
    };
  },
  methods: {
    handleCancel() {
      this.clearForm();
      this.$refs["ruleForm"].clearValidate();
      this.$emit("handleClose", false);
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (valid) {
          const params = {
            name: this.detail.name,
            from: this.detail.from,
            sseUrl: this.ruleForm.serverUrl,
            desc: this.detail.desc,
            mcpSquareId: this.detail.mcpSquareId
          }
          this.publishLoading = true
          setCreate(params).then((res) => {
            if(res.code === 0){
              this.$message.success(this.$t('common.info.publish'))
              this.publishLoading = false
              this.handleCancel()
              // UpdateSendButtonStatus
              this.$emit("getIsCanSendStatus");
            }
          }).finally(() => this.publishLoading = false)
        }
      })
    },
    clearForm(){
      this.ruleForm.serverUrl = ''
      this.mcpList = []
    },
    handleTools() {
      this.toolsLoading = true
      this.$refs['ruleForm'].validate((valid) => {
        if (valid) {
          getTools({
            serverUrl: this.ruleForm.serverUrl,
          }).then((res) => {
            this.mcpList = res.data.tools || []
          }).finally(() => this.toolsLoading = false)
        }
      });
    },
  },
  computed: {
    isGetMCP() {
      return !isValidURL(this.ruleForm.serverUrl);
    },
  },
};
</script>
<style lang="scss" scoped>
.add-dialog {
  .el-button.is-disabled {
    &:active {
      background: transparent !important;
      border-color: #ebeef5 !important;
    }
  }
  .mcpList {
    list-style: none;
    li {
      padding: 10px;
      border: 1px solid #ddd;
      border-radius: 5px;
      margin-bottom: 10px;
      background: #fff;
    }
  }
  .description-text .el-form-item__content{
    line-height: 24px!important;
    padding: 10px 0;
  }
}
</style>