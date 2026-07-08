<template>
  <div class="global-filter-wrapper">
    <span class="global-filter-label">
      {{ $t('statisticsDashboard.global') }}:
    </span>
    <el-select
      v-model="filterParams.orgIds"
      :placeholder="$t('statisticsDashboard.selectOrg')"
      :class="[
        'no-border-select',
        'scroll-select',
        { 'hide-tag-close': isOrgSelectedAll },
      ]"
      style="margin-left: 15px; width: 240px"
      multiple
      filterable
      clearable
      collapse-tags
      @change="handleOrgChange"
    >
      <el-option :label="$t('statisticsDashboard.all')" :value="ALL" />
      <el-option
        v-for="item in orgList"
        :key="item.id"
        :label="item.name"
        :value="item.id"
      />
    </el-select>
    <el-select
      v-model="filterParams.userIds"
      :placeholder="$t('statisticsDashboard.selectUser')"
      :class="[
        'no-border-select',
        'scroll-select',
        { 'hide-tag-close': isUserSelectedAll },
      ]"
      style="margin-left: 15px; width: 240px"
      multiple
      collapse-tags
      filterable
      clearable
      @change="handleUserChange"
    >
      <el-option :label="$t('statisticsDashboard.all')" :value="ALL" />
      <el-option
        v-for="item in userList"
        :key="item.userId"
        :label="item.username"
        :value="item.userId"
      />
    </el-select>
  </div>
</template>

<script>
import { fetchOrgs, fetchUsers } from '@/api/statisticsDashboard';
import { ALL } from '../constants';

export default {
  name: 'GlobalFilter',
  data() {
    return {
      ALL,
      orgList: [],
      userList: [],
      filterParams: {
        orgIds: [ALL],
        userIds: [ALL],
      },
    };
  },
  computed: {
    isOrgSelectedAll() {
      return this.filterParams.orgIds.includes(ALL);
    },
    isUserSelectedAll() {
      return this.filterParams.userIds.includes(ALL);
    },
  },
  mounted() {
    this.fetchOrgList();
    this.fetchUserList();
  },
  methods: {
    async fetchOrgList() {
      const res = await fetchOrgs();
      this.orgList = res.data ? res.data.list || [] : [];
    },
    async fetchUserList() {
      const res = await fetchUsers();
      this.userList = res.data ? res.data.list || [] : [];
    },
    handleOrgChange(vals) {
      if (!vals || !vals.length) {
        this.filterParams.orgIds = [ALL];
      } else {
        const lastVal = vals[vals.length - 1];
        if (lastVal === ALL) {
          this.filterParams.orgIds = [ALL];
        } else {
          const allIndex = this.filterParams.orgIds.indexOf(ALL);
          if (allIndex !== -1) {
            this.filterParams.orgIds.splice(allIndex, 1);
          }
        }
      }
      this.emitChange();
    },
    handleUserChange(vals) {
      if (!vals || !vals.length) {
        this.filterParams.userIds = [ALL];
      } else {
        const lastVal = vals[vals.length - 1];
        if (lastVal === ALL) {
          this.filterParams.userIds = [ALL];
        } else {
          const allIndex = this.filterParams.userIds.indexOf(ALL);
          if (allIndex !== -1) {
            this.filterParams.userIds.splice(allIndex, 1);
          }
        }
      }
      this.emitChange();
    },
    emitChange() {
      this.$emit('change', {
        orgIds: this.filterParams.orgIds,
        userIds: this.filterParams.userIds,
      });
    },
    reset() {
      this.filterParams.orgIds = [ALL];
      this.filterParams.userIds = [ALL];
      this.emitChange();
    },
  },
};
</script>

<style lang="scss" scoped>
.global-filter-wrapper {
  display: flex;
  align-items: center;
  padding: 5px 24px;
}
.hide-tag-close {
  ::v-deep .el-tag__close {
    display: none !important;
  }
}
</style>
