<template>
  <div class="ranking-card" v-loading="loading">
    <div class="ranking-title">
      {{ title }}
    </div>
    <div class="ranking-list">
      <div
        v-for="(item, index) in list"
        :key="index"
        class="ranking-item"
        :class="{ 'org-dimension': dimension === 'org' }"
      >
        <template v-if="dimension === 'org'">
          <div
            class="ranking-index"
            :class="{ 'ranking-index-first': index === 0 }"
          >
            {{ index + 1 }}
          </div>
          <div class="ranking-info">
            <span class="ranking-name">{{ item.name }}</span>
          </div>
        </template>
        <template v-else>
          <div class="ranking-avatar-wrap">
            <img
              class="ranking-avatar"
              :src="
                item.avatar?.path ? avatarSrc(item.avatar.path) : defaultAvatar
              "
              alt=""
            />
            <div
              v-if="index < 3"
              class="ranking-badge"
              :class="`ranking-badge-${index + 1}`"
            >
              {{ index + 1 }}
            </div>
          </div>
          <div class="ranking-info">
            <div class="ranking-name-row">
              <span class="ranking-name">{{ item.name }}</span>
              <span
                v-if="dimension === 'model' && item.modelTypeName"
                class="ranking-tag"
              >
                {{ item.modelTypeName }}
              </span>
            </div>
            <div class="ranking-desc">
              <template v-if="dimension === 'model'">
                <div v-if="item.provider">
                  {{ item.provider }}
                </div>
                <div v-if="item.publisher">
                  {{ $t('statisticsDashboard.publisher') }}：{{
                    item.publisher
                  }}
                </div>
              </template>
              <template v-else-if="dimension === 'user'">
                <span v-if="item.orgName">
                  {{ item.orgName }}
                </span>
              </template>
            </div>
          </div>
        </template>
        <div class="ranking-value">{{ formatAmount(item.value) }}</div>
      </div>
      <div v-if="!list.length" class="ranking-empty">
        <el-empty :description="$t('common.noData')"></el-empty>
      </div>
    </div>
  </div>
</template>

<script>
import { formatAmount, avatarSrc, getModelDefaultIcon } from '@/utils/util.js';
import { MODEL_TYPE_OBJ, PROVIDER_OBJ } from '@/views/modelAccess/constants';

export default {
  props: {
    title: {
      type: String,
      default: '',
    },
    dimension: {
      type: String,
      default: 'model',
    },
    data: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
    modelMap: {
      type: Object,
      default: () => ({}),
    },
  },
  computed: {
    defaultAvatar() {
      return require('@/assets/imgs/avatar_default.png');
    },
    list() {
      const data = Array.isArray(this.data) ? this.data : [];
      return data.map(row => {
        if (this.dimension === 'model') {
          return {
            name: row.model || '--',
            value: row.totalTokens,
            modelId: row.modelId,
            provider: PROVIDER_OBJ[row.provider] || row.provider || '--',
            publisher: row.modelCreatorUserName || '--',
            modelTypeName: this.getModelTypeName(row.modelType),
            avatar: row.modelAvatar,
          };
        } else if (this.dimension === 'user') {
          return {
            name: row.userName || '--',
            value: row.totalTokens,
            userId: row.userId,
            orgName: row.orgName || '--',
            avatar: row.avatar,
          };
        } else if (this.dimension === 'org') {
          return {
            name: row.orgName || '--',
            value: row.totalTokens,
            orgId: row.orgId,
          };
        }
      });
    },
  },
  methods: {
    avatarSrc,
    formatAmount,
    getModelTypeName(modelType) {
      return MODEL_TYPE_OBJ[modelType] || modelType || '';
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsRankingCard.scss';
</style>
