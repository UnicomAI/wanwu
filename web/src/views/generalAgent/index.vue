<template>
  <div class="general-agent-page">
    <!-- 主内容区 -->
    <div
      :class="{
        'has-sidebar': !sidebarCollapsed,
        'has-workspace': panelVisible && activeWorkspace,
        'has-preview': previewVisible || shouldShowSkillTabs,
      }"
      class="agent-main-content"
    >
      <!-- 左侧会话列表 - 固定宽度，可折叠 -->
      <div :class="['sidebar', { collapsed: sidebarCollapsed }]">
        <div class="sidebar-header">
          <el-button
            class="new-chat-btn"
            type="primary"
            @click="initNewConversation"
          >
            <i class="el-icon-plus"></i>
            {{ $t('generalAgent.sidebar.newChat') }}
          </el-button>
        </div>

        <div class="sidebar-divider"></div>

        <div
          ref="conversationList"
          class="conversation-list"
          @scroll="handleConversationListScroll"
        >
          <div
            v-for="item in conversationList"
            :key="item.threadId"
            :class="[
              'conversation-item',
              { active: currentThreadId === item.threadId },
            ]"
            @click="selectConversation(item.threadId, item)"
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

      <!-- 中间区域：主消息区域（自适应）或 40% 宽度 -->
      <div class="center-panel">
        <!-- 顶部标题栏 -->
        <div class="header">
          <div class="header-left">
            <button
              v-if="mode === 'skill'"
              class="header-icon-btn"
              :title="$t('menu.back')"
              :aria-label="$t('menu.back')"
              @click="$router.back()"
            >
              <i class="el-icon-arrow-left"></i>
            </button>

            <div v-if="mode !== 'skill'" class="header-btns">
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

          <div v-if="isSkillType" class="header-right">
            <!-- Skill 模式下不再使用传统的 WorkspacePanel 入口，改为由 SkillTabs 集成代码编辑 -->
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
                :disabled="isStreaming || currentPreviewIsStreaming"
                :isDIP="selectedMode?.value === 'DIP Agent'"
                :placeholder="inputPlaceholder"
                :before-enter-submit="beforeEnterSubmit"
                @keydown-enter="handleKeyDown"
              />
            </div>

            <!-- 底部工具栏：配置按钮 + 发送按钮 -->
            <div class="input-toolbar">
              <div class="toolbar-left">
                <div
                  v-show="selectedMode?.value !== 'DIP Agent'"
                  class="config-btn"
                  @click="showConfigDialog = true"
                >
                  <i class="el-icon-setting"></i>
                  <span>{{ $t('generalAgent.header.config') }}</span>
                </div>
                <!-- 导入Skill按钮 -->
                <div
                  v-show="isSkillType && !currentThreadId"
                  class="config-btn"
                  @click="openImportSkillDialog"
                >
                  <i class="el-icon-download"></i>
                  <span>{{ $t('generalAgent.header.importSkill') }}</span>
                </div>
                <!-- 已选模式标签 -->
                <div v-show="selectedMode" class="mode-btn mode-btn-selected">
                  <img
                    v-if="selectedMode?.avatar"
                    :src="selectedMode?.avatar"
                    alt=""
                    class="mode-avatar"
                  />
                  <i v-else :class="selectedMode?.icon"></i>
                  <span>{{ selectedMode?.label }}</span>
                  <i
                    v-show="
                      mode !== 'skill' &&
                      (!isSkillType || (isSkillType && !currentThreadId))
                    "
                    class="el-icon-close"
                    @click="removeMode(selectedMode.value)"
                  ></i>
                </div>
              </div>
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
          <!-- 可选模式按钮区域 - 仅在未选择模式且有数据时显示 -->
          <div
            v-if="!selectedMode && Object.keys(modeOptions).length > 0"
            class="mode-buttons"
          >
            <!-- 模式按钮列表 -->
            <div
              v-for="(mode, key) in modeOptions"
              v-show="
                mode.value !== MAIN_CHAT_MODES.SKILL_MODE || !currentThreadId
              "
              :key="key"
              class="mode-btn"
              @click="addMode(mode.value)"
            >
              <img
                v-if="mode?.avatar"
                :src="mode?.avatar"
                alt=""
                class="mode-avatar"
              />
              <i v-else :class="mode?.icon"></i>
              <span>{{ mode?.label }}</span>
            </div>
          </div>
          <div class="input-footer">
            <span>{{ $t('generalAgent.header.footer') }}</span>
          </div>
        </div>
      </div>

      <!-- 右侧区域：工作空间 + 预览（固定宽度或 60%） -->
      <div class="right-area">
        <!-- Workspace 面板 -->
        <transition
          :name="
            shouldShowSkillTabs || previewVisible
              ? 'workspace-slide-inner'
              : 'workspace-slide'
          "
        >
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

        <!-- 预览面板：文件预览或 SkillTabs -->
        <div v-if="previewVisible || shouldShowSkillTabs" class="preview-panel">
          <!-- 文件预览抽屉 -->
          <file-preview-drawer
            v-if="previewVisible"
            :blob="previewBlob"
            :file-name="previewFileName"
            :loading="previewLoading"
            :visible.sync="previewVisible"
            @close="previewVisible = false"
          />

          <!--
              预览会话面板生命周期说明：
              1. 切换会话或 previewId 变化时，skillTabsKey 会变化，组件会卸载重建，用于重置上一轮预览会话及 SSE 状态。
              2. 打开文件预览抽屉时，仅通过 v-show 隐藏面板，不卸载组件，避免中断正在进行的预览 SSE。
              3. SkillTabs 和文件预览抽屉互斥显示，占用相同空间
            -->
          <SkillTabs
            ref="skillTabs"
            v-if="shouldMountSkillTabs"
            v-show="shouldShowSkillTabs"
            :key="skillTabsKey"
            :skillPreviewParams="skillPreviewParams"
            @view-workspace="handleViewWorkspace"
            @refresh-workspace="loadWorkspaceFiles"
            @security-review="handleSecurityReview"
          />
        </div>
      </div>

      <!-- 配置弹窗 -->
      <configDialog
        ref="configDialog"
        :visible.sync="showConfigDialog"
        :agent-id="selectedMode?.value ?? ''"
      />

      <!-- skill导入弹框 -->
      <SkillDialog
        ref="importSkillDialogRef"
        @submit="handleSkillImportSubmit"
      />
    </div>
  </div>
</template>

<script>
import MessageItem from './components/MessageItem.vue';
import WorkspacePanel from './components/WorkspacePanel.vue';
import FilePreviewDrawer from './components/FilePreviewDrawer.vue';
import ConfigDialog from './components/ConfigDialog.vue';
import MentionInput from './components/MentionInput.vue';
import SkillTabs from './components/skills/SkillTabs.vue';
import SkillDialog from './components/skills/SkillDialog.vue';
import ModelSelect from '@/components/modelSelect.vue';
import GAFileUpload from './components/GAFileUpload.vue';
import {
  cancelGeneralAgentConversation,
  cancelGeneralAgentSkillConversation,
  chatGeneralAgentConversation,
  chatGeneralAgentSkillConversation,
  checkGeneralAgentConversationConfig,
  connectGeneralAgentConversation,
  connectGeneralAgentSkillConversation,
  createGeneralAgentConversation,
  createGeneralAgentSkillConversation,
  deleteGeneralAgentConversation,
  downloadGeneralAgentWorkspace,
  getGeneralAgentConversationConfig,
  getGeneralAgentConversationDetail,
  getGeneralAgentConversationList,
  getGeneralAgentConversationPending,
  getGeneralAgentSkillConversationPending,
  getGeneralAgentWorkspace,
  importGeneralAgentSkillConversation,
  previewGeneralAgentWorkspace,
  refreshGeneralAgentSkillConversation,
  updateGeneralAgentConversationConfig,
} from '@/api/generalAgent';
import { getCustomSkillInfo } from '@/api/templateSquare';
import { selectModelList } from '@/api/modelAccess';
import { avatarSrc, resDownloadFile } from '@/utils/util';
import { mapActions, mapGetters } from 'vuex';
import { SSEEventParser } from './utils/sse-parser';
// 引入工具函数
import {
  aggregateEventsToMessages,
  aggregateInputMessagesToUserMessages,
} from './utils/message-aggregator';
// 引入 Mixins
import streamStateManager from './mixins/streamStateManager';
import messageManager from './mixins/messageManager';
import fileManager from './mixins/fileManager';
import scrollController from './mixins/scrollController';
import modeManager from './mixins/modeManager';
import skillManager from './mixins/skillManager';

// 引入常量
import { MAIN_CHAT_MODES } from './constants';

export default {
  name: 'GeneralAgent',
  components: {
    MessageItem,
    WorkspacePanel,
    FilePreviewDrawer,
    ConfigDialog: ConfigDialog,
    MentionInput,
    SkillTabs,
    ModelSelect,
    GAFileUpload,
    SkillDialog, // skill导入弹框
  },
  mixins: [
    skillManager,
    streamStateManager,
    messageManager,
    fileManager,
    scrollController,
    modeManager,
  ],
  props: {
    mode: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      MAIN_CHAT_MODES,
      sidebarCollapsed: true,
      conversationList: [],
      currentThreadId: '',
      pageNo: 1,
      pageSize: 50,
      hasMoreConversations: true,
      isLoadingMoreConversations: false,
      isLoadingHistory: false,

      inputMessage: '',
      selectedModel: '',
      modelList: [],
      modelLoading: false,
      showConfigDialog: false,

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
      previewBlob: null, // 只存储 blob
      workspaceRect: null,
      resizeObserver: null,

      // skill 相关
      chatType: '', // skill 则 selectedMode.value 为 Skill Chat Agent
      customSkillId: '',
      previewId: '',
      skillChatMode: 'normal', // normal | import | convert
      pendingSkillWorkspaceRefresh: false,
    };
  },
  computed: {
    ...mapGetters('user', ['commonInfo']),

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
      return avatarSrc(this.commonInfo?.data?.generalAgent?.logo?.path);
    },
    welcomeText() {
      return this.commonInfo?.data?.generalAgent?.welcomeText;
    },

    currentWorkspaceTree() {
      if (!this.activeWorkspace) return null;
      const key = `${this.activeWorkspace.threadId}-${this.activeWorkspace.runId}`;
      return this.workspaceTrees[key] || null;
    },

    currentTitle() {
      if (!this.currentThreadId) return '';
      return (
        this.currentConversation?.title ||
        this.$t('generalAgent.index.newConversation')
      );
    },
    canSend() {
      const hasContent =
        this.inputMessage.trim() || this.uploadedFiles.length > 0;
      const hasModel = !!this.selectedModel;
      // 检查当前会话或预览面板是否正在流式传输
      return (
        hasContent &&
        hasModel &&
        !this.isStreaming &&
        !this.currentPreviewIsStreaming
      );
    },
    isEmptyConversation() {
      return this.messageList.length === 0;
    },
    inputPlaceholder() {
      // 如果有选中的模式，使用模式的 placeholder
      if (this.selectedMode) {
        const modeConfig = this.modeOptions[this.selectedMode.value];
        if (modeConfig && modeConfig.placeholder) {
          return modeConfig.placeholder;
        }
      }
      // 默认 placeholder
      return this.$t('generalAgent.header.placeholder');
    },
    // skill预览相关参数
    skillPreviewParams() {
      return {
        customSkillId: this.customSkillId,
        previewId: this.previewId,
        threadId: this.currentThreadId,
      };
    },
    // 创建skill模式
    isSkillType() {
      return this.chatType === 'skill';
    },
    isActiveSkillMode() {
      return (
        this.isSkillType &&
        this.selectedMode?.value === MAIN_CHAT_MODES.SKILL_MODE
      );
    },
    // 当前会话的skill预览是否正在流式传输
    currentPreviewIsStreaming() {
      if (!this.previewId) return false;
      return this.previewStreamingIds.includes(this.previewId);
    },
    // 控制组件是否挂载：会话身份存在时才挂载，切换会话/previewId 时配合 key 卸载重建以重置 SSE。
    shouldMountSkillTabs() {
      return (
        this.isActiveSkillMode &&
        !!this.currentThreadId &&
        !!this.customSkillId &&
        !!this.previewId
      );
    },
    // 控制组件是否可见：文件预览打开时只隐藏 SkillTabs，不卸载，避免中断内部预览 SSE。
    shouldShowSkillTabs() {
      return this.shouldMountSkillTabs && !this.previewVisible;
    },
    skillTabsKey() {
      return `${this.currentThreadId}_${this.previewId}`;
    },
    // 当前选中的会话对象
    currentConversation() {
      if (!this.currentThreadId) return null;
      return this.conversationList.find(
        c => c.threadId === this.currentThreadId,
      );
    },
  },
  watch: {
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
    shouldShowSkillTabs(val) {
      if (val) {
        this.sidebarCollapsed = true;
        if (this.pendingSkillWorkspaceRefresh) {
          this.$nextTick(() => this.refreshSkillWorkspace());
        }
      }
    },
  },
  mounted() {
    this.initFromRoute();
    this.fetchModelList();
    this.fetchConversationList();
    this.initUserInfo();
    this.setupResizeObserver();
    this.fetchModeOptions().then(() => {
      if (this.isSkillType) {
        this.addMode(MAIN_CHAT_MODES.SKILL_MODE);
      }
    });
  },
  beforeDestroy() {
    this.resetWorkspace();
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }
  },
  methods: {
    ...mapActions('user', ['getPermissionInfo', 'getCommonInfo']),

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

    async initUserInfo() {
      if (localStorage.getItem('access_cert')) {
        await this.getPermissionInfo();
        await this.getCommonInfo();
      }
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

    syncConversationRoute(threadId = '', conversation = null) {
      const query = { ...this.$route.query };
      const isSkillConversation = conversation
        ? !!conversation.isSkillConversation
        : this.isSkillType;

      delete query.chatMode;

      if (threadId) {
        query.threadId = threadId;

        if (isSkillConversation) {
          const skillId = conversation?.skillId || this.customSkillId;
          query.chatType = 'skill';
          if (skillId) {
            query.customSkillId = skillId;
          } else {
            delete query.customSkillId;
          }
        } else {
          delete query.chatType;
          delete query.customSkillId;
        }
      } else {
        delete query.threadId;

        if (isSkillConversation) {
          query.chatType = 'skill';
          if (this.customSkillId) {
            query.customSkillId = this.customSkillId;
          } else {
            delete query.customSkillId;
          }
        } else {
          delete query.chatType;
          delete query.customSkillId;
        }
      }

      this.replaceRouteQuery(query);
    },

    shouldPreserveSkillContext() {
      return this.mode === 'skill';
    },

    initFromRoute() {
      const { chatType, customSkillId, chatMode, threadId } =
        this.$route.query || {};

      let initialChatMode = '';

      if (chatType) {
        this.chatType = chatType;
      }
      if (customSkillId) {
        this.customSkillId = customSkillId;
      }
      if (chatMode) {
        initialChatMode = chatMode;
        this.skillChatMode = 'normal';

        // 防止页面刷新再次进入自动触发模式
        const newQuery = { ...this.$route.query };
        delete newQuery.chatMode;
        this.$router.replace({ query: newQuery }).catch(() => {});
      }

      // 根据 customSkillId 自动拉取并补全状态
      if (this.isSkillType && this.customSkillId) {
        // 如果是从某模式进入（如转换、导入），则由 fetchCustomSkillInfo 补全后触发启动
        this.fetchCustomSkillInfo(threadId || '', initialChatMode);
        return;
      }

      if (threadId) {
        // 普通 General Agent 路由恢复逻辑
        this.selectConversation(threadId);
      } else {
        this.initNewConversation();
      }
    },
    // 从会话列表设置skill参数
    setSkillParamsByConversation(c_data) {
      if (c_data.skillId) {
        this.customSkillId = c_data.skillId;
      }
      if (c_data.previewId) {
        this.previewId = c_data.previewId;
      }
    },

    startSkillPresetConversation(mode) {
      const promptKey =
        mode === 'convert'
          ? 'generalAgent.skill.convertPrompt'
          : mode === 'import'
            ? 'generalAgent.skill.importPrompt'
            : 'generalAgent.skill.defaultPrompt';
      const content = this.$t(promptKey);
      const userMessage = this.buildUserMessage(content);
      const currentConversation = this.conversationList.find(
        item => item.threadId === this.currentThreadId,
      );

      if (currentConversation) {
        this.$set(currentConversation, 'isSkillConversation', true);
        this.$set(currentConversation, 'skillId', this.customSkillId);
        this.$set(currentConversation, 'previewId', this.previewId);
      } else if (this.currentThreadId) {
        this.conversationList.unshift({
          threadId: this.currentThreadId,
          title: content.slice(0, 50),
          createdAt: new Date().toISOString(),
          isSkillConversation: true,
          skillId: this.customSkillId,
          previewId: this.previewId,
        });
        this.scrollConversationListToTop();
      }

      this.syncConversationRoute(
        this.currentThreadId,
        this.currentConversation || {
          isSkillConversation: true,
          skillId: this.customSkillId,
        },
      );
      this.ensureMessageList(this.currentThreadId);
      this.addUserMessage(this.currentThreadId, content);
      this.$nextTick(() => this.scrollToBottom());

      this.skillChatMode = mode;
      return this.startStreaming(userMessage);
    },

    /**
     * 根据 customSkillId 拉取详情并补全/自愈会话状态
     * @param {string} routeThreadId 路由中带入的 threadId（如果有）
     * @param {string} initialChatMode 预设模式（如 convert/import）
     */
    async fetchCustomSkillInfo(routeThreadId = '', initialChatMode = '') {
      if (!this.customSkillId) return;

      this.isLoadingHistory = true;
      try {
        const res = await getCustomSkillInfo({ skillId: this.customSkillId });

        if (res.code === 0 && res.data) {
          const { previewId, threadId } = res.data;

          // 1. 设置预览 ID
          if (previewId) this.previewId = previewId;

          // 优先级：接口活跃记录 > 路由穿透带入
          let targetThreadId = threadId || routeThreadId;
          let isNewRefresh = false;

          // 若目标对话不存在（如刚转换完还没有任何会话），主动刷新获取新会话
          if (!targetThreadId) {
            targetThreadId = await this.refreshSkillConversation();
            isNewRefresh = true;
          }

          if (targetThreadId) {
            if (initialChatMode) {
              // 处理特殊模式（如 convert/import）的流式启动
              this.currentThreadId = targetThreadId;
              this.resetScrollState();
              this.clearModes();
              this.addMode(MAIN_CHAT_MODES.SKILL_MODE);
              this.hidePanel();
              this.clearMessages(targetThreadId);
              this.loadConfig();

              this.$nextTick(() => {
                this.startSkillPresetConversation(initialChatMode);
              });
            } else if (isNewRefresh) {
              // 刚刷新后的新会话且无 initialChatMode，激活模式并加载配置即可
              this.addMode(MAIN_CHAT_MODES.SKILL_MODE);
              this.loadConfig();
            } else {
              // 非新创建会话，则走正常加载逻辑
              this.selectConversation(targetThreadId);
            }
          }
        } else {
          this.$message.error(
            this.$t('generalAgent.skill.invalidSkill') || '无效技能',
          );
          this.$router.go(-1);
        }
      } catch (error) {
        console.error('fetchCustomSkillInfo error:', error);
        this.$message.error(
          this.$t('generalAgent.skill.loadDetailError') || '获取技能详情失败',
        );
        this.$router.go(-1);
      } finally {
        this.isLoadingHistory = false;
      }
    },

    /**
     * 刷新 Skill 专用对话
     * 当 threadId 为空或对话历史丢失时，为已有 customSkillId 重新创建编辑对话
     * @returns {Promise<string>} 新创建的 threadId
     */
    async refreshSkillConversation() {
      if (!this.customSkillId) {
        console.warn('refreshSkillConversation: customSkillId 为空，无法刷新');
        return '';
      }

      this.isLoadingHistory = true;
      try {
        const res = await refreshGeneralAgentSkillConversation({
          skillId: this.customSkillId,
        });

        if (res.code === 0 && res.data) {
          const { customSkillId, threadId, previewId } = res.data;

          // 更新本地状态
          this.customSkillId = customSkillId || this.customSkillId;
          this.currentThreadId = threadId;
          if (previewId) this.previewId = previewId;

          // 清空消息列表，准备空白会话
          this.clearMessages(threadId);

          // 主动设置模型配置
          await this.syncModelConfigAfterRefresh(threadId);

          // 将新会话加入左侧会话列表
          const newConversation = {
            threadId,
            title: this.$t('generalAgent.index.newConversation'),
            createdAt: new Date().toISOString(),
            isSkillConversation: true,
            skillId: this.customSkillId,
            previewId: this.previewId,
          };
          this.conversationList.unshift(newConversation);
          this.scrollConversationListToTop();
          this.syncConversationRoute(threadId, newConversation);

          return threadId;
        } else {
          this.$message.error(res.msg || '刷新 Skill 对话失败');
          return '';
        }
      } catch (error) {
        console.error('refreshSkillConversation error:', error);
        this.$message.error('刷新 Skill 对话失败，请重试');
        return '';
      } finally {
        this.isLoadingHistory = false;
      }
    },

    /**
     * refresh 后主动同步模型配置到新对话
     * 使用当前已选模型或模型列表中的第一个模型
     */
    async syncModelConfigAfterRefresh(threadId) {
      const modelId = this.selectedModel || this.modelList[0]?.modelId;
      if (!modelId || !threadId) return;

      const modelConfig = this.modelList.find(m => m.modelId === modelId);
      if (!modelConfig) return;

      try {
        await updateGeneralAgentConversationConfig({
          threadId,
          modelConfig: {
            modelId: modelConfig.modelId,
            model: modelConfig.model,
            provider: modelConfig.provider,
            displayName: modelConfig.displayName,
            modelType: modelConfig.modelType || 'llm',
            config: modelConfig.config || {},
          },
        });
      } catch (error) {
        console.error('syncModelConfigAfterRefresh error:', error);
      }
    },
    setupResizeObserver() {
      if (typeof ResizeObserver === 'undefined') return;
      this.resizeObserver = new ResizeObserver(() => {
        this.updateWorkspaceRect();
      });
      // 监听整个页面容器
      const pageEl = this.$el;
      if (pageEl) {
        this.resizeObserver.observe(pageEl);
      }
      // 也监听 sidebar 变化
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
          // 如果 workspace 不可见，使用 mainContent 的右边界
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
          if (!this.selectedModel)
            this.selectedModel = this.modelList[0].modelId;
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
        const res = await getGeneralAgentConversationList({
          pageNo: this.pageNo,
          pageSize: this.pageSize,
        });
        if (res.code === 0) {
          const list = res.data?.list || [];
          if (loadMore) {
            this.conversationList = [...this.conversationList, ...list];
          } else {
            this.conversationList = list;
          }
          // 返回数量不足一页，说明已到底
          if (list.length < this.pageSize) {
            this.hasMoreConversations = false;
          }
        } else if (loadMore) this.pageNo -= 1; // 请求失败时回退页码
      } catch {
        if (loadMore) this.pageNo -= 1;
      } finally {
        this.isLoadingMoreConversations = false;
      }
    },

    /** 将指定会话移到列表顶部，用于历史会话有新消息时按更新时间置顶 */
    moveConversationToTop(threadId) {
      if (!threadId) return;
      const index = this.conversationList.findIndex(
        c => c.threadId === threadId,
      );
      if (index <= 0) return; // 已在顶部或不在列表中
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
      // 距离底部 100vh 时触发加载
      const threshold = window.innerHeight;
      const distanceToBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
      if (distanceToBottom < threshold) {
        this.fetchConversationList(true);
      }
    },

    initNewConversation() {
      const shouldPreserveSkillContext = this.shouldPreserveSkillContext();
      this.currentThreadId = '';
      if (shouldPreserveSkillContext) {
        this.chatType = 'skill';
      } else {
        this.chatType = '';
        this.customSkillId = '';
        this.previewId = '';
      }
      this.syncConversationRoute();
      this.clearMessages('');
      // 重置滚动状态
      this.resetScrollState();
      // 重置模式选择
      this.clearModes();
      if (shouldPreserveSkillContext) {
        this.chatType = 'skill';
      }
      // 关闭工作区面板
      this.hidePanel();
      this.$nextTick(() => {
        if (this.modelList && this.modelList.length > 0) {
          const defaultModel = this.modelList[0];
          this.selectedModel = defaultModel?.modelId || '';
        }
      });
    },

    async createConversationWithTitle(title) {
      if (!this.modelList || this.modelList.length === 0) {
        this.$message.warning(this.$t('generalAgent.error.modelListLoading'));
        return null;
      }

      // 使用用户选择的模型，如果没有选择则使用第一个模型
      const selectedModelConfig = this.selectedModel
        ? this.modelList.find(m => m.modelId === this.selectedModel)
        : this.modelList[0];

      const modelConfig = {
        modelId: selectedModelConfig?.modelId,
        model: selectedModelConfig?.model,
        provider: selectedModelConfig?.provider,
        displayName: selectedModelConfig?.displayName,
        modelType: selectedModelConfig?.modelType,
        config: selectedModelConfig?.config,
      };

      const creationApi = this.isSkillType
        ? createGeneralAgentSkillConversation
        : createGeneralAgentConversation;

      const res = await creationApi({
        title: title || this.$t('generalAgent.index.newConversation'),
        modelConfig,
      });

      if (res.code === 0) {
        const threadId = res.data?.threadId;
        if (threadId) {
          this.currentThreadId = threadId;

          // 保存 skill 相关的额外 ID
          if (this.isSkillType) {
            this.customSkillId = res.data.customSkillId || '';
            this.previewId = res.data.previewId || '';
          }

          const oldMessages = this.messagesMap[''] || [];
          this.$set(this.messagesMap, threadId, oldMessages);
          this.$delete(this.messagesMap, '');

          this.selectedModel = modelConfig.modelId;
          const newConversation = {
            threadId,
            title: title || this.$t('generalAgent.index.newConversation'),
            createdAt: new Date().toISOString(),
            isSkillConversation: this.isSkillType,
            ...(this.isSkillType
              ? {
                  skillId: this.customSkillId,
                  previewId: this.previewId,
                }
              : {}),
          };
          this.conversationList.unshift(newConversation);
          this.scrollConversationListToTop();
          this.syncConversationRoute(threadId, newConversation);
          return threadId;
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

    selectConversation(threadId, conversation = null) {
      if (!threadId) return;

      const targetConversation =
        conversation ||
        this.conversationList.find(item => item.threadId === threadId) ||
        null;

      this.syncConversationRoute(threadId, targetConversation);

      if (this.currentThreadId === threadId) {
        if (targetConversation?.isSkillConversation) {
          this.chatType = 'skill';
          this.setSkillParamsByConversation(targetConversation);
        } else if (targetConversation) {
          this.chatType = '';
          this.customSkillId = '';
          this.previewId = '';
        }
        return;
      }
      // 切换会话时，只切换 currentThreadId，不中止 SSE 流
      // SSE 流会继续在后台运行，切换回来时能继续显示
      this.currentThreadId = threadId;
      this.isLoadingHistory = true;
      this.resetScrollState();
      // 重置模式选择
      this.clearModes();
      // 检查当前会话是否是 skill 会话
      const currentConversation =
        targetConversation || this.currentConversation;
      if (currentConversation?.isSkillConversation) {
        this.chatType = 'skill';
        this.addMode(MAIN_CHAT_MODES.SKILL_MODE);
        this.setSkillParamsByConversation(currentConversation);
      } else if (currentConversation) {
        this.chatType = '';
        this.customSkillId = '';
        this.previewId = '';
      }
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
        const res = await getGeneralAgentConversationDetail({
          threadId: this.currentThreadId,
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
          // Skill 会话且接口正常返回但历史为空（list 为 null）
          // 说明 threadId 对应的对话已被删除，调用 refresh 重建
          if (this.isSkillType && this.customSkillId) {
            this.isLoadingHistory = false;
            await this.refreshSkillConversation();
            return;
          }
          this.isLoadingHistory = false;
        } else {
          this.isLoadingHistory = false;
        }
        this.loadConfig();

        // detail 加载完成后，检查是否有运行中的对话需要重连 connect
        if (this.isSkillType) {
          await this.checkAndResumeSkillPending(this.currentThreadId);
        } else {
          await this.checkAndResumePending(this.currentThreadId);
        }
      } finally {
        this.isLoadingHistory = false;
      }
    },

    // 检查会话是否有运行中的对话，有则走 connect 重连恢复实时流。
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

    // Skill 主会话断线恢复：左侧/中间编辑会话只用 threadId 区分。
    async checkAndResumeSkillPending(threadId) {
      if (!threadId) return;

      const streaming = this.streamingMap[threadId];
      if (streaming && streaming.isStreaming) return;

      let hasPending = false;
      let pendingMessages = [];
      try {
        const res = await getGeneralAgentSkillConversationPending({ threadId });
        hasPending =
          res.code === 0 && res.data?.hasPendingConversation === true;
        pendingMessages = Array.isArray(res.data?.messages)
          ? res.data.messages
          : [];
      } catch (error) {
        console.error('check skill pending conversation failed:', error);
        return;
      }

      if (!hasPending) return;

      this.appendPendingInputMessages(threadId, pendingMessages);

      const { abortController, assistantMessage } =
        this.initStreamState(threadId);
      this.addAssistantMessage(threadId, assistantMessage);
      this.currentStage = 'understanding';
      this.resetScrollState();

      const parser = new SSEEventParser();
      let isUserAborted = false;

      try {
        await connectGeneralAgentSkillConversation({
          threadId,
          onMessage: event => {
            this.handleSSEEvent(event, assistantMessage, parser, threadId);
          },
          onError: error => {
            console.error('Skill reconnect SSE Error:', error);
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
        console.error('Skill reconnect stream error:', error);
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
      const res = await getGeneralAgentConversationConfig({
        threadId: this.currentThreadId,
      });
      if (res.code === 0 && res.data) {
        if (res.data.modelConfig) {
          const modelConfig = res.data.modelConfig;
          this.selectedModel = modelConfig.modelId || modelConfig.model;
        }
      }
    },

    handleKeyDown(e) {
      if (e.shiftKey) return;
      e.preventDefault();
      this.sendMessage({ skipUnsavedCheck: true });
    },

    async beforeEnterSubmit() {
      const content = this.inputMessage.trim();
      if (!content && this.uploadedFiles.length === 0) return false;
      if (this.currentPreviewIsStreaming) return false;

      const currentStreaming = this.streamingMap[this.currentThreadId];
      if (currentStreaming && currentStreaming.isStreaming) return false;

      return this.confirmSkillUnsavedBeforeSend();
    },

    handleModelChange(value) {
      if (!this.currentThreadId) {
        return;
      }
      const selectedModelConfig = this.modelList.find(m => m.modelId === value);
      updateGeneralAgentConversationConfig({
        threadId: this.currentThreadId,
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

    async sendMessage(options = {}) {
      const skipUnsavedCheck = options?.skipUnsavedCheck === true;
      const hasContentOverride = typeof options?.contentOverride === 'string';
      const content = hasContentOverride
        ? options.contentOverride.trim()
        : this.inputMessage.trim();
      const messageFiles = hasContentOverride ? [] : this.uploadedFiles;
      if (!content && messageFiles.length === 0) return;
      if (this.currentPreviewIsStreaming) return;

      // 检查当前会话是否正在流式传输
      const currentStreaming = this.streamingMap[this.currentThreadId];
      if (currentStreaming && currentStreaming.isStreaming) return;

      // 在发送消息前进行本地配置校验
      await this.$refs.configDialog?.fetchToolList();

      // 先执行校验，检查是否有错误（不弹窗）
      const isValid = this.$refs.configDialog?.validateTools();

      if (!isValid) {
        // 有错误，打开弹窗显示错误提示
        this.showConfigDialog = true;
        await this.$nextTick();
        return;
      }

      if (!this.currentThreadId) {
        const title = content.slice(0, 50);
        const threadId = await this.createConversationWithTitle(title);
        if (!threadId) {
          this.$message.error(
            this.$t('generalAgent.error.createConversationFailed'),
          );
          return;
        }
      }

      // 检查配置是否满足条件（在发送消息前）
      const checkRes = await checkGeneralAgentConversationConfig({
        agentId: this.selectedMode?.value ?? '',
        threadId: this.currentThreadId,
      });

      if (checkRes.code === 0 && checkRes.data) {
        const { meet, modelMeet, toolsMeet } = checkRes.data;

        // 如果配置不满足，检查具体哪些项不满足
        if (!meet) {
          // 检查模型是否满足
          if (!modelMeet) {
            this.$message.warning(
              this.$t('generalAgent.error.modelNotAvailable'),
            );
            return;
          }

          // 检查工具是否满足
          if (toolsMeet && Array.isArray(toolsMeet)) {
            const unmetTools = toolsMeet.filter(category => !category.meet);
            if (unmetTools.length > 0) {
              this.showConfigDialog = true;
              this.$nextTick(async () => {
                await this.$refs.configDialog?.fetchToolList();
                this.$refs.configDialog?.validateTools();
              });
              return;
            }
          }
        }
      }

      if (this.currentPreviewIsStreaming) return;
      if (!skipUnsavedCheck) {
        const canSendAfterUnsavedCheck =
          await this.confirmSkillUnsavedBeforeSend();
        if (!canSendAfterUnsavedCheck) {
          return;
        }
      }

      const userMessage = this.buildUserMessage(content, messageFiles);
      this.ensureMessageList(this.currentThreadId);
      this.addUserMessage(this.currentThreadId, content, messageFiles);

      if (!hasContentOverride) {
        this.clearFiles();
        this.$refs.mentionInput?.clear();
      }
      this.$nextTick(() => this.scrollToBottom());

      await this.startStreaming(userMessage);
    },

    async handleSecurityReview() {
      if (this.mainIsStreaming || this.previewIsStreaming) return;
      await this.sendMessage({
        contentOverride: this.$t(
          'generalAgent.skill.skillWorkBench.securityReview.message',
        ),
      });
    },

    async confirmSkillUnsavedBeforeSend() {
      if (!this.isSkillType) return true;

      const skillTabs = this.$refs.skillTabs;
      if (
        !skillTabs ||
        !skillTabs.hasUnsavedFiles ||
        !skillTabs.hasUnsavedFiles()
      ) {
        return true;
      }

      try {
        await this.$confirm(
          '您有未保存的文件内容，继续对话将丢弃这些内容。是否继续？',
          '提示',
          {
            confirmButtonText: '继续对话',
            cancelButtonText: '取消',
            type: 'warning',
          },
        );
      } catch (e) {
        return false;
      }

      const results = skillTabs.discardUnsavedFiles
        ? await skillTabs.discardUnsavedFiles()
        : [];
      const hasFailed = results.some(item => item && item.failed);
      if (hasFailed) {
        this.$message.warning('未保存内容丢弃失败，请保存后重试');
        return false;
      }
      return true;
    },

    async startStreaming(userMessage) {
      if (this.currentPreviewIsStreaming) return;

      if (!this.currentThreadId) {
        this.$message.error(
          this.$t('generalAgent.error.conversationIdNotExist'),
        );
        return;
      }

      const streamingThreadId = this.currentThreadId;
      const agentId = this.selectedMode?.value ?? '';
      const currentMode = this.skillChatMode || 'normal';

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

      const chatParams = {
        threadId: streamingThreadId,
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
          // 使用 mixin 的方法来清理状态，确保全局状态也被更新
          this.cleanupStreamState(streamingThreadId);
          assistantMessage.isStreaming = false;
          // 清理所有 fragments 的 isStreaming 状态
          this.setFragmentsNotStreaming(assistantMessage.fragments);
        },
        signal: abortController.signal,
      };

      try {
        if (this.isSkillType) {
          // Skill 专用对话流
          await chatGeneralAgentSkillConversation({
            ...chatParams,
            customSkillId: this.customSkillId,
            mode: currentMode,
          });
        } else {
          // 普通对话流
          await chatGeneralAgentConversation({
            ...chatParams,
            agentId,
          });
        }
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
        if (this.isSkillType) {
          this.skillChatMode = 'normal';
        }
        // 统一清理（非用户主动中止）；用户中止由 stopStreaming 处理
        this.finalizeStream(streamingThreadId, assistantMessage, isUserAborted);
        if (
          !isUserAborted &&
          this.isSkillType &&
          this.currentThreadId === streamingThreadId
        ) {
          this.requestSkillWorkspaceRefresh();
        }
      }
    },

    requestSkillWorkspaceRefresh() {
      if (!this.isSkillType) return;
      if (!this.customSkillId || !this.previewId) {
        this.pendingSkillWorkspaceRefresh = true;
        return;
      }
      this.$nextTick(() => this.refreshSkillWorkspace());
    },

    refreshSkillWorkspace() {
      if (!this.isSkillType) return;
      if (!this.customSkillId || !this.previewId) {
        this.pendingSkillWorkspaceRefresh = true;
        return;
      }

      const skillTabs = this.$refs.skillTabs;
      if (skillTabs && skillTabs.refreshWorkspace) {
        skillTabs.refreshWorkspace();
        this.pendingSkillWorkspaceRefresh = false;
        return;
      }

      this.pendingSkillWorkspaceRefresh = true;
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

    // 重新生成 - 找到上一条用户消息并重新发送
    handleRegenerate(message) {
      if (this.isStreaming || this.currentPreviewIsStreaming) return;

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

    // Human-in-the-Loop: 用户回答问题
    handleQuestionReply(event) {
      // QuestionBlock 已处理 API 调用和状态更新
      // 此处仅用于日志或后续扩展
      console.log('[GeneralAgent] Question replied:', event);
    },

    // Human-in-the-Loop: 用户拒绝问题
    handleQuestionReject(event) {
      // QuestionBlock 已处理 API 调用和状态更新
      // 此处仅用于日志或后续扩展
      console.log('[GeneralAgent] Question rejected:', event);
    },

    async handleDeleteConversation(item) {
      await this.$confirm(
        this.$t('generalAgent.index.confirmDeleteConversation'),
        this.$t('common.button.tip'),
        {
          type: 'warning',
        },
      );
      const res = await deleteGeneralAgentConversation({
        threadId: item.threadId,
      });
      if (res.code === 0) {
        this.$message.success(this.$t('common.info.delete'));
        if (this.currentThreadId === item.threadId) {
          this.clearMessages(this.currentThreadId);
          this.currentThreadId = '';
          this.chatType = '';
          this.customSkillId = '';
          this.previewId = '';
          this.syncConversationRoute();
          this.hidePanel();
        }
        // 保留已加载的多页数据与滚动位置
        const index = this.conversationList.findIndex(
          c => c.threadId === item.threadId,
        );
        if (index !== -1) {
          this.conversationList.splice(index, 1);
        }
      }
    },

    // 处理停止按钮点击
    handleStopClick() {
      this.stopGeneralConversation(this.currentThreadId);
    },

    // 停止主对话：先通知服务端 cancel，再本地中断 SSE。
    async stopGeneralConversation(threadId) {
      if (threadId) {
        try {
          if (this.isSkillType) {
            await cancelGeneralAgentSkillConversation({ threadId });
          } else {
            await cancelGeneralAgentConversation({ threadId });
          }
        } catch (error) {
          // 取消失败不阻塞本地清理
          console.error('cancel conversation failed:', error);
        }
      }
      this.stopStreaming(threadId);
    },

    // 打开skill导入弹框
    openImportSkillDialog() {
      if (!this.selectedModel) {
        this.$message.warning(this.$t('generalAgent.skill.selectModelFirst'));
        return;
      }
      this.$refs.importSkillDialogRef?.openDialog();
    },
    async handleSkillImportSubmit(formData) {
      const dialogRef = this.$refs.importSkillDialogRef;
      dialogRef.btnLoading = true;
      try {
        const selectedModelConfig =
          this.modelList.find(m => m.modelId === this.selectedModel) ||
          this.modelList[0];
        const modelConfig = {
          modelId: selectedModelConfig?.modelId,
          model: selectedModelConfig?.model,
          provider: selectedModelConfig?.provider,
          displayName: selectedModelConfig?.displayName,
          modelType: selectedModelConfig?.modelType,
          config: selectedModelConfig?.config,
        };

        const { avatar, zipUrl } = formData;
        const res = await importGeneralAgentSkillConversation({
          avatar,
          zipUrl,
          modelConfig,
        });

        if (res.code === 0) {
          const { threadId, customSkillId, previewId } = res.data || {};
          this.$message.success(this.$t('common.message.success'));
          dialogRef.dialogVisible = false;

          if (threadId) {
            this.currentThreadId = threadId;
            this.customSkillId = customSkillId || '';
            this.previewId = previewId || '';
            this.startSkillPresetConversation('import');
          }
        }
      } catch (error) {
        console.error(error);
      } finally {
        dialogRef.btnLoading = false;
      }
    },
    // 打开skill工作空间
    handleViewSkillWorkspace() {
      // skill时runId可为空
      this.handleViewWorkspace({
        threadId: this.currentThreadId,
        runId: '',
        fileCount: 0,
        totalSize: 0,
      });
    },
  },
};
</script>

<style lang="scss" scoped>
@import './styles/variables';

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
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    border-bottom: 1px solid #f0f0f0;

    .new-chat-btn {
      width: 100%;
      border-radius: 12px;
      background: $wga-primary;
      border-color: $wga-primary;
      font-weight: 500;

      &:hover {
        background: $wga-primary-dark;
        border-color: $wga-primary-dark;
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

// 中间区域：主对话（自适应或 40%）
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

// 右侧区域：工作空间 + 预览（固定宽度或 60%）
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

  .header-right {
    display: flex;
    align-items: center;
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

    &.workspace-entry-btn {
      min-width: 80px;
      padding: 0 12px;
      gap: 6px;
      font-size: 13px;
      font-weight: 500;

      &:hover {
        background: $wga-primary;
        color: #fff;
        border-color: $wga-primary;
      }
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

        i {
          font-size: 32px;
          color: #10a37f;
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

  .config-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
    color: $wga-text-secondary;
    background: transparent;
    border: 1px solid $wga-border;
    transition: all 0.2s;

    &:hover {
      background: rgba($wga-primary, 0.08);
      color: $wga-primary;
      border-color: rgba($wga-primary, 0.3);
    }

    i {
      font-size: 16px;
    }

    .el-badge {
      margin-left: 4px;
    }
  }

  // 模式按钮基础样式
  .mode-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border-radius: 20px;
    border: 1px solid $wga-border;
    background: #fff;
    color: $wga-text-secondary;
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s;
    user-select: none;

    .mode-avatar {
      width: 18px;
      height: 18px;
      border-radius: 50%;
      object-fit: cover;
    }

    i {
      font-size: 14px;
    }

    &:hover {
      background: rgba($wga-primary, 0.06);
      border-color: rgba($wga-primary, 0.3);
      color: $wga-primary;
    }

    .el-icon--right {
      margin-left: 0;
      font-size: 12px;
    }
  }

  // 已选模式样式
  .mode-btn-selected {
    background: rgba($wga-primary, 0.08);
    border-color: rgba($wga-primary, 0.3);
    color: $wga-primary;
    cursor: default;

    &:hover {
      background: rgba($wga-primary, 0.12);
      border-color: rgba($wga-primary, 0.4);
    }

    .el-icon-close {
      font-size: 12px;
      cursor: pointer;
      padding: 2px;
      border-radius: 50%;
      transition: all 0.2s;
      margin-left: 4px;

      &:hover {
        background: rgba($wga-primary, 0.2);
      }
    }
  }

  .input-footer {
    text-align: center;
    font-size: 12px;
    color: $wga-text-muted;
    margin-top: 12px;
  }

  // 模式按钮容器
  .mode-buttons {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    margin-top: 16px;
    flex-wrap: wrap;
  }
}

// Workspace 面板过渡动画 — 无预览时从页面右侧滑入
.workspace-slide-enter-active,
.workspace-slide-leave-active {
  transition: all 0.3s ease;
}

.workspace-slide-enter,
.workspace-slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

// Workspace 面板过渡动画 — 有预览时从 right-area 左侧滑入
.workspace-slide-inner-enter-active,
.workspace-slide-inner-leave-active {
  transition:
    width 0.3s ease,
    opacity 0.3s ease;
  overflow: hidden;
}

.workspace-slide-inner-enter,
.workspace-slide-inner-leave-to {
  width: 0 !important;
  opacity: 0;
  overflow: hidden;
}
</style>
