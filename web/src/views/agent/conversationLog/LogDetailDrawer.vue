<template>
  <el-drawer
    :visible.sync="visible"
    :title="$t('agent.log.detailDrawer.title')"
    direction="rtl"
    size="70%"
    @closed="handleClosed"
  >
    <div v-loading="loading" class="log-detail-drawer">
      <div class="log-detail-drawer__filter">
        <span class="log-detail-drawer__label">
          {{ $t('agent.log.detailDrawer.feedback') }}
        </span>
        <el-select
          v-model="feedbackType"
          class="log-detail-drawer__select"
          @change="handleFeedbackChange"
        >
          <el-option
            v-for="item in feedbackOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>

      <div class="log-detail-drawer__messages">
        <stream-message-field
          v-show="filteredHistoryData.length"
          ref="streamMessageField"
          chat-type="agent"
          :support-clear="false"
          :support-stop="false"
          :support-answer-feedback="false"
          :model-session-status="1"
          :default-url="defaultUrl"
          :answer-operation-options="{
            showCopy: true,
            showFeedback: true,
            feedbackReadonly: true,
            feedbackDisplayMode: 'selectedOnly',
            showFeedbackContent: true,
            showRefresh: false,
            showStop: false,
            showDelete: false,
            showTip: false,
          }"
        />
        <el-empty
          v-if="!loading && !filteredHistoryData.length"
          class="log-detail-drawer__empty"
          :description="$t('common.noData')"
        />
      </div>
    </div>
  </el-drawer>
</template>

<script>
import streamMessageField from '@/components/stream/streamMessageField';
import { transformAgentHistory } from '@/utils/agentHistoryTransformer';

const ALL_OPTION_VALUE = '__all__';

export default {
  name: 'AgentConversationLogDetailDrawer',
  components: {
    streamMessageField,
  },
  props: {
    defaultUrl: {
      type: String,
      default: '',
    },
    getDetail: {
      type: Function,
      required: true,
    },
  },
  data() {
    return {
      visible: false,
      loading: false,
      log: null,
      detailData: null,
      historyData: [],
      feedbackType: ALL_OPTION_VALUE,
    };
  },
  computed: {
    feedbackOptions() {
      return [
        {
          label: this.$t('agent.log.all'),
          value: ALL_OPTION_VALUE,
        },
        {
          label: this.$t('agent.log.detailDrawer.like'),
          value: 1,
        },
        {
          label: this.$t('agent.log.detailDrawer.dislike'),
          value: 2,
        },
      ];
    },
    filteredHistoryData() {
      if (this.feedbackType === ALL_OPTION_VALUE) return this.historyData;
      return this.historyData.filter(
        item => item.feedback === this.feedbackType,
      );
    },
  },
  methods: {
    showDrawer({ appId, appType, conversationId, log }) {
      this.log = log || null;
      this.detailData = null;
      this.historyData = [];
      this.feedbackType = ALL_OPTION_VALUE;
      this.visible = true;
      this.loadDetail({ appId, appType, conversationId });
    },
    async loadDetail({ appId, appType, conversationId }) {
      if (!appId || !appType || !conversationId) return;

      this.loading = true;
      try {
        const res = await this.getDetail({
          appId,
          appType,
          conversationId,
          pageNo: 1,
          pageSize: 1000,
        });
        if (res && res.code === 0) {
          this.detailData = res.data || null;
          this.historyData = transformAgentHistory(
            (this.detailData && this.detailData.list) || [],
          );
          this.replaceHistory();
          this.$emit('loaded', this.detailData);
        } else {
          this.$message.error(res?.msg || this.$t('common.message.error'));
        }
      } catch (error) {
        this.$message.error(error?.message || this.$t('common.message.error'));
      } finally {
        this.loading = false;
      }
    },
    handleFeedbackChange(value) {
      this.feedbackType = value;
      this.replaceHistory();
      this.$emit('feedback-change', value);
    },
    replaceHistory() {
      this.$nextTick(() => {
        this.$refs.streamMessageField.replaceHistory(this.filteredHistoryData);
      });
    },
    handleClosed() {
      this.log = null;
      this.detailData = null;
      this.historyData = [];
    },
  },
};
</script>

<style scoped lang="scss">
::v-deep .el-drawer__header {
  margin-bottom: 0;
  color: #303133;
  font-weight: 600;
  span {
    font-size: 16px;
  }
}

::v-deep .el-drawer__close-btn {
  font-size: 16px;
}

.log-detail-drawer {
  padding: 20px;

  &__filter {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  &__label {
    color: #4b5563;
    font-size: 14px;
  }

  &__select {
    width: 240px;
  }

  &__messages {
    height: calc(100vh - 150px);
    margin-top: 16px;

    ::v-deep .session {
      height: 100%;
    }
  }

  &__empty {
    display: flex;
    align-items: center;
    flex-direction: column;
    justify-content: center;
    height: 100%;
  }
}
</style>
