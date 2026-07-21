<template>
  <div
    :class="[mode]"
    class="page-wrapper wrap-fullheight conversation-pane right-page-content-body"
  >
    <div class="page-title">
      <img class="page-title-img" src="@/assets/imgs/model.svg" alt="" />
      <span class="page-title-name">{{ title }}</span>
    </div>
    <div class="page-content">
      <slot name="nav"></slot>
      <div class="chat-wrapper">
        <div class="chat-header">
          <template v-if="isModelComparison">
            <el-button
              size="mini"
              type="primary"
              @click="beforeOpenAddModelDialog"
            >
              {{ $t('modelExprience.addModel') }}
            </el-button>
            <el-button size="mini" type="primary" @click="handleExitComparison">
              {{ $t('modelExprience.exitModelComparison') }}
            </el-button>
          </template>
          <el-button
            v-else
            size="mini"
            type="primary"
            @click="openAddModelDialog"
          >
            {{ $t('modelExprience.modelComparison') }}
          </el-button>
        </div>
        <div class="chat-content">
          <ModelChatPane
            v-for="chat in modelChatList"
            :key="chat.modelId"
            :ref="el => setSessionRef(chat.modelId, el)"
            :modelId="chat.modelId"
            :sessionId="chat.sessionId"
            :modelExperienceId="modelExperienceId"
            :modelDetail="chat.modelDetail"
            :modelSetting="chat.modelSetting"
            :pending="chat.pending"
            :supportDelete="modelChatList.length > 1"
            :supportStop="modelChatList.length > 1"
            :modelSessionStatus.sync="chat.sessionStatus"
            @modelSet="openModelSetDialog(chat.modelId, chat.modelSetting)"
            @modelReplace="openReplaceModelDialog(chat.modelId)"
            @modelDelete="handleModelDelete(chat.modelId)"
            @queryCopy="handleSetQuery"
          />
        </div>
        <div class="chat-footer">
          <div v-show="isGenerating" class="stop-box">
            <span class="stop" @click="preStop">
              <img
                class="stop-icon mdl"
                :src="require('@/assets/imgs/stop.png')"
              />
              <span class="mdl">{{ $t('modelExprience.stop') }}</span>
            </span>
          </div>
          <div v-show="!isInit && !isGenerating" class="stop-box">
            <span class="stop" @click="refresh">
              <img
                class="stop-icon mdl"
                :src="require('@/assets/imgs/refresh.png')"
              />
              <span class="mdl">{{ $t('agent.refresh') }}</span>
            </span>
          </div>
          <StreamInputField
            ref="editable"
            type="webChat"
            :fileLimit="100"
            :supportReminder="true"
            :visibleClearHistory="false"
            :visibleUpload="visibleUpload"
            :maxImageSize="maxImageSize"
            :maxPicNum="currentMaxPicNum"
            :fileTypeArr="['image/*']"
            @preSend="preSend"
          />
        </div>
      </div>
    </div>
    <!-- 模型设置 -->
    <ModelSetDialog
      @setModelSet="setModelSet"
      ref="modelSetDialog"
      :modelform="editModel.modelSetting"
      :limitMaxTokens="editModel.limitMaxTokens"
      :append-to-body="true"
    />
    <!-- 模型选择 -->
    <SelectModelDialog
      ref="selectModelDialog"
      :modelOptions="modelOptions"
      @submit="handleModelSelect"
    />
  </div>
</template>
<script>
import { generateChatConfig } from './helper';
import StreamInputField from '@/components/stream/streamInputField.vue';
import ModelChatPane from './components/ModelChatPane.vue';
import ModelSetDialog from '@/views/agent/components/modelSetDialog.vue';
import SelectModelDialog from './components/SelectModelDialog.vue';
import { createAndUpdateChat, getExprienceDetail } from '@/api/modelExprience';
import { getModelDetail } from '@/api/modelAccess';
import { md } from '@/mixins/markdown-it';
export default {
  name: 'ConversationPane',
  props: {
    mode: {
      type: String,
      default: 'modelExprience', // modelComparison
    },
    modelExperienceId: {
      // 模型体验id（历史记录列表的id）
      type: String,
      default: '0',
    },
    comparisonIds: {
      // 模型对比模式下，对比的模型id列表
      type: Array,
      default: () => [],
    },
    modelOptions: {
      // 可选模型列表
      type: Array,
      default: () => [],
    },
  },
  components: {
    ModelSetDialog,
    StreamInputField,
    ModelChatPane,
    SelectModelDialog,
  },
  data() {
    return {
      loading: false,
      sessionRefs: {},
      modelChatList: [],
      editModel: {
        modelId: '',
        modelSetting: {},
        limitMaxTokens: 4096,
      },
      isChatGenerating: false, // 是否正在创建会话
      isInit: true,
    };
  },
  watch: {
    modelOptions: {
      handler() {
        if (this.isModelExprience && !this.modelChatList.length) {
          this.initPage();
        }
      },
      deep: false,
      immediate: false,
    },
  },
  computed: {
    visibleUpload() {
      return this.modelChatList.every(
        item => item?.modelDetail?.config?.visionSupport === 'support',
      );
    },
    currentMaxPicNum() {
      return this.visibleUpload ? 3 : -1;
    },
    maxImageSize() {
      const sizes = this.modelChatList
        .map(item => item?.modelDetail?.config?.maxImageSize)
        .filter(v => v); // 过滤出有限制的 size，如果是 0：默认不限制尺寸
      return sizes.length > 0 ? Math.min(...sizes) : null;
    },
    title() {
      return this.mode === 'modelExprience'
        ? this.$t('modelExprience.title')
        : this.$t('modelExprience.modelComparison');
    },
    isGenerating() {
      return this.modelChatList.some(item => item.sessionStatus === 0);
    },
    isModelComparison() {
      return this.mode === 'modelComparison';
    },
    isModelExprience() {
      return this.mode === 'modelExprience';
    },
  },
  mounted() {
    this.initPage();
  },
  methods: {
    initPage() {
      const comparisonIds = this.$route.query.comparisonIds || '';
      const modelId =
        this.$route.query.modelId || this.modelOptions[0]?.modelId;
      if (this.$route.query.modelId || comparisonIds) {
        this.$router.replace({ query: {} });
      }
      if (this.isModelExprience) {
        // 页面初始化的时候，如果url上携带了comparisonIds，则需要同步打开模型对比界面
        comparisonIds &&
          this.$emit('openModelComparison', comparisonIds.split(','));
        if (!modelId) {
          return;
        }
      }
      this.modelChatList = [];
      this.$nextTick(() => {
        (this.isModelExprience ? [modelId] : this.comparisonIds)
          .filter(id => !!id)
          .forEach(id => {
            this.modelChatList.push(
              generateChatConfig({
                modelId: id,
              }),
            );
            this.fetchModelDetail(id).then(result => {
              this.modelChatList.some((item, index) => {
                if (item.modelId === id) {
                  this.isModelComparison && (result.sessionId = this.$guid());
                  this.$set(
                    this.modelChatList,
                    index,
                    generateChatConfig(result),
                  );
                  return true;
                }
                return false;
              });
            });
          });
      });
    },
    // 获取历史对话记录详情（暴露父级组件调用）
    async initConversation(chatRecord) {
      this.modelChatList = [];
      await this.$nextTick();
      this.modelChatList = [
        generateChatConfig({
          sessionId: chatRecord.sessionId,
          modelExperienceId: chatRecord.id,
          modelId: chatRecord.modelId,
          title: chatRecord.title,
          modelSetting: JSON.parse(chatRecord.modelSetting),
        }),
      ];
      this.fetchExprienceDetail(chatRecord.id, chatRecord.modelId);
    },
    // 请求历史记录的问答对话信息
    fetchExprienceDetail(modelExperienceId, modelId) {
      this.isInit = true;
      getExprienceDetail({
        modelId,
        modelExperienceId,
      }).then(res => {
        if (res.code === 0) {
          this.sessionRefs = {};
          const resultList = this.assembleHistoryQnA(res);
          this.$nextTick(() => {
            Object.entries(this.sessionRefs).forEach(([key, sessionRef]) => {
              if (modelId !== key || !sessionRef) {
                return;
              }
              sessionRef.initHistoryList(resultList);
              if (
                this.isModelExprience &&
                modelExperienceId &&
                modelExperienceId !== '0' &&
                sessionRef.reconnectModelExperienceStream
              ) {
                sessionRef.reconnectModelExperienceStream();
              }
            });
          });
          this.fetchModelDetail(modelId).then(result => {
            this.modelChatList.some((item, index) => {
              if (item.modelId === modelId) {
                this.modelChatList.splice(index, 1, { ...item, ...result });
                return true;
              }
              return false;
            });
          });
        }
      });
    },
    // 获取模型详情
    fetchModelDetail(modelId) {
      return new Promise((resolve, reject) => {
        getModelDetail({ modelId }).then(res => {
          if (res.code === 0) {
            resolve({
              modelId,
              model: res.data.model,
              modelType: res.data.modelType,
              modelDetail: res.data,
              pending: false,
            });
            return;
          }
          reject();
        });
      });
    },
    // 组装历史记录中的问答对话内容
    assembleHistoryQnA(res) {
      const resultList = [];
      if (Array.isArray(res.data.list)) {
        for (let i = 0; i < res.data.list.length; i++) {
          let item = res.data.list[i];
          let result = null;
          if (item.role === 'user') {
            result = {
              query: item.originalContent,
              fileList: item.requestFiles || [],
              thinkText: this.$t('modelExprience.thinking'),
              searchList: [],
              isOpen: true,
              qa_type: 0, // 为了组件复用，前端加了标识
            };
            // 因为问题和答案在数组的相邻两个对象中，所以这里需要i++，以方便读取答案元素
            i++;
            if (res.data.list[i]) {
              if (res.data.list[i].role !== 'user') {
                let item = res.data.list[i];
                result = {
                  ...(result || {}),
                  response: md.render(item.originalContent),
                  stableReasoningChunks: item.reasoningContent
                    ? [md.render(item.reasoningContent.replace(/\\n/g, '\n'))]
                    : [],
                  activeReasoning: '',
                  oriResponse: item.originalContent,
                };
              } else if (res.data.list[i].role === 'user') {
                i--;
              }
            }
          }
          result && resultList.push(result);
        }
      }
      return resultList;
    },
    // 打开模型配置弹窗
    openModelSetDialog(modelId, modelSetting) {
      this.editModel.modelId = modelId;
      // 从模型详情中获取 maxTokens 限制
      const chatModel = this.modelChatList.find(
        item => item.modelId === modelId,
      );
      const config = chatModel?.modelDetail?.config || {};
      const maxTokens = config.maxTokens;
      this.editModel.limitMaxTokens =
        maxTokens && maxTokens > 0 ? maxTokens : 4096;
      Object.keys(modelSetting).forEach(key => {
        if (key === 'thinkingSupport') {
          if (modelSetting[key] === 'support') {
            this.editModel.modelSetting.thinkingEnable = true;
          } else {
            delete this.editModel.modelSetting.thinkingEnable;
          }
        } else {
          this.editModel.modelSetting[key] = modelSetting[key];
        }
      });
      delete this.editModel.modelSetting.thinkingSupport;
      this.$refs.modelSetDialog.showDialog();
    },
    // 更新模型配置
    async setModelSet(val) {
      const chatRecorder = this.modelChatList.find(
        item => item.modelId === this.editModel.modelId,
      );
      if (!chatRecorder) {
        return;
      }
      chatRecorder.modelSetting = { ...val }; // 暂不做响应式处理
      // 无模型体验id || 在模型对比模式下：不需要调接口更新对话的模型配置
      if (!this.modelExperienceId || this.mode !== 'modelExprience') {
        return;
      }
      await createAndUpdateChat({
        id: this.modelExperienceId,
        modelId: chatRecorder.modelId,
        modelType: chatRecorder.modelType,
        modelSetting: chatRecorder.modelSetting,
        sessionId: chatRecorder.sessionId,
        title: chatRecorder.title.substring(0, 12),
      });
      this.$emit('refreshHistoryList');
    },
    handleModelDelete(modelId) {
      this.modelChatList.some((item, index) => {
        if (item.modelId === modelId) {
          this.modelChatList.splice(index, 1);
          delete this.sessionRefs[modelId];
          return true;
        }
        return false;
      });
    },
    openReplaceModelDialog(modelId) {
      this.$refs.selectModelDialog.openDialog({
        mode: 'replace',
        current: [modelId],
        disabledSelected: this.modelChatList
          .map(item => item.modelId)
          .filter(id => id !== modelId),
      });
    },
    beforeOpenAddModelDialog() {
      if (this.modelChatList.length >= 4) {
        this.$message.warning(
          this.$t('modelExprience.tip.maxSelectModel').replace('@', 4),
        );
        return;
      }
      this.openAddModelDialog();
    },
    openAddModelDialog() {
      const modelIds = this.modelChatList.map(item => item.modelId);
      if (this.mode === 'modelComparison') {
        // 模型对比模式下，该操作为新增对比模型，之前已经选择的模型禁用，在模型选择弹窗中不允许操作了。
        this.$refs.selectModelDialog.openDialog({
          mode: 'add',
          current: [],
          disabledSelected: modelIds,
        });
      } else {
        // 模型体验模式下，该操作为选择所有的需要对比的模型，应该是所有的模型都可以选择，且模型选中了当前正在体验的模型
        this.$refs.selectModelDialog.openDialog({
          mode: 'add',
          current: modelIds,
          disabledSelected: [],
        });
      }
    },
    openSelectModelDialog() {
      this.$refs.selectModelDialog.openDialog({
        mode: 'create',
      });
    },
    async handleModelSelect(ids, config) {
      // 模型体验模式下，新建会话时选择模型
      if (config.mode === 'create') {
        this.isInit = true;
        this.modelChatList = [];
        await this.$nextTick();
        this.modelChatList = [
          generateChatConfig({
            modelId: ids[0],
          }),
        ];
        this.fetchModelDetail(ids[0]).then(result => {
          this.modelChatList = [
            generateChatConfig({
              modelId: ids[0],
              ...result,
            }),
          ];
        });
        return;
      }

      // 模型体验模式&&不为替换（即选择模型进行模型对比的场景）
      if (this.mode === 'modelExprience' && config.mode !== 'replace') {
        this.$emit('openModelComparison', ids);
        return;
      }

      // 替换模型（包含模型体验和模型对比）
      if (config.mode === 'replace') {
        const oldModelId = config.originalModelIds[0];
        const newModelId = ids[0];
        const index = this.modelChatList.findIndex(
          item => item.modelId === oldModelId,
        );
        const oldRecorder = this.modelChatList[index];
        const newRecorder = generateChatConfig({
          modelId: newModelId,
          sessionId: oldRecorder.sessionId || '',
          pending: true,
        });
        this.modelChatList.splice(index, 1, newRecorder);

        // 如果是模型体验 & 已经建立了会话：需要进行更新历史记录更新；
        if (this.isModelExprience && this.modelExperienceId) {
          await createAndUpdateChat({
            id: this.modelExperienceId,
            sessionId: newRecorder.sessionId,
            modelId: newModelId,
            modelType: config.modelList[0].modelType,
            modelSetting: newRecorder.modelSetting,
          });
          this.fetchExprienceDetail(this.modelExperienceId, newModelId);
          this.$emit('refreshHistoryList');
        } else {
          // 只需获取模型信息即可
          this.fetchModelDetail(newModelId).then(result => {
            this.modelChatList.splice(index, 1, {
              ...newRecorder,
              ...result,
            });
          });
        }
      } else {
        // 模型对比环境下，新增模型
        ids.forEach(id => {
          this.modelChatList.push(
            generateChatConfig({
              modelId: id,
            }),
          );
          this.fetchModelDetail(id).then(result => {
            this.modelChatList.some((item, index) => {
              if (item.modelId === id) {
                this.modelChatList.splice(
                  index,
                  1,
                  generateChatConfig({ ...result, sessionId: this.$guid() }),
                );
                return true;
              }
              return false;
            });
          });
        });
      }
    },
    setSessionRef(modelId, el) {
      this.sessionRefs[modelId] = el;
      return this.sessionRefs[modelId];
    },
    preStop() {
      Object.entries(this.sessionRefs).forEach(([, sessionRef]) => {
        sessionRef && sessionRef.preStop();
      });
    },
    refresh() {
      Object.entries(this.sessionRefs).forEach(([, sessionRef]) => {
        sessionRef && sessionRef.refresh();
      });
    },
    async preSend() {
      if (this.isChatGenerating) {
        this.$message.warning(this.$t('modelExprience.warning.chatGenerating'));
        return;
      }
      this.isInit = false;
      const inputVal = this.$refs['editable'].getPrompt();
      const fileList = [...this.$refs['editable'].getFileList()];
      const fileInfo = [...this.$refs['editable'].getFileIdList()];
      if (!inputVal) {
        this.$message.warning(this.$t('modelExprience.warning.rejectEmpty'));
        return;
      }
      if (this.isGenerating) {
        this.$message.warning(this.$t('modelExprience.warning.chatGenerating'));
        return;
      }
      let isNewChat = false;
      // 模型体验 && 未创建对话
      if (this.isModelExprience && !this.modelExperienceId) {
        const chatVo = this.modelChatList[0];
        if (!chatVo) {
          return;
        }
        isNewChat = true;
        chatVo.sessionId = this.$guid();
        this.isChatGenerating = true;
        // 先占位
        this.modelChatList.forEach(item => {
          item.sessionStatus = 0;
          this.sessionRefs[item.modelId] &&
            this.sessionRefs[item.modelId].beforeCreateChat({
              inputVal,
              fileList,
            });
        });
        this.$refs.editable.clearInput();
        const result = await createAndUpdateChat({
          modelId: chatVo.modelId,
          modelType: chatVo.modelType,
          modelSetting: chatVo.modelSetting,
          sessionId: chatVo.sessionId,
          title: inputVal,
        }).finally(() => {
          this.isChatGenerating = false;
        });
        this.$emit('update:modelExperienceId', result.data.id);
        this.$emit('refreshHistoryList');
      }
      this.$nextTick(() => {
        this.modelChatList.forEach(item => {
          if (this.sessionRefs[item.modelId]) {
            // 把之前占位的数据清除，开始进入真正的发送流程
            isNewChat && this.sessionRefs[item.modelId].afterCreateChat();
            this.sessionRefs[item.modelId].autoScroll = true;
            this.sessionRefs[item.modelId].preSend(
              inputVal,
              fileList,
              fileInfo,
            );
          }
        });
        this.$refs.editable.clearInput();
        this.$refs.editable.clearFile();
        // 已有会话发送消息后，通知父组件判断是否需要刷新列表置顶
        if (!isNewChat) {
          this.$emit('pinCurrentChat');
        }
      });
    },
    handleExitComparison() {
      this.$confirm(
        this.$t('modelExprience.warning.exitModelComparison'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(() => {
          this.$emit('closeModelComparison');
        })
        .catch(() => {});
    },
    handleSetQuery(query) {
      this.$refs['editable'].setPrompt(query);
    },
  },
};
</script>
<style lang="scss" scoped>
.conversation-pane {
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  &.modelExprience {
    width: 100%;
  }
  &.modelComparison {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;
    z-index: 999;
  }
  .page-title {
    flex-shrink: 0;
  }
  .page-content {
    flex: 1;
    display: flex;
    padding: 20px;
    overflow: hidden;
    .chat-wrapper {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
      .chat-header {
        flex-shrink: 0;
        text-align: right;
        margin-bottom: 12px;
      }
      .chat-content {
        flex: 1;
        display: flex;
        gap: 24px;
        overflow: hidden;
        > div {
          flex: 1;
        }
      }
      .chat-footer {
        flex-shrink: 0;
        .stop {
          display: flex;
          justify-content: center;
          align-items: center;
          cursor: pointer;
          margin-bottom: 5px;
        }
        .stop-icon {
          width: 18px;
          margin-right: 3px;
        }
        .mdl {
          font-size: 14px;
        }
      }
    }
  }
}
</style>
