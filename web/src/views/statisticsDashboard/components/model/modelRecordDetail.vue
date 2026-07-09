<template>
  <el-dialog
    :visible.sync="dialogVisible"
    append-to-body
    :close-on-click-modal="false"
    width="720px"
    custom-class="model-record-detail-dialog"
  >
    <div slot="title" class="dialog-title-wrap">
      <span class="dialog-title">{{ title }}</span>
      <el-tag
        v-if="currentRow.status === '成功'"
        type="success"
        size="small"
        class="status-tag"
      >
        {{ currentRow.status }}
      </el-tag>
      <el-tag
        v-else-if="currentRow.status"
        type="danger"
        size="small"
        class="status-tag"
      >
        {{ currentRow.status }}
      </el-tag>
    </div>
    <div v-if="currentRow" class="detail-body">
      <!-- 基本信息 -->
      <div class="section info-section">
        <div class="info-grid">
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.modelName') }}:</label>
            <span>{{ currentRow.model || '--' }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.provider') }}:</label>
            <span>{{ providerName }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('modelAccess.table.modelType') }}:</label>
            <span>
              <span :class="['type-tag', modelTypeTagClass]">
                {{ modelTypeName }}
              </span>
            </span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appName') }}:</label>
            <span>{{ currentRow.appName || '--' }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.modelTypeLabel') }}:</label>
            <span>{{ currentRow.modelTypeLabel || '--' }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appType') }}:</label>
            <span>
              <span
                v-if="currentRow.appType"
                :class="['type-tag', appTypeTagClass]"
              >
                {{ appTypeName }}
              </span>
              <span v-else>--</span>
            </span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.org') }}:</label>
            <span>{{ currentRow.orgName || '--' }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appSystem') }}:</label>
            <span>{{ currentRow.appSystem || '--' }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appName') }}:</label>
            <span>{{ currentRow.appName || '--' }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.appType') }}:</label>
            <span>
              <span
                v-if="currentRow.appType"
                :class="['type-tag', appTypeTagClass]"
              >
                {{ appTypeName }}
              </span>
              <span v-else>--</span>
            </span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.creator') }}:</label>
            <span>{{ currentRow.creator || currentRow.userName || '--' }}</span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.createTime') }}:</label>
            <span>
              {{ currentRow.createTime || currentRow.callTime || '--' }}
            </span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.userDept') }}:</label>
            <span>
              {{ currentRow.userDept || currentRow.userName || '--' }}
            </span>
          </div>
          <div class="info-item">
            <label>{{ $t('statisticsDashboard.callTime') }}:</label>
            <span>{{ currentRow.callTime || '--' }}</span>
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
              {{ $t('statisticsDashboard.promptTokens') }}
            </div>
            <div class="token-value purple">
              {{ formatAmount(currentRow.promptTokens) }}
            </div>
          </div>
          <div class="token-card">
            <div class="token-label">
              {{ $t('statisticsDashboard.completionTokens') }}
            </div>
            <div class="token-value blue">
              {{ formatAmount(currentRow.completionTokens) }}
            </div>
          </div>
          <div class="token-card">
            <div class="token-label">
              {{ $t('statisticsDashboard.totalTokens') }}
            </div>
            <div class="token-value primary">
              {{ formatAmount(currentRow.totalTokens) }}
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

    <!-- 底部导航 -->
    <div slot="footer" class="dialog-footer">
      <span class="page-info">
        {{ currentIndex + 1 }} / {{ list.length }}
        {{ $t('statisticsDashboard.pieces') }}
      </span>
      <el-button-group>
        <el-button
          size="mini"
          icon="el-icon-s-home"
          :disabled="currentIndex === 0"
          @click="goFirst"
        />
        <el-button
          size="mini"
          icon="el-icon-arrow-left"
          :disabled="currentIndex === 0"
          @click="goPrev"
        />
        <el-button
          size="mini"
          icon="el-icon-arrow-right"
          :disabled="currentIndex === list.length - 1"
          @click="goNext"
        />
        <el-button
          size="mini"
          icon="el-icon-s-promotion"
          :disabled="currentIndex === list.length - 1"
          @click="goLast"
        />
      </el-button-group>
      <el-button size="mini" @click="dialogVisible = false">
        {{ $t('common.button.cancel') }}
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
    list: {
      type: Array,
      default: () => [],
    },
    startIndex: {
      type: Number,
      default: 0,
    },
    modelMap: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      currentIndex: 0,
    };
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
      if (
        this.list.length > 0 &&
        this.currentIndex >= 0 &&
        this.currentIndex < this.list.length
      ) {
        return this.list[this.currentIndex];
      }
      return this.row;
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
    statusTagType() {
      return this.currentRow?.status === '成功' ? 'success' : 'danger';
    },
    isFailed() {
      return this.currentRow?.status && this.currentRow.status !== '成功';
    },
  },
  watch: {
    visible(val) {
      if (val) {
        this.currentIndex = this.startIndex;
      }
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
    goFirst() {
      this.currentIndex = 0;
    },
    goPrev() {
      if (this.currentIndex > 0) {
        this.currentIndex--;
      }
    },
    goNext() {
      if (this.currentIndex < this.list.length - 1) {
        this.currentIndex++;
      }
    },
    goLast() {
      this.currentIndex = this.list.length - 1;
    },
  },
};
</script>

<style lang="scss" scoped>
.model-record-detail-dialog {
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
  }
}

.detail-body {
  max-height: 60vh;
  overflow-y: auto;
  padding-right: 4px;
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
  align-items: center;
  font-size: 13px;
  line-height: 1.5;

  label {
    color: #909399;
    width: 100px;
    flex-shrink: 0;
    text-align: right;
    margin-right: 8px;
  }

  span {
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
    color: #f56c6c;
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
      color: #409eff;
    }

    &.primary {
      color: #67c23a;
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
  color: #f56c6c;
  word-break: break-all;
  padding: 8px 12px;
  background: #fff;
  border-radius: 4px;
}

.content-box {
  font-size: 13px;
  color: #606266;
  line-height: 1.6;
  padding: 12px;
  background: #fff;
  border-radius: 4px;
  border: 1px solid #ebeef5;
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

.page-info {
  font-size: 13px;
  color: #909399;
  margin-right: auto;
}

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.5;
}

.tag-blue {
  color: $tag_color;
  background: $tag_bg;
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
