<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/rag">
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

    <!-- 模型配置 -->
    <DetailCard :title="ragDetail('modelConfig')">
      <InfoGrid :items="modelItems" />
    </DetailCard>

    <!-- 对话配置 -->
    <DetailCard :title="ragDetail('dialogConfig')">
      <div class="dialog-grid">
        <div class="dialog-row is-stack">
          <span class="dialog-label">{{ ragDetail('recommendQuestion') }}</span>
          <div class="dialog-value">
            <template v-if="recommendQuestions.length">
              <div
                v-for="(q, i) in recommendQuestions"
                :key="`rq-${i}`"
                class="recommend-item"
              >
                <span class="recommend-index">Q{{ i + 1 }}</span>
                <span class="recommend-text">{{ q }}</span>
              </div>
            </template>
            <div v-else class="empty-text">{{ ragDetail('noData') }}</div>
          </div>
        </div>
      </div>
    </DetailCard>

    <!-- 关联知识库 -->
    <DetailCard :title="ragDetail('knowledgeBase')">
      <template v-if="hasKnowledge">
        <div class="recall-config">
          <InfoGrid :items="knowledgeRecallItems" />
        </div>
        <div class="card-list">
          <div
            v-for="(kb, i) in knowledgeList"
            :key="`kb-${i}`"
            class="resource-card is-clickable"
            @click="goKnowledge(kb)"
          >
            <img
              :src="avatarSrc(kb.avatar?.path, knowledgeDefaultIcon)"
              class="resource-avatar"
            />
            <div class="resource-info">
              <div class="resource-name">{{ kb.name }}</div>
              <div v-if="kb.description" class="resource-desc">
                {{ kb.description }}
              </div>
              <div class="resource-meta">
                <span class="meta-tag">
                  {{
                    kb.share
                      ? $t('knowledgeManage.public')
                      : $t('knowledgeManage.private')
                  }}
                </span>
                <span v-if="kb.orgName" class="meta-tag">{{ kb.orgName }}</span>
                <span v-if="kb.external === 1" class="meta-tag">
                  {{ $t('knowledgeManage.ribbon.external') }}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="canExpandKnowledge" class="view-more">
          <span
            class="view-more-btn"
            @click="knowledgeExpanded = !knowledgeExpanded"
          >
            {{
              knowledgeExpanded ? ragDetail('collapse') : ragDetail('viewMore')
            }}
            <i
              :class="
                knowledgeExpanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'
              "
            ></i>
          </span>
        </div>
      </template>
      <div v-else class="empty-text">{{ ragDetail('noData') }}</div>
    </DetailCard>

    <!-- 关联问答库 -->
    <DetailCard :title="ragDetail('qaDatabase')">
      <template v-if="hasQa">
        <div class="recall-config">
          <InfoGrid :items="qaRecallItems" />
        </div>
        <div class="card-list">
          <div
            v-for="(kb, i) in qaList"
            :key="`qa-${i}`"
            class="resource-card is-clickable"
            @click="goKnowledge(kb)"
          >
            <img
              :src="avatarSrc(kb.avatar?.path, knowledgeDefaultIcon)"
              class="resource-avatar"
            />
            <div class="resource-info">
              <div class="resource-name">{{ kb.name }}</div>
              <div v-if="kb.description" class="resource-desc">
                {{ kb.description }}
              </div>
              <div class="resource-meta">
                <span class="meta-tag">
                  {{
                    kb.share
                      ? $t('knowledgeManage.public')
                      : $t('knowledgeManage.private')
                  }}
                </span>
                <span v-if="kb.orgName" class="meta-tag">{{ kb.orgName }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="canExpandQa" class="view-more">
          <span class="view-more-btn" @click="qaExpanded = !qaExpanded">
            {{ qaExpanded ? ragDetail('collapse') : ragDetail('viewMore') }}
            <i
              :class="qaExpanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'"
            ></i>
          </span>
        </div>
      </template>
      <div v-else class="empty-text">{{ ragDetail('noData') }}</div>
    </DetailCard>

    <!-- 安全护栏 -->
    <DetailCard :title="ragDetail('safetyGuardrail')">
      <InfoGrid :items="safetyItems" />
    </DetailCard>

    <!-- 视觉问答 -->
    <DetailCard :title="ragDetail('vision')">
      <InfoGrid :items="visionItems" />
    </DetailCard>
  </DetailLayout>
</template>

<script>
import { getAdminRagBase, getAdminRagDetail } from '@/api/adminCenter';
import { avatarSrc } from '@/utils/util';
import DetailLayout from '../components/DetailLayout.vue';
import DetailHeader from '../components/DetailHeader.vue';
import DetailCard from '../components/DetailCard.vue';
import InfoGrid from '../components/InfoGrid.vue';
import VisibleUsers from '../components/VisibleUsers.vue';
import detailMixin from '../mixins/detailMixin';

const VIEW_MORE_THRESHOLD = 6;

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
      knowledgeDefaultIcon: require('@/assets/imgs/knowledgeIcon.png'),
      knowledgeExpanded: false,
      qaExpanded: false,
      moduleTitleKey: 'adminCenter.pageModules.appDevelopment.rag.title',
    };
  },
  computed: {
    ragId() {
      return this.$route.query.appId;
    },
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/rag.svg'),
      );
    },
    headerTags() {
      const tags = [];
      if (this.base.publishStatus)
        tags.push({ text: this.statusText, type: '' });
      return tags;
    },
    modelConfig() {
      return this.detail.modelConfig ?? {};
    },
    modelLlm() {
      const cfg = this.modelConfig.config;
      if (cfg && typeof cfg === 'object') return cfg;
      return this.modelConfig;
    },
    modelConfigTags() {
      const tags = [];
      if (this.modelConfig.modelType)
        tags.push({ text: this.modelConfig.modelType });
      if (this.modelConfig.provider)
        tags.push({ text: this.modelConfig.provider });
      return tags;
    },
    modelItems() {
      const c = this.modelLlm;
      return [
        {
          label: this.ragDetail('model'),
          model: {
            icon: this.convertModelIcon(this.modelConfig.avatar?.path),
            name: this.modelConfig.displayName,
            tags: this.modelConfigTags,
          },
        },
        {
          label: this.ragDetail('temperature'),
          value: this.numText(c.temperature),
        },
        { label: this.ragDetail('topP'), value: this.numText(c.topP) },
        {
          label: this.ragDetail('frequencyPenalty'),
          value: this.numText(c.frequencyPenalty),
        },
        {
          label: this.ragDetail('thinking'),
          value: this.boolText(c.thinkingEnable),
        },
      ];
    },
    knowledgeConfig() {
      return this.detail.knowledgeBaseConfig ?? {};
    },
    knowledgeRecall() {
      return this.knowledgeConfig.config ?? {};
    },
    knowledgeList() {
      const list = this.knowledgeConfig.knowledgebases || [];
      return this.knowledgeExpanded ? list : list.slice(0, VIEW_MORE_THRESHOLD);
    },
    hasKnowledge() {
      return (this.knowledgeConfig.knowledgebases || []).length > 0;
    },
    canExpandKnowledge() {
      return (
        (this.knowledgeConfig.knowledgebases || []).length > VIEW_MORE_THRESHOLD
      );
    },
    matchTypeText() {
      const map = {
        vector: this.ragDetail('matchTypeOptions.vector'),
        text: this.ragDetail('matchTypeOptions.text'),
        mix: this.ragDetail('matchTypeOptions.mix'),
      };
      return (
        map[this.knowledgeRecall.matchType] ??
        this.knowledgeRecall.matchType ??
        '-'
      );
    },
    isMixMatch() {
      return this.knowledgeRecall.matchType === 'mix';
    },
    knowledgeRecallItems() {
      const c = this.knowledgeRecall;
      const items = [
        { label: this.ragDetail('matchType'), value: this.matchTypeText },
        { label: this.ragDetail('topK'), value: this.numText(c.topK) },
        {
          label: this.ragDetail('scoreThreshold'),
          value: this.numText(c.threshold),
        },
      ];
      if (this.isMixMatch) {
        items.push(
          {
            label: this.ragDetail('semanticsPriority'),
            value: this.numText(c.semanticsPriority),
          },
          {
            label: this.ragDetail('keywordPriority'),
            value: this.numText(c.keywordPriority),
          },
        );
      }
      items.push({
        label: this.ragDetail('useGraph'),
        value: this.boolText(c.useGraph),
      });
      if (this.rerankConfig.displayName) {
        items.push({
          label: this.ragDetail('rerankModel'),
          model: {
            icon: this.convertModelIcon(this.rerankConfig.avatar?.path),
            name: this.rerankConfig.displayName,
            tags: this.rerankModelTags,
          },
        });
      }
      return items;
    },
    rerankConfig() {
      return this.detail.rerankConfig ?? {};
    },
    rerankModelTags() {
      const tags = [];
      if (this.rerankConfig.modelType)
        tags.push({ text: this.rerankConfig.modelType });
      if (this.rerankConfig.provider)
        tags.push({ text: this.rerankConfig.provider });
      return tags;
    },
    qaConfig() {
      return this.detail.qaKnowledgeBaseConfig ?? {};
    },
    qaRecall() {
      return this.qaConfig.config ?? {};
    },
    qaList() {
      const list = this.qaConfig.knowledgebases || [];
      return this.qaExpanded ? list : list.slice(0, VIEW_MORE_THRESHOLD);
    },
    hasQa() {
      return (this.qaConfig.knowledgebases || []).length > 0;
    },
    canExpandQa() {
      return (this.qaConfig.knowledgebases || []).length > VIEW_MORE_THRESHOLD;
    },
    qaMatchTypeText() {
      const map = {
        vector: this.ragDetail('matchTypeOptions.vector'),
        text: this.ragDetail('matchTypeOptions.text'),
        mix: this.ragDetail('matchTypeOptions.mix'),
      };
      return map[this.qaRecall.matchType] ?? this.qaRecall.matchType ?? '-';
    },
    qaRecallItems() {
      const c = this.qaRecall;
      const items = [
        { label: this.ragDetail('matchType'), value: this.qaMatchTypeText },
        { label: this.ragDetail('topK'), value: this.numText(c.topK) },
        {
          label: this.ragDetail('scoreThreshold'),
          value: this.numText(c.threshold),
        },
      ];
      if (c.maxHistory !== undefined) {
        items.push({
          label: this.ragDetail('maxHistory'),
          value: this.numText(c.maxHistory),
        });
      }
      if (this.qaRerankConfig.displayName) {
        items.push({
          label: this.ragDetail('rerankModel'),
          model: {
            icon: this.convertModelIcon(this.qaRerankConfig.avatar?.path),
            name: this.qaRerankConfig.displayName,
            tags: this.qaRerankModelTags,
          },
        });
      }
      return items;
    },
    qaRerankConfig() {
      return this.detail.qaRerankConfig ?? {};
    },
    qaRerankModelTags() {
      const tags = [];
      if (this.qaRerankConfig.modelType)
        tags.push({ text: this.qaRerankConfig.modelType });
      if (this.qaRerankConfig.provider)
        tags.push({ text: this.qaRerankConfig.provider });
      return tags;
    },
    safetyConfig() {
      return this.detail.safetyConfig ?? {};
    },
    safetyItems() {
      const items = [
        {
          label: this.$t('adminCenter.common.status'),
          value: this.boolText(this.safetyConfig.enable),
        },
      ];
      if (this.safetyConfig.enable) {
        items.push({
          label: this.ragDetail('sensitiveTables'),
          tags: (this.safetyConfig.tables || []).map(t => ({
            text: t.tableName,
            url: t.tableId
              ? `/adminCenter/safety/detail?tableId=${t.tableId}`
              : '',
          })),
        });
      }
      return items;
    },
    recommendQuestions() {
      return (this.detail.recommendQuestion || []).filter(Boolean);
    },
    visionConfig() {
      return this.detail.visionConfig ?? {};
    },
    visionItems() {
      return [
        {
          label: this.$t('adminCenter.common.status'),
          value: this.boolText(this.visionConfig.picNum === 1),
        },
      ];
    },
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    ragDetail(key) {
      return this.$t(
        `adminCenter.pageModules.appDevelopment.rag.detail.${key}`,
      );
    },
    goKnowledge(kb) {
      if (!kb?.id) return;
      const routeUrl = this.$router.resolve({
        path: '/adminCenter/knowledge/detail',
        query: { knowledgeId: kb.id },
      });
      window.open(routeUrl.href, '_blank');
    },
    fetchData() {
      if (!this.ragId) return;
      const params = { ragId: this.ragId };
      getAdminRagBase(params).then(res => {
        this.base = res?.data ?? {};
      });
      getAdminRagDetail(params).then(res => {
        this.detail = res?.data ?? {};
      });
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/views/adminCenter/styles/common.scss';
</style>
