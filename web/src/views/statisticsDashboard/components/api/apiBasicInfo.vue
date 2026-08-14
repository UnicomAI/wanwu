<template>
  <div class="api-basic-info">
    <div class="section-title">
      {{ $t('statisticsDashboard.apiBasicInfo') }}
    </div>
    <el-table
      :data="[apiData]"
      :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
      style="width: 100%"
    >
      <el-table-column
        prop="apiName"
        :label="$t('statisticsDashboard.apiName')"
        align="left"
        min-width="140"
      >
        <template slot-scope="scope">
          {{ scope.row.apiName || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="apiKey"
        label="API Key"
        align="left"
        min-width="160"
      >
        <template slot-scope="scope">
          <span>
            {{
              scope.row.apiKey ? scope.row.apiKey.slice(0, 6) + '******' : '--'
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
          {{ scope.row.orgName || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="userName"
        :label="$t('statisticsDashboard.userName')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.userName || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="methodPath"
        :label="$t('statisticsDashboard.apiPath')"
        align="left"
        min-width="200"
      >
        <template slot-scope="scope">
          {{ scope.row.methodPath || '--' }}
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
export default {
  props: {
    apiInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  computed: {
    apiData() {
      return this.apiInfo || {};
    },
  },
};
</script>

<style lang="scss" scoped>
.api-basic-info {
  margin-top: 10px;
  margin-bottom: 16px;
  .section-title {
    font-size: 14px;
    font-weight: 600;
    color: $color_title;
    margin-bottom: 12px;
  }
}
</style>
