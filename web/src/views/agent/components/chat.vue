<template>
  <div class="full-content flex">
    <el-main class="scroll" style="padding: 20px 0">
      <div class="smart-center" style="padding: 0">
        <!--开场白设置-->
        <div v-show="echo" class="session rl echo">
          <streamGreetingField
            :editForm="editForm"
            sessionItemWidth="100%"
            @setProloguePrompt="setProloguePrompt"
          />
        </div>
        <!--对话-->
        <div v-show="!echo" class="center-session">
          <streamMessageField
            ref="session-com"
            class="component"
            :chatType="'agent'"
            :sessionStatus="sessionStatus"
            :recommendConfig="recommendConfig"
            :supportClear="false"
            :supportAnswerFeedback="chatType !== 'test'"
            @clearHistory="clearHistory"
            @refresh="refresh"
            @queryCopy="queryCopy"
            @answer-feedback="submitAnswerFeedback"
            @handleRecommendClick="handleRecommendClick"
            @sub-conversion-toggle="onSubConversionToggle"
            @delConversationQA="handleDelConversationQA"
            :defaultUrl="editForm.avatar.path"
          >
            <template #afterContent="{ responseFiles }">
              <div class="product-card-list">
                <ProductFileCard
                  v-for="fileItem in responseFiles"
                  :key="fileItem.fileUrl"
                  :info="fileItem"
                />
              </div>
            </template>
          </streamMessageField>
        </div>
        <!--停止生成-重新生成-->
        <div class="center-editable">
          <div v-show="stopBtShow" class="stop-box">
            <span v-show="sessionStatus === 0" class="stop" @click="preStop">
              <img
                class="stop-icon mdl"
                :src="require('@/assets/imgs/stop.png')"
              />
              <span class="mdl">{{ $t('agent.stop') }}</span>
            </span>
            <span v-show="sessionStatus !== 0" class="stop" @click="refresh">
              <img
                class="stop-icon mdl"
                :src="require('@/assets/imgs/refresh.png')"
              />
              <span class="mdl">{{ $t('agent.refresh') }}</span>
            </span>
          </div>
          <!-- 输入框 -->
          <streamInputField
            ref="editable"
            source="perfectReminder"
            :fileTypeArr="fileTypeArr"
            :type="type"
            :hasHistory="hasHistory"
            :maxImageSize="maxImageSize"
            :maxPicNum="maxPicNum"
            :maxFileNum="maxFileNum"
            :maxFileSize="maxFileSize"
            @preSend="preSend"
            @setSessionStatus="setSessionStatus"
            @clearHistory="handleClearHistory"
            @inputHeightChange="handleInputHeightChange"
          />
          <!-- 版权信息 -->
          <div v-if="appUrlInfo" class="appUrlInfo">
            <span v-if="appUrlInfo.copyrightEnable">
              {{ $t('app.copyright') }}: {{ appUrlInfo.copyright }}
            </span>
            <span v-if="appUrlInfo.privacyPolicyEnable">
              {{ $t('app.privacyPolicy') }}:
              <a
                :href="appUrlInfo.privacyPolicy"
                target="_blank"
                style="color: var(--color)"
              >
                {{ appUrlInfo.privacyPolicy }}
              </a>
            </span>
            <span v-if="appUrlInfo.disclaimerEnable">
              {{ $t('app.disclaimer') }}: {{ appUrlInfo.disclaimer }}
            </span>
          </div>
        </div>
      </div>
    </el-main>
  </div>
</template>

<script>
import streamMessageField from '@/components/stream/streamMessageField';
import streamInputField from '@/components/stream/streamInputField';
import streamGreetingField from '@/components/stream/streamGreetingField';
import { getXClientId, formatReqUrl } from '@/utils/util.js';
import {
  delConversation,
  createConversation,
  getConversationHistory,
  delOpenurlConversation,
  openurlConversation,
  OpenurlConverHistory,
  getRecommendQuestionUrl,
  getConversationDraftHistory,
  getPendingConversation,
  getOpenurlPendingConversation,
  delConversationDraft,
  clearConversation,
  openurlConverDel,
  openurlAgentfeedback,
  agentfeedback,
} from '@/api/agent';
import sseMethod from '@/mixins/sseMethod';
import { transformAgentHistory } from '@/utils/agentHistoryTransformer';
import { mapGetters, mapState } from 'vuex';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import ProductFileCard from './productFileCard.vue';

export default {
  inject: {
    getHeaderConfig: {
      default: () => null,
    },
  },
  props: {
    editForm: {
      type: Object,
      default: null,
    },
    chatType: {
      type: String,
      default: '',
    },
    type: {
      type: String,
      default: 'agentChat',
    },
    appUrlInfo: {
      type: Object,
      default: null,
    },
    assistantId: {
      type: String,
      default: '',
    },
    maxImageSize: {
      type: [Number, String],
      required: false,
      default: null,
    },
    maxPicNum: {
      type: Number,
      required: false,
      default: -1,
    },
    maxFileNum: {
      type: Number,
      required: false,
      default: -1,
    },
    maxFileSize: {
      type: [Number, String],
      required: false,
      default: null,
    },
  },
  components: {
    streamMessageField,
    streamInputField,
    streamGreetingField,
    ProductFileCard,
  },
  mixins: [sseMethod],
  computed: {
    ...mapGetters('app', ['sessionStatus']),
    ...mapGetters('menu', ['basicInfo']),
    ...mapGetters('user', ['commonInfo']),
    ...mapState('user', ['userInfo']),
    hasHistory() {
      return !this.echo;
    },
  },
  data() {
    return {
      echo: true,
      fileTypeArr: ['doc/*', 'image/*'],
      hasDrawer: false,
      drawer: true,
      fileId: [],
      recommendConfig: {
        reqController: new AbortController(),
        list: [],
        loading: false,
      },
      recommendTimer: null,
      draftReconnectRequested: false,
    };
  },
  methods: {
    getOpenurlStreamHeaders() {
      return {
        'X-Client-ID': getXClientId(),
      };
    },
    getOpenurlRequestConfig() {
      return {
        headers: this.getOpenurlStreamHeaders(),
        isOpenUrl: true,
      };
    },
    getStreamAssistantId() {
      if (this.type === 'webChat') {
        return (
          this.assistantId ||
          this.$route.params.id ||
          this.$route.query.id ||
          this.editForm.assistantId
        );
      }
      return this.editForm.assistantId;
    },
    disconnectCurrentStreamForNavigation() {
      if (this.recommendTimer) {
        clearInterval(this.recommendTimer);
        this.recommendTimer = null;
      }
      if (this.recommendConfig.loading) {
        this.recommendConfig.reqController.abort();
        this.recommendConfig.reqController = new AbortController();
        this.recommendConfig.loading = false;
      }
      this.stopEventSource();
      this._print && this._print.stop();
      this._reasoningPrint && this._reasoningPrint.stop();
      this.clearActiveAgentStreamParams && this.clearActiveAgentStreamParams();
      this.setStoreSessionStatus(-1);
      this.stopBtShow = false;
    },
    createConversion() {
      if (this.echo) {
        this.$message({
          type: 'info',
          message: this.$t('app.switchSession'),
          customClass: 'dark-message',
          iconClass: 'none',
          duration: 1500,
        });
        return;
      }
      this.disconnectCurrentStreamForNavigation();
      this.conversationId = '';
      this.echo = true;
      this.clearHistory();
      this.$emit('setHistoryStatus');
    },
    //切换对话
    conversationClick(n) {
      this.disconnectCurrentStreamForNavigation();

      this.$emit('setHistoryStatus');
      this.amswerNum = 0;
      n.active = true;
      this.clearPageHistory();
      this.echo = false;
      this.conversationId = n.conversationId;
      this.getConversationDetail(this.conversationId, true);
    },
    async getConversationDetail(id, loading) {
      loading && this.$refs['session-com'].doLoading();
      let res = null;
      if (this.type === 'agentChat') {
        res = await getConversationHistory({
          assistantId: this.getStreamAssistantId(),
          conversationId: id,
          pageSize: 1000,
          pageNo: 1,
        });
      } else {
        const config = this.getHeaderConfig();
        res = await OpenurlConverHistory(
          { conversationId: id },
          this.getStreamAssistantId(),
          config,
        );
      }

      if (res.code === 0) {
        let history = this.convertHistoryData(res.data.list);

        this.$refs['session-com'].replaceHistory(history);
        this.$nextTick(() => {
          this.connectPendingStream({
            conversationId: id,
            draft: false,
          });
        });
      }
    },
    //删除对话
    async preDelConversation(n) {
      if (this.sessionStatus === 0) {
        return;
      }
      let res = null;
      if (this.type === 'agentChat') {
        res = await delConversation({
          assistantId: this.getStreamAssistantId(),
          conversationId: n.conversationId,
        });
      } else {
        const config = this.getHeaderConfig();
        res = await delOpenurlConversation(
          { conversationId: n.conversationId },
          this.getStreamAssistantId(),
          config,
        );
      }

      if (res.code === 0) {
        this.$emit('conversationDeleted', n);
        if (this.conversationId === n.conversationId) {
          this.conversationId = '';
          this.$refs['session-com'].clearData();
        }
        this.echo = true;
      }
    },
    /*------会话------*/
    async preSend(val, fileList, fileInfo) {
      if (this.recommendTimer) {
        clearInterval(this.recommendTimer);
        this.recommendTimer = null;
      }
      if (this.recommendConfig.loading) {
        this.recommendConfig.reqController.abort();
        this.recommendConfig.reqController = new AbortController();
      }
      this.recommendConfig.list = [];
      this.recommendConfig.loading = false;
      this.inputVal = val || this.$refs['editable'].getPrompt();
      this.fileId = fileInfo || [];
      this.isTestChat = this.chatType === 'test';
      this.fileList = fileList || this.$refs['editable'].getFileList();
      if (!this.inputVal) {
        this.$message.warning(this.$t('agent.inputContent'));
        return;
      }
      if (!this.verifiyFormParams()) {
        return;
      }
      //如果是新会话，先创建
      if (!this.conversationId && this.chatType === 'chat') {
        let res = null;
        if (this.type === 'agentChat') {
          res = await createConversation({
            prompt: this.inputVal,
            assistantId: this.editForm.assistantId,
          });
        } else {
          const config = this.getHeaderConfig();
          res = await openurlConversation(
            { prompt: this.inputVal },
            this.getStreamAssistantId(),
            config,
          );
        }

        if (res.code === 0) {
          this.conversationId = res.data.conversationId;
          this.$emit('reloadList', true);
          this.setParams();
        }
      } else {
        if (this.chatType === 'chat' && this.conversationId) {
          this.$emit('conversationPromoted', {
            conversationId: this.conversationId,
          });
        }
        this.setParams();
      }
    },
    verifiyFormParams() {
      if (this.chatType === 'chat') return true;
      const { matchType, priorityMatch, rerankModelId } =
        this.editForm.knowledgeBaseConfig.config;
      const isMixPriorityMatch = matchType === 'mix' && priorityMatch;
      const knowledgebasesLength =
        this.editForm.knowledgeBaseConfig.knowledgebases.length;
      const conditions = [
        {
          check: !this.editForm.modelParams,
          message: this.$t('agent.form.selectModel'),
        },
        {
          check:
            knowledgebasesLength > 0
              ? !isMixPriorityMatch && !rerankModelId
              : false,
          message: this.$t('knowledgeManage.hitTest.selectRerankModel'),
        },
        {
          check: !this.editForm.prologue,
          message: this.$t('agent.form.inputPrologue'),
        },
      ];
      for (const condition of conditions) {
        if (condition.check) {
          this.$message.warning(condition.message);
          return false;
        }
      }
      return true;
    },
    setParams() {
      const fileInfo = JSON.parse(
        JSON.stringify(this.$refs['editable'].getFileIdList()),
      );
      let fileId = !fileInfo.length
        ? this.fileId
        : fileInfo.map(file => {
            return {
              fileName: file.oldFileName || file.fileName,
              fileSize: file.fileSize,
              fileUrl: file.fileUrl,
            };
          });
      // this.useSearch = this.$refs['editable'].sendUseSearch();
      this.setSseParams({
        conversationId: this.conversationId,
        fileInfo: fileId,
        assistantId: this.getStreamAssistantId(),
      });
      this.doSend();
      this.echo = false;
    },
    /*--右侧提示词--*/
    showDrawer() {
      this.drawer = true;
    },
    hideDrawer() {
      this.drawer = false;
    },
    async getReminderList(cb) {
      let res = await getTemplateList({ pageNo: 0, pageSize: 0, title: '' });
      if (res.code === 0) {
        this.reminderList = res.data.list || [];
        cb && cb();
      }
    },
    reminderClick(n) {
      this.$refs['editable'].setPrompt(n.prompt);
    },
    // 打印结束回调
    onMainPrintEnd() {
      const history = this.$refs['session-com'].getSessionData().history;
      const lastMessage = history[history.length - 1];

      // 只有当最后一条消息存在且 finish 状态为 1 (真正结束) 时才触发推荐
      if (
        lastMessage &&
        lastMessage.finish === 1 &&
        this.editForm.recommendConfig &&
        this.editForm.recommendConfig.recommendEnable &&
        this.editForm.recommendConfig.modelConfig.modelId
      ) {
        this.recommendConfig.list = [];
        this.getRecommendQuestion();
      }
    },
    handleRecommendClick(val) {
      this.preSend(val);
    },
    // 请求推荐问题
    getRecommendQuestion() {
      const history = this.$refs['session-com'].getSessionData().history;
      const lastUserMessage = history
        .slice()
        .reverse()
        .find(item => item.query);
      const query = lastUserMessage ? lastUserMessage.query : '';
      const signal = this.recommendConfig.reqController.signal;

      class RetriableError extends Error {}
      class FatalError extends Error {
        constructor(message = '', detail = {}) {
          super(message);
          this.name = 'FatalError';
          Object.assign(this, detail);
        }
      }

      const params = {
        query: query,
        assistantId: this.getStreamAssistantId(),
        conversationId: this.conversationId,
        trial: this.chatType === 'test' ? true : false,
      };

      this.recommendConfig.loading = true;

      let currentBuffer = ''; // 用于暂存当前正在拼接的问题片段
      let baseList = []; // 用于存储已经确认完成的问题
      let contentQueue = []; // 字符队列，用于模拟打字机效果
      let isFinished = false; // 标记 SSE 是否已结束接收

      const getResponseErrorInfo = async response => {
        const fallbackMessage =
          response.statusText || `HTTP ${response.status}`;

        try {
          const errorData = await response.clone().json();
          return {
            message:
              errorData.msg ||
              errorData.message ||
              errorData.error ||
              fallbackMessage,
            data: errorData,
          };
        } catch (jsonError) {
          try {
            const text = await response.clone().text();
            return {
              message: text || fallbackMessage,
              data: text,
            };
          } catch (textError) {
            return {
              message: fallbackMessage,
              data: null,
            };
          }
        }
      };
      const normalizeFatalError = error => {
        if (error instanceof FatalError) {
          return error;
        }

        const message =
          (error && (error.msg || error.message || error.error)) ||
          (typeof error === 'string' ? error : '') ||
          '连接错误';

        return new FatalError(message, {
          status: error && error.status,
          statusText: error && error.statusText,
          data: error && error.data !== undefined ? error.data : error,
          cause: error,
        });
      };
      if (this.recommendTimer) {
        clearInterval(this.recommendTimer);
        this.recommendTimer = null;
      }

      // 核心处理逻辑：从队列中取字符并更新 UI
      const processQueue = () => {
        if (contentQueue.length > 0) {
          const item = contentQueue.shift();
          const { char, type } = item;
          currentBuffer += char;
          const delimiter = currentBuffer.includes('\\n')
            ? '\\n'
            : currentBuffer.includes('\n')
              ? '\n'
              : null;

          if (delimiter) {
            // 使用分隔符拆分内容
            const parts = currentBuffer.split(delimiter);
            // 除了最后一部分外，前面的部分都是已经接收完整的
            for (let i = 0; i < parts.length - 1; i++) {
              const finishedContent = parts[i].trim();
              if (finishedContent) {
                baseList.push({
                  content: finishedContent,
                  type: type,
                });
              }
            }
            // 将最后一部分（可能还不完整）留回缓冲区
            currentBuffer = parts[parts.length - 1];
          }

          // 实时渲染展示列表（已完成列表 + 当前正在输入的问题）
          const displayList = [...baseList];
          if (currentBuffer.trim()) {
            displayList.push({
              content: currentBuffer.trim(),
              type: type,
            });
          }
          this.recommendConfig.list = displayList;
        } else if (isFinished) {
          // 如果数据接收完毕且队列已空，执行最后收尾
          clearInterval(this.recommendTimer);
          this.recommendTimer = null;

          // 处理缓冲区剩余的内容
          const finalContent = currentBuffer.trim();
          if (finalContent) {
            // 获取最后一个元素的类型，如果没有则默认为 answer
            const lastType =
              this.recommendConfig.list.length > 0
                ? this.recommendConfig.list[
                    this.recommendConfig.list.length - 1
                  ].type
                : 'answer';

            baseList.push({
              content: finalContent,
              type: lastType,
            });
          }

          this.recommendConfig.list = [...baseList];
          this.recommendConfig.loading = false;
          currentBuffer = '';
        }
      };

      const api = getRecommendQuestionUrl(this.type, params.assistantId);
      let headers = {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + this.token,
        'x-user-id': this.userInfo.uid,
        'x-org-id': this.userInfo.orgId,
      };

      // webchat场景使用不同的请求配置
      if (this.type === 'webChat') {
        headers = {
          'X-Client-ID': this.getHeaderConfig
            ? this.getHeaderConfig().headers['X-Client-ID']
            : '',
        };
        delete params.assistantId;
        delete params.trial;
      }

      const _this = this;
      fetchEventSource(formatReqUrl(api), {
        method: 'POST',
        signal,
        openWhenHidden: true,
        headers,
        body: JSON.stringify(params),
        async onopen(response) {
          const contentType = response.headers.get('content-type') || '';
          if (response.ok && contentType.includes('text/event-stream')) {
            console.log('连接成功，开始获取数据...');
          } else if (
            response.status >= 400 &&
            response.status < 500 &&
            response.status !== 429
          ) {
            _this.recommendConfig.loading = false;
            const errorInfo = await getResponseErrorInfo(response);
            throw new FatalError(errorInfo.message, {
              status: response.status,
              statusText: response.statusText,
              data: errorInfo.data,
            });
          } else {
            throw new RetriableError();
          }
        },

        onmessage: msgData => {
          if (msgData.data && msgData.data !== '[DONE]') {
            try {
              const _data = JSON.parse(msgData.data);
              const choice = _data.choices && _data.choices[0];
              if (choice) {
                const content = choice.delta && choice.delta.content;
                const contentType = choice.contentType || 'answer';

                if (content) {
                  // 将内容拆分为带类型信息的字符对象存入队列
                  const items = content.split('').map(char => ({
                    char,
                    type: contentType,
                  }));
                  contentQueue.push(...items);

                  if (!this.recommendTimer) {
                    this.recommendTimer = setInterval(processQueue, 30);
                  }
                }

                if (['stop', 'accidentStop'].includes(choice.finish_reason)) {
                  isFinished = true;
                  if (!this.recommendTimer) {
                    processQueue();
                  }
                }
              }
            } catch (e) {
              console.error('解析推荐问题失败', e);
            }
          }
          if (msgData.event === 'FatalError') {
            isFinished = true;
            throw new FatalError(msgData.data);
          }
        },
        async onclose() {
          console.log('连接关闭...');
          isFinished = true;
          if (!_this.recommendTimer) {
            processQueue();
          }
          return false;
        },
        onerror(event) {
          isFinished = true;
          _this.recommendConfig.loading = false;
          throw normalizeFatalError(event);
        },
      });
    },
    // 转换智能体历史记录数据
    convertHistoryData(data) {
      return transformAgentHistory(data);
    },
    // 获取草稿页会话历史
    async _getConversationDraftHistory() {
      this.echo = false;
      this.$refs['session-com'].doLoading();
      try {
        const res = await getConversationDraftHistory({
          assistantId: this.editForm.assistantId,
          pageSize: 30,
          pageNo: 1,
        });

        if (res.code === 0) {
          let history = this.convertHistoryData(res.data.list);
          if (!history.length) {
            this.echo = true;
          }
          this.$refs['session-com'].replaceHistory(history);
          this.$nextTick(() => {
            this.connectPendingStream({ draft: true });
          });
        }
      } catch (error) {
        this.$refs['session-com'].stopLoading();
        this.echo = true;
        throw error;
      }
    },
    // 清空会话
    // 格式化历史会话数据
    formatUnderwayHistory(data) {
      if (!data) return null;

      const base = this.convertHistoryData([data])[0] || {};
      const fileList =
        data.requestFiles ||
        data.fileList ||
        data.fileInfo ||
        base.fileList ||
        [];

      return {
        ...base,
        query: data.prompt || data.query || base.query || '',
        conversationId: data.conversationId || base.conversationId || '',
        detailId: data.detailId || base.detailId || data.id || '',
        fileList,
        requestFiles: fileList,
        requestFileUrls: data.requestFileUrls || [],
        response: '',
        oriResponse: '',
        pendingResponse: '',
        responseLoading: true,
        pending: true,
        finish: 0,
        searchList: [],
        subConversions: [],
        messageSequence: [],
        responseFiles: [],
        gen_file_url_list: [],
        isOpen: true,
        toolText: this.$t('agent.tooled'),
        thinkText: this.$t('agent.thinked'),
        showScrollBtn: null,
      };
    },
    // 重连sse
    async connectPendingStream({ conversationId = '', draft = false } = {}) {
      const assistantId = this.getStreamAssistantId();
      if (
        !['agentChat', 'webChat'].includes(this.type) ||
        (draft && this.draftReconnectRequested) ||
        !assistantId
      ) {
        return;
      }

      if (!draft && !conversationId) return;
      if (draft) {
        this.draftReconnectRequested = true;
      }

      try {
        const requestData = {
          assistantId,
          draft,
        };
        if (!draft) {
          requestData.conversationId = conversationId;
        }

        const requestConfig =
          this.type === 'webChat' ? this.getOpenurlRequestConfig() : {};

        const res =
          this.type === 'webChat'
            ? await getOpenurlPendingConversation(
                requestData.assistantId,
                {
                  conversationId: requestData.conversationId,
                },
                requestConfig,
              )
            : await getPendingConversation(requestData, requestConfig);
        if (
          res.code !== 0 ||
          !res.data ||
          res.data.hasPendingConversation !== true
        ) {
          return;
        }

        const underwayHistory = this.formatUnderwayHistory({
          ...res.data,
          conversationId: res.data.conversationId || conversationId,
        });
        if (!underwayHistory || !underwayHistory.conversationId) return;

        const sessionCom = this.$refs['session-com'];
        if (!sessionCom) return;

        const history = sessionCom.getSessionData().history || [];
        const lastIndex = history.length;
        sessionCom.pushHistory(underwayHistory);
        this.echo = false;

        this.connectEventSource({
          assistantId,
          conversationId: underwayHistory.conversationId,
          query: underwayHistory.query,
          lastIndex,
          fileList: underwayHistory.fileList,
          headers:
            this.type === 'webChat' ? this.getOpenurlStreamHeaders() : null,
        });
      } catch (error) {
        console.warn('[agent chat] get underway conversation failed', error);
      }
    },
    connectDraftStream() {
      return this.connectPendingStream({ draft: true });
    },
    // 清空会话
    async handleClearHistory() {
      const history = this.$refs['session-com'].session_data.history;
      if (!history || !history.length) return;
      if (this.chatType === 'test') {
        const res = await delConversationDraft({
          assistantId: this.editForm.assistantId,
        });
        if (res.code === 0) {
          this.clearHistory();
        }
      } else if (this.chatType === 'chat') {
        let res = null;
        if (this.type === 'webChat') {
          res = await openurlConverDel(
            {
              conversationId: this.conversationId,
            },
            this.getStreamAssistantId(),
            this.getHeaderConfig(),
          );
        } else {
          res = await clearConversation({
            assistantId: this.getStreamAssistantId(),
            conversationId: this.conversationId,
          });
        }
        if (res.code === 0) {
          this.clearHistory();
        }
      } else {
        this.clearHistory();
      }
    },
    // 处理输入框高度变化
    handleInputHeightChange(height) {
      this.$refs['session-com'] &&
        this.$refs['session-com'].setHistoryBoxHeight(height);
    },
    // 处理子会话手动切换
    onSubConversionToggle({ id, isOpen }) {
      this.setSubConversionUserToggle(id, isOpen);
    },
    // 删除单条会话(sse使用detailId，历史记录使用id)
    async handleDelConversationQA(qaId) {
      let res = null;
      if (this.chatType === 'test') {
        res = await delConversationDraft({
          assistantId: this.editForm.assistantId,
          detailId: qaId,
        });
      } else if (this.chatType === 'chat') {
        if (this.type === 'webChat') {
          res = await openurlConverDel(
            {
              conversationId: this.conversationId,
              detailId: qaId,
            },
            this.getStreamAssistantId(),
            this.getHeaderConfig(),
          );
        } else {
          res = await clearConversation({
            assistantId: this.getStreamAssistantId(),
            conversationId: this.conversationId,
            detailId: qaId,
          });
        }
      } else {
        throw new Error('不支持的会话类型');
      }

      if (res && res.code === 0) {
        const sessionCom = this.$refs['session-com'];
        if (!sessionCom) return;

        const history = sessionCom.getSessionData().history || [];

        if (history.length > 0) {
          const lastItem = history[history.length - 1];
          if (lastItem.detailId === qaId || lastItem.id === qaId) {
            this.stopBtShow = false;
          }
        }

        const nextHistory = history.filter(
          item => item.detailId !== qaId && item.id !== qaId,
        );

        this.echo = !nextHistory.length;
        sessionCom.replaceHistory(nextHistory);
      }
    },
    // 提交答案反馈
    async submitAnswerFeedback(
      { feedbackType, feedbackContent = '', conversationId, detailId },
      onSuccess,
    ) {
      const feedbackTypeMap = { up: 1, down: 2 };
      const type = feedbackType === 0 ? 0 : feedbackTypeMap[feedbackType];
      const currentAssistantId = this.getStreamAssistantId();
      if (
        ![0, 1, 2].includes(type) ||
        !detailId ||
        !(conversationId || this.conversationId) ||
        !currentAssistantId
      )
        return;

      const feedbackData = {
        conversationId: conversationId || this.conversationId,
        detailId,
        feedbackContent,
        feedbackType: type,
      };

      try {
        const res =
          this.type === 'webChat'
            ? await openurlAgentfeedback(
                currentAssistantId,
                feedbackData,
                this.getHeaderConfig(),
              )
            : await agentfeedback({
                ...feedbackData,
                assistantId: currentAssistantId,
              });
        if (!res || res.code !== 0) {
          this.$message.error(res?.msg || this.$t('common.message.error'));
          return;
        }
        if (typeof onSuccess === 'function') {
          onSuccess();
        }
      } catch (error) {
        this.$message.error(error?.message || this.$t('common.message.error'));
      }
    },
  },
  mounted() {
    // 获取草稿页会话历史(延迟请求避免阻塞其他接口)
    if (this.chatType === 'test') {
      setTimeout(() => {
        this._getConversationDraftHistory();
      }, 1000);
    }
  },
  beforeDestroy() {
    this.stopEventSource();
    if (this.recommendTimer) {
      clearInterval(this.recommendTimer);
      this.recommendTimer = null;
    }
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/chat.scss';
.appUrlInfo {
  margin-top: 10px;
  display: flex;
  justify-content: center;
  span {
    cursor: pointer;
    color: #bbb;
    margin-right: 15px;
  }
}
.product-card-list {
  display: flex;
  flex-wrap: wrap;
  margin-top: 10px;
  gap: 10px;
  flex: 1;
}
</style>
