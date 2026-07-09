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
