<template>
  <div class="general-agent-page">
    <!-- 主内容区 -->
    <div
      :class="{
        'has-sidebar': !sidebarCollapsed,
        'has-workspace': panelVisible && activeWorkspace,
        'has-preview': previewVisible,
      }"
      class="agent-main-content"
    >
      <!-- 左侧会话列表 - 固定宽度，可折叠 -->
      <div :class="['sidebar', { collapsed: sidebarCollapsed }]">
        <div class="sidebar-header">
          <div class="history-toolbar">
            <el-tooltip
              :content="$t('generalAgent.sidebar.newChat')"
              effect="dark"
              placement="bottom"
            >
              <el-button
                circle
                class="create-conversation-btn"
                size="small"
                type="primary"
                @click="initNewConversation"
              >
                <svg-icon class-name="create-icon" icon-class="message-plus" />
              </el-button>
            </el-tooltip>
            <el-input
              v-model="searchText"
              :placeholder="$t('generalAgent.sidebar.search') + '...'"
              class="history-search-input"
              clearable
              size="small"
              @clear="handleConversationSearch"
              @keyup.enter.native="handleConversationSearch"
            >
              <i slot="prefix" class="el-input__icon el-icon-search"></i>
            </el-input>
          </div>
        </div>

        <div class="sidebar-divider"></div>

        <div
          ref="conversationList"
          class="conversation-list"
          @scroll="handleConversationListScroll"
        >
          <div
            v-for="item in conversationList"
            :key="item.conversationId"
            :class="[
              'conversation-item',
              { active: currentThreadId === item.conversationId },
            ]"
            @click="selectConversation(item.conversationId, item)"
          >
            <i class="el-icon-chat-dot-round"></i>
            <span class="conversation-title">{{ item.title }}</span>
            <i
              class="el-icon-delete conversation-delete"
              @click.stop="handleDeleteConversation(item)"
            ></i>
          </div>
          <!-- 加载更多 -->
          <div
            v-if="isLoadingMoreConversations"
            class="conversation-loading-more"
          >
            <i class="el-icon-loading"></i>
          </div>
          <div
            v-else-if="!hasMoreConversations && conversationList.length > 0"
            class="conversation-no-more"
          >
            {{ $t('generalAgent.sidebar.noMore') }}
          </div>
        </div>
      </div>

      <!-- 中间区域：主消息区域 -->
      <div class="center-panel">
        <!-- 顶部标题栏 -->
        <div class="header">
          <div class="header-left">
            <button
              :aria-label="$t('menu.back')"
              :title="$t('menu.back')"
              class="header-icon-btn"
              @click="$router.push('/explore')"
            >
              <i class="el-icon-arrow-left"></i>
            </button>

            <div class="header-btns">
              <button class="header-icon-btn" @click="toggleSidebar">
                <i
                  :class="
                    sidebarCollapsed ? 'el-icon-s-unfold' : 'el-icon-s-fold'
                  "
                ></i>
              </button>
              <button class="header-icon-btn" @click="initNewConversation">
                <i class="el-icon-plus"></i>
              </button>
            </div>

            <div class="header-title">{{ currentTitle }}</div>
          </div>
        </div>

        <!-- 消息区域 - 独立滚动 -->
        <div
          ref="messageArea"
          :class="[
            'message-area',
            { empty: isEmptyConversation && !isLoadingHistory },
          ]"
          @scroll="handleMessageAreaScroll"
        >
          <!-- 加载历史记录中 -->
          <div v-if="isLoadingHistory" class="history-loading">
            <i class="el-icon-loading"></i>
            <span>{{ $t('generalAgent.sidebar.loading') }}</span>
          </div>

          <!-- 消息列表 -->
          <div
            v-else-if="messageList.length > 0 || isStreaming"
            class="message-list"
          >
            <message-item
              v-for="(msg, index) in messageList"
              :key="msg.id || index"
              :is-last-message="index === messageList.length - 1"
              :message="msg"
              :thread-id="currentThreadId"
              @regenerate="handleRegenerate"
              @view-workspace="handleViewWorkspace"
              @download-all="handleDownloadAll"
              @question-reply="handleQuestionReply"
              @question-reject="handleQuestionReject"
            />
          </div>

          <div ref="scrollAnchor"></div>
        </div>

        <!-- 底部输入区 -->
        <div
          :class="[
            'input-area',
            { 'is-centered': isEmptyConversation && !isLoadingHistory },
          ]"
        >
          <!-- 滚动到底部按钮 -->
          <transition name="scroll-btn-fade">
            <button
              v-if="showScrollToBottom"
              class="scroll-to-bottom-btn"
              @click="handleScrollToBottomClick"
            >
              <svg
                fill="none"
                height="16"
                stroke="currentColor"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                viewBox="0 0 24 24"
                width="16"
              >
                <polyline points="6,9 12,15 18,9"></polyline>
              </svg>
            </button>
          </transition>
          <!-- 欢迎词 - 仅居中时显示 -->
          <div
            v-if="isEmptyConversation && !isLoadingHistory"
            class="welcome-section"
          >
            <div class="welcome-avatar">
              <img
                v-if="assistantAvatar"
                :src="assistantAvatar"
                alt="Assistant"
              />
            </div>
            <div class="welcome-title">
              {{ welcomeText || $t('generalAgent.header.welcomeTitle') }}
            </div>
          </div>

          <div class="input-container">
            <!-- 模型选择 -->
            <div style="margin-bottom: 12px">
              <ModelSelect
                v-model="selectedModel"
                :filterable="true"
                :loading="modelLoading"
                :options="modelList"
                :placeholder="$t('common.model.select')"
                :style="{ width: modelSelectWidth }"
                @change="handleModelChange"
              />
            </div>

            <!-- 文件预览 -->
            <div v-if="uploadedFiles.length > 0" class="file-preview">
              <!-- 图片文件 -->
              <div
                v-for="(file, index) in uploadedFiles"
                :key="index"
                :class="['echo-img-box', { 'is-uploading': file.uploading }]"
              >
                <!-- 图片类型 -->
                <el-image
                  v-if="file.type && file.type.startsWith('image/')"
                  :preview-src-list="[file.displayUrl]"
                  :src="file.displayUrl"
                  class="echo-img"
                ></el-image>
                <!-- 文档类型 -->
                <div v-else class="echo-doc-box">
                  <img
                    :src="require('@/assets/imgs/fileicon.png')"
                    class="docIcon"
                  />
                  <div class="docInfo">
                    <p class="docInfo_name">
                      {{ $t('knowledgeManage.fileName') }}：{{ file.fileName }}
                    </p>
                    <p class="docInfo_size">
                      {{ $t('knowledgeManage.fileSize') }}：{{
                        file.size > 1024
                          ? (file.size / (1024 * 1024)).toFixed(2) + ' MB'
                          : (file.size || 0) + ' bytes'
                      }}
                    </p>
                  </div>
                </div>
                <!-- 删除按钮 -->
                <i
                  class="el-icon-close echo-close"
                  @click="removeFile(index)"
                ></i>
              </div>
            </div>

            <!-- 输入框 -->
            <div class="input-wrapper">
              <MentionInput
                ref="mentionInput"
                v-model="inputMessage"
                :before-enter-submit="beforeEnterSubmit"
                :disable-mention="true"
                :disabled="isStreaming"
                :isDIP="true"
                :placeholder="inputPlaceholder"
                @keydown-enter="handleKeyDown"
              />
            </div>

            <!-- 底部工具栏：发送按钮 -->
            <div class="input-toolbar">
              <div class="toolbar-left"></div>
              <div class="toolbar-right">
                <GAFileUpload
                  :fileTypeArr="['doc/*', 'image/*', 'audio/*']"
                  type="wga"
                  @setFileId="handleSetFileId"
                >
                  <template #default="{ openDialog }">
                    <el-tooltip
                      :content="$t('generalAgent.header.uploadFile')"
                      placement="top"
                    >
                      <i
                        class="action-icon el-icon-paperclip"
                        @click="openDialog"
                      ></i>
                    </el-tooltip>
                  </template>
                </GAFileUpload>
                <el-button
                  v-show="isStreaming"
                  circle
                  class="send-btn stop-btn"
                  @click="handleStopClick"
                >
                  <svg
                    class="stop-icon"
                    height="16"
                    viewBox="0 0 24 24"
                    width="16"
                  >
                    <rect height="12" rx="2" width="12" x="6" y="6" />
                  </svg>
                </el-button>
                <el-button
                  v-show="!isStreaming"
                  :disabled="!canSend"
                  circle
                  class="send-btn"
                  type="primary"
                  @click="sendMessage"
                >
                  <svg
                    class="send-icon"
                    height="18"
                    viewBox="0 0 24 24"
                    width="18"
                  >
                    <path
                      d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"
                      fill="currentColor"
                    />
                  </svg>
                </el-button>
              </div>
            </div>
          </div>
          <div class="input-footer">
            <span>{{ $t('generalAgent.header.footer') }}</span>
          </div>
        </div>
      </div>

      <!-- 右侧区域：工作空间 + 预览 -->
      <div class="right-area">
        <!-- Workspace 面板 -->
        <transition name="workspace-slide">
          <workspace-panel
            v-if="panelVisible && activeWorkspace"
            ref="workspacePanel"
            :initial-data="currentWorkspaceTree"
            :run-id="activeWorkspace.runId"
            :thread-id="activeWorkspace.threadId"
            @close="hidePanel"
            @preview-file="handlePreviewFile"
          />
        </transition>

        <!-- 预览面板：文件预览 -->
        <div v-if="previewVisible" class="preview-panel">
          <file-preview-drawer
            :blob="previewBlob"
            :file-name="previewFileName"
            :loading="previewLoading"
            :visible.sync="previewVisible"
            @close="previewVisible = false"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import MessageItem from '@/views/generalAgent/components/MessageItem.vue';
import WorkspacePanel from '@/views/generalAgent/components/WorkspacePanel.vue';
import FilePreviewDrawer from '@/views/generalAgent/components/FilePreviewDrawer.vue';
import MentionInput from '@/views/generalAgent/components/MentionInput.vue';
import ModelSelect from '@/components/modelSelect.vue';
import GAFileUpload from '@/views/generalAgent/components/GAFileUpload.vue';
import {
  cancelGeneralAgentConversation,
  connectGeneralAgentConversation,
  downloadGeneralAgentWorkspace,
  getGeneralAgentConversationPending,
  getGeneralAgentWorkspace,
  previewGeneralAgentWorkspace,
} from '@/api/generalAgent';
import {
  chatDigitalEmployeeConversation,
  createDigitalEmployeeConversation,
  deleteDigitalEmployeeConversation,
  getDigitalEmployeeConversationConfig,
  getDigitalEmployeeConversationDetail,
  getDigitalEmployeeConversationList,
  getDigitalEmployeeDetail,
  updateDigitalEmployeeConversationConfig,
} from '@/api/digitalEmployee';
import { selectModelList } from '@/api/modelAccess';
import { avatarSrc, resDownloadFile } from '@/utils/util';
import { SSEEventParser } from '@/views/generalAgent/utils/sse-parser';
import {
  aggregateEventsToMessages,
  aggregateInputMessagesToUserMessages,
} from '@/views/generalAgent/utils/message-aggregator';
// 引入 Mixins
import streamStateManager from '@/views/generalAgent/mixins/streamStateManager';
import messageManager from '@/views/generalAgent/mixins/messageManager';
import fileManager from '@/views/generalAgent/mixins/fileManager';
import scrollController from '@/views/generalAgent/mixins/scrollController';

export default {
  name: 'DigitalEmployee',
  components: {
    MessageItem,
    WorkspacePanel,
    FilePreviewDrawer,
    MentionInput,
    ModelSelect,
    GAFileUpload,
  },
  mixins: [streamStateManager, messageManager, fileManager, scrollController],
  data() {
    return {
      // 数字员工（固定绑定，不可切换）
      employeeId: '',
      employeeDetail: null,

      sidebarCollapsed: true,
      conversationList: [],
      currentThreadId: '',
      currentConversationTitle: '',
      searchText: '',
      pageNo: 1,
      pageSize: 50,
      hasMoreConversations: true,
      isLoadingMoreConversations: false,
      isLoadingHistory: false,

      inputMessage: '',
      selectedModel: '',
      modelList: [],
      modelLoading: false,

      currentRunId: '',
      currentStage: '',

      // Workspace 相关
      activeWorkspace: null,
      workspaceTrees: {},
      panelVisible: false,

      // 文件预览
      previewVisible: false,
      previewLoading: false,
      previewFileName: '',
      previewBlob: null,
      workspaceRect: null,
      resizeObserver: null,
    };
  },
  computed: {
    modelSelectWidth() {
      const selected = this.modelList.find(
        m => m.modelId === this.selectedModel,
      );
      const text = selected?.displayName || this.$t('common.model.select');
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      ctx.font = '14px sans-serif';
      const textWidth = ctx.measureText(text).width;
      return `${Math.ceil(textWidth) + 68}px`;
    },

    assistantAvatar() {
      return avatarSrc(this.employeeDetail?.avatar?.path);
    },
    welcomeText() {
      return this.employeeDetail?.name || '';
    },

    currentWorkspaceTree() {
      if (!this.activeWorkspace) return null;
      const key = `${this.activeWorkspace.threadId}-${this.activeWorkspace.runId}`;
      return this.workspaceTrees[key] || null;
    },

    currentTitle() {
      if (this.currentThreadId) {
        return (
          this.currentConversationTitle ||
          this.employeeDetail?.name ||
          this.$t('generalAgent.index.newConversation')
        );
      }
      return this.employeeDetail?.name || '';
    },
    canSend() {
      const hasContent =
        this.inputMessage.trim() || this.uploadedFiles.length > 0;
      const hasModel = !!this.selectedModel;
      return hasContent && hasModel && !this.isStreaming;
    },
    isEmptyConversation() {
      return this.messageList.length === 0;
    },
    inputPlaceholder() {
      // 优先使用后端 detail 接口返回的 placeholder，未配置时回退到通用文案
      return (
        this.employeeDetail?.placeholder ||
        this.$t('generalAgent.header.placeholder')
      );
    },
    // 当前选中的会话对象
    currentConversation() {
      if (!this.currentThreadId) return null;
      return this.conversationList.find(
        c => c.conversationId === this.currentThreadId,
      );
    },
  },
  watch: {
    currentThreadId(val) {
      if (!val) this.currentConversationTitle = '';
    },
    currentConversation(val) {
      if (val?.title) this.currentConversationTitle = val.title;
    },
    panelVisible(val) {
      if (val && this.activeWorkspace) {
        this.loadWorkspaceFiles();
        this.$nextTick(() => this.updateWorkspaceRect());
      } else if (!val) {
        // 工作空间关闭时，关闭文件预览
        this.previewVisible = false;
      }
    },
    previewVisible(val) {
      if (val) {
        this.sidebarCollapsed = true;
        this.$nextTick(() => this.updateWorkspaceRect());
      }
    },
  },
  mounted() {
    this.initFromRoute();
    this.fetchModelList();
    this.fetchConversationList();
    this.setupResizeObserver();
  },
  beforeDestroy() {
    this.resetWorkspace();
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }
  },
  methods: {
    // Workspace 面板控制
    handleWorkspaceActivity(content) {
      if (!content) return;
      const { runId, threadId, fileCount, totalSize, timestamp } = content;
      this.activeWorkspace = {
        runId,
        threadId,
        fileCount: fileCount || 0,
        totalSize: totalSize || 0,
        timestamp: timestamp || Date.now(),
      };
    },
    setActiveWorkspace(payload) {
      this.activeWorkspace = payload;
    },
    showPanel() {
      this.panelVisible = true;
    },
    hidePanel() {
      this.panelVisible = false;
    },
    setWorkspaceTree({ threadId, runId, data }) {
      const key = `${threadId}-${runId}`;
      this.$set(this.workspaceTrees, key, {
        files: data.files || [],
        fileCount: data.fileCount || 0,
        totalSize: data.totalSize || 0,
        isDisplay: data.isDisplay || false,
        loaded: true,
        loading: false,
      });
    },
    resetWorkspace() {
      this.activeWorkspace = null;
      this.workspaceTrees = {};
      this.panelVisible = false;
    },

    replaceRouteQuery(query) {
      const currentQuery = this.$route.query || {};
      const currentKeys = Object.keys(currentQuery);
      const nextKeys = Object.keys(query);
      const isSameQuery =
        currentKeys.length === nextKeys.length &&
        nextKeys.every(key => currentQuery[key] === query[key]);

      if (isSameQuery) return;

      this.$router
        .replace({
          path: this.$route.path,
          query,
        })
        .catch(() => {});
    },

    syncConversationRoute(threadId = '') {
      const query = { ...this.$route.query };
      if (threadId) {
        query.threadId = threadId;
      } else {
        delete query.threadId;
      }
      this.replaceRouteQuery(query);
    },

    initFromRoute() {
      const { employeeId, threadId } = this.$route.query || {};
      this.employeeId = employeeId || '';
      if (!this.employeeId) {
        this.$message.warning('缺少数字员工ID');
        this.$router.replace('/explore');
        return;
      }
      this.fetchEmployeeDetail();
      if (threadId) {
        this.selectConversation(threadId);
      } else {
        this.initNewConversation();
      }
    },

    async fetchEmployeeDetail() {
      if (!this.employeeId) return;
      try {
        const res = await getDigitalEmployeeDetail({
          employeeId: this.employeeId,
        });
        if (res.code === 0 && res.data) {
          this.employeeDetail = res.data;
        }
      } catch (error) {
        console.error('fetch employee detail error:', error);
      }
    },

    setupResizeObserver() {
      if (typeof ResizeObserver === 'undefined') return;
      this.resizeObserver = new ResizeObserver(() => {
        this.updateWorkspaceRect();
      });
      const pageEl = this.$el;
      if (pageEl) {
        this.resizeObserver.observe(pageEl);
      }
      const sidebar = pageEl?.querySelector('.sidebar');
      if (sidebar) {
        this.resizeObserver.observe(sidebar);
      }
    },

    updateWorkspaceRect() {
      this.$nextTick(() => {
        const workspaceEl = this.$refs.workspacePanel?.$el;
        if (workspaceEl) {
          this.workspaceRect = workspaceEl.getBoundingClientRect();
        } else {
          const mainContent = this.$el?.querySelector('.agent-main-content');
          if (mainContent) {
            const rect = mainContent.getBoundingClientRect();
            this.workspaceRect = { left: rect.right };
          }
        }
      });
    },

    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },

    async fetchModelList() {
      this.modelLoading = true;
      try {
        const res = await selectModelList();
        if (res.code === 0 && res.data?.list) {
          this.modelList = res.data.list.map(model => ({
            modelId: model.modelId,
            displayName: model.displayName,
            model: model.model,
            provider: model.provider,
            modelType: model.modelType,
            config: model.config,
            avatar: model.avatar,
            tags: model.tags || [],
          }));
          if (!this.selectedModel) {
            this.selectedModel = this.modelList[0]?.modelId || '';
          }
        }
      } finally {
        this.modelLoading = false;
      }
    },

    async fetchConversationList(loadMore = false) {
      if (this.isLoadingMoreConversations) return;

      if (loadMore) {
        if (!this.hasMoreConversations) return;
        this.pageNo += 1;
      } else {
        this.pageNo = 1;
        this.hasMoreConversations = true;
      }

      this.isLoadingMoreConversations = true;
      try {
        const res = await getDigitalEmployeeConversationList({
          employeeId: this.employeeId,
          pageNo: this.pageNo,
          pageSize: this.pageSize,
          searchText: this.searchText.trim(),
        });
        if (res.code === 0) {
          const list = res.data?.list || [];
          if (loadMore) {
            this.conversationList = [...this.conversationList, ...list];
          } else {
            this.conversationList = list;
          }
          if (list.length < this.pageSize) {
            this.hasMoreConversations = false;
          }
        } else if (loadMore) this.pageNo -= 1;
      } catch {
        if (loadMore) this.pageNo -= 1;
      } finally {
        this.isLoadingMoreConversations = false;
      }
    },

    /** 将指定会话移到列表顶部，用于历史会话有新消息时按更新时间置顶 */
    moveConversationToTop(conversationId) {
      if (!conversationId) return;
      const index = this.conversationList.findIndex(
        c => c.conversationId === conversationId,
      );
      if (index <= 0) return;
      const [conversation] = this.conversationList.splice(index, 1);
      this.conversationList.unshift(conversation);
      this.scrollConversationListToTop();
    },

    /** 历史对话列表滚动到顶部 */
    scrollConversationListToTop() {
      const el = this.$refs.conversationList;
      if (!el) return;
      el.scrollTop = 0;
    },

    /** 侧边栏会话列表滚动加载 */
    handleConversationListScroll(e) {
      const el = e.target;
      if (!el) return;
      const threshold = window.innerHeight;
      const distanceToBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
      if (distanceToBottom < threshold) {
        this.fetchConversationList(true);
      }
    },

    /** 按关键字搜索历史会话，重置分页后重新拉取 */
    handleConversationSearch() {
      this.searchText = this.searchText.trim();
      this.fetchConversationList(false);
    },

    initNewConversation() {
      this.currentThreadId = '';
      this.currentConversationTitle = '';
      this.syncConversationRoute();
      this.clearMessages('');
      this.resetScrollState();
      this.hidePanel();
      this.$nextTick(() => {
        if (this.modelList && this.modelList.length > 0) {
          this.selectedModel = this.modelList[0]?.modelId || '';
        }
      });
    },

    buildCurrentModelConfig() {
      const selectedModelConfig = this.selectedModel
        ? this.modelList.find(m => m.modelId === this.selectedModel)
        : this.modelList[0];
      return {
        modelId: selectedModelConfig?.modelId,
        model: selectedModelConfig?.model,
        provider: selectedModelConfig?.provider,
        displayName: selectedModelConfig?.displayName,
        modelType: selectedModelConfig?.modelType,
        config: selectedModelConfig?.config,
      };
    },

    async createConversationWithTitle(title) {
      if (!this.modelList || this.modelList.length === 0) {
        this.$message.warning(this.$t('generalAgent.error.modelListLoading'));
        return null;
      }

      const res = await createDigitalEmployeeConversation({
        employeeId: this.employeeId,
        title: title || this.$t('generalAgent.index.newConversation'),
        modelConfig: this.buildCurrentModelConfig(),
      });

      if (res.code === 0) {
        const conversationId = res.data?.conversationId;
        if (conversationId) {
          this.currentThreadId = conversationId;

          const oldMessages = this.messagesMap[''] || [];
          this.$set(this.messagesMap, conversationId, oldMessages);
          this.$delete(this.messagesMap, '');

          this.selectedModel = this.buildCurrentModelConfig().modelId;
          const newConversation = {
            conversationId,
            employeeId: this.employeeId,
            title: title || this.$t('generalAgent.index.newConversation'),
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          };
          this.conversationList.unshift(newConversation);
          this.scrollConversationListToTop();
          this.syncConversationRoute(conversationId);
          return conversationId;
        } else {
          this.$message.error(this.$t('generalAgent.error.createFailed'));
        }
      } else {
        this.$message.error(
          res.msg || this.$t('generalAgent.error.createError'),
        );
      }
      return null;
    },

    selectConversation(conversationId) {
      if (!conversationId) return;

      this.syncConversationRoute(conversationId);

      if (this.currentThreadId === conversationId) {
        return;
      }
      // 切换会话时，只切换 currentThreadId，不中止 SSE 流
      this.currentThreadId = conversationId;
      this.isLoadingHistory = true;
      this.resetScrollState();
      this.hidePanel();
      this.fetchHistory();
    },

    async fetchHistory() {
      if (!this.currentThreadId) return;

      const streaming = this.streamingMap[this.currentThreadId];
      if (streaming && streaming.isStreaming) {
        this.isLoadingHistory = false;
        this.loadConfig();
        return;
      }

      try {
        const res = await getDigitalEmployeeConversationDetail({
          conversationId: this.currentThreadId,
          pageNo: 1,
          pageSize: 100,
        });

        if (res.code === 0 && res.data?.list) {
          const allMessages = [];
          res.data.list.forEach(run => {
            // 后端返回的是 events 字段，需要聚合为消息
            if (run.events && Array.isArray(run.events)) {
              const messages = aggregateEventsToMessages(run.events);
              allMessages.push(...messages);
            }
            if (run.runId) this.currentRunId = run.runId;
          });

          this.$set(this.messagesMap, this.currentThreadId, allMessages);
          // 先关闭加载状态，让消息列表渲染
          this.isLoadingHistory = false;
          // 等待 DOM 渲染完成后滚动到底部
          this.$nextTick(() => {
            requestAnimationFrame(() => {
              this.scrollToBottom(true);
            });
          });
        } else if (res.code === 0 && !res.data?.list) {
          this.isLoadingHistory = false;
        } else {
          this.isLoadingHistory = false;
        }
        this.loadConfig();
        // detail 加载完成后，检查是否有运行中的对话需要重连 connect（断线保持）
        await this.checkAndResumePending(this.currentThreadId);
      } finally {
        this.isLoadingHistory = false;
      }
    },

    // 检查会话是否有运行中的对话，有则走 connect 重连恢复实时流（对齐通用智能体断线重连）
    appendPendingInputMessages(threadId, inputMessages) {
      const pendingMessages =
        aggregateInputMessagesToUserMessages(inputMessages);
      if (!pendingMessages.length) return;

      const messages = this.ensureMessageList(threadId);
      const existingIds = new Set(messages.map(msg => msg.id).filter(Boolean));

      pendingMessages.forEach(message => {
        if (existingIds.has(message.id)) return;
        messages.push(message);
        existingIds.add(message.id);
      });
    },

    async checkAndResumePending(threadId) {
      if (!threadId) return;

      // 已在流式中则不再重连
      const streaming = this.streamingMap[threadId];
      if (streaming && streaming.isStreaming) return;

      let hasPending = false;
      let pendingMessages = [];
      try {
        const res = await getGeneralAgentConversationPending({ threadId });
        hasPending =
          res.code === 0 && res.data?.hasPendingConversation === true;
        pendingMessages = Array.isArray(res.data?.messages)
          ? res.data.messages
          : [];
      } catch (error) {
        console.error('check pending conversation failed:', error);
        return;
      }

      if (!hasPending) return;

      this.appendPendingInputMessages(threadId, pendingMessages);

      // 初始化流式状态，承载进行中那一轮的助手消息
      const { abortController, assistantMessage } =
        this.initStreamState(threadId);
      this.addAssistantMessage(threadId, assistantMessage);
      this.currentStage = 'understanding';
      this.resetScrollState();

      const parser = new SSEEventParser();
      let isUserAborted = false;

      try {
        await connectGeneralAgentConversation({
          threadId,
          onMessage: event => {
            this.handleSSEEvent(event, assistantMessage, parser, threadId);
          },
          onError: error => {
            console.error('Reconnect SSE Error:', error);
            if (this.currentThreadId === threadId) {
              this.$message.error(
                this.$t('generalAgent.error.chatRequestFailed'),
              );
            }
            this.cleanupStreamState(threadId);
            assistantMessage.isStreaming = false;
            this.setFragmentsNotStreaming(assistantMessage.fragments);
          },
          signal: abortController.signal,
        });
      } catch (error) {
        console.error('Reconnect stream error:', error);
        isUserAborted = error.name === 'AbortError';
        if (!isUserAborted && this.currentThreadId === threadId) {
          this.$message.error(
            this.$t('generalAgent.error.sendMessageFailed') +
              (error.message || error),
          );
        }
      } finally {
        this.finalizeStream(threadId, assistantMessage, isUserAborted);
      }

      return true;
    },

    async loadConfig() {
      if (!this.currentThreadId) return;
      try {
        const res = await getDigitalEmployeeConversationConfig({
          conversationId: this.currentThreadId,
        });
        if (res.code === 0 && res.data) {
          if (res.data.modelConfig?.modelId) {
            this.selectedModel = res.data.modelConfig.modelId;
          }
        }
      } catch (error) {
        console.error('load config error:', error);
      }
    },

    handleKeyDown(e) {
      if (e.shiftKey) return;
      e.preventDefault();
      this.sendMessage();
    },

    async beforeEnterSubmit() {
      const content = this.inputMessage.trim();
      if (!content && this.uploadedFiles.length === 0) return false;

      const currentStreaming = this.streamingMap[this.currentThreadId];
      if (currentStreaming && currentStreaming.isStreaming) return false;

      return true;
    },

    handleModelChange(value) {
      if (!this.currentThreadId) return;
      const selectedModelConfig = this.modelList.find(m => m.modelId === value);
      updateDigitalEmployeeConversationConfig({
        conversationId: this.currentThreadId,
        modelConfig: {
          modelId: value,
          model: selectedModelConfig?.model,
          provider: selectedModelConfig?.provider,
          displayName: selectedModelConfig?.displayName,
          modelType: selectedModelConfig?.modelType || 'llm',
          config: selectedModelConfig?.config || {},
        },
      });
    },

    async sendMessage() {
      const content = this.inputMessage.trim();
      const messageFiles = this.uploadedFiles;
      if (!content && messageFiles.length === 0) return;

      // 检查当前会话是否正在流式传输
      const currentStreaming = this.streamingMap[this.currentThreadId];
      if (currentStreaming && currentStreaming.isStreaming) return;

      if (!this.currentThreadId) {
        const title = content.slice(0, 50);
        const conversationId = await this.createConversationWithTitle(title);
        if (!conversationId) {
          this.$message.error(
            this.$t('generalAgent.error.createConversationFailed'),
          );
          return;
        }
      }

      const userMessage = this.buildUserMessage(content, messageFiles);
      this.ensureMessageList(this.currentThreadId);
      this.addUserMessage(this.currentThreadId, content, messageFiles);

      this.clearFiles();
      this.$refs.mentionInput?.clear();
      this.$nextTick(() => this.scrollToBottom());

      await this.startStreaming(userMessage);
    },

    async startStreaming(userMessage) {
      if (!this.currentThreadId) {
        this.$message.error(
          this.$t('generalAgent.error.conversationIdNotExist'),
        );
        return;
      }

      const streamingThreadId = this.currentThreadId;

      // 历史会话有新消息时置顶，与后端按更新时间排序保持一致
      this.moveConversationToTop(streamingThreadId);

      // 使用 mixin 初始化流式状态
      const { abortController, assistantMessage } =
        this.initStreamState(streamingThreadId);

      // 添加消息到对应会话的消息列表
      this.addAssistantMessage(streamingThreadId, assistantMessage);

      this.currentStage = 'understanding';
      this.resetScrollState();

      const parser = new SSEEventParser();
      let isUserAborted = false;

      try {
        await chatDigitalEmployeeConversation({
          employeeId: this.employeeId,
          conversationId: streamingThreadId,
          messages: [userMessage],
          onMessage: event => {
            this.handleSSEEvent(
              event,
              assistantMessage,
              parser,
              streamingThreadId,
            );
          },
          onError: error => {
            console.error('SSE Error:', error);
            if (this.currentThreadId === streamingThreadId) {
              this.$message.error(
                this.$t('generalAgent.error.chatRequestFailed'),
              );
            }
            this.cleanupStreamState(streamingThreadId);
            assistantMessage.isStreaming = false;
            this.setFragmentsNotStreaming(assistantMessage.fragments);
          },
          signal: abortController.signal,
        });
      } catch (error) {
        console.error('Stream error:', error);
        // 判断是否是用户主动中止
        isUserAborted = error.name === 'AbortError';

        if (!isUserAborted && this.currentThreadId === streamingThreadId) {
          this.$message.error(
            this.$t('generalAgent.error.sendMessageFailed') +
              (error.message || error),
          );
        }
      } finally {
        // 统一清理（非用户主动中止）；用户中止由 stopStreaming 处理
        this.finalizeStream(streamingThreadId, assistantMessage, isUserAborted);
      }
    },

    // 重新生成 - 找到上一条用户消息并重新发送
    handleRegenerate(message) {
      if (this.isStreaming) return;

      // 找到这条助手消息的索引
      const messageIndex = this.messageList.findIndex(m => m.id === message.id);
      if (messageIndex <= 0) return;

      // 找到上一条用户消息
      let userMessage = null;
      for (let i = messageIndex - 1; i >= 0; i--) {
        if (this.messageList[i].role === 'user') {
          userMessage = this.messageList[i];
          break;
        }
      }

      if (!userMessage) return;

      // 删除当前助手消息
      this.removeMessage(this.currentThreadId, message.id);

      // 构建请求消息
      const requestMessage = this.buildRequestMessage(userMessage);

      // 直接调用 startStreaming
      this.$nextTick(() => {
        this.startStreaming(requestMessage);
      });
    },

    // Human-in-the-Loop 回调（数字员工固定 EnableHumanInTheLoop=false，此处留空）
    handleQuestionReply() {},
    handleQuestionReject() {},

    async handleDeleteConversation(item) {
      try {
        await this.$confirm(
          this.$t('generalAgent.index.confirmDeleteConversation'),
          this.$t('common.button.tip'),
          { type: 'warning' },
        );
      } catch (e) {
        return;
      }
      const res = await deleteDigitalEmployeeConversation({
        conversationId: item.conversationId,
      });
      if (res.code === 0) {
        this.$message.success(this.$t('common.info.delete'));
        if (this.currentThreadId === item.conversationId) {
          this.clearMessages(this.currentThreadId);
          this.currentThreadId = '';
          this.currentConversationTitle = '';
          this.syncConversationRoute();
          this.hidePanel();
        }
        const index = this.conversationList.findIndex(
          c => c.conversationId === item.conversationId,
        );
        if (index !== -1) {
          this.conversationList.splice(index, 1);
        }
      }
    },

    // 处理停止按钮点击：先通知服务端 cancel，再本地中断 SSE（对齐通用智能体断线保持停止流程）
    async handleStopClick() {
      if (this.currentThreadId) {
        try {
          await cancelGeneralAgentConversation({
            threadId: this.currentThreadId,
          });
        } catch (error) {
          // 取消失败不阻塞本地清理
          console.error('cancel conversation failed:', error);
        }
      }
      this.stopStreaming(this.currentThreadId);
    },

    async loadWorkspaceFiles() {
      if (!this.activeWorkspace || !this.currentThreadId) return;
      try {
        const res = await getGeneralAgentWorkspace({
          threadId: this.currentThreadId,
          runId: this.activeWorkspace.runId,
        });
        if (res.code === 0 && res.data) {
          this.setWorkspaceTree({
            threadId: this.currentThreadId,
            runId: this.activeWorkspace.runId,
            data: res.data,
          });
        }
      } catch (error) {
        console.error('loadWorkspaceFiles error:', error);
      }
    },

    handleViewWorkspace(data) {
      this.setActiveWorkspace({
        runId: data.runId,
        threadId: data.threadId || this.currentThreadId,
        fileCount: data.fileCount || 0,
        totalSize: data.totalSize || 0,
        timestamp: Date.now(),
      });
      this.showPanel();
    },

    async handlePreviewFile(data) {
      const { file, filePath, threadId, runId } = data;

      this.previewFileName = file.name;
      this.previewVisible = true;
      this.previewLoading = true;
      this.previewBlob = null;

      try {
        this.previewBlob = await previewGeneralAgentWorkspace({
          threadId,
          runId,
          path: filePath,
        });
      } finally {
        this.previewLoading = false;
      }
    },

    // 下载整个工作空间
    async handleDownloadAll() {
      try {
        const blob = await downloadGeneralAgentWorkspace({
          threadId: this.currentThreadId,
          runId: this.currentRunId,
          path: '',
        });
        resDownloadFile(blob, this.$t('generalAgent.index.workspaceZip'));
        this.$message.success(
          this.$t('generalAgent.workspace.downloadSuccess'),
        );
      } catch (error) {
        console.error(
          this.$t('generalAgent.index.downloadWorkspaceFailed'),
          error,
        );
        this.$message.error(this.$t('generalAgent.workspace.downloadFailed'));
      }
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/views/generalAgent/styles/variables';

.general-agent-page {
  display: flex;
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #f5f7fa;
  overflow: hidden;
  padding: 16px;
  box-sizing: border-box;

  ::v-deep .model-select .el-input__inner {
    border-radius: 12px;
  }
}

.agent-main-content {
  flex: 1;
  display: flex;
  min-width: 0;
  min-height: 0;
  position: relative;
  overflow: hidden;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

  // 没有工作空间且没有预览时，right-area 宽度为 0
  &:not(.has-workspace):not(.has-preview) {
    .right-area {
      width: 0;
      overflow: hidden;
    }

    .message-list,
    .input-area .input-container {
      max-width: $message-main-max-width;
    }
  }

  // 有工作空间时，左侧自适应，右侧工作空间宽度
  &.has-workspace:not(.has-preview) {
    .right-area {
      width: 400px;
    }

    .message-list,
    .input-area .input-container {
      max-width: $message-workspace-max-width;
    }
  }

  // 有预览时，左侧 40%，右侧 60%
  &.has-preview {
    .center-panel {
      width: 40%;
    }

    .right-area {
      width: 60%;
    }

    .message-list,
    .input-area .input-container {
      max-width: 100%;
    }
  }

  // 有工作空间 + 预览时，右侧同时包含两者
  &.has-workspace.has-preview {
    .right-area {
      width: 60%;
    }
  }
}

// 左侧会话列表 - 固定宽度
.sidebar {
  flex: none;
  display: flex;
  flex-direction: column;
  width: 240px;
  height: 100%;
  flex-shrink: 0;
  background: #f9fafb;
  border-right: 1px solid #f0f0f0;
  transition: width 0.3s ease;
  overflow: hidden;

  &.collapsed {
    width: 0;
    border-right: none;
  }

  .sidebar-header {
    flex-shrink: 0;
    padding: 12px 22px;
    border-bottom: 1px solid #f0f0f0;

    .history-toolbar {
      display: flex;
      align-items: center;
      gap: 8px;

      .create-conversation-btn {
        flex: 0 0 auto;
        padding: 6px;

        .create-icon {
          font-size: 16px;
        }
      }

      ::v-deep .history-search-input {
        flex: 1;
        min-width: 0;

        .el-input__inner {
          border-radius: 12px;
        }
      }
    }
  }

  .sidebar-divider {
    height: 1px;
    background: #f0f0f0;
    flex-shrink: 0;
  }

  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
    min-height: 0;

    &::-webkit-scrollbar {
      width: 4px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background: #d1d5db;
      border-radius: 2px;
    }

    .conversation-loading-more,
    .conversation-no-more {
      text-align: center;
      padding: 12px 0;
      color: $wga-text-muted;
      font-size: 12px;
    }

    .conversation-loading-more i {
      font-size: 18px;
    }
  }

  .conversation-item {
    display: flex;
    align-items: center;
    padding: 12px 14px;
    border-radius: 10px;
    cursor: pointer;
    margin-bottom: 4px;
    transition: background-color 0.2s;

    &:hover {
      background: rgba($wga-primary, 0.08);

      .conversation-delete {
        opacity: 1;
      }
    }

    &.active {
      background: rgba($wga-primary, 0.12);

      .conversation-title {
        font-weight: 500;
      }
    }

    i:first-child {
      margin-right: 10px;
      color: $wga-text-muted;
      font-size: 16px;
    }

    .conversation-title {
      flex: 1;
      font-size: 14px;
      color: $wga-text;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .conversation-delete {
      opacity: 0;
      color: $wga-text-muted;
      padding: 4px;
      font-size: 16px;
      transition: all 0.2s;
      cursor: pointer;

      &:hover {
        color: #f56c6c;
      }
    }
  }
}

// 中间区域：主对话
.center-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  position: relative;
  overflow: hidden;
  background: #fff;
  transition: width 0.3s ease;
}

// 右侧区域：工作空间 + 预览
.right-area {
  flex: none;
  display: flex;
  min-width: 0;
  min-height: 0;
  position: relative;
  overflow: hidden;
  transition: width 0.3s ease;

  // Workspace 面板：固定宽度
  ::v-deep .workspace-panel {
    border-left: 1px solid #f0f0f0;
  }

  // 预览面板：占据剩余空间
  .preview-panel {
    flex: 1;
    min-width: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
    border-left: 1px solid #f0f0f0;
  }
}

.header {
  flex: none;
  height: $header-height;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  border-radius: 12px 12px 0 0;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-btns {
    display: flex;
    gap: 12px;
  }

  .header-title {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    font-size: 16px;
    font-weight: 600;
    color: $wga-text;
  }

  .header-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border: 1px solid $wga-border;
    background: #fff;
    border-radius: 8px;
    cursor: pointer;
    color: $wga-text-muted;
    transition: all 0.2s;

    &:hover {
      border-color: $wga-primary;
      color: $wga-primary;
      background: rgba($wga-primary, 0.05);
    }

    i {
      font-size: 16px;
    }
  }
}

.message-area {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  background: #fff;
  position: relative;

  &.empty {
    flex: none;
    min-height: 0;
    height: 0;
    overflow: hidden;
  }

  .message-list {
    max-width: $message-max-width;
    width: 100%;
    box-sizing: border-box;
    margin: 0 auto;
    padding: 24px;
    min-height: 100%;
  }

  .history-loading {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #909399;
    font-size: 14px;
    gap: 12px;
    background: #fff;
    z-index: 10;

    i {
      font-size: 32px;
      color: #10a37f;
    }
  }
}

.scroll-to-bottom-btn {
  position: absolute;
  bottom: calc(100% + 10px);
  left: 50%;
  transform: translateX(-50%);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  color: #10a37f;
  border: 1px solid #10a37f;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.2s ease;
  z-index: 100;

  &:hover {
    background: #10a37f;
    color: #fff;
    transform: translateX(-50%) translateY(-2px);
    box-shadow: 0 4px 12px rgba(16, 163, 127, 0.4);
  }

  svg {
    width: 16px;
    height: 16px;
  }
}

.scroll-btn-fade-enter-active,
.scroll-btn-fade-leave-active {
  transition: all 0.3s ease;
}

.scroll-btn-fade-enter,
.scroll-btn-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(20px);
}

.input-area {
  position: relative;
  flex: none;
  background: #fff;
  padding: 16px 24px 24px;
  border-radius: 0 0 12px 12px;

  &.is-centered {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    border-top: none;
    padding: 0 24px;

    .input-container {
      max-width: 800px;
      width: 100%;
    }

    .welcome-section {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin-bottom: 32px;

      .welcome-avatar {
        width: 72px;
        height: 72px;
        border-radius: 20px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 20px;
        background: #fff;
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
        overflow: hidden;

        img {
          width: 100%;
          height: 100%;
          border-radius: 20px;
          object-fit: cover;
        }
      }

      .welcome-title {
        font-size: 28px;
        color: $wga-text;
        font-weight: 600;
      }
    }
  }

  &:not(.is-centered) {
    border-top: none;
  }

  .input-container {
    max-width: $message-max-width;
    width: 100%;
    box-sizing: border-box;
    margin: 0 auto;
    background: #fff;
    border-radius: 16px;
    border: 1px solid #e5e7eb;
    padding: 16px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
    transition:
      border-color 0.2s,
      box-shadow 0.2s;

    &:focus-within {
      border-color: $wga-primary;
      box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
    }
  }

  .file-preview {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;

    .echo-img-box {
      position: relative;
      display: inline-block;

      // 图片样式
      .echo-img {
        width: 48px;
        height: 48px;
        border-radius: 8px;
        cursor: pointer;
      }

      // 文档样式
      .echo-doc-box {
        background: #fff;
        min-width: 200px;
        max-width: 300px;
        border: 1px solid #dcdfe6;
        border-radius: 5px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 50px 10px 5px;

        .docIcon {
          width: 30px;
          height: 30px;
          flex-shrink: 0;
        }

        .docInfo {
          flex: 1;
          margin-left: 8px;
          overflow: hidden;

          .docInfo_name {
            color: #333;
            font-size: 13px;
            margin: 0;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }

          .docInfo_size {
            color: #bbbbbb;
            font-size: 12px;
            margin: 4px 0 0 0;
          }
        }
      }

      // 关闭按钮
      .echo-close {
        position: absolute;
        top: -6px;
        right: -6px;
        width: 18px;
        height: 18px;
        background: #ef4444;
        color: #fff;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        font-size: 12px;
        transition: transform 0.2s;
        z-index: 10;

        &:hover {
          transform: scale(1.1);
        }
      }
    }

    &.is-uploading {
      .echo-close {
        display: none;
      }
    }
  }

  .input-wrapper {
    ::v-deep .el-textarea {
      .el-textarea__inner {
        background: transparent;
        border: none;
        padding: 0;
        resize: none;
        font-size: 16px;
        line-height: 1.6;
        color: $wga-text;

        &::placeholder {
          color: #9ca3af;
        }
      }
    }
  }

  .input-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #f3f4f6;

    .toolbar-left {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .toolbar-right {
      display: flex;
      align-items: center;
      gap: 8px;

      .action-icon {
        font-size: 18px;
        color: $wga-text-muted;
        cursor: pointer;
        padding: 8px;
        border-radius: 8px;
        transition: all 0.2s;

        &:hover {
          color: $wga-primary;
          background: rgba($wga-primary, 0.08);
        }
      }

      .send-btn {
        padding: 10px;
        border: none;
        color: #5147ff;
        line-height: 0;
        display: inline-flex;
        justify-content: center;
        align-items: center;

        &:hover {
          background-color: rgba(87, 104, 161, 0.08);
        }

        .send-icon {
          width: 18px;
          height: 18px;
          fill: currentColor;
        }

        .stop-icon {
          width: 16px;
          height: 16px;
          fill: currentColor;
        }
      }

      .stop-btn {
        background: #fff !important;
        border: 1px solid #333 !important;
        color: #333 !important;

        &:hover,
        &:focus {
          background: #333 !important;
          border-color: #333 !important;
          color: #fff !important;
        }
      }
    }
  }

  .input-footer {
    text-align: center;
    font-size: 12px;
    color: $wga-text-muted;
    margin-top: 12px;
  }
}

// Workspace 面板过渡动画 — 从页面右侧滑入
.workspace-slide-enter-active,
.workspace-slide-leave-active {
  transition: all 0.3s ease;
}

.workspace-slide-enter,
.workspace-slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
