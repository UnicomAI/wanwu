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
          <el-radio-button :label="'list'">
            {{ $t('statisticsDashboard.tabStatistics') }}
          </el-radio-button>
          <el-radio-button :label="'record'">
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
            prop="apiName"
            :label="'API Key' + $t('statisticsDashboard.name')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.apiName || '--' }}
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
            :label="$t('statisticsDashboard.callCount')"
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
            prop="avgFirstTokenLatency"
            :label="$t('statisticsDashboard.avgFirstTokenLatency')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgFirstTokenLatency) }}ms
            </template>
          </el-table-column>
          <el-table-column
            prop="avgCosts"
            :label="$t('statisticsDashboard.avgCosts')"
            align="left"
            sortable="custom"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.avgCosts) }}ms
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
        >
          <el-table-column
            prop="apiName"
            :label="'API Key' + $t('statisticsDashboard.name')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.apiName || '--' }}
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
            prop="costs"
            :label="$t('statisticsDashboard.costs')"
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.costs) }}ms
            </template>
          </el-table-column>
          <el-table-column
            prop="firstTokenLatency"
            :label="$t('statisticsDashboard.firstTokenLatency')"
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.firstTokenLatency) }}ms
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
      const response = await exportApiData(
        this.formatParams(this.params),
        this.type,
      );
      resDownloadFile(
        response,
        `${this.type === 'list' ? this.$t('statisticsDashboard.apiStatistics') : this.$t('statisticsDashboard.apiDetail')}.xlsx`,
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
  .status-tag {
    font-weight: 500;
    border-radius: 12px;
  }
}
</style>
