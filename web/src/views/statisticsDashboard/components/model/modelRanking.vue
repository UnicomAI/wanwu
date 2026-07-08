<template>
  <div class="ranking-card" v-loading="loading">
    <div class="ranking-title">
      {{ title }}
    </div>
    <div class="ranking-list">
      <div v-for="(item, index) in list" :key="index" class="ranking-item">
        <div class="ranking-index">{{ index + 1 }}</div>
        <div class="ranking-info">
          <span class="ranking-name">{{ item.name }}</span>
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
  },
  computed: {
    list() {
      const groupMap = {};
      this.data.forEach(row => {
        let key = '';
        let name = '';
        if (this.dimension === 'model') {
          key = row.model || '未知';
          name = row.model || '未知';
        } else if (this.dimension === 'user') {
          key = row.userName || '未知';
          name = row.userName || '未知';
        } else if (this.dimension === 'org') {
          key = row.orgName || '未知';
          name = row.orgName || '未知';
        }
        if (!groupMap[key]) {
          groupMap[key] = { name, value: 0 };
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
  },
};
</script>

<style lang="scss" scoped>
.ranking-card {
  width: 100%;
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
      }
      .ranking-info {
        flex: 1;
        .ranking-name {
          font-size: 14px;
          color: #303133;
        }
      }
      .ranking-value {
        font-size: 14px;
        color: #303133;
        font-weight: bold;
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
