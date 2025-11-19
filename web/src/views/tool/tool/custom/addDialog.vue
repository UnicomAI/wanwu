<template>
  <el-dialog
    :title="title"
    :visible.sync="dialogBasicVisible"
    width="50%"
    :before-close="cancel">
    <div class="action">
      <el-form
        :model="form"
        :rules="rules"
        ref="form">
        <el-col :span="24" class="left-col">
          <div class="action-form">
            <div class="block prompt-box" v-show="!dialogDetailVisible && !dialogToolVisible">
              <p class="block-title rl">{{ $t('tool.custom.avatar') }}</p>
              <upload-avatar
                :avatar="form.avatar"
                :default-avatar="defaultAvatar"
                @update-avatar="handleUpdateAvatar"
              />
            </div>
            <div class="block prompt-box" v-show="!dialogDetailVisible">
              <p class="block-title required-label rl">
                {{ dialogToolVisible ? $t('tool.custom.app') : $t('tool.custom.tool') }}</p>
              <el-form-item prop="name">
                <el-input class="name-input" v-model="form.name"
                          :placeholder="$t('common.input.placeholder') + (dialogToolVisible ? $t('tool.custom.app') : $t('tool.custom.tool'))"/>
              </el-form-item>
            </div>
            <div class="block prompt-box" v-show="!dialogToolVisible">
              <p class="block-title required-label rl">{{ $t('tool.custom.desc') }}</p>
              <div v-show="dialogDetailVisible">{{ form.description }}</div>
              <el-form-item prop="description" v-show="!dialogDetailVisible">
                <el-input class="name-input" v-model="form.description"
                          :placeholder="$t('common.input.placeholder') + $t('tool.custom.desc')"/>
              </el-form-item>
            </div>
            <div class="block prompt-box" v-show="!dialogDetailVisible">
              <p class="block-title required-label rl">{{ $t('tool.custom.apiAuth') }}</p>
              <div class="rl" @click="preAuthorize">
                <el-form-item prop="apiAuth">
                  <div class="api-key">{{ authTypeMap[form.apiAuth.authType] }}</div>
                </el-form-item>
                <img class="auth-icon" :src="require('@/assets/imgs/auth.png')" alt=""/>
              </div>
            </div>

            <div class="block prompt-box" v-show="!dialogDetailVisible">
              <p class="block-title required-label rl">Schema</p>
              <div class="rl">
                <div class="flex" style="margin-bottom: 10px">
                  <el-select v-model="example" :placeholder="$t('tool.custom.schema')" style="width:100%;"
                             @change="exampleChange">
                    <!--<el-option label="Import Template Example" value="json"></el-option>-->
                    <el-option :label="'JSON' + $t('tool.custom.example')" value="json"></el-option>
                    <el-option :label="'YAML' + $t('tool.custom.example')" value="yaml"></el-option>
                  </el-select>
                </div>
                <el-form-item prop="schema">
                  <el-input
                    class="schema-textarea"
                    v-model="form.schema"
                    @blur="listenerSchema"
                    :placeholder="$t('tool.custom.schemaHint')"
                    type="textarea"/>
                </el-form-item>
              </div>
            </div>

            <div class="block prompt-box">
              <p class="block-title required-label rl">{{ $t('tool.custom.api') }}</p>
              <div class="api-list">
                <el-form-item prop="apiTable">
                  <el-table
                    ref="apiTable"
                    :data="apiList"
                    border
                    size="mini"
                    class="api-table"
                    :header-cell-style="{ textAlign: 'center' }"
                  >
                    <el-table-column
                      v-if="dialogToolVisible"
                      type="selection"
                      width="55"
                      align="center">
                    </el-table-column>
                    <el-table-column
                      prop="name"
                      label="Name">
                    </el-table-column>
                    <el-table-column
                      prop="method"
                      label="Method">
                    </el-table-column>
                    <el-table-column
                      prop="path"
                      label="Path">
                    </el-table-column>
                  </el-table>
                </el-form-item>
              </div>
            </div>
            <div class="block prompt-box" v-show="!dialogDetailVisible">
              <p class="block-title rl">{{ $t('tool.custom.privacy') }}</p>
              <el-form-item prop="privacyPolicy">
                <el-input
                  class="name-input"
                  v-model="form.privacyPolicy"
                  :placeholder="$t('tool.custom.privacyHint')"/>
              </el-form-item>
            </div>
          </div>
        </el-col>
      </el-form>

      <!--Authentication Dialog-->
      <el-dialog
        :title="$t('tool.custom.auth.title')"
        :visible.sync="dialogAuthVisible"
        width="600px"
        append-to-body
        :close-on-click-modal="false"
        @close="beforeApiAuthClose"
      >
        <div class="action-form">
          <el-form :rules="apiAuthRules" ref="apiAuthForm" :inline="false" :model="form.apiAuth">
            <el-form-item :label="$t('tool.custom.auth.authType')">
              <el-select v-model="form.apiAuth.authType">
                <el-option label="None" value="none"/>
                <el-option :label="$t('tool.custom.auth.headerType')" value="api_key_header"/>
                <el-option :label="$t('tool.custom.auth.queryType')" value="api_key_query"/>
              </el-select>
            </el-form-item>
            <!--Request头-->
            <div v-show="form.apiAuth.authType === 'api_key_header'">
              <el-form-item :label="$t('tool.custom.auth.prefix')" prop="apiKeyHeaderPrefix">
                <el-select v-model="form.apiAuth.apiKeyHeaderPrefix">
                  <el-option label="Basic" value="basic"/>
                  <el-option label="Bearer" value="bearer"/>
                  <el-option label="Custom" value="custom"/>
                </el-select>
              </el-form-item>
              <el-form-item prop="apiKeyHeader">
                <template #label>
                  {{ $t('tool.custom.auth.header') }}
                  <el-tooltip
                    effect="dark"
                    :content="$t('tool.custom.auth.headerHint')"
                    placement="top-start"
                  >
                    <span class="el-icon-question tips"/>
                  </el-tooltip>
                </template>
                <el-input
                  class="desc-input"
                  v-model="form.apiAuth.apiKeyHeader"
                  placeholder="Authorization"
                  clearable
                />
              </el-form-item>
              <el-form-item :label="$t('tool.custom.auth.value')" prop="apiKeyValue">
                <el-input class="desc-input" v-model="form.apiAuth.apiKeyValue" placeholder="API key" clearable/>
              </el-form-item>
            </div>
            <!--QueryParameter-->
            <div v-show="form.apiAuth.authType === 'api_key_query'">
              <el-form-item prop="apiKeyQueryParam">
                <template #label>
                  {{ $t('tool.custom.auth.query') }}
                  <el-tooltip
                    effect="dark"
                    :content="$t('tool.custom.auth.queryHint')"
                    placement="top-start"
                  >
                    <span class="el-icon-question tips"/>
                  </el-tooltip>
                </template>
                <el-input
                  class="desc-input"
                  v-model="form.apiAuth.apiKeyQueryParam"
                  clearable
                />
              </el-form-item>
              <el-form-item :label="$t('tool.custom.auth.value')" prop="apiKeyValue">
                <el-input class="desc-input" v-model="form.apiAuth.apiKeyValue" placeholder="API key" clearable/>
              </el-form-item>
            </div>
          </el-form>
        </div>
        <span slot="footer" class="dialog-footer">
                <el-button @click="beforeApiAuthClose">{{ $t('common.button.cancel') }}</el-button>
                <el-button type="primary" @click="listenerApiKey">{{ $t('common.button.confirm') }}</el-button>
            </span>
      </el-dialog>
    </div>
    <span slot="footer" class="dialog-footer" v-show="!dialogDetailVisible">
        <el-button @click="cancel">{{ $t('common.button.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="submit"
          :loading="loading">{{ $t('common.button.confirm') }}</el-button>
    </span>
    <span slot="footer" class="dialog-footer" v-show="dialogDetailVisible">
        <el-button
          type="primary"
          @click="dialogDetailVisible = false; title = $t('tool.custom.editTitle')">{{
            $t('common.button.edit')
          }}</el-button>
    </span>
  </el-dialog>
</template>

<script>
import {getCustom, addCustom, editCustom, getSchema, addOpenapi} from "@/api/mcp";
import {schemaConfig} from '@/utils/schema.conf';
import uploadAvatar from "@/components/uploadAvatar.vue";

export default {
  components: {uploadAvatar},
  data() {
    const validateApiAuthFields = (rule, value, callback) => {
      if (this.form.apiAuth.authType === 'api_key_header' &&
        (!this.form.apiAuth.apiKeyValue || !this.form.apiAuth.apiKeyHeader)) {
        callback(new Error(rule.message));
      } else if (this.form.apiAuth.authType === 'api_key_query' &&
        (!this.form.apiAuth.apiKeyValue || !this.form.apiAuth.apiKeyQueryParam)) {
        callback(new Error(rule.message));
      } else {
        callback();
      }
    }
    const validateApiTableFields = (rule, value, callback) => {
      if (this.dialogToolVisible && this.$refs.apiTable.selection.length === 0) {
        callback(new Error(rule.message));
      } else {
        callback();
      }
    }
    return {
      dialogBasicVisible: false,
      dialogDetailVisible: false,
      dialogToolVisible: false,
      title: '',
      apiList: [],
      defaultAvatar: require("@/assets/imgs/toolImg.png"),
      example: '',
      form: {
        description: '',
        customToolId: '',
        mcpServerId: '',
        name: '',
        schema: '',
        privacyPolicy: '',
        apiAuth: {
          authType: 'none',
          apiKeyValue: '',
          apiKeyHeader: '',
          apiKeyHeaderPrefix: "basic",
          apiKeyQueryParam: '',
        },
        avatar: {
          key: "",
          path: ""
        },
      },
      //AuthenticateForm
      dialogAuthVisible: false,
      rules: {
        description: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        name: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        schema: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        apiAuth: [{validator: validateApiAuthFields, message: this.$t('tool.custom.apiAuthHint'), trigger: 'blur'}],
        apiTable: [{validator: validateApiTableFields, message: this.$t('tool.custom.apiHint'), trigger: 'blur'}],
      },
      apiAuthRules: {
        apiKeyValue: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        apiKeyHeader: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        apiKeyQueryParam: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        apiKeyHeaderPrefix: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
      },
      schemaConfig: schemaConfig,
      loading: false
    }
  },
  computed: {
    authTypeMap() {
      return {
        none: 'None',
        api_key_header: this.$t('tool.custom.auth.headerType'),
        api_key_query: this.$t('tool.custom.auth.queryType')
      }
    },
  },
  methods: {
    showDialog(customToolId, dialogDetailVisible) {
      this.dialogDetailVisible = dialogDetailVisible
      this.dialogBasicVisible = true
      if (customToolId) {
        if (!dialogDetailVisible) this.title = this.$t('tool.custom.editTitle')
        const params = {
          customToolId: customToolId
        }
        getCustom(params)
          .then((res) => {
            const {list, ...form} = res.data
            this.form = form
            if (dialogDetailVisible) {
              this.title = form.name
            }
            this.listenerSchema()
          })
      } else this.title = this.$t('tool.custom.addTitle')
    },
    showToolDialog(mcpServerId) {
      this.form.mcpServerId = mcpServerId
      this.dialogBasicVisible = true
      this.dialogToolVisible = true
      this.form.description = ' '
      this.title = this.$t('tool.custom.toolTitle')
    },
    handleUpdateAvatar(avatar) {
      this.form = {...this.form, avatar: avatar};
    },
    exampleChange(value) {
      this.form.schema = this.schemaConfig[value]
      this.listenerSchema()
    },
    beforeApiAuthClose() {
      this.$refs.apiAuthForm.clearValidate()
      this.dialogAuthVisible = false
    },
    listenerApiKey() {
      this.rules.apiAuth[0].validator({}, null, (error) => {
        if (!error) {
          this.dialogAuthVisible = false;
        }
      });
    },
    listenerSchema() {
      const params = JSON.stringify({
        schema: this.form.schema
      })
      getSchema(params)
        .then((res) => {
          this.apiList = res.data.list || []
        })
    },
    preAuthorize() {
      this.form.apiAuth.apiKeyHeaderPrefix = this.form.apiAuth.apiKeyHeaderPrefix || "basic"
      this.dialogAuthVisible = true
    },
    submit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return;
        this.loading = true
        switch (this.form.apiAuth.authType) {
          case 'none':
            this.form.apiAuth.apiKeyValue = ''
            this.form.apiAuth.apiKeyHeader = ''
            this.form.apiAuth.apiKeyHeaderPrefix = ''
            this.form.apiAuth.apiKeyQueryParam = ''
            break
          case 'api_key_query':
            this.form.apiAuth.apiKeyHeader = ''
            this.form.apiAuth.apiKeyHeaderPrefix = ''
            break
          case 'api_key_header':
            this.form.apiAuth.apiKeyQueryParam = ''
            break
        }
        const params = {
          ...this.form
        }
        if (this.form.customToolId) {
          editCustom(params)
            .then((res) => {
              if (res.code === 0) {
                this.$message.success(this.$t('common.info.edit'))
                this.$emit("handleFetch", false)
                this.cancel()
              }
            }).finally(() => this.loading = false)
        } else {
          delete params.customToolId
          if (this.dialogToolVisible) {
            params.methodNames = this.$refs.apiTable.selection.map(item => item.name)
            addOpenapi(params).then((res) => {
              if (res.code === 0) {
                this.$message.success(this.$t('common.info.create'))
                this.$emit("handleFetch", false)
                this.cancel()
              }
            }).finally(() => this.loading = false)
          } else {
            addCustom(params).then((res) => {
              if (res.code === 0) {
                this.$message.success(this.$t('common.info.create'))
                this.$emit("handleFetch", false)
                this.cancel()
              }
            }).finally(() => this.loading = false)
          }
        }
      })
    },
    cancel() {
      this.$emit("handleClose", false)
      this.loading = false
      this.dialogBasicVisible = false
      this.dialogDetailVisible = false
      this.dialogToolVisible = false
      this.apiList = []
      this.example = ''
      this.title = ''
      this.$refs.form.clearValidate()
      this.form = {
        description: '',
        customToolId: '',
        mcpServerId: '',
        name: '',
        schema: '',
        privacyPolicy: '',
        apiAuth: {
          authType: 'none',
          apiKeyValue: '',
          apiKeyHeader: '',
          apiKeyHeaderPrefix: "basic",
          apiKeyQueryParam: '',
        }
      }
    }
  },

}
</script>

<style lang="scss" scoped>
/deep/ .el-radio__input.is-checked .el-radio__inner {
  border-color: #D33A3A !important;
  background: transparent !important;
}

/deep/ .el-radio__input.is-checked .el-radio__inner::after {
  background: #eb0a0b !important;
  width: 7px !important;
  height: 7px !important;
}

::selection {
  color: #1a2029;
  background: #c8deff;
}

.left-col {
  // background-color: #fafafa;
  overflow: auto;
  height: 100%;

  .left-col-header {
    width: 100%;
    padding: 30px 40px;
    text-align: center;

    .back-icon {
      position: absolute;
      left: 35px;
      font-size: 14px;
      cursor: pointer;
      border-radius: 20px;
      border: 1px solid #e1e1e1;
      padding: 6px;
      color: #444;

      &:hover {
        font-weight: bold;
      }
    }

    .header-title {
      font-size: 18px;
      font-weight: bold;
      color: #303133;
    }

    .bt-box {
      position: absolute;
      width: 160px;
      height: 30px;
      right: 20px;
      top: 0;
      bottom: 0;
      margin: auto;

      .del-bt {
        margin-left: 10px;
      }
    }
  }

  .action-form {
    padding: 0 40px;

    /deep/ .schema-textarea {
      .el-textarea__inner {
        height: 200px !important;
      }
    }

    .api-key {
      background-color: transparent !important;
      border: 1px solid #d3d7dd !important;
      padding: 0 15px;
      -webkit-appearance: none;
      background-image: none;
      border-radius: 4px;
      box-sizing: border-box;
      color: #606266;
      display: inline-block;
      height: 40px;
      line-height: 40px;
      outline: 0;
      transition: border-color .2s cubic-bezier(.645, .045, .355, 1);
      width: 100%;
    }

    .auth-icon {
      position: absolute;
      right: 0;
      height: 39px;
      top: 0;
      cursor: pointer;
      border-left: 1px solid #d3d7dd;
      padding: 7px 9px;
    }
  }

}

.right-col {
  height: 100%;
  // background-color: #f6f7f9;
  .right-title {
    line-height: 84px;
    font-size: 18px;
    font-weight: bold;
    text-align: center;
    color: #303133;
  }

  .smart-center {
    min-width: 0;
    height: calc(100% - 84px);
    flex: 1;
    background-size: 100% 100%;
    position: relative;
  }
}

/* General */
.action {
  position: relative;
  height: 100%;

  /deep/ .el-input__inner, /deep/ .el-textarea__inner {
    background-color: transparent !important;
    border: 1px solid #d3d7dd !important;
    font-family: 'Microsoft YaHei', Arial, sans-serif;
    padding: 15px;
  }

  .flex {
    width: 100%;
    display: flex;
    justify-content: space-between;
  }

  .block {
    margin-bottom: 20px;

    .block-title {
      line-height: 30px;
      font-size: 15px;
      font-weight: bold;
    }

    .required-label::after {
      content: '*';
      position: absolute;
      color: #eb0a0b;
      font-size: 20px;
      margin-left: 4px;
    }

    .block-tip {
      color: #919eac;
    }
  }

  .el-input__count {
    color: #909399;
    // background: #fafafa;
    position: absolute;
    font-size: 12px;
    bottom: 5px;
    right: 10px;
  }
}

.action-form /deep/ .el-form-item__label {
  display: block;
  width: 100%;
  text-align: left;
  float: none;
}

.api-list {
  .api-table /deep/ .el-table__body tr td,
  .api-table /deep/ .el-table__header tr th,
  .api-table /deep/ .el-table__body tr:hover > td {
    background-color: transparent !important;
  }
}

.tips {
  margin-left: 2px;
}
</style>
