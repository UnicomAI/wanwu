<template>
  <div class="app-basic-info">
    <div class="section-title">
      {{ $t('statisticsDashboard.appBasicInfo') }}
    </div>
    <el-table
      :data="[appData]"
      :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
      style="width: 100%"
    >
      <el-table-column
        prop="source"
        :label="$t('statisticsDashboard.source')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.sourceName || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="module"
        :label="$t('statisticsDashboard.module')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          <AppTypeTag :app-type="scope.row.module" />
        </template>
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
          <AppTypeTag :app-type="scope.row.appType" />
        </template>
      </el-table-column>
      <el-table-column
        prop="moduleCreatorUserName"
        :label="$t('statisticsDashboard.appAuthor')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.moduleCreatorUserName || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="moduleCreatorOrgName"
        :label="$t('statisticsDashboard.appAuthorOrg')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.moduleCreatorOrgName || '--' }}
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import AppTypeTag from './appTypeTag.vue';

export default {
  components: { AppTypeTag },
  props: {
    appInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  computed: {
    appData() {
      return this.appInfo || {};
    },
  },
};
</script>

<style lang="scss" scoped>
.app-basic-info {
  margin-top: 10px;
  margin-bottom: 16px;
  .section-title {
    font-size: 14px;
    font-weight: 600;
    color: #333;
    margin-bottom: 12px;
  }
}
</style>
