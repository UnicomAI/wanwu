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
            :visibleUpload="editForm.visionsupport === 'support'"
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
import { mapGetters } from 'vuex';

export default {
  props: {
    chatType: {
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
    };
  },
  created() {},
  methods: {
    handleRecommendedQuestion(question) {
      this.inputVal = question;
      this.preSend(question);
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
      // this.setParams()
      this.setSseParams({
        ragId: this.editForm.appId,
        fileInfo: requestFileInfo,
        question: this.inputVal,
        // RAG 配置里的最长上下文轮数，取 知识库 / 问答库 两处配置的较大值。
        // 0 表示不携带历史；sendRagEventSource 会据此裁剪 history 并在发请求前剔除该字段。
        maxHistory: Math.max(
          this.editForm.knowledgeBaseConfig?.config?.maxHistory || 0,
          this.editForm.qaKnowledgeBaseConfig?.config?.maxHistory || 0,
        ),
      });
      this.doragSend();
      this.echo = false;
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
