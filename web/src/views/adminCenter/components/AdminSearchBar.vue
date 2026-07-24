<template>
  <div class="admin-search-bar">
    <div class="admin-search-bar__fields">
      <div class="admin-search-bar__field admin-search-bar__field--user">
        <div class="admin-search-bar__label">
          {{ $t('adminCenter.columns.user') }}
        </div>
        <el-select
          v-model="localForm.userIdList"
          class="admin-search-bar__control"
          multiple
          filterable
          collapse-tags
          :loading="userLoading"
          @change="handleMultiSelectChange('userIdList', $event)"
        >
          <el-option
            v-for="item in userSelectOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>

      <!-- <div class="admin-search-bar__field">
        <div class="admin-search-bar__label">
          {{ $t('adminCenter.columns.publishScope') }}
        </div>
        <el-select
          v-model="localForm.publishScope"
          class="admin-search-bar__control"
          multiple
          collapse-tags
          @change="handleMultiSelectChange('publishScope', $event)"
        >
          <el-option
            v-for="item in publishScopeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div> -->

      <slot name="extra" :form="localForm"></slot>

      <div class="admin-search-bar__actions">
        <el-button type="primary" size="mini" @click="handleSearch">
          {{ $t('adminCenter.actions.search') }}
        </el-button>
        <el-button size="mini" @click="handleReset">
          {{ $t('adminCenter.actions.reset') }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { debounce } from 'lodash-es';
import { queryOrgUsers } from '@/api/permission/org';
import { ALL_VALUE, SCOPE_TYPE_LIST } from '../constants.js';

const ORG_USER_OPTIONS_DEBOUNCE = 1000;

export default {
  name: 'AdminSearchBar',
  props: {
    value: {
      type: Object,
      default: () => ({}),
    },
    orgIdList: {
      type: Array,
      default: () => [],
    },
    preserveUserSelection: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    const defaultForm = this.createDefaultForm();
    const localForm = this.normalizeForm(this.value);
    return {
      userLoading: false,
      userOptions: [],
      latestUserOptionsRequestId: 0,
      debouncedLoadUserOptions: null,
      localForm,
      previousSelections: {
        userIdList: [...localForm.userIdList],
        publishScope: [
          ...(this.value.publishScope || defaultForm.publishScope),
        ],
      },
    };
  },
  computed: {
    currentOrgIdList() {
      return this.orgIdList;
    },
    userSelectOptions() {
      return [this.allOption, ...this.userOptions];
    },
    allOption() {
      return {
        label: this.$t('adminCenter.options.common.all'),
        value: ALL_VALUE,
      };
    },
    publishScopeOptions() {
      return [this.allOption, ...SCOPE_TYPE_LIST];
    },
  },
  watch: {
    value: {
      deep: true,
      handler(value) {
        this.syncFormFromValue(value);
      },
    },
    currentOrgIdList: {
      deep: true,
      immediate: true,
      handler() {
        this.scheduleLoadUserOptions();
      },
    },
  },
  methods: {
    createDefaultForm() {
      return {
        userIdList: [ALL_VALUE],
        publishScope: [ALL_VALUE],
      };
    },
    normalizeForm(value = {}) {
      const defaultForm = this.createDefaultForm();
      return {
        userIdList: value.userIdList || value.userIds || defaultForm.userIdList,
        publishScope: value.publishScope || defaultForm.publishScope,
      };
    },
    syncFormFromValue(value) {
      const nextForm = this.normalizeForm(value);
      if (JSON.stringify(nextForm) === JSON.stringify(this.localForm)) {
        return;
      }
      this.localForm = nextForm;
      this.previousSelections = {
        userIdList: [...nextForm.userIdList],
        publishScope: [...nextForm.publishScope],
      };
    },
    scheduleLoadUserOptions() {
      const requestId = ++this.latestUserOptionsRequestId;
      const orgIdList = [...this.currentOrgIdList];
      if (!this.debouncedLoadUserOptions) {
        this.debouncedLoadUserOptions = debounce(
          (id, ids) => this.loadUserOptions(id, ids),
          ORG_USER_OPTIONS_DEBOUNCE,
          { leading: true, trailing: true },
        );
      }
      this.debouncedLoadUserOptions(requestId, orgIdList);
    },
    async loadUserOptions(requestId, orgIdList) {
      this.userLoading = true;
      try {
        const userOptions = await this.fetchUserOptions(orgIdList);
        if (requestId !== this.latestUserOptionsRequestId) return;
        this.userOptions = userOptions;
        const previousUserIdList = [...this.localForm.userIdList];
        const availableUserIds = new Set(
          userOptions.map(user => String(user.value)),
        );
        const nextUserIdList =
          this.preserveUserSelection && !previousUserIdList.includes(ALL_VALUE)
            ? previousUserIdList.filter(userId =>
                availableUserIds.has(String(userId)),
              )
            : [ALL_VALUE];
        this.setFieldSelection(
          'userIdList',
          nextUserIdList.length ? nextUserIdList : [ALL_VALUE],
        );
        if (
          JSON.stringify(previousUserIdList) !==
          JSON.stringify(this.localForm.userIdList)
        ) {
          this.emitChange(false);
        }
      } catch (error) {
        if (requestId !== this.latestUserOptionsRequestId) return;
        this.userOptions = [];
      } finally {
        if (requestId === this.latestUserOptionsRequestId) {
          this.userLoading = false;
          this.$emit('user-options-loaded');
        }
      }
    },
    async fetchUserOptions(orgIdList) {
      const isAllOrg = orgIdList.includes(-1);
      const res = await queryOrgUsers({
        orgIdList: orgIdList.filter(orgId => orgId !== -1),
        isAllOrg,
      });
      const users = (res.data && res.data.users) || [];
      return users.map(user => ({
        label: user.name,
        value: user.id,
      }));
    },
    handleMultiSelectChange(field, values) {
      const previous = this.previousSelections[field] || [ALL_VALUE];
      let nextValues = values;

      if (!values.length) {
        nextValues = [ALL_VALUE];
      } else if (values.includes(ALL_VALUE) && values.length > 1) {
        nextValues = previous.includes(ALL_VALUE)
          ? values.filter(item => item !== ALL_VALUE)
          : [ALL_VALUE];
      }

      this.setFieldSelection(field, nextValues);
      this.emitChange();
    },
    setFieldSelection(field, values) {
      this.localForm[field] = values;
      this.previousSelections[field] = [...values];
    },
    emitChange(emitSearch = true) {
      const form = { ...this.localForm };
      this.$emit('input', form);
      if (emitSearch) this.$emit('change', form);
    },
    handleSearch() {
      this.$emit('search', { ...this.localForm });
    },
    handleReset() {
      const defaultForm = this.createDefaultForm();
      this.localForm = defaultForm;
      this.previousSelections = {
        userIdList: [...defaultForm.userIdList],
        publishScope: [...defaultForm.publishScope],
      };
      this.$emit('reset', { ...this.localForm });
      this.emitChange();
    },
  },
  beforeDestroy() {
    this.debouncedLoadUserOptions?.cancel();
  },
};
</script>

<style lang="scss">
.admin-search-bar {
  width: 100%;
  border: 1px solid #e5e7eb;
  border-radius: 15px;
  padding: 30px 18px;
}

.admin-search-bar__fields {
  display: flex;
  width: 100%;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 16px 28px;
  min-width: 0;
}

.admin-search-bar__field {
  flex: 0 0 160px;
}

.admin-search-bar__field--user {
  flex-basis: 180px;
}

.admin-search-bar__field--name {
  flex-basis: 200px;
}

.admin-search-bar__label {
  margin-bottom: 8px;
  color: #4b5563;
  font-size: 13px;
  line-height: 18px;
}

.admin-search-bar__control {
  width: 100%;
}

.admin-search-bar__actions {
  display: flex;
  flex: 0 0 auto;
  align-self: flex-end;
  gap: 8px;
  margin-left: auto;
}

.admin-search-bar__actions .el-button {
  min-width: 72px;
}
</style>
