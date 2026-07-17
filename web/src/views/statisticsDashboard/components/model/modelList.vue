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
            prop="modelCreatorUserName"
            :label="$t('statisticsDashboard.publisher')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.modelCreatorUserName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="modelCreatorOrgName"
            :label="$t('statisticsDashboard.fromOrg')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.modelCreatorOrgName || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="modelType"
            :label="$t('modelAccess.table.modelType')"
            align="left"
            width="110px"
          >
            <template slot-scope="scope">
              <ModelTypeTag :model-type="scope.row.modelType" />
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
              {{ formatSec(scope.row.avgCosts) }}
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
            width="160"
            align="center"
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
              <ModelTypeTag :model-type="scope.row.modelType" />
            </template>
          </el-table-column>
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
              <AppTypeTag :app-type="scope.row.module" />
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
              <AppTypeTag :app-type="scope.row.appType" />
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
            prop="totalTokens"
            :label="$t('statisticsDashboard.totalTokens')"
            align="left"
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
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.promptTokens) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="completionTokens"
            :label="$t('statisticsDashboard.completionTokens')"
            align="left"
          >
            <template slot-scope="scope">
              {{ formatAmount(scope.row.completionTokens) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="callTime"
            :label="$t('statisticsDashboard.callTime')"
            align="left"
          >
            <template slot-scope="scope">
              {{ scope.row.calledAt || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="status"
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
            width="110"
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
                {{ $t('common.table.showDetail') }}
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
    />
    <ModelAppModal
      :visible.sync="appModalVisible"
      :params="params"
      :model-info="currentRow"
    />
    <ModelRecordDetail :visible.sync="detailVisible" :row="currentRow" />
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import { formatAmount, formatSec, resDownloadFile } from '@/utils/util.js';
import { fetchModelList, exportModelData } from '@/api/statisticsDashboard';
import ModelUserModal from './modelUserModal.vue';
import ModelAppModal from './modelAppModal.vue';
import ModelRecordDetail from './modelRecordDetail.vue';
import ModelTypeTag from './modelTypeTag.vue';
import AppTypeTag from '../app/appTypeTag.vue';
import { PROVIDER_OBJ } from '@/views/modelAccess/constants';

export default {
  components: {
    Pagination,
    ModelUserModal,
    ModelAppModal,
    ModelRecordDetail,
    ModelTypeTag,
    AppTypeTag,
  },
  props: {
    params: {},
  },
  data() {
    return {
      listApi: fetchModelList,
      loading: false,
      tableData: [],
      providerObj: PROVIDER_OBJ,
      type: 'list',
      sortField: '',
      sortOrder: '',
      userModalVisible: false,
      appModalVisible: false,
      detailVisible: false,
      currentRow: null,
    };
  },
  methods: {
    formatSec,
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
    showDetail(row) {
      this.currentRow = row;
      this.detailVisible = true;
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
@import '@/style/statisticsTag.scss';
.table-wrap {
  margin-top: 24px;
  .add-bt {
    margin: 0 0 16px;
    float: right;
  }
}
</style>
