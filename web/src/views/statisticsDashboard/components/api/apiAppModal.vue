<template>
  <el-dialog
    :title="$t('statisticsDashboard.apiAppUsageStats')"
    :visible.sync="visible"
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
          prop="name"
          :label="$t('statisticsDashboard.apiName')"
          align="left"
          min-width="140"
        >
          <template slot-scope="scope">
            {{ apiInfo.name || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="apiKey"
          :label="'API Key'"
          align="left"
          min-width="160"
        >
          <template slot-scope="scope">
            <span>
              {{
                apiInfo.apiKey ? apiInfo.apiKey.slice(0, 6) + '******' : '--'
              }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="orgName"
          :label="$t('statisticsDashboard.org')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ apiInfo.orgName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="userName"
          :label="$t('statisticsDashboard.userName')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ apiInfo.userName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="methodPath"
          :label="$t('statisticsDashboard.apiPath')"
          align="left"
          min-width="200"
        >
          <template slot-scope="scope">
            {{ apiInfo.methodPath || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="source"
          :label="$t('statisticsDashboard.source')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">{{ scope.row.source || '--' }}</template>
        </el-table-column>
        <el-table-column
          prop="module"
          :label="$t('statisticsDashboard.module')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">{{ scope.row.module || '--' }}</template>
        </el-table-column>
        <el-table-column
          prop="appName"
          :label="$t('statisticsDashboard.appName')"
          align="left"
          min-width="140"
        >
          <template slot-scope="scope">
            {{ scope.row.appName || '--' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="appType"
          :label="$t('statisticsDashboard.appType')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            <span :class="['type-tag', getAppTypeTagClass(scope.row)]">
              {{ appTypeObj[scope.row.appType] || scope.row.appType || '--' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="author"
          :label="$t('statisticsDashboard.author')"
          align="left"
          min-width="120"
        >
          <template slot-scope="scope">
            {{ scope.row.author || scope.row.userName || '--' }}
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
          :label="$t('statisticsDashboard.appCallCount')"
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
          :label="$t('statisticsDashboard.appCallFailure')"
          align="left"
          sortable="custom"
          min-width="120"
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
          min-width="150"
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
          min-width="150"
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
          min-width="120"
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
import { fetchApiAppList, exportApiAppData } from '@/api/statisticsDashboard';
import { AppType, AGENT, CHAT, RAG, WORKFLOW } from '@/utils/commonSet';

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
    apiInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      listApi: fetchApiAppList,
      loading: false,
      tableData: [],
      appTypeObj: AppType,
    };
  },
  computed: {
    modalParams() {
      const { apiKeyIds, methodPaths, ...rest } = this.params || {};
      return {
        ...rest,
        apiKeyId: this.apiInfo?.apiKeyId || this.apiInfo?.keyId || '',
        apiKey: this.apiInfo?.apiKey || '',
        methodPath: this.apiInfo?.methodPath || '',
      };
    },
  },
  methods: {
    formatAmount,
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
      const response = await exportApiAppData(this.modalParams);
      resDownloadFile(
        response,
        `${this.$t('statisticsDashboard.apiAppUsageStats')}.xlsx`,
      );
    },
    getAppTypeTagClass(row) {
      const typeTag = {
        [AGENT]: 'tag-purple',
        [WORKFLOW]: 'tag-green',
        [RAG]: 'tag-blue',
        [CHAT]: 'tag-orange',
      };
      return typeTag[row.appType] || 'tag-gray';
    },
  },
};
</script>

<style lang="scss" scoped>
.modal-toolbar {
  margin-bottom: 16px;
}

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.5;
}

.tag-blue {
  color: $tag_color;
  background: $tag_bg;
}

.tag-green {
  color: #67c23a;
  background: rgba(103, 194, 58, 0.1);
}

.tag-orange {
  color: #e6a23c;
  background: rgba(230, 162, 60, 0.1);
}

.tag-purple {
  color: #a55fef;
  background: rgba(165, 95, 239, 0.1);
}

.tag-gray {
  color: #909399;
  background: rgba(144, 147, 153, 0.1);
}
</style>
