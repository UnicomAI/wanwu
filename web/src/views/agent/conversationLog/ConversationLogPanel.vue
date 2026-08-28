<template>
  <div class="conversation-log-panel">
    <div class="action-bar">
      <div class="action-bar__fields">
        <div class="action-bar__field action-bar__field--title">
          <div class="action-bar__label">
            {{ $t('agent.log.conversationTitle') }}
          </div>
          <el-input
            v-model="queryParams.name"
            class="action-bar__control"
            prefix-icon="el-icon-search"
            clearable
            :placeholder="$t('agent.log.conversationTitlePlaceholder')"
            @keyup.enter.native="handleSearch"
          />
        </div>

        <div class="action-bar__field">
          <div class="action-bar__label">{{ $t('agent.log.source') }}</div>
          <el-select
            v-model="queryParams.source"
            class="action-bar__control"
            multiple
            collapse-tags
            :placeholder="$t('agent.log.sourcePlaceholder')"
            @change="handleMultiSelectChange('source', $event)"
          >
            <el-option
              v-for="item in sourceOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </div>

        <div class="action-bar__field action-bar__field--date">
          <div class="action-bar__label">{{ $t('agent.log.date') }}</div>
          <el-date-picker
            v-model="queryParams.dateRange"
            class="action-bar__control"
            type="daterange"
            value-format="yyyy-MM-dd"
            :range-separator="$t('agent.log.dateRangeSeparator')"
            :start-placeholder="$t('agent.log.startDate')"
            :end-placeholder="$t('agent.log.endDate')"
            size="mini"
          />
        </div>

        <div class="action-bar__field action-bar__field--user">
          <div class="action-bar__label">{{ $t('agent.log.user') }}</div>
          <el-select
            v-model="queryParams.userIds"
            class="action-bar__control"
            multiple
            filterable
            collapse-tags
            :loading="userLoading"
            :placeholder="$t('agent.log.userPlaceholder')"
            @change="handleMultiSelectChange('userIds', $event)"
          >
            <el-option
              v-for="item in userSelectOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </div>
      </div>

      <div class="action-bar__actions">
        <el-button type="primary" size="mini" @click="handleSearch">
          {{ $t('agent.log.search') }}
        </el-button>
        <el-button size="mini" @click="handleReset">
          {{ $t('agent.log.reset') }}
        </el-button>
        <el-dropdown
          class="action-bar__export"
          trigger="hover"
          @command="handleExportCommand"
        >
          <el-button type="primary" size="mini" :loading="exportLoading">
            {{ $t('agent.log.export') }}
            <i class="el-icon-arrow-down el-icon--right" />
          </el-button>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item command="all">
              {{ $t('agent.log.exportAll') }}
            </el-dropdown-item>
            <el-dropdown-item command="selected">
              {{ $t('agent.log.exportSelected') }}
            </el-dropdown-item>
            <el-dropdown-item command="records">
              {{ $t('agent.log.exportRecord.title') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </div>
    </div>
    <el-table
      ref="logTable"
      v-loading="listLoading"
      :data="tableData"
      class="main-table"
      row-key="logId"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <el-table-column type="selection" :reserve-selection="true" width="48" />
      <el-table-column
        :label="$t('agent.log.columns.source')"
        width="100"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ getSourceLabel(row.source) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="version"
        :label="$t('agent.log.columns.version')"
        width="100"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.version) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="userId"
        :label="$t('agent.log.columns.user')"
        width="120"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.userId) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="title"
        :label="$t('agent.log.columns.title')"
        min-width="180"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.title) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="conversationId"
        :label="$t('agent.log.columns.conversationId')"
        min-width="180"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.conversationId) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="messageCount"
        :label="$t('agent.log.columns.messageCount')"
        width="100"
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.messageCount) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="createAt"
        :label="$t('agent.log.columns.createAt')"
        width="180"
        sortable="custom"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.createAt) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="updateAt"
        :label="$t('agent.log.columns.updateAt')"
        width="180"
        sortable="custom"
        show-overflow-tooltip
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.updateAt) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="avgCosts"
        :label="$t('agent.log.columns.avgCosts')"
        width="160"
        sortable="custom"
      >
        <template slot-scope="{ row }">
          {{ formatMilliseconds(row.avgCosts) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="avgFirstTokenLatency"
        :label="$t('agent.log.columns.avgFirstTokenLatency')"
        width="190"
        sortable="custom"
      >
        <template slot-scope="{ row }">
          {{ formatMilliseconds(row.avgFirstTokenLatency) }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('agent.log.columns.feedback')" width="140">
        <template slot-scope="{ row }">
          <div class="feedback-cell">
            <svg-icon icon-class="thumb-up" class="feedback-cell__icon" />
            <span>{{ displayValue(row.likeCount) }}</span>
            <svg-icon icon-class="thumb-down" class="feedback-cell__icon" />
            <span>{{ displayValue(row.dislikeCount) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column
        prop="errorCount"
        :label="$t('agent.log.columns.errorCount')"
        width="100"
      >
        <template slot-scope="{ row }">
          {{ displayValue(row.errorCount) }}
        </template>
      </el-table-column>
      <el-table-column
        :label="$t('agent.log.columns.action')"
        width="80"
        fixed="right"
      >
        <template slot-scope="{ row }">
          <el-button type="text" @click="handleDetail(row)">
            {{ $t('agent.log.detail') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      class="message-pagination"
      background
      layout="total, sizes, prev, pager, next, jumper"
      :current-page="pagination.pageNo"
      :page-size="pagination.pageSize"
      :page-sizes="[10, 20, 50, 100]"
      :total="pagination.total"
      @size-change="handlePageSizeChange"
      @current-change="handlePageChange"
    />
    <export-record-dialog ref="exportRecordDialog" />
    <log-detail-drawer
      ref="logDetailDrawer"
      :default-url="avatarPath"
      :get-detail="requestService.getDetail"
    />
  </div>
</template>

<script>
import { exportConversationLogs } from '@/api/chatLog';
import { CONVERSATION_LOG_SOURCE_OPTIONS } from './constants';
import ExportRecordDialog from './ExportRecordDialog.vue';
import LogDetailDrawer from './LogDetailDrawer.vue';
const ALL_OPTION_VALUE = '__all__';
const getCurrentDate = () => {
  const date = new Date();
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

export default {
  name: 'ConversationLogPanel',
  components: {
    ExportRecordDialog,
    LogDetailDrawer,
  },
  props: {
    appId: {
      type: String,
      required: true,
    },
    appType: {
      type: String,
      default: 'agent',
    },
    avatarPath: {
      type: String,
      default: '',
    },
    requestService: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      userLoading: false,
      userOptions: [],
      queryParams: this.createDefaultQueryParams(),
      multiSelectValues: {
        source: [ALL_OPTION_VALUE],
        userIds: [ALL_OPTION_VALUE],
      },
      listLoading: false,
      exportLoading: false,
      tableData: [],
      selectedLogs: [],
      pagination: {
        pageNo: 1,
        pageSize: 10,
        total: 0,
      },
    };
  },
  computed: {
    sourceOptions() {
      return [
        {
          label: this.$t('agent.log.all'),
          value: ALL_OPTION_VALUE,
        },
        ...CONVERSATION_LOG_SOURCE_OPTIONS.map(item => ({
          value: item.value,
          label: this.$t(`agent.log.sourceOptions.${item.labelKey}`),
        })),
      ];
    },
    userSelectOptions() {
      return [
        {
          label: this.$t('agent.log.all'),
          value: ALL_OPTION_VALUE,
        },
        ...this.userOptions,
      ];
    },
  },
  created() {
    this.loadUserOptions();
    this.loadLogList();
  },
  methods: {
    createDefaultQueryParams() {
      const currentDate = getCurrentDate();
      return {
        name: '',
        source: [ALL_OPTION_VALUE],
        dateRange: [currentDate, currentDate],
        userIds: [ALL_OPTION_VALUE],
        orderBy: '',
        orderType: '',
      };
    },
    async loadUserOptions() {
      if (!this.appId) return;

      this.userLoading = true;
      try {
        const res = await this.requestService.getUsers({
          appId: this.appId,
          appType: this.appType,
        });
        if (res && res.code === 0) {
          this.userOptions = ((res.data && res.data.users) || []).map(user => ({
            label: user.name,
            value: user.id,
          }));
        }
      } catch (error) {
        this.userOptions = [];
      } finally {
        this.userLoading = false;
      }
    },
    handleMultiSelectChange(field, values) {
      const previousValues = this.multiSelectValues[field] || [];
      let nextValues = values;

      if (!values.length) {
        nextValues = [ALL_OPTION_VALUE];
      } else if (values.includes(ALL_OPTION_VALUE) && values.length > 1) {
        nextValues = previousValues.includes(ALL_OPTION_VALUE)
          ? values.filter(value => value !== ALL_OPTION_VALUE)
          : [ALL_OPTION_VALUE];
      }

      this.$set(this.queryParams, field, nextValues);
      this.$set(this.multiSelectValues, field, [...nextValues]);
    },
    getQueryParams() {
      const [startDate = '', endDate = ''] = this.queryParams.dateRange || [];
      const normalizeMultiValue = values =>
        (values || []).filter(value => value !== ALL_OPTION_VALUE);

      return {
        appId: this.appId,
        appType: this.appType,
        name: this.queryParams.name,
        source: normalizeMultiValue(this.queryParams.source),
        userIds: normalizeMultiValue(this.queryParams.userIds),
        startDate,
        endDate,
        orderBy: this.queryParams.orderBy,
        orderType: this.queryParams.orderType,
        pageNo: this.pagination.pageNo,
        pageSize: this.pagination.pageSize,
      };
    },
    async loadLogList() {
      if (!this.appId) {
        this.tableData = [];
        this.pagination.total = 0;
        return;
      }

      this.listLoading = true;
      try {
        const res = await this.requestService.getList(this.getQueryParams());
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
        this.listLoading = false;
      }
    },
    handleSearch() {
      this.pagination.pageNo = 1;
      this.clearSelectedLogs();
      this.loadLogList();
    },
    handleReset() {
      this.queryParams = this.createDefaultQueryParams();
      this.multiSelectValues = {
        source: [ALL_OPTION_VALUE],
        userIds: [ALL_OPTION_VALUE],
      };
      this.pagination.pageNo = 1;
      this.$nextTick(() => {
        this.$refs.logTable && this.$refs.logTable.clearSort();
      });
      this.loadLogList();
    },
    handleExportCommand(command) {
      const actions = {
        all: this.handleExportAll,
        selected: this.handleExportSelected,
        records: this.showExportRecords,
      };
      actions[command] && actions[command]();
    },
    async createExportTask(logIds = []) {
      if (!this.appId || this.exportLoading) return;

      this.exportLoading = true;
      try {
        const res = await exportConversationLogs({
          appId: this.appId,
          appType: this.appType,
          logIds,
        });
        if (res && (res.code === 0 || res.code === 200)) {
          this.$message.success(this.$t('agent.log.exportSuccess'));
          this.showExportRecords();
        } else {
          this.$message.error(res?.msg || this.$t('common.message.error'));
        }
      } catch (error) {
        this.$message.error(error?.message || this.$t('common.message.error'));
      } finally {
        this.exportLoading = false;
      }
    },
    handleExportAll() {
      this.createExportTask();
    },
    handleExportSelected() {
      const logIds = this.selectedLogs.map(item => item.logId).filter(Boolean);
      if (!logIds.length) {
        this.$message.warning(this.$t('agent.log.selectLogsFirst'));
        return;
      }
      this.createExportTask(logIds);
    },
    showExportRecords() {
      if (!this.appId) return;
      this.$refs.exportRecordDialog.showDialog({
        appId: this.appId,
        appType: this.appType,
      });
    },
    handleSelectionChange(rows) {
      this.selectedLogs = rows;
    },
    clearSelectedLogs() {
      this.selectedLogs = [];
      this.$refs.logTable && this.$refs.logTable.clearSelection();
    },
    handleDetail(row) {
      this.$refs.logDetailDrawer.showDrawer({
        appId: this.appId,
        appType: this.appType,
        conversationId: row.conversationId,
        log: row,
      });
    },
    getSourceLabel(source) {
      const sourceOption = CONVERSATION_LOG_SOURCE_OPTIONS.find(
        item => item.value === source,
      );
      return sourceOption
        ? this.$t(`agent.log.sourceOptions.${sourceOption.labelKey}`)
        : this.displayValue(source);
    },
    displayValue(value) {
      return value === undefined || value === null || value === ''
        ? '-'
        : value;
    },
    formatMilliseconds(value) {
      return value === undefined || value === null || value === ''
        ? '-'
        : `${value} ms`;
    },
    handleSortChange({ prop, order }) {
      const sortableFields = [
        'createAt',
        'updateAt',
        'avgCosts',
        'avgFirstTokenLatency',
      ];
      if (!sortableFields.includes(prop)) return;

      this.queryParams.orderBy = order ? prop : '';
      this.queryParams.orderType =
        order === 'ascending' ? 'asc' : order === 'descending' ? 'des' : '';
      this.pagination.pageNo = 1;
      this.loadLogList();
    },
    handlePageSizeChange(pageSize) {
      this.pagination.pageSize = pageSize;
      this.handleSearch();
    },
    handlePageChange(pageNo) {
      this.pagination.pageNo = pageNo;
      this.loadLogList();
    },
  },
};
</script>

<style scoped lang="scss">
.action-bar {
  display: flex;
  align-items: flex-end;
  gap: 24px;
  padding: 20px 18px;

  &__fields {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    min-width: 0;
    gap: 16px 28px;
  }

  &__field {
    flex: 0 0 200px;
    min-width: 0;

    &--date {
      flex-basis: 230px;
    }

    &--user {
      flex-basis: 208px;
    }
  }

  &__label {
    margin-bottom: 8px;
    color: #4b5563;
    font-size: 13px;
    line-height: 18px;
  }

  &__control {
    width: 100%;
    &.el-input__inner {
      height: 30px;
    }
  }

  &__actions {
    display: flex;
    flex: 0 0 auto;
    gap: 16px;
  }

  &__actions .el-button {
    min-width: 80px;
  }
}

@media screen and (max-width: 1200px) {
  .action-bar {
    align-items: stretch;
    flex-direction: column;

    &__actions {
      justify-content: flex-end;
    }
  }
}
.main-table {
  margin-top: 8px;

  &__title--unread {
    color: $color;
  }

  &__expanded-content {
    padding: 8px 48px;
    color: #606266;
    line-height: 22px;
    white-space: pre-wrap;
  }

  ::v-deep(th.el-table__cell .cell) {
    height: 34px;
    line-height: 34px;
  }
}

.feedback-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #606266;

  &__icon {
    color: #909399;
    font-size: 16px;

    &:nth-of-type(2) {
      margin-left: 6px;
    }
  }
}
.message-pagination {
  margin-top: 16px;
  text-align: right;
}
</style>
