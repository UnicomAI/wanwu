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
            <div>{{ currentRow.provider || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('modelAccess.table.modelType') }}</label>
            <div>
              {{ modelTypeName }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.modelPublisher') }}</label>
            <div>{{ currentRow.modelCreatorUserName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.modelOrg') }}</label>
            <div>{{ currentRow.modelCreatorOrgName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.source') }}</label>
            <div>
              {{ currentRow.source || '--' }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.module') }}</label>
            <div>
              {{ currentRow.module || '--' }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appName') }}</label>
            <div>{{ currentRow.appName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appType') }}</label>
            <div>
              {{ appTypeName }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appAuthor') }}</label>
            <div>{{ currentRow.moduleCreatorUserName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appAuthorOrg') }}</label>
            <div>{{ currentRow.moduleCreatorOrgName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.user') }}</label>
            <div>
              {{ currentRow.userName || '--' }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.userOrg') }}</label>
            <div>
              {{ currentRow.orgName || '--' }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.callTime') }}</label>
            <div>{{ currentRow.calledAt || '--' }}</div>
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
            <label>{{ $t('statisticsDashboard.costs') }}</label>
            <span class="perf-value">
              {{ formatSec(currentRow.costs) }}
            </span>
          </div>
          <div class="perf-item">
            <label>{{ $t('statisticsDashboard.firstTokenLatency') }}</label>
            <span class="perf-value">
              {{ formatSec(currentRow.firstTokenLatency) }}
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
          {{ currentRow.failureReason || '--' }}
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
import { formatAmount, formatSec } from '@/utils/util.js';
import { MODEL_TYPE_OBJ } from '@/views/modelAccess/constants';
import { TotalTypeObj } from '@/utils/commonSet';

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
    modelTypeName() {
      return (
        MODEL_TYPE_OBJ[this.currentRow?.modelType] ||
        this.currentRow?.modelType ||
        '--'
      );
    },
    appTypeName() {
      return (
        TotalTypeObj[this.currentRow?.appType] ||
        this.currentRow?.appType ||
        '--'
      );
    },
    isFailed() {
      return this.currentRow.isSuccess === false;
    },
  },
  methods: {
    formatAmount,
    formatSec,
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsDetail.scss';
</style>
