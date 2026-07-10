<template>
  <div>
    <div class="table-wrap list-common wrap-fullheight">
      <div class="table-box" style="margin-top: 36px">
        <el-button
          class="add-bt"
          size="mini"
          type="primary"
          @click="exportData"
        >
          <span>{{ $t('common.button.export') }}</span>
        </el-button>
        <el-radio-group v-model="type" size="mini" @change="handleRadio">
          <el-radio-button :label="'list'">
            {{ $t('statisticsDashboard.apiStatistics') }}
          </el-radio-button>
          <el-radio-button :label="'record'">
            {{ $t('statisticsDashboard.apiDetail') }}
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
            prop="name"
            :label="'API Key' + $t('statisticsDashboard.name')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.name || '--' }}
            </template>
          </el-table-column>
          <el-table-column prop="apiKey" label="API Key" align="left">
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
            prop="methodPath"
            :label="$t('statisticsDashboard.apiPath')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.methodPath || '--' }}
            </template>
          </el-table-column>
          <!--<el-table-column
            prop="model"
            :label="$t('statisticsDashboard.model')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.model || '&#45;&#45;' }}
            </template>
          </el-table-column>-->
          <el-table-column
            prop="callCount"
            :label="$t('statisticsDashboard.appCallCount')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.callCount) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="callFailure"
            :label="$t('statisticsDashboard.appCallFailure')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.callFailure) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="avgStreamCosts"
            :label="$t('statisticsDashboard.avgStreamCosts')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgStreamCosts) }}ms
            </template>
          </el-table-column>
          <el-table-column
            prop="avgNonStreamCosts"
            :label="$t('statisticsDashboard.avgCosts')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgNonStreamCosts) }}ms
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
            width="90"
            align="center"
            :label="$t('common.table.operation')"
          >
            <template slot-scope="scope">
              <el-button
                class="btn-app"
                size="mini"
                icon="el-icon-s-grid"
                @click="showAppModal(scope.row)"
              >
                {{ $t('common.button.app') }}
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
            prop="name"
            :label="'API Key' + $t('statisticsDashboard.name')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.name || '--' }}
            </template>
          </el-table-column>
          <el-table-column prop="apiKey" label="API Key" align="left">
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
            prop="methodPath"
            :label="$t('statisticsDashboard.apiPath')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.methodPath || '--' }}
            </template>
          </el-table-column>
          <!--<el-table-column
            prop="model"
            :label="$t('statisticsDashboard.model')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.model || '&#45;&#45;' }}
            </template>
          </el-table-column>-->
          <el-table-column
            prop="callTime"
            :label="$t('statisticsDashboard.callTime')"
            align="left"
            sortable="custom"
          ></el-table-column>
          <el-table-column
            prop="responseStatus"
            :label="$t('statisticsDashboard.responseStatus')"
            align="left"
          ></el-table-column>
          <el-table-column
            prop="streamCosts"
            :label="$t('statisticsDashboard.streamCosts')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.streamCosts) }}ms
            </template>
          </el-table-column>
          <el-table-column
            prop="nonStreamCosts"
            :label="$t('statisticsDashboard.nonStreamCosts')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.nonStreamCosts) }}ms
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
    <ApiAppModal
      :visible.sync="appModalVisible"
      :params="params"
      :api-info="currentRow"
    />
    <RecordDetail :visible.sync="detailVisible" :row="currentRow" />
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import { formatAmount, resDownloadFile } from '@/utils/util.js';
import { fetchApiList, exportApiData } from '@/api/statisticsDashboard';
import { PROVIDER_OBJ } from '@/views/modelAccess/constants';
import RecordDetail from './recordDetail.vue';
import ApiAppModal from './apiAppModal.vue';

export default {
  components: { Pagination, RecordDetail, ApiAppModal },
  props: {
    params: {},
  },
  data() {
    return {
      listApi: fetchApiList,
      loading: false,
      tableData: [],
      providerObj: PROVIDER_OBJ,
      type: 'list',
      sortField: '',
      sortOrder: '',
      appModalVisible: false,
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
    showAppModal(row) {
      this.currentRow = row;
      this.appModalVisible = true;
    },
    showDetail(row) {
      this.currentRow = row;
      this.detailVisible = true;
    },
    refreshData(data) {
      this.tableData = data;
    },
    async exportData() {
      const response = await exportApiData(this.params, this.type);
      resDownloadFile(
        response,
        `${this.type === 'list' ? this.$t('statisticsDashboard.apiStatistics') : this.$t('statisticsDashboard.apiDetail')}.xlsx`,
      );
    },
  },
};
</script>

<style lang="scss" scoped>
.table-wrap {
  padding: 0 12px;
  .add-bt {
    margin: 0 0 16px;
    float: right;
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
