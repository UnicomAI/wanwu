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
      >
        <el-table-column
          prop="source"
          :label="$t('statisticsDashboard.source')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.source || '--' }}
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
          prop="callCount"
          :label="
            $t('statisticsDashboard.appCallCount') +
            ` (${$t('statisticsDashboard.frequency')})`
          "
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
          :label="
            $t('statisticsDashboard.appCallFailure') +
            ` (${$t('statisticsDashboard.frequency')})`
          "
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
          prop="avgStreamCosts"
          :label="$t('statisticsDashboard.avgFirstCosts') + ` (ms)`"
          align="left"
          sortable="custom"
          min-width="170"
        >
          <template slot-scope="scope">
            {{ formatTime(scope.row.avgStreamCosts, 'avgStreamCosts') }}
          </template>
        </el-table-column>
        <el-table-column
          prop="avgNonStreamCosts"
          :label="$t('statisticsDashboard.avgCosts') + ` (ms)`"
          align="left"
          sortable="custom"
          min-width="150"
        >
          <template slot-scope="scope">
            {{ formatTime(scope.row.avgNonStreamCosts, 'avgCosts') }}
          </template>
        </el-table-column>
        <el-table-column
          prop="streamCount"
          :label="
            $t('statisticsDashboard.streamCount') +
            ` (${$t('statisticsDashboard.frequency')})`
          "
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
          :label="
            $t('statisticsDashboard.nonStreamCount') +
            ` (${$t('statisticsDashboard.frequency')})`
          "
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
import { formatAmount, resDownloadFile } from '@/utils/util.js';
import { fetchAppUserList, exportAppUserData } from '@/api/statisticsDashboard';

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
    modalParams() {
      const { apps, ...rest } = this.params || {};
      return {
        ...rest,
        appId: this.appInfo?.appId || this.appInfo?.app || '',
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
      const response = await exportAppUserData(this.modalParams);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.userUsageStats')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
.modal-toolbar {
  margin-bottom: 16px;
}
</style>
