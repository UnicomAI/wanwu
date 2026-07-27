<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/skill">
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

    <DetailCard
      :title="$t('adminCenter.pageModules.resourcePool.skill.detail.overview')"
    >
      <SkillDetail
        v-if="skillDetail?.skillMarkdown"
        :detail="skillDetail"
        :recommendList="[]"
        :visibleDownload="false"
        :visibleHistory="false"
        :visibleVariableConfig="false"
        readonly
      />
      <EmptyState
        v-else
        :tip="$t('adminCenter.pageModules.resourcePool.skill.detail.emptyTip')"
      />
    </DetailCard>
  </DetailLayout>
</template>

<script>
import { getAdminSkillBase, getAdminSkillDetail } from '@/api/adminCenter';
import { avatarSrc } from '@/utils/util';
import SkillDetail from '@/components/skills/skillDetail.vue';
import DetailLayout from '../components/DetailLayout.vue';
import DetailHeader from '../components/DetailHeader.vue';
import DetailCard from '../components/DetailCard.vue';
import InfoGrid from '../components/InfoGrid.vue';
import VisibleUsers from '../components/VisibleUsers.vue';
import EmptyState from '../components/EmptyState.vue';
import detailMixin from '../mixins/detailMixin';

export default {
  components: {
    DetailLayout,
    DetailHeader,
    DetailCard,
    InfoGrid,
    VisibleUsers,
    EmptyState,
    SkillDetail,
  },
  mixins: [detailMixin],
  data() {
    return {
      base: {},
      skillDetail: {},
      moduleTitleKey: 'adminCenter.pageModules.resourcePool.skill.title',
    };
  },
  computed: {
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/custom-skill-default-icon.png'),
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
    const skillId = this.$route.query.skillId;
    if (!skillId) return;
    const skillType = this.$route.query.type || '';
    const params = { skillId, skillType };
    this.fetchBase(params);
    this.fetchSkillDetail(params);
  },
  methods: {
    fetchBase(params) {
      getAdminSkillBase(params).then(res => {
        this.base = res?.data ?? {};
      });
    },
    fetchSkillDetail(params) {
      getAdminSkillDetail(params).then(res => {
        this.skillDetail = res?.data ?? {};
      });
    },
  },
};
</script>

<style lang="scss" scoped>
::v-deep .tempSquare-detail {
  padding: 0 !important;
}
</style>
