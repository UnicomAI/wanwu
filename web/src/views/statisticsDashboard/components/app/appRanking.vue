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
        <div class="ranking-value">{{ formatAmount(item.value) }}</div>
      </div>
      <div v-if="!list.length" class="ranking-empty">
        <el-empty :description="$t('common.noData')"></el-empty>
      </div>
    </div>
  </div>
</template>

<script>
import { formatAmount, avatarSrc } from '@/utils/util.js';

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
      const data = Array.isArray(this.data) ? this.data : [];
      return data.map(row => ({
        name: row.appName || '--',
        value: row.callCount,
        appId: row.appId,
        userName: row.moduleCreatorUserName || '--',
        avatar: row.avatar,
      }));
    },
  },
  methods: {
    avatarSrc,
    formatAmount,
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsRankingCard.scss';
</style>
