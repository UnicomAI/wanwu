<template>
  <div
    :id="readonly ? '' : 'timeScroll'"
    :class="[
      'mcp-detail',
      { 'page-wrapper': !readonly, 'readonly-embed': readonly },
    ]"
  >
    <span v-if="!readonly" class="back" @click="back">
      {{ $t('menu.back') + $t('menu.resource') }}
    </span>
    <div v-if="!readonly" class="mcp-title">
      <img
        class="logo"
        :src="
          detail.avatar && detail.avatar.path
            ? avatarSrc(detail.avatar.path)
            : defaultAvatar
        "
        alt=""
      />
      <div :class="['info', { fold: foldStatus }]">
        <p class="name">{{ detail.name }}</p>
        <p v-if="detail.desc && detail.desc.length > 260" class="desc">
          {{ foldStatus ? detail.desc : detail.desc.slice(0, 268) + '...' }}
          <span class="arrow" v-show="detail.desc.length > 260" @click="fold">
            {{
              foldStatus ? $t('common.button.fold') : $t('common.button.detail')
            }}
          </span>
        </p>
        <p v-else class="desc">{{ detail.desc }}</p>
      </div>
    </div>
    <div class="mcp-main">
      <div class="info">
        <!-- tabs -->
        <div class="tabs">
          <div
            :class="['tab', { active: tabActive === 0 }]"
            @click="tabActive = 0"
          >
            SSE URL及工具
          </div>
          <div style="display: inline-block">
            <div
              :class="['tab', { active: tabActive === 1 }]"
              @click="tabActive = 1"
            >
              Streamable HTTP
            </div>
          </div>
        </div>

        <div v-if="tabActive === 0">
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">SSE URL:</p>
              <el-input
                class="sse-url"
                v-model="detail.sseUrl"
                :readonly="true"
                style="margin-right: 20px"
              />
            </div>
          </div>
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">{{ $t('tool.server.detail.example') }}</p>
              <el-input
                class="schema-textarea"
                v-model="detail.sseExample"
                :readonly="true"
                type="textarea"
              />
            </div>
          </div>
        </div>
        <div v-if="tabActive === 1">
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">Streamable HTTP:</p>
              <el-input
                class="sse-url"
                v-model="detail.streamableUrl"
                :readonly="true"
                style="margin-right: 20px"
              />
            </div>
          </div>
          <div class="tool bg-border">
            <div class="tool-item">
              <p class="title">{{ $t('tool.server.detail.example') }}</p>
              <el-input
                class="schema-textarea"
                v-model="detail.streamableExample"
                :readonly="true"
                type="textarea"
              />
            </div>
          </div>
        </div>
        <div class="tool bg-border">
          <div class="tool-item">
            <div style="display: flex; align-items: center">
              <p class="title">{{ $t('tool.server.bind.title') }}</p>
              <el-tooltip
                style="margin-left: 3px"
                effect="dark"
                :content="$t('tool.server.bind.hint')"
                placement="right"
                popper-class="tooltip"
              >
                <span class="el-icon-question question-tips" />
              </el-tooltip>
            </div>
            <div v-if="!readonly">
              <el-button
                size="mini"
                @click="$refs.toolDialog.showDialog(detail)"
              >
                {{ $t('tool.server.bind.action') }}
              </el-button>
              <el-button
                size="mini"
                @click="$refs.addDialog.showToolDialog(mcpServerId)"
              >
                {{ $t('common.button.add') }}
              </el-button>
            </div>
            <el-table :data="detail.tools" style="width: 100%">
              <el-table-column
                :label="$t('tool.server.bind.methodName')"
                prop="methodName"
                width="100"
              >
                <template #default="scope">
                  <el-input
                    :readonly="!scope.row.isEditing"
                    v-model="scope.row.methodName"
                    :placeholder="
                      $t('common.input.placeholder') +
                      $t('tool.server.bind.methodName')
                    "
                  ></el-input>
                </template>
              </el-table-column>
              <el-table-column
                :label="$t('tool.server.bind.name')"
                prop="name"
                width="100"
              />
              <el-table-column :label="$t('tool.server.bind.type')" width="100">
                <template #default="scope">
                  <div>
                    {{ appTypeMap[scope.row.type] || scope.row.type }}
                  </div>
                </template>
              </el-table-column>
              <el-table-column :label="$t('tool.server.bind.desc')" prop="desc">
                <template #default="scope">
                  <el-input
                    :readonly="!scope.row.isEditing"
                    v-model="scope.row.desc"
                    :placeholder="
                      $t('common.input.placeholder') +
                      $t('tool.server.bind.desc')
                    "
                  ></el-input>
                </template>
              </el-table-column>
              <el-table-column
                v-if="!readonly"
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
                  <el-button size="mini" @click="handleDeleteTool(scope.row)">
                    {{ $t('common.button.delete') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>

        <div v-if="!readonly" class="tool bg-border">
          <div class="tool-item">
            <p class="title">{{ $t('tool.server.detail.apiKey') }}</p>
            <el-button
              v-if="!readonly"
              style="width: 100px"
              size="mini"
              type="primary"
              :disabled="detail.hasCustom"
              @click="handleCreateApiKey"
            >
              {{ $t('tool.server.detail.action') }}
            </el-button>
            <el-table
              v-if="apiKeyList.length"
              :data="apiKeyList"
              style="width: 100%"
            >
              <el-table-column
                :label="$t('tool.server.detail.key')"
                prop="apiKey"
                width="300"
              ></el-table-column>
              <el-table-column
                :label="$t('tool.server.detail.createTime')"
                prop="createdAt"
              />
              <el-table-column
                v-if="!readonly"
                :label="$t('tool.server.detail.operate')"
                width="200"
              >
                <template slot-scope="scope">
                  <copyIcon
                    :text="scope.row.apiKey"
                    :showIcon="false"
                    size="mini"
                  />
                  <el-button size="mini" @click="handleDeleteApiKey(scope.row)">
                    {{ $t('common.button.delete') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>
    </div>
    <addDialog v-if="!readonly" ref="addDialog" @handleFetch="fetchList" />
    <toolDialog v-if="!readonly" ref="toolDialog" @handleFetch="fetchList" />
  </div>
</template>
<script>
import { getServer, editServerTool, deleteServerTool } from '@/api/mcp';
import { createApiKey, delApiKey, getApiKeyList } from '@/api/appspace';
import { avatarSrc } from '@/utils/util';
import CopyIcon from '@/components/copyIcon.vue';
import addDialog from '@/views/tool/tool/custom/addDialog.vue';
import toolDialog from './toolDialog.vue';

const APPTYPE_MCPSERVER = 'mcpserver';
export default {
  name: 'McpServiceServerDetail',
  components: { CopyIcon, addDialog, toolDialog },
  props: {
    readonly: {
      type: Boolean,
      default: false,
    },
    embeddedDetail: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      tabActive: 0,
      defaultAvatar: require('@/assets/imgs/mcp_active.svg'),
      mcpServerId: '',
      detail: {},
      apiKeyList: [],
      foldStatus: false,
    };
  },
  watch: {
    $route: {
      handler() {
        this.initData();
      },
      // 深度观察监听
      deep: true,
    },
    embeddedDetail: {
      handler(val) {
        if (!this.readonly) return;
        this.detail = val || {};
        this.apiKeyList = val?.apiKeyList || [];
      },
      immediate: true,
    },
  },
  computed: {
    appTypeMap() {
      return {
        agent: this.$t('menu.app.agent'),
        rag: this.$t('menu.app.rag'),
        workflow: this.$t('menu.app.workflow'),
        custom: this.$t('menu.app.custom'),
        openapi: this.$t('menu.app.openapi'),
        builtin: this.$t('menu.app.builtIn'),
      };
    },
  },
  mounted() {
    this.initData();
  },
  methods: {
    avatarSrc,
    initData() {
      if (this.readonly) return;
      this.mcpServerId = this.$route.query.mcpServerId;
      this.tabActive = 0;
      getServer({ mcpServerId: this.mcpServerId }).then(res => {
        this.detail = res.data || {};
        this.detail.tools = (this.detail.tools || []).map(tool => ({
          ...tool,
          isEditing: false,
        }));
      });

      getApiKeyList({
        appId: this.mcpServerId,
        appType: APPTYPE_MCPSERVER,
      }).then(res => {
        this.apiKeyList = res.data || [];
      });

      //滚动到顶部
      const main = document.querySelector('.el-main > .page-container');
      if (main) main.scrollTop = 0;
    },
    fold() {
      this.foldStatus = !this.foldStatus;
    },
    fetchList() {
      this.initData();
    },
    handleEditTool(row) {
      editServerTool(row).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.edit'));
          row.isEditing = false;
        }
      });
    },
    handleDeleteTool(row) {
      deleteServerTool(row).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.delete'));
          this.detail.tools = this.detail.tools.filter(
            item => item.mcpServerToolId !== row.mcpServerToolId,
          );
        }
      });
    },
    handleCreateApiKey() {
      createApiKey({
        appId: this.mcpServerId,
        appType: APPTYPE_MCPSERVER,
      }).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          this.apiKeyList = [...this.apiKeyList, res.data];
        }
      });
    },
    handleDeleteApiKey(row) {
      this.$confirm(
        this.$t('tool.server.detail.deleteHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(() => {
        delApiKey({ apiId: row.apiId }).then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delete'));
            this.apiKeyList = this.apiKeyList.filter(
              item => item.apiId !== row.apiId,
            );
          }
        });
      });
    },
    back() {
      this.$router.push({ path: '/mcpService?mcp=server' });
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/tabs.scss';
@import '@/style/squareDetail.scss';

.mcp-detail {
  .mcp-main {
    .info {
      width: 100%;
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

        ::v-deep .el-table {
          margin-top: 10px;
        }

        .schema-textarea {
          ::v-deep .el-textarea__inner {
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
  }
}

.tooltip {
  max-width: 500px !important;
}

.mcp-detail.readonly-embed {
  padding: 0;
  overflow: visible;
}
</style>
