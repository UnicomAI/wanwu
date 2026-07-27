<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/tool">
    <DetailHeader
      :avatar="avatarUrl"
      :description="base.desc"
      :name="base.name"
    />

    <DetailCard :title="$t('adminCenter.common.basicInfo')">
      <InfoGrid :items="basicItems" />
      <VisibleUsers :users="visibleUsers" />
    </DetailCard>

    <DetailCard :title="$t('adminCenter.common.config')">
      <el-form
        class="tool-config-form"
        label-position="left"
        label-width="120px"
      >
        <el-form-item :label="$t('tool.custom.apiAuth')">
          <span class="config-value">{{ authTypeText }}</span>
        </el-form-item>
        <el-form-item label="Schema">
          <pre class="schema-code">{{ prettySchema }}</pre>
        </el-form-item>
        <el-form-item :label="$t('tool.custom.api')">
          <el-table
            :data="config.apiList"
            :header-cell-style="{ textAlign: 'center' }"
            border
            size="mini"
          >
            <el-table-column
              :label="$t('adminCenter.common.name')"
              prop="name"
            />
            <el-table-column
              :label="
                $t('adminCenter.pageModules.resourcePool.tool.detail.apiMethod')
              "
              prop="method"
            />
            <el-table-column
              :label="
                $t('adminCenter.pageModules.resourcePool.tool.detail.apiPath')
              "
              prop="path"
            />
          </el-table>
        </el-form-item>
        <el-form-item :label="$t('tool.custom.privacy')">
          <span class="config-value">{{ config.privacyPolicy || '-' }}</span>
        </el-form-item>
      </el-form>
    </DetailCard>
  </DetailLayout>
</template>

<script>
import { getAdminToolBase, getAdminToolDetail } from '@/api/adminCenter';
import { avatarSrc } from '@/utils/util';
import DetailLayout from '../components/DetailLayout.vue';
import DetailHeader from '../components/DetailHeader.vue';
import DetailCard from '../components/DetailCard.vue';
import InfoGrid from '../components/InfoGrid.vue';
import VisibleUsers from '../components/VisibleUsers.vue';
import detailMixin from '../mixins/detailMixin';

export default {
  components: {
    DetailLayout,
    DetailHeader,
    DetailCard,
    InfoGrid,
    VisibleUsers,
  },
  mixins: [detailMixin],
  data() {
    return {
      base: {},
      config: {},
      moduleTitleKey: 'adminCenter.pageModules.resourcePool.tool.title',
    };
  },
  computed: {
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/toolImg.png'),
      );
    },
    toolId() {
      return this.$route.query.toolId;
    },
    toolType() {
      return this.$route.query.type || 'custom';
    },
    authTypeText() {
      const map = {
        none: this.$t('tool.custom.auth.noneType'),
        api_key_header: this.$t('tool.custom.auth.headerType'),
        api_key_query: this.$t('tool.custom.auth.queryType'),
      };
      const type = this.config.apiAuth?.authType;
      return type ? (map[type] ?? type) : '-';
    },
    prettySchema() {
      if (!this.config.schema) return '-';
      try {
        return JSON.stringify(JSON.parse(this.config.schema), null, 2);
      } catch (e) {
        return this.config.schema;
      }
    },
  },
  mounted() {
    this.fetchBase();
  },
  methods: {
    fetchBase() {
      if (!this.toolId) return;
      const params = { toolId: this.toolId, type: this.toolType };
      Promise.all([getAdminToolBase(params), getAdminToolDetail(params)]).then(
        ([baseRes, detailRes]) => {
          this.base = baseRes?.data ?? {};
          this.config = detailRes?.data ?? {};
        },
      );
    },
  },
};
</script>

<style lang="scss" scoped>
.tool-config-form {
  .config-value {
    font-size: 14px;
    color: #303133;
    word-break: break-all;
  }

  .schema-code {
    margin: 0;
    padding: 16px;
    background: #f5f7fa;
    border-radius: 4px;
    font-family: 'Courier New', Consolas, monospace;
    font-size: 13px;
    color: #303133;
    line-height: 1.6;
    max-height: 360px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }
}
</style>
