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
            @change="handleOrganizationChange"
          />
        </div>
        <div class="page-container-right">
          <div class="group-header">
            <h2>
              {{ $t('adminCenter.pageModules.resourcePool.skill.title') }}
            </h2>
          </div>
          <admin-search-bar
            v-model="commonSearchForm"
            :org-id-list="selectedOrganizationIds"
            :preserve-user-selection="isRestoringListState"
            @user-options-loaded="handleUserOptionsLoaded"
            @change="handleCommonSearchFormChange"
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
                  v-model="skillSearchForm.name"
                  class="admin-search-bar__control"
                  clearable
                  prefix-icon="el-icon-search"
                  :placeholder="$t('adminCenter.placeholders.skillName')"
                  @keyup.enter.native="handleSearch(commonSearchForm)"
                />
              </div>
              <!-- <div class="admin-search-bar__field">
                <div class="admin-search-bar__label">
                  {{ $t('adminCenter.columns.publishStatus') }}
                </div>
                <el-select
                  v-model="skillSearchForm.publishStatus"
                  class="admin-search-bar__control"
                  multiple
                  collapse-tags
                  @change="
                    handleSkillMultiSelectChange('publishStatus', $event)
                  "
                >
                  <el-option
                    v-for="item in publishStatusOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </div> -->
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
import adminListQueryMixin from '../mixins/adminListQueryMixin';
import { getAdminSkillPageList } from '@/api/adminCenter';
import {
  ALL_VALUE,
  DRAFT,
  PUBLISHED,
  PUBLISH_STATUS_LIST,
  normalizeAdminListParams,
} from '@/views/adminCenter/constants.js';
import AdminDataTable from '@/views/adminCenter/components/AdminDataTable.vue';
import AdminPageHeader from '@/views/adminCenter/components/AdminPageHeader.vue';
import AdminSearchBar from '@/views/adminCenter/components/AdminSearchBar.vue';
import OrganizationTreeSelect from '@/views/adminCenter/components/OrganizationTreeSelect.vue';

export default {
  mixins: [adminListQueryMixin],
  components: {
    AdminPageHeader,
    AdminDataTable,
    AdminSearchBar,
    OrganizationTreeSelect,
  },
  data() {
    return {
      listStateType: 'skill',
      listStateSearchFormKey: 'skillSearchForm',
      listStatePreviousSelectionsKey: 'previousSkillSelections',
      currentOrganizationId: '1',
      selectedOrganizationIds: [],
      commonSearchForm: {
        userIdList: [ALL_VALUE],
        publishScope: [ALL_VALUE],
      },
      skillSearchForm: {
        name: '',
        publishStatus: [ALL_VALUE],
      },
      previousSkillSelections: {
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
          prop: 'isPublished',
          label: this.$t('adminCenter.columns.publishStatus'),
          type: 'status',
          minWidth: 110,
          formatter: row => this.getPublishStatus(row),
        },
        // {
        //   prop: 'publishStatus',
        //   label: this.$t('adminCenter.columns.publishStatus'),
        //   type: 'status',
        //   minWidth: 110,
        //   statusMap: {
        //     [PUBLISHED]: {
        //       label: this.$t('adminCenter.options.publishStatus.published'),
        //       className: 'admin-data-table__status--published',
        //     },
        //     [DRAFT]: {
        //       label: this.$t('adminCenter.options.publishStatus.draft'),
        //       className: 'admin-data-table__status--draft',
        //     },
        //   },
        // },
        // {
        //   prop: 'publishScope',
        //   label: this.$t('adminCenter.columns.publishScope'),
        //   type: 'ellipsis',
        //   minWidth: 110,
        //   formatter: row => this.getPublishScopeName(row.publishScope),
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
    getPublishStatus(row) {
      if (typeof row.isPublished === 'boolean') {
        return row.isPublished ? PUBLISHED : DRAFT;
      }
      return '';
    },
    getPublishScopeName(scope) {
      const scopeMap = {
        private: this.$t('adminCenter.options.publishScope.private'),
        organization: this.$t('adminCenter.options.publishScope.organization'),
        public: this.$t('adminCenter.options.publishScope.public'),
        global: this.$t('adminCenter.options.publishScope.global'),
      };
      return scopeMap[scope] || scope;
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
        name: this.skillSearchForm.name,
        publishStatus: this.skillSearchForm.publishStatus,
        orgIdList,
        isAllOrg,
      });
      return params;
    },
    handleSkillMultiSelectChange(field, values) {
      const previous = this.previousSkillSelections[field] || [ALL_VALUE];
      values = Array.isArray(values) ? values : [];
      let nextValues = values;

      if (!values.length) {
        nextValues = [ALL_VALUE];
      } else if (values.includes(ALL_VALUE) && values.length > 1) {
        nextValues = previous.includes(ALL_VALUE)
          ? values.filter(item => item !== ALL_VALUE)
          : [ALL_VALUE];
      }

      this.skillSearchForm[field] = nextValues;
      this.previousSkillSelections[field] = [...nextValues];
      this.tablePage = 1;
      this.fetchTableData();
    },
    async requestTableData(commonForm = this.commonSearchForm, requestId) {
      this.tableLoading = true;
      try {
        const res = await getAdminSkillPageList(
          this.createListParams(commonForm),
        );
        if (requestId !== this.latestRequestId) return;
        const data = res.data || {};
        this.tableData = data.list || [];
        this.tableTotal = data.total || 0;
      } finally {
        if (requestId === this.latestRequestId) {
          this.tableLoading = false;
        }
      }
    },
    handleSearch(commonForm) {
      this.tablePage = 1;
      this.fetchTableData(commonForm);
    },
    handleReset(commonForm) {
      this.skipNextCommonSearch = true;
      this.selectedOrganizationIds = [];
      this.skillSearchForm = {
        name: '',
        publishStatus: [ALL_VALUE],
      };
      this.previousSkillSelections = {
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
      this.saveListState();
      this.$router.push({
        path: `/adminCenter/skill/detail?skillId=${row.skillId || row.id}`,
      });
    },
  },
  mounted() {
    if (!this.isRestoringListState) this.fetchTableData();
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
