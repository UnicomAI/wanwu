<template>
  <el-dialog
    :title="$t('statisticsDashboard.appUsageStats')"
    :visible.sync="visible"
    width="90%"
    top="5vh"
    :close-on-click-modal="false"
    custom-class="usage-detail-modal"
    append-to-body
    @open="handleOpen"
    @close="handleClose"
  >
    <div class="list-common">
      <div class="modal-toolbar">
        <el-button type="primary" size="mini" @click="exportData">
          <span>{{ $t('common.button.export') }}</span>
        </el-button>
      </div>
      <el-table
        :data="tableData"
        v-loading="loading"
        :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
        style="width: 100%"
      >
        <el-table-column
          prop="modelName"
          :label="$t('statisticsDashboard.modelName')"
          align="left"
          min-width="140"
        >
          <template slot-scope="scope">{{ getModelName(scope.row) }}</template>
        </el-table-column>
        <el-table-column
          prop="provider"
          :label="$t('statisticsDashboard.provider')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ getProviderName(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="publisher"
          :label="$t('statisticsDashboard.publisher')"
          align="left"
          min-width="140"
        >
          <template slot-scope="scope">
            {{ getPublisherName(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="orgName"
          :label="$t('statisticsDashboard.org')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">{{ getOrgName(scope.row) }}</template>
        </el-table-column>
        <el-table-column
          prop="modelType"
          :label="$t('modelAccess.table.modelType')"
          align="left"
          width="110px"
        >
          <template slot-scope="scope">
            <span :class="['type-tag', getModelTypeTagClass(scope.row)]">
              {{ getModelTypeName(scope.row) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="source"
          :label="$t('statisticsDashboard.source')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">{{ scope.row.source || '--' }}</template>
        </el-table-column>
        <el-table-column
          prop="module"
          :label="$t('statisticsDashboard.module')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">{{ scope.row.module || '--' }}</template>
        </el-table-column>
        <el-table-column
          prop="appName"
          :label="$t('statisticsDashboard.appName')"
          align="left"
          min-width="140"
        >
          <template slot-scope="scope">
            {{ scope.row.appName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="appType"
          :label="$t('statisticsDashboard.appType')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            <span :class="['type-tag', getAppTypeTagClass(scope.row)]">
              {{ appTypeObj[scope.row.appType] || scope.row.appType || '--' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="author"
          :label="$t('statisticsDashboard.author')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.author || scope.row.userName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="orgName"
          :label="$t('statisticsDashboard.org')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.orgName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="totalTokens"
          :label="$t('statisticsDashboard.totalTokens')"
          align="left"
          sortable="custom"
          min-width="120"
        >
          <template slot-scope="scope">
            <span class="token-purple">
              {{ formatAmount(scope.row.totalTokens) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="promptTokens"
          :label="$t('statisticsDashboard.promptTokens')"
          align="left"
          sortable="custom"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ formatAmount(scope.row.promptTokens) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="completionTokens"
          :label="$t('statisticsDashboard.completionTokens')"
          align="left"
          sortable="custom"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ formatAmount(scope.row.completionTokens) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="callCount"
          :label="$t('statisticsDashboard.callCount')"
          align="left"
          sortable="custom"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ formatAmount(scope.row.callCount) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="failureRate"
          :label="$t('statisticsDashboard.failureRate')"
          align="left"
          sortable="custom"
          min-width="110"
        >
          <template slot-scope="scope">{{ scope.row.failureRate }}%</template>
        </el-table-column>
        <el-table-column
          prop="callFailure"
          :label="$t('statisticsDashboard.callFailure')"
          align="left"
          sortable="custom"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ formatAmount(scope.row.callFailure) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="avgCosts"
          :label="$t('statisticsDashboard.avgCosts')"
          align="left"
          sortable="custom"
          min-width="150"
        >
          <template slot-scope="scope">
            {{ formatTime(scope.row.avgCosts, 'avgCosts') }}
          </template>
        </el-table-column>
        <el-table-column
          prop="avgFirstTokenLatency"
          :label="$t('statisticsDashboard.avgFirstTokenLatency')"
          align="left"
          sortable="custom"
          min-width="170"
        >
          <template slot-scope="scope">
            {{
              formatTime(scope.row.avgFirstTokenLatency, 'avgFirstTokenLatency')
            }}
          </template>
        </el-table-column>
      </el-table>
      <Pagination
        class="pagination"
        ref="pagination"
        :listApi="listApi"
        @refreshData="refreshData"
      />
    </div>
  </el-dialog>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import { formatAmount, resDownloadFile } from '@/utils/util.js';
import {
  fetchModelAppList,
  exportModelAppData,
} from '@/api/statisticsDashboard';
import { AppType, AGENT, CHAT, RAG, WORKFLOW } from '@/utils/commonSet';
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

export default {
  components: { Pagination },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    params: {
      type: Object,
      default: () => ({}),
    },
    modelInfo: {
      type: Object,
      default: () => ({}),
    },
    modelMap: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      listApi: fetchModelAppList,
      loading: false,
      tableData: [],
      appTypeObj: AppType,
    };
  },
  computed: {
    modalParams() {
      const { models, ...rest } = this.params || {};
      return {
        ...rest,
        modelId: this.modelInfo?.modelId || this.modelInfo?.model || '',
      };
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
    handleOpen() {
      this.$nextTick(() => {
        if (this.$refs.pagination) {
          this.$refs.pagination.pageNo = 1;
          this.$refs.pagination.pageSize = 10;
          this.$refs.pagination.searchInfo = {};
        }
        this.getTableData({ ...this.modalParams, pageNo: 1 });
      });
    },
    handleClose() {
      this.$emit('update:visible', false);
      this.tableData = [];
    },
    async getTableData(params) {
      if (this.$refs.pagination) {
        this.loading = true;
        try {
          this.tableData = await this.$refs.pagination.getTableData(params);
        } finally {
          this.loading = false;
        }
      }
    },
    refreshData(data) {
      this.tableData = data;
    },
    async exportData() {
      const response = await exportModelAppData(this.modalParams);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.appUsageStats')}.xlsx`,
      );
    },
    getModelId(row) {
      return (
        row.modelId ||
        this.modelInfo?.modelId ||
        row.model ||
        this.modelInfo?.model ||
        ''
      );
    },
    getModelName(row) {
      return row.modelName || row.model || this.modelInfo?.model || '--';
    },
    getProviderName(row) {
      const provider = row.provider || this.modelInfo?.provider || '';
      return PROVIDER_OBJ[provider] || provider || '--';
    },
    getPublisherName(row) {
      if (row.publisher) {
        return row.publisher;
      }
      const orgName = row.orgName || this.modelInfo?.orgName || '';
      const userName = row.userName || '';
      if (orgName && userName) {
        return `${orgName} ${userName}`;
      }
      return orgName || userName || '--';
    },
    getOrgName(row) {
      return row.orgName || this.modelInfo?.orgName || '--';
    },
    getModelTypeName(row) {
      const modelId = this.getModelId(row);
      const modelInfo = this.modelMap[modelId] || {};
      const type =
        row.modelType || modelInfo.modelType || this.modelInfo?.modelType || '';
      return MODEL_TYPE_OBJ[type] || '--';
    },
    getModelTypeTagClass(row) {
      const modelId = this.getModelId(row);
      const modelInfo = this.modelMap[modelId] || {};
      const type =
        row.modelType || modelInfo.modelType || this.modelInfo?.modelType || '';
      if (type === LLM) return 'tag-blue';
      if ([RERANK, MULTIMODAL_RERANK].includes(type)) return 'tag-orange';
      if ([EMBEDDING, MULTIMODAL_EMBEDDING].includes(type)) return 'tag-green';
      if ([OCR, ASR, GUI, PDF_PARSER].includes(type)) return 'tag-purple';
      return 'tag-gray';
    },
    getAppTypeTagClass(row) {
      const typeTag = {
        [AGENT]: 'tag-purple',
        [WORKFLOW]: 'tag-green',
        [RAG]: 'tag-blue',
        [CHAT]: 'tag-orange',
      };
      return typeTag[row.appType] || 'tag-gray';
    },
  },
};
</script>

<style lang="scss" scoped>
.modal-toolbar {
  margin-bottom: 16px;
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

.token-purple {
  color: #5951e7;
  font-weight: 500;
}
</style>
