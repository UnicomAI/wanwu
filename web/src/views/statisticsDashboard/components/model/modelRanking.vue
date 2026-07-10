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
              :src="item.avatar || defaultAvatar"
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
        <div class="ranking-value">{{ formatValue(item.value) }}</div>
      </div>
      <div v-if="!list.length" class="ranking-empty">
        {{ $t('common.noData') }}
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
      const groupMap = {};
      this.data.forEach(row => {
        let key = '';
        let item = {};
        if (this.dimension === 'model') {
          const modelId = row.modelId || row.model;
          const modelInfo = this.modelMap[modelId] || {};
          key = row.model || '--';
          item = {
            name: row.model || '--',
            value: 0,
            modelId,
            provider: PROVIDER_OBJ[row.provider] || row.provider || '--',
            publisher: this.buildPublisher(row),
            modelTypeName: this.getModelTypeName(
              row.modelType || modelInfo.modelType,
            ),
            avatar: this.getModelAvatar(modelInfo, row),
          };
        } else if (this.dimension === 'user') {
          key = row.userId || row.userName || '--';
          item = {
            name: row.userName || '--',
            value: 0,
            userId: row.userId,
            orgName: row.orgName || '--',
            avatar: this.getUserAvatar(row),
          };
        } else if (this.dimension === 'org') {
          key = row.orgId || row.orgName || '--';
          item = {
            name: row.orgName || '--',
            value: 0,
            orgId: row.orgId,
          };
        }
        if (!groupMap[key]) {
          groupMap[key] = item;
        }
        groupMap[key].value += Number(row.totalTokens || 0);
      });
      return Object.values(groupMap)
        .sort((a, b) => b.value - a.value)
        .slice(0, 5);
    },
  },
  methods: {
    formatValue(val) {
      return formatAmount(val);
    },
    buildPublisher(row) {
      const orgName = row.orgName || '';
      const userName = row.userName || '';
      if (orgName && userName) {
        return `${orgName} ${userName}`;
      }
      return orgName || userName || '--';
    },
    getModelTypeName(modelType) {
      return MODEL_TYPE_OBJ[modelType] || modelType || '';
    },
    getModelAvatar(modelInfo, row) {
      const path = row.modelAvatar || row.avatar || modelInfo?.avatar?.path;
      return path ? avatarSrc(path) : getModelDefaultIcon();
    },
    getUserAvatar(row) {
      const path = row.userAvatar || row.avatar;
      return path ? avatarSrc(path) : '';
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsRankingCard.scss';
</style>
