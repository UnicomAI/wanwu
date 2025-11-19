<template>
  <div class="mcp-detail" id="timeScroll">
    <span class="back" @click="back">{{ $t('menu.back') + $t('menu.tool') }}</span>
    <div class="mcp-title">
      <img class="logo" :src="(detail.avatar && detail.avatar.path) ? avatarSrc(detail.avatar.path) : defaultAvatar" alt=""/>
      <div :class="['info',{fold:foldStatus}]">
        <p class="name">{{ detail.name }}</p>
        <p v-if="detail.desc && detail.desc.length > 260" class="desc">
          {{ foldStatus ? detail.desc : detail.desc.slice(0, 268) + '...' }}
          <span class="arrow" v-show="detail.desc.length > 260" @click="fold">
            {{ foldStatus ? $t('common.button.fold') : $t('common.button.detail') }}
          </span>
        </p>
        <p v-else class="desc">{{ detail.desc }}</p>
      </div>
    </div>
    <div class="mcp-main">
      <div class="info">
        <!-- tabs -->
        <div class="mcp-tabs">
          <div :class="['mcp-tab',{ 'active': tabActive === 0 }]" @click="tabActive = 0">SSE URL及Tool</div>
          <div style="display: inline-block">
            <div :class="['mcp-tab',{ 'active': tabActive === 1 }]" @click="tabActive = 1">Streamable HTTP</div>
          </div>
        </div>

        <div v-if="tabActive === 0" >
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">SSE URL:</p>
              <el-input
                class="sse-url"
                v-model="detail.sseUrl"
                :readonly="true"
                style="margin-right: 20px"/>
            </div>
          </div>
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">{{ $t('tool.server.detail.example') }}</p>
              <el-input
                class="schema-textarea"
                v-model="detail.sseExample"
                :readonly="true"
                type="textarea"/>
            </div>
          </div>
        </div>
        <div v-if="tabActive === 1" >
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">Streamable HTTP:</p>
              <el-input
                class="sse-url"
                v-model="detail.streamableUrl"
                :readonly="true"
                style="margin-right: 20px"/>
            </div>
          </div>
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">{{ $t('tool.server.detail.example') }}</p>
              <el-input
                class="schema-textarea"
                v-model="detail.streamableExample"
                :readonly="true"
                type="textarea"/>
            </div>
          </div>
        </div>
        <div class="tool bg-border">
          <div class="tool-item">
            <div style="display: flex; align-items: center;">
              <p class="title">{{ $t('tool.server.bind.title') }}</p>
              <el-tooltip
                style="margin-left: 3px;"
                effect="dark"
                :content="$t('tool.server.bind.hint')"
                placement="right"
                popper-class="tooltip">
                <span class="el-icon-question question-tips"/>
              </el-tooltip>
            </div>
            <div>
              <el-button
                size="mini"
                @click="$refs.toolDialog.showDialog(detail)"
              >{{ $t('tool.server.bind.action') }}
              </el-button>
              <el-button
                size="mini"
                @click="$refs.addDialog.showToolDialog(mcpServerId)"
              >{{ $t('common.button.add') }}
              </el-button>
            </div>
            <el-table
              :data="detail.tools"
              style="width: 100%"
            >
              <el-table-column
                :label="$t('tool.server.bind.methodName')"
                prop="methodName"
                width="100"
              >
                <template #default="scope">
                  <el-input
                    :readonly="!scope.row.isEditing"
                    v-model="scope.row.methodName"
                    :placeholder="$t('common.input.placeholder') + $t('tool.server.bind.methodName')"
                  ></el-input>
                </template>
              </el-table-column>
              <el-table-column
                :label="$t('tool.server.bind.name')"
                prop="name"
                width="100"
              />
              <el-table-column
                :label="$t('tool.server.bind.type')"
                width="100"
              >
                <template #default="scope">
                  <div>
                    {{ appTypeMap[scope.row.type] || scope.row.type }}
                  </div>
                </template>
              </el-table-column>
              <el-table-column
                :label="$t('tool.server.bind.desc')"
                prop="desc"
              >
                <template #default="scope">
                  <el-input
                    :readonly="!scope.row.isEditing"
                    v-model="scope.row.desc"
                    :placeholder="$t('common.input.placeholder') + $t('tool.server.bind.desc')"
                  ></el-input>
                </template>
              </el-table-column>
              <el-table-column
                :label="$t('tool.server.bind.operate')"
                width="200"
              >
                <template #default="scope">
                  <el-button
                    v-if="scope.row.isEditing"
                    size="mini"
                    type="primary"
                    @click="handleEditTool(scope.row)"
                  >
                    {{ $t('common.confirm.confirm') }}
                  </el-button>
                  <el-button
                    v-else
                    size="mini"
                    @click="scope.row.isEditing = true"
                  >
                    {{ $t('common.button.edit') }}
                  </el-button>
                  <el-button
                    size="mini"
                    @click="handleDeleteTool(scope.row)"
                  >{{ $t('common.button.delete') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>


        <div class="tool bg-border">
          <div class="tool-item">
            <p class="title">{{$t('tool.server.detail.apiKey')}}</p>
            <el-button style="width: 100px" size="mini" type="primary" :disabled="detail.hasCustom"
                       @click="handleCreateApiKey">
              {{$t('tool.server.detail.action')}}
            </el-button>
            <el-table
              :data="apiKeyList"
              style="width: 100%"
            >
              <el-table-column
                :label="$t('tool.server.detail.key')"
                prop="apiKey"
                width="300"
              >
              </el-table-column>
              <el-table-column
                :label="$t('tool.server.detail.createTime')"
                prop="createdAt"
              />
              <el-table-column
                :label="$t('tool.server.detail.operate')"
                width="200"
              >
                <template slot-scope="scope">
                  <copyIcon :text="scope.row.apiKey" :showIcon="false" size="mini"/>
                  <el-button
                    size="mini"
                    @click="handleDeleteApiKey(scope.row)"
                  >{{ $t('common.button.delete') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>
    </div>
    <addDialog ref="addDialog" @handleFetch="fetchList"/>
    <toolDialog ref="toolDialog" @handleFetch="fetchList"/>
  </div>
</template>
<script>
import {getServer, editServerTool, deleteServerTool} from "@/api/mcp"
import {createApiKey, delApiKey, getApiKeyList} from "@/api/appspace"
import {avatarSrc} from "@/utils/util"
import CopyIcon from "@/components/copyIcon.vue";
import addDialog from "@/views/tool/tool/custom/addDialog.vue";
import toolDialog from "./toolDialog.vue";

const APPTYPE_MCPSERVER = 'mcpserver'
export default {
  components: {CopyIcon, addDialog, toolDialog},
  data() {
    return {
      tabActive: 0,
      defaultAvatar: require("@/assets/imgs/mcp_active.svg"),
      mcpServerId: '',
      detail: {},
      apiKeyList: [],
      foldStatus: false,
    };
  },
  watch: {
    $route: {
      handler() {
        this.initData()
      },
      // 深度ObserveListen
      deep: true
    }
  },
  computed: {
    appTypeMap() {
      return {
        'agent': this.$t('menu.app.agent'),
        'rag': this.$t('menu.app.rag'),
        'workflow': this.$t('menu.app.workflow'),
        'custom': this.$t('menu.app.custom'),
        'openapi': this.$t('menu.app.openapi'),
        'builtin': this.$t('menu.app.builtIn')
      }
    }
  },
  mounted() {
    this.initData()
  },
  methods: {
    avatarSrc,
    initData() {
      this.mcpServerId = this.$route.query.mcpServerId
      this.tabActive = 0
      getServer({mcpServerId: this.mcpServerId}).then((res) => {
        this.detail = res.data || {}
        this.detail.tools = (this.detail.tools || []).map(tool => ({...tool, isEditing: false}))
      })

      getApiKeyList({
        appId: this.mcpServerId,
        appType: APPTYPE_MCPSERVER
      }).then((res) => {
        this.apiKeyList = res.data || []
      })

      //scrolltotop
      const main = document.querySelector(".el-main > .page-container")
      if (main) main.scrollTop = 0
    },
    fold() {
      this.foldStatus = !this.foldStatus
    },
    fetchList() {
      this.initData()
    },
    handleEditTool(row) {
      editServerTool(row).then((res) => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.edit'))
          row.isEditing = false
        }
      })
    },
    handleDeleteTool(row) {
      deleteServerTool(row).then((res) => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delInfo'))
            this.detail.tools = this.detail.tools.filter(item => item.mcpServerToolId !== row.mcpServerToolId)
          }
        }
      )
    },
    handleCreateApiKey() {
      createApiKey({
        appId: this.mcpServerId,
        appType: APPTYPE_MCPSERVER
      }).then((res) => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'))
          this.apiKeyList = [...this.apiKeyList, res.data]
        }
      })
    },
    handleDeleteApiKey(row) {
      this.$confirm(
        this.$t("tool.server.detail.deleteHint"),
        this.$t("common.confirm.title"),
        {
          confirmButtonText: this.$t("common.confirm.confirm"),
          cancelButtonText: this.$t("common.confirm.cancel"),
          type: "warning",
        }
      ).then(() => {
        delApiKey({apiId: row.apiId}).then((res) => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delInfo'));
            this.apiKeyList = this.apiKeyList.filter(item => item.apiId !== row.apiId)
          }
        });
      })
    },
    back() {
      this.$router.push({path: '/tool?type=mcp&mcp=server'})
    },
  },
};
</script>
<style lang="scss">
.mcp-detail {
  padding: 20px;
  overflow: auto;
  margin: 20px;

  .back {
    color: $color;
    cursor: pointer;
  }

  .mcp-title {
    padding: 20px 0;
    display: flex;
    border-bottom: 1px solid #bfbfbf;

    .logo {
      width: 54px;
      height: 54px;
      object-fit: cover;
    }

    .info {
      position: relative;
      width: 100%;
      margin-left: 15px;

      .name {
        font-size: 16px;
        color: #5D5D5D;
        font-weight: bold;
      }

      .desc {
        margin-top: 10px;
        line-height: 22px;
        color: #9F9F9F;
        word-break: break-all;
      }

      .arrow {
        position: absolute;
        display: block;
        right: 0;
        bottom: -5px;
        cursor: pointer;
        color: $color;
        margin-left: 10px;
        font-size: 13px;
      }
    }

    .fold {
      height: auto;
    }
  }

  .mcp-main {
    display: flex;
    margin: 10px 0 0 0;

    .info {
      width: 100%;
      margin-right: 20px;

      .mcp-tabs{
        margin: 20px 0 0 0;
        .mcp-tab{
          display: inline-block;
          vertical-align: middle;
          width: 160px;
          height: 40px;
          border-bottom: 1px solid #333;
          line-height: 40px;
          text-align: center;
          cursor: pointer;
        }
        .active{
          background: #333;
          color: #fff;
          font-weight: bold;
        }
      }

      .tool {
        .tool-item {
          border-bottom: 1px solid #eee;

          .title {
            font-weight: bold;
            line-height: 46px;
          }

          .tool-item-bg {
            background: inherit;
            background-color: rgba(249, 249, 249, 1);
            border: none;
            border-radius: 10px;
            padding: 20px;
          }
        }

        .tool-item:last-child {
          border-bottom: none;
        }

        .schema-textarea {
          .el-textarea__inner {
            height: 200px !important;
          }
        }

        .install-intro-item {
          p {
            line-height: 26px;
            color: #333;
          }

          .install-intro-title {
            color: $color;
            margin-top: 10px;
            font-weight: bold;
          }
        }
      }
    }

    .right-recommend {
      width: 400px;
      overflow-y: auto;
      border-left: 1px solid #eee;
      padding: 20px;
      max-height: 900px;

      .recommend-item {
        position: relative;
        border: 1px solid $border_color;
        background: $color_opacity;
        margin-bottom: 15px;
        border-radius: 10px;
        padding: 20px 20px 20px 80px;
        text-align: left;
        cursor: pointer;

        .logo {
          width: 46px;
          height: 46px;
          object-fit: cover;
          position: absolute;
          left: 20px;
          border: 1px solid #fff;
          border-radius: 4px;
        }

        .name {
          color: #5D5D5D;
          font-weight: bold;
        }

        .intro {
          height: 36px;
          color: #5D5D5D;
          margin-top: 8px;
          font-size: 13px;
          overflow: hidden;
        }
      }
    }
  }

  .bg-border {
    margin-top: 20px;
    /*min-height: calc(100vh - 360px);*/
    background-color: rgba(255, 255, 255, 1);
    box-sizing: border-box;
    /*border:1px solid rgba(208, 167, 167, 1);*/
    border-radius: 10px;
    padding: 10px 20px;
    box-shadow: 2px 2px 15px $color_opacity;
  }
}

.el-button.is-disabled {
  background: #f9f9f9 !important;
}

.tooltip {
  max-width: 500px !important;
}
</style>
