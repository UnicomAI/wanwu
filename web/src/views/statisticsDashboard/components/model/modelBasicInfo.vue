<template>
  <div class="model-basic-info">
    <div class="section-title">
      {{ $t('statisticsDashboard.modelBasicInfo') }}
    </div>
    <el-table
      :data="[modelData]"
      :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
      style="width: 100%"
    >
      <el-table-column
        prop="model"
        :label="$t('statisticsDashboard.modelName')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.model || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="provider"
        :label="$t('statisticsDashboard.provider')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ providerObj[scope.row.provider] || scope.row.provider || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="modelCreatorUserName"
        :label="$t('statisticsDashboard.modelPublisher')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.modelCreatorUserName || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="modelCreatorOrgName"
        :label="$t('statisticsDashboard.modelOrg')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          {{ scope.row.modelCreatorOrgName || '--' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="modelType"
        :label="$t('statisticsDashboard.modelType')"
        align="left"
        min-width="120"
      >
        <template slot-scope="scope">
          <ModelTypeTag :model-type="scope.row.modelType" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import ModelTypeTag from './modelTypeTag.vue';
import { PROVIDER_OBJ } from '@/views/modelAccess/constants';

export default {
  components: { ModelTypeTag },
  props: {
    modelInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      providerObj: PROVIDER_OBJ,
    };
  },
  computed: {
    modelData() {
      return this.modelInfo || {};
    },
  },
};
</script>

<style lang="scss" scoped>
.model-basic-info {
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
