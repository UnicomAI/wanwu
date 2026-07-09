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
.ranking-card {
  width: 100%;
  background: #fff;
  border-radius: 5px;
  padding: 20px;
  box-shadow: 0 2px 8px 0 rgba(0, 0, 0, 0.06);
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
      .ranking-index {
        width: 24px;
        height: 24px;
        line-height: 24px;
        text-align: center;
        border-radius: 50%;
        background: #f0f2f5;
        color: #666;
        font-size: 12px;
        margin-right: 12px;
        flex-shrink: 0;
        &.ranking-index-first {
          background: #4c7cf6;
          color: #fff;
        }
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
          .ranking-tag {
            flex-shrink: 0;
            padding: 0 6px;
            height: 18px;
            line-height: 18px;
            background: $tag_bg;
            color: $tag_color;
            font-size: 11px;
            border-radius: 2px;
          }
        }
        .ranking-desc {
          font-size: 12px;
          color: #909399;
          line-height: 1.4;
          span {
            margin-right: 12px;
            &:last-child {
              margin-right: 0;
            }
          }
        }
      }
      .ranking-value {
        font-size: 14px;
        color: #303133;
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
