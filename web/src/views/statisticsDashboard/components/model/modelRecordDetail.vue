<template>
  <el-dialog
    :visible.sync="dialogVisible"
    append-to-body
    :close-on-click-modal="false"
    width="720px"
    custom-class="dashboard-record-detail-dialog"
  >
    <div slot="title" class="dialog-title-wrap">
      <span class="dialog-title">{{ title }}</span>
      <el-tag
        v-if="currentRow.status === 'success'"
        type="success"
        size="small"
        class="status-tag"
      >
        {{ $t('statisticsDashboard.success') }}
      </el-tag>
      <el-tag
        v-else-if="currentRow.status === 'error'"
        type="danger"
        size="small"
        class="status-tag"
      >
        {{ $t('statisticsDashboard.error') }}
      </el-tag>
    </div>
    <div v-if="currentRow" class="detail-body">
      <!-- 基本信息 -->
      <div class="section info-section">
        <div class="info-grid">
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.modelName') }}</label>
            <div>{{ currentRow.model || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.provider') }}</label>
            <div>{{ providerName }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('modelAccess.table.modelType') }}</label>
            <div>
              <span :class="['type-tag', modelTypeTagClass]">
                {{ modelTypeName }}
              </span>
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appName') }}</label>
            <div>{{ currentRow.appName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.modelTypeLabel') }}</label>
            <div>{{ currentRow.modelTypeLabel || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appType') }}</label>
            <div>
              <span
                v-if="currentRow.appType"
                :class="['type-tag', appTypeTagClass]"
              >
                {{ appTypeName }}
              </span>
              <span v-else>--</span>
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.org') }}</label>
            <div>{{ currentRow.orgName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appSystem') }}</label>
            <div>{{ currentRow.appSystem || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appName') }}</label>
            <div>{{ currentRow.appName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appType') }}</label>
            <div>
              <span
                v-if="currentRow.appType"
                :class="['type-tag', appTypeTagClass]"
              >
                {{ appTypeName }}
              </span>
              <span v-else>--</span>
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.creator') }}</label>
            <div>{{ currentRow.creator || currentRow.userName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.createTime') }}</label>
            <div>
              {{ currentRow.createTime || currentRow.callTime || '--' }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.userDept') }}</label>
            <div>
              {{ currentRow.userDept || currentRow.userName || '--' }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.callTime') }}</label>
            <div>{{ currentRow.callTime || '--' }}</div>
          </div>
        </div>
      </div>

      <!-- Token 统计 -->
      <div class="section token-section">
        <div class="section-title">
          {{ $t('statisticsDashboard.tokenStats') }}
        </div>
        <div class="token-cards">
          <div class="token-card">
            <div class="token-label">
              {{ $t('statisticsDashboard.totalTokens') }}
            </div>
            <div class="token-value purple">
              {{ formatAmount(currentRow.totalTokens) }}
            </div>
          </div>
          <div class="token-card">
            <div class="token-label">
              {{ $t('statisticsDashboard.promptTokens') }}
            </div>
            <div class="token-value blue">
              {{ formatAmount(currentRow.promptTokens) }}
            </div>
          </div>
          <div class="token-card">
            <div class="token-label">
              {{ $t('statisticsDashboard.completionTokens') }}
            </div>
            <div class="token-value primary">
              {{ formatAmount(currentRow.completionTokens) }}
            </div>
          </div>
        </div>
      </div>

      <!-- 性能指标 -->
      <div class="section perf-section">
        <div class="section-title">
          {{ $t('statisticsDashboard.perfMetrics') }}
        </div>
        <div class="perf-grid">
          <div class="perf-item">
            <label>{{ $t('statisticsDashboard.totalCosts') }}</label>
            <span class="perf-value">
              {{ formatTime(currentRow.avgCosts, 'avgCosts') }}
            </span>
            <span class="perf-unit">
              {{ $t('statisticsDashboard.singleMode') }}
            </span>
          </div>
          <div class="perf-item">
            <label>{{ $t('statisticsDashboard.firstTokenTime') }}</label>
            <span class="perf-value">
              {{
                formatTime(
                  currentRow.avgFirstTokenLatency,
                  'avgFirstTokenLatency',
                )
              }}
            </span>
            <span class="perf-unit">
              {{ $t('statisticsDashboard.streamMode') }}
            </span>
          </div>
        </div>
      </div>

      <!-- 失败原因 -->
      <div v-if="isFailed" class="section fail-section">
        <div class="section-title fail-title">
          {{ $t('statisticsDashboard.failReason') }}
        </div>
        <div class="fail-content">
          {{ currentRow.failReason || currentRow.errorMsg || '--' }}
        </div>
      </div>

      <!-- 输入内容 -->
      <div class="section content-section">
        <div class="section-title">
          {{ $t('statisticsDashboard.reqContent') }}
        </div>
        <div class="content-box">{{ currentRow.requestBody || '--' }}</div>
      </div>

      <!-- 输出内容 -->
      <div class="section content-section">
        <div class="section-title">
          {{ $t('statisticsDashboard.resContent') }}
        </div>
        <div class="content-box">{{ currentRow.responseBody || '--' }}</div>
      </div>
    </div>

    <div slot="footer" class="dialog-footer">
      <el-button size="mini" @click="dialogVisible = false">
        {{ $t('common.button.close') }}
      </el-button>
    </div>
  </el-dialog>
</template>

<script>
import { formatAmount } from '@/utils/util.js';
import {
  MODEL_TYPE_OBJ,
  PROVIDER_OBJ,
  LLM,
  RERANK,
  EMBEDDING,
  OCR,
  GUI,
  PDF_PARSER,
  ASR,
  MULTIMODAL_RERANK,
  MULTIMODAL_EMBEDDING,
} from '@/views/modelAccess/constants';
import { AGENT, AppType, CHAT, RAG, WORKFLOW } from '@/utils/commonSet';

export default {
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    row: {
      type: Object,
      default: () => ({}),
    },
    modelMap: {
      type: Object,
      default: () => ({}),
    },
  },
  computed: {
    dialogVisible: {
      get() {
        return this.visible;
      },
      set(val) {
        this.$emit('update:visible', val);
      },
    },
    currentRow() {
      return this.row || {};
    },
    title() {
      const name = this.currentRow?.model || '';
      return name
        ? `${name} - ${this.$t('statisticsDashboard.detailTitle')}`
        : this.$t('statisticsDashboard.detailTitle');
    },
    providerName() {
      const provider = this.currentRow?.provider || '';
      return PROVIDER_OBJ[provider] || provider || '--';
    },
    modelTypeName() {
      const modelId = this.currentRow?.modelId || this.currentRow?.model || '';
      const modelInfo = this.modelMap[modelId] || {};
      const type = modelInfo.modelType || this.currentRow?.modelType || '';
      return MODEL_TYPE_OBJ[type] || '--';
    },
    modelTypeTagClass() {
      const modelId = this.currentRow?.modelId || this.currentRow?.model || '';
      const modelInfo = this.modelMap[modelId] || {};
      const type = modelInfo.modelType || this.currentRow?.modelType || '';
      if (type === LLM) return 'tag-blue';
      if ([RERANK, MULTIMODAL_RERANK].includes(type)) return 'tag-orange';
      if ([EMBEDDING, MULTIMODAL_EMBEDDING].includes(type)) return 'tag-green';
      if ([OCR, ASR, GUI, PDF_PARSER].includes(type)) return 'tag-purple';
      return 'tag-gray';
    },
    appTypeName() {
      return (
        AppType[this.currentRow?.appType] || this.currentRow?.appType || '--'
      );
    },
    appTypeTagClass() {
      const typeTag = {
        [AGENT]: 'tag-purple',
        [WORKFLOW]: 'tag-green',
        [RAG]: 'tag-blue',
        [CHAT]: 'tag-orange',
      };
      return typeTag[this.currentRow?.appType] || 'tag-gray';
    },
    isFailed() {
      return this.currentRow?.status && this.currentRow.status !== 'success';
    },
  },
  methods: {
    formatAmount,
    formatTime(val, type) {
      if (!val) return '0';
      const num = Number(val);
      if (type === 'avgCosts' && num >= 1000) {
        return (num / 1000).toFixed(1) + 's';
      }
      return num + 'ms';
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsDetail.scss';
</style>
