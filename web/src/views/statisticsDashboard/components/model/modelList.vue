<template>
  <div>
    <div class="table-wrap list-common wrap-fullheight">
      <div class="table-box">
        <el-button
          class="add-bt"
          size="mini"
          type="primary"
          @click="exportData"
        >
          <span>{{ $t('common.button.export') }}</span>
        </el-button>
        <el-radio-group v-model="type" size="mini" @change="handleRadio">
          <el-radio-button label="list">
            {{ $t('statisticsDashboard.modelStatistics') }}
          </el-radio-button>
          <el-radio-button label="record">
            {{ $t('statisticsDashboard.modelDetail') }}
          </el-radio-button>
        </el-radio-group>
        <el-table
          v-if="type === 'list'"
          ref="listTable"
          :data="tableData"
          :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
          v-loading="loading"
          style="width: 100%"
          @sort-change="handleSortChange"
        >
          <el-table-column
            prop="model"
            :label="$t('statisticsDashboard.modelName')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.model || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="provider"
            :label="$t('statisticsDashboard.provider')"
            align="left"
          >
            <template slot-scope="scope">
              {{
                providerObj[scope.row.provider] || scope.row.provider || '--'
              }}
            </template>
          </el-table-column>
          <el-table-column
            prop="publisher"
            :label="$t('statisticsDashboard.publisher')"
            align="left"
          >
            <template slot-scope="scope">
              {{ getPublisher(scope.row) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="orgName"
            :label="$t('statisticsDashboard.org')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.orgName || '--' }}
            </template>
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
            prop="totalTokens"
            :label="$t('statisticsDashboard.totalTokens')"
            align="left"
            sortable="custom"
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
          >
            <template slot-scope="scope">{{ scope.row.failureRate }}%</template>
          </el-table-column>
          <el-table-column
            prop="callFailure"
            :label="$t('statisticsDashboard.callFailure')"
            align="left"
            sortable="custom"
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
          >
            <template slot-scope="scope">
              {{
                formatTime(
                  scope.row.avgFirstTokenLatency,
                  'avgFirstTokenLatency',
                )
              }}
            </template>
          </el-table-column>
          <el-table-column
            width="180"
            align="left"
            :label="$t('common.table.operation')"
          >
            <template slot-scope="scope">
              <el-button
                class="btn-user"
                size="mini"
                icon="el-icon-user"
                @click="showUserDetail(scope.row)"
              >
                {{ $t('statisticsDashboard.userName') }}
              </el-button>
              <el-button
                class="btn-app"
                size="mini"
                icon="el-icon-s-grid"
                @click="showAppDetail(scope.row)"
              >
                {{ $t('statisticsDashboard.app') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-table
          v-else
          ref="recordTable"
          :data="tableData"
          :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
          v-loading="loading"
          style="width: 100%"
          @sort-change="handleSortChange"
        >
          <el-table-column
            prop="model"
            :label="$t('statisticsDashboard.modelName')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.model || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="modelType"
            :label="$t('modelAccess.table.modelType')"
            align="left"
          >
            <template slot-scope="scope">
              <span :class="['type-tag', getModelTypeTagClass(scope.row)]">
                {{ getModelTypeName(scope.row) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            prop="appName"
            :label="$t('statisticsDashboard.appName')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.appName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="appType"
            :label="$t('statisticsDashboard.appType')"
            align="left"
          >
            <template slot-scope="scope">
              <span :class="['type-tag', getAppTypeTagClass(scope.row)]">
                {{ appTypeObj[scope.row.appType] || scope.row.appType || '--' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            prop="userName"
            :label="$t('statisticsDashboard.userName')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.userName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="totalTokens"
            :label="$t('statisticsDashboard.totalTokens')"
            align="left"
            sortable="custom"
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
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.completionTokens) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="avgCosts"
            :label="$t('statisticsDashboard.avgCosts')"
            align="left"
            sortable="custom"
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
          >
            <template slot-scope="scope">
              {{
                formatTime(
                  scope.row.avgFirstTokenLatency,
                  'avgFirstTokenLatency',
                )
              }}
            </template>
          </el-table-column>
          <el-table-column
            prop="callTime"
            :label="$t('statisticsDashboard.callTime')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ scope.row.callTime || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="status"
            :label="$t('common.status')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.status || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            width="120"
            align="left"
            :label="$t('common.table.operation')"
          >
            <template slot-scope="scope">
              <el-button
                class="btn-detail"
                size="mini"
                icon="el-icon-view"
                @click="showDetail(scope.row)"
              >
                {{ $t('common.table.detail') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <Pagination
        class="pagination"
        ref="pagination"
        :listApi="listApi"
        @refreshData="refreshData"
      />
    </div>

    <ModelUserModal
      :visible.sync="userModalVisible"
      :params="params"
      :model-info="currentRow"
      :model-map="modelMap"
    />
    <ModelAppModal
      :visible.sync="appModalVisible"
      :params="params"
      :model-info="currentRow"
      :model-map="modelMap"
    />
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import { formatAmount, resDownloadFile } from '@/utils/util.js';
import { fetchModelList, exportModelData } from '@/api/statisticsDashboard';
import ModelUserModal from './modelUserModal.vue';
import ModelAppModal from './modelAppModal.vue';
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
  components: { Pagination, ModelUserModal, ModelAppModal },
  props: {
    params: {},
    modelMap: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      listApi: fetchModelList,
      loading: false,
      tableData: [],
      providerObj: PROVIDER_OBJ,
      appTypeObj: AppType,
      type: 'list',
      sortField: '',
      sortOrder: '',
      userModalVisible: false,
      appModalVisible: false,
      currentRow: null,
    };
  },
  methods: {
    formatTime(val, type) {
      if (!val) return '0';
      const num = Number(val);
      if (type === 'avgCosts' && num >= 1000) {
        return (num / 1000).toFixed(1) + 's';
      }
      return num + 'ms';
    },
    formatAmount,
    handleRadio(val) {
      this.type = val;
      this.sortField = '';
      this.sortOrder = '';
      this.$nextTick(() => {
        const table = this.$refs.listTable || this.$refs.recordTable;
        if (table) {
          table.clearSort();
        }
      });
      this.getTableData({ ...this.params, pageNo: 1 });
    },
    handleSortChange({ prop, order }) {
      this.sortField = prop || '';
      this.sortOrder =
        order === 'ascending' ? 'asc' : order === 'descending' ? 'desc' : '';
      this.getTableData({ ...this.params, pageNo: 1 });
    },
    async getTableData(params) {
      if (this.$refs.pagination) {
        this.loading = true;
        try {
          this.tableData = await this.$refs.pagination.getTableData({
            ...params,
            type: this.type,
            sortField: this.sortField,
            sortOrder: this.sortOrder,
          });
        } finally {
          this.loading = false;
        }
      }
    },
    refreshData(data) {
      this.tableData = data;
    },
    async exportData() {
      const response = await exportModelData(this.params, this.type);
      const fileName =
        this.type === 'list'
          ? this.$t('statisticsDashboard.modelStatistics')
          : this.$t('statisticsDashboard.modelDetail');
      resDownloadFile(response, `${fileName}.xlsx`);
    },
    getModelTypeName(row) {
      const modelId = row.modelId || row.model;
      const modelInfo = this.modelMap[modelId] || {};
      return MODEL_TYPE_OBJ[modelInfo.modelType || row.modelType] || '--';
    },
    getModelTypeTagClass(row) {
      const modelId = row.modelId || row.model;
      const modelInfo = this.modelMap[modelId] || {};
      const type = modelInfo.modelType || row.modelType || '';
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
    getPublisher(row) {
      if (row.publisher) {
        return row.publisher;
      }
      const orgName = row.orgName || '';
      const userName = row.userName || '';
      if (orgName && userName) {
        return `${orgName} ${userName}`;
      }
      return orgName || userName || '--';
    },
    showDetail() {
      this.$message.info(this.$t('statisticsDashboard.detailTitle'));
    },
    showUserDetail(row) {
      this.currentRow = row;
      this.userModalVisible = true;
    },
    showAppDetail(row) {
      this.currentRow = row;
      this.appModalVisible = true;
    },
  },
};
</script>

<style lang="scss" scoped>
.table-wrap {
  margin-top: 24px;
  .add-bt {
    margin: 0 0 16px;
    float: right;
  }
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

.btn-user {
  color: #366fed;
  border: 1px solid #dcebfe;
  background: #eff6ff;
  padding: 5px 8px;

  &:hover,
  &:focus {
    color: #366fed;
    border: 1px solid #dcebfe;
    background: #eff6ff !important;
  }
}

.btn-app {
  color: #a55fef;
  border: 1px solid #eedfff;
  background: #faf5ff;
  padding: 5px 8px;

  &:hover,
  &:focus {
    color: #a55fef;
    border: 1px solid #eedfff;
    background: #faf5ff !important;
  }
}

.btn-detail {
  color: #5951e7;
  border: 1px solid #d2dafe;
  background: #eef2ff;
  padding: 5px 8px;

  &:hover,
  &:focus {
    color: #5951e7;
    border: 1px solid #d2dafe;
    background: #eef2ff !important;
  }
}
</style>
