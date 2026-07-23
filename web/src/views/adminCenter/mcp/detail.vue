<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/mcp">
    <DetailHeader
      :avatar="avatarUrl"
      :description="base.desc"
      :name="base.name"
      :tags="headerTags"
    />

    <DetailCard :title="$t('adminCenter.common.basicInfo')">
      <InfoGrid :items="basicItems" />
      <VisibleUsers :users="visibleUsers" />
    </DetailCard>

    <DetailCard :title="$t('adminCenter.common.config')">
      <McpServiceServerDetail
        v-if="mcpType === 'mcpserver'"
        :embedded-detail="config"
        readonly
      />
      <McpServiceCustomDetail v-else :embedded-detail="config" readonly />
    </DetailCard>
  </DetailLayout>
</template>

<script>
import {
  getAdminMcpBase,
  getAdminMcpCustomDetail,
  getAdminMcpServerDetail,
  getAdminMcpToolList,
} from '@/api/adminCenter';
import { avatarSrc } from '@/utils/util';
import McpServiceCustomDetail from '@/views/mcpManagementPublic/detail.vue';
import McpServiceServerDetail from '@/views/tool/mcp/server/detail.vue';
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
    McpServiceCustomDetail,
    McpServiceServerDetail,
  },
  mixins: [detailMixin],
  data() {
    return {
      base: {},
      config: {},
      moduleTitleKey: 'adminCenter.pageModules.resourcePool.mcp.title',
    };
  },
  computed: {
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/mcp_active.svg'),
      );
    },
    mcpId() {
      return this.$route.query.mcpId;
    },
    mcpType() {
      return this.$route.query.type;
    },
    typeText() {
      const map = {
        mcp: this.$t(
          'adminCenter.pageModules.resourcePool.mcp.detail.typeCustom',
        ),
        mcpserver: this.$t(
          'adminCenter.pageModules.resourcePool.mcp.detail.typeServer',
        ),
      };
      return map[this.base.type] ?? this.base.type ?? '-';
    },
    headerTags() {
      const tags = [];
      if (this.base.publishStatus) {
        tags.push({ text: this.statusText, type: '' });
      }
      if (this.base.type) {
        tags.push({ text: this.typeText, type: 'info' });
      }
      return tags;
    },
  },
  mounted() {
    this.fetchBase();
  },
  methods: {
    fetchBase() {
      if (!this.mcpId || !this.mcpType) return;
      getAdminMcpBase({ mcpId: this.mcpId, type: this.mcpType }).then(res => {
        this.base = res?.data ?? {};
        this.fetchConfig();
      });
    },
    async fetchConfig() {
      const fetcher =
        this.mcpType === 'mcpserver'
          ? getAdminMcpServerDetail
          : getAdminMcpCustomDetail;
      const res = await fetcher({ mcpId: this.mcpId });
      this.config = res?.data ?? {};
      if (this.config?.tools) return;
      const toolRes = await getAdminMcpToolList({
        mcpId: this.mcpId,
        type: this.mcpType,
      });
      this.config = { ...this.config, tools: toolRes?.data?.tools ?? {} };
    },
  },
};
</script>
