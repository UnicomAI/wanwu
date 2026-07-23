<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/prompt">
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
      <div class="config-row">
        <span class="config-label">
          {{ $t('adminCenter.pageModules.resourcePool.prompt.detail.prompt') }}
        </span>
        <div class="prompt-text input-box">
          {{ detail.prompt }}
        </div>
      </div>
    </DetailCard>
  </DetailLayout>
</template>

<script>
import { getAdminPromptBase, getAdminPromptDetail } from '@/api/adminCenter';
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
      detail: {},
      moduleTitleKey: 'adminCenter.pageModules.resourcePool.prompt.title',
    };
  },
  computed: {
    customPromptId() {
      return this.$route.query.promptId;
    },
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/prompt.png'),
      );
    },
    headerTags() {
      const tags = [];
      if (this.base.publishStatus) {
        tags.push({ text: this.statusText, type: '' });
      }
      return tags;
    },
  },
  mounted() {
    this.fetchBase();
  },
  methods: {
    fetchBase() {
      if (!this.customPromptId) return;
      getAdminPromptBase({ customPromptId: this.customPromptId }).then(res => {
        this.base = res?.data ?? {};
      });
      getAdminPromptDetail({ customPromptId: this.customPromptId }).then(
        res => {
          this.detail = res?.data ?? {};
        },
      );
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/views/adminCenter/styles/common.scss';

.config-row {
  display: flex;
  align-items: center;
}

.config-label {
  color: #909399;
  font-size: 14px;
  width: 120px;
  flex-shrink: 0;
}
</style>
