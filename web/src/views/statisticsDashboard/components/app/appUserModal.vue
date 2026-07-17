<template>
  <el-dialog
    :title="$t('statisticsDashboard.userUsageStats')"
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
          prop="userName"
          :label="$t('statisticsDashboard.user')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.userName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="orgName"
          :label="$t('statisticsDashboard.fromOrg')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.orgName || '--' }}
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
          prop="failureRate"
          :label="$t('statisticsDashboard.failureRate')"
          align="left"
          sortable="custom"
          min-width="110"
        >
          <template slot-scope="scope">{{ scope.row.failureRate }}%</template>
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
import AppTypeTag from './appTypeTag.vue';
import { formatAmount, formatSec, resDownloadFile } from '@/utils/util.js';
import { fetchAppUserList, exportAppUserData } from '@/api/statisticsDashboard';

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
    appInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      listApi: fetchAppUserList,
      loading: false,
      tableData: [],
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
    appUserParams() {
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
            ...this.appUserParams,
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
      const response = await exportAppUserData(this.appUserParams);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.appStatistics')}_${this.$t('statisticsDashboard.userUsageStats')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsTag.scss';
</style>
