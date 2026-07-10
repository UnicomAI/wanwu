<template>
  <div class="ranking-card" v-loading="loading">
    <div class="ranking-title">
      {{ title }}
    </div>
    <div class="ranking-list">
      <div v-for="(item, index) in list" :key="index" class="ranking-item">
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
          </div>
          <div class="ranking-desc">
            <div v-if="item.description">
              {{ item.description }}
            </div>
            <div v-if="item.userName">
              {{ $t('statisticsDashboard.publisher') }}:{{ item.userName }}
            </div>
          </div>
        </div>
        <div class="ranking-value">{{ formatValue(item.value) }}</div>
      </div>
      <div v-if="!list.length" class="ranking-empty">
        {{ $t('common.noData') }}
      </div>
    </div>
  </div>
</template>

<script>
import { formatAmount } from '@/utils/util.js';

export default {
  props: {
    title: {
      type: String,
      default: '',
    },
    data: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    defaultAvatar() {
      return require('@/assets/imgs/avatar_default.png');
    },
    list() {
      const groupMap = {};
      const data = Array.isArray(this.data) ? this.data : [];
      data.forEach(row => {
        if (!row) return;
        const key = row.apiKeyId || row.name || '--';
        if (!groupMap[key]) {
          groupMap[key] = {
            name: row.name || '--',
            value: 0,
            apiKeyId: row.apiKeyId,
            description: row.description || '',
            userName: row.userName || '',
          };
        }
        groupMap[key].value += Number(row.callCount || 0);
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
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsRankingCard.scss';
.ranking-card {
  width: 40% !important;
}
</style>
