<template>
  <div class="agent-mobile-wrapper right-page-content-body">
    <!-- 移动端菜单按钮 -->
    <div class="mobile-menu-btn" @click="toggleMobileMenu" v-if="!showAside">
      <img src="@/assets/imgs/historyList.png" class="mobile-menu-img" />
    </div>
    <!-- 移动端遮罩层 -->
    <div
      class="mobile-overlay"
      :class="{ show: showMobileMenu && isMobile }"
      @click="closeMobileMenu"
      v-if="isMobile"
    ></div>
    <CommonLayout
      :aside-title="asideTitle"
      :isButton="true"
      :asideWidth="asideWidth"
      @handleBtnClick="handleBtnClick"
      :class="[chatType === 'webChat' ? 'chatBg' : '']"
      :showAside="showAside"
      @aside-scroll="handleHistoryScroll"
    >
      <template #aside-content>
        <transition name="fade">
          <div class="explore-aside-app">
            <div
              v-for="(n, i) in historyList"
              class="appList"
              :class="['appList', { active: n.active }]"
              @click="historyClick(n)"
              @touchstart="historyClick(n)"
              @mouseenter="mouseEnter(n)"
              @mouseleave="mouseLeave(n)"
              :key="n.conversationId || 'history' + i"
            >
              <span class="appName">
                <span class="appTag"></span>
                {{ n.title }}
              </span>
              <span
                class="el-icon-delete appDelete"
                v-if="n.hover || n.active"
                @click.stop="deleteConversation(n)"
              ></span>
            </div>
            <div
              v-if="historyPageConf.loading && historyList.length"
              class="history-loading"
            >
              <i class="el-icon-loading"></i>
            </div>
          </div>
        </transition>
      </template>
      <template #main-content>
        <div class="app-content">
          <Chat
            :chatType="'chat'"
            :editForm="editForm"
            :assistantId="assistantId"
            :appUrlInfo="appUrlInfo"
            :type="chatType"
            :maxPicNum="currentMaxPicNum"
            :maxImageSize="currentMaxImageSize"
            :maxFileNum="10"
            :maxFileSize="100"
            ref="agentChat"
            @reloadList="reloadList"
            @conversationDeleted="handleConversationDeleted"
            @conversationPromoted="handleConversationPromoted"
            @setHistoryStatus="setHistoryStatus"
          />
        </div>
      </template>
    </CommonLayout>
  </div>
</template>
<script>
import CommonLayout from '@/components/exploreContainer.vue';
import Chat from './components/chat.vue';
import { mapGetters } from 'vuex';
import {
  getAgentPublishedInfo,
  getOpenurlInfo,
  getOpenurlAgentLlm,
  OpenurlConverList,
  getConversationlist,
} from '@/api/agent';
import { getApiKeyRoot } from '@/api/appspace';
import { selectModelList } from '@/api/modelAccess';
import sseMethod from '@/mixins/sseMethod';
import { MULTIPLE_AGENT, SINGLE_AGENT } from '@/views/agent/constants';
import { guid, getXClientId } from '@/utils/util';
export default {
  name: 'ExploreAgent',
  components: { CommonLayout, Chat },
  mixins: [sseMethod],
  provide() {
    return {
      getHeaderConfig: this.headerConfig,
    };
  },
  data() {
    return {
      showAside: false,
      asideWidth: '260px',
      apiURL: '',
      asideTitle: this.$t('app.createConversation'),
      assistantId: '',
      historyList: [],
      historyPageConf: {
        pageNo: 1,
        pageSize: 50,
        total: 0,
        hasMore: true,
        loading: false,
      },
      appUrlInfo: {},
      modelOptions: [],
      editForm: {
        assistantId: '',
        category: SINGLE_AGENT,
        avatar: {},
        name: '',
        desc: '',
        prologue: '',
        recommendQuestion: [],
        modelConfig: {},
        recommendConfig: {
          recommendEnable: false,
          modelConfig: {
            config: {
              temperature: 0.7,
              temperatureEnable: true,
              topP: 1,
              topPEnable: true,
              frequencyPenalty: 0,
              frequencyPenaltyEnable: true,
              presencePenalty: 0,
              presencePenaltyEnable: true,
              maxTokens: 512,
              maxTokensEnable: true,
            },
            displayName: '',
            model: '',
            modelId: '',
            modelType: '',
            provider: '',
          },
          promptEnable: false,
          prompt: '',
          maxHistory: 3,
        },
      },
      chatType: 'agentChat',
      apiStrategies: {
        agentChat_info: getAgentPublishedInfo,
        webChat_info: getOpenurlInfo,
        agentChat_converstionList: getConversationlist,
        webChat_converstionList: OpenurlConverList,
      },
      xClientId: '',
      isMobile: false,
      showMobileMenu: false,
    };
  },
  computed: {
    ...mapGetters('app', ['sessionStatus']),
    currentModelId() {
      return this.editForm.modelConfig?.modelId || '';
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
    const id = this.$route.query.id || this.$route.params.id;
    if (id) {
      this.assistantId = id;
      this.editForm.assistantId = id;
    }
    if (this.$route.path.includes('/webChat')) {
      this.chatType = 'webChat';
      this.initXClientId();
    } else {
      this.chatType = 'agentChat';
    }
    this.getDetail();
    this.getList();
  },
  mounted() {
    //检查是否是移动端
    this.checkMobile();
    window.addEventListener('resize', this.checkMobile);
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkMobile);
  },
  methods: {
    checkMobile() {
      this.isMobile = window.innerWidth < 768;
      if (this.isMobile) {
        this.showMobileMenu = false;
        this.showAside = false;
      } else {
        this.showAside = true;
      }
      this.$nextTick(() => {
        this.ensureHistoryScrollable();
      });
    },
    toggleMobileMenu() {
      this.showMobileMenu = true;
      this.showAside = true;
      this.$nextTick(() => {
        this.ensureHistoryScrollable();
      });
    },
    closeMobileMenu() {
      this.showMobileMenu = false;
      this.showAside = false;
    },
    initXClientId() {
      let xClientId = getXClientId();
      if (!xClientId) {
        xClientId = guid();
        localStorage.setItem('xClientId', xClientId);
      }
      this.xClientId = xClientId;
      return xClientId;
    },
    // 重置会话列表，创建新会话后可激活第一条
    reloadList(activateFirst) {
      this.getList({ reset: true, activateFirst });
    },
    // 按当前会话类型请求指定页的历史会话
    fetchHistoryPage(pageNo) {
      const params = {
        pageNo,
        pageSize: this.historyPageConf.pageSize,
      };
      if (this.chatType === 'agentChat') {
        return getConversationlist({
          assistantId: this.assistantId,
          ...params,
        });
      }
      const config = this.headerConfig();
      return OpenurlConverList(this.assistantId, params, config);
    },
    // 给历史会话补齐侧栏交互状态
    normalizeHistoryItem(item) {
      return { ...item, hover: false, active: false };
    },
    // 追加分页数据并按 conversationId 去重
    mergeHistoryList(list) {
      const existedMap = new Map(
        this.historyList.map(item => [item.conversationId, item]),
      );
      list.forEach(item => {
        const oldItem = existedMap.get(item.conversationId);
        if (oldItem) {
          Object.assign(oldItem, {
            ...item,
            hover: oldItem.hover,
            active: oldItem.active,
          });
        } else {
          this.historyList.push(item);
        }
      });
    },
    // 根据 total 或本页数量判断是否还有更多
    updateHistoryHasMore(rawList) {
      if (this.historyPageConf.total > 0) {
        this.historyPageConf.hasMore =
          this.historyList.length < this.historyPageConf.total;
        return;
      }
      if (
        this.chatType === 'webChat' &&
        rawList.length > this.historyPageConf.pageSize
      ) {
        this.historyPageConf.hasMore = false;
        return;
      }
      this.historyPageConf.hasMore =
        rawList.length >= this.historyPageConf.pageSize;
    },
    headerConfig() {
      const config = {
        headers: { 'X-Client-ID': this.initXClientId() },
      };
      return config;
    },
    async getModelData() {
      if (this.chatType === 'webChat') {
        return this.getOpenurlModelData();
      }
      try {
        const res = await selectModelList();
        if (res.code === 0) {
          this.modelOptions = res.data.list || [];
          this.applyModelFullConfig();
        }
      } catch (error) {
        console.warn('[agent chat] get model list failed', error);
      }
    },
    async getOpenurlModelData() {
      try {
        const config = this.headerConfig();
        const res = await getOpenurlAgentLlm(this.assistantId, config);
        if (res.code === 0) {
          const modelConfig = res.data?.modelConfig || res.data || {};
          this.modelOptions = Object.keys(modelConfig).length
            ? [modelConfig]
            : [];
          this.$set(this.editForm, 'modelConfig', {
            ...(this.editForm.modelConfig || {}),
            ...modelConfig,
            fullConfig: modelConfig.config || {},
          });
        }
        return res;
      } catch (error) {
        console.warn('[agent webChat] get openurl llm config failed', error);
        return null;
      }
    },
    applyModelFullConfig() {
      const modelId = this.currentModelId;
      if (!modelId) return;
      const selectedModel = this.modelOptions.find(
        item => item.modelId === modelId,
      );
      if (!selectedModel) return;
      this.$set(this.editForm, 'modelConfig', {
        ...(this.editForm.modelConfig || {}),
        fullConfig: selectedModel.config || {},
      });
    },
    async getDetail() {
      let res = null;
      let data = null;
      if (this.chatType === 'agentChat') {
        res = await getAgentPublishedInfo({
          assistantId: this.editForm.assistantId,
        });
      } else {
        const config = this.headerConfig();
        res = await getOpenurlInfo(this.assistantId, config);
      }
      if (res.code === 0) {
        if (this.chatType === 'agentChat') {
          data = res.data;
          this.editForm.category = data.category;
        } else {
          data = res.data.assistant;
          this.appUrlInfo = data.appUrlInfo;
        }
        this.editForm.avatar = data.avatar;
        this.editForm.name = data.name;
        this.editForm.desc = data.desc;
        this.editForm.prologue = data.prologue;
        this.editForm.recommendQuestion = data.recommendQuestion.map(item => ({
          value: item,
        }));
        this.editForm.recommendConfig = data.recommendConfig;
        this.$set(this.editForm, 'modelConfig', data.modelConfig || {});
        await this.getModelData();
      }
    },
    // 获取历史会话列表，支持首次加载和滚动加载
    async getList(options = {}) {
      const opts =
        typeof options === 'boolean' ? { activateFirst: options } : options;
      const loadMore = !!opts.loadMore;
      const activateFirst = !!opts.activateFirst;
      if (this.historyPageConf.loading) return;
      if (loadMore && !this.historyPageConf.hasMore) return;

      const pageNo = loadMore ? this.historyPageConf.pageNo + 1 : 1;
      this.historyPageConf.loading = true;
      try {
        const res = await this.fetchHistoryPage(pageNo);
        if (res.code === 0) {
          const rawList = res.data?.list || [];
          this.historyPageConf.total = Number(res.data?.total) || 0;
          const list = rawList.map(item => this.normalizeHistoryItem(item));

          if (loadMore) {
            this.mergeHistoryList(list);
          } else {
            this.historyList = list;
          }

          this.historyPageConf.pageNo = pageNo;
          if (activateFirst && this.historyList[0]) {
            this.historyList[0].active = true;
          }
          this.updateHistoryHasMore(rawList);
        } else if (!loadMore) {
          this.historyList = [];
          this.historyPageConf.hasMore = false;
          this.historyPageConf.total = 0;
        }
      } catch (error) {
        console.warn('[agent chat] get conversation list failed', error);
        if (!loadMore) {
          this.historyList = [];
          this.historyPageConf.hasMore = false;
          this.historyPageConf.total = 0;
        }
      } finally {
        this.historyPageConf.loading = false;
      }
      if (!loadMore && !opts.skipEnsure) {
        this.$nextTick(() => {
          this.ensureHistoryScrollable();
        });
      }
    },
    // 获取侧栏滚动容器，用于判断首屏是否已出现滚动条
    getHistoryScrollEl() {
      return this.$el && this.$el.querySelector('.aside-content');
    },
    // 首屏列表不足一屏时自动补拉，避免无滚动条无法触底加载
    async ensureHistoryScrollable() {
      const el = this.getHistoryScrollEl();
      if (!el) return;

      let loadCount = 0;
      const maxAutoLoadCount = 5;
      while (
        this.historyPageConf.hasMore &&
        !this.historyPageConf.loading &&
        el.scrollHeight <= el.clientHeight &&
        loadCount < maxAutoLoadCount
      ) {
        loadCount += 1;
        await this.getList({ loadMore: true, skipEnsure: true });
        await this.$nextTick();
      }
    },
    // 侧栏滚动接近底部时加载下一页
    handleHistoryScroll(event) {
      const el = event && event.target;
      if (!el || this.historyPageConf.loading || !this.historyPageConf.hasMore)
        return;
      const distanceToBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
      if (distanceToBottom < 120) {
        this.getList({ loadMore: true });
      }
    },
    // 已有会话再次提问后置顶，保持侧栏顺序与最近对话一致
    handleConversationPromoted(item) {
      const conversationId = item && item.conversationId;
      const index = this.historyList.findIndex(
        history => history.conversationId === conversationId,
      );
      if (index <= 0) return;
      const [current] = this.historyList.splice(index, 1);
      this.historyList.unshift(current);
      this.$nextTick(() => {
        const el = this.getHistoryScrollEl();
        if (el) {
          el.scrollTo({ top: 0, behavior: 'smooth' });
        }
      });
    },
    // 删除成功后本地移除并补位，保持已加载列表连续
    async handleConversationDeleted(item) {
      const conversationId = item && item.conversationId;
      if (!conversationId) {
        this.getList({ reset: true });
        return;
      }

      this.historyList = this.historyList.filter(
        history => history.conversationId !== conversationId,
      );
      if (this.historyPageConf.total > 0) {
        this.historyPageConf.total = Math.max(
          this.historyPageConf.total - 1,
          this.historyList.length,
        );
      }

      if (
        this.historyPageConf.hasMore ||
        (this.historyPageConf.total > 0 &&
          this.historyList.length < this.historyPageConf.total)
      ) {
        await this.refillHistoryAfterDelete();
      }
      this.$nextTick(() => {
        this.ensureHistoryScrollable();
      });
    },
    // 删除后重新拉当前最后页，补齐 offset 分页产生的缺口
    async refillHistoryAfterDelete() {
      if (this.historyPageConf.loading) return;
      this.historyPageConf.loading = true;
      try {
        const res = await this.fetchHistoryPage(
          Math.max(this.historyPageConf.pageNo, 1),
        );
        if (res.code === 0) {
          const rawList = res.data?.list || [];
          this.historyPageConf.total =
            Number(res.data?.total) || this.historyPageConf.total;
          const list = rawList.map(item => this.normalizeHistoryItem(item));
          this.mergeHistoryList(list);
          this.updateHistoryHasMore(rawList);
        }
      } catch (error) {
        console.warn('[agent chat] refill conversation list failed', error);
      } finally {
        this.historyPageConf.loading = false;
      }
    },
    setHistoryStatus() {
      this.historyList.forEach(m => {
        m.active = false;
      });
    },
    historyClick(n) {
      //切换对话
      n.hover = true;
      this.$refs['agentChat'].conversationClick(n);
      if (this.isMobile) {
        this.showMobileMenu = false;
        this.showAside = false;
      }
    },
    deleteConversation(n) {
      this.$refs['agentChat'].preDelConversation(n);
    },
    handleBtnClick() {
      //新建对话
      this.$refs['agentChat'].createConversion();
      if (this.isMobile) {
        this.showMobileMenu = false;
        this.showAside = false;
      }
    },
    mouseEnter(n) {
      n.hover = true;
    },
    mouseLeave(n) {
      n.hover = false;
    },
    apiKeyRootUrl() {
      const data = { appId: this.editForm.assistantId, appType: 'agent' };
      getApiKeyRoot(data).then(res => {
        if (res.code === 0) {
          this.apiURL = res.data || '';
        }
      });
    },
    openApiDialog() {
      this.$refs.apiKeyDialog.showDialog();
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/chat.scss';
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter,
.fade-leave-to {
  opacity: 0;
}
.chatBg {
  background: linear-gradient(
    1deg,
    rgb(255, 255, 255) 42%,
    rgb(255, 255, 255) 42%,
    rgb(235, 237, 254) 98%,
    rgb(238, 240, 255) 98%
  );
}
.active {
  background-color: $color_opacity !important;
  .appTag {
    background-color: $color !important;
  }
}
.agent-mobile-wrapper {
  width: 100%;
  height: 100%;
  position: relative;
  .mobile-menu-btn {
    display: none;
    position: fixed;
    top: 5px;
    z-index: 1001;
    border-radius: 4px;
    padding: 5px 12px;
    cursor: pointer;
    .mobile-menu-img {
      width: 24px;
    }
  }
  .mobile-overlay {
    display: none;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 999;
    transition: opacity 0.3s ease;

    &.show {
      display: block;
    }
  }
}
.explore-aside-app {
  .history-loading {
    padding: 10px 0;
    text-align: center;
    color: $color;
  }
  .appList:hover {
    background-color: $color_opacity !important;
  }
  .appList {
    margin: 10px 20px;
    padding: 10px;
    border-radius: 6px;
    margin-bottom: 6px;
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
    position: relative;
    .appDelete {
      color: $color;
      margin-right: -5px;
      cursor: pointer;
    }
    .appName {
      display: block;
      max-width: 130px;
      overflow: hidden;
      white-space: nowrap;
      pointer-events: none;
      text-overflow: ellipsis;
      .appTag {
        display: inline-block;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #ccc;
      }
    }
  }
}
.app-content {
  width: 100%;
  height: 100%;
}

// weburl适配移动端样式
::v-deep .chatBg,
::v-deep .explore-container {
  @media (max-width: 768px) {
    .el-aside {
      position: fixed !important;
      top: 0 !important;
      left: 0 !important;
      height: 100vh !important;
      z-index: 1000 !important;
      transition: transform 0.3s ease !important;
      border-radius: 0 !important;
      box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15) !important;
      width: 60vw !important;
      .mobile-menu-open & {
        transform: translateX(0) !important;
      }
    }

    .el-main {
      width: 99% !important;
      padding-top: 16px;
      margin-left: 0 !important;
      .center-editable {
        left: 0;
        right: 0;
      }
      .center-session .history-box {
        padding: 0;
      }
      .session-answer .session-answer-wrapper {
        padding-left: 0;
      }
      .session .session-item {
        padding-right: 0;
      }
      .edtable--wrap {
        z-index: 99;
        .editable--send {
          padding: 5px 12px;
          span img {
            width: 12px;
            height: 12px;
          }
        }
      }
    }
    &.el-container {
      width: 100% !important;
    }
  }
}

@media (max-width: 768px) {
  .agent-mobile-wrapper .mobile-menu-btn {
    display: block;
  }
}
</style>
