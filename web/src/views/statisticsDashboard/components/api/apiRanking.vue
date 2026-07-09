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
.ranking-card {
  width: 40%;
  background: #fff;
  border-radius: 5px;
  padding: 20px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.08);
  .ranking-title {
    font-size: 14px;
    font-weight: bold;
    margin-bottom: 16px;
  }
  .ranking-list {
    .ranking-item {
      display: flex;
      align-items: center;
      padding: 10px 0;
      border-bottom: 1px solid #f0f0f0;
      &:last-child {
        border-bottom: none;
      }
      .ranking-avatar-wrap {
        position: relative;
        width: 40px;
        height: 40px;
        margin-right: 12px;
        flex-shrink: 0;
        .ranking-avatar {
          width: 40px;
          height: 40px;
          border-radius: 50%;
          object-fit: cover;
        }
        .ranking-badge {
          position: absolute;
          top: -4px;
          right: -4px;
          width: 16px;
          height: 16px;
          line-height: 16px;
          text-align: center;
          border-radius: 50%;
          color: #fff;
          font-size: 10px;
          font-weight: bold;
          border: 2px solid #fff;
          box-sizing: content-box;
          &-1 {
            background: #ff4d4f;
          }
          &-2 {
            background: #ff7a45;
          }
          &-3 {
            background: #ffa940;
          }
        }
      }
      .ranking-info {
        flex: 1;
        min-width: 0;
        .ranking-name-row {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 4px;
          .ranking-name {
            font-size: 14px;
            color: #303133;
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }
        }
        .ranking-desc {
          font-size: 12px;
          color: #909399;
          line-height: 1.4;
          > div {
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }
        }
      }
      .ranking-value {
        font-size: 14px;
        color: $color;
        font-weight: bold;
        margin-left: 12px;
        white-space: nowrap;
      }
    }
    .ranking-empty {
      text-align: center;
      padding: 30px 0;
      color: #909399;
      font-size: 14px;
    }
  }
}
</style>
