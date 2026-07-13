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
            prop="orgName"
            :label="$t('statisticsDashboard.org')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.orgName || '--' }}
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
            prop="callCount"
            :label="
              $t('statisticsDashboard.appCallCount') +
              ` (${$t('statisticsDashboard.frequency')})`
            "
            align="left"
            sortable="custom"
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
          />
          <el-table-column
            prop="avgStreamCosts"
            :label="$t('statisticsDashboard.avgFirstCosts') + ` (ms)`"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgStreamCosts) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="avgNonStreamCosts"
            :label="$t('statisticsDashboard.avgCosts') + ` (ms)`"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgNonStreamCosts) }}
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
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.nonStreamCount) }}
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
                icon="el-icon-s-grid"
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
              {{ scope.row.source || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="module"
            :label="$t('statisticsDashboard.module')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.module || '--' }}
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
            :label="$t('statisticsDashboard.userName')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.userName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="modelName"
            :label="$t('statisticsDashboard.modelName')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.modelName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="streamCosts"
            :label="$t('statisticsDashboard.streamCosts') + ` (ms)`"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.streamCosts) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="nonStreamCosts"
            :label="$t('statisticsDashboard.nonStreamCosts') + ` (ms)`"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.nonStreamCosts) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="callTime"
            :label="$t('statisticsDashboard.callTime')"
            align="left"
            sortable="custom"
          />
          <el-table-column
            prop="status"
            :label="$t('statisticsDashboard.responseStatus')"
            align="left"
          >
            <template slot-scope="scope">
              <el-tag
                v-if="scope.row.status === '成功'"
                type="success"
                size="small"
              >
                {{ scope.row.status }}
              </el-tag>
              <el-tag v-else-if="scope.row.status" type="danger" size="small">
                {{ scope.row.status }}
              </el-tag>
              <span v-else>--</span>
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
import { formatAmount, resDownloadFile } from '@/utils/util.js';
import {
  fetchAppList,
  exportAppData,
  fetchAppRecordList,
  exportAppRecordData,
} from '@/api/statisticsDashboard';
import { AppType } from '@/utils/commonSet';
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
      appTypeObj: AppType,
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
      this.listApi = val === 'list' ? fetchAppList : fetchAppRecordList;
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
      if (this.type === 'list') {
        const response = await exportAppData(this.params);
        resDownloadFile(
          response,
          `${this.$t('statisticsDashboard.appStatistics')}.xlsx`,
        );
      } else {
        const response = await exportAppRecordData(this.params);
        resDownloadFile(
          response,
          `${this.$t('statisticsDashboard.appDetail')}.xlsx`,
        );
      }
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

.btn-model {
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
