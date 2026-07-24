<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/modelConfig">
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
      <InfoGrid :items="configItems" />
    </DetailCard>
  </DetailLayout>
</template>

<script>
import { getAdminModelBase, getAdminModelDetail } from '@/api/adminCenter';
import { FUNC_CALLING, SUPPORT_LIST } from '@/views/modelAccess/constants.js';
import { avatarSrc, getModelDefaultIcon } from '@/utils/util';
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
      detail: {},
      moduleTitleKey: 'adminCenter.pageModules.modelService.title',
    };
  },
  computed: {
    avatarUrl() {
      return avatarSrc(this.base.avatar?.path, getModelDefaultIcon());
    },
    headerTags() {
      const tags = [];
      if (this.base.publishStatus) {
        tags.push({ text: this.statusText, type: '' });
      }
      return tags;
    },
    config() {
      return this.detail?.config ?? {};
    },
    configItems() {
      const d = this.detail;
      const c = this.config;
      const label = key =>
        this.$t(`adminCenter.pageModules.modelService.detail.${key}`);
      return [
        { label: label('modelId'), value: d.model },
        {
          label: label('functionCall'),
          value: this.optionText(FUNC_CALLING, c.functionCalling),
        },
        {
          label: label('vision'),
          value: this.optionText(SUPPORT_LIST, c.visionSupport),
        },
        { label: label('contextLength'), value: c.contextSize },
        { label: label('maxToken'), value: c.maxTokens },
        { label: label('inferenceUrl'), value: c.endpointUrl },
        {
          label: label('reasoningSwitch'),
          value: this.optionText(SUPPORT_LIST, c.thinkingSupport),
        },
        { label: 'uuid', value: d.uuid },
      ].filter(
        it => it.value !== undefined && it.value !== null && it.value !== '',
      );
    },
  },
  mounted() {
    const modelId = this.$route.query.modelId;
    if (!modelId) return;
    this.fetchBase(modelId);
    this.fetchDetail(modelId);
  },
  methods: {
    optionText(list, value) {
      const item = list.find(it => it.key === value);
      return item ? item.name : value;
    },
    fetchBase(modelId) {
      getAdminModelBase({ modelId }).then(res => {
        this.base = res?.data ?? {};
      });
    },
    fetchDetail(modelId) {
      getAdminModelDetail({ modelId }).then(res => {
        this.detail = res?.data ?? {};
      });
    },
  },
};
</script>
