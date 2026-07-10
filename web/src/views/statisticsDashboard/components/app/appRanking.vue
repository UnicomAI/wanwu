<template>
  <div class="ranking-card" v-loading="loading">
    <div class="ranking-title">
      {{ title }}
    </div>
    <div class="ranking-list">
      <div v-for="(item, index) in list" :key="index" class="ranking-item">
        <template>
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
              <template>
                <div v-if="item.userName">
                  {{ $t('statisticsDashboard.publisher') }}：{{ item.userName }}
                </div>
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
import { formatAmount, avatarSrc } from '@/utils/util.js';
import { AGENT, WORKFLOW } from '@/utils/commonSet';

export default {
  props: {
    title: {
      type: String,
      default: '',
    },
    dimension: {
      type: String,
      default: 'app',
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
        let key = '';
        let item = {};
        key = row.appId || row.appName || '--';
        item = {
          name: row.appName || '--',
          value: 0,
          appId: row.appId,
          userName: row.userName || '',
          avatar: this.getAppAvatar(row),
        };
        if (!groupMap[key]) {
          groupMap[key] = item;
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
    getAppAvatar(row) {
      const path = row.avatar;
      return path ? avatarSrc(path) : '';
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsRankingCard.scss';
</style>
