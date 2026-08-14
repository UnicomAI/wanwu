<template>
  <!-- 远景大模型 -->
  <div class="full-content flex">
    <el-main class="scroll">
      <div class="smart-center">
        <!--基础配置回显-->
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
            :chatType="'rag'"
            :sessionStatus="sessionStatus"
            :supportClear="false"
            @clearHistory="clearHistory"
            @refresh="refresh"
            @queryCopy="queryCopy"
            @delConversationQA="handleDelConversationQA"
            @handleRecommendedQuestion="handleRecommendedQuestion"
            :defaultUrl="editForm.avatar.path"
          />
        </div>
        <!--输入框-->
        <div class="center-editable">
          <div v-show="stopBtShow" class="stop-box">
            <span v-show="sessionStatus === 0" class="stop" @click="preStop">
              <img
                class="stop-icon mdl"
                :src="require('@/assets/imgs/stop.png')"
              />
              <span class="mdl">{{ $t('agent.stop') }}</span>
            </span>
          </div>
          <streamInputField
            ref="editable"
            source="perfectReminder"
            :fileTypeArr="fileTypeArr"
            :type="'ragChat'"
            :hasHistory="hasHistory"
            :maxImageSize="maxImageSize"
            :maxPicNum="maxPicNum"
            :visibleUpload="
              editForm.visionsupport === 'support' &&
              editForm.visionConfig.picNum === 1
            "
            @preSend="preSend"
            @setSessionStatus="setSessionStatus"
            @clearHistory="clearHistory"
            @inputHeightChange="handleInputHeightChange"
          />
        </div>
      </div>
    </el-main>
  </div>
</template>

<script>
import streamMessageField from '@/components/stream/streamMessageField';
import streamGreetingField from '@/components/stream/streamGreetingField';
import streamInputField from '@/components/stream/streamInputField';
import sseMethod from '@/mixins/sseMethod';
import { md } from '@/mixins/markdown-it';
import { convertLatexSyntax, parseSub } from '@/utils/util';
import {
  clearRagConversation,
  createRagConversation,
  deleteRagConversation,
  deleteRagDraftConversation,
  getRagDraftConversationDetail,
} from '@/api/rag';
import { mapGetters } from 'vuex';

export default {
  props: {
    chatType: {
      type: String,
      default: '',
    },
    activeConversationId: {
      type: String,
      default: '',
    },
    editForm: {
      type: Object,
      default: null,
    },
    type: {
      type: String,
      default: 'agentChat',
    },
    maxImageSize: {
      type: [Number, String],
      required: false,
      default: null,
    },
    maxPicNum: {
      type: Number,
      required: false,
      default: 1,
    },
  },
  components: {
    streamGreetingField,
    streamMessageField,
    streamInputField,
  },
  mixins: [sseMethod],
  computed: {
    ...mapGetters('app', ['sessionStatus']),
    ...mapGetters('menu', ['basicInfo']),
    ...mapGetters('user', ['commonInfo']),
    hasHistory() {
      return !this.echo;
    },
  },
  data() {
    return {
      echo: true,
      basicForm: {
        avatar: '',
        instructions: '',
        name: '',
        description: '',
      },
      expandForm: {
        starterPrompts: [],
      },
      fileTypeArr: ['image/*'],
      draftHistoryRequestId: 0,
    };
  },
  created() {},
  watch: {
    activeConversationId(value) {
      this.conversationId = value || '';
    },
  },
  methods: {
    // 切换历史会话时终止旧会话的流，避免其运行态残留到新会话界面。
    disconnectCurrentStreamForNavigation() {
      this.stopEventSource();
      this.ctrlAbort = null;
      this._print && this._print.stop();
      this._reasoningPrint && this._reasoningPrint.stop();
      this.setStoreSessionStatus(-1);
      this.stopBtShow = false;
    },
    parseHistoryJson(value) {
      if (!value) return [];
      if (Array.isArray(value)) return value;
      try {
        return JSON.parse(value);
      } catch (error) {
        console.warn(
          '[rag chat] parse conversation history field failed',
          error,
        );
        return [];
      }
    },
    renderHistoryContent(content, index) {
      const normalizedContent = convertLatexSyntax(content || '');
      return md.render(parseSub(normalizedContent, index));
    },
    formatKnowledgeSearchList(list) {
      return list.map(item => ({
        ...item,
        snippet: item.snippet ? md.render(item.snippet) : '',
      }));
    },
    formatQaSearchList(list) {
      return list.map(item => {
        const rawSnippet =
          item.snippet ||
          (item.question || item.answer
            ? `**Q:** ${item.question || ''}\n\n**A:** ${item.answer || ''}`
            : '');
        return {
          ...item,
          title: item.user_kb_name || item.title || '',
          snippet: rawSnippet ? md.render(rawSnippet) : '',
        };
      });
    },
    buildHistoryRagSteps({
      searchList,
      qaSearchList,
      reasoningContent,
      reasoningTimeCost,
      searchListTimeCost,
      qaSearchListTimeCost,
      shouldRenderQaStep,
      shouldRenderKnowledgeStep,
      shouldRenderReasoningStep,
      qaErrMessage,
      createdAt,
    }) {
      const timestamp = Number(createdAt) || 0;
      const hasQaKnowledgeBase = Boolean(
        this.editForm.qaKnowledgeBaseConfig?.knowledgebases?.length,
      );
      const hasKnowledgeBase = Boolean(
        this.editForm.knowledgeBaseConfig?.knowledgebases?.length,
      );
      const buildStep = (type, timeCost, shouldRender, errorMessage = '') => {
        const startAt = Number(timeCost?.start);
        const endAt = Number(timeCost?.end);
        const hasTimeCost =
          timeCost?.start != null &&
          timeCost?.end != null &&
          Number.isFinite(startAt) &&
          Number.isFinite(endAt);
        if (!hasTimeCost && !shouldRender) return null;

        // 与 SSE 的 closeStep 一致：以结束时间减开始时间，保留三位小数。
        const resolvedStartAt = hasTimeCost ? startAt : timestamp;
        const resolvedEndAt = hasTimeCost ? endAt : timestamp;
        return {
          type,
          status: errorMessage ? 'error' : 'done',
          errorMessage,
          startAt: resolvedStartAt,
          endAt: resolvedEndAt,
          duration: hasTimeCost
            ? `${((resolvedEndAt - (resolvedStartAt || resolvedEndAt)) / 1000).toFixed(3)}s`
            : '',
        };
      };

      // 详情接口空检索结果且 qaErrMessage 有值时，按 SSE 的 rag_qa_error 回显。
      const qaErrorMessage = qaSearchList.length ? '' : qaErrMessage || '';
      // qaSearchList 为 null 表示该轮没有进入问答库检索；即使返回了耗时也不展示气泡。
      const qaStep = shouldRenderQaStep
        ? buildStep(
            'qa_search',
            qaSearchListTimeCost,
            qaSearchList.length || hasQaKnowledgeBase || qaErrorMessage,
            qaErrorMessage,
          )
        : null;
      return [
        qaStep,
        shouldRenderKnowledgeStep
          ? buildStep(
              'knowledge_search',
              searchListTimeCost,
              searchList.length || hasKnowledgeBase,
            )
          : null,
        shouldRenderReasoningStep
          ? buildStep('thinking', reasoningTimeCost, reasoningContent)
          : null,
      ].filter(Boolean);
    },
    formatConversationHistory(list) {
      return (list || []).map((item, index) => {
        const rawSearchList = this.parseHistoryJson(item.searchList);
        const rawQaSearchList = this.parseHistoryJson(item.qaSearchList);
        const searchList = this.formatKnowledgeSearchList(rawSearchList);
        const qaSearchList = this.formatQaSearchList(rawQaSearchList);
        const reasoningContent = item.reasoningContent || '';
        // 详情接口使用 fileName/fileSize，消息组件使用 name/size。
        // 归一化后可复用组件已有的图片缩略图与普通文件渲染分支。
        const requestFiles = (item.requestFiles || []).map(file => ({
          ...file,
          name: file.name || file.fileName || '',
          size: file.size ?? file.fileSize ?? 0,
          imgUrl: file.imgUrl || file.fileUrl || '',
        }));
        return {
          detailId: item.detailId || '',
          isHistory: true,
          conversationId: item.conversationId || this.conversationId,
          query: item.prompt || '',
          response: this.renderHistoryContent(item.response, index),
          oriResponse: item.response || '',
          stableReasoningChunks: reasoningContent
            ? [this.renderHistoryContent(reasoningContent, index)]
            : [],
          activeReasoning: '',
          searchList,
          qaSearchList,
          ragSteps: this.buildHistoryRagSteps({
            searchList,
            qaSearchList,
            reasoningContent,
            reasoningTimeCost: item.reasoningTimeCost,
            searchListTimeCost: item.searchListTimeCost,
            qaSearchListTimeCost: item.qaSearchListTimeCost,
            // qaSearchList 为 null 表示该轮未进入问答库检索，连同 qaErrMessage 一并忽略。
            shouldRenderQaStep: item.qaSearchList != null,
            // searchList 为 null 表示该轮未进入知识库检索；即使有耗时也不展示气泡。
            shouldRenderKnowledgeStep: item.searchList != null,
            // reasoningContent 为 null 表示该轮未进入深度思考；即使有耗时也不展示气泡。
            shouldRenderReasoningStep: item.reasoningContent != null,
            qaErrMessage:
              item.qaSearchList === '' ? item.qaErrMessage || '' : '',
            createdAt: item.createdAt,
          }),
          fileList: requestFiles,
          requestFiles,
          // 与 SSE RUN_ERROR 对齐：通用错误文案做标题，后端原始错误放详情。
          error: Boolean(item.errMessage),
          errResponse: item.errMessage ? this.$t('sse.error') : '',
          errorDetail: item.errMessage || '',
          responseLoading: false,
          finish: 1,
          isOpen: true,
          createdAt: item.createdAt,
        };
      });
    },
    loadConversationHistory(list) {
      const history = this.formatConversationHistory(list);
      this.echo = history.length === 0;
      this.$refs['session-com']?.replaceHistory(history);
    },
    async loadDraftConversationHistory() {
      const ragId = this.editForm?.appId;
      const requestId = ++this.draftHistoryRequestId;
      if (!ragId) {
        this.loadConversationHistory([]);
        return;
      }
      try {
        const res = await getRagDraftConversationDetail({
          ragId,
          pageNo: 1,
          pageSize: 1000,
        });
        // 切换草稿或后发请求已开始时，忽略过期响应。
        if (
          requestId !== this.draftHistoryRequestId ||
          ragId !== this.editForm?.appId
        )
          return;
        if (res.code !== 0) {
          this.$message.error(res.msg || this.$t('sse.error'));
          return;
        }
        this.loadConversationHistory(res.data?.list || []);
      } catch (error) {
        if (requestId !== this.draftHistoryRequestId) return;
        console.warn('[rag chat] get draft conversation detail failed', error);
        this.loadConversationHistory([]);
      }
    },
    async clearHistory() {
      if (this.sessionStatus === 0) return;
      try {
        if (this.chatType === 'chat' && this.conversationId) {
          const res = await clearRagConversation({
            ragId: this.editForm.appId,
            conversationId: this.conversationId,
          });
          if (res.code !== 0) {
            this.$message.error(res.msg || this.$t('sse.error'));
            return;
          }
        } else if (this.chatType === 'test' && this.editForm?.appId) {
          const res = await deleteRagDraftConversation({
            ragId: this.editForm.appId,
          });
          if (res.code !== 0) {
            this.$message.error(res.msg || this.$t('sse.error'));
            return;
          }
        }
      } catch (error) {
        console.warn('[rag chat] clear conversation failed', error);
        this.$message.error(this.$t('sse.error'));
        return;
      }
      this.stopBtShow = false;
      this.clearPageHistory();
    },
    handleRecommendedQuestion(question) {
      this.inputVal = question;
      this.preSend(question);
    },
    createConversation() {
      if (this.sessionStatus === 0) return;
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
      this.conversationId = '';
      this.echo = true;
      this.clearHistory();
      this.$emit('conversation-changed', '');
    },
    async preSend(val, fileList, fileInfo) {
      this.inputVal = val || this.$refs['editable'].getPrompt();
      this.fileList = fileList || this.$refs['editable'].getFileList();
      if (!this.inputVal) {
        this.$message.warning(this.$t('agent.inputContent'));
        return;
      }
      if (!this.verifiyFormParams()) {
        return;
      }
      const requestFileInfo = Array.isArray(fileInfo)
        ? fileInfo
        : this.$refs['editable'].getFileIdList();

      if (this.chatType === 'chat' && !this.conversationId) {
        try {
          const res = await createRagConversation({
            ragId: this.editForm.appId,
            prompt: this.inputVal,
          });
          if (res.code !== 0 || !res.data?.conversationId) {
            this.$message.error(res.msg || this.$t('sse.error'));
            return;
          }
          this.conversationId = res.data.conversationId;
          this.$emit('conversation-created', this.conversationId);
        } catch (error) {
          console.warn('[rag chat] create conversation failed', error);
          this.$message.error(this.$t('sse.error'));
          return;
        }
      }

      const sseParams = {
        ragId: this.editForm.appId,
        fileInfo: requestFileInfo,
        question: this.inputVal,
      };
      if (this.chatType === 'chat') {
        sseParams.conversationId = this.conversationId;
      }
      this.setSseParams(sseParams);
      this.doragSend();
      this.echo = false;
    },
    async handleDelConversationQA(detailId) {
      if (!detailId || this.sessionStatus === 0) return;
      const isDraft = this.chatType === 'test';
      if (!isDraft && !this.conversationId) return;
      try {
        const res = isDraft
          ? await deleteRagDraftConversation({
              ragId: this.editForm.appId,
              detailId,
            })
          : await deleteRagConversation({
              ragId: this.editForm.appId,
              conversationId: this.conversationId,
              detailId,
            });
        if (res.code !== 0) {
          this.$message.error(res.msg || this.$t('sse.error'));
          return;
        }
        const sessionCom = this.$refs['session-com'];
        const history = sessionCom?.getSessionData().history || [];
        const nextHistory = history.filter(
          item => item.detailId !== detailId && item.id !== detailId,
        );
        sessionCom?.replaceHistory(nextHistory);
        this.echo = nextHistory.length === 0;
      } catch (error) {
        console.warn('[rag chat] delete conversation detail failed', error);
        this.$message.error(this.$t('sse.error'));
      }
    },
    verifiyFormParams() {
      if (this.chatType === 'chat') return true;
      const { matchType, priorityMatch, rerankModelId } =
        this.editForm.knowledgeBaseConfig.config;
      const qArerankModelId =
        this.editForm.qaKnowledgeBaseConfig.config.rerankModelId;
      const isMixPriorityMatch = matchType === 'mix' && priorityMatch;
      const knowledgebasesLength =
        this.editForm.knowledgeBaseConfig.knowledgebases.length;

      const conditions = [
        {
          check: !this.editForm.modelParams,
          message: this.$t('knowledgeManage.create.selectModel'),
        },
        {
          check:
            knowledgebasesLength > 0
              ? !isMixPriorityMatch && !rerankModelId
              : false,
          message: this.$t('knowledgeManage.hitTest.selectRerankModel'),
        },
        {
          check:
            this.editForm.qaKnowledgeBaseConfig.knowledgebases.length === 0 &&
            this.editForm.knowledgeBaseConfig.knowledgebases.length === 0,
          message: this.$t('app.selectKnowledge'),
        },
        {
          check:
            this.editForm.qaKnowledgeBaseConfig.knowledgebases.length > 0 &&
            !qArerankModelId,
          message: this.$t('knowledgeManage.hitTest.selectQaRerankModel'),
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
      let fileId = this.getFileIdList() || this.fileId;
      this.useSearch = this.$refs['editable'].sendUseSearch();
      this.modelParams = this.$refs['editable'].getModelInfo();
      this.isBigModel = true;
      this.setSseParams({ conversationId: this.conversationId, fileId });
      this.doSend();
      this.echo = false;
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
    // 处理输入框高度变化
    handleInputHeightChange(height) {
      this.$refs['session-com'] &&
        this.$refs['session-com'].setHistoryBoxHeight(height);
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/chat.scss';
</style>
