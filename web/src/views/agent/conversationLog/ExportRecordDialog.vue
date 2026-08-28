<template>
  <el-dialog
    :visible.sync="visible"
    :close-on-click-modal="false"
    top="10vh"
    width="70%"
    @closed="handleClosed"
  >
    <template #title>
      <h1 class="dialog-title">{{ $t('agent.log.exportRecord.title') }}</h1>
    </template>

    <div class="export-record-filter">
      <el-date-picker
        v-model="queryParams.dateRange"
        type="daterange"
        size="mini"
        value-format="yyyy-MM-dd"
        :range-separator="$t('agent.log.dateRangeSeparator')"
        :start-placeholder="$t('agent.log.startDate')"
        :end-placeholder="$t('agent.log.endDate')"
      />
      <el-button type="primary" size="mini" @click="handleSearch">
        {{ $t('agent.log.search') }}
      </el-button>
    </div>

    <el-table v-loading="loading" :data="tableData" border>
      <el-table-column
        prop="exportTime"
        :label="$t('agent.log.exportRecord.exportTime')"
        min-width="180"
        show-overflow-tooltip
      />
      <el-table-column
        prop="fileName"
        :label="$t('agent.log.exportRecord.fileName')"
        min-width="180"
        show-overflow-tooltip
      />
      <el-table-column
        prop="author"
        :label="$t('agent.log.exportRecord.author')"
        min-width="140"
        show-overflow-tooltip
      />
      <el-table-column
        :label="$t('agent.log.exportRecord.status')"
        min-width="140"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ getStatusText(row.status) }}
        </template>
      </el-table-column>
      <el-table-column
        :label="$t('agent.log.exportRecord.action')"
        fixed="right"
        width="150"
      >
        <template slot-scope="{ row }">
          <el-button
            type="text"
            size="mini"
            :disabled="!row.filePath || !isExportFinished(row.status)"
            @click="handleDownload(row)"
          >
            {{ $t('agent.log.exportRecord.download') }}
          </el-button>
          <el-divider direction="vertical" />
          <el-button
            type="text"
            size="mini"
            :disabled="isExportProcessing(row.status)"
            @click="handleDelete(row)"
          >
            {{ $t('agent.log.exportRecord.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div v-if="pagination.total" class="pagination-wrapper">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :current-page="pagination.pageNo"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @current-change="handlePageChange"
      />
    </div>
  </el-dialog>
</template>

<script>
import {
  deleteConversationLogExportRecords,
  getConversationLogExportRecordList,
} from '@/api/chatLog';

const EXPORT_STATUS = {
  PENDING: 0,
  PROCESSING: 1,
  FINISHED: 2,
  FAILED: 3,
};

const getCurrentDate = () => {
  const date = new Date();
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

export default {
  name: 'AgentConversationLogExportRecordDialog',
  data() {
    return {
      visible: false,
      loading: false,
      appId: '',
      appType: '',
      queryParams: this.createDefaultQueryParams(),
      tableData: [],
      pagination: {
        pageNo: 1,
        pageSize: 10,
        total: 0,
      },
    };
  },
  methods: {
    showDialog({ appId, appType }) {
      this.appId = appId;
      this.appType = appType;
      this.queryParams = this.createDefaultQueryParams();
      this.pagination.pageNo = 1;
      this.visible = true;
      this.fetchRecordList();
    },
    async fetchRecordList() {
      if (!this.appId || !this.appType) return;

      const [startDate = '', endDate = ''] = this.queryParams.dateRange || [];
      this.loading = true;
      try {
        const res = await getConversationLogExportRecordList({
          appId: this.appId,
          appType: this.appType,
          startDate,
          endDate,
          pageNo: this.pagination.pageNo,
          pageSize: this.pagination.pageSize,
        });
        if (res && res.code === 0) {
          const data = res.data || {};
          this.tableData = data.list || [];
          this.pagination.total = data.total || 0;
        } else {
          this.$message.error(res?.msg || this.$t('common.message.error'));
        }
      } catch (error) {
        this.$message.error(error?.message || this.$t('common.message.error'));
      } finally {
        this.loading = false;
      }
    },
    handlePageChange(pageNo) {
      this.pagination.pageNo = pageNo;
      this.fetchRecordList();
    },
    handleSearch() {
      this.pagination.pageNo = 1;
      this.fetchRecordList();
    },
    createDefaultQueryParams() {
      const currentDate = getCurrentDate();
      return {
        dateRange: [currentDate, currentDate],
      };
    },
    getStatusText(status) {
      const statusKey = {
        [EXPORT_STATUS.PENDING]: 'pending',
        [EXPORT_STATUS.PROCESSING]: 'processing',
        [EXPORT_STATUS.FINISHED]: 'finished',
        [EXPORT_STATUS.FAILED]: 'failed',
      }[status];
      return statusKey
        ? this.$t(`agent.log.exportRecord.statusOptions.${statusKey}`)
        : '-';
    },
    isExportFinished(status) {
      return status === EXPORT_STATUS.FINISHED;
    },
    isExportProcessing(status) {
      return status === EXPORT_STATUS.PROCESSING;
    },
    handleDownload(row) {
      if (!row.filePath || !this.isExportFinished(row.status)) return;
      window.open(row.filePath, '_blank');
    },
    handleDelete(row) {
      if (this.isExportProcessing(row.status)) return;
      this.$confirm(
        this.$t('agent.log.exportRecord.deleteConfirm'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(async () => {
          const res = await deleteConversationLogExportRecords({
            exportRecordIds: [row.exportRecordId],
          });
          if (res && res.code === 0) {
            this.$message.success(this.$t('common.info.delete'));
            if (this.tableData.length === 1 && this.pagination.pageNo > 1) {
              this.pagination.pageNo -= 1;
            }
            this.fetchRecordList();
          } else {
            this.$message.error(res?.msg || this.$t('common.message.error'));
          }
        })
        .catch(() => {});
    },
    handleClosed() {
      this.tableData = [];
      this.pagination.total = 0;
    },
  },
};
</script>

<style scoped lang="scss">
.pagination-wrapper {
  margin-top: 16px;
  text-align: right;
}

.export-record-filter {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.dialog-title {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
  color: #1f2329;
}
</style>
