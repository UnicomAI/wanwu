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
            {{ $t('statisticsDashboard.tabStatistics') }}
          </el-radio-button>
          <el-radio-button label="record">
            {{ $t('statisticsDashboard.tabDetail') }}
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
            prop="source"
            :label="$t('statisticsDashboard.source')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.sourceName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="module"
            :label="$t('statisticsDashboard.module')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.moduleName || '--' }}
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
              {{ appTypeObj[scope.row.appType] || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="moduleCreatorUserName"
            :label="$t('statisticsDashboard.appAuthor')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.moduleCreatorUserName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="moduleCreatorOrgName"
            :label="$t('statisticsDashboard.fromOrg')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.moduleCreatorOrgName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="avgFirstTokenLatency"
            :label="$t('statisticsDashboard.avgFirstTokenLatency')"
            align="left"
            sortable="custom"
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
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.nonStreamCount) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="callCount"
            :label="$t('statisticsDashboard.callCountTotal')"
            align="left"
            sortable="custom"
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
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.failureRate) }}%
            </template>
          </el-table-column>
          <el-table-column
            width="160"
            align="center"
            :label="$t('common.table.operation')"
          >
            <template slot-scope="scope">
              <el-button
                class="btn-user"
                size="mini"
                icon="el-icon-user"
                @click="showUserModal(scope.row)"
              >
                {{ $t('common.button.user') }}
              </el-button>
              <el-button
                class="btn-model"
                size="mini"
                icon="el-icon-menu"
                @click="showModelModal(scope.row)"
              >
                {{ $t('common.button.model') }}
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
            prop="source"
            :label="$t('statisticsDashboard.source')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.sourceName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="module"
            :label="$t('statisticsDashboard.module')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.moduleName || '--' }}
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
              {{ appTypeObj[scope.row.appType] || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="userName"
            :label="$t('statisticsDashboard.user')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.userName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="costs"
            :label="$t('statisticsDashboard.costs')"
            align="left"
          >
            <template slot-scope="scope">
              {{ formatSec(scope.row.costs) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="firstTokenLatency"
            :label="$t('statisticsDashboard.firstTokenLatency')"
            align="left"
          >
            <template slot-scope="scope">
              {{ formatSec(scope.row.firstTokenLatency) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="calledAt"
            :label="$t('statisticsDashboard.callTime')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.calledAt || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="isSuccess"
            :label="$t('statisticsDashboard.status')"
            align="left"
            width="70"
          >
            <template slot-scope="scope">
              <el-tag
                v-if="scope.row.isSuccess"
                type="success"
                size="small"
                class="status-tag"
              >
                {{ $t('statisticsDashboard.success') }}
              </el-tag>
              <el-tag
                v-else-if="scope.row.isSuccess === false"
                type="danger"
                size="small"
                class="status-tag"
              >
                {{ $t('statisticsDashboard.error') }}
              </el-tag>
              <div v-else>--</div>
            </template>
          </el-table-column>
          <el-table-column
            width="100"
            align="center"
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
    <AppUserModal
      :visible.sync="userModalVisible"
      :params="params"
      :app-info="currentRow"
    />
    <AppModelModal
      :visible.sync="modelModalVisible"
      :params="params"
      :app-info="currentRow"
    />
    <AppRecordDetail :visible.sync="detailVisible" :row="currentRow" />
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import { formatAmount, formatSec, resDownloadFile } from '@/utils/util.js';
import { fetchAppList, exportAppData } from '@/api/statisticsDashboard';
import { TotalTypeObj } from '@/utils/commonSet';
import AppUserModal from './appUserModal.vue';
import AppModelModal from './appModelModal.vue';
import AppRecordDetail from './appRecordDetail.vue';

export default {
  components: { Pagination, AppUserModal, AppModelModal, AppRecordDetail },
  props: {
    params: {},
  },
  data() {
    return {
      listApi: fetchAppList,
      loading: false,
      tableData: [],
      appTypeObj: TotalTypeObj,
      type: 'list',
      sortField: '',
      sortOrder: '',
      userModalVisible: false,
      modelModalVisible: false,
      detailVisible: false,
      currentRow: null,
    };
  },
  methods: {
    formatAmount,
    formatSec,
    handleRadio(val) {
      this.type = val;
      this.sortField = '';
      this.sortOrder = '';
      // 列表数据初始化
      this.tableData = [];
      if (this.$refs.pagination) this.$refs.pagination.total = 0;
      this.$nextTick(() => {
        const table = this.$refs.listTable || this.$refs.recordTable;
        if (table) {
          table.clearSort();
        }
      });
      this.getTableData({ ...this.params, pageNo: 1 });
    },
    formatParams(params) {
      return {
        ...params,
        sortField: this.sortField,
        sortOrder: this.sortOrder,
      };
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
          this.tableData = await this.$refs.pagination.getTableData(
            this.formatParams({
              ...params,
              type: this.type,
            }),
          );
        } finally {
          this.loading = false;
        }
      }
    },
    refreshData(data) {
      this.tableData = data;
    },
    showUserModal(row) {
      this.currentRow = row;
      this.userModalVisible = true;
    },
    showModelModal(row) {
      this.currentRow = row;
      this.modelModalVisible = true;
    },
    showDetail(row) {
      this.currentRow = row;
      this.detailVisible = true;
    },
    async exportData() {
      const response = await exportAppData(
        this.formatParams(this.params),
        this.type,
      );
      resDownloadFile(
        response,
        this.type === 'list'
          ? `${this.$t('statisticsDashboard.appStatistics')}.xlsx`
          : `${this.$t('statisticsDashboard.appDetail')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsTag.scss';
.table-wrap {
  margin-top: 24px;
  .add-bt {
    margin: 0 0 16px;
    float: right;
  }
}
</style>
