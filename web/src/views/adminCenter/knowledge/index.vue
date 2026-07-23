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
              {{ $t('adminCenter.pageModules.resourcePool.knowledge.title') }}
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
                  v-model="knowledgeSearchForm.name"
                  class="admin-search-bar__control"
                  clearable
                  prefix-icon="el-icon-search"
                  :placeholder="$t('adminCenter.placeholders.knowledgeName')"
                  @keyup.enter.native="handleSearch(commonSearchForm)"
                />
              </div>
              <div class="admin-search-bar__field basis-240">
                <div class="admin-search-bar__label">
                  {{ $t('adminCenter.columns.knowledgeType') }}
                </div>
                <el-cascader
                  v-model="knowledgeSearchForm.category"
                  class="admin-search-bar__control"
                  :options="knowledgeTypeOptions"
                  :props="knowledgeTypeProps"
                  collapse-tags
                  clearable
                  @change="handleKnowledgeTypeChange"
                />
              </div>
              <div class="admin-search-bar__field">
                <div class="admin-search-bar__label">
                  {{ $t('adminCenter.columns.knowledgeSource') }}
                </div>
                <el-select
                  v-model="knowledgeSearchForm.external"
                  class="admin-search-bar__control"
                  @change="handleKnowledgeSourceChange"
                >
                  <el-option
                    v-for="item in knowledgeSourceOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </div>
              <!-- <div class="admin-search-bar__field">
                <div class="admin-search-bar__label">
                  {{ $t('adminCenter.columns.publishStatus') }}
                </div>
                <el-select
                  v-model="knowledgeSearchForm.publishStatus"
                  class="admin-search-bar__control"
                  multiple
                  collapse-tags
                  @change="
                    handleKnowledgeMultiSelectChange('publishStatus', $event)
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
              <span v-if="isExternalKnowledge(row)">--</span>
              <el-button
                v-else
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
import { getAdminKnowledgePageList } from '@/api/adminCenter';
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
import {
  INTERNAL,
  EXTERNAL,
  KNOWLEDGE,
  QA,
  MULTIMODAL,
} from '@/views/knowledge/constants';

const KNOWLEDGE_PARENT_VALUE = 'knowledge';
const KNOWLEDGE_CATEGORY_VALUES = [KNOWLEDGE, MULTIMODAL, QA];
const ALL_EXTERNAL_VALUE = -1;

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
      listStateType: 'knowledge',
      listStateSearchFormKey: 'knowledgeSearchForm',
      listStatePreviousSelectionsKey: 'previousKnowledgeSelections',
      currentOrganizationId: '1',
      selectedOrganizationIds: [],
      commonSearchForm: {
        userIdList: [ALL_VALUE],
        publishScope: [ALL_VALUE],
      },
      knowledgeSearchForm: {
        name: '',
        category: [ALL_VALUE],
        external: ALL_EXTERNAL_VALUE,
        publishStatus: [ALL_VALUE],
      },
      previousKnowledgeSelections: {
        category: [ALL_VALUE],
        publishStatus: [ALL_VALUE],
      },
      knowledgeTypeProps: {
        multiple: true,
        emitPath: false,
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
        ...PUBLISH_STATUS_LIST.filter(item => item.value !== DRAFT),
      ];
    },
    knowledgeSourceOptions() {
      return [
        {
          label: this.$t('adminCenter.options.common.all'),
          value: ALL_EXTERNAL_VALUE,
        },
        {
          label: this.$t('knowledgeManage.internal'),
          value: INTERNAL,
        },
        {
          label: this.$t('knowledgeManage.external'),
          value: EXTERNAL,
        },
      ];
    },
    knowledgeTypeOptions() {
      return [
        {
          label: this.$t('adminCenter.options.common.all'),
          value: ALL_VALUE,
        },
        {
          label: this.$t(
            'adminCenter.pageModules.resourcePool.knowledge.title',
          ),
          value: KNOWLEDGE_PARENT_VALUE,
          children: [
            {
              label: this.$t('knowledgeManage.textKnowledgeDatabase.title'),
              value: KNOWLEDGE,
            },
            {
              label: this.$t('knowledgeManage.multiKnowledgeDatabase.title'),
              value: MULTIMODAL,
            },
          ],
        },
        {
          label: this.$t('knowledgeManage.qaDatabase.title'),
          value: QA,
        },
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
          prop: 'description',
          label: this.$t('adminCenter.columns.description'),
          type: 'ellipsis',
          minWidth: 150,
        },
        {
          prop: 'category',
          label: this.$t('adminCenter.columns.knowledgeType'),
          type: 'ellipsis',
          minWidth: 120,
          formatter: row => this.getKnowledgeTypeName(row.category),
        },
        {
          prop: 'external',
          label: this.$t('adminCenter.columns.knowledgeSource'),
          type: 'ellipsis',
          minWidth: 120,
          formatter: row => this.getKnowledgeSourceName(row.external),
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
    getKnowledgeTypeName(type) {
      const typeMap = {
        [KNOWLEDGE]: this.$t('knowledgeManage.textKnowledgeDatabase.title'),
        [QA]: this.$t('knowledgeManage.qaDatabase.title'),
        [MULTIMODAL]: this.$t('knowledgeManage.multiKnowledgeDatabase.title'),
      };
      return typeMap[type] || type;
    },
    getKnowledgeSourceName(source) {
      const sourceMap = {
        [INTERNAL]: this.$t('knowledgeManage.internal'),
        [EXTERNAL]: this.$t('knowledgeManage.external'),
      };
      return sourceMap[source] || source;
    },
    isExternalKnowledge(row) {
      return Number(row && row.external) === EXTERNAL;
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
        name: this.knowledgeSearchForm.name,
        publishStatus: this.knowledgeSearchForm.publishStatus,
        orgIdList,
        isAllOrg,
      });
      const category = this.normalizeCategoryValues(
        this.knowledgeSearchForm.category,
      );
      if (category.length) {
        params.category = category;
      }

      params.external = this.knowledgeSearchForm.external;

      return params;
    },
    normalizeCategoryValues(values) {
      if (!Array.isArray(values) || values.includes(ALL_VALUE)) {
        return [];
      }
      const categorySet = new Set();
      values.forEach(value => {
        if (value === KNOWLEDGE_PARENT_VALUE) {
          categorySet.add(KNOWLEDGE);
          categorySet.add(MULTIMODAL);
          return;
        }
        if (KNOWLEDGE_CATEGORY_VALUES.includes(value)) {
          categorySet.add(value);
        }
      });
      return [...categorySet];
    },
    handleKnowledgeTypeChange(values) {
      const previous = this.previousKnowledgeSelections.category || [ALL_VALUE];
      values = Array.isArray(values) ? values : [];
      const selectedCategories = values.filter(value =>
        KNOWLEDGE_CATEGORY_VALUES.includes(value),
      );
      const hasAll = values.includes(ALL_VALUE);
      const hadAll = previous.includes(ALL_VALUE);
      const allCategoriesSelected = KNOWLEDGE_CATEGORY_VALUES.every(value =>
        selectedCategories.includes(value),
      );
      let nextValues = selectedCategories;

      if (!values.length) {
        nextValues = [ALL_VALUE];
      } else if (hasAll && !hadAll) {
        nextValues = [ALL_VALUE, ...KNOWLEDGE_CATEGORY_VALUES];
      } else if (allCategoriesSelected) {
        nextValues = [ALL_VALUE, ...KNOWLEDGE_CATEGORY_VALUES];
      } else if (hasAll) {
        nextValues = selectedCategories.length
          ? selectedCategories
          : [ALL_VALUE];
      }

      this.knowledgeSearchForm.category = nextValues;
      this.previousKnowledgeSelections.category = [...nextValues];
      this.tablePage = 1;
      this.fetchTableData();
    },
    handleKnowledgeSourceChange() {
      this.tablePage = 1;
      this.fetchTableData();
    },
    handleKnowledgeMultiSelectChange(field, values) {
      const previous = this.previousKnowledgeSelections[field] || [ALL_VALUE];
      values = Array.isArray(values) ? values : [];
      let nextValues = values;

      if (!values.length) {
        nextValues = [ALL_VALUE];
      } else if (values.includes(ALL_VALUE) && values.length > 1) {
        nextValues = previous.includes(ALL_VALUE)
          ? values.filter(item => item !== ALL_VALUE)
          : [ALL_VALUE];
      }

      this.knowledgeSearchForm[field] = nextValues;
      this.previousKnowledgeSelections[field] = [...nextValues];
      this.tablePage = 1;
      this.fetchTableData();
    },
    async requestTableData(commonForm = this.commonSearchForm, requestId) {
      this.tableLoading = true;
      try {
        const res = await getAdminKnowledgePageList(
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
      this.knowledgeSearchForm = {
        name: '',
        category: [ALL_VALUE],
        external: ALL_EXTERNAL_VALUE,
        publishStatus: [ALL_VALUE],
      };
      this.previousKnowledgeSelections = {
        category: [ALL_VALUE],
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
        path: `/adminCenter/knowledge/detail?knowledgeId=${row.knowledgeId}`,
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

::v-deep(.el-cascader) {
  line-height: 30px;
  .el-input--suffix {
    line-height: 30px;
  }
  .el-input__inner {
    height: 30px !important;
  }
}
</style>
