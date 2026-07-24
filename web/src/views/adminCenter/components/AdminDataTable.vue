<template>
  <div class="admin-data-table">
    <el-table
      v-loading="loading"
      class="admin-data-table__table"
      :data="data"
      :row-key="rowKey"
      :empty-text="emptyText"
      :max-height="resolvedMaxHeight"
      @sort-change="handleSortChange"
    >
      <el-table-column
        v-for="column in columns"
        :key="column.prop || column.type"
        :prop="column.prop"
        :label="column.label"
        :width="column.width"
        :min-width="column.minWidth"
        :fixed="column.fixed"
        :align="column.align || 'left'"
        :sortable="column.sortable || false"
        :show-overflow-tooltip="isOverflowTooltipColumn(column)"
      >
        <template slot-scope="scope">
          <slot
            v-if="$scopedSlots[`cell-${column.prop}`]"
            :name="`cell-${column.prop}`"
            :row="scope.row"
            :value="getCellValue(scope.row, column)"
            :column="column"
          />
          <slot
            v-else-if="column.type === 'actions' && $scopedSlots.actions"
            name="actions"
            :row="scope.row"
            :column="column"
          />
          <div
            v-else-if="column.type === 'modelName'"
            class="admin-data-table__model"
          >
            <span class="admin-data-table__model-icon">
              <i :class="scope.row.icon || 'el-icon-picture-outline'"></i>
            </span>
            <span
              class="admin-data-table__ellipsis admin-data-table__model-name"
            >
              {{ formatCellValue(scope.row, column) }}
            </span>
          </div>
          <div
            v-else-if="column.type === 'avatarName'"
            class="admin-data-table__avatar-name"
          >
            <admin-avatar
              :path="getAvatarPath(scope.row, column)"
              :local-path="getLocalAvatar(scope.row, column)"
              :size="column.avatarSize || 26"
            />
            <span
              class="admin-data-table__ellipsis admin-data-table__avatar-name-text"
            >
              {{ formatCellValue(scope.row, column) }}
            </span>
          </div>
          <span
            v-else-if="column.type === 'status'"
            :class="[
              'admin-data-table__status',
              getStatusMeta(getCellValue(scope.row, column), column).className,
            ]"
          >
            {{ getStatusMeta(getCellValue(scope.row, column), column).label }}
          </span>
          <span
            v-else-if="column.type === 'ellipsis'"
            class="admin-data-table__ellipsis"
          >
            {{ formatCellValue(scope.row, column) }}
          </span>
          <span v-else>{{ formatCellValue(scope.row, column) }}</span>
        </template>
      </el-table-column>
    </el-table>

    <div class="admin-data-table__pagination">
      <el-pagination
        background
        :current-page="page"
        :page-size="pageSize"
        :page-sizes="pageSizes"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script>
import { DRAFT, PUBLISHED } from '../constants.js';
import AdminAvatar from './avatar.vue';

export default {
  name: 'AdminDataTable',
  components: {
    AdminAvatar,
  },
  props: {
    data: {
      type: Array,
      default: () => [],
    },
    columns: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
    total: {
      type: Number,
      default: 0,
    },
    page: {
      type: Number,
      default: 1,
    },
    pageSize: {
      type: Number,
      default: 10,
    },
    pageSizes: {
      type: Array,
      default: () => [5, 10, 20, 50],
    },
    rowKey: {
      type: [String, Function],
      default: 'id',
    },
    emptyText: {
      type: String,
      default: '',
    },
    fitViewport: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      viewportMaxHeight: undefined,
    };
  },
  computed: {
    resolvedMaxHeight() {
      return this.fitViewport ? this.viewportMaxHeight : undefined;
    },
    statusMap() {
      return {
        [PUBLISHED]: {
          label: this.$t('adminCenter.options.publishStatus.published'),
          className: 'admin-data-table__status--published',
        },
        [DRAFT]: {
          label: this.$t('adminCenter.options.publishStatus.draft'),
          className: 'admin-data-table__status--draft',
        },
      };
    },
  },
  watch: {
    data() {
      this.scheduleViewportMaxHeightUpdate();
    },
    total() {
      this.scheduleViewportMaxHeightUpdate();
    },
  },
  mounted() {
    if (this.fitViewport) {
      this.scheduleViewportMaxHeightUpdate();
      window.addEventListener('resize', this.updateViewportMaxHeight);
    }
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.updateViewportMaxHeight);
  },
  methods: {
    scheduleViewportMaxHeightUpdate() {
      if (this.fitViewport) {
        this.$nextTick(this.updateViewportMaxHeight);
      }
    },
    updateViewportMaxHeight() {
      const pagination = this.$el.querySelector(
        '.admin-data-table__pagination',
      );
      const paginationHeight = pagination
        ? pagination.getBoundingClientRect().height
        : 0;
      const { top: tableTop } = this.$el.getBoundingClientRect();
      const styles = window.getComputedStyle(this.$el);
      const verticalBorder =
        parseFloat(styles.borderTopWidth) +
        parseFloat(styles.borderBottomWidth);
      const bottomGap = 40;
      const availableHeight =
        window.innerHeight -
        tableTop -
        paginationHeight -
        verticalBorder -
        bottomGap +
        5;

      this.viewportMaxHeight = Math.max(240, Math.ceil(availableHeight));
    },
    isOverflowTooltipColumn(column) {
      return ['ellipsis', 'modelName', 'avatarName'].includes(column.type);
    },
    getCellValue(row, column) {
      if (typeof column.formatter === 'function') {
        return column.formatter(row, column);
      }
      return this.getValueByPath(row, column.valueProp || column.prop);
    },
    getValueByPath(row, path) {
      if (!path) {
        return undefined;
      }
      return String(path)
        .split('.')
        .reduce((value, key) => (value == null ? undefined : value[key]), row);
    },
    getAvatarPath(row, column) {
      return this.getValueByPath(row, column.avatarPathProp || 'avatar.path');
    },
    getLocalAvatar(row, column) {
      return this.getValueByPath(row, column.localAvatarProp);
    },
    formatCellValue(row, column) {
      const value = this.getCellValue(row, column);
      if (value === undefined || value === null || value === '') {
        return column.emptyText || '-';
      }
      return String(value);
    },
    getStatusMeta(value, column = {}) {
      const statusMap = column.statusMap || this.statusMap;
      return (
        statusMap[value] || {
          label: value || '-',
          className: 'admin-data-table__status--default',
        }
      );
    },
    handleCurrentChange(page) {
      this.$emit('update:page', page);
      this.$emit('page-change', page);
    },
    handleSizeChange(size) {
      this.$emit('update:pageSize', size);
      this.$emit('update:page', 1);
      this.$emit('size-change', size);
      this.$emit('page-change', 1);
    },
    handleSortChange(sort) {
      this.$emit('sort-change', sort);
    },
  },
};
</script>

<style lang="scss">
.admin-data-table {
  width: 100%;
  margin-top: 18px;
  overflow: hidden;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
}

.admin-data-table__table {
  width: 100%;
}

.admin-data-table__table::before {
  display: none;
}

.admin-data-table__table th.el-table__cell {
  background: #f8fafc;
  color: #667085;
  font-weight: 600;
}

.admin-data-table__table .el-table__cell {
  padding: 12px 0;
}

.admin-data-table__model,
.admin-data-table__avatar-name {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.admin-data-table__model-icon {
  display: inline-flex;
  width: 26px;
  height: 26px;
  flex: 0 0 26px;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  background: #f2f4f7;
  color: #667085;
  font-size: 14px;
}

.admin-data-table__model-name,
.admin-data-table__avatar-name-text {
  color: #1f2937;
  font-weight: 600;
}

.admin-data-table__ellipsis {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-data-table__status {
  display: inline-flex;
  height: 22px;
  align-items: center;
  justify-content: center;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 22px;
  white-space: nowrap;
}

.admin-data-table__status--published {
  background: #dcfce7;
  color: #15803d;
}

.admin-data-table__status--draft {
  background: #f2f4f7;
  color: #667085;
}

.admin-data-table__status--default {
  background: #eef2ff;
  color: #375dfb;
}

.admin-data-table__action {
  padding: 0;
  color: #2563eb;
  font-weight: 600;
}

.admin-data-table__pagination {
  display: flex;
  justify-content: flex-end;
  padding: 12px 16px;
  border-top: 1px solid #edf0f5;
  background: #fff;
}
</style>
