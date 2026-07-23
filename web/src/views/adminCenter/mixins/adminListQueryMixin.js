import { cloneDeep, debounce } from 'lodash-es';
import { ALL_VALUE } from '../constants.js';

const TABLE_REQUEST_DEBOUNCE = 1000;

export default {
  data() {
    return {
      latestRequestId: 0,
      debouncedRequestTableData: null,
      skipNextCommonSearch: false,
      isRestoringListState: false,
    };
  },
  methods: {
    saveListState() {
      if (!this.listStateType || !this.listStateSearchFormKey) return;

      this.$store.commit('adminCenter/SET_ACTIVE_LIST', {
        type: this.listStateType,
        filters: {
          currentOrganizationId: this.currentOrganizationId,
          selectedOrganizationIds: cloneDeep(this.selectedOrganizationIds),
          commonSearchForm: cloneDeep(this.commonSearchForm),
          searchForm: cloneDeep(this[this.listStateSearchFormKey]),
          tablePage: this.tablePage,
          tablePageSize: this.tablePageSize,
          tableSort: cloneDeep(this.tableSort),
        },
      });
    },
    restoreListState() {
      if (!this.listStateType || !this.listStateSearchFormKey) return;

      const savedList = this.$store.getters['adminCenter/activeList'];
      if (!savedList || savedList.type !== this.listStateType) return;

      const { filters } = savedList;
      this.currentOrganizationId = filters.currentOrganizationId;
      this.selectedOrganizationIds = cloneDeep(filters.selectedOrganizationIds);
      this.commonSearchForm = cloneDeep(filters.commonSearchForm);
      this[this.listStateSearchFormKey] = cloneDeep(filters.searchForm);
      this.tablePage = filters.tablePage;
      this.tablePageSize = filters.tablePageSize;
      this.tableSort = cloneDeep(filters.tableSort);

      if (this.listStatePreviousSelectionsKey) {
        this[this.listStatePreviousSelectionsKey] = Object.entries(
          this[this.listStateSearchFormKey],
        ).reduce((selections, [key, value]) => {
          if (Array.isArray(value)) selections[key] = [...value];
          return selections;
        }, {});
      }
      this.isRestoringListState = true;
    },
    handleUserOptionsLoaded() {
      if (!this.isRestoringListState) return;
      this.isRestoringListState = false;
      this.fetchTableData();
    },
    handleOrganizationChange() {
      if (this.isRestoringListState) return;
      this.commonSearchForm = {
        ...this.commonSearchForm,
        userIdList: [ALL_VALUE],
      };
      this.tablePage = 1;
      this.fetchTableData();
    },
    handleCommonSearchFormChange(commonForm) {
      if (this.isRestoringListState) return;
      this.commonSearchForm = { ...commonForm };
      if (this.skipNextCommonSearch) {
        this.skipNextCommonSearch = false;
        return;
      }
      this.tablePage = 1;
      this.fetchTableData();
    },
    fetchTableData(commonForm = this.commonSearchForm) {
      this.latestRequestId += 1;
      return this.debouncedRequestTableData({ ...commonForm });
    },
  },
  created() {
    this.debouncedRequestTableData = debounce(
      commonForm => this.requestTableData(commonForm, ++this.latestRequestId),
      TABLE_REQUEST_DEBOUNCE,
      { leading: true, trailing: true },
    );
    this.restoreListState();
  },
  beforeDestroy() {
    this.debouncedRequestTableData?.cancel();
  },
};
