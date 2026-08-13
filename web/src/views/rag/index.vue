<template>
  <CommonLayout
    :isButton="true"
    :asideWidth="'260px'"
    :aside-title="$t('app.createConversation')"
    :show-title="false"
    class="right-page-content-body"
    @handleBtnClick="handleBtnClick"
    @aside-scroll="handleHistoryScroll"
  >
    <template #aside-header>
      <div class="history-toolbar">
        <el-tooltip
          :content="$t('app.createConversation')"
          placement="bottom"
          effect="dark"
        >
          <el-button
            circle
            class="create-conversation-btn"
            type="primary"
            size="small"
            @click="handleBtnClick"
          >
            <svg-icon class-name="create-icon" icon-class="message-plus" />
          </el-button>
        </el-tooltip>
        <el-input
          v-model="searchText"
          class="history-search-input"
          clearable
          size="small"
          :placeholder="$t('square.search') + '...'"
          @keyup.enter.native="handleHistorySearch"
          @clear="handleHistorySearch"
        >
          <i slot="prefix" class="el-input__icon el-icon-search"></i>
        </el-input>
      </div>
    </template>
    <template #aside-content>
      <div class="explore-aside-app">
        <div
          v-for="(item, index) in historyList"
          :key="item.conversationId || `history-${index}`"
          :class="['appList', { active: item.active }]"
          @click="historyClick(item)"
          @mouseenter="item.hover = true"
          @mouseleave="item.hover = false"
        >
          <span class="appName">
            <span class="appTag"></span>
            {{ convertTitle(item.title) }}
          </span>
          <span
            v-if="item.hover || item.active"
            class="el-icon-delete appDelete"
            @click.stop="deleteConversation(item)"
          ></span>
        </div>
        <div
          v-if="historyPageConf.loading && historyList.length"
          class="history-loading"
        >
          <i class="el-icon-loading"></i>
        </div>
      </div>
    </template>
    <template #main-content>
      <div class="app-content">
        <Chat
          ref="ragChat"
          :active-conversation-id="activeConversationId"
          :editForm="editForm"
          :chatType="'chat'"
          :maxPicNum="currentMaxPicNum"
          :maxImageSize="currentMaxImageSize"
          @conversation-created="handleConversationCreated"
          @conversation-changed="handleConversationChanged"
        />
      </div>
    </template>
  </CommonLayout>
</template>
<script>
import CommonLayout from '@/components/exploreContainer.vue';
import Chat from './components/chat.vue';
import {
  deleteRagConversation,
  getRagConversationDetail,
  getRagConversationList,
  getRagPublishedInfo,
} from '@/api/rag';
import { selectModelList } from '@/api/modelAccess';
export default {
  name: 'ExploreRag',
  components: { CommonLayout, Chat },
  data() {
    return {
      editForm: {
        appId: '',
        avatar: {},
        name: '',
        desc: '',
        modelParams: '',
        visionsupport: '',
        modelConfig: {},
        visionConfig: {
          picNum: 0,
        },
        knowledgeBaseConfig: { config: {}, knowledgebases: [] },
        qaKnowledgeBaseConfig: { config: {}, knowledgebases: [] },
        recommendQuestion: [],
      },
      modelOptions: [],
      searchText: '',
      activeConversationId: '',
      historyList: [],
      historyPageConf: {
        pageNo: 1,
        pageSize: 50,
        total: 0,
        hasMore: true,
        loading: false,
      },
    };
  },
  computed: {
    currentModelId() {
      return (
        this.editForm.modelConfig?.modelId || this.editForm.modelParams || ''
      );
    },
    currentModelInfo() {
      return (
        this.modelOptions.find(item => item.modelId === this.currentModelId) ||
        this.editForm.modelConfig ||
        {}
      );
    },
    currentModelFullConfig() {
      return (
        this.currentModelInfo.fullConfig || this.currentModelInfo.config || {}
      );
    },
    currentMaxPicNum() {
      const visionSupport = this.currentModelFullConfig.visionSupport;
      if (visionSupport === 'support') return 3;
      if (visionSupport) return -1;
      return 1;
    },
    currentMaxImageSize() {
      const size = Number(this.currentModelFullConfig.maxImageSize);
      return size > 0 ? size : null;
    },
  },
  created() {
    if (this.$route.query.id) {
      this.editForm.appId = this.$route.query.id;
      this.getDetail();
      this.getHistoryList();
    }
  },

  methods: {
    /** 查询发布态 RAG 的历史会话，支持搜索和触底分页。 */
    async getHistoryList({
      loadMore = false,
      activateConversationId = '',
    } = {}) {
      if (!this.editForm.appId || this.historyPageConf.loading) return;
      if (loadMore && !this.historyPageConf.hasMore) return;

      const pageNo = loadMore ? this.historyPageConf.pageNo + 1 : 1;
      this.historyPageConf.loading = true;
      try {
        const res = await getRagConversationList({
          ragId: this.editForm.appId,
          pageNo,
          pageSize: this.historyPageConf.pageSize,
          searchText: this.searchText.trim(),
        });
        if (res.code !== 0) {
          if (!loadMore) this.resetHistoryList();
          return;
        }

        const activeConversationId =
          activateConversationId || this.activeConversationId;
        const list = (res.data?.list || []).map(item => ({
          ...item,
          active: item.conversationId === activeConversationId,
          hover: false,
        }));
        this.historyPageConf.total = Number(res.data?.total) || 0;
        this.historyPageConf.pageNo = pageNo;
        const currentCount = loadMore ? this.historyList.length : 0;
        this.historyPageConf.hasMore =
          currentCount + list.length < this.historyPageConf.total;

        if (loadMore) {
          const historyMap = new Map(
            this.historyList.map(item => [item.conversationId, item]),
          );
          list.forEach(item => {
            const existing = historyMap.get(item.conversationId);
            if (existing) {
              Object.assign(existing, item, {
                active: existing.active,
                hover: existing.hover,
              });
            } else {
              this.historyList.push(item);
            }
          });
        } else {
          this.historyList = list;
        }
      } catch (error) {
        console.warn('[rag chat] get conversation list failed', error);
        if (!loadMore) this.resetHistoryList();
      } finally {
        this.historyPageConf.loading = false;
      }
    },
    resetHistoryList() {
      this.historyList = [];
      this.historyPageConf.pageNo = 1;
      this.historyPageConf.total = 0;
      this.historyPageConf.hasMore = false;
    },
    handleHistorySearch() {
      this.searchText = this.searchText.trim();
      this.getHistoryList();
    },
    handleHistoryScroll(event) {
      const el = event && event.target;
      if (!el || this.historyPageConf.loading || !this.historyPageConf.hasMore)
        return;
      if (el.scrollHeight - el.scrollTop - el.clientHeight < 120) {
        this.getHistoryList({ loadMore: true });
      }
    },
    async historyClick(item) {
      this.$refs.ragChat?.disconnectCurrentStreamForNavigation();
      this.activeConversationId = item.conversationId;
      this.historyList.forEach(history => {
        history.active = history.conversationId === item.conversationId;
      });
      try {
        const res = await getRagConversationDetail({
          ragId: this.editForm.appId,
          conversationId: item.conversationId,
          pageNo: 1,
          pageSize: 1000,
        });
        if (res.code !== 0) {
          this.$message.error(res.msg || this.$t('sse.error'));
          return;
        }
        // 用户快速切换会话时，忽略已过期请求的详情响应。
        if (this.activeConversationId !== item.conversationId) return;
        this.$refs.ragChat?.loadConversationHistory(res.data?.list || []);
      } catch (error) {
        console.warn('[rag chat] get conversation detail failed', error);
      }
    },
    handleBtnClick() {
      this.activeConversationId = '';
      this.$refs.ragChat && this.$refs.ragChat.createConversation();
    },
    async handleConversationCreated(conversationId) {
      this.activeConversationId = conversationId;
      await this.getHistoryList({ activateConversationId: conversationId });
    },
    handleConversationChanged(conversationId) {
      this.activeConversationId = conversationId;
      this.historyList.forEach(history => {
        history.active = history.conversationId === conversationId;
      });
    },
    async deleteConversation(item) {
      if (!item?.conversationId || this.$refs.ragChat?.sessionStatus === 0)
        return;
      try {
        const res = await deleteRagConversation({
          ragId: this.editForm.appId,
          conversationId: item.conversationId,
        });
        if (res.code !== 0) {
          this.$message.error(res.msg || this.$t('sse.error'));
          return;
        }
        const isActive = item.conversationId === this.activeConversationId;
        this.historyList = this.historyList.filter(
          history => history.conversationId !== item.conversationId,
        );
        this.historyPageConf.total = Math.max(
          this.historyPageConf.total - 1,
          this.historyList.length,
        );
        if (isActive) {
          this.activeConversationId = '';
          this.$refs.ragChat?.createConversation();
        }
        await this.getHistoryList({
          activateConversationId: this.activeConversationId,
        });
      } catch (error) {
        console.warn('[rag chat] delete conversation failed', error);
        this.$message.error(this.$t('sse.error'));
      }
    },
    convertTitle(title) {
      const value = (title || '').trim();
      return value || this.$t('app.defaultConversationTitle');
    },
    async getDetail() {
      const res = await getRagPublishedInfo({ ragId: this.editForm.appId });
      if (res.code === 0) {
        this.editForm.avatar = res.data.avatar;
        this.editForm.name = res.data.name;
        this.editForm.desc = res.data.desc;
        this.editForm.visionConfig = res.data.visionConfig || { picNum: 0 };
        this.$set(this.editForm, 'modelConfig', res.data.modelConfig || {});
        this.editForm.modelParams = res.data.modelConfig?.modelId || '';
        if (res.data.knowledgeBaseConfig) {
          this.editForm.knowledgeBaseConfig = res.data.knowledgeBaseConfig;
        }
        if (res.data.qaKnowledgeBaseConfig) {
          this.editForm.qaKnowledgeBaseConfig = res.data.qaKnowledgeBaseConfig;
        }
        this.editForm.recommendQuestion = res.data.recommendQuestion?.map(
          item => ({
            value: item,
          }),
        );
        await this.getModelData();
      }
    },
    async getModelData() {
      try {
        const res = await selectModelList();
        if (res.code === 0) {
          this.modelOptions = res.data.list || [];
          this.applyModelFullConfig();
        }
      } catch (error) {
        console.warn('[rag chat] get model list failed', error);
      }
    },
    applyModelFullConfig() {
      const modelId = this.currentModelId;
      if (!modelId) return;
      const selectedModel = this.modelOptions.find(
        item => item.modelId === modelId,
      );
      if (!selectedModel) return;
      this.editForm.visionsupport = selectedModel.config?.visionSupport || '';
      this.$set(this.editForm, 'modelConfig', {
        ...(this.editForm.modelConfig || {}),
        fullConfig: selectedModel.config || {},
      });
    },
    goBack() {
      this.$router.go(-1);
    },
  },
};
</script>
<style lang="scss" scoped>
::v-deep {
  .apikeyBtn {
    padding: 11px 10px;
    border: 1px solid $btn_bg;
    color: $btn_bg;
    display: flex;
    align-items: center;
    img {
      height: 14px;
    }
  }
}
.app-content {
  width: 100%;
  height: 100%;
  position: relative;
  .app-header-api {
    width: 100%;
    padding: 10px;
    position: absolute;
    z-index: 999;
    top: 0;
    left: 0;
    border-bottom: 1px solid #eaeaea;
    display: flex;
    justify-content: space-between;
    align-content: center;
    .app_name {
      font-size: 18px;
      font-weight: bold;
      color: $color_title;
      display: flex;
      align-items: center;
      .goBack {
        font-weight: bold;
        font-size: 16px;
        cursor: pointer;
        margin-right: 15px;
        color: #333;
      }
    }
    .header-api-box {
      display: flex;
      .header-api-url {
        padding: 6px 10px;
        background: #fff;
        margin: 0 10px;
        border-radius: 6px;
        .root-url {
          background-color: #eceefe;
          color: $color;
          border: none;
        }
      }
    }
  }
}
.history-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  .create-conversation-btn {
    flex: none;
  }
  .history-search-input {
    min-width: 0;
  }
}
.explore-aside-app {
  .history-loading {
    padding: 10px 0;
    text-align: center;
    color: $color;
  }
  .appList {
    display: flex;
    align-items: center;
    margin: 10px 20px 6px;
    padding: 10px;
    border-radius: 6px;
    cursor: pointer;
    justify-content: space-between;
    &:hover,
    &.active {
      background-color: $color_opacity;
    }
    &.active .appTag {
      background-color: $color;
    }
    .appDelete {
      flex: none;
      margin-left: 8px;
      color: $color;
      cursor: pointer;
    }
  }
  .appName {
    display: block;
    max-width: 100%;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    .appTag {
      display: inline-block;
      width: 8px;
      height: 8px;
      margin-right: 8px;
      border-radius: 50%;
      background: #ccc;
    }
  }
}
</style>
