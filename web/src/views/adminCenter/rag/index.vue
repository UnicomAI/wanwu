<template>
  <div class="page-wrapper">
    <div class="page-wrapper-content">
      <admin-page-header :title="$t('adminCenter.title')" />
      <div class="page-container">
        <div class="page-container-left">
          <div class="group-header">
            <div class="blue-bar"></div>
            <h2>{{ $t('adminCenter.common.orgStructure') }}</h2>
          </div>
          <organization-tree-select
            v-model="selectedOrganizationIds"
            :current-key.sync="currentOrganizationId"
          />
        </div>
        <div class="page-container-right">
          <div class="group-header">
            <h2>
              {{ $t('adminCenter.pageModules.appDevelopment.rag.title') }}
            </h2>
          </div>
          <admin-search-bar
            v-model="commonSearchForm"
            :org-id-list="selectedOrganizationIds"
            @search="handleSearch"
            @reset="handleReset"
          >
            <template #extra>
              <div
                class="admin-search-bar__field admin-search-bar__field--name"
              >
                <div class="admin-search-bar__label">
                  {{ $t('adminCenter.columns.name') }}
                </div>
                <el-input
                  v-model="ragSearchForm.name"
                  class="admin-search-bar__control"
                  clearable
                  prefix-icon="el-icon-search"
                  :placeholder="$t('adminCenter.placeholders.ragName')"
                  @keyup.enter.native="handleSearch(commonSearchForm)"
                />
              </div>
              <div class="admin-search-bar__field">
                <div class="admin-search-bar__label">
                  {{ $t('adminCenter.columns.publishStatus') }}
                </div>
                <el-select
                  v-model="ragSearchForm.publishStatus"
                  class="admin-search-bar__control"
                  multiple
                  collapse-tags
                  @change="handleRagMultiSelectChange('publishStatus', $event)"
                >
                  <el-option
                    v-for="item in publishStatusOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </div>
            </template>
          </admin-search-bar>
          <admin-data-table
            fit-viewport
            :data="tableData"
            :columns="tableColumns"
            :loading="tableLoading"
            :total="tableTotal"
            :page.sync="tablePage"
            :page-size.sync="tablePageSize"
            @sort-change="handleTableSortChange"
            @page-change="handleTablePageChange"
          >
            <template #actions="{ row }">
              <el-button
                type="text"
                class="admin-data-table__action"
                @click="handleViewDetail(row)"
              >
                {{ $t('adminCenter.actions.viewDetail') }}
              </el-button>
            </template>
          </admin-data-table>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getAdminRagPageList } from '@/api/adminCenter';
import {
  ALL_VALUE,
  PUBLISH_STATUS_LIST,
  normalizeAdminListParams,
} from '@/views/adminCenter/constants.js';
import AdminDataTable from '@/views/adminCenter/components/AdminDataTable.vue';
import AdminPageHeader from '@/views/adminCenter/components/AdminPageHeader.vue';
import AdminSearchBar from '@/views/adminCenter/components/AdminSearchBar.vue';
import OrganizationTreeSelect from '@/views/adminCenter/components/OrganizationTreeSelect.vue';

export default {
  components: {
    AdminPageHeader,
    AdminDataTable,
    AdminSearchBar,
    OrganizationTreeSelect,
  },
  data() {
    return {
      currentOrganizationId: '1',
      selectedOrganizationIds: [],
      commonSearchForm: {
        userIdList: [ALL_VALUE],
        publishScope: [ALL_VALUE],
      },
      ragSearchForm: {
        name: '',
        publishStatus: [ALL_VALUE],
      },
      previousRagSelections: {
        publishStatus: [ALL_VALUE],
      },
      tableLoading: false,
      tablePage: 1,
      tablePageSize: 10,
      tableTotal: 0,
      tableSort: {
        prop: '',
        order: '',
      },
      tableData: [],
    };
  },
  computed: {
    publishStatusOptions() {
      return [
        {
          label: this.$t('adminCenter.options.common.all'),
          value: ALL_VALUE,
        },
        ...PUBLISH_STATUS_LIST,
      ];
    },
    tableColumns() {
      return [
        {
          prop: 'name',
          label: this.$t('adminCenter.columns.name'),
          type: 'avatarName',
          avatarPathProp: 'avatar.path',
          minWidth: 150,
        },
        {
          prop: 'desc',
          label: this.$t('adminCenter.columns.description'),
          type: 'ellipsis',
          minWidth: 150,
        },
        {
          prop: 'ownerUserName',
          label: this.$t('adminCenter.columns.creator'),
          type: 'ellipsis',
          minWidth: 90,
          formatter: row => row.ownerUserName || row.userName || row.creator,
        },
        {
          prop: 'ownerOrgName',
          label: this.$t('adminCenter.columns.organization'),
          type: 'ellipsis',
          minWidth: 110,
          formatter: row => row.ownerOrgName || row.orgName,
        },
        {
          prop: 'publishStatus',
          label: this.$t('adminCenter.columns.publishStatus'),
          type: 'status',
          minWidth: 110,
          formatter: row => {
            return !row.publishType
              ? this.$t('adminCenter.options.publishStatus.draft')
              : this.$t('adminCenter.options.publishStatus.published');
          },
        },
        // {
        //   prop: 'publishType',
        //   label: this.$t('adminCenter.columns.publishScope'),
        //   type: 'ellipsis',
        //   minWidth: 110,
        //   formatter: row => this.getPublishScopeName(row.publishType),
        // },
        {
          prop: 'updatedAt',
          label: this.$t('adminCenter.columns.updateTime'),
          // sortable: 'custom',
          minWidth: 160,
          formatter: row => row.updatedAt || row.updateTime,
        },
        {
          prop: 'actions',
          label: this.$t('adminCenter.columns.actions'),
          type: 'actions',
          width: 100,
          fixed: 'right',
        },
      ];
    },
  },
  methods: {
    getPublishScopeName(scope) {
      const scopeMap = {
        private: this.$t('adminCenter.options.publishScope.private'),
        organization: this.$t('adminCenter.options.publishScope.organization'),
        public: this.$t('adminCenter.options.publishScope.public'),
        global: this.$t('adminCenter.options.publishScope.global'),
      };
      return scopeMap[scope] || scope || '-';
    },
    createListParams(commonForm = this.commonSearchForm) {
      const orgIdList = this.selectedOrganizationIds.filter(
        orgId => orgId !== -1,
      );
      const isAllOrg =
        this.selectedOrganizationIds.includes(-1) || orgIdList.length === 0;
      const params = normalizeAdminListParams({
        pageNo: this.tablePage,
        pageSize: this.tablePageSize,
        ...commonForm,
        name: this.ragSearchForm.name,
        publishStatus: this.ragSearchForm.publishStatus,
        orgIdList,
        isAllOrg,
      });
      return params;
    },
    handleRagMultiSelectChange(field, values) {
      const previous = this.previousRagSelections[field] || [ALL_VALUE];
      values = Array.isArray(values) ? values : [];
      let nextValues = values;

      if (!values.length) {
        nextValues = [ALL_VALUE];
      } else if (values.includes(ALL_VALUE) && values.length > 1) {
        nextValues = previous.includes(ALL_VALUE)
          ? values.filter(item => item !== ALL_VALUE)
          : [ALL_VALUE];
      }

      this.ragSearchForm[field] = nextValues;
      this.previousRagSelections[field] = [...nextValues];
    },
    async fetchTableData(commonForm = this.commonSearchForm) {
      this.tableLoading = true;
      try {
        const res = await getAdminRagPageList(
          this.createListParams(commonForm),
        );
        const data = res.data || {};
        this.tableData = data.list || [];
        this.tableTotal = data.total || 0;
      } finally {
        this.tableLoading = false;
      }
    },
    handleSearch(commonForm) {
      this.tablePage = 1;
      this.fetchTableData(commonForm);
    },
    handleReset(commonForm) {
      this.ragSearchForm = {
        name: '',
        publishStatus: [ALL_VALUE],
      };
      this.previousRagSelections = {
        publishStatus: [ALL_VALUE],
      };
      this.tablePage = 1;
      this.fetchTableData(commonForm);
    },
    handleTablePageChange() {
      this.fetchTableData();
    },
    handleTableSortChange(sort) {
      this.tableSort = {
        prop: sort.prop,
        order: sort.order,
      };
      this.tablePage = 1;
      this.fetchTableData();
    },
    handleViewDetail(row) {
      this.$router.push({
        path: `/adminCenter/rag/detail?appId=${row.appId || row.id}`,
      });
    },
  },
  mounted() {
    this.fetchTableData();
  },
};
</script>

<style lang="scss" scoped>
@import '@/views/adminCenter/styles/common.scss';
@import '@/views/adminCenter/styles/list.scss';

.page-container-left {
  display: flex;
  flex-direction: column;

  .group-header {
    margin-bottom: 16px;
  }
}

.page-container-right {
  min-width: 0;
}
</style>
