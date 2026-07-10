<template>
  <el-dialog
    :visible.sync="dialogVisible"
    append-to-body
    :close-on-click-modal="false"
    width="720px"
    custom-class="api-record-detail-dialog"
  >
    <div slot="title" class="dialog-title-wrap">
      <span class="dialog-title">{{ title }}</span>
      <el-tag v-if="isSuccess" type="success" size="small" class="status-tag">
        {{ $t('statisticsDashboard.success') }}
      </el-tag>
      <el-tag
        v-else-if="isFailed"
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
            <label>{{ 'API Key' + $t('statisticsDashboard.name') }}</label>
            <div>{{ currentRow.name || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ 'API Key' }}</label>
            <div>
              {{
                currentRow.apiKey
                  ? currentRow.apiKey.slice(0, 6) + '******'
                  : '--'
              }}
            </div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.org') }}</label>
            <div>{{ currentRow.orgName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.userName') }}</label>
            <div>{{ currentRow.userName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.apiPath') }}</label>
            <div>{{ currentRow.methodPath || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.responseStatus') }}</label>
            <div>{{ currentRow.responseStatus || '--' }}</div>
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
            <label>{{ $t('statisticsDashboard.streamCosts') }}</label>
            <span class="perf-value">
              {{ formatTime(currentRow.streamCosts) }}
            </span>
            <span class="perf-unit">
              {{ $t('statisticsDashboard.streamMode') }}
            </span>
          </div>
          <div class="perf-item">
            <label>{{ $t('statisticsDashboard.nonStreamCosts') }}</label>
            <span class="perf-value">
              {{ formatTime(currentRow.nonStreamCosts) }}
            </span>
            <span class="perf-unit">
              {{ $t('statisticsDashboard.singleMode') }}
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
          {{
            currentRow.failReason ||
            currentRow.errorMsg ||
            currentRow.responseStatus ||
            '--'
          }}
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
      const name = this.currentRow?.name || '';
      return name
        ? `${name} - ${this.$t('statisticsDashboard.detailTitle')}`
        : this.$t('statisticsDashboard.detailTitle');
    },
    responseStatus() {
      return this.currentRow?.responseStatus;
    },
    isSuccess() {
      const status = this.responseStatus;
      if (status === undefined || status === null || status === '') {
        return false;
      }
      return /^2/.test(String(status));
    },
    isFailed() {
      const status = this.responseStatus;
      if (status === undefined || status === null || status === '') {
        return false;
      }
      return !this.isSuccess;
    },
  },
  methods: {
    formatAmount,
    formatTime(val) {
      if (!val) return '0';
      const num = Number(val);
      if (num >= 1000) {
        return (num / 1000).toFixed(1) + 's';
      }
      return num + 'ms';
    },
  },
};
</script>

<style lang="scss" scoped>
.api-record-detail-dialog {
  ::v-deep .el-dialog__body {
    padding: 0 20px;
  }
}

.dialog-title-wrap {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-right: 36px;

  .dialog-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }

  .status-tag {
    font-weight: 500;
    border-radius: 12px;
  }
}

.detail-body {
  margin-top: -22px;
}

.section {
  margin-bottom: 16px;
  padding: 16px;
  background: #f8f9fb;
  border-radius: 8px;

  &.info-section {
    background: transparent;
    padding: 0;
  }

  &.token-section {
    background: #f5f7ff;
  }

  &.perf-section {
    background: #fff8f0;
  }

  &.fail-section {
    background: #fff2f0;
  }
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;
}

.info-item {
  display: flex;
  align-items: flex-start;
  font-size: 13px;
  line-height: 1.6;

  label {
    color: #909399;
    width: 100px;
    flex-shrink: 0;
    text-align: right;
    margin-right: 8px;
  }

  div {
    color: #303133;
    flex: 1;
    min-width: 0;
    word-break: break-all;
  }
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;

  &.fail-title {
    color: #bd3d3e;
  }
}

.token-cards {
  display: flex;
  gap: 16px;
}

.token-card {
  flex: 1;
  text-align: center;
  padding: 16px 8px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

  .token-label {
    font-size: 12px;
    color: #909399;
    margin-bottom: 8px;
  }

  .token-value {
    font-size: 22px;
    font-weight: 700;

    &.purple {
      color: #5951e7;
    }

    &.blue {
      color: #2563eb;
    }

    &.primary {
      color: #9233e9;
    }
  }
}

.perf-grid {
  display: flex;
  gap: 24px;
}

.perf-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  width: 50%;
  background: #fff;
  padding: 15px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

  label {
    color: #909399;
  }

  .perf-value {
    color: #e6a23c;
    font-weight: 600;
    font-size: 16px;
  }

  .perf-unit {
    color: #909399;
    font-size: 12px;
  }
}

.fail-content {
  font-size: 13px;
  color: #bd3d3e;
  word-break: break-all;
  padding: 8px 12px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.content-box {
  font-size: 13px;
  color: #606266;
  line-height: 1.6;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  white-space: pre-wrap;
  word-break: break-all;
  min-height: 60px;
  max-height: 160px;
  overflow-y: auto;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
}
</style>
