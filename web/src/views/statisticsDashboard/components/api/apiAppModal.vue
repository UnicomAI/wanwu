<template>
  <el-dialog
    :title="$t('statisticsDashboard.apiAppUsageStats')"
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
          prop="name"
          :label="$t('statisticsDashboard.apiName')"
          align="left"
          min-width="140"
        >
          <template slot-scope="scope">
            {{ scope.row.apiName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="apiKey"
          :label="'API Key'"
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
            <AppTypeTag :app-type="scope.row.module" />
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
          :label="$t('statisticsDashboard.appAuthor')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.moduleCreatorUserName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="orgName"
          :label="$t('statisticsDashboard.appAuthorOrg')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.moduleCreatorOrgName || '--' }}
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
          min-width="120"
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
          min-width="120"
        >
          <template slot-scope="scope">
            {{ formatSec(scope.row.avgFirstTokenLatency) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="streamCount"
          :label="$t('statisticsDashboard.streamCount')"
          align="left"
          sortable="custom"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ formatAmount(scope.row.streamCount) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="nonStreamCount"
          :label="$t('statisticsDashboard.nonStreamCount')"
          align="left"
          sortable="custom"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ formatAmount(scope.row.nonStreamCount) }}
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
import AppTypeTag from '../app/appTypeTag.vue';
import { formatAmount, formatSec, resDownloadFile } from '@/utils/util.js';
import { fetchApiAppList, exportApiAppData } from '@/api/statisticsDashboard';

export default {
  components: { Pagination, AppTypeTag },
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
      listApi: fetchApiAppList,
      loading: false,
      tableData: [],
      sortField: '',
      sortOrder: '',
    };
  },
  computed: {
    apiAppParams() {
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
            ...this.apiAppParams,
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
      const response = await exportApiAppData(this.apiAppParams);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.apiStatistics')}_${this.$t('statisticsDashboard.apiAppUsageStats')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsTag.scss';
</style>
