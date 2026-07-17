<template>
  <el-dialog
    :title="$t('statisticsDashboard.apiModelUsageStats')"
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
        @sort-change="handleSortChange"
      >
        <el-table-column
          prop="apiName"
          :label="$t('statisticsDashboard.apiName')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.apiName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="apiKey"
          label="API Key"
          align="left"
          min-width="160"
        >
          <template slot-scope="scope">
            <span>
              {{
                scope.row.apiKey
                  ? scope.row.apiKey.slice(0, 6) + '******'
                  : '--'
              }}
            </span>
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
          prop="userName"
          :label="$t('statisticsDashboard.userName')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.userName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="methodPath"
          :label="$t('statisticsDashboard.apiPath')"
          align="left"
          min-width="200"
        >
          <template slot-scope="scope">
            {{ scope.row.methodPath || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="model"
          :label="$t('statisticsDashboard.modelName')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.model || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="provider"
          :label="$t('statisticsDashboard.provider')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ providerObj[scope.row.provider] || scope.row.provider || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="modelType"
          :label="$t('statisticsDashboard.modelType')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            <ModelTypeTag :model-type="scope.row.modelType" />
          </template>
        </el-table-column>
        <el-table-column
          prop="modelCreatorUserName"
          :label="$t('statisticsDashboard.modelPublisher')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.modelCreatorUserName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="modelCreatorOrgName"
          :label="$t('statisticsDashboard.fromOrg')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.modelCreatorOrgName || '--' }}
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
            {{ formatSec(scope.row.avgCosts) }}
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
            {{ formatSec(scope.row.avgFirstTokenLatency) }}
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
import ModelTypeTag from '../model/modelTypeTag.vue';
import { formatAmount, formatSec, resDownloadFile } from '@/utils/util.js';
import {
  fetchApiModelList,
  exportApiModelData,
} from '@/api/statisticsDashboard';
import { PROVIDER_OBJ } from '@/views/modelAccess/constants';

export default {
  components: { Pagination, ModelTypeTag },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    params: {
      type: Object,
      default: () => ({}),
    },
    apiInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      listApi: fetchApiModelList,
      loading: false,
      tableData: [],
      providerObj: PROVIDER_OBJ,
      sortField: '',
      sortOrder: '',
    };
  },
  computed: {
    apiModelParams() {
      return {
        ...this.params,
        sortField: this.sortField,
        sortOrder: this.sortOrder,
        apiKeyId: this.apiInfo?.apiKeyId || this.apiInfo?.keyId || '',
        methodPath: this.apiInfo?.methodPath || '',
      };
    },
  },
  methods: {
    formatAmount,
    formatSec,
    handleOpen() {
      this.$nextTick(() => {
        this.initTableData();
      });
    },
    handleClose() {
      this.$emit('update:visible', false);
      this.tableData = [];
    },
    initTableData() {
      this.getTableData({ pageNo: 1 });
    },
    async getTableData(params) {
      if (this.$refs.pagination) {
        this.loading = true;
        try {
          this.tableData = await this.$refs.pagination.getTableData({
            ...this.apiModelParams,
            ...params,
          });
        } finally {
          this.loading = false;
        }
      }
    },
    refreshData(data) {
      this.tableData = data;
    },
    handleSortChange({ prop, order }) {
      this.sortField = prop || '';
      this.sortOrder =
        order === 'ascending' ? 'asc' : order === 'descending' ? 'desc' : '';
      this.initTableData();
    },
    async exportData() {
      const response = await exportApiModelData(this.apiModelParams);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.apiStatistics')}_${this.$t('statisticsDashboard.apiModelUsageStats')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsTag.scss';
</style>
