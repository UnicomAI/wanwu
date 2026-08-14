<template>
  <div class="page-wrapper">
    <div class="page-title">
      <i class="el-icon-arrow-left" @click="$router.go(-1)" />
      <svg-icon class="page-title-icon" icon-class="bell" />
      <span class="page-title-name">{{ $t('messageCenter.title') }}</span>
      <div class="message-center-title-actions">
        <el-checkbox
          v-model="queryParams.onlyUnread"
          @change="handleFilterChange"
          class="message-center-title-actions__filter"
        >
          {{ $t('messageCenter.onlyUnreadMessages') }}({{ unreadTotal }})
        </el-checkbox>
        <el-button
          type="primary"
          size="small"
          :disabled="!unreadTotal"
          @click="handleMarkAllAsRead"
        >
          {{ $t('messageCenter.markAllAsRead') }}
        </el-button>
      </div>
    </div>
    <div class="page-wrapper-content">
      <div class="taglist_warp">
        <button
          v-for="tab in messageTabs"
          :key="tab.name"
          type="button"
          class="tagList"
          :class="{ active: tab.name === queryParams.type }"
          @click="selectMessageTab(tab)"
        >
          {{ tab.label }}({{ tab.unread }})
        </button>
      </div>

      <div class="action-bar">
        <el-button
          type="primary"
          size="small"
          :loading="deleteLoading"
          :disabled="!selectedMessages.length"
          @click="handleDel"
        >
          {{ $t('common.button.delete') }}
        </el-button>
        <el-input
          v-model="queryParams.keyword"
          class="action-bar__search"
          prefix-icon="el-icon-search"
          clearable
          :placeholder="$t('messageCenter.searchPlaceholder')"
          @keyup.enter.native="handleSearch"
        />
      </div>
      <el-table
        ref="messageTable"
        v-loading="listLoading"
        :data="tableData"
        class="message-table"
        row-key="id"
        :expand-row-keys="expandedRowKeys"
        @selection-change="handleSelectionChange"
        @expand-change="handleExpandChange"
      >
        <el-table-column
          type="selection"
          :reserve-selection="true"
          width="48"
        />
        <el-table-column type="expand" width="48">
          <template slot-scope="{ row }">
            <product-service-message-content
              v-if="row.category === 2"
              :msg-text="row.msgText"
              :actions="row.actions"
            />
            <div v-else class="message-table__expanded-content">
              {{ row.msgText }}
            </div>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('messageCenter.columns.content')"
          min-width="360"
          show-overflow-tooltip
        >
          <template slot-scope="{ row }">
            <span :class="{ 'message-table__title--unread': !row.isRead }">
              {{ row.title }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="updateAt"
          :label="$t('messageCenter.columns.updateAt')"
          width="180"
        />
        <el-table-column :label="$t('messageCenter.columns.type')" width="150">
          <template slot-scope="{ row }">
            {{ getMessageTypeLabel(row.category) }}
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
    </div>
  </div>
</template>

<script>
import {
  getMessageList,
  getUnreadMessageCount,
  markAllMessagesAsRead,
  markMessageAsRead,
  deleteMessages,
} from '@/api/messageCenter';
import ProductServiceMessageContent from './components/ProductServiceMessageContent.vue';
export default {
  components: { ProductServiceMessageContent },
  data() {
    return {
      tableData: [],
      listLoading: false,
      deleteLoading: false,
      selectedMessages: [],
      expandedRowKeys: [],
      pendingExpandMessageId: '',
      markingMessageIds: [],
      isPageMounted: false,
      pagination: {
        pageNo: 1,
        pageSize: 20,
        total: 0,
      },
      // 未读消息总数
      unreadTotal: 0,
      // 各类消息未读数
      typeUnreads: {
        announcement: 0,
        productService: 0,
      },
      queryParams: {
        onlyUnread: false,
        type: 'productService',
        keyword: '',
      },
    };
  },
  computed: {
    messageTabs() {
      return [
        // {
        //   name: 'all',
        //   label: this.$t('messageCenter.tabs.all'),
        //   unread: this.unreadTotal,
        // },
        // {
        //   name: 'announcement',
        //   label: this.$t('messageCenter.tabs.announcement'),
        //   unread: this.typeUnreads.announcement,
        // },
        {
          name: 'productService',
          label: this.$t('messageCenter.tabs.productService'),
          unread: this.typeUnreads.productService,
        },
      ];
    },
  },
  watch: {
    '$route.params': {
      handler() {
        if (this.syncPopoverNavigationParams() && this.isPageMounted) {
          this.fetchMessages(true);
        }
      },
      deep: true,
    },
  },
  methods: {
    syncPopoverNavigationParams() {
      const { from, messageId, category, onlyUnread } =
        this.$route.params || {};
      if (from !== 'messageCenterPopover' || !messageId) return false;

      const categoryTabMap = {
        1: 'announcement',
        2: 'productService',
      };
      this.queryParams.type = categoryTabMap[Number(category)] || 'all';
      this.queryParams.onlyUnread = Boolean(onlyUnread);
      this.pendingExpandMessageId = messageId;
      this.resetPage();
      return true;
    },
    /* 切换消息分类并重新查询第一页。 */
    selectMessageTab(tab) {
      this.queryParams.type = tab.name;
      this.resetPage();
      this.fetchMessages(true);
    },
    /* 将接口返回的消息类别转换为页面显示文案。 */
    getMessageCategory() {
      const categoryMap = {
        all: 0,
        announcement: 1,
        productService: 2,
      };
      return categoryMap[this.queryParams.type] || 0;
    },
    getMessageTypeLabel(category) {
      const categoryMap = {
        1: 'announcement',
        2: 'productService',
        3: 'ticket',
      };
      return this.$t(`messageCenter.tabs.${categoryMap[category] || 'ticket'}`);
    },
    normalizeMessage(message) {
      return {
        id: message.messageId,
        title: message.title,
        msgText: message.content,
        category: message.category,
        actions: message.actions || [],
        updateAt: this.formatReceivedAt(message.receivedAt),
        isRead: message.isRead,
      };
    },
    formatReceivedAt(receivedAt) {
      const date = new Date(Number(receivedAt));
      if (Number.isNaN(date.getTime())) return '';

      const pad = value => String(value).padStart(2, '0');
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
        date.getDate(),
      )} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(
        date.getSeconds(),
      )}`;
    },
    async fetchUnreadMessageCount() {
      try {
        const res = await getUnreadMessageCount();
        if (res.code !== 0) return;

        const data = res.data || {};
        const byCategory = data.byCategory || {};
        this.unreadTotal = data.total || 0;
        this.typeUnreads = {
          announcement: byCategory.announcement || 0,
          productService: byCategory.productService || 0,
        };
      } catch (error) {
        this.unreadTotal = 0;
        this.typeUnreads = {
          announcement: 0,
          productService: 0,
        };
      }
    },
    /* 预留消息列表请求，并在重新查询时清空已选项。 */
    async fetchMessages(resetSelection = false) {
      if (resetSelection) this.clearSelectedMessages();

      this.listLoading = true;
      try {
        const res = await getMessageList({
          category: this.getMessageCategory(),
          onlyUnread: this.queryParams.onlyUnread,
          keyword: this.queryParams.keyword.trim(),
          pageNo: this.pagination.pageNo,
          pageSize: this.pagination.pageSize,
        });
        if (res.code !== 0) return;

        const data = res.data || {};
        this.tableData = (data.list || []).map(message =>
          this.normalizeMessage(message),
        );
        this.pagination.total = Number(data.total) || 0;
        this.pagination.pageNo = Number(data.pageNo) || this.pagination.pageNo;
        this.pagination.pageSize =
          Number(data.pageSize) || this.pagination.pageSize;
        await this.expandAndMarkTargetMessage();
      } catch (error) {
        this.tableData = [];
        this.pagination.total = 0;
      } finally {
        this.listLoading = false;
      }
    },
    async expandAndMarkTargetMessage() {
      const messageId = this.pendingExpandMessageId;
      if (!messageId) return;

      const message = this.tableData.find(item => item.id === messageId);
      if (!message) return;

      this.pendingExpandMessageId = '';
      this.expandedRowKeys = [messageId];
      await this.$nextTick();

      await this.markMessageAsReadIfNeeded(message);
    },
    handleExpandChange(row, expandedRows) {
      const isExpanded = expandedRows.some(item => item.id === row.id);
      if (isExpanded) this.markMessageAsReadIfNeeded(row);
    },
    async markMessageAsReadIfNeeded(message) {
      if (
        !message ||
        message.isRead ||
        this.markingMessageIds.includes(message.id)
      ) {
        return;
      }

      this.markingMessageIds.push(message.id);
      try {
        const res = await markMessageAsRead({ messageId: message.id });
        if (res.code === 0) message.isRead = true;
      } finally {
        this.markingMessageIds = this.markingMessageIds.filter(
          id => id !== message.id,
        );
        this.fetchUnreadMessageCount();
      }
    },
    handleSelectionChange(selection) {
      this.selectedMessages = selection;
    },
    /* 清空跨页已选消息，并重置表格选择状态。 */
    clearSelectedMessages() {
      this.selectedMessages = [];
      this.$nextTick(() => {
        const table = this.$refs.messageTable;
        if (table) table.clearSelection();
      });
    },
    /* 二次确认后将当前组织的全部未读消息标记为已读。 */
    async handleMarkAllAsRead() {
      try {
        await this.$confirm(
          this.$t('messageCenter.markAllAsReadConfirm'),
          this.$t('messageCenter.title'),
          { type: 'warning' },
        );
      } catch (error) {
        return;
      }

      const res = await markAllMessagesAsRead();
      if (res.code !== 0) return;

      await this.fetchUnreadMessageCount();
      this.fetchMessages(true);
    },
    /* 按关键词重新查询消息列表。 */
    handleSearch() {
      this.resetPage();
      this.fetchMessages(true);
    },
    /* 按未读状态重新查询消息列表。 */
    handleFilterChange() {
      this.resetPage();
      this.fetchMessages(true);
    },
    /* 更新每页条数并重新查询第一页。 */
    handlePageSizeChange(pageSize) {
      this.pagination.pageSize = pageSize;
      this.resetPage();
      this.fetchMessages(true);
    },
    /* 切换页码并查询对应页数据。 */
    handlePageChange(pageNo) {
      this.pagination.pageNo = pageNo;
      this.fetchMessages();
    },
    /* 将分页状态恢复到第一页。 */
    resetPage() {
      this.pagination.pageNo = 1;
    },
    /* 删除勾选的消息，并同步刷新列表和未读数量。 */
    async handleDel() {
      const ids = this.selectedMessages
        .map(message => message.id)
        .filter(Boolean);
      if (!ids.length) return;

      try {
        await this.$confirm(
          this.$t('messageCenter.deleteConfirm'),
          this.$t('common.confirm.title'),
          { type: 'warning' },
        );
      } catch (error) {
        return;
      }

      this.deleteLoading = true;
      try {
        const res = await deleteMessages({ ids });
        if (res.code !== 0) return;

        this.$message.success(this.$t('common.message.success'));
        this.expandedRowKeys = this.expandedRowKeys.filter(
          id => !ids.includes(id),
        );
        await Promise.all([
          this.fetchMessages(true),
          this.fetchUnreadMessageCount(),
        ]);
      } finally {
        this.deleteLoading = false;
      }
    },
  },
  mounted() {
    this.syncPopoverNavigationParams();
    this.isPageMounted = true;
    this.fetchUnreadMessageCount();
    this.fetchMessages(true);
  },
};
</script>

<style lang="scss" scoped>
.page-title {
  display: flex;
  align-items: center;

  .el-icon-arrow-left {
    margin-right: 10px;
    font-size: 15px;
    cursor: pointer;
    color: $color_title;
  }
  .page-title-icon {
    font-size: 18px;
    vertical-align: middle;
    color: $color_title;
    margin-right: 10px;
  }
}

.page-wrapper-content {
  padding: 10px 30px 20px;
}

.taglist_warp {
  display: flex;

  .tagList {
    height: 36px;
    margin: 10px;
    padding: 0 3px;
    border: 0;
    border-bottom: 2.5px solid rgba(255, 255, 255, 0);
    background: transparent;
    line-height: 36px;
    cursor: pointer;
  }

  .tagList:first-child {
    margin-left: 0;
  }

  .active {
    font-weight: bold;
    color: $color;
    border-bottom: 2.5px solid $color !important;
  }
}
.message-center-title-actions {
  display: flex;
  align-items: center;
  margin-left: 24px;

  &__filter {
    margin-right: 16px;
    font-size: 14px;
    font-weight: normal;
  }
}

.action-bar {
  display: flex;
  align-items: center;
  padding: 8px 0;

  &__search {
    width: 280px;
    margin-left: auto;
  }
}
.message-table {
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
}

.message-pagination {
  margin-top: 16px;
  text-align: right;
}

::v-deep(.el-table__expanded-cell) {
  background-color: #f4f8ff;
  border-left: 3px solid #3b82f6;
  &:hover {
    background-color: #f4f8ff !important;
  }
}
</style>
