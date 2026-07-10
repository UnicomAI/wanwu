<template>
  <el-dialog
    :visible.sync="dialogVisible"
    append-to-body
    :close-on-click-modal="false"
    width="720px"
    custom-class="app-record-detail-dialog"
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
            <label>{{ $t('statisticsDashboard.org') }}</label>
            <div>{{ currentRow.orgName || '--' }}</div>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.userName') }}</label>
            <div>{{ currentRow.userName || '--' }}</div>
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
      const name = this.currentRow?.appName || '';
      return name
        ? `${name} - ${this.$t('statisticsDashboard.detailTitle')}`
        : this.$t('statisticsDashboard.detailTitle');
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
.app-record-detail-dialog {
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

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.5;
}

.tag-blue {
  color: $tag_color !important;
  background: $tag_bg !important;
}

.tag-green {
  color: #67c23a;
  background: rgba(103, 194, 58, 0.1);
}

.tag-orange {
  color: #e6a23c;
  background: rgba(230, 162, 60, 0.1);
}

.tag-purple {
  color: #a55fef;
  background: rgba(165, 95, 239, 0.1);
}

.tag-gray {
  color: #909399;
  background: rgba(144, 147, 153, 0.1);
}
</style>
