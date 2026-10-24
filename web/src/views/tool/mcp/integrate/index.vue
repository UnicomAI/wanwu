<template>
  <div class="mcp-content-box customize">
    <div class="mcp-content">
      <McpOAuthPanel @connected="fetchList" />
      <div class="card-search card-search-cust">
        <div>
          <p class="card-search-des" style="display: flex; align-items: center">
            <span>{{ $t('tool.integrate.slogan') }}</span>
            <LinkIcon type="mcp" />
          </p>
        </div>
        <div>
          <search-input
            :placeholder="$t('tool.integrate.search')"
            ref="searchInput"
            @handleSearch="fetchList"
          />
          <el-button size="mini" type="primary" @click="handleAddMCP">
            {{ $t('common.button.import') }}
          </el-button>
        </div>
      </div>

      <div class="card-box scroll-card-container">
        <div class="card card-item-create">
          <div class="app-card-create" @click="handleAddMCP">
            <div class="create-img-wrap">
              <img
                class="create-img"
                src="@/assets/imgs/card_create_icon_mcp.svg"
                alt=""
              />
            </div>
            <span>{{ $t('tool.integrate.create') }}</span>
          </div>
        </div>
        <div
          v-if="list && list.length"
          class="card"
          v-for="(item, index) in list"
          :key="index"
          @click.stop="handleClick(item)"
        >
          <div class="card-title">
            <img
              class="common-card-logo"
              :src="
                item.avatar && item.avatar.path
                  ? avatarSrc(item.avatar.path)
                  : defaultAvatar
              "
              alt=""
            />
            <div class="mcp_detailBox">
              <span class="mcp_name">{{ item.name }}</span>
              <span class="mcp_from from_tag">
                <label>
                  {{ item.from }}
                </label>
              </span>
            </div>
            <el-dropdown placement="bottom">
              <span class="el-dropdown-link">
                <i class="el-icon-more" @click.stop />
              </span>
              <el-dropdown-menu slot="dropdown" style="margin-top: -10px">
                <el-dropdown-item @click.native="handleEdit(item)">
                  {{ $t('common.button.edit') }}
                </el-dropdown-item>
                <el-dropdown-item @click.native="handleDelete(item)">
                  {{ $t('common.button.delete') }}
                </el-dropdown-item>
                <el-dropdown-item @click.native="handleConvertToSkill(item)">
                  {{ $t('common.button.convertToSkill') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
          <div class="card-des">{{ item.desc }}</div>
        </div>
      </div>

      <!--<div class="no-list" v-if="list.length === 0 && is">
        <div>
          <i class="el-icon-circle-plus-outline" @click="handleAddMCP"></i>
          <span>添加你的第一个MCP Server</span>
        </div>
      </div>-->
      <el-empty
        class="noData"
        v-if="!(list && list.length)"
        :description="$t('common.noData')"
      ></el-empty>
    </div>
    <addDialog
      :dialogVisible="addOpen"
      :title="addTitle"
      :initialData="dialogParams"
      @handleFetch="fetchList()"
      @handleClose="handleClose"
    ></addDialog>
    <ConvertSkillDialog ref="convertSkillDialog" />
  </div>
</template>
<script>
import McpOAuthPanel from '@/components/McpOAuthPanel.vue';
import addDialog from './addDialog.vue';
import SearchInput from '@/components/searchInput.vue';
import ConvertSkillDialog from '@/components/skills/convertSkillDialog.vue';
import { getList, setDelete } from '@/api/mcp';
import LinkIcon from '@/components/linkIcon.vue';
import { avatarSrc } from '@/utils/util';
export default {
  name: 'McpIntegrate',
  components: { McpOAuthPanel, LinkIcon, SearchInput, addDialog, ConvertSkillDialog },
  data() {
    return {
      defaultAvatar: require('@/assets/imgs/mcp_active.svg'),
      addOpen: false, // 自定义添加mcp开关
      addTitle: '',
      dialogParams: {
        name: '',
        from: '',
        sseUrl: '',
        streamableUrl: '',
        transport: 'sse',
        desc: '',
        apiAuth: {
          authType: 'none',
          apiKeyValue: '',
          apiKeyHeader: '',
          apiKeyHeaderPrefix: 'basic',
          apiKeyQueryParam: '',
        },
        customParams: [],
      }, // 添加自定义mcp参数
      list: [],
    };
  },
  mounted() {
    this.fetchList();
  },
  methods: {
    avatarSrc,
    fetchList() {
      const searchInput = this.$refs.searchInput;
      const params = {
        name: searchInput.value,
      };
      getList(params).then(res => {
        this.list = res.data.list || [];
      });
    },
    handleClick(val) {
      // mcpSquareId 有值 MCP广场, 否则自定义
      this.$router.push({
        path: `/mcpService/detail/custom?mcpId=${val.mcpId}&mcpSquareId=${val.mcpSquareId}`,
      });
    },
    handleAddMCP() {
      this.addOpen = true;
      this.addTitle = this.$t('tool.integrate.addTitle');
    },
    formatCustomParams(headers = {}) {
      return Object.entries(headers).map(([name, value]) => ({ name, value }));
    },
    handleEdit(item) {
      if (item.apiAuth && item.apiAuth.authType === 'oauth') { this.$message.info('请在上方 OAuth 授权区域管理此连接'); return; }
      this.addOpen = true;
      this.addTitle = this.$t('tool.integrate.editTitle');
      this.dialogParams = {
        ...item,
        customParams: this.formatCustomParams(item.headers || {}),
      };
    },
    handleClose() {
      this.addOpen = false;
      this.dialogParams = {
        name: '',
        from: '',
        sseUrl: '',
        streamableUrl: '',
        transport: 'sse',
        desc: '',
        apiAuth: {
          authType: 'none',
          apiKeyValue: '',
          apiKeyHeader: '',
          apiKeyHeaderPrefix: 'basic',
          apiKeyQueryParam: '',
        },
        customParams: [],
      };
    },
    handleDelete(item) {
      this.$confirm(
        item.apiAuth && item.apiAuth.authType === 'oauth' ? '删除此连接也会解绑其 OAuth 授权，并停用共享该授权的其他连接。是否继续？' : this.$t('tool.integrate.deleteHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          dangerouslyUseHTMLString: true,
          type: 'warning',
          center: true,
        },
      ).then(async () => {
        setDelete({
          mcpId: item.mcpId,
        }).then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delete'));
            this.fetchList();
          } else {
            this.$message.error(res.msg || this.$t('common.info.deleteErr'));
          }
        });
      });
    },
    // 转化为 Skill
    handleConvertToSkill(item) {
      if (item.apiAuth && item.apiAuth.authType === 'oauth') { this.$message.info('OAuth 连接请直接作为 MCP 使用，不导出授权凭据到 Skill'); return; }
      this.$refs.convertSkillDialog.open({
        id: item.mcpId,
        type: 'mcp',
      });
    },
  },
};
</script>
<style lang="scss">
.card-search-cust {
  text-align: left !important;

  .radio-box {
    margin: 20px 0 0 0 !important;
  }
}
.mcp-content-box .noData {
  width: 100%;
  text-align: center;
  margin-top: -60px;
  ::v-deep .el-empty__description p {
    color: #b3b1bc;
  }
}
</style>
