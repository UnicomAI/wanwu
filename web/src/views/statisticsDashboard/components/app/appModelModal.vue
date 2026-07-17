<template>
  <el-dialog
    :title="$t('statisticsDashboard.modelUsageStats')"
    :visible.sync="dialogVisible"
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
          prop="source"
          :label="$t('statisticsDashboard.source')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.sourceName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="module"
          :label="$t('statisticsDashboard.module')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.moduleName || '--' }}
          </template>
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
            <AppTypeTag :app-type="scope.row.appType" />
          </template>
        </el-table-column>
        <el-table-column
          prop="moduleCreatorUserName"
          :label="$t('statisticsDashboard.author')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.moduleCreatorUserName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="moduleCreatorOrgName"
          :label="$t('statisticsDashboard.authorOrg')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.moduleCreatorOrgName || '--' }}
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
          min-width="150"
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
import AppTypeTag from './appTypeTag.vue';
import { formatAmount, formatSec, resDownloadFile } from '@/utils/util.js';
import {
  fetchAppModelList,
  exportAppModelData,
} from '@/api/statisticsDashboard';
import { PROVIDER_OBJ } from '@/views/modelAccess/constants';

export default {
  components: { Pagination, AppTypeTag, ModelTypeTag },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    params: {
      type: Object,
      default: () => ({}),
    },
    appInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      listApi: fetchAppModelList,
      loading: false,
      tableData: [],
      providerObj: PROVIDER_OBJ,
      sortField: '',
      sortOrder: '',
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
    appModelParams() {
      const { appId, source, module, moduleCreatorOrgId, moduleCreatorUserId } =
        this.appInfo || {};
      return {
        ...this.params,
        sortField: this.sortField,
        sortOrder: this.sortOrder,
        appId,
        source,
        module,
        moduleCreatorOrgId,
        moduleCreatorUserId,
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
            ...this.appModelParams,
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
      const response = await exportAppModelData(this.appModelParams);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.appStatistics')}_${this.$t('statisticsDashboard.modelUsageStats')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsTag.scss';
</style>
